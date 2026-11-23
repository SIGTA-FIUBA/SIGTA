package input

import (
	"context"

	"backend/src/auth/domain"
)

// People gives other modules the names of the users, so they show a person
// and not an internal id (ADR-021).
type People interface {
	// Names returns the name of each id that is a registered user. An id
	// that is not one is left out.
	Names(ctx context.Context, ids []string) (map[string]string, error)
}

const (
	PeopleSearchLimit   = 10
	PeopleSearchMinimum = 2
)

// Directory lets the auth module show who each person is, and an
// administrator find someone to grant a membership or the administration.
type Directory interface {
	// Search returns, for an administrator, the users whose name or email
	// contains text, without telling case or accents apart, ordered by name.
	// It returns at most PeopleSearchLimit, and nobody for a text shorter
	// than PeopleSearchMinimum, which almost everyone would match. The limit
	// is per search: an administrator may see anyone who signed in (DU-46).
	Search(ctx context.Context, administratorID, text string) ([]domain.User, error)
	// Users returns each id that is a registered user. An id that is not one
	// is left out.
	Users(ctx context.Context, ids []string) (map[string]domain.User, error)
}
