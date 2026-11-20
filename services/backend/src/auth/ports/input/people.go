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

// Directory lets the auth module show who each person is, and an
// administrator find someone to grant a membership or the administration.
type Directory interface {
	// Find returns the user registered with email, for an administrator. It
	// fails with domain.ErrUserNotFound if nobody signed in with it.
	Find(ctx context.Context, administratorID, email string) (domain.User, error)
	// Users returns each id that is a registered user. An id that is not one
	// is left out.
	Users(ctx context.Context, ids []string) (map[string]domain.User, error)
}
