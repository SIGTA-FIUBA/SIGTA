package domain

import (
	"errors"
	"time"
)

var (
	ErrUnknownUserKind   = errors.New("unknown kind of user")
	ErrAccessDenied      = errors.New("access denied")
	ErrProcedureNotFound = errors.New("procedure not found")
	ErrInvalidPermission = errors.New("invalid staff permission")
)

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

func (p Profile) Covers(step Profile) bool {
	switch p {
	case ProfileStaffApprover:
		return step == ProfileStaffApprover || step == ProfileStaffOperative
	case ProfileStaffOperative:
		return step == ProfileStaffOperative
	}
	return false
}

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

type DeniedAttempt struct {
	UserID    string
	Action    string
	Procedure string
	At        time.Time
}
