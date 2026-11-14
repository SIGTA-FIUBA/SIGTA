package domain

import (
	"errors"
	"time"
)

var (
	ErrDocumentNotFound     = errors.New("document not found")
	ErrDocumentNotRequested = errors.New("the open step does not ask for that document")
	ErrDocumentType         = errors.New("the file type is not accepted for that document")
	ErrDocumentTooLarge     = errors.New("the file exceeds the maximum size of that document")
	ErrEmptyDocument        = errors.New("the file is empty")
)

// Document is a file attached to a procedure for one of the documents its
// procedure type asks for. The file itself is in the document store, under
// Digest; Document is only what is recorded about it.
type Document struct {
	ID        string
	Procedure string
	// Requirement is the document of the procedure type that the file covers.
	Requirement string
	Step        string
	// FileName is what the browser sent, cleaned of paths and invisible
	// characters but still chosen by whoever uploaded the file.
	FileName string
	// MediaType is read from the content, never from the file name or what
	// the browser declares.
	MediaType  string
	Size       int64
	Digest     string
	UploadedBy string
	UploadedAt time.Time
}

// Current keeps the last document attached for each requirement, in the
// order they were attached. An attached document is never removed: a new
// file for the same requirement replaces it as the current one.
func Current(attached []Document) []Document {
	last := map[string]int{}
	for i, d := range attached {
		last[d.Requirement] = i
	}
	var current []Document
	for i, d := range attached {
		if last[d.Requirement] == i {
			current = append(current, d)
		}
	}
	return current
}
