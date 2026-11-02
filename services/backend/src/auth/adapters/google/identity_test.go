package google_test

import (
	"backend/src/auth/adapters/google"
	out "backend/src/auth/ports/output"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"golang.org/x/oauth2"
)

const (
	clientID     = "sigta-client"
	clientSecret = "sigta-secret"
	redirectURL  = "https://sigta.test/api/auth/callback"
	hostedDomain = "fi.uba.ar"
	issuedCode   = "issued-code"
)

type fakeGoogle struct {
	t         *testing.T
	server    *httptest.Server
	key       *rsa.PrivateKey
	signer    *rsa.PrivateKey
	challenge string
	claims    map[string]any
	noIDToken bool
}

func newFakeGoogle(t *testing.T) *fakeGoogle {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeGoogle{t: t, key: key, signer: key}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/openid-configuration", f.discovery)
	mux.HandleFunc("GET /keys", f.keys)
	mux.HandleFunc("POST /token", f.token)
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeGoogle) discovery(w http.ResponseWriter, _ *http.Request) {
	f.write(w, http.StatusOK, map[string]any{
		"issuer":                                f.server.URL,
		"authorization_endpoint":                f.server.URL + "/auth",
		"token_endpoint":                        f.server.URL + "/token",
		"jwks_uri":                              f.server.URL + "/keys",
		"id_token_signing_alg_values_supported": []string{"RS256"},
	})
}

func (f *fakeGoogle) keys(w http.ResponseWriter, _ *http.Request) {
	f.write(w, http.StatusOK, jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
		{Key: &f.key.PublicKey, KeyID: "k1", Algorithm: string(jose.RS256), Use: "sig"},
	}})
}

func (f *fakeGoogle) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		f.t.Error(err)
	}
	id, secret, ok := r.BasicAuth()
	if !ok {
		id, secret = r.PostForm.Get("client_id"), r.PostForm.Get("client_secret")
	}
	sum := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
	verifierMatches := base64.RawURLEncoding.EncodeToString(sum[:]) == f.challenge
	if id != clientID || secret != clientSecret || r.PostForm.Get("code") != issuedCode || !verifierMatches {
		f.write(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	response := map[string]any{"access_token": "access", "token_type": "Bearer", "expires_in": 3600}
	if !f.noIDToken {
		response["id_token"] = f.sign()
	}
	f.write(w, http.StatusOK, response)
}

func (f *fakeGoogle) sign() string {
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: f.signer},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "k1"),
	)
	if err != nil {
		f.t.Fatal(err)
	}
	payload, err := json.Marshal(f.claims)
	if err != nil {
		f.t.Fatal(err)
	}
	signed, err := signer.Sign(payload)
	if err != nil {
		f.t.Fatal(err)
	}
	token, err := signed.CompactSerialize()
	if err != nil {
		f.t.Fatal(err)
	}
	return token
}

func (f *fakeGoogle) write(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		f.t.Error(err)
	}
}

