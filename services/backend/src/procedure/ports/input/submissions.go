package input

import (
	"context"
	"io"
	"iter"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/submission"
)

// SubmittedFile is a file for one of the documents the opening steps ask for.
type SubmittedFile struct {
	// Requirement is the document of the procedure type that the file covers.
	Requirement string
	FileName    string
	Content     io.Reader
}

// Submission is what the applicant sends to start a procedure. Files yields
// the files as they arrive: the content of each one is read before the next
// is asked for, so a request can be read without keeping it whole.
type Submission struct {
	ProcedureType int64
	Values        map[string]submission.Value
	Files         iter.Seq2[SubmittedFile, error]
}

type Submissions interface {
	// Submit starts a procedure of the last version of the type, in the name
	// of the person, with what the opening steps ask for, and leaves it on
	// the step after them. Either all of it is recorded or nothing is. It
	// fails with a *submission.Error, which errors.Is matches with
	// submission.ErrInvalid, or with output.ErrNoVersion if the type is not
	// published.
	Submit(ctx context.Context, userID string, s Submission) (domain.Procedure, error)
}
