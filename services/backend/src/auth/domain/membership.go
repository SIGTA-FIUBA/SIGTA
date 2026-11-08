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
// and anyone else is an applicant (ADR-017).
type Role string

const (
	RoleApplicant Role = "applicant"
	RoleStaff     Role = "staff"
)

type Profile string

const (
	ProfileStaffOperative Profile = "staff_operative"
	ProfileStaffApprover  Profile = "staff_approver"
)

func (p Profile) Known() bool {
	return p == ProfileStaffOperative || p == ProfileStaffApprover
}

// Covers tells whether someone with this profile can do what needs the other:
// an approver also does the operative steps.
func (p Profile) Covers(needed Profile) bool {
	switch p {
	case ProfileStaffApprover:
		return needed == ProfileStaffApprover || needed == ProfileStaffOperative
	case ProfileStaffOperative:
		return needed == ProfileStaffOperative
	}
	return false
}

// Membership makes a person staff of an area with a profile. Scope, when not
// empty, restricts it to some procedure types. A zero Until means it holds
// until it is ended.
type Membership struct {
	ID        string
	UserID    string
	Area      string
	Profile   Profile
	Scope     []int64
	From      time.Time
	Until     time.Time
	GrantedBy string
	EndedBy   string
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
// someone can have a higher profile on some procedure types.
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

// Permissions is what a person may do at a given moment: only their active
// memberships count. Being an administrator gives no access to procedures.
type Permissions struct {
	UserID        string
	Administrator bool
	Memberships   []Membership
}

func NewPermissions(userID string, administrator bool, memberships []Membership, at time.Time) Permissions {
	active := slices.DeleteFunc(slices.Clone(memberships), func(m Membership) bool { return !m.ActiveAt(at) })
	return Permissions{UserID: userID, Administrator: administrator, Memberships: active}
}

// Holds tells whether an active membership in the area covers the procedure
// type with at least the profile.
func (p Permissions) Holds(area string, procedureType int64, profile Profile) bool {
	return slices.ContainsFunc(p.Memberships, func(m Membership) bool {
		return m.Area == area && m.Covers(procedureType) && m.Profile.Covers(profile)
	})
}

func (p Permissions) Role() Role {
	if len(p.Memberships) > 0 {
		return RoleStaff
	}
	return RoleApplicant
}

type DeniedAttempt struct {
	UserID string
	Action string
	At     time.Time
}
