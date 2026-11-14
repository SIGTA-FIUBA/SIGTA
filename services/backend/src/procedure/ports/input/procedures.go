package input

import (
	"context"
	"slices"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/detail"
)

// ProcedureView is a procedure with the name and the steps of the version it
// started with, and the steps the person can act on now.
type ProcedureView struct {
	Procedure domain.Procedure
	Name      string
	Steps     []detail.Step
	Actions   []string
}

// CanAct says whether the person can act on the open step.
func (v ProcedureView) CanAct() bool {
	return slices.Contains(v.Actions, v.Procedure.OpenStep)
}

// ProcedureDetails opens a procedure as Procedures does, with the file that
// covers each document its steps ask for.
type ProcedureDetails interface {
	Detail(ctx context.Context, userID, procedureID string) (ProcedureView, error)
}

type Procedures interface {
	Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error)
	Open(ctx context.Context, userID, procedureID string) (ProcedureView, error)
	List(ctx context.Context, userID string) ([]domain.Procedure, error)
	Act(ctx context.Context, userID, procedureID, stepID string) error
}
