package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/auth/adapters/dev"
	"backend/src/auth/adapters/memory"
	"backend/src/auth/adapters/rest"
	"backend/src/auth/application"
	"backend/src/auth/domain"
	ports "backend/src/auth/ports/input"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	auth := application.NewAuthService(
		application.AuthConfig{IdleTimeout: time.Hour, AllowedDomain: "fi.uba.ar"},
		dev.NewIdentityProvider(srv.URL+"/auth/callback", ""),
		memory.NewUserRepository(
			domain.User{ID: "u1", Email: "titular@fi.uba.ar", Name: "Titular"},
			domain.User{ID: "u2", Email: "operativo@fi.uba.ar", Name: "Operativo"},
			domain.User{ID: "u3", Email: "admin@fi.uba.ar", Name: "Administración"},
		),
		memory.NewSessionStore(),
		time.Now,
	)
	areaRepository := memory.NewAreaRepository(domain.Area{ID: "office", Name: "Secretaría", Permissions: []domain.Permission{{ID: "review", Name: "Revisar"}, {ID: "approve", Name: "Aprobar"}}})
	administrators := memory.NewAdministrators("u3")
	membershipRepository := memory.NewMembershipRepository(domain.Membership{ID: "m1", UserID: "u2", Area: "office", Permissions: []string{"review"}})
	areas := application.NewAreaService(areaRepository, administrators, membershipRepository, time.Now)
	memberships := application.NewMembershipService(areaRepository, administrators, membershipRepository, func() string { return "m2" }, time.Now)
	server := rest.NewAuthHandler(auth, memberships, srv.URL+"/me", srv.URL+"/login", false)
	administration := rest.NewAdministrationHandler(auth, areas, memberships)
	api.Mount(mux, api.NewServer(server, nil, administration), slog.Default())
	return srv
}

// withoutRedirects leaves a client on each redirect, so a test can read it.
func withoutRedirects(client *http.Client) *http.Client {
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return client
}

// expectError checks that a response follows the contract's Error, and returns
// its message.
func expectError(t *testing.T, res *http.Response, err error, want int) string {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := res.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != want {
		t.Fatalf("%s %s: got %d, want %d", res.Request.Method, res.Request.URL.Path, res.StatusCode, want)
	}
	if got := res.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}
	var body api.Error
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body.Message
}

func expectLoginError(t *testing.T, srv *httptest.Server, res *http.Response, err error) string {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if err := res.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusFound {
		t.Fatalf("callback: got %d, want a redirect", res.StatusCode)
	}
	location, err := url.Parse(res.Header.Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if location.Scheme+"://"+location.Host+location.Path != srv.URL+"/login" {
		t.Fatalf("redirected to %s, want the login page", location)
	}
	for _, c := range res.Cookies() {
		if c.Name == "sigta_session" {
			t.Fatal("a failed login set a session")
		}
	}
	return location.Query().Get("error")
}

// expiresLoginState tells whether a response removes the pending login.
func expiresLoginState(res *http.Response) bool {
	for _, c := range res.Cookies() {
		if c.Name == "sigta_login_state" && c.MaxAge < 0 {
			return true
		}
	}
	return false
}

func newClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar}
}

func get(t *testing.T, client *http.Client, url string) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return client.Do(req)
}

// post sends body as JSON; a nil body sends an empty request.
func post(t *testing.T, client *http.Client, url string, body io.Reader) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return client.Do(req)
}

func expectStatus(t *testing.T, res *http.Response, err error, want int) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	if err := res.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != want {
		t.Fatalf("%s %s: got %d, want %d", res.Request.Method, res.Request.URL.Path, res.StatusCode, want)
	}
}

