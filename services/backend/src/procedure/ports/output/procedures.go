package output

import (
	"context"

	"backend/src/procedure/domain"
)

type ProcedureRepository interface {
	// Add records the procedure together with the values entered when it was
	// submitted, as one change.
	Add(ctx context.Context, p domain.Procedure, values ...domain.FieldValue) error
	Get(ctx context.Context, id string) (domain.Procedure, error)
	List(ctx context.Context) ([]domain.Procedure, error)
	// Values returns the values of the procedure in the order they were
	// recorded.
	Values(ctx context.Context, procedureID string) ([]domain.FieldValue, error)
	Advance(ctx context.Context, p domain.Procedure, e domain.Event) error
	Events(ctx context.Context, procedureID string) ([]domain.Event, error)
}
