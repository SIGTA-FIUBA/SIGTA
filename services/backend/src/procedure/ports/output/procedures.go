package output

import (
	"context"

	"backend/src/procedure/domain"
)

type ProcedureRepository interface {
	// Add records the procedure together with the events of the steps done to
	// start it and the values entered, as one change.
	Add(ctx context.Context, p domain.Procedure, events []domain.Event, values ...domain.FieldValue) error
	Get(ctx context.Context, id string) (domain.Procedure, error)
	List(ctx context.Context) ([]domain.Procedure, error)
	// Values returns the values of the procedure in the order they were
	// recorded.
	Values(ctx context.Context, procedureID string) ([]domain.FieldValue, error)
	Advance(ctx context.Context, p domain.Procedure, e domain.Event) error
	// RequestCorrection adds a pending correction request to the procedure of
	// the event, together with the event, as one change. It fails with
	// engine.ErrStepNotOpen if the procedure is no longer on the step the
	// request was made on.
	RequestCorrection(ctx context.Context, r domain.CorrectionRequest, e domain.Event) error
	// Record appends an event that does not move the procedure, and fails
	// with domain.ErrProcedureNotFound if there is no such procedure.
	Record(ctx context.Context, e domain.Event) error
	Events(ctx context.Context, procedureID string) ([]domain.Event, error)
}
