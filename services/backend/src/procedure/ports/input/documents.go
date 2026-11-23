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

// AttachedDocument is Current when no later document of its procedure covers
// the same requirement.
type AttachedDocument struct {
	domain.Document
	Current bool
}

// Documents answers domain.ErrProcedureNotFound to whoever cannot see the
// procedure, so that the answer does not reveal whether it exists.
type Documents interface {
	// Attach is for whoever can act on the open step, and only for a document
	// that step asks for. It fails with domain.ErrAccessDenied,
	// domain.ErrDocumentNotRequested, domain.ErrDocumentTypeNotAccepted,
	// domain.ErrDocumentTooLarge or domain.ErrDocumentEmpty.
	Attach(ctx context.Context, userID string, u Upload) (domain.Document, error)
	// List returns the documents in the order they were attached, including
	// the ones that a later file replaced.
	List(ctx context.Context, userID, procedureID string) ([]AttachedDocument, error)
	// Open fails with domain.ErrDocumentNotFound if the procedure has no
	// document with that identifier. The caller closes the content.
	Open(ctx context.Context, userID, procedureID, documentID string) (domain.Document, io.ReadCloser, error)
}
