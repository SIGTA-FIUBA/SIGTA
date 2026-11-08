package output

import (
	"context"

	"backend/src/procedure/domain/inbox"
)

// ProcedureSummaries reads a summary of each procedure in progress, for the
// inbox. Closed procedures are not included.
type ProcedureSummaries interface {
	InProgress(ctx context.Context) ([]inbox.Summary, error)
}
