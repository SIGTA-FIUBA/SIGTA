package application_test

import (
	"application"
	"context"
	"dev"
	"domain"
	"errors"
	"memory"
	"net/url"
	out "output"
	"strings"
	"testing"
	"time"
)

const linkURL = "http://app.test/login/link#token="

var (
	applicant = domain.User{ID: "u1", Email: "titular@fi.uba.ar", Name: "Titular", Role: domain.RoleApplicant}
	operative = domain.User{ID: "u2", Email: "operativo@fi.uba.ar", Name: "Operativo", Role: domain.RoleStaffOperative}
)

type fixture struct {
	t      *testing.T
	auth   *application.AuthService
	mailer *outbox
	now    time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, mailer: &outbox{}, now: time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)}
	f.auth = application.NewAuthService(
		application.AuthConfig{
			IdleTimeout:   30 * time.Minute,
			LoginLinkTTL:  15 * time.Minute,
			AllowedDomain: "fi.uba.ar",
			LoginLinkURL:  linkURL,
		},
		dev.NewIdentityProvider("http://api.test/auth/callback"),
		memory.NewUserRepository(applicant, operative),
		memory.NewSessionStore(),
		memory.NewLoginLinkStore(),
		f.mailer,
		func() time.Time { return f.now },
	)
	return f
}

func (f *fixture) advance(d time.Duration) { f.now = f.now.Add(d) }

// login goes through the provider like a browser would: BeginLogin, then the
// callback with the code and state the provider sends back.
func (f *fixture) login(email string) (string, error) {
	code, state, pending := f.begin(email)
	return f.auth.CompleteLogin(context.Background(), code, state, pending)
}

func (f *fixture) begin(email string) (code, state string, pending out.PendingLogin) {
	f.t.Helper()
	location, pending, err := f.auth.BeginLogin(email)
	if err != nil {
		f.t.Fatal(err)
	}
	u, err := url.Parse(location)
	if err != nil {
		f.t.Fatal(err)
	}
	return u.Query().Get("code"), u.Query().Get("state"), pending
}

type outbox struct{ links map[string]string }

func (o *outbox) SendLoginLink(_ context.Context, to, link string) error {
	if o.links == nil {
		o.links = make(map[string]string)
	}
	o.links[to] = strings.TrimPrefix(link, linkURL)
	return nil
}

