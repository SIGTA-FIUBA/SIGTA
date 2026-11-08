package input

import (
	"context"

	"backend/src/procedure/domain"
)

type ProcedureView struct {
	Procedure domain.Procedure
	Actions   []string
}

type Procedures interface {
	Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error)
	Open(ctx context.Context, userID, procedureID string) (ProcedureView, error)
	List(ctx context.Context, userID string) ([]domain.Procedure, error)
	Act(ctx context.Context, userID, procedureID, stepID string) error
}
