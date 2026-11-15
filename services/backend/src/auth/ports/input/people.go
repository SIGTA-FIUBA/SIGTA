package input

import "context"

// People gives other modules the names of the users, so they show a person
// and not an internal id (ADR-021).
type People interface {
	// Names returns the name of each id that is a registered user. An id
	// that is not one is left out.
	Names(ctx context.Context, ids []string) (map[string]string, error)
}
