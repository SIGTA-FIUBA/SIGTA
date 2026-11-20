package output

import (
	"context"
	"errors"

	"backend/src/procedure/domain/draft"
)

var (
	ErrDraftNotFound = errors.New("procedure type draft not found")
	ErrDraftExists   = errors.New("the procedure type already has a draft")
	// ErrDraftChanged means the draft changed since it was read.
	ErrDraftChanged = errors.New("the draft changed since it was read")
)

// ProcedureTypeDraftRepository keeps the only open draft of each procedure type.
// Unlike a version, a draft changes and is deleted when it is published or
// discarded.
type ProcedureTypeDraftRepository interface {
	// Get fails with ErrDraftNotFound if the procedure type has no draft.
	Get(ctx context.Context, procedureType int64) (draft.Draft, error)
	// List returns every draft, ordered by procedure type.
	List(ctx context.Context) ([]draft.Draft, error)
	// Add fails with ErrDraftExists if the procedure type already has a
	// draft, also when another one is added at the same time.
	Add(ctx context.Context, d draft.Draft) error
	// Update replaces the draft of d.ProcedureTypeID only if the stored one is
	// at the revision before d.Revision, and fails with ErrDraftChanged if not,
	// or with ErrDraftNotFound if there is none.
	Update(ctx context.Context, d draft.Draft) error
	// Remove deletes the draft only if it is still at revision, and fails like
	// Update.
	Remove(ctx context.Context, procedureType int64, revision int) error
}
