package input

import (
	"context"
	"slices"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/detail"
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
// started with, the file that covers each document its steps ask for, whether
// the person can act on the open step and the steps done so far.
type ProcedureDetail struct {
	Procedure domain.Procedure
	Name      string
	State     domain.State
	Steps     []detail.Step
	CanAct    bool
	Referral  *detail.Referral
	// History holds the steps done, in the order they were recorded.
	History []domain.Event
}

// ProcedureDetails answers domain.ErrProcedureNotFound to whoever cannot see
// the procedure, as Procedures does.
type ProcedureDetails interface {
	Detail(ctx context.Context, userID, procedureID string) (ProcedureDetail, error)
}

type StepConfirmation struct {
	Step               string
	Observation        string
	Outcome            string
	ExternalFileNumber string
}

type Procedures interface {
	Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error)
	Open(ctx context.Context, userID, procedureID string) (ProcedureView, error)
	List(ctx context.Context, userID string) ([]domain.Procedure, error)
	Act(ctx context.Context, userID, procedureID, stepID string) error
	Confirm(ctx context.Context, userID, procedureID string, c StepConfirmation) (domain.Event, error)
}
