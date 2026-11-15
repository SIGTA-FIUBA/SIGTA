package main

import (
	"errors"
	"strings"
	"testing"
)

func TestWithSecretFilesReadsTheFile(t *testing.T) {
	getenv, err := withSecretFiles(env(map[string]string{
		"SIGTA_DATABASE_URL_FILE": "/run/files/database_url",
		"SIGTA_ENV":               "production",
	}), readFrom(map[string]string{"/run/files/database_url": "postgres://backend@db/sigta\n"}))
	if err != nil {
		t.Fatal(err)
	}
	if got := getenv("SIGTA_DATABASE_URL"); got != "postgres://backend@db/sigta" {
		t.Errorf("SIGTA_DATABASE_URL = %q, want the file without its trailing newline", got)
	}
	if got := getenv("SIGTA_ENV"); got != "production" {
		t.Errorf("SIGTA_ENV = %q, want production: other variables pass through", got)
	}
}

func TestWithSecretFilesKeepsAVariableWithoutFile(t *testing.T) {
	getenv, err := withSecretFiles(env(map[string]string{"SIGTA_GOOGLE_CLIENT_SECRET": "secret"}), readFrom(nil))
	if err != nil {
		t.Fatal(err)
	}
	if got := getenv("SIGTA_GOOGLE_CLIENT_SECRET"); got != "secret" {
		t.Errorf("SIGTA_GOOGLE_CLIENT_SECRET = %q, want secret", got)
	}
}

func TestWithSecretFilesRejects(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			"both ways",
			map[string]string{"SIGTA_DATABASE_URL": "postgres://db/sigta", "SIGTA_DATABASE_URL_FILE": "/run/files/database_url"},
			"both",
		},
		{"missing file", map[string]string{"SIGTA_DATABASE_URL_FILE": "/run/files/missing"}, "SIGTA_DATABASE_URL_FILE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := withSecretFiles(env(tc.env), readFrom(map[string]string{"/run/files/database_url": "postgres://db/sigta"}))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %v, want one that mentions %q", err, tc.want)
			}
		})
	}
}

func readFrom(files map[string]string) func(string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		content, ok := files[path]
		if !ok {
			return nil, errors.New("no such file")
		}
		return []byte(content), nil
	}
}
