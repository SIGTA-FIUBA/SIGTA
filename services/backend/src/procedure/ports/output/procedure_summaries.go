package output

import (
	"context"

	"backend/src/procedure/domain/inbox"
)

// ProcedureSummaries reads a summary of each procedure in progress, for the
// inbox. Closed procedures are not included. It returns only what the access
// can see: a store with a query language applies it in the query, so that
// nothing else leaves the store.
type ProcedureSummaries interface {
	InProgress(ctx context.Context, access inbox.Access) ([]inbox.Summary, error)
}

// Permissions tells what a person can see of the procedures. Until per-type
// permissions exist it derives the access from the person's role.
type Permissions interface {
	AccessOf(ctx context.Context, person inbox.Person) (inbox.Access, error)
}
