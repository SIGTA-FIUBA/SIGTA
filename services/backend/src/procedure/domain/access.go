package domain

import (
	"errors"
	"slices"
	"time"
)

var (
	ErrAccessDenied       = errors.New("access denied")
	ErrProcedureNotFound  = errors.New("procedure not found")
	ErrInvalidMembership  = errors.New("invalid membership")
	ErrMembershipNotFound = errors.New("membership not found")
)

// Role is derived, never stored: a person with an active membership is staff,
// and anyone else is an applicant (ADR-017).
type Role string

const (
	RoleApplicant Role = "applicant"
	RoleStaff     Role = "staff"
)

func (p Profile) Known() bool {
	return p == ProfileStaffOperative || p == ProfileStaffApprover
}

func (p Profile) Covers(step Profile) bool {
	switch p {
	case ProfileStaffApprover:
		return step == ProfileStaffApprover || step == ProfileStaffOperative
	case ProfileStaffOperative:
		return step == ProfileStaffOperative
	}
	return false
}

// Membership makes a person staff of an area with a profile. Scope, when not
// empty, restricts it to some procedure types; it names types and not
// versions. A zero Until means it holds until it is ended.
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

func (m Membership) ActiveAt(at time.Time) bool {
	return !at.Before(m.From) && (m.Until.IsZero() || at.Before(m.Until))
}

func (m Membership) covers(procedureType int64) bool {
	return len(m.Scope) == 0 || slices.Contains(m.Scope, procedureType)
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

type Actor struct {
	UserID        string
	Administrator bool
	Memberships   []Membership
}

func NewActor(userID string, administrator bool, memberships []Membership, at time.Time) Actor {
	active := slices.DeleteFunc(slices.Clone(memberships), func(m Membership) bool { return !m.ActiveAt(at) })
	return Actor{UserID: userID, Administrator: administrator, Memberships: active}
}

func (a Actor) Role() Role {
	if len(a.Memberships) > 0 {
		return RoleStaff
	}
	return RoleApplicant
}

func (a Actor) CanSee(p Procedure, c Content) bool {
	if a.Role() == RoleApplicant {
		return p.Applicant == a.UserID
	}
	for _, s := range c.Definition.Steps {
		if slices.ContainsFunc(a.Memberships, func(m Membership) bool { return m.Area == s.Area && m.covers(p.ProcedureType) }) {
			return true
		}
	}
	return false
}

// CanActOn tells whether the actor may act on a step of the procedure. Steps
// of the applicant area belong to the applicant of the procedure; any other
// area, including the legacy external kind, belongs to its members.
func (a Actor) CanActOn(p Procedure, step Step, area Area) bool {
	if area.Kind == AreaApplicant {
		return a.Role() == RoleApplicant && p.Applicant == a.UserID
	}
	return slices.ContainsFunc(a.Memberships, func(m Membership) bool {
		return m.Area == step.Area && m.covers(p.ProcedureType) && m.Profile.Covers(step.Profile)
	})
}

type DeniedAttempt struct {
	UserID    string
	Action    string
	Procedure string
	At        time.Time
}
