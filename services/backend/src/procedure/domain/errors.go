package domain

import (
	"errors"
	"strings"
)

var (
	ErrInvalidDefinition = errors.New("invalid procedure type definition")
	ErrInvalidAreas      = errors.New("invalid areas")
	ErrInvalidAuthorship = errors.New("invalid authorship")
	ErrAlteredContent    = errors.New("content does not match its digest")
	ErrUnknownFormat     = errors.New("unknown content format")
)

// InvalidError is what Validate returns: the cause of the failure, which
// errors.Is matches, and every problem found.
type InvalidError struct {
	Cause    error
	Problems []Problem
}

func (e *InvalidError) Error() string {
	texts := make([]string, len(e.Problems))
	for i, p := range e.Problems {
		texts[i] = p.String()
	}
	return e.Cause.Error() + ": " + strings.Join(texts, "\n")
}

func (e *InvalidError) Unwrap() error { return e.Cause }

func ProblemsOf(err error) []Problem {
	if invalid, ok := errors.AsType[*InvalidError](err); ok {
		return invalid.Problems
	}
	return nil
}
