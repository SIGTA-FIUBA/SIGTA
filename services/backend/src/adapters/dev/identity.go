// Package dev has adapters for local development that need no external
// services. They must never be wired in production.
package dev

import (
	"context"
	"errors"
	"net/url"
	out "output"
)

// IdentityProvider logs in as whatever email the login hint names, without
// asking for anything else. It stands in for the institutional OpenID Connect
// login during development.
type IdentityProvider struct {
	callbackURL string
}

func NewIdentityProvider(callbackURL string) *IdentityProvider {
	return &IdentityProvider{callbackURL: callbackURL}
}

func (p *IdentityProvider) AuthorizationURL(state, loginHint string) string {
	query := url.Values{"state": {state}, "code": {loginHint}}
	return p.callbackURL + "?" + query.Encode()
}

func (p *IdentityProvider) Exchange(_ context.Context, code string) (out.Identity, error) {
	if code == "" {
		return out.Identity{}, errors.New("dev identity provider needs an email as login_hint")
	}
	return out.Identity{Subject: "dev|" + code, Email: code}, nil
}
