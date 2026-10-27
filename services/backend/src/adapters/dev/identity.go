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
	out "output"
	"strings"
)

// IdentityProvider logs in as whatever email the login hint names, without
// asking for anything else. It stands in for the institutional OpenID Connect
// login during development, and checks PKCE like a real provider would: the
// code it issues only works with the verifier of the login that asked for it.
type IdentityProvider struct {
	callbackURL string
}

func NewIdentityProvider(callbackURL string) *IdentityProvider {
	return &IdentityProvider{callbackURL: callbackURL}
}

func (p *IdentityProvider) AuthorizationURL(pending out.PendingLogin, loginHint string) string {
	// The code carries the challenge, as a real provider would remember it.
	code := codeChallenge(pending.CodeVerifier) + "." + loginHint
	query := url.Values{"state": {pending.State}, "code": {code}}
	return p.callbackURL + "?" + query.Encode()
}

func (p *IdentityProvider) Exchange(_ context.Context, code string, pending out.PendingLogin) (out.Identity, error) {
	challenge, email, ok := strings.Cut(code, ".")
	if !ok || email == "" {
		return out.Identity{}, errors.New("dev identity provider needs an email as login_hint")
	}
	if subtle.ConstantTimeCompare([]byte(challenge), []byte(codeChallenge(pending.CodeVerifier))) != 1 {
		return out.Identity{}, errors.New("PKCE code verifier does not match the challenge")
	}
	return out.Identity{Subject: "dev|" + email, Email: email}, nil
}

// codeChallenge is the PKCE S256 challenge for a verifier (RFC 7636).
func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
