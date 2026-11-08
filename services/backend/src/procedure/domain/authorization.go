package domain

import (
	"errors"
	"slices"
	"time"
)

var (
	ErrAccessDenied      = errors.New("access denied")
	ErrProcedureNotFound = errors.New("procedure not found")
)

func (p Profile) Covers(step Profile) bool {
	switch p {
	case ProfileStaffApprover:
		return step == ProfileStaffApprover || step == ProfileStaffOperative
	case ProfileStaffOperative:
		return step == ProfileStaffOperative
	}
	return false
}

// StaffMembership is what procedure needs of an active membership, which auth
// keeps (ADR-017). An empty Scope covers every procedure type.
type StaffMembership struct {
	Area    string
	Profile Profile
	Scope   []int64
}

func (m StaffMembership) covers(procedureType int64) bool {
	return len(m.Scope) == 0 || slices.Contains(m.Scope, procedureType)
}

// Actor is a person at the moment they act on procedures, with their active
// memberships. Without any, they are an applicant.
type Actor struct {
	UserID      string
	Memberships []StaffMembership
}

func (a Actor) IsStaff() bool {
	return len(a.Memberships) > 0
}

func (a Actor) CanSee(p Procedure, c Content) bool {
	if !a.IsStaff() {
		return p.Applicant == a.UserID
	}
	for _, s := range c.Definition.Steps {
		if slices.ContainsFunc(a.Memberships, func(m StaffMembership) bool { return m.Area == s.Area && m.covers(p.ProcedureType) }) {
			return true
		}
	}
	return false
}

// CanActOn gives the steps of the applicant area to the applicant of the
// procedure, and those of any other area, including the legacy external kind,
// to its members.
func (a Actor) CanActOn(p Procedure, step Step, area Area) bool {
	if area.Kind == AreaApplicant {
		return !a.IsStaff() && p.Applicant == a.UserID
	}
	return slices.ContainsFunc(a.Memberships, func(m StaffMembership) bool {
		return m.Area == step.Area && m.covers(p.ProcedureType) && m.Profile.Covers(step.Profile)
	})
}

type DeniedAttempt struct {
	UserID    string
	Action    string
	Procedure string
	At        time.Time
}
