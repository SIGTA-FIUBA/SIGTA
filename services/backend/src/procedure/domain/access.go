package domain

import (
	"errors"
	"time"
)

var (
	ErrUnknownUserKind = errors.New("unknown kind of user")
	ErrAccessDenied    = errors.New("access denied")
	// ErrProcedureNotFound is also what someone gets for a procedure they
	// cannot see, so that the answer does not reveal whether it exists.
	ErrProcedureNotFound = errors.New("procedure not found")
	ErrInvalidPermission = errors.New("invalid staff permission")
)

// UserKind tells what a person is in SIGTA. The kinds exclude each other:
// only an applicant starts procedures, and only staff act on them.
type UserKind string

const (
	KindApplicant     UserKind = "applicant"
	KindStaff         UserKind = "staff"
	KindAdministrator UserKind = "administrator"
)

func (k UserKind) Known() bool {
	switch k {
	case KindApplicant, KindStaff, KindAdministrator:
		return true
	}
	return false
}

// Covers: the approval profile also does what the operative does.
func (p Profile) Covers(step Profile) bool {
	switch p {
	case ProfileStaffApprover:
		return step == ProfileStaffApprover || step == ProfileStaffOperative
	case ProfileStaffOperative:
		return step == ProfileStaffOperative
	}
	return false
}

// Permission gives a staff member a profile on a procedure type. It refers to
// the procedure type and not to a version, so publishing a new version keeps
// it.
type Permission struct {
	UserID        string
	ProcedureType int64
	Profile       Profile
}

type Actor struct {
	UserID      string
	Kind        UserKind
	Permissions []Permission
}

func (a Actor) profileOn(procedureType int64) (Profile, bool) {
	for _, p := range a.Permissions {
		if p.ProcedureType == procedureType {
			return p.Profile, true
		}
	}
	return "", false
}

// CanSee lets an applicant see their own procedures and staff see those of
// the procedure types they have a permission on.
func (a Actor) CanSee(p Procedure) bool {
	switch a.Kind {
	case KindApplicant:
		return p.Applicant == a.UserID
	case KindStaff:
		_, ok := a.profileOn(p.ProcedureType)
		return ok
	}
	return false
}

// CanActOn tells whether the actor can act on a step of the procedure. Steps
// of the applicant area have no profile and belong to the applicant of the
// procedure; steps of an external area are done outside SIGTA.
func (a Actor) CanActOn(p Procedure, step Step, area Area) bool {
	switch area.Kind {
	case AreaApplicant:
		return a.Kind == KindApplicant && p.Applicant == a.UserID
	case AreaOffice:
		if a.Kind != KindStaff {
			return false
		}
		profile, ok := a.profileOn(p.ProcedureType)
		return ok && profile.Covers(step.Profile)
	}
	return false
}

// DeniedAttempt records a request that was refused for lack of permission.
// Procedure is empty when the attempt was not on an existing procedure.
type DeniedAttempt struct {
	UserID    string
	Action    string
	Procedure string
	At        time.Time
}
