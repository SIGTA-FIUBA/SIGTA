// Package dev has adapters for local development that need no external
// services. They must never be wired in production.
package dev

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"

	"backend/src/auth/domain"
	out "backend/src/auth/ports/output"
)

// IdentityProvider logs in as whatever email the login hint names, or as
// defaultEmail without a hint, without asking for anything else. It stands in
// for the institutional OpenID Connect login during development, and checks
// PKCE like a real provider would: the code it issues only works with the
// verifier of the login that asked for it.
type IdentityProvider struct {
	callbackURL  string
	defaultEmail string
}

func NewIdentityProvider(callbackURL, defaultEmail string) *IdentityProvider {
	return &IdentityProvider{callbackURL: callbackURL, defaultEmail: defaultEmail}
}

func (p *IdentityProvider) AuthorizationURL(pending domain.PendingLogin, loginHint string) string {
	if loginHint == "" {
		loginHint = p.defaultEmail
	}
	// The code carries the challenge, as a real provider would remember it.
	code := pkceS256Challenge(pending.CodeVerifier) + "." + loginHint
	query := url.Values{"state": {pending.State}, "code": {code}}
	return p.callbackURL + "?" + query.Encode()
}

func (p *IdentityProvider) Exchange(_ context.Context, code string, pending domain.PendingLogin) (out.Identity, error) {
	challenge, email, ok := strings.Cut(code, ".")
	if !ok || email == "" {
		return out.Identity{}, errors.New("dev identity provider needs an email as login_hint")
	}
	if subtle.ConstantTimeCompare([]byte(challenge), []byte(pkceS256Challenge(pending.CodeVerifier))) != 1 {
		return out.Identity{}, errors.New("PKCE code verifier does not match the challenge")
	}
	return out.Identity{Subject: "dev|" + email, Email: email}, nil
}

func pkceS256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
