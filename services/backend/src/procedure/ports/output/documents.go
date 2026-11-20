package output

import (
	"context"

	"backend/src/procedure/domain"
)

// DocumentRepository reads the documents that ProcedureRepository records
// when a procedure starts or a document is attached. A recorded document is
// never changed or removed.
type DocumentRepository interface {
	// List returns the documents of a procedure in the order they were added.
	List(ctx context.Context, procedureID string) ([]domain.Document, error)
	// Get fails with domain.ErrDocumentNotFound if the procedure has no
	// document with that identifier.
	Get(ctx context.Context, procedureID, documentID string) (domain.Document, error)
	// Digests returns the digests of the files that the documents of every
	// procedure name, each once and in no order.
	Digests(ctx context.Context) ([]string, error)
}
