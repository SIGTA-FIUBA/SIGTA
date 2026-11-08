package acceptance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"backend/src/api"
	"backend/src/auth/adapters/dev"
	"backend/src/auth/adapters/memory"
	authrest "backend/src/auth/adapters/rest"
	"backend/src/auth/application"
	"backend/src/auth/domain"
)

const (
	baseURL     = "http://sigta.test"
	loginPage   = baseURL + "/login"
	idleTimeout = 30 * time.Minute
)

// authenticationWorld is the state of one scenario of US06. Each scenario goes
// through the REST adapter, which it calls in process.
type authenticationWorld struct {
	domain      string
	users       *memory.UserRepository
	memberships *memory.MembershipRepository
	areas       []domain.Area
	now         time.Time
	handler     http.Handler

	client *http.Client

	// The result of the last request.
	status     int
	body       string
	location   string
	sessionSet bool
	// callback is the last provider response the browser arrived with.
	callback string
}

func registerAuthenticationSteps(sc *godog.ScenarioContext) {
	w := &authenticationWorld{
		users:       memory.NewUserRepository(),
		memberships: memory.NewMembershipRepository(),
		now:         time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC),
	}

	sc.Given(`^only accounts from the "([^"]*)" domain are allowed$`, w.givenDomain)
	sc.Given(`^SIGTA has "([^"]*)" registered as an applicant$`, w.givenApplicant)
	sc.Given(`^SIGTA has "([^"]*)" registered as staff of the "([^"]*)" area$`, w.givenStaff)
	sc.Given(`^"([^"]*)" completed a sign-in$`, w.givenSignedIn)
	sc.Given(`^"([^"]*)" has an open session$`, w.givenSignedIn)

	sc.When(`^"([^"]*)" signs in through the identity provider$`, w.signIn)
	sc.When(`^the identity provider rejects a sign-in$`, w.signInRejected)
	sc.When(`^"([^"]*)" cancels the sign-in at the identity provider$`, w.signInCancelled)
	sc.When(`^a provider response arrives with a state different from the one the browser started$`, w.signInForged)
	sc.When(`^the same provider response arrives again in that browser$`, w.repeatCallback)
	sc.When(`^(\d+) minutes pass(?: without requests)?$`, w.pass)
	sc.When(`^they make a request every (\d+) minutes$`, w.keepActive)

	sc.Then(`^a session is opened$`, w.thenSessionOpened)
	sc.Then(`^their role is "([^"]*)"$`, w.thenRole)
	sc.Then(`^no session is opened$`, w.thenNoSession)
	sc.Then(`^they are asked to try again$`, w.thenSentBack(api.Rejected))
	sc.Then(`^they are asked to sign in with their institutional account$`, w.thenSentBack(api.DomainNotAllowed))
	sc.Then(`^the session expires$`, w.thenSessionExpired)
	sc.Then(`^they have to sign in again$`, w.thenTold("Iniciá sesión"))
	sc.Then(`^the session stays open$`, w.thenSessionOpen)
}

// roundTrip serves requests in process, so a scenario opens no port and shares
// its clock with the service without synchronization.
type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// server is built on first use, once the Background has set the domain.
func (w *authenticationWorld) server() http.Handler {
	if w.handler == nil {
		auth := application.NewAuthService(
			application.AuthConfig{IdleTimeout: idleTimeout, AllowedDomain: w.domain},
			dev.NewIdentityProvider(baseURL+"/auth/callback", ""),
			w.users,
			memory.NewSessionStore(),
			func() time.Time { return w.now },
		)
		memberships := application.NewMembershipService(
			memory.NewAreaRepository(w.areas...),
			memory.NewAdministrators(),
			w.memberships,
			memory.NewDeniedAttemptLog(),
			func() string { return "M-new" },
			func() time.Time { return w.now },
		)
		server := authrest.NewAuthHandler(auth, memberships, baseURL+"/", loginPage, false)
		w.handler = api.Mount(http.NewServeMux(), api.NewServer(server, nil), slog.New(slog.DiscardHandler))
	}
	return w.handler
}

// newBrowser keeps cookies like a browser, but stops at each redirect so that
// a step can see it.
func (w *authenticationWorld) newBrowser() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	w.client = &http.Client{
		Jar: jar,
		Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			recorder := httptest.NewRecorder()
			w.server().ServeHTTP(recorder, r)
			return recorder.Result(), nil
		}),
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return nil
}

