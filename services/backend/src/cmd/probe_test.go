package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbe(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		healthy bool
	}{
		{"answers", http.StatusOK, true},
		{"fails", http.StatusInternalServerError, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer server.Close()
			if err := probe(server.URL); (err == nil) != tc.healthy {
				t.Fatalf("probe = %v, want healthy %v", err, tc.healthy)
			}
		})
	}
}

func TestProbeWithoutServer(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	url := server.URL
	server.Close()
	if err := probe(url); err == nil {
		t.Fatal("probe succeeded without a server")
	}
}
