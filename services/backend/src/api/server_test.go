package api_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/src/api"
)

func newRouter() http.Handler {
	return api.Mount(http.NewServeMux(), nil, slog.Default())
}

func TestRejectsTextThatCannotBeStored(t *testing.T) {
	tests := map[string]struct{ method, target, body string }{
		"a NUL in the path":             {http.MethodGet, "/procedures/P%00", ""},
		"a path that is not UTF-8":      {http.MethodGet, "/procedures/P%ff", ""},
		"a NUL in the query":            {http.MethodGet, "/auth/login?login_hint=%00", ""},
		"a NUL in a string of the body": {http.MethodPost, "/me", `{"name":"a\u0000b"}`},
		"a NUL in a key of the body":    {http.MethodPost, "/me", `{"a\u0000":"b"}`},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if rec := serve(t, tt.method, tt.target, tt.body); rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

// TestPassesTextThatCanBeStored reaches the router, which knows no POST on
// /me, with a body that only names a NUL.
func TestPassesTextThatCanBeStored(t *testing.T) {
	if rec := serve(t, http.MethodPost, "/me", `{"name":"a\\u0000b"}`); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func serve(t *testing.T, method, target, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequestWithContext(t.Context(), method, target, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	newRouter().ServeHTTP(rec, r)
	return rec
}

func TestRejectsUnknownRoutesAndMethods(t *testing.T) {
	tests := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/unknown", http.StatusNotFound},
		{http.MethodPost, "/me", http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newRouter().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), tt.method, tt.path, nil))

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