func TestProviderLoginThroughToProfileAndLogout(t *testing.T) {
	srv := newServer(t)
	client := newClient(t)

	// /auth/login -> dev provider -> /auth/callback -> /me, following redirects.
	res, err := get(t, client, srv.URL+"/auth/login?login_hint="+url.QueryEscape("operativo@fi.uba.ar"))
	expectStatus(t, res, err, http.StatusOK)

	res, err = post(t, client, srv.URL+"/auth/logout", nil)
	expectStatus(t, res, err, http.StatusNoContent)

	res, err = get(t, client, srv.URL+"/me")
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestPendingLoginWorksOnce(t *testing.T) {
	srv := newServer(t)
	client := withoutRedirects(newClient(t))

	res, err := get(t, client, srv.URL+"/auth/login?login_hint="+url.QueryEscape("operativo@fi.uba.ar"))
	expectStatus(t, res, err, http.StatusFound)
	callback := res.Header.Get("Location")

	res, err = get(t, client, callback)
	expectStatus(t, res, err, http.StatusFound)
	if !expiresLoginState(res) {
		t.Fatal("the successful callback left the pending login")
	}
	if len(res.Cookies()) != 2 {
		t.Fatalf("got cookies %v, want the session and the expired pending login", res.Cookies())
	}

	res, err = get(t, client, callback)
	if reason := expectLoginError(t, srv, res, err); reason != "state_mismatch" {
		t.Fatalf("second callback: reason = %q, want %q", reason, "state_mismatch")
	}
}

func TestFailedCallbackExpiresThePendingLogin(t *testing.T) {
	srv := newServer(t)
	client := withoutRedirects(newClient(t))

	res, err := get(t, client, srv.URL+"/auth/login?login_hint=titular@example.com")
	expectStatus(t, res, err, http.StatusFound)

	res, err = get(t, client, res.Header.Get("Location"))
	expectLoginError(t, srv, res, err)
	if !expiresLoginState(res) {
		t.Fatal("the failed callback left the pending login")
	}
}

func TestCallbackTellsWhyTheLoginFailed(t *testing.T) {
	srv := newServer(t)
	client := withoutRedirects(newClient(t))

	tests := []struct {
		name     string
		login    string
		callback func(url.Values)
		want     string
	}{
		// The development provider rejects a login without a hint.
		{"rejected by the provider", "/auth/login", nil, "rejected"},
		{"cancelled at the provider", "/auth/login?login_hint=operativo@fi.uba.ar", func(q url.Values) {
			q.Del("code")
			q.Set("error", "access_denied")
		}, "rejected"},
		{"outside the domain", "/auth/login?login_hint=titular@example.com", nil, "domain_not_allowed"},
		{"state of another login", "/auth/login?login_hint=operativo@fi.uba.ar", func(q url.Values) { q.Set("state", "forged") }, "state_mismatch"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := get(t, client, srv.URL+tt.login)
			expectStatus(t, res, err, http.StatusFound)
			callback, err := url.Parse(res.Header.Get("Location"))
			if err != nil {
				t.Fatal(err)
			}
			if tt.callback != nil {
				query := callback.Query()
				tt.callback(query)
				callback.RawQuery = query.Encode()
			}

			res, err = get(t, client, callback.String())
			if reason := expectLoginError(t, srv, res, err); reason != tt.want {
				t.Fatalf("reason = %q, want %q", reason, tt.want)
			}
		})
	}
}

func TestInvalidRequestsAnswerWithTheContractError(t *testing.T) {
	srv := newServer(t)

	res, err := get(t, newClient(t), srv.URL+"/auth/callback?code=x")
	if got := expectError(t, res, err, http.StatusBadRequest); got != "El pedido no es válido." {
		t.Fatalf("message = %q, want %q", got, "El pedido no es válido.")
	}
}

type brokenAuth struct{ ports.Authenticator }

func (brokenAuth) Logout(context.Context, string) error { return errors.New("secret detail") }

func TestInternalErrorsDoNotExposeTheirMessage(t *testing.T) {
	handler := api.Mount(http.NewServeMux(), api.NewServer(rest.NewAuthHandler(brokenAuth{}, nil, "", "", false), nil, nil), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	res, err := post(t, newClient(t), srv.URL+"/auth/logout", nil)
	if got := expectError(t, res, err, http.StatusInternalServerError); strings.Contains(got, "secret") {
		t.Fatalf("message = %q, want one without %q", got, "secret")
	}
}

func TestCallbackWithoutMatchingStateIsRejected(t *testing.T) {
	srv := newServer(t)

	// A callback the browser did not start, as in a login CSRF.
	res, err := get(t, withoutRedirects(newClient(t)), srv.URL+"/auth/callback?code=operativo@fi.uba.ar&state=forged")
	if reason := expectLoginError(t, srv, res, err); reason != "state_mismatch" {
		t.Fatalf("forged state: reason = %q, want %q", reason, "state_mismatch")
	}
}
