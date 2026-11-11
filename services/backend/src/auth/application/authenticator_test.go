package application_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"backend/src/auth/adapters/dev"
	"backend/src/auth/adapters/memory"
	"backend/src/auth/application"
	"backend/src/auth/domain"
	out "backend/src/auth/ports/output"
)

var (
	applicant = domain.User{ID: "u1", Email: "titular@fi.uba.ar", Name: "Titular"}
	operative = domain.User{ID: "u2", Email: "operativo@fi.uba.ar", Name: "Operativo"}
)

type fixture struct {
	t     *testing.T
	auth  *application.AuthService
	users *memory.UserRepository
	now   time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		t:     t,
		users: memory.NewUserRepository(applicant, operative),
		now:   time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC),
	}
	f.auth = application.NewAuthService(
		application.AuthConfig{
			IdleTimeout:   30 * time.Minute,
			AllowedDomain: "fi.uba.ar",
		},
		dev.NewIdentityProvider("http://api.test/auth/callback", ""),
		f.users,
		memory.NewSessionStore(),
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

func (f *fixture) begin(email string) (code, state string, pending domain.PendingLogin) {
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

func TestCompleteLoginRegistersANewIdentityAsApplicant(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	var ids []string
	for range 2 {
		token, err := f.login("Nuevo@fi.uba.ar")
		if err != nil {
			t.Fatal(err)
		}
		user, err := f.auth.Authenticate(ctx, token)
		if err != nil {
			t.Fatal(err)
		}
		if user.Email != "nuevo@fi.uba.ar" {
			t.Fatalf("got %+v, want the normalized email", user)
		}
		ids = append(ids, user.ID)
	}
	// The second login finds the profile the first one registered.
	if ids[0] != ids[1] {
		t.Fatalf("got two profiles %v for the same person", ids)
	}
}

func TestCompleteLoginOnlyAcceptsTheAllowedDomain(t *testing.T) {
	f := newFixture(t)

	// Lookalike domains are rejected too.
	for _, email := range []string{"titular@example.com", "operativo@evilfi.uba.ar", "operativo@fi.uba.ar.evil.com"} {
		t.Run(email, func(t *testing.T) {
			if _, err := f.login(email); !errors.Is(err, domain.ErrDomainNotAllowed) {
				t.Errorf("login(%q) = %v, want ErrDomainNotAllowed", email, err)
			}
		})
	}
}

func TestCompleteLoginRejectsWhatTheProviderRejects(t *testing.T) {
	f := newFixture(t)

	_, err := f.login("")
	if !errors.Is(err, domain.ErrLoginRejected) {
		t.Fatalf("got %v, want ErrLoginRejected", err)
	}
}

// outsideDomain stands for a provider that checks the domain itself, as the
// Google adapter does with the hd claim.
type outsideDomain struct{ out.IdentityProvider }

func (outsideDomain) Exchange(context.Context, string, domain.PendingLogin) (out.Identity, error) {
	return out.Identity{}, fmt.Errorf("%w: hosted domain %q", domain.ErrDomainNotAllowed, "uba.ar")
}

func TestCompleteLoginKeepsTheDomainRejectionOfTheProvider(t *testing.T) {
	auth := application.NewAuthService(
		application.AuthConfig{AllowedDomain: "fi.uba.ar"},
		outsideDomain{},
		memory.NewUserRepository(applicant),
		memory.NewSessionStore(),
		time.Now,
	)
	pending := domain.PendingLogin{State: "s", Nonce: "n", CodeVerifier: "v"}

	_, err := auth.CompleteLogin(context.Background(), "code", "s", pending)
	if !errors.Is(err, domain.ErrDomainNotAllowed) || !errors.Is(err, domain.ErrLoginRejected) {
		t.Fatalf("got %v, want ErrDomainNotAllowed within ErrLoginRejected", err)
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
	if _, err := f.auth.Authenticate(ctx, token); !errors.Is(err, domain.ErrSessionExpired) {
		t.Fatalf("got %v, want ErrSessionExpired", err)
	}
	if _, err := f.auth.Authenticate(ctx, token); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expired session still stored: %v", err)
	}
}

func TestLogoutClosesTheSession(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	token, err := f.login(operative.Email)
	if err != nil {
		t.Fatal(err)
	}

	if err := f.auth.Logout(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := f.auth.Authenticate(ctx, token); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("got %v, want ErrSessionNotFound", err)
	}
}

func TestUnknownSessionTokenIsRejected(t *testing.T) {
	f := newFixture(t)

	tests := map[string]string{
		"empty token":   "",
		"unknown token": "not-a-session",
	}
	for name, token := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := f.auth.Authenticate(context.Background(), token); !errors.Is(err, domain.ErrSessionNotFound) {
				t.Errorf("Authenticate(%q) = %v, want ErrSessionNotFound", token, err)
			}
		})
	}
}

func TestBeginLoginCreatesAFreshPendingLoginEachTime(t *testing.T) {
	f := newFixture(t)

	_, _, a := f.begin(operative.Email)
	_, _, b := f.begin(operative.Email)
	for _, p := range []domain.PendingLogin{a, b} {
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
		pending domain.PendingLogin
	}{
		"forged state":        {"forged", pending},
		"no pending login":    {state, domain.PendingLogin{}},
		"empty state on both": {"", domain.PendingLogin{Nonce: pending.Nonce, CodeVerifier: pending.CodeVerifier}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := f.auth.CompleteLogin(ctx, code, c.state, c.pending); !errors.Is(err, domain.ErrLoginStateMismatch) {
				t.Errorf("CompleteLogin(state %q) = %v, want ErrLoginStateMismatch", c.state, err)
			}
		})
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
	if !errors.Is(err, domain.ErrLoginRejected) {
		t.Fatalf("got %v, want ErrLoginRejected", err)
	}
}
