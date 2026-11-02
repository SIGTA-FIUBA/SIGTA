package rest_test

import (
	"backend/src/adapters/dev"
	"backend/src/adapters/memory"
	"backend/src/adapters/rest"
	"backend/src/application"
	"backend/src/domain"
	"context"
	"io"
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
	rest.HandlerFromMux(rest.NewStrictHandler(api, nil), mux)
	return srv, mail
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
