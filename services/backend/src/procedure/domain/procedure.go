package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

var (
	ErrAccessDenied      = errors.New("access denied")
	ErrProcedureNotFound = errors.New("procedure not found")
)

// A Procedure keeps the digest of its version as well as the number, so that
// a version lost and published again with the same number is noticed.
type Procedure struct {
	ID            string
	ProcedureType int64
	Version       int
	VersionDigest string
	Applicant     string
	// OpenStep is the step the procedure waits for, empty once it is closed.
	OpenStep string
	Referral *Referral
}

type Referral struct {
	From  string
	To    string
	Since time.Time
}

func (p Procedure) Pin() version.Pin {
	return version.Pin{ProcedureTypeID: p.ProcedureType, Number: p.Version, Digest: p.VersionDigest}
}

// State is where a procedure is, derived from its open step: the procedure
// type does not declare states. It is not the result of the procedure.
type State string

const (
	StateWithApplicant State = "with_applicant"
	StateInOffice      State = "in_office"
	StateClosed        State = "closed"
)

func (p Procedure) Closed() bool {
	return p.OpenStep == ""
}

// StateOf fails if the open step is not in def, the definition of the version
// the procedure started with.
func StateOf(p Procedure, def definition.ProcedureTypeDefinition) (State, error) {
	if p.Closed() {
		return StateClosed, nil
	}
	i := slices.IndexFunc(def.Steps, func(s definition.Step) bool { return s.ID == p.OpenStep })
	if i < 0 {
		return "", fmt.Errorf("procedure %s: open step %q is not in version %d", p.ID, p.OpenStep, p.Version)
	}
	if def.Steps[i].Role == definition.RoleApplicant {
		return StateWithApplicant, nil
	}
	return StateInOffice, nil
}

func (p Procedure) DaysWithoutReturn(now time.Time) (int, bool) {
	if p.Referral == nil {
		return 0, false
	}
	since := p.Referral.Since
	from := time.Date(since.Year(), since.Month(), since.Day(), 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return int(to.Sub(from).Hours() / 24), true
}
