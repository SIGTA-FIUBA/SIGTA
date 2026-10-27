package ports

import (
	"context"
	"domain"
	out "output"
)

type Authenticator interface {
	// BeginLogin returns where to send the browser and the pending login it
	// must bring back to CompleteLogin.
	BeginLogin(loginHint string) (url string, pending out.PendingLogin, err error)
	CompleteLogin(ctx context.Context, code, state string, pending out.PendingLogin) (sessionToken string, err error)
	RequestLoginLink(ctx context.Context, email string) error
	RedeemLoginLink(ctx context.Context, token string) (sessionToken string, err error)
	Authenticate(ctx context.Context, sessionToken string) (domain.User, error)
	Logout(ctx context.Context, sessionToken string) error
}
