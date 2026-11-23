package domain

import (
	"errors"
	"time"
)

var (
	ErrLoginStateMismatch = errors.New("login callback does not match the login this browser started")
	ErrLoginRejected      = errors.New("identity provider rejected the login")
	ErrDomainNotAllowed   = errors.New("email domain not allowed for this login")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
)

// PendingLogin is what a login started with the identity provider needs to
// finish: it is created when the browser leaves for the provider and checked
// when it comes back to the callback.
type PendingLogin struct {
	// State ties the callback to the browser that started the login.
	State string
	// Nonce ties the ID token to this login, so an old token cannot be
	// replayed.
	Nonce string
	// CodeVerifier is the PKCE secret: the provider only gets its hash up
	// front, so a leaked code is useless without it.
	CodeVerifier string
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
