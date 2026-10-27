package ports

import (
	"context"
	"domain"
)

type Authenticator interface {
	BeginLogin(state, loginHint string) string
	CompleteLogin(ctx context.Context, code string) (sessionToken string, err error)
	RequestLoginLink(ctx context.Context, email string) error
	RedeemLoginLink(ctx context.Context, token string) (sessionToken string, err error)
	Authenticate(ctx context.Context, sessionToken string) (domain.User, error)
	Logout(ctx context.Context, sessionToken string) error
}
