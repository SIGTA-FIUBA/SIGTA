package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
	out "backend/src/auth/ports/output"
)

var _ input.Authenticator = (*AuthService)(nil)

type AuthConfig struct {
	IdleTimeout   time.Duration
	AllowedDomain string
}

type AuthService struct {
	config   AuthConfig
	identity out.IdentityProvider
	users    out.UserRepository
	sessions out.SessionStore
	now      func() time.Time
}

func NewAuthService(
	config AuthConfig,
	identity out.IdentityProvider,
	users out.UserRepository,
	sessions out.SessionStore,
	now func() time.Time,
) *AuthService {
	return &AuthService{
		config:   config,
		identity: identity,
		users:    users,
		sessions: sessions,
		now:      now,
	}
}

func (s *AuthService) BeginLogin(loginHint string) (string, domain.PendingLogin, error) {
	pending := domain.PendingLogin{State: newToken(), Nonce: newToken(), CodeVerifier: newToken()}
	return s.identity.AuthorizationURL(pending, loginHint), pending, nil
}

// CompleteLogin registers someone new to SIGTA as an applicant: only a
// membership makes them staff (ADR-017).
func (s *AuthService) CompleteLogin(ctx context.Context, code, state string, pending domain.PendingLogin) (string, error) {
	if pending.State == "" || subtle.ConstantTimeCompare([]byte(state), []byte(pending.State)) != 1 {
		return "", domain.ErrLoginStateMismatch
	}
	identity, err := s.identity.Exchange(ctx, code, pending)
	if err != nil {
		return "", fmt.Errorf("%w: %w", domain.ErrLoginRejected, err)
	}
	email := domain.NormalizeEmail(identity.Email)
	if !domain.InDomain(email, s.config.AllowedDomain) {
		return "", domain.ErrDomainNotAllowed
	}
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrUnknownUser) {
		user, err = s.register(ctx, email, identity.Name)
	}
	if err != nil {
		return "", fmt.Errorf("find user %s: %w", email, err)
	}
	return s.openSession(ctx, user.ID)
}

func (s *AuthService) register(ctx context.Context, email, name string) (domain.User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = email
	}
	user, err := s.users.Register(ctx, domain.User{ID: newToken(), Email: email, Name: name})
	if err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}
	return user, nil
}

// Authenticate extends the session, which expires only after IdleTimeout
// without requests.
func (s *AuthService) Authenticate(ctx context.Context, sessionToken string) (domain.User, error) {
	if sessionToken == "" {
		return domain.User{}, domain.ErrSessionNotFound
	}
	tokenHash := hashToken(sessionToken)
	session, err := s.sessions.Find(ctx, tokenHash)
	if err != nil {
		return domain.User{}, fmt.Errorf("find session: %w", err)
	}
	now := s.now()
	if session.IdleSince(now, s.config.IdleTimeout) {
		if err := s.sessions.Delete(ctx, tokenHash); err != nil {
			return domain.User{}, fmt.Errorf("delete expired session: %w", err)
		}
		return domain.User{}, domain.ErrSessionExpired
	}
	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		return domain.User{}, fmt.Errorf("find user %s: %w", session.UserID, err)
	}
	if err := s.sessions.Touch(ctx, tokenHash, now); err != nil {
		return domain.User{}, fmt.Errorf("touch session: %w", err)
	}
	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return nil
	}
	if err := s.sessions.Delete(ctx, hashToken(sessionToken)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *AuthService) openSession(ctx context.Context, userID string) (string, error) {
	token := newToken()
	session := domain.Session{
		TokenHash:  hashToken(token),
		UserID:     userID,
		LastSeenAt: s.now(),
	}
	if err := s.sessions.Save(ctx, session); err != nil {
		return "", fmt.Errorf("save session: %w", err)
	}
	return token, nil
}

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // Never fails since Go 1.24.
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