func TestCompleteLoginResolvesProfileFromSIGTA(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	// The provider sends back the email with different casing; the profile
	// still comes from SIGTA's own records.
	token, err := f.login("Operativo@FI.UBA.AR")
	if err != nil {
		t.Fatal(err)
	}
	user, err := f.auth.Authenticate(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	if user != operative {
		t.Fatalf("got %+v, want %+v", user, operative)
	}
}

func TestCompleteLoginRejectsIdentityUnknownToSIGTA(t *testing.T) {
	f := newFixture(t)

	_, err := f.login("nadie@fi.uba.ar")
	if !errors.Is(err, domain.ErrUnknownUser) {
		t.Fatalf("got %v, want ErrUnknownUser", err)
	}
}

func TestCompleteLoginOnlyAcceptsTheAllowedDomain(t *testing.T) {
	f := newFixture(t)

	// Lookalike domains are rejected too.
	for _, email := range []string{"titular@example.com", "operativo@evilfi.uba.ar", "operativo@fi.uba.ar.evil.com"} {
		if _, err := f.login(email); !errors.Is(err, out.ErrDomainNotAllowed) {
			t.Fatalf("%s: got %v, want ErrDomainNotAllowed", email, err)
		}
	}
}

func TestCompleteLoginRejectsWhatTheProviderRejects(t *testing.T) {
	f := newFixture(t)

	_, err := f.login("")
	if !errors.Is(err, out.ErrLoginRejected) {
		t.Fatalf("got %v, want ErrLoginRejected", err)
	}
}

func TestSessionExpiresAfterInactivity(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	token, err := f.login(operative.Email)
	if err != nil {
		t.Fatal(err)
	}

	// Each request within the timeout extends the session.
	for range 3 {
		f.advance(29 * time.Minute)
		if _, err := f.auth.Authenticate(ctx, token); err != nil {
			t.Fatalf("active session rejected: %v", err)
		}
	}

	f.advance(30 * time.Minute)
	if _, err := f.auth.Authenticate(ctx, token); !errors.Is(err, out.ErrSessionExpired) {
		t.Fatalf("got %v, want ErrSessionExpired", err)
	}
	if _, err := f.auth.Authenticate(ctx, token); !errors.Is(err, out.ErrSessionNotFound) {
		t.Fatalf("expired session still stored: %v", err)
	}
}

func TestLogoutClosesTheSession(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	token, _ := f.login(operative.Email)

	if err := f.auth.Logout(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := f.auth.Authenticate(ctx, token); !errors.Is(err, out.ErrSessionNotFound) {
		t.Fatalf("got %v, want ErrSessionNotFound", err)
	}
}

func TestUnknownSessionTokenIsRejected(t *testing.T) {
	f := newFixture(t)

	for _, token := range []string{"", "not-a-session"} {
		if _, err := f.auth.Authenticate(context.Background(), token); !errors.Is(err, out.ErrSessionNotFound) {
			t.Fatalf("token %q: got %v, want ErrSessionNotFound", token, err)
		}
	}
}

func TestLoginLinkOpensASessionOnce(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if err := f.auth.RequestLoginLink(ctx, " Titular@FI.UBA.AR "); err != nil {
		t.Fatal(err)
	}
	link := f.mailer.links[applicant.Email]
	if link == "" {
		t.Fatal("no login link sent")
	}

	token, err := f.auth.RedeemLoginLink(ctx, link)
	if err != nil {
		t.Fatal(err)
	}
	if user, err := f.auth.Authenticate(ctx, token); err != nil || user.ID != applicant.ID {
		t.Fatalf("got %+v, %v", user, err)
	}

	if _, err := f.auth.RedeemLoginLink(ctx, link); !errors.Is(err, out.ErrInvalidLoginLink) {
		t.Fatalf("reused link: got %v, want ErrInvalidLoginLink", err)
	}
}

func TestLoginLinkExpires(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	_ = f.auth.RequestLoginLink(ctx, applicant.Email)

	f.advance(15 * time.Minute)
	if _, err := f.auth.RedeemLoginLink(ctx, f.mailer.links[applicant.Email]); !errors.Is(err, out.ErrInvalidLoginLink) {
		t.Fatalf("got %v, want ErrInvalidLoginLink", err)
	}
}

func TestLoginLinkIsOnlyForApplicants(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	for _, email := range []string{operative.Email, "nadie@fi.uba.ar"} {
		if err := f.auth.RequestLoginLink(ctx, email); err != nil {
			t.Fatalf("%s: %v", email, err)
		}
	}
	if len(f.mailer.links) != 0 {
		t.Fatalf("sent links to %v", f.mailer.links)
	}
}

func TestLoginLinkOnlyAcceptsTheAllowedDomain(t *testing.T) {
	f := newFixture(t)

	err := f.auth.RequestLoginLink(context.Background(), "titular@example.com")
	if !errors.Is(err, out.ErrDomainNotAllowed) {
		t.Fatalf("got %v, want ErrDomainNotAllowed", err)
	}
	if len(f.mailer.links) != 0 {
		t.Fatalf("sent links to %v", f.mailer.links)
	}
}

func TestBeginLoginCreatesAFreshPendingLoginEachTime(t *testing.T) {
	f := newFixture(t)

	_, _, a := f.begin(operative.Email)
	_, _, b := f.begin(operative.Email)
	for _, p := range []out.PendingLogin{a, b} {
		if p.State == "" || p.Nonce == "" || p.CodeVerifier == "" {
			t.Fatalf("incomplete pending login %+v", p)
		}
	}
	if a.State == b.State || a.Nonce == b.Nonce || a.CodeVerifier == b.CodeVerifier {
		t.Fatalf("pending logins share values: %+v, %+v", a, b)
	}
}

func TestCompleteLoginRejectsACallbackThisBrowserDidNotStart(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	code, state, pending := f.begin(operative.Email)

	cases := map[string]struct {
		state   string
		pending out.PendingLogin
	}{
		"forged state":        {"forged", pending},
		"no pending login":    {state, out.PendingLogin{}},
		"empty state on both": {"", out.PendingLogin{Nonce: pending.Nonce, CodeVerifier: pending.CodeVerifier}},
	}
	for name, c := range cases {
		if _, err := f.auth.CompleteLogin(ctx, code, c.state, c.pending); !errors.Is(err, out.ErrLoginStateMismatch) {
			t.Fatalf("%s: got %v, want ErrLoginStateMismatch", name, err)
		}
	}
}

func TestCompleteLoginRejectsAStolenCode(t *testing.T) {
	f := newFixture(t)

	// An attacker who got the victim's code and state from the URL, but not
	// their cookie, completes it with a login of their own.
	code, state, _ := f.begin(operative.Email)
	_, _, attacker := f.begin(operative.Email)
	attacker.State = state

	_, err := f.auth.CompleteLogin(context.Background(), code, state, attacker)
	if !errors.Is(err, out.ErrLoginRejected) {
		t.Fatalf("got %v, want ErrLoginRejected", err)
	}
}
