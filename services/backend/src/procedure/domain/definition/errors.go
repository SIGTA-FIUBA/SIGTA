package definition

import (
	"errors"
	"strings"
)

var (
	ErrInvalidDefinition = errors.New("invalid procedure type definition")
	ErrInvalidAreas      = errors.New("invalid areas")
)

// ValidationError is what Validate returns: the cause of the failure, which
// errors.Is matches, and every problem found.
type ValidationError struct {
	Cause    error
	Problems []ValidationProblem
}

func (e *ValidationError) Error() string {
	texts := make([]string, len(e.Problems))
	for i, p := range e.Problems {
		texts[i] = p.String()
	}
	return e.Cause.Error() + ": " + strings.Join(texts, "\n")
}

func (e *ValidationError) Unwrap() error { return e.Cause }
