package domain

import (
	"errors"
	"slices"
	"time"
)

var (
	ErrInvalidMembership  = errors.New("invalid membership")
	ErrMembershipNotFound = errors.New("membership not found")
	ErrMembershipEnded    = errors.New("the membership has ended")
	ErrAccessDenied       = errors.New("access denied")
)

// Membership puts a person in an area with some of its roles and loose
// permissions. A zero Until means it holds until it is ended. GrantedAt is
// when it was granted, which From may follow; it is zero in those granted
// before SIGTA recorded it.
type Membership struct {
	ID          string
	UserID      string
	Area        string
	Roles       []string
	Permissions []string
	Scope       []int64
	From        time.Time
	Until       time.Time
	GrantedBy   string
	GrantedAt   time.Time
	EndedBy     string
}

// Covers tells whether the scope includes the procedure type. An empty scope
// covers every type, including those published later.
func (m Membership) Covers(procedureType int64) bool {
	return len(m.Scope) == 0 || slices.Contains(m.Scope, procedureType)
}

func (m Membership) ActiveAt(at time.Time) bool {
	return !at.Before(m.From) && (m.Until.IsZero() || at.Before(m.Until))
}

// Overlaps tells whether two memberships of the same person would hold at once
// in the same area with the same scope. A different scope is allowed, so that
// someone can have more permissions on some procedure types.
func (m Membership) Overlaps(other Membership) bool {
	if m.UserID != other.UserID || m.Area != other.Area || !sameScope(m.Scope, other.Scope) {
		return false
	}
	startsBeforeOtherEnds := other.Until.IsZero() || m.From.Before(other.Until)
	endsAfterOtherStarts := m.Until.IsZero() || other.From.Before(m.Until)
	return startsBeforeOtherEnds && endsAfterOtherStarts
}

func sameScope(a, b []int64) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(slices.Compact(a), slices.Compact(b))
}

// Access is what a person may do at a given moment: only their active
// memberships count, with the permissions that their roles and loose
// permissions give. Being an administrator gives no access to procedures.
type Access struct {
	Administrator bool
	Memberships   []Membership
	granted       [][]string
}

// NewAccess takes the areas to resolve the roles of each membership: a role
// counts with the permissions it has now.
func NewAccess(administrator bool, memberships []Membership, areas []Area, at time.Time) Access {
	active := slices.DeleteFunc(slices.Clone(memberships), func(m Membership) bool { return !m.ActiveAt(at) })
	granted := make([][]string, len(active))
	for i, m := range active {
		if j := slices.IndexFunc(areas, func(a Area) bool { return a.ID == m.Area }); j >= 0 {
			granted[i] = areas[j].GrantedBy(m.Roles, m.Permissions)
		}
	}
	return Access{Administrator: administrator, Memberships: active, granted: granted}
}

// Holds tells whether the memberships in the area that cover the procedure
// type give, together, every permission asked for.
func (a Access) Holds(area string, procedureType int64, permissions []string) bool {
	granted := map[string]bool{}
	for i, m := range a.Memberships {
		if m.Area == area && m.Covers(procedureType) {
			for _, p := range a.granted[i] {
				granted[p] = true
			}
		}
	}
	return len(permissions) > 0 && !slices.ContainsFunc(permissions, func(p string) bool { return !granted[p] })
}

// IsStaff tells whether the person has an active membership. Staff can also
// start procedures, and are the applicants of those (ADR-020).
func (a Access) IsStaff() bool {
	return len(a.Memberships) > 0
}
