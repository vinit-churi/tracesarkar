package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// throttleServer has one real account, so a correct password can be told
// apart from a wrong one.
func throttleServer(t *testing.T) *Server {
	t.Helper()
	accounts := newFakeAccounts()
	srv := authServer(t, accounts)
	rec := postJSON(t, srv, "/v1/auth/register",
		`{"email":"real@example.org","password":"correct-horse-battery","name":"R"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("could not seed an account: %d %s", rec.Code, rec.Body.String())
	}
	return srv
}

func attempt(t *testing.T, srv *Server, email, password, ip string) int {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	req := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if ip != "" {
		req.Header.Set("X-Forwarded-For", ip)
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	return rec.Code
}

// Guessing a password is the cheapest attack on this platform, and nothing
// stopped it.
func TestWrongPasswordsAreThrottledPerAccount(t *testing.T) {
	srv := throttleServer(t)

	var refused bool
	for i := range failedLoginLimit + 2 {
		code := attempt(t, srv, "victim@example.org", "guess", "203.0.113.9")
		if code == http.StatusTooManyRequests {
			refused = true
			if i < failedLoginLimit {
				t.Fatalf("refused on attempt %d, before the limit of %d",
					i+1, failedLoginLimit)
			}
			break
		}
	}
	if !refused {
		t.Fatalf("%d wrong passwords in a row were all accepted for trying",
			failedLoginLimit+2)
	}
}

// Locking one account must not lock the others, or anyone can deny service to
// everyone by guessing at a single address.
func TestThrottlingOneAccountDoesNotLockAnother(t *testing.T) {
	srv := throttleServer(t)

	for range failedLoginLimit + 2 {
		attempt(t, srv, "victim@example.org", "guess", "203.0.113.9")
	}
	if code := attempt(t, srv, "someone.else@example.org", "guess", "198.51.100.4"); code == http.StatusTooManyRequests {
		t.Error("a second account was locked out by attempts on the first")
	}
}

// A refusal must say when to come back, and must not say whether the account
// exists. The limit applies identically either way — otherwise the throttle
// itself becomes the oracle it was meant to protect.
func TestARefusalSaysWhenAndNotWhether(t *testing.T) {
	srv := throttleServer(t)

	var rec *httptest.ResponseRecorder
	for range failedLoginLimit + 2 {
		body := `{"email":"nobody@example.org","password":"guess"}`
		req := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "203.0.113.9")
		rec = httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			break
		}
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatal("an address that does not exist was never throttled")
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("no Retry-After on a refusal")
	}

	var body struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	lower := strings.ToLower(body.Error.Message)
	for _, leak := range []string{"no such", "not found", "does not exist", "unknown account"} {
		if strings.Contains(lower, leak) {
			t.Errorf("the refusal reveals whether the account exists: %q", body.Error.Message)
		}
	}
}

// A successful sign-in must not be counted against the person. Someone who
// signs in on a phone and a laptop every morning is not an attacker.
func TestSucceedingDoesNotCountTowardTheLimit(t *testing.T) {
	srv := throttleServer(t)

	for range failedLoginLimit * 3 {
		body := `{"email":"real@example.org","password":"correct-horse-battery"}`
		req := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "203.0.113.9")
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, req)
		if rec.Code == http.StatusTooManyRequests {
			t.Fatal("a correct password was throttled")
		}
	}
}
