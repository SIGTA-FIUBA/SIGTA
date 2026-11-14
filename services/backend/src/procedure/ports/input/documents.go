package input

import (
	"context"
	"io"

	"backend/src/procedure/domain"
)

type Upload struct {
	ProcedureID string
	// Requirement is the document of the procedure type that the file covers.
	Requirement string
	FileName    string
	Content     io.Reader
}

// Documents answers domain.ErrProcedureNotFound to whoever cannot see the
// procedure, so that the answer does not reveal whether it exists.
type Documents interface {
	// Attach is for whoever can act on the open step, and only for a document
	// that step asks for. It fails with domain.ErrAccessDenied,
	// domain.ErrDocumentNotRequested, domain.ErrDocumentType,
	// domain.ErrDocumentTooLarge or domain.ErrEmptyDocument.
	Attach(ctx context.Context, userID string, u Upload) (domain.Document, error)
	List(ctx context.Context, userID, procedureID string) ([]domain.Document, error)
	// Open fails with domain.ErrDocumentNotFound if the procedure has no
	// document with that identifier. The caller closes the content.
	Open(ctx context.Context, userID, procedureID, documentID string) (domain.Document, io.ReadCloser, error)
}
