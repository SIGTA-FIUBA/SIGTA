package input

import (
	"context"
	"time"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/draft"
	"backend/src/procedure/domain/version"
)

// DraftView is a draft with what its definition still lacks against the
// current areas, so that whoever edits it sees the problems before proposing
// it.
type DraftView struct {
	Draft    draft.Draft
	Problems []definition.ValidationProblem
	// What the person who asked can do with it, so that a screen offers only
	// that. CanEdit is to edit, propose and discard it; CanReview, to approve
	// or reject it once proposed, which whoever edited or proposed it cannot.
	CanEdit   bool
	CanReview bool
}

// EditableProcedureTypes is what a person can edit or review: the procedure
// types with a published version or a draft where they hold either
// permission, whether they can create a new one, and the areas that a step
// can belong to, with the permissions a step can ask for.
type EditableProcedureTypes struct {
	CanCreate      bool
	ProcedureTypes []EditableProcedureType
	Areas          []definition.Area
}

// EditableProcedureType names a procedure type by its latest version or, if
// it has none, by its draft. Latest is 0 for a procedure type without
// versions. CanOpenDraft tells whether the person can open the draft of a new
// version: they hold the permission to edit, there is a version, and there
// is no draft.
type EditableProcedureType struct {
	ProcedureTypeID int64
	Name            string
	Latest          int
	PublishedAt     time.Time
	CanOpenDraft    bool
	Draft           *DraftView
}

// ProcedureTypeDrafts is how a procedure type is edited and published
// (ADR-013). Proposing a draft and creating, editing or discarding one ask
// for one permission, and approving or rejecting it for another; seeing a
// draft asks for either. Every method fails with domain.ErrAccessDenied
// without them. A change names the revision it was made on and fails with
// output.ErrDraftChanged if the draft changed since then, and with
// output.ErrDraftNotFound if the procedure type has no draft.
type ProcedureTypeDrafts interface {
	// Create starts the draft of a new procedure type, with a number that no
	// other procedure type or draft has.
	Create(ctx context.Context, userID string, d definition.ProcedureTypeDefinition) (DraftView, error)
	// Open starts the draft of a new version of a published procedure type,
	// from its latest version. It fails with output.ErrVersionNotFound if the
	// procedure type has none, and with output.ErrDraftExists if it already
	// has a draft.
	Open(ctx context.Context, userID string, procedureType int64) (DraftView, error)
	Get(ctx context.Context, userID string, procedureType int64) (DraftView, error)
	// List returns the drafts the person can see, ordered by procedure type.
	List(ctx context.Context, userID string) ([]DraftView, error)
	// Editable returns what the person can edit or review, ordered by
	// procedure type. Without either permission it is empty.
	Editable(ctx context.Context, userID string) (EditableProcedureTypes, error)
	// Edit fails with draft.ErrDraftProposed while the draft is proposed.
	Edit(ctx context.Context, userID string, procedureType int64, revision int, d definition.ProcedureTypeDefinition) (DraftView, error)
	// Propose fails with draft.ErrDraftProposed if it is already proposed, and with
	// a *definition.ValidationError if the definition is not valid.
	Propose(ctx context.Context, userID string, procedureType int64, revision int) (DraftView, error)
	// Reject takes the draft back to editing. It fails with
	// draft.ErrDraftNotProposed, draft.ErrSelfApproval or draft.ErrRejectionReasonMissing.
	Reject(ctx context.Context, userID string, procedureType int64, revision int, reason string) (DraftView, error)
	// Approve publishes the draft as a new version and deletes it. It fails
	// with draft.ErrDraftNotProposed, draft.ErrSelfApproval, and as
	// ProcedureTypes.Publish does, for example with an *OutdatedBaseError if
	// another version was published since the draft was opened.
	Approve(ctx context.Context, userID string, procedureType int64, revision int) (version.ProcedureTypeVersion, error)
	// Discard deletes a draft that is not proposed, and fails with
	// draft.ErrDraftProposed if it is.
	Discard(ctx context.Context, userID string, procedureType int64, revision int) error
}
