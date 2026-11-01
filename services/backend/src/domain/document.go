package domain

import (
	"errors"
	"slices"
	"time"
)

var (
	ErrNoProcedureAccess   = errors.New("no access to that procedure")
	ErrUnknownRequirement  = errors.New("the procedure has no such document requirement")
	ErrRequirementRequired = errors.New("an applicant's document must cover a requirement")
	ErrEmptyFile           = errors.New("the file is empty")
	ErrFileTooLarge        = errors.New("the file is larger than the requirement allows")
	ErrFileTypeNotAccepted = errors.New("the requirement does not accept that file type")
	ErrDocumentNotFound    = errors.New("document not found")
)

// DocumentRule is what a document requirement accepts, as its procedure type
// configures it.
type DocumentRule struct {
	// AcceptedTypes are media types, such as "application/pdf", matched
	// against the type detected from the file's content.
	AcceptedTypes []string
	// MaxSize is in bytes.
	MaxSize int64
}

func (r DocumentRule) Accepts(contentType string) bool {
	return slices.Contains(r.AcceptedTypes, contentType)
}

// Document is a file attached to a procedure. Its content lives in the
// document storage; this record is what the database keeps.
type Document struct {
	ID          string
	ProcedureID string
	// RequirementID is empty when the document covers no requirement, such as
	// a resolution document the Secretaría attaches.
	RequirementID string
	Filename      string
	// ContentType is detected from the content, never taken from the client.
	ContentType string
	Size        int64
	UploadedBy  string
	UploadedAt  time.Time
}

// FileRejectedError carries the rule a file broke, so the person can be told
// what the requirement accepts. It unwraps to ErrFileTooLarge or
// ErrFileTypeNotAccepted.
type FileRejectedError struct {
	Reason error
	Rule   DocumentRule
}

func (e *FileRejectedError) Error() string { return e.Reason.Error() }

func (e *FileRejectedError) Unwrap() error { return e.Reason }
