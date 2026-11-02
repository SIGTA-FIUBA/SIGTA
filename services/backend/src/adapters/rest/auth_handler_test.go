package rest_test

import (
	"backend/src/adapters/dev"
	"backend/src/adapters/memory"
	"backend/src/adapters/rest"
	"backend/src/application"
	"backend/src/domain"
	ports "backend/src/ports/input"
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
)

type outbox struct{ last string }

func (o *outbox) SendLoginLink(_ context.Context, _, link string) error {
	o.last = link
	return nil
}

func newServer(t *testing.T) (*httptest.Server, *outbox) {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mail := &outbox{}
	auth := application.NewAuthService(
		application.AuthConfig{IdleTimeout: time.Hour, LoginLinkTTL: time.Hour, AllowedDomain: "fi.uba.ar", LoginLinkURL: "#token="},
		dev.NewIdentityProvider(srv.URL+"/auth/callback"),
		memory.NewUserRepository(
			domain.User{ID: "u1", Email: "titular@fi.uba.ar", Name: "Titular", Role: domain.RoleApplicant},
			domain.User{ID: "u2", Email: "operativo@fi.uba.ar", Name: "Operativo", Role: domain.RoleStaffOperative},
		),
		memory.NewSessionStore(),
		memory.NewLoginLinkStore(),
		mail,
		time.Now,
	)
	api := rest.NewServer(rest.NewAuthHandler(auth, srv.URL+"/me", false))
	rest.Mount(mux, api, slog.Default())
	return srv, mail
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
		t.Fatalf("Content-Type = %q", got)
	}
	var body rest.Error
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body.Message
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
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar}
}

func get(t *testing.T, client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return client.Do(req)
}

// post sends body as JSON; a nil body sends an empty request.
func post(t *testing.T, client *http.Client, url string, body io.Reader) (*http.Response, error) {
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
	srv, _ := newServer(t)
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
	srv, _ := newServer(t)
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
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestFailedCallbackExpiresThePendingLogin(t *testing.T) {
	srv, _ := newServer(t)
	client := withoutRedirects(newClient(t))

	res, err := get(t, client, srv.URL+"/auth/login?login_hint=nadie@fi.uba.ar")
	expectStatus(t, res, err, http.StatusFound)

	res, err = get(t, client, res.Header.Get("Location"))
	expectStatus(t, res, err, http.StatusUnauthorized)
	if !expiresLoginState(res) {
		t.Fatal("the failed callback left the pending login")
	}
}

func TestCallbackTellsARejectedLoginFromAnUnknownUser(t *testing.T) {
	srv, _ := newServer(t)

	// The development provider rejects a login without a hint.
	res, err := get(t, newClient(t), srv.URL+"/auth/login")
	rejected := expectError(t, res, err, http.StatusUnauthorized)

	res, err = get(t, newClient(t), srv.URL+"/auth/login?login_hint=nadie@fi.uba.ar")
	unknown := expectError(t, res, err, http.StatusUnauthorized)

	if rejected == unknown || strings.Contains(rejected, "cuenta de SIGTA") {
		t.Fatalf("rejected = %q, unknown = %q", rejected, unknown)
	}
}

func TestInvalidRequestsAnswerWithTheContractError(t *testing.T) {
	srv, _ := newServer(t)
	client := newClient(t)

	tests := []struct {
		name string
		do   func() (*http.Response, error)
	}{
		{"callback without code", func() (*http.Response, error) {
			return get(t, client, srv.URL+"/auth/callback?state=x")
		}},
		{"invalid email", func() (*http.Response, error) {
			return post(t, client, srv.URL+"/auth/login-links", strings.NewReader(`{"email":"no"}`))
		}},
		{"broken JSON", func() (*http.Response, error) {
			return post(t, client, srv.URL+"/auth/login-links", strings.NewReader(`{`))
		}},
		{"empty token", func() (*http.Response, error) {
			return post(t, client, srv.URL+"/auth/login-links/redeem", strings.NewReader(`{}`))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.do()
			if got := expectError(t, res, err, http.StatusBadRequest); got != "El pedido no es válido." {
				t.Fatalf("message = %q", got)
			}
		})
	}
}

type brokenAuth struct{ ports.Authenticator }

func (brokenAuth) Logout(context.Context, string) error { return errors.New("secret detail") }

func TestInternalErrorsDoNotExposeTheirMessage(t *testing.T) {
	handler := rest.Mount(http.NewServeMux(), rest.NewServer(rest.NewAuthHandler(brokenAuth{}, "", false)), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	res, err := post(t, newClient(t), srv.URL+"/auth/logout", nil)
	if got := expectError(t, res, err, http.StatusInternalServerError); strings.Contains(got, "secret") {
		t.Fatalf("message = %q", got)
	}
}

func TestCallbackWithoutMatchingStateIsRejected(t *testing.T) {
	srv, _ := newServer(t)

	// A callback the browser did not start, as in a login CSRF.
	res, err := get(t, newClient(t), srv.URL+"/auth/callback?code=operativo@fi.uba.ar&state=forged")
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestUnknownIdentityGetsNoSession(t *testing.T) {
	srv, _ := newServer(t)

	res, err := get(t, newClient(t), srv.URL+"/auth/login?login_hint=nadie@fi.uba.ar")
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestProviderLoginOutsideTheAllowedDomainGetsNoSession(t *testing.T) {
	srv, _ := newServer(t)

	res, err := get(t, newClient(t), srv.URL+"/auth/login?login_hint=titular@example.com")
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestLoginLinkThroughToProfile(t *testing.T) {
	srv, mail := newServer(t)
	client := newClient(t)

	res, err := post(t, client, srv.URL+"/auth/login-links", strings.NewReader(`{"email":"titular@fi.uba.ar"}`))
	expectStatus(t, res, err, http.StatusAccepted)

	body := `{"token":"` + strings.TrimPrefix(mail.last, "#token=") + `"}`
	res, err = post(t, client, srv.URL+"/auth/login-links/redeem", strings.NewReader(body))
	expectStatus(t, res, err, http.StatusNoContent)

	res, err = get(t, client, srv.URL+"/me")
	expectStatus(t, res, err, http.StatusOK)

	// A second redemption fails even from a fresh browser.
	res, err = post(t, newClient(t), srv.URL+"/auth/login-links/redeem", strings.NewReader(body))
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestLoginLinkRequestDoesNotRevealAccounts(t *testing.T) {
	srv, mail := newServer(t)

	res, err := post(t, newClient(t), srv.URL+"/auth/login-links", strings.NewReader(`{"email":"nadie@fi.uba.ar"}`))
	expectStatus(t, res, err, http.StatusAccepted)
	if mail.last != "" {
		t.Fatalf("sent %q to an unknown email", mail.last)
	}
}

func TestLoginLinkOutsideTheAllowedDomainIsRejected(t *testing.T) {
	srv, mail := newServer(t)

	res, err := post(t, newClient(t), srv.URL+"/auth/login-links", strings.NewReader(`{"email":"titular@example.com"}`))
	expectStatus(t, res, err, http.StatusBadRequest)
	if mail.last != "" {
		t.Fatalf("sent %q outside the allowed domain", mail.last)
	}
}
