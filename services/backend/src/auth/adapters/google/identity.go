// Package google logs in with the Google account of the UBA, through OpenID
// Connect.
package google

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"backend/src/auth/domain"
	out "backend/src/auth/ports/output"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const Issuer = "https://accounts.google.com"

const httpTimeout = 10 * time.Second

type Config struct {
	// Issuer is Google's unless a test points it at a fake provider.
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	HostedDomain string
}

type IdentityProvider struct {
	oauth        oauth2.Config
	verifier     *oidc.IDTokenVerifier
	hostedDomain string
	client       *http.Client
	logger       *slog.Logger
}

// NewIdentityProvider reads Google's discovery document, so it needs the
// network and fails if Google cannot be reached.
func NewIdentityProvider(ctx context.Context, cfg Config, logger *slog.Logger) (*IdentityProvider, error) {
	if cfg.Issuer == "" || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "" || cfg.HostedDomain == "" {
		return nil, errors.New("google identity provider needs issuer, client ID, client secret, redirect URL and hosted domain")
	}
	client := &http.Client{Timeout: httpTimeout}
	provider, err := oidc.NewProvider(oidc.ClientContext(ctx, client), cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("discover %s: %w", cfg.Issuer, err)
	}
	return &IdentityProvider{
		oauth: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     provider.Endpoint(),
			// The profile lives in SIGTA: Google only says who the person is.
			Scopes: []string{oidc.ScopeOpenID, "email", "profile"},
		},
		verifier:     provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		hostedDomain: strings.ToLower(cfg.HostedDomain),
		client:       client,
		logger:       logger,
	}, nil
}

func (p *IdentityProvider) AuthorizationURL(pending domain.PendingLogin, loginHint string) string {
	options := []oauth2.AuthCodeOption{
		oidc.Nonce(pending.Nonce),
		oauth2.S256ChallengeOption(pending.CodeVerifier),
		// hd only narrows the accounts Google offers; Exchange is what checks
		// the domain.
		oauth2.SetAuthURLParam("hd", p.hostedDomain),
	}
	if loginHint != "" {
		options = append(options, oauth2.SetAuthURLParam("login_hint", loginHint))
	}
	return p.oauth.AuthCodeURL(pending.State, options...)
}

type claims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	HostedDomain  string `json:"hd"`
	Name          string `json:"name"`
}

func (p *IdentityProvider) Exchange(ctx context.Context, code string, pending domain.PendingLogin) (out.Identity, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client)
	token, err := p.oauth.Exchange(ctx, code, oauth2.VerifierOption(pending.CodeVerifier))
	if err != nil {
		return out.Identity{}, fmt.Errorf("exchange the code: %w", err)
	}
	raw, _ := token.Extra("id_token").(string)
	if raw == "" {
		return out.Identity{}, errors.New("the token response has no ID token")
	}
	idToken, err := p.verifier.Verify(ctx, raw)
	if err != nil {
		return out.Identity{}, fmt.Errorf("verify the ID token: %w", err)
	}
	if pending.Nonce == "" || subtle.ConstantTimeCompare([]byte(idToken.Nonce), []byte(pending.Nonce)) != 1 {
		return out.Identity{}, errors.New("the ID token nonce does not match the pending login")
	}
	var c claims
	if err := idToken.Claims(&c); err != nil {
		return out.Identity{}, fmt.Errorf("read the ID token claims: %w", err)
	}
	if !c.EmailVerified {
		return out.Identity{}, errors.New("the ID token email is not verified")
	}
	// A Google account can have an email of the domain without belonging to
	// its Workspace; only Workspace accounts carry hd.
	if !strings.EqualFold(c.HostedDomain, p.hostedDomain) {
		p.logger.Warn("google login outside the hosted domain", "hd", c.HostedDomain, "want", p.hostedDomain)
		return out.Identity{}, fmt.Errorf("%w: hosted domain %q", domain.ErrDomainNotAllowed, c.HostedDomain)
	}
	return out.Identity{Subject: idToken.Subject, Email: c.Email, Name: c.Name}, nil
}
