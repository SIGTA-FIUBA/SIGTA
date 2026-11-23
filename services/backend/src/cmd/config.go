package main

import (
	"cmp"
	"fmt"
	"net/url"
	"strings"
)

const (
	envDevelopment = "development"
	envProduction  = "production"
)

type config struct {
	env           string
	publicURL     string
	appURL        string
	allowedDomain string
	googleClient  struct{ id, secret string }
	// databaseURL connects as the application role. Without it, development
	// keeps everything in memory.
	databaseURL string
	// documentsDir keeps the files of the procedures. Without it, development
	// uses a directory under the system's temporary one.
	documentsDir string
}

func (c config) development() bool { return c.env == envDevelopment }

func (c config) callbackURL() string { return c.publicURL + "/auth/callback" }

func (c config) loginURL() string { return c.appURL + "login" }

func (c config) secureCookies() bool { return strings.HasPrefix(c.publicURL, "https://") }

// loadConfig reads the environment and refuses a configuration that would
// start half working. Development only has defaults for a local start.
func loadConfig(getenv func(string) string) (config, error) {
	c := config{
		env:           getenv("SIGTA_ENV"),
		publicURL:     getenv("SIGTA_PUBLIC_URL"),
		appURL:        getenv("SIGTA_APP_URL"),
		allowedDomain: getenv("SIGTA_ALLOWED_DOMAIN"),
		databaseURL:   getenv("SIGTA_DATABASE_URL"),
		documentsDir:  getenv("SIGTA_DOCUMENTS_DIR"),
	}
	c.googleClient.id = getenv("SIGTA_GOOGLE_CLIENT_ID")
	c.googleClient.secret = getenv("SIGTA_GOOGLE_CLIENT_SECRET")

	if c.env != envDevelopment && c.env != envProduction {
		return config{}, fmt.Errorf("SIGTA_ENV must be %s or %s, not %q", envDevelopment, envProduction, c.env)
	}
	if c.development() {
		c.publicURL = cmp.Or(c.publicURL, "http://localhost:8080")
		c.appURL = cmp.Or(c.appURL, "http://localhost:5173/")
		c.allowedDomain = cmp.Or(c.allowedDomain, "fi.uba.ar")
	}

	var missing []string
	for _, v := range []struct{ name, value string }{
		{"SIGTA_PUBLIC_URL", c.publicURL},
		{"SIGTA_APP_URL", c.appURL},
		{"SIGTA_ALLOWED_DOMAIN", c.allowedDomain},
		{"SIGTA_GOOGLE_CLIENT_ID", c.googleClient.id},
		{"SIGTA_GOOGLE_CLIENT_SECRET", c.googleClient.secret},
	} {
		if v.value == "" {
			missing = append(missing, v.name)
		}
	}
	if !c.development() && c.databaseURL == "" {
		missing = append(missing, "SIGTA_DATABASE_URL")
	}
	if (!c.development() || c.databaseURL != "") && c.documentsDir == "" {
		missing = append(missing, "SIGTA_DOCUMENTS_DIR")
	}

	if len(missing) > 0 {
		return config{}, fmt.Errorf("missing %s", strings.Join(missing, ", "))
	}
	if !secureOrLocal(c.publicURL) {
		return config{}, fmt.Errorf("SIGTA_PUBLIC_URL must be https or localhost for Google, not %s", c.publicURL)
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
