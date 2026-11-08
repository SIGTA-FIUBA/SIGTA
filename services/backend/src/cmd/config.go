package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"backend/src/auth/domain"
)

const (
	envDevelopment = "development"
	envProduction  = "production"

	identityDev    = "dev"
	identityGoogle = "google"
)

type config struct {
	env string
	// identityProvider is dev or google, chosen apart from env so Google can
	// be tried locally.
	identityProvider string
	publicURL        string
	appURL           string
	allowedDomain    string
	googleClient     struct{ id, secret string }
	usersFile        string
}

func (c config) development() bool { return c.env == envDevelopment }

func (c config) callbackURL() string { return c.publicURL + "/auth/callback" }

func (c config) loginURL() string { return c.appURL + "login" }

// secureCookies is off only for local development over plain HTTP.
func (c config) secureCookies() bool { return strings.HasPrefix(c.publicURL, "https://") }

// loadConfig reads the environment and refuses a configuration that would
// start half working or wire a development adapter outside development.
func loadConfig(getenv func(string) string) (config, error) {
	c := config{
		env:              getenv("SIGTA_ENV"),
		identityProvider: getenv("SIGTA_IDENTITY_PROVIDER"),
		publicURL:        getenv("SIGTA_PUBLIC_URL"),
		appURL:           getenv("SIGTA_APP_URL"),
		allowedDomain:    getenv("SIGTA_ALLOWED_DOMAIN"),
		usersFile:        getenv("SIGTA_USERS_FILE"),
	}
	c.googleClient.id = getenv("SIGTA_GOOGLE_CLIENT_ID")
	c.googleClient.secret = getenv("SIGTA_GOOGLE_CLIENT_SECRET")

	if c.env != envDevelopment && c.env != envProduction {
		return config{}, fmt.Errorf("SIGTA_ENV must be %s or %s, not %q", envDevelopment, envProduction, c.env)
	}
	if c.development() {
		c.identityProvider = or(c.identityProvider, identityDev)
		c.publicURL = or(c.publicURL, "http://localhost:8080")
		c.appURL = or(c.appURL, "http://localhost:5173/")
		c.allowedDomain = or(c.allowedDomain, "fi.uba.ar")
	}
	c.identityProvider = or(c.identityProvider, identityGoogle)

	var missing []string
	for _, v := range []struct{ name, value string }{
		{"SIGTA_PUBLIC_URL", c.publicURL},
		{"SIGTA_APP_URL", c.appURL},
		{"SIGTA_ALLOWED_DOMAIN", c.allowedDomain},
	} {
		if v.value == "" {
			missing = append(missing, v.name)
		}
	}
	if !c.development() && c.usersFile == "" {
		missing = append(missing, "SIGTA_USERS_FILE")
	}

	switch c.identityProvider {
	case identityDev:
		if !c.development() {
			return config{}, errors.New("the dev identity provider logs anyone in; it only runs with SIGTA_ENV=development")
		}
	case identityGoogle:
		if c.googleClient.id == "" {
			missing = append(missing, "SIGTA_GOOGLE_CLIENT_ID")
		}
		if c.googleClient.secret == "" {
			missing = append(missing, "SIGTA_GOOGLE_CLIENT_SECRET")
		}
		if c.publicURL != "" && !secureOrLocal(c.publicURL) {
			return config{}, fmt.Errorf("SIGTA_PUBLIC_URL must be https or localhost for Google, not %s", c.publicURL)
		}
	default:
		return config{}, fmt.Errorf("SIGTA_IDENTITY_PROVIDER must be %s or %s, not %q", identityDev, identityGoogle, c.identityProvider)
	}

	if len(missing) > 0 {
		return config{}, fmt.Errorf("missing %s", strings.Join(missing, ", "))
	}
	if !strings.HasSuffix(c.appURL, "/") {
		return config{}, fmt.Errorf("SIGTA_APP_URL must end in a slash, not %s", c.appURL)
	}
	return c, nil
}

func secureOrLocal(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme == "https" || (u.Scheme == "http" && u.Hostname() == "localhost")
}

func or(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// userEntry is one person in SIGTA_USERS_FILE. The file holds real names and
// emails, so it lives with the deployment's secrets and never in the
// repository. Someone without memberships is an applicant.
type userEntry struct {
	ID            string            `json:"id"`
	Email         string            `json:"email"`
	Name          string            `json:"name"`
	Administrator bool              `json:"administrator,omitempty"`
	Memberships   []membershipEntry `json:"memberships,omitempty"`
}

type membershipEntry struct {
	Area    string  `json:"area"`
	Profile string  `json:"profile"`
	Scope   []int64 `json:"scope,omitempty"`
}

// people is what SIGTA_USERS_FILE loads, until users and memberships are
// stored in PostgreSQL and administered from the interface.
type people struct {
	users          []domain.User
	administrators []string
	memberships    []domain.Membership
}

// loadUsers reads the people who can log in. It rejects the whole file on any
// invalid entry, so a typo does not silently leave someone out.
func loadUsers(path, allowedDomain string, areas []domain.Area) (people, error) {
	file, err := os.Open(path) //nolint:gosec // G304: the path comes from configuration.
	if err != nil {
		return people{}, err
	}
	defer func() { _ = file.Close() }()
	var entries []userEntry
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&entries); err != nil {
		return people{}, fmt.Errorf("%s: %w", path, err)
	}

	known := map[string]bool{}
	for _, a := range areas {
		known[a.ID] = true
	}
	ids, emails := map[string]bool{}, map[string]bool{}
	var problems []error
	var loaded people
	for i, e := range entries {
		user := domain.User{ID: e.ID, Email: domain.NormalizeEmail(e.Email), Name: strings.TrimSpace(e.Name)}
		switch {
		case user.ID == "" || ids[user.ID]:
			problems = append(problems, fmt.Errorf("entry %d: empty or repeated id %q", i, user.ID))
		case !domain.InDomain(user.Email, allowedDomain) || emails[user.Email]:
			problems = append(problems, fmt.Errorf("entry %d: email outside %s or repeated", i, allowedDomain))
		case user.Name == "":
			problems = append(problems, fmt.Errorf("entry %d: empty name", i))
		}
		ids[user.ID], emails[user.Email] = true, true
		loaded.users = append(loaded.users, user)
		if e.Administrator {
			loaded.administrators = append(loaded.administrators, user.ID)
		}
		for j, m := range e.Memberships {
			membership := domain.Membership{
				ID:        fmt.Sprintf("%s-%d", user.ID, j+1),
				UserID:    user.ID,
				Area:      m.Area,
				Profile:   domain.Profile(m.Profile),
				Scope:     m.Scope,
				GrantedBy: "SIGTA_USERS_FILE",
			}
			switch {
			case !known[membership.Area]:
				problems = append(problems, fmt.Errorf("entry %d: membership %d: unknown area %q", i, j+1, m.Area))
			case !membership.Profile.Known():
				problems = append(problems, fmt.Errorf("entry %d: membership %d: unknown profile %q", i, j+1, m.Profile))
			case slices.ContainsFunc(loaded.memberships, membership.Overlaps):
				problems = append(problems, fmt.Errorf("entry %d: membership %d: repeats area %q with the same scope", i, j+1, m.Area))
			}
			loaded.memberships = append(loaded.memberships, membership)
		}
	}
	if len(problems) > 0 {
		return people{}, fmt.Errorf("%s: %w", path, errors.Join(problems...))
	}
	return loaded, nil
}
