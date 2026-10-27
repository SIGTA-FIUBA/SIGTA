package ports

import (
	"context"
	"domain"
	"time"
)

// IdentityProvider delegates the login to an external provider, such as the
// institutional OpenID Connect login. SIGTA never sees a password.
type IdentityProvider interface {
	// AuthorizationURL is where the browser goes to log in. The provider sends
	// it back to the callback with a code and the same state.
	AuthorizationURL(state, loginHint string) string
	// Exchange trades the code from the callback for the identity it proves.
	Exchange(ctx context.Context, code string) (Identity, error)
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
}

type SessionStore interface {
	Save(ctx context.Context, session Session) error
	Find(ctx context.Context, tokenHash string) (Session, error)
	Touch(ctx context.Context, tokenHash string, seenAt time.Time) error
	Delete(ctx context.Context, tokenHash string) error
}

type LoginLinkStore interface {
	Save(ctx context.Context, link LoginLink) error
	// Consume deletes the link and returns it, so it can be used only once even
	// under concurrent redemptions.
	Consume(ctx context.Context, tokenHash string) (LoginLink, error)
}

type Mailer interface {
	SendLoginLink(ctx context.Context, to, link string) error
}
