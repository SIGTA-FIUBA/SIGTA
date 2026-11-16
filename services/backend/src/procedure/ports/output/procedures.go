package output

import (
	"context"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/engine"
)

type ProcedureRepository interface {
	Add(ctx context.Context, p domain.Procedure) error
	Get(ctx context.Context, id string) (domain.Procedure, error)
	List(ctx context.Context) ([]domain.Procedure, error)
	Advance(ctx context.Context, p domain.Procedure, e engine.Event) error
	Events(ctx context.Context, procedureID string) ([]engine.Event, error)
}
