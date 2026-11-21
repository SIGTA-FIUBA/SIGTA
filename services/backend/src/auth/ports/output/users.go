package output

import (
	"context"

	"backend/src/auth/domain"
)

// UserRepository stores users. Emails arrive already normalized.
type UserRepository interface {
	// GetByEmail and Get fail with domain.ErrUserNotFound if there is no
	// such user.
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Get(ctx context.Context, id string) (domain.User, error)
	// Search returns up to limit users whose name or email contains text,
	// compared as domain.FoldForSearch folds them, ordered by the folded
	// name and then the email. text arrives folded.
	Search(ctx context.Context, text string, limit int) ([]domain.User, error)
	// Register adds a user and returns it. If the email is already registered
	// it returns that user instead, so two first logins of the same person at
	// once do not create two users. It fails if the id is taken by another
	// email.
	Register(ctx context.Context, user domain.User) (domain.User, error)
}
