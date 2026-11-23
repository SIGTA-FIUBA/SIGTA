package output

import (
	"context"

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
