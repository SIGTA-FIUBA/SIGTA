package main

import (
	"strings"
	"testing"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func production() map[string]string {
	return map[string]string{
		"SIGTA_ENV":                  "production",
		"SIGTA_PUBLIC_URL":           "https://sigta.test/api",
		"SIGTA_APP_URL":              "https://sigta.test/",
		"SIGTA_ALLOWED_DOMAIN":       "fi.uba.ar",
		"SIGTA_GOOGLE_CLIENT_ID":     "client",
		"SIGTA_GOOGLE_CLIENT_SECRET": "secret",
		"SIGTA_DATABASE_URL":         "postgres://backend@db/sigta",
		"SIGTA_DOCUMENTS_DIR":        "/var/lib/sigta/documents",
	}
}

func TestDevelopmentNeedsOnlyItsEnv(t *testing.T) {
	cfg, err := loadConfig(env(map[string]string{"SIGTA_ENV": "development"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.identityProvider != identityDev || cfg.allowedDomain != "fi.uba.ar" || cfg.databaseURL != "" || cfg.documentsDir != "" {
		t.Fatalf("provider %q, domain %q, database %q, documents %q; want %q, %q, none and none", cfg.identityProvider, cfg.allowedDomain, cfg.databaseURL, cfg.documentsDir, identityDev, "fi.uba.ar")
	}
	if cfg.callbackURL() != "http://localhost:8080/auth/callback" || cfg.loginURL() != "http://localhost:5173/login" {
		t.Fatalf("callback %s, login %s; want http://localhost:8080/auth/callback and http://localhost:5173/login", cfg.callbackURL(), cfg.loginURL())
	}
}

func TestProductionUsesGoogle(t *testing.T) {
	cfg, err := loadConfig(env(production()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.identityProvider != identityGoogle || cfg.callbackURL() != "https://sigta.test/api/auth/callback" {
		t.Fatalf("provider %q, callback %s; want %q and https://sigta.test/api/auth/callback", cfg.identityProvider, cfg.callbackURL(), identityGoogle)
	}
	if !cfg.secureCookies() {
		t.Error("secureCookies() = false over HTTPS, want true")
	}
}

func TestGoogleCanBeTriedInDevelopment(t *testing.T) {
	cfg, err := loadConfig(env(map[string]string{
		"SIGTA_ENV":                  "development",
		"SIGTA_IDENTITY_PROVIDER":    "google",
		"SIGTA_PUBLIC_URL":           "http://localhost:8000/api",
		"SIGTA_GOOGLE_CLIENT_ID":     "client",
		"SIGTA_GOOGLE_CLIENT_SECRET": "secret",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.identityProvider != identityGoogle {
		t.Fatalf("provider %q, want %q", cfg.identityProvider, identityGoogle)
	}
	if cfg.secureCookies() {
		t.Error("secureCookies() = true over plain HTTP, want false")
	}
}

// A temporary directory of documents does not outlive the server, and their
// records in the database would name files that are gone.
func TestDevelopmentWithADatabaseNeedsTheDocumentsDirectory(t *testing.T) {
	_, err := loadConfig(env(map[string]string{"SIGTA_ENV": "development", "SIGTA_DATABASE_URL": "postgres://backend@db/sigta"}))
	if err == nil || !strings.Contains(err.Error(), "SIGTA_DOCUMENTS_DIR") {
		t.Fatalf("loadConfig() = %v, want an error about SIGTA_DOCUMENTS_DIR", err)
	}
}

func TestRejectedConfigs(t *testing.T) {
	tests := []struct {
		name   string
		change func(map[string]string)
		says   string
	}{
		{"no env", func(e map[string]string) { delete(e, "SIGTA_ENV") }, "SIGTA_ENV"},
		{"unknown env", func(e map[string]string) { e["SIGTA_ENV"] = "staging" }, "SIGTA_ENV"},
		{"dev provider outside development", func(e map[string]string) { e["SIGTA_IDENTITY_PROVIDER"] = "dev" }, "development"},
		{"unknown provider", func(e map[string]string) { e["SIGTA_IDENTITY_PROVIDER"] = "github" }, "SIGTA_IDENTITY_PROVIDER"},
		{"no client ID", func(e map[string]string) { delete(e, "SIGTA_GOOGLE_CLIENT_ID") }, "SIGTA_GOOGLE_CLIENT_ID"},
		{"no client secret", func(e map[string]string) { delete(e, "SIGTA_GOOGLE_CLIENT_SECRET") }, "SIGTA_GOOGLE_CLIENT_SECRET"},
		{"no database", func(e map[string]string) { delete(e, "SIGTA_DATABASE_URL") }, "SIGTA_DATABASE_URL"},
		{"no documents directory", func(e map[string]string) { delete(e, "SIGTA_DOCUMENTS_DIR") }, "SIGTA_DOCUMENTS_DIR"},
		{"no public URL", func(e map[string]string) { delete(e, "SIGTA_PUBLIC_URL") }, "SIGTA_PUBLIC_URL"},
		{"no app URL", func(e map[string]string) { delete(e, "SIGTA_APP_URL") }, "SIGTA_APP_URL"},
		{"no domain", func(e map[string]string) { delete(e, "SIGTA_ALLOWED_DOMAIN") }, "SIGTA_ALLOWED_DOMAIN"},
		{"plain HTTP", func(e map[string]string) { e["SIGTA_PUBLIC_URL"] = "http://sigta.test/api" }, "https"},
		{"app URL without slash", func(e map[string]string) { e["SIGTA_APP_URL"] = "https://sigta.test" }, "slash"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := production()
			tt.change(values)
			_, err := loadConfig(env(values))
			if err == nil {
				t.Fatal("the config was accepted")
			}
			if !strings.Contains(err.Error(), tt.says) {
				t.Fatalf("error %q does not mention %q", err, tt.says)
			}
		})
	}
}
