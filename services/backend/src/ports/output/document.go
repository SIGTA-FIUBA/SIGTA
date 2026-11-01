package output

import (
	"backend/src/domain"
	"context"
	"io"
)

// DocumentStore keeps the content of documents outside the database.
type DocumentStore interface {
	// Put stores everything content yields under key and returns how many
	// bytes it wrote.
	Put(ctx context.Context, key string, content io.Reader) (int64, error)
	// Open fails with domain.ErrDocumentNotFound if nothing is stored under
	// key.
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete succeeds if nothing is stored under key.
	Delete(ctx context.Context, key string) error
}

type DocumentRepository interface {
	Save(ctx context.Context, document domain.Document) error
	// FindByID fails with domain.ErrDocumentNotFound.
	FindByID(ctx context.Context, id string) (domain.Document, error)
}

// DocumentRequirements reads the document rules of a procedure's type. The
// procedure model (SIG-12) will implement it.
type DocumentRequirements interface {
	// DocumentRule fails with domain.ErrUnknownRequirement if the procedure's
	// type has no document requirement with that ID.
	DocumentRule(ctx context.Context, procedureID, requirementID string) (domain.DocumentRule, error)
}

// ProcedureAccess decides who can act on a procedure. Authorization (SIG-14)
// will implement it.
type ProcedureAccess interface {
	// CanAccess is false, not an error, for a procedure that does not exist,
	// so a denial does not reveal which procedures exist.
	CanAccess(ctx context.Context, user domain.User, procedureID string) (bool, error)
}
