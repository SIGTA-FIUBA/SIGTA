package ports

import (
	"errors"
	"time"
)

var (
	ErrLoginRejected    = errors.New("identity provider rejected the login")
	ErrDomainNotAllowed = errors.New("email domain not allowed for this login")
	ErrInvalidLoginLink = errors.New("login link is invalid, expired or already used")
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionExpired   = errors.New("session expired")
)

// Identity is what an identity provider asserts about whoever just logged in.
type Identity struct {
	Subject string
	Email   string
}

// Session is an open session. Only the hash of its token is stored, so a
// leaked database does not leak usable sessions.
type Session struct {
	TokenHash  string
	UserID     string
	LastSeenAt time.Time
}

func (s Session) IdleSince(now time.Time, timeout time.Duration) bool {
	return now.Sub(s.LastSeenAt) >= timeout
}

// LoginLink is a single-use link sent by email. Only the hash of its token is
// stored.
type LoginLink struct {
	TokenHash string
	UserID    string
	ExpiresAt time.Time
}

func (l LoginLink) ExpiredAt(now time.Time) bool {
	return !now.Before(l.ExpiresAt)
}
