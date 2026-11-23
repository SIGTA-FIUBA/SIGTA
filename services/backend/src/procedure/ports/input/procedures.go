package input

import (
	"context"
	"slices"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/domain/submission"
)

// ProcedureView is a procedure with the steps the person can act on now.
type ProcedureView struct {
	Procedure domain.Procedure
	Actions   []string
}

// CanAct says whether the person can act on the open step.
func (v ProcedureView) CanAct() bool {
	return slices.Contains(v.Actions, v.Procedure.OpenStep)
}

// ProcedureDetail is a procedure with the name and the steps of the version it
// started with, the value and the file that cover each field and document its
// steps ask for, whether the person can act on the open step, the corrections
// the applicant has pending and the steps done so far.
type ProcedureDetail struct {
	Procedure domain.Procedure
	Name      string
	State     domain.State
	Steps     []detail.Step
	CanAct    bool
	// AsApplicant is set when the person is the applicant, who sees only
	// part of the procedure (detail.Applicant).
	AsApplicant        bool
	Referral           *detail.Referral
	CorrectionRequests []domain.CorrectionRequest
	// History holds the steps done, in the order they were recorded.
	History []domain.Event
}

// ProcedureDetails answers domain.ErrProcedureNotFound to whoever cannot see
// the procedure, as Procedures does.
type ProcedureDetails interface {
	Detail(ctx context.Context, userID, procedureID string) (ProcedureDetail, error)
	// Mine returns the procedures that the person started, in progress and
	// closed, the most recently moved first.
	Mine(ctx context.Context, userID string) ([]MyProcedure, error)
	// Receipt is for the applicant alone: to staff of the areas the
	// procedure passes through it is not found either.
	Receipt(ctx context.Context, userID, procedureID string) (Receipt, error)
}

// Receipt is the proof the applicant takes away of their procedure: its
// detail as they see it and when it was issued.
type Receipt struct {
	ProcedureDetail
	IssuedAt time.Time
}

// MyProcedure is a procedure as its applicant sees it in their list: where it
// is, not what happened inside each area, and once it closes, the documents
// that resolve it.
type MyProcedure struct {
	Procedure   domain.Procedure
	Name        string
	StepName    string // empty once the procedure is closed
	State       domain.State
	Resolutions []detail.ResolutionFile
}

type StepConfirmation struct {
	Step               string
	Observation        string
	Outcome            string
	ExternalFileNumber string
}

// CorrectionRequest is what staff asks the applicant for on the open step:
// fields and documents of the procedure type, and why.
type CorrectionRequest struct {
	Step      string
	Reason    string
	Fields    []string
	Documents []string
}

type Procedures interface {
	Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error)
	Open(ctx context.Context, userID, procedureID string) (ProcedureView, error)
	List(ctx context.Context, userID string) ([]domain.Procedure, error)
	Act(ctx context.Context, userID, procedureID, stepID string) error
	Confirm(ctx context.Context, userID, procedureID string, c StepConfirmation) (domain.Event, error)
	// RequestCorrection is up to whoever can act on the open step, as a
	// confirmation is.
	RequestCorrection(ctx context.Context, userID, procedureID string, r CorrectionRequest) (domain.Event, error)
	// EnterValue and AnswerCorrection are up to the applicant, while a
	// correction is pending: they load what it asks for, and then send it.
	EnterValue(ctx context.Context, userID, procedureID, fieldID string, v submission.Value) (domain.FieldValue, error)
	AnswerCorrection(ctx context.Context, userID, procedureID string) (domain.Event, error)
}
