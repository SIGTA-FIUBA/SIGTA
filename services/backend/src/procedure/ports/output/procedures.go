package output

import (
	"context"

	"backend/src/procedure/domain"
)

type ProcedureRepository interface {
	Add(ctx context.Context, p domain.Procedure) error
	Get(ctx context.Context, id string) (domain.Procedure, error)
	List(ctx context.Context) ([]domain.Procedure, error)
}
