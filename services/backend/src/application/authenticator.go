package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"domain"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	out "output"
	"strings"
	"time"
)

type AuthConfig struct {
	// IdleTimeout closes a session after this long without requests.
	IdleTimeout time.Duration
	// LoginLinkTTL is how long an emailed login link stays valid.
	LoginLinkTTL time.Duration
	// AllowedDomain is the only email domain that can log in, such as
	// "fi.uba.ar", whether through the identity provider or a login link.
	AllowedDomain string
	// LoginLinkURL is the frontend page that redeems a link; the token is
	// appended to it.
	LoginLinkURL string
}

type AuthService struct {
	config   AuthConfig
	identity out.IdentityProvider
	users    out.UserRepository
	sessions out.SessionStore
	links    out.LoginLinkStore
	mailer   out.Mailer
	now      func() time.Time
}

func NewAuthService(
	config AuthConfig,
	identity out.IdentityProvider,
	users out.UserRepository,
	sessions out.SessionStore,
	links out.LoginLinkStore,
	mailer out.Mailer,
	now func() time.Time,
) *AuthService {
	return &AuthService{
		config:   config,
		identity: identity,
		users:    users,
		sessions: sessions,
		links:    links,
		mailer:   mailer,
		now:      now,
	}
}

func (s *AuthService) BeginLogin(state, loginHint string) string {
	return s.identity.AuthorizationURL(state, loginHint)
}

// CompleteLogin opens a session for whoever the provider vouches for, as long
// as their email is in the allowed domain and SIGTA already knows them:
// profiles are managed in SIGTA, not created on first login.
func (s *AuthService) CompleteLogin(ctx context.Context, code string) (string, error) {
	identity, err := s.identity.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("%w: %w", out.ErrLoginRejected, err)
	}
	email := domain.NormalizeEmail(identity.Email)
	if !s.allowed(email) {
		return "", out.ErrDomainNotAllowed
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	return s.openSession(ctx, user.ID)
}

// RequestLoginLink emails a single-use link to an applicant, as an
// alternative to the identity provider. It returns nil for unknown emails so
// the response does not reveal who has an account.
func (s *AuthService) RequestLoginLink(ctx context.Context, email string) error {
	email = domain.NormalizeEmail(email)
	if !s.allowed(email) {
		return out.ErrDomainNotAllowed
	}
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrUnknownUser) {
		return nil
	}
	if err != nil {
		return err
	}
	// Staff log in with their institutional account only.
	if user.Role != domain.RoleApplicant {
		return nil
	}

	token, err := newToken()
	if err != nil {
		return err
	}
	link := out.LoginLink{
		TokenHash: hashToken(token),
		UserID:    user.ID,
		ExpiresAt: s.now().Add(s.config.LoginLinkTTL),
	}
	if err := s.links.Save(ctx, link); err != nil {
		return err
	}
	return s.mailer.SendLoginLink(ctx, user.Email, s.config.LoginLinkURL+token)
}

func (s *AuthService) RedeemLoginLink(ctx context.Context, token string) (string, error) {
	link, err := s.links.Consume(ctx, hashToken(token))
	if err != nil {
		return "", err
	}
	if link.ExpiredAt(s.now()) {
		return "", out.ErrInvalidLoginLink
	}
	return s.openSession(ctx, link.UserID)
}

// Authenticate resolves the profile behind a session token and extends the
// session, which expires only after IdleTimeout without requests.
func (s *AuthService) Authenticate(ctx context.Context, sessionToken string) (domain.User, error) {
	if sessionToken == "" {
		return domain.User{}, out.ErrSessionNotFound
	}
	tokenHash := hashToken(sessionToken)
	session, err := s.sessions.Find(ctx, tokenHash)
	if err != nil {
		return domain.User{}, err
	}
	now := s.now()
	if session.IdleSince(now, s.config.IdleTimeout) {
		if err := s.sessions.Delete(ctx, tokenHash); err != nil {
			return domain.User{}, err
		}
		return domain.User{}, out.ErrSessionExpired
	}
	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		return domain.User{}, err
	}
	if err := s.sessions.Touch(ctx, tokenHash, now); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return nil
	}
	return s.sessions.Delete(ctx, hashToken(sessionToken))
}

func (s *AuthService) allowed(email string) bool {
	return strings.HasSuffix(email, "@"+s.config.AllowedDomain)
}

func (s *AuthService) openSession(ctx context.Context, userID string) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	session := out.Session{
		TokenHash:  hashToken(token),
		UserID:     userID,
		LastSeenAt: s.now(),
	}
	if err := s.sessions.Save(ctx, session); err != nil {
		return "", err
	}
	return token, nil
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
