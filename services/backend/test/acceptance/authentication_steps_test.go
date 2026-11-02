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

	"backend/src/adapters/dev"
	"backend/src/adapters/memory"
	"backend/src/adapters/rest"
	"backend/src/application"
	"backend/src/domain"
)

const (
	baseURL     = "http://sigta.test"
	linkPrefix  = "#token="
	idleTimeout = 30 * time.Minute
	linkTTL     = 15 * time.Minute
)

// authenticationWorld is the state of one scenario of US06. Each scenario goes
// through the REST adapter, which it calls in process.
type authenticationWorld struct {
	domain  string
	users   *memory.UserRepository
	now     time.Time
	links   map[string]string
	handler http.Handler

	client *http.Client
	email  string

	// The outcome of the last request.
	status     int
	body       string
	location   string
	sessionSet bool
	// callback is the last provider response the browser arrived with.
	callback string
}

func registerAuthenticationSteps(sc *godog.ScenarioContext) {
	w := &authenticationWorld{
		users: memory.NewUserRepository(),
		now:   time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC),
		links: map[string]string{},
	}

	sc.Given(`^only accounts from the "([^"]*)" domain are allowed$`, w.givenDomain)
	sc.Given(`^SIGTA has "([^"]*)" registered as an applicant$`, w.givenApplicant)
	sc.Given(`^SIGTA has "([^"]*)" registered as staff with the "([^"]*)" profile$`, w.givenStaff)
	sc.Given(`^"([^"]*)" completed a sign-in$`, w.givenSignedIn)
	sc.Given(`^"([^"]*)" has an open session$`, w.givenSignedIn)
	sc.Given(`^"([^"]*)" already used their sign-in link$`, w.givenUsedLink)
	sc.Given(`^their profile changed to "([^"]*)"$`, w.givenProfileChanged)

	sc.Step(`^"([^"]*)" (?:asks|asked) for a sign-in link$`, w.askForLink)
	sc.When(`^"([^"]*)" signs in through the identity provider$`, w.signIn)
	sc.When(`^the identity provider rejects a sign-in$`, w.signInRejected)
	sc.When(`^a provider response arrives with a state different from the one the browser started$`, w.signInForged)
	sc.When(`^the same provider response arrives again in that browser$`, w.repeatCallback)
	sc.When(`^(\d+) minutes pass(?: without requests)?$`, w.pass)
	sc.When(`^they make a request every (\d+) minutes$`, w.keepActive)
	sc.When(`^they open (?:it again|the link)$`, w.openLink)

	sc.Then(`^a session is opened$`, w.thenSessionOpened)
	sc.Then(`^their profile is "([^"]*)"$`, w.thenProfile)
	sc.Then(`^no session is opened$`, w.thenNoSession)
	sc.Then(`^they are told SIGTA has no account for that identity$`, w.thenTold("No hay una cuenta de SIGTA"))
	sc.Then(`^they are asked to try again$`, w.thenTold("Volvé a intentarlo"))
	sc.Then(`^they are asked to sign in with their institutional account$`, w.thenTold("cuenta institucional"))
	sc.Then(`^they are asked to use their institutional email$`, w.thenTold("correo institucional"))
	sc.Then(`^the session expires$`, w.thenSessionExpired)
	sc.Then(`^they have to sign in again$`, w.thenTold("Iniciá sesión"))
	sc.Then(`^the session stays open$`, w.thenSessionOpen)
	sc.Then(`^they receive a single-use link by email$`, w.thenLinkReceived)
	sc.Then(`^opening it opens a session$`, w.thenLinkOpensSession)
	sc.Then(`^the link no longer opens a session$`, w.thenLinkRejected)
	sc.Then(`^the response is the same as for an existing account$`, w.thenSameResponse)
	sc.Then(`^no email is sent$`, w.thenNoEmail)
}

// roundTrip serves requests in process, so a scenario opens no port and shares
// its clock with the service without synchronization.
type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func (w *authenticationWorld) SendLoginLink(_ context.Context, to, link string) error {
	w.links[to] = strings.TrimPrefix(link, linkPrefix)
	return nil
}

