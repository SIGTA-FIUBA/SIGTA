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
	// CorrectionRequests holds what the applicant has been asked for and has
	// not answered yet. While there is one, the procedure waits for them on
	// the same open step.
	CorrectionRequests []CorrectionRequest
	// StartedAt is when the applicant started it, and MovedAt when it reached
	// its open step.
	StartedAt time.Time
	MovedAt   time.Time
}

// CorrectionRequest asks the applicant for something specific without closing
// or restarting the procedure: what is missing, out of the fields and
// documents of its procedure type, and why.
type CorrectionRequest struct {
	Step      string
	By        string
	At        time.Time
	Reason    string
	Fields    []string
	Documents []string
}

type Referral struct {
	From  string
	To    string
	Since time.Time
}

func (p Procedure) Pin() version.Pin {
	return version.Pin{ProcedureTypeID: p.ProcedureType, Number: p.Version, Digest: p.VersionDigest}
}

// State is where a procedure is, derived from its open step and its pending
// correction requests: the procedure type does not declare states. It is not the result of the procedure.
type State string

const (
	StateWithApplicant State = "with_applicant"
	StateInOffice      State = "in_office"
	StateReferred      State = "referred"
	StateClosed        State = "closed"
)

func (p Procedure) Closed() bool {
	return p.OpenStep == ""
}

// AwaitingApplicant tells whether the applicant has a correction to make.
func (p Procedure) AwaitingApplicant() bool {
	return len(p.CorrectionRequests) > 0
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
	if def.Steps[i].Role == definition.RoleApplicant || p.AwaitingApplicant() {
		return StateWithApplicant, nil
	}
	if p.Referral != nil {
		return StateReferred, nil
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
