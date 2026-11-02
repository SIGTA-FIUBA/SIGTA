package memory

import (
	"context"
	"sync"
	"time"

	"backend/src/auth/domain"
)

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]domain.Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]domain.Session)}
}

func (s *SessionStore) Save(_ context.Context, session domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.TokenHash] = session
	return nil
}

func (s *SessionStore) Find(_ context.Context, tokenHash string) (domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[tokenHash]
	if !ok {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return session, nil
}

func (s *SessionStore) Touch(_ context.Context, tokenHash string, seenAt time.Time) error {
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

func (s *SessionStore) Delete(_ context.Context, tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, tokenHash)
	return nil
}

type LoginLinkStore struct {
	mu    sync.Mutex
	links map[string]domain.LoginLink
}

func NewLoginLinkStore() *LoginLinkStore {
	return &LoginLinkStore{links: make(map[string]domain.LoginLink)}
}

func (s *LoginLinkStore) Save(_ context.Context, link domain.LoginLink) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links[link.TokenHash] = link
	return nil
}

func (s *LoginLinkStore) Consume(_ context.Context, tokenHash string) (domain.LoginLink, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	link, ok := s.links[tokenHash]
	if !ok {
		return domain.LoginLink{}, domain.ErrInvalidLoginLink
	}
	delete(s.links, tokenHash)
	return link, nil
}