// server is built on first use, once the Background has set the domain.
func (w *authenticationWorld) server() http.Handler {
	if w.handler == nil {
		auth := application.NewAuthService(
			application.AuthConfig{IdleTimeout: idleTimeout, LoginLinkTTL: linkTTL, AllowedDomain: w.domain, LoginLinkURL: linkPrefix},
			dev.NewIdentityProvider(baseURL+"/auth/callback"),
			w.users,
			memory.NewSessionStore(),
			memory.NewLoginLinkStore(),
			w,
			func() time.Time { return w.now },
		)
		api := rest.NewServer(rest.NewAuthHandler(auth, baseURL+"/", false))
		w.handler = rest.Mount(http.NewServeMux(), api, slog.New(slog.DiscardHandler))
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

// do takes a path or a full URL, and records the outcome.
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

func (w *authenticationWorld) add(email string, role domain.Role) {
	w.users.Add(domain.User{ID: email, Email: email, Name: "Persona de prueba", Role: role})
}

func (w *authenticationWorld) givenDomain(domain string) error {
	w.domain = domain
	return nil
}

func (w *authenticationWorld) givenApplicant(email string) error {
	w.add(email, domain.RoleApplicant)
	return nil
}

func (w *authenticationWorld) givenStaff(email, profile string) error {
	w.add(email, domain.Role(profile))
	return nil
}

func (w *authenticationWorld) givenSignedIn(email string) error {
	if err := w.signIn(email); err != nil {
		return err
	}
	return w.expectStatus(http.StatusFound)
}

func (w *authenticationWorld) givenUsedLink(email string) error {
	if err := w.askForLink(email); err != nil {
		return err
	}
	if err := w.openLink(); err != nil {
		return err
	}
	return w.expectStatus(http.StatusNoContent)
}

func (w *authenticationWorld) givenProfileChanged(profile string) error {
	user, err := w.users.FindByEmail(context.Background(), w.email)
	if err != nil {
		return err
	}
	user.Role = domain.Role(profile)
	w.users.Add(user)
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
	w.email = email
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

func (w *authenticationWorld) requestLink(email string) error {
	return w.do(http.MethodPost, "/auth/login-links", fmt.Sprintf(`{"email":%q}`, email))
}

func (w *authenticationWorld) askForLink(email string) error {
	w.email = email
	if err := w.newBrowser(); err != nil {
		return err
	}
	return w.requestLink(email)
}

func (w *authenticationWorld) openLink() error {
	return w.do(http.MethodPost, "/auth/login-links/redeem", fmt.Sprintf(`{"token":%q}`, w.links[w.email]))
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

func (w *authenticationWorld) thenProfile(want string) error {
	var user rest.User
	if err := json.Unmarshal([]byte(w.body), &user); err != nil {
		return fmt.Errorf("%q: %w", w.body, err)
	}
	if string(user.Role) != want {
		return fmt.Errorf("the profile is %q, want %q", user.Role, want)
	}
	return nil
}

func (w *authenticationWorld) thenNoSession() error {
	if w.sessionSet {
		return errors.New("the response set a session")
	}
	return w.expectStatus(http.StatusUnauthorized)
}

func (w *authenticationWorld) thenTold(message string) func() error {
	return func() error {
		var got rest.Error
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

func (w *authenticationWorld) thenLinkReceived() error {
	if err := w.expectStatus(http.StatusAccepted); err != nil {
		return err
	}
	if w.links[w.email] == "" {
		return fmt.Errorf("no link sent to %s", w.email)
	}
	return nil
}

func (w *authenticationWorld) thenLinkOpensSession() error {
	if err := w.openLink(); err != nil {
		return err
	}
	if err := w.expectStatus(http.StatusNoContent); err != nil {
		return err
	}
	return w.thenSessionOpen()
}

func (w *authenticationWorld) thenLinkRejected() error {
	if err := w.openLink(); err != nil {
		return err
	}
	return w.thenNoSession()
}

// thenSameResponse asks again for the applicant of the Background, who does
// have an account.
func (w *authenticationWorld) thenSameResponse() error {
	status, body := w.status, w.body
	if err := w.requestLink("applicant@fi.uba.ar"); err != nil {
		return err
	}
	if w.status != status || w.body != body {
		return fmt.Errorf("got %d %q for an unknown account and %d %q for an existing one", status, body, w.status, w.body)
	}
	return nil
}

func (w *authenticationWorld) thenNoEmail() error {
	if link, sent := w.links[w.email]; sent {
		return fmt.Errorf("sent %q to %s", link, w.email)
	}
	return nil
}
