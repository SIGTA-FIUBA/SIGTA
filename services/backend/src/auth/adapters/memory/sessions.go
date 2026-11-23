package memory

import (
	"context"
	"sync"
	"time"

	"backend/src/auth/domain"
)

type SessionRepository struct {
	mu       sync.Mutex
	sessions map[string]domain.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{sessions: make(map[string]domain.Session)}
}

func (s *SessionRepository) Save(_ context.Context, session domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.TokenHash] = session
	return nil
}

func (s *SessionRepository) Get(_ context.Context, tokenHash string) (domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[tokenHash]
	if !ok {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return session, nil
}

func (s *SessionRepository) Touch(_ context.Context, tokenHash string, seenAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[tokenHash]
	if !ok {
		return domain.ErrSessionNotFound
	}
	session.LastSeenAt = seenAt
	s.sessions[tokenHash] = session
	return nil
}

func (s *SessionRepository) Delete(_ context.Context, tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, tokenHash)
	return nil
}
