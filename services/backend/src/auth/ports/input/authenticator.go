package input

import (
	"context"

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
