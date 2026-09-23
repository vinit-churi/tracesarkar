package archive

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := New(Options{
		Endpoint:  srv.URL,
		Bucket:    "tracesarkar",
		Region:    "auto",
		AccessKey: "ak",
		Secret:    "sk",
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestKeyForLaysOutBySourceAndDate(t *testing.T) {
	when := time.Date(2026, 9, 23, 4, 5, 6, 0, time.UTC)
	got := KeyFor("bmc_roads_api", when, "abc123", ".json")
	want := "archive/bmc_roads_api/2026/09/23/abc123.json"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestKeyForDefaultsExtensionToBin(t *testing.T) {
	when := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	if got := KeyFor("s", when, "h", ""); got != "archive/s/2026/01/02/h.bin" {
		t.Errorf("got %q", got)
	}
}

func TestPutSendsSignedRequestToBucketPath(t *testing.T) {
	var gotPath, gotAuth, gotType, gotBody string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	})

	if err := c.Put(context.Background(), "archive/x/1.json", []byte(`{"a":1}`), "application/json"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	if gotPath != "/tracesarkar/archive/x/1.json" {
		t.Errorf("path: got %q", gotPath)
	}
	if !strings.HasPrefix(gotAuth, "AWS4-HMAC-SHA256 Credential=ak/") {
		t.Errorf("authorization: got %q", gotAuth)
	}
	if gotType != "application/json" {
		t.Errorf("content-type: got %q", gotType)
	}
	if gotBody != `{"a":1}` {
		t.Errorf("body: got %q", gotBody)
	}
}

func TestPutReturnsErrorOnRejection(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "<Error><Code>AccessDenied</Code></Error>")
	})

	err := c.Put(context.Background(), "k", []byte("b"), "text/plain")
	if err == nil {
		t.Fatal("expected an error on 403")
	}
	if !strings.Contains(err.Error(), "403") || !strings.Contains(err.Error(), "AccessDenied") {
		t.Errorf("error should carry status and body: %v", err)
	}
}

func TestPutNeverLeaksTheSecretInErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	c, err := New(Options{Endpoint: srv.URL, Bucket: "b", Region: "auto", AccessKey: "ak", Secret: "topsecretvalue"})
	if err != nil {
		t.Fatal(err)
	}

	err = c.Put(context.Background(), "k", []byte("b"), "text/plain")
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "topsecretvalue") {
		t.Fatalf("error leaked the secret: %v", err)
	}
}

func TestExistsReportsPresenceFromHead(t *testing.T) {
	var method string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		if r.URL.Path == "/tracesarkar/present" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	ok, err := c.Exists(context.Background(), "present")
	if err != nil || !ok {
		t.Fatalf("present: ok=%v err=%v", ok, err)
	}
	if method != http.MethodHead {
		t.Errorf("expected HEAD, got %s", method)
	}

	ok, err = c.Exists(context.Background(), "absent")
	if err != nil || ok {
		t.Fatalf("absent: ok=%v err=%v", ok, err)
	}
}

func TestGetReturnsBody(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "stored bytes")
	})

	got, err := c.Get(context.Background(), "k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "stored bytes" {
		t.Errorf("got %q", got)
	}
}

func TestNewRequiresEndpointAndBucket(t *testing.T) {
	if _, err := New(Options{Bucket: "b", AccessKey: "a", Secret: "s"}); err == nil {
		t.Error("expected an error without an endpoint")
	}
	if _, err := New(Options{Endpoint: "https://x", AccessKey: "a", Secret: "s"}); err == nil {
		t.Error("expected an error without a bucket")
	}
}
