package domain

import (
	"errors"
	"slices"
	"time"
)

var (
	ErrInvalidMembership  = errors.New("invalid membership")
	ErrMembershipNotFound = errors.New("membership not found")
	ErrAccessDenied       = errors.New("access denied")
)

// Role is derived, never stored: a person with an active membership is staff,
// and anyone else is an applicant (ADR-020).
type Role string

const (
	RoleApplicant Role = "applicant"
	RoleStaff     Role = "staff"
)

// A zero Until means the membership holds until it is ended.
type Membership struct {
	ID          string
	UserID      string
	Area        string
	Permissions []string
	Scope       []int64
	From        time.Time
	Until       time.Time
	GrantedBy   string
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
// memberships count. Being an administrator gives no access to procedures.
type Access struct {
	UserID        string
	Administrator bool
	Memberships   []Membership
}

func NewAccess(userID string, administrator bool, memberships []Membership, at time.Time) Access {
	active := slices.DeleteFunc(slices.Clone(memberships), func(m Membership) bool { return !m.ActiveAt(at) })
	return Access{UserID: userID, Administrator: administrator, Memberships: active}
}

// Holds tells whether the memberships in the area that cover the procedure
// type give, together, every permission asked for.
func (a Access) Holds(area string, procedureType int64, permissions []string) bool {
	granted := map[string]bool{}
	for _, m := range a.Memberships {
		if m.Area == area && m.Covers(procedureType) {
			for _, p := range m.Permissions {
				granted[p] = true
			}
		}
	}
	return len(permissions) > 0 && !slices.ContainsFunc(permissions, func(p string) bool { return !granted[p] })
}

func (a Access) Role() Role {
	if len(a.Memberships) > 0 {
		return RoleStaff
	}
	return RoleApplicant
}
