package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var (
	// ErrNotSignedIn means the email has no user yet. SIGTA keeps no
	// password, so a person becomes a user by signing in.
	ErrNotSignedIn = errors.New("nobody signed in with that email")
	// ErrHasAdministrators means the first administrator was granted
	// already: the others are granted in SIGTA, which records who did it.
	ErrHasAdministrators = errors.New("there is an administrator already: this command only grants the first one")
)

// GrantFirstAdministrator makes the user with email the first administrator,
// as the initial grant of the system. It takes no author: nothing the one who
// runs it types would prove who did. The server never calls it.
func GrantFirstAdministrator(ctx context.Context, users output.UserRepository, administrators output.FirstAdministratorRepository, email string, newID func() string, at time.Time) (domain.User, error) {
	email = domain.NormalizeEmail(email)
	user, err := users.GetByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, fmt.Errorf("%w: %s has to sign in to SIGTA once first", ErrNotSignedIn, email)
	}
	if err != nil {
		return domain.User{}, err
	}
	granted, err := administrators.GrantFirst(ctx, domain.Administrator{ID: newID(), UserID: user.ID, From: at})
	if err != nil {
		return domain.User{}, err
	}
	if !granted {
		return domain.User{}, ErrHasAdministrators
	}
	return user, nil
}
