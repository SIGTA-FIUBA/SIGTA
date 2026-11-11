package main

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"backend/src/auth/domain"
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
		"SIGTA_USERS_FILE":           "/run/secrets/users.json",
	}
}

func TestDevelopmentNeedsOnlyItsEnv(t *testing.T) {
	cfg, err := loadConfig(env(map[string]string{"SIGTA_ENV": "development"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.identityProvider != identityDev || cfg.allowedDomain != "fi.uba.ar" || cfg.usersFile != "" {
		t.Fatalf("provider %q, domain %q, users file %q; want %q, %q and none", cfg.identityProvider, cfg.allowedDomain, cfg.usersFile, identityDev, "fi.uba.ar")
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
		{"no users file", func(e map[string]string) { delete(e, "SIGTA_USERS_FILE") }, "SIGTA_USERS_FILE"},
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

func usersFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "users.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

var testAreas = []domain.Area{{ID: "office", Name: "Secretaría", Permissions: []domain.Permission{{ID: "review", Name: "Revisar"}, {ID: "approve", Name: "Aprobar"}}}}

func TestLoadUsers(t *testing.T) {
	path := usersFile(t, `[
		{"id": "u1", "email": " Operativo@FI.UBA.AR ", "name": "Operativo de prueba", "memberships": [{"area": "office", "permissions": ["review"], "scope": [2]}]},
		{"id": "u2", "email": "titular@fi.uba.ar", "name": "Titular de prueba"},
		{"id": "u3", "email": "admin@fi.uba.ar", "name": "Administración de prueba", "administrator": true}
	]`)

	loaded, err := loadUsers(path, "fi.uba.ar", testAreas)
	if err != nil {
		t.Fatal(err)
	}
	want := domain.User{ID: "u1", Email: "operativo@fi.uba.ar", Name: "Operativo de prueba"}
	if len(loaded.users) != 3 || loaded.users[0] != want {
		t.Fatalf("users = %+v, want 3 users starting with %+v", loaded.users, want)
	}
	if !slices.Equal(loaded.administrators, []string{"u3"}) {
		t.Fatalf("administrators = %v, want [u3]", loaded.administrators)
	}
	wantMembership := domain.Membership{ID: "u1-1", UserID: "u1", Area: "office", Permissions: []string{"review"}, Scope: []int64{2}, GrantedBy: "SIGTA_USERS_FILE"}
	if len(loaded.memberships) != 1 || !reflect.DeepEqual(loaded.memberships[0], wantMembership) {
		t.Fatalf("memberships = %+v, want only %+v", loaded.memberships, wantMembership)
	}
}

func TestLoadUsersRejectsTheWholeFile(t *testing.T) {
	tests := map[string]string{
		"repeated id":         `[{"id":"u1","email":"a@fi.uba.ar","name":"A"},{"id":"u1","email":"b@fi.uba.ar","name":"B"}]`,
		"empty id":            `[{"id":"","email":"a@fi.uba.ar","name":"A"}]`,
		"repeated email":      `[{"id":"u1","email":"a@fi.uba.ar","name":"A"},{"id":"u2","email":"A@fi.uba.ar","name":"B"}]`,
		"email outside":       `[{"id":"u1","email":"a@fi.uba.ar.evil.com","name":"A"}]`,
		"empty name":          `[{"id":"u1","email":"a@fi.uba.ar","name":" "}]`,
		"old role key":        `[{"id":"u1","email":"a@fi.uba.ar","name":"A","role":"staff_operative"}]`,
		"unknown area":        `[{"id":"u1","email":"a@fi.uba.ar","name":"A","memberships":[{"area":"dean","permissions":["review"]}]}]`,
		"unknown permission":  `[{"id":"u1","email":"a@fi.uba.ar","name":"A","memberships":[{"area":"office","permissions":["sign"]}]}]`,
		"no permissions":      `[{"id":"u1","email":"a@fi.uba.ar","name":"A","memberships":[{"area":"office","permissions":[]}]}]`,
		"old profile key":     `[{"id":"u1","email":"a@fi.uba.ar","name":"A","memberships":[{"area":"office","profile":"staff_operative"}]}]`,
		"repeated membership": `[{"id":"u1","email":"a@fi.uba.ar","name":"A","memberships":[{"area":"office","permissions":["review"]},{"area":"office","permissions":["approve"]}]}]`,
		"unknown key":         `[{"id":"u1","email":"a@fi.uba.ar","name":"A","password":"x"}]`,
		"not a list of users": `{"id":"u1"}`,
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			if loaded, err := loadUsers(usersFile(t, content), "fi.uba.ar", testAreas); err == nil {
				t.Fatalf("loaded %+v", loaded)
			}
		})
	}
}

func TestLoadUsersNeedsTheFile(t *testing.T) {
	if _, err := loadUsers(filepath.Join(t.TempDir(), "missing.json"), "fi.uba.ar", testAreas); err == nil {
		t.Fatal("loaded a missing file")
	}
}