// do takes a path or a full URL, and records the result.
func (w *authenticationWorld) do(method, target, body string) error {
	if strings.HasPrefix(target, "/") {
		target = baseURL + target
	}
	req, err := http.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	if err != nil {
		return err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := w.client.Do(req)
	if err != nil {
		return err
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := res.Body.Close(); err != nil {
		return err
	}
	w.status, w.body, w.location = res.StatusCode, string(raw), res.Header.Get("Location")
	w.sessionSet = false
	for _, c := range res.Cookies() {
		if c.Name == "sigta_session" && c.Value != "" {
			w.sessionSet = true
		}
	}
	return nil
}

func (w *authenticationWorld) expectStatus(want int) error {
	if w.status != want {
		return fmt.Errorf("got status %d with %q, want %d", w.status, w.body, want)
	}
	return nil
}

func (w *authenticationWorld) add(email string) {
	w.users.Add(domain.User{ID: email, Email: email, Name: "Persona de prueba"})
}

func (w *authenticationWorld) givenDomain(allowed string) error {
	w.domain = allowed
	return nil
}

func (w *authenticationWorld) givenApplicant(email string) error {
	w.add(email)
	return nil
}

func (w *authenticationWorld) givenStaff(email, area string) error {
	w.add(email)
	w.areas = append(w.areas, domain.Area{ID: areaID(area), Name: area})
	return w.memberships.Save(context.Background(), domain.Membership{
		ID:      "M-" + email,
		UserID:  email,
		Area:    areaID(area),
		Profile: domain.ProfileStaffOperative,
	})
}

func (w *authenticationWorld) givenSignedIn(email string) error {
	if err := w.signIn(email); err != nil {
		return err
	}
	if err := w.expectStatus(http.StatusFound); err != nil {
		return err
	}
	if !w.sessionSet {
		return fmt.Errorf("%s got no session: %s", email, w.location)
	}
	return nil
}

// begin leaves for the provider and returns where it sends the browser back.
func (w *authenticationWorld) begin(target string) (string, error) {
	if err := w.newBrowser(); err != nil {
		return "", err
	}
	if err := w.do(http.MethodGet, target, ""); err != nil {
		return "", err
	}
	return w.location, w.expectStatus(http.StatusFound)
}

func (w *authenticationWorld) signIn(email string) error {
	callback, err := w.begin("/auth/login?login_hint=" + url.QueryEscape(email))
	if err != nil {
		return err
	}
	w.callback = callback
	return w.do(http.MethodGet, callback, "")
}

// signInRejected relies on the development provider, which rejects a login
// that names no email.
func (w *authenticationWorld) signInRejected() error {
	callback, err := w.begin("/auth/login")
	if err != nil {
		return err
	}
	return w.do(http.MethodGet, callback, "")
}

func (w *authenticationWorld) signInCancelled(email string) error {
	callback, err := w.begin("/auth/login?login_hint=" + url.QueryEscape(email))
	if err != nil {
		return err
	}
	cancelled, err := url.Parse(callback)
	if err != nil {
		return err
	}
	query := cancelled.Query()
	query.Del("code")
	query.Set("error", "access_denied")
	cancelled.RawQuery = query.Encode()
	return w.do(http.MethodGet, cancelled.String(), "")
}

func (w *authenticationWorld) signInForged() error {
	callback, err := w.begin("/auth/login?login_hint=" + url.QueryEscape("applicant@fi.uba.ar"))
	if err != nil {
		return err
	}
	forged, err := url.Parse(callback)
	if err != nil {
		return err
	}
	query := forged.Query()
	query.Set("state", "forged")
	forged.RawQuery = query.Encode()
	return w.do(http.MethodGet, forged.String(), "")
}

func (w *authenticationWorld) repeatCallback() error {
	return w.do(http.MethodGet, w.callback, "")
}

func (w *authenticationWorld) pass(minutes int) error {
	w.now = w.now.Add(time.Duration(minutes) * time.Minute)
	return nil
}

func (w *authenticationWorld) keepActive(minutes int) error {
	for range 3 {
		if err := w.pass(minutes); err != nil {
			return err
		}
		if err := w.thenSessionOpen(); err != nil {
			return err
		}
	}
	return nil
}

func (w *authenticationWorld) currentUser() error {
	return w.do(http.MethodGet, "/me", "")
}

func (w *authenticationWorld) thenSessionOpened() error {
	if err := w.expectStatus(http.StatusFound); err != nil {
		return err
	}
	if !w.sessionSet {
		return errors.New("the response set no session")
	}
	return w.thenSessionOpen()
}

func (w *authenticationWorld) thenSessionOpen() error {
	if err := w.currentUser(); err != nil {
		return err
	}
	return w.expectStatus(http.StatusOK)
}

func (w *authenticationWorld) thenRole(want string) error {
	var user api.User
	if err := json.Unmarshal([]byte(w.body), &user); err != nil {
		return fmt.Errorf("%q: %w", w.body, err)
	}
	if string(user.Role) != want {
		return fmt.Errorf("the role is %q, want %q", user.Role, want)
	}
	return nil
}

// thenNoSession expects the browser back at the login page without a session.
func (w *authenticationWorld) thenNoSession() error {
	if w.sessionSet {
		return errors.New("the response set a session")
	}
	_, err := w.loginError()
	return err
}

func (w *authenticationWorld) loginError() (api.LoginError, error) {
	location, err := url.Parse(w.location)
	if err != nil {
		return "", err
	}
	if location.Scheme+"://"+location.Host+location.Path != loginPage {
		return "", fmt.Errorf("redirected to %s, want the login page", w.location)
	}
	return api.LoginError(location.Query().Get("error")), nil
}

func (w *authenticationWorld) thenSentBack(want api.LoginError) func() error {
	return func() error {
		got, err := w.loginError()
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("sent back with %q, want %q", got, want)
		}
		return nil
	}
}

func (w *authenticationWorld) thenTold(message string) func() error {
	return func() error {
		var got api.Error
		if err := json.Unmarshal([]byte(w.body), &got); err != nil {
			return fmt.Errorf("%q: %w", w.body, err)
		}
		if !strings.Contains(got.Message, message) {
			return fmt.Errorf("got %q, want it to say %q", got.Message, message)
		}
		return nil
	}
}

func (w *authenticationWorld) thenSessionExpired() error {
	if err := w.currentUser(); err != nil {
		return err
	}
	return w.expectStatus(http.StatusUnauthorized)
}