func newProvider(t *testing.T, f *fakeGoogle) *google.IdentityProvider {
	t.Helper()
	p, err := google.NewIdentityProvider(t.Context(), google.Config{
		Issuer:       f.server.URL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		HostedDomain: hostedDomain,
	}, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func newPending() out.PendingLogin {
	return out.PendingLogin{State: "state", Nonce: "nonce", CodeVerifier: oauth2.GenerateVerifier()}
}

func authorize(t *testing.T, f *fakeGoogle, p *google.IdentityProvider, pending out.PendingLogin) url.Values {
	t.Helper()
	u, err := url.Parse(p.AuthorizationURL(pending, ""))
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	f.challenge = query.Get("code_challenge")
	now := time.Now()
	f.claims = map[string]any{
		"iss":            f.server.URL,
		"aud":            clientID,
		"sub":            "1234567890",
		"email":          "titular@fi.uba.ar",
		"email_verified": true,
		"hd":             hostedDomain,
		"nonce":          query.Get("nonce"),
		"iat":            now.Unix(),
		"exp":            now.Add(time.Hour).Unix(),
	}
	return query
}

func TestLoginWithAnInstitutionalAccount(t *testing.T) {
	f := newFakeGoogle(t)
	p := newProvider(t, f)
	pending := newPending()
	authorize(t, f, p, pending)

	identity, err := p.Exchange(t.Context(), issuedCode, pending)
	if err != nil {
		t.Fatal(err)
	}
	if identity != (out.Identity{Subject: "1234567890", Email: "titular@fi.uba.ar"}) {
		t.Fatalf("identity = %+v", identity)
	}
}

func TestAuthorizationURL(t *testing.T) {
	f := newFakeGoogle(t)
	p := newProvider(t, f)
	pending := newPending()

	u, err := url.Parse(p.AuthorizationURL(pending, "titular@fi.uba.ar"))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(pending.CodeVerifier))
	want := map[string]string{
		"response_type":         "code",
		"client_id":             clientID,
		"redirect_uri":          redirectURL,
		"scope":                 "openid email profile",
		"state":                 pending.State,
		"nonce":                 pending.Nonce,
		"code_challenge":        base64.RawURLEncoding.EncodeToString(sum[:]),
		"code_challenge_method": "S256",
		"hd":                    hostedDomain,
		"login_hint":            "titular@fi.uba.ar",
	}
	query := u.Query()
	for key, value := range want {
		if got := query.Get(key); got != value {
			t.Errorf("%s = %q, want %q", key, got, value)
		}
	}
	if query.Has("code_verifier") {
		t.Error("the authorization URL leaks the PKCE verifier")
	}
	if u.Scheme+"://"+u.Host+u.Path != f.server.URL+"/auth" {
		t.Errorf("authorization endpoint = %s", u)
	}
}

func TestAuthorizationURLWithoutHint(t *testing.T) {
	p := newProvider(t, newFakeGoogle(t))

	u, err := url.Parse(p.AuthorizationURL(newPending(), ""))
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Has("login_hint") {
		t.Fatalf("login_hint in %s", u)
	}
}

func TestRejectedLogins(t *testing.T) {
	other, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name           string
		change         func(f *fakeGoogle, pending *out.PendingLogin)
		domainRejected bool
	}{
		{"PKCE verifier of another login", func(_ *fakeGoogle, pending *out.PendingLogin) {
			pending.CodeVerifier = oauth2.GenerateVerifier()
		}, false},
		{"nonce of another login", func(_ *fakeGoogle, pending *out.PendingLogin) { pending.Nonce = "other" }, false},
		{"empty nonce", func(f *fakeGoogle, pending *out.PendingLogin) {
			pending.Nonce = ""
			f.claims["nonce"] = ""
		}, false},
		{"signed with another key", func(f *fakeGoogle, _ *out.PendingLogin) { f.signer = other }, false},
		{"another issuer", func(f *fakeGoogle, _ *out.PendingLogin) { f.claims["iss"] = "https://evil.test" }, false},
		{"another audience", func(f *fakeGoogle, _ *out.PendingLogin) { f.claims["aud"] = "other-client" }, false},
		{"expired", func(f *fakeGoogle, _ *out.PendingLogin) {
			f.claims["exp"] = time.Now().Add(-time.Minute).Unix()
		}, false},
		{"email not verified", func(f *fakeGoogle, _ *out.PendingLogin) { f.claims["email_verified"] = false }, false},
		{"no ID token", func(f *fakeGoogle, _ *out.PendingLogin) { f.noIDToken = true }, false},
		{"account outside any Workspace, with an email of the domain", func(f *fakeGoogle, _ *out.PendingLogin) { delete(f.claims, "hd") }, true},
		{"account of another Workspace", func(f *fakeGoogle, _ *out.PendingLogin) { f.claims["hd"] = "uba.ar" }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFakeGoogle(t)
			p := newProvider(t, f)
			pending := newPending()
			authorize(t, f, p, pending)
			tt.change(f, &pending)

			_, err := p.Exchange(t.Context(), issuedCode, pending)
			if err == nil {
				t.Fatal("the login was accepted")
			}
			if got := errors.Is(err, out.ErrDomainNotAllowed); got != tt.domainRejected {
				t.Fatalf("domain rejection = %v, want %v: %v", got, tt.domainRejected, err)
			}
		})
	}
}

func TestHostedDomainIgnoresCase(t *testing.T) {
	f := newFakeGoogle(t)
	p := newProvider(t, f)
	pending := newPending()
	authorize(t, f, p, pending)
	f.claims["hd"] = "FI.UBA.AR"

	if _, err := p.Exchange(t.Context(), issuedCode, pending); err != nil {
		t.Fatal(err)
	}
}

func TestNewIdentityProviderNeedsTheWholeConfig(t *testing.T) {
	f := newFakeGoogle(t)
	complete := google.Config{Issuer: f.server.URL, ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL, HostedDomain: hostedDomain}
	tests := map[string]func(*google.Config){
		"issuer":        func(c *google.Config) { c.Issuer = "" },
		"client ID":     func(c *google.Config) { c.ClientID = "" },
		"client secret": func(c *google.Config) { c.ClientSecret = "" },
		"redirect URL":  func(c *google.Config) { c.RedirectURL = "" },
		"hosted domain": func(c *google.Config) { c.HostedDomain = "" },
	}
	for missing, change := range tests {
		t.Run(missing, func(t *testing.T) {
			cfg := complete
			change(&cfg)
			if _, err := google.NewIdentityProvider(t.Context(), cfg, slog.New(slog.DiscardHandler)); err == nil {
				t.Fatalf("started without %s", missing)
			}
		})
	}
}

func TestNewIdentityProviderFailsWithoutDiscovery(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	cfg := google.Config{Issuer: server.URL, ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL, HostedDomain: hostedDomain}
	if _, err := google.NewIdentityProvider(t.Context(), cfg, slog.New(slog.DiscardHandler)); err == nil {
		t.Fatal("started without a discovery document")
	}
}
