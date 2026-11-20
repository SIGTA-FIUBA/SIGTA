package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestServeLetsTheRequestsInFlightFinish stops the server while a request is
// still running: the request gets its answer before serve returns.
func TestServeLetsTheRequestsInFlightFinish(t *testing.T) {
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(started)
			<-release
			w.WriteHeader(http.StatusNoContent)
		}),
		ReadHeaderTimeout: time.Second,
	}
	stop, cancel := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- serve(stop, server, func() error { return server.Serve(listener) }) }()

	answered := make(chan int, 1)
	go func() {
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+listener.Addr().String(), nil)
		if err != nil {
			t.Error(err)
			answered <- 0
			return
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Error(err)
			answered <- 0
			return
		}
		_ = response.Body.Close()
		answered <- response.StatusCode
	}()
	<-started
	cancel()
	select {
	case err := <-served:
		t.Fatalf("serve() = %v before the request in flight finished", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	if status := <-answered; status != http.StatusNoContent {
		t.Errorf("the request in flight got %d, want %d", status, http.StatusNoContent)
	}
	if err := <-served; err != nil {
		t.Errorf("serve() = %v, want nil", err)
	}
}

// TestStoppedDuringStartup checks that a start interrupted by a signal ends
// without an error, and that one that fails on its own keeps it.
func TestStoppedDuringStartup(t *testing.T) {
	failed := errors.New("the database did not answer")
	if err := stoppedDuringStartup(t.Context(), failed); !errors.Is(err, failed) {
		t.Errorf("without a signal = %v, want %v", err, failed)
	}
	stop, cancel := context.WithCancel(t.Context())
	cancel()
	if err := stoppedDuringStartup(stop, failed); err != nil {
		t.Errorf("after a signal = %v, want nil", err)
	}
}

func TestNowIsAsPostgreSQLKeepsIt(t *testing.T) {
	if got := now(); got.Location() != time.UTC || got.Nanosecond()%1000 != 0 {
		t.Errorf("now() = %v, want UTC to the microsecond", got)
	}
}
