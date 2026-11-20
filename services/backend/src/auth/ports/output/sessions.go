package output

import (
	"context"
	"time"

	"backend/src/auth/domain"
)

// SessionRepository stores open sessions by the hash of their token.
type SessionRepository interface {
	Save(ctx context.Context, session domain.Session) error
	// Get and Touch fail with domain.ErrSessionNotFound if there is no such
	// session.
	Get(ctx context.Context, tokenHash string) (domain.Session, error)
	Touch(ctx context.Context, tokenHash string, seenAt time.Time) error
	// Delete does not fail if the session is already gone.
	Delete(ctx context.Context, tokenHash string) error
}
