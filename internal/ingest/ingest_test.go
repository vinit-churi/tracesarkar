package ingest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchRecordsEveryAttemptIncludingFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	f := NewFetcher(srv.Client(), "TraceSarkar-test/0.1")
	_, err := f.Fetch(context.Background(), srv.URL)

	if err == nil {
		t.Fatal("a 500 must be an error")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected an *HTTPError, got %T", err)
	}
	if httpErr.StatusCode != 500 {
		t.Errorf("status: got %d", httpErr.StatusCode)
	}
}

func TestFetchReturnsBodyStatusAndHash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	f := NewFetcher(srv.Client(), "TraceSarkar-test/0.1")
	res, err := f.Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	if string(res.Body) != `{"ok":true}` {
		t.Errorf("body: got %s", res.Body)
	}
	if res.StatusCode != 200 {
		t.Errorf("status: got %d", res.StatusCode)
	}
	if len(res.SHA256) != 64 {
		t.Errorf("sha256 should be hex-encoded: got %q", res.SHA256)
	}
	if res.ContentType != "application/json" {
		t.Errorf("content type: got %q", res.ContentType)
	}
	if res.Duration <= 0 {
		t.Error("duration should be measured")
	}
}

func TestFetchSendsHonestUserAgent(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	f := NewFetcher(srv.Client(), "TraceSarkar/0.1 (+https://example.org/crawler; contact@example.org)")
	if _, err := f.Fetch(context.Background(), srv.URL); err != nil {
		t.Fatal(err)
	}
	if got != "TraceSarkar/0.1 (+https://example.org/crawler; contact@example.org)" {
		t.Errorf("user agent: got %q", got)
	}
}

func TestFetchRetriesOnServerErrorThenSucceeds(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	f := NewFetcher(srv.Client(), "ua")
	f.Retries = 3
	f.Backoff = time.Millisecond

	res, err := f.Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 attempts, got %d", calls)
	}
	if res.Attempts != 3 {
		t.Errorf("attempts should be reported: got %d", res.Attempts)
	}
}

func TestFetchDoesNotRetryClientErrors(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	f := NewFetcher(srv.Client(), "ua")
	f.Retries = 3
	f.Backoff = time.Millisecond

	if _, err := f.Fetch(context.Background(), srv.URL); err == nil {
		t.Fatal("expected an error")
	}
	if calls != 1 {
		t.Errorf("a 404 should not be retried: %d attempts", calls)
	}
}

func TestFetchRejectsAnEmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	f := NewFetcher(srv.Client(), "ua")
	f.Retries = 1
	if _, err := f.Fetch(context.Background(), srv.URL); err == nil {
		t.Fatal("an empty response is a failure, not a successful zero-row run")
	}
}
