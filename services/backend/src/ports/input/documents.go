package input

import (
	"backend/src/domain"
	"context"
	"io"
)

type Upload struct {
	ProcedureID string
	// RequirementID is empty for a document that covers no requirement.
	RequirementID string
	Filename      string
	Content       io.Reader
}

type Documents interface {
	Upload(ctx context.Context, user domain.User, upload Upload) (domain.Document, error)
	// Open returns the document and its content, which the caller closes.
	Open(ctx context.Context, user domain.User, procedureID, documentID string) (domain.Document, io.ReadCloser, error)
}
