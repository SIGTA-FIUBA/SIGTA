package output

import (
	"context"
	"errors"

	"backend/src/procedure/domain"
)

var (
	ErrProcedureExists = errors.New("procedure already exists")
	ErrDocumentExists  = errors.New("document already exists")
)

// Start is everything a procedure records when it starts: the events of the
// steps done to start it, the values entered and the documents attached.
type Start struct {
	Procedure domain.Procedure
	Events    []domain.Event
	Values    []domain.FieldValue
	Documents []domain.Document
}

// ProcedureRepository keeps the procedures with their history, values and
// documents. Each change is recorded whole or not at all.
type ProcedureRepository interface {
	// Add records the start of a procedure, whose version is already stored
	// and whose documents name it. It fails with ErrProcedureExists if there
	// is a procedure with that identifier, and with ErrDocumentExists if a
	// document identifier is already taken.
	Add(ctx context.Context, s Start) error
	Get(ctx context.Context, id string) (domain.Procedure, error)
	// List returns the procedures ordered by identifier.
	List(ctx context.Context) ([]domain.Procedure, error)
	// Values returns the values of the procedure in the order they were
	// recorded.
	Values(ctx context.Context, procedureID string) ([]domain.FieldValue, error)
	// Advance moves the procedure as p says and appends e, the confirmation
	// of its open step. It fails with engine.ErrStepNotOpen if e.Step is no
	// longer the open step, and with engine.ErrAwaitingApplicant if the
	// procedure has a pending correction request.
	Advance(ctx context.Context, p domain.Procedure, e domain.Event) error
	// Attach records a document together with e, its event in the history,
	// which does not move the procedure. It fails with
	// domain.ErrProcedureNotFound if there is no such procedure, with
	// engine.ErrStepNotOpen if d.Step is no longer the open step, and with
	// ErrDocumentExists if the document identifier is already taken.
	Attach(ctx context.Context, d domain.Document, e domain.Event) error
	// RequestCorrection adds a pending correction request to the procedure of
	// the event, together with the event, as one change. It fails with
	// domain.ErrProcedureNotFound if there is no such procedure, and with
	// engine.ErrStepNotOpen if the procedure is no longer on the step the
	// request was made on.
	RequestCorrection(ctx context.Context, r domain.CorrectionRequest, e domain.Event) error
	Events(ctx context.Context, procedureID string) ([]domain.Event, error)
}
