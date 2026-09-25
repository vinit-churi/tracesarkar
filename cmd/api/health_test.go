package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthCheckPassesOnAHealthyServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	if err := healthCheck(context.Background(), srv.URL); err != nil {
		t.Fatalf("a 200 must pass: %v", err)
	}
}

func TestHealthCheckFailsOnAnUnhealthyServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	err := healthCheck(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("a 503 must fail, or the orchestrator will keep routing to it")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("the status should be in the message: %v", err)
	}
}

func TestHealthCheckFailsWhenNothingIsListening(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // nothing is listening now

	if err := healthCheck(context.Background(), url); err == nil {
		t.Fatal("a refused connection must fail")
	}
}

func TestHealthCheckGivesUpRatherThanHanging(t *testing.T) {
	// A server that accepts the connection and then stalls must not hold the
	// healthcheck open forever; Docker would read that as "still starting".
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	if err := healthCheck(ctx, srv.URL); err == nil {
		t.Fatal("a stalled server must fail the check")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("gave up too late: %s", elapsed)
	}
}
