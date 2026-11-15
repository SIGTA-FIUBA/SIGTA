package output

import (
	"context"
	"time"

	"backend/src/procedure/domain/inbox"
)

// ProcedureSummaries reads a summary of each procedure in progress, for the
// inbox. Closed procedures are not included. It returns only what the access
// can see: a store with a query language applies it in the query, so that
// nothing else leaves the store.
type ProcedureSummaries interface {
	InProgress(ctx context.Context, access inbox.Access) ([]inbox.Summary, error)
	// ClosedSince reads the procedures the access can see that were closed at
	// or after since.
	ClosedSince(ctx context.Context, access inbox.Access, since time.Time) ([]inbox.Summary, error)
	// Find reads one procedure, in progress or closed, if the access can see
	// it; ok is false otherwise.
	Find(ctx context.Context, access inbox.Access, id string) (summary inbox.Summary, ok bool, err error)
}

// ProcedureViews keeps when each person last saw each procedure, so that the
// list marks what moved since.
type ProcedureViews interface {
	SeenBy(ctx context.Context, personID string) (inbox.Seen, error)
	See(ctx context.Context, personID, procedureID string, at time.Time) error
}

// Permissions tells what a person can see of the procedures. Until per-type
// permissions exist it derives the access from the person's role.
type Permissions interface {
	AccessOf(ctx context.Context, person inbox.Person) (inbox.Access, error)
}
