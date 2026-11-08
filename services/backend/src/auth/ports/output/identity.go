package output

import (
	"context"
	"time"

	"backend/src/auth/domain"
)

type Identity struct {
	Subject string
	Email   string
	// Name is the display name, when the provider has one.
	Name string
}

// IdentityProvider delegates the login to an external provider, such as the
// institutional OpenID Connect login. SIGTA never sees a password.
type IdentityProvider interface {
	// AuthorizationURL is where the browser goes to log in. It carries the
	// pending login's state, nonce and PKCE code challenge; the provider sends
	// the browser back to the callback with a code and the same state.
	AuthorizationURL(pending domain.PendingLogin, loginHint string) string
	// Exchange trades the code from the callback for the identity it proves,
	// sending the PKCE code verifier and checking the ID token's nonce
	// against the pending login.
	Exchange(ctx context.Context, code string, pending domain.PendingLogin) (Identity, error)
}

// UserRepository stores users. Emails arrive already normalized.
type UserRepository interface {
	// FindByEmail and FindByID fail with domain.ErrUnknownUser if there is no
	// such user.
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
	// Register adds a user and returns it. If the email is already registered
	// it returns that user instead, so two first logins of the same person at
	// once do not create two users.
	Register(ctx context.Context, user domain.User) (domain.User, error)
}

// SessionStore stores open sessions by the hash of their token.
type SessionStore interface {
	Save(ctx context.Context, session domain.Session) error
	// Find and Touch fail with domain.ErrSessionNotFound if there is no such
	// session.
	Find(ctx context.Context, tokenHash string) (domain.Session, error)
	Touch(ctx context.Context, tokenHash string, seenAt time.Time) error
	// Delete does not fail if the session is already gone.
	Delete(ctx context.Context, tokenHash string) error
}
