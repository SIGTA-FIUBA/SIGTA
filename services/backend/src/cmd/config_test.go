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

func development() map[string]string {
	return map[string]string{
		"SIGTA_ENV":                  "development",
		"SIGTA_GOOGLE_CLIENT_ID":     "client",
		"SIGTA_GOOGLE_CLIENT_SECRET": "secret",
	}
}

func TestDevelopmentNeedsOnlyItsEnvAndTheGoogleClient(t *testing.T) {
	cfg, err := loadConfig(env(development()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.allowedDomain != "fi.uba.ar" || cfg.databaseURL != "" || cfg.documentsDir != "" {
		t.Fatalf("domain %q, database %q, documents %q; want %q, none and none", cfg.allowedDomain, cfg.databaseURL, cfg.documentsDir, "fi.uba.ar")
	}
	if cfg.callbackURL() != "http://localhost:8080/auth/callback" || cfg.loginURL() != "http://localhost:5173/login" {
		t.Fatalf("callback %s, login %s; want http://localhost:8080/auth/callback and http://localhost:5173/login", cfg.callbackURL(), cfg.loginURL())
	}
	if cfg.secureCookies() {
		t.Error("secureCookies() = true over plain HTTP, want false")
	}
}

func TestDevelopmentNeedsTheGoogleClient(t *testing.T) {
	_, err := loadConfig(env(map[string]string{"SIGTA_ENV": "development"}))
	if err == nil || !strings.Contains(err.Error(), "SIGTA_GOOGLE_CLIENT_ID") || !strings.Contains(err.Error(), "SIGTA_GOOGLE_CLIENT_SECRET") {
		t.Fatalf("loadConfig() = %v, want an error about the Google client", err)
	}
}

func TestProduction(t *testing.T) {
	cfg, err := loadConfig(env(production()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.development() || cfg.callbackURL() != "https://sigta.test/api/auth/callback" {
		t.Fatalf("development %t, callback %s; want false and https://sigta.test/api/auth/callback", cfg.development(), cfg.callbackURL())
	}
	if !cfg.secureCookies() {
		t.Error("secureCookies() = false over HTTPS, want true")
	}
}

// A temporary directory of documents does not outlive the server, and their
// records in the database would name files that are gone.
func TestDevelopmentWithADatabaseNeedsTheDocumentsDirectory(t *testing.T) {
	values := development()
	values["SIGTA_DATABASE_URL"] = "postgres://backend@db/sigta"
	_, err := loadConfig(env(values))
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
		{"no client ID", func(e map[string]string) { delete(e, "SIGTA_GOOGLE_CLIENT_ID") }, "SIGTA_GOOGLE_CLIENT_ID"},
		{"no client secret", func(e map[string]string) { delete(e, "SIGTA_GOOGLE_CLIENT_SECRET") }, "SIGTA_GOOGLE_CLIENT_SECRET"},
		{"no database", func(e map[string]string) { delete(e, "SIGTA_DATABASE_URL") }, "SIGTA_DATABASE_URL"},
		{"no documents directory", func(e map[string]string) { delete(e, "SIGTA_DOCUMENTS_DIR") }, "SIGTA_DOCUMENTS_DIR"},
		{"no public URL", func(e map[string]string) { delete(e, "SIGTA_PUBLIC_URL") }, "SIGTA_PUBLIC_URL"},
		{"no app URL", func(e map[string]string) { delete(e, "SIGTA_APP_URL") }, "SIGTA_APP_URL"},
		{"no domain", func(e map[string]string) { delete(e, "SIGTA_ALLOWED_DOMAIN") }, "SIGTA_ALLOWED_DOMAIN"},
		{"plain HTTP", func(e map[string]string) { e["SIGTA_PUBLIC_URL"] = "http://sigta.test/api" }, "https"},
		{"development over plain HTTP", func(e map[string]string) {
			e["SIGTA_ENV"] = "development"
			e["SIGTA_PUBLIC_URL"] = "http://sigta.test/api"
		}, "https"},
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
