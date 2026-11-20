package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

var _ output.SessionRepository = (*SessionRepository)(nil)

// SessionRepository uses single statements, so it needs no transaction.
type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (s *SessionRepository) Save(ctx context.Context, session domain.Session) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO sessions (token_hash, user_id, last_seen_at) VALUES ($1, $2, $3)",
		session.TokenHash, session.UserID, session.LastSeenAt)
	if err != nil {
		return fmt.Errorf("save a session of user %s: %w", session.UserID, err)
	}
	return nil
}

func (s *SessionRepository) Get(ctx context.Context, tokenHash string) (domain.Session, error) {
	session := domain.Session{TokenHash: tokenHash}
	err := s.pool.QueryRow(ctx, "SELECT user_id, last_seen_at FROM sessions WHERE token_hash = $1", tokenHash).
		Scan(&session.UserID, &session.LastSeenAt)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.Session{}, domain.ErrSessionNotFound
	case err != nil:
		return domain.Session{}, fmt.Errorf("read a session: %w", err)
	}
	session.LastSeenAt = session.LastSeenAt.UTC()
	return session, nil
}

func (s *SessionRepository) Touch(ctx context.Context, tokenHash string, seenAt time.Time) error {
	tag, err := s.pool.Exec(ctx, "UPDATE sessions SET last_seen_at = $2 WHERE token_hash = $1", tokenHash, seenAt)
	switch {
	case err != nil:
		return fmt.Errorf("touch a session: %w", err)
	case tag.RowsAffected() == 0:
		return domain.ErrSessionNotFound
	}
	return nil
}

func (s *SessionRepository) Delete(ctx context.Context, tokenHash string) error {
	if _, err := s.pool.Exec(ctx, "DELETE FROM sessions WHERE token_hash = $1", tokenHash); err != nil {
		return fmt.Errorf("delete a session: %w", err)
	}
	return nil
}
