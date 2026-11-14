package input

import (
	"context"
	"errors"

	"backend/src/auth/domain"
)

type Authenticator interface {
	// BeginLogin returns where to send the browser and the pending login that only
	// that browser should hold and bring back to CompleteLogin.
	BeginLogin(loginHint string) (url string, pending domain.PendingLogin, err error)
	CompleteLogin(ctx context.Context, code, state string, pending domain.PendingLogin) (sessionToken string, err error)
	Authenticate(ctx context.Context, sessionToken string) (domain.User, error)
	Logout(ctx context.Context, sessionToken string) error
}

// Unauthenticated tells whether Authenticate failed because there is no valid
// session, so that other modules answer 401 without reading auth/domain
// (ADR-021).
func Unauthenticated(err error) bool {
	return errors.Is(err, domain.ErrSessionNotFound) ||
		errors.Is(err, domain.ErrSessionExpired) ||
		errors.Is(err, domain.ErrUnknownUser)
}
