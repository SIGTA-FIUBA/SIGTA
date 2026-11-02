package rest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/src/adapters/rest"
)

func newRouter() http.Handler {
	api := rest.NewServer(nil)
	return rest.HandlerFromMux(rest.NewStrictHandler(api, nil), http.NewServeMux())
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
