package rest_test

import (
	"application"
	"context"
	"dev"
	"domain"
	"memory"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"rest"
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

func expectStatus(t *testing.T, res *http.Response, err error, want int) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != want {
		t.Fatalf("%s %s: got %d, want %d", res.Request.Method, res.Request.URL.Path, res.StatusCode, want)
	}
}

func TestProviderLoginThroughToProfileAndLogout(t *testing.T) {
	srv, _ := newServer(t)
	client := newClient(t)

	// /auth/login -> dev provider -> /auth/callback -> /me, following redirects.
	res, err := client.Get(srv.URL + "/auth/login?login_hint=" + url.QueryEscape("operativo@fi.uba.ar"))
	expectStatus(t, res, err, http.StatusOK)

	res, err = client.Post(srv.URL+"/auth/logout", "", nil)
	expectStatus(t, res, err, http.StatusNoContent)

	res, err = client.Get(srv.URL + "/me")
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestCallbackWithoutMatchingStateIsRejected(t *testing.T) {
	srv, _ := newServer(t)

	// A callback the browser did not start, as in a login CSRF.
	res, err := newClient(t).Get(srv.URL + "/auth/callback?code=operativo@fi.uba.ar&state=forged")
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestUnknownIdentityGetsNoSession(t *testing.T) {
	srv, _ := newServer(t)

	res, err := newClient(t).Get(srv.URL + "/auth/login?login_hint=nadie@fi.uba.ar")
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestProviderLoginOutsideTheAllowedDomainGetsNoSession(t *testing.T) {
	srv, _ := newServer(t)

	res, err := newClient(t).Get(srv.URL + "/auth/login?login_hint=titular@example.com")
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestLoginLinkThroughToProfile(t *testing.T) {
	srv, mail := newServer(t)
	client := newClient(t)

	res, err := client.Post(srv.URL+"/auth/login-links", "application/json", strings.NewReader(`{"email":"titular@fi.uba.ar"}`))
	expectStatus(t, res, err, http.StatusAccepted)

	body := `{"token":"` + strings.TrimPrefix(mail.last, "#token=") + `"}`
	res, err = client.Post(srv.URL+"/auth/login-links/redeem", "application/json", strings.NewReader(body))
	expectStatus(t, res, err, http.StatusNoContent)

	res, err = client.Get(srv.URL + "/me")
	expectStatus(t, res, err, http.StatusOK)

	// A second redemption fails even from a fresh browser.
	res, err = newClient(t).Post(srv.URL+"/auth/login-links/redeem", "application/json", strings.NewReader(body))
	expectStatus(t, res, err, http.StatusUnauthorized)
}

func TestLoginLinkRequestDoesNotRevealAccounts(t *testing.T) {
	srv, mail := newServer(t)

	res, err := newClient(t).Post(srv.URL+"/auth/login-links", "application/json", strings.NewReader(`{"email":"nadie@fi.uba.ar"}`))
	expectStatus(t, res, err, http.StatusAccepted)
	if mail.last != "" {
		t.Fatalf("sent %q to an unknown email", mail.last)
	}
}

func TestLoginLinkOutsideTheAllowedDomainIsRejected(t *testing.T) {
	srv, mail := newServer(t)

	res, err := newClient(t).Post(srv.URL+"/auth/login-links", "application/json", strings.NewReader(`{"email":"titular@example.com"}`))
	expectStatus(t, res, err, http.StatusBadRequest)
	if mail.last != "" {
		t.Fatalf("sent %q outside the allowed domain", mail.last)
	}
}
