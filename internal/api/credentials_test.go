package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/mail"
)

var errMailDown = errors.New("the mail route is down")

// recorder is a mail.Sender that keeps what it was asked to send.
type recorder struct {
	mu   sync.Mutex
	sent []mail.Message
	err  error
}

func (r *recorder) Describe() string { return "recorder" }

func (r *recorder) Send(_ context.Context, m mail.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	r.sent = append(r.sent, m)
	return nil
}

func (r *recorder) last(t *testing.T) mail.Message {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.sent) == 0 {
		t.Fatal("no mail was sent")
	}
	return r.sent[len(r.sent)-1]
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sent)
}

// tokenFrom pulls the one-time token out of the link in the body, which is
// what a person would do.
func tokenFrom(t *testing.T, m mail.Message) string {
	t.Helper()
	i := strings.Index(m.Body, "token=")
	if i < 0 {
		t.Fatalf("no token in the mail:\n%s", m.Body)
	}
	tok := m.Body[i+len("token="):]
	if j := strings.IndexAny(tok, " \n\r"); j >= 0 {
		tok = tok[:j]
	}
	if tok == "" {
		t.Fatalf("empty token in the mail:\n%s", m.Body)
	}
	return tok
}

func credServer(t *testing.T, accounts *fakeAccounts, sender mail.Sender) *Server {
	t.Helper()
	issuer, err := auth.NewIssuer("a-test-signing-secret-32-bytes!!!", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{}, Token: "test-token",
		Account: "account-1", Accounts: accounts, Issuer: issuer,
		Credentials: accounts, Mail: sender,
		PublicBaseURL: "https://tracesarkar.org",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

// registered returns a server with one account whose password is known.
func registered(t *testing.T, sender mail.Sender) (*Server, *fakeAccounts) {
	t.Helper()
	accounts := newFakeAccounts()
	srv := credServer(t, accounts, sender)
	rec := postJSON(t, srv, "/v1/auth/register",
		`{"email":"person@example.org","password":"a-reasonable-passphrase","name":"A Person"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", rec.Code, rec.Body.String())
	}
	return srv, accounts
}

// The answer must not depend on whether the address is registered. On a civic
// platform, "is this person a user here" is worth something to the wrong
// people, and a forgotten-password form that answers it is the easiest place
// to ask.
func TestForgotPasswordAnswersTheSameForAStranger(t *testing.T) {
	sender := &recorder{}
	srv, _ := registered(t, sender)

	known := postJSON(t, srv, "/v1/auth/password/forgot", `{"email":"person@example.org"}`)
	unknown := postJSON(t, srv, "/v1/auth/password/forgot", `{"email":"nobody@example.org"}`)

	if known.Code != unknown.Code {
		t.Errorf("a known address answered %d and an unknown one %d",
			known.Code, unknown.Code)
	}
	if known.Code != http.StatusAccepted {
		t.Errorf("status is %d, want 202", known.Code)
	}
	if known.Body.String() != unknown.Body.String() {
		t.Errorf("the bodies differ:\nknown:   %s\nunknown: %s",
			known.Body.String(), unknown.Body.String())
	}
	// And only the real address gets a mail.
	if sender.count() != 1 {
		t.Errorf("%d mails were sent, want 1", sender.count())
	}
	if to := sender.last(t).To; to != "person@example.org" {
		t.Errorf("the mail went to %q", to)
	}
	// The response must not leak the token either.
	if strings.Contains(known.Body.String(), "token") {
		t.Errorf("the response body mentions a token: %s", known.Body.String())
	}
}

func TestAResetLinkSetsANewPasswordExactlyOnce(t *testing.T) {
	sender := &recorder{}
	srv, _ := registered(t, sender)

	if rec := postJSON(t, srv, "/v1/auth/password/forgot",
		`{"email":"person@example.org"}`); rec.Code != http.StatusAccepted {
		t.Fatalf("forgot: %d", rec.Code)
	}
	token := tokenFrom(t, sender.last(t))

	rec := postJSON(t, srv, "/v1/auth/password/reset",
		`{"token":"`+token+`","password":"a-different-passphrase"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("reset: %d %s", rec.Code, rec.Body.String())
	}
	// No session comes back. Somebody who just proved control of a mailbox
	// has not yet proved they know the password they set, and signing them in
	// anyway makes the mailbox the credential.
	if strings.Contains(rec.Body.String(), `"token"`) {
		t.Errorf("the reset returned a session token: %s", rec.Body.String())
	}

	// The new password works and the old one does not.
	if got := postJSON(t, srv, "/v1/auth/login",
		`{"email":"person@example.org","password":"a-different-passphrase"}`); got.Code != http.StatusOK {
		t.Errorf("the new password does not work: %d %s", got.Code, got.Body.String())
	}
	if got := postJSON(t, srv, "/v1/auth/login",
		`{"email":"person@example.org","password":"a-reasonable-passphrase"}`); got.Code != http.StatusUnauthorized {
		t.Errorf("the old password still works: %d", got.Code)
	}

	// The same link again does nothing.
	again := postJSON(t, srv, "/v1/auth/password/reset",
		`{"token":"`+token+`","password":"a-third-passphrase"}`)
	if again.Code == http.StatusOK {
		t.Error("the same reset link worked twice")
	}
}

func TestAResetRefusesWhatItShould(t *testing.T) {
	sender := &recorder{}
	srv, _ := registered(t, sender)
	postJSON(t, srv, "/v1/auth/password/forgot", `{"email":"person@example.org"}`)
	good := tokenFrom(t, sender.last(t))

	cases := []struct {
		name, body string
		want       int
	}{
		{"a forged token", `{"token":"not-a-real-token","password":"a-reasonable-passphrase"}`,
			http.StatusBadRequest},
		{"no token", `{"token":"","password":"a-reasonable-passphrase"}`, http.StatusBadRequest},
		{"a password that is too short", `{"token":"` + good + `","password":"short"}`,
			http.StatusBadRequest},
		{"no password", `{"token":"` + good + `","password":""}`, http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if rec := postJSON(t, srv, "/v1/auth/password/reset", c.body); rec.Code != c.want {
				t.Errorf("status %d, want %d: %s", rec.Code, c.want, rec.Body.String())
			}
		})
	}
	// A refused weak password must not have spent the link.
	if rec := postJSON(t, srv, "/v1/auth/password/reset",
		`{"token":"`+good+`","password":"a-reasonable-passphrase"}`); rec.Code != http.StatusOK {
		t.Errorf("a refused attempt spent the link: %d %s", rec.Code, rec.Body.String())
	}
}

// A reset exists because a password may have been stolen. Leaving the thief's
// session alive is worse than not resetting at all.
func TestResettingAPasswordEndsEverySession(t *testing.T) {
	sender := &recorder{}
	srv, _ := registered(t, sender)

	signIn := postJSON(t, srv, "/v1/auth/login",
		`{"email":"person@example.org","password":"a-reasonable-passphrase"}`)
	var got struct{ Token string }
	if err := json.Unmarshal(signIn.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Token == "" {
		t.Fatal("no session token")
	}
	if code := getWith(t, srv, "/v1/auth/me", got.Token); code != http.StatusOK {
		t.Fatalf("the session does not work to begin with: %d", code)
	}

	postJSON(t, srv, "/v1/auth/password/forgot", `{"email":"person@example.org"}`)
	token := tokenFrom(t, sender.last(t))
	if rec := postJSON(t, srv, "/v1/auth/password/reset",
		`{"token":"`+token+`","password":"a-different-passphrase"}`); rec.Code != http.StatusOK {
		t.Fatalf("reset: %d %s", rec.Code, rec.Body.String())
	}

	if code := getWith(t, srv, "/v1/auth/me", got.Token); code != http.StatusUnauthorized {
		t.Errorf("the old session still works after a reset: %d", code)
	}
}

func TestSigningOutEverywhereEndsEverySessionAndKeepsThePassword(t *testing.T) {
	srv, _ := registered(t, &recorder{})

	first := signInToken(t, srv, "a-reasonable-passphrase")
	second := signInToken(t, srv, "a-reasonable-passphrase")

	rec := postWith(t, srv, "/v1/auth/sessions/revoke", "", first)
	if rec.Code != http.StatusOK {
		t.Fatalf("revoke: %d %s", rec.Code, rec.Body.String())
	}
	for name, tok := range map[string]string{"the session that asked": first, "another session": second} {
		if code := getWith(t, srv, "/v1/auth/me", tok); code != http.StatusUnauthorized {
			t.Errorf("%s still works: %d", name, code)
		}
	}
	// Not a lockout: the password is untouched.
	if code := postJSON(t, srv, "/v1/auth/login",
		`{"email":"person@example.org","password":"a-reasonable-passphrase"}`).Code; code != http.StatusOK {
		t.Errorf("signing out everywhere locked the account out: %d", code)
	}
	// And revoking needs a session of its own.
	if rec := postJSON(t, srv, "/v1/auth/sessions/revoke", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("revoking without a session answered %d", rec.Code)
	}
}

func TestVerifyingAnAddressMarksTheAccountAndNeedsASession(t *testing.T) {
	sender := &recorder{}
	srv, accounts := registered(t, sender)
	session := signInToken(t, srv, "a-reasonable-passphrase")

	if rec := postJSON(t, srv, "/v1/auth/email/verify/request", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("an unauthenticated request answered %d", rec.Code)
	}
	if rec := postWith(t, srv, "/v1/auth/email/verify/request", "", session); rec.Code != http.StatusAccepted {
		t.Fatalf("request: %d %s", rec.Code, rec.Body.String())
	}
	token := tokenFrom(t, sender.last(t))

	// Confirming is not a sign-in: the link is followed in a browser that may
	// have no session at all.
	if rec := postJSON(t, srv, "/v1/auth/email/verify",
		`{"token":"`+token+`"}`); rec.Code != http.StatusOK {
		t.Fatalf("verify: %d %s", rec.Code, rec.Body.String())
	}
	if !accounts.verified["account-new"] {
		t.Error("the account was not marked verified")
	}
	// Once only.
	if rec := postJSON(t, srv, "/v1/auth/email/verify",
		`{"token":"`+token+`"}`); rec.Code == http.StatusOK {
		t.Error("the same verification link worked twice")
	}
}

// A verification link must not be presentable as a reset link, or the weaker
// of the two windows sets the strength of both.
func TestACredentialIsOnlyGoodForThePurposeItWasMintedFor(t *testing.T) {
	sender := &recorder{}
	srv, _ := registered(t, sender)
	session := signInToken(t, srv, "a-reasonable-passphrase")

	postWith(t, srv, "/v1/auth/email/verify/request", "", session)
	verifyToken := tokenFrom(t, sender.last(t))

	if rec := postJSON(t, srv, "/v1/auth/password/reset",
		`{"token":"`+verifyToken+`","password":"a-brand-new-passphrase"}`); rec.Code == http.StatusOK {
		t.Error("a verification link reset a password")
	}

	postJSON(t, srv, "/v1/auth/password/forgot", `{"email":"person@example.org"}`)
	resetToken := tokenFrom(t, sender.last(t))
	if rec := postJSON(t, srv, "/v1/auth/email/verify",
		`{"token":"`+resetToken+`"}`); rec.Code == http.StatusOK {
		t.Error("a reset link verified an address")
	}
}

// Mail that cannot be sent must not leave a live credential behind with
// nobody holding it, and must not tell the caller whether the address exists.
func TestAMailFailureStillAnswersTheSameWay(t *testing.T) {
	sender := &recorder{err: errMailDown}
	srv, _ := registered(t, sender)
	rec := postJSON(t, srv, "/v1/auth/password/forgot", `{"email":"person@example.org"}`)
	if rec.Code != http.StatusAccepted {
		t.Errorf("status %d, want 202 even when mail is down: %s", rec.Code, rec.Body.String())
	}
}

// With no mail route at all the endpoints have to refuse rather than mint a
// credential nobody will ever receive.
func TestWithNoMailRouteTheEndpointsRefuseCleanly(t *testing.T) {
	accounts := newFakeAccounts()
	issuer, _ := auth.NewIssuer("a-test-signing-secret-32-bytes!!!", time.Hour)
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{}, Token: "test-token",
		Account: "account-1", Accounts: accounts, Issuer: issuer,
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec := postJSON(t, srv, "/v1/auth/password/forgot",
		`{"email":"person@example.org"}`); rec.Code != http.StatusNotImplemented {
		t.Errorf("status %d, want 501", rec.Code)
	}
}

// Helpers --------------------------------------------------------------------

func signInToken(t *testing.T, srv *Server, password string) string {
	t.Helper()
	rec := postJSON(t, srv, "/v1/auth/login",
		`{"email":"person@example.org","password":"`+password+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body.String())
	}
	var got struct{ Token string }
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Token == "" {
		t.Fatal("no session token")
	}
	return got.Token
}

func getWith(t *testing.T, srv *Server, path, token string) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	return rec.Code
}

func postWith(t *testing.T, srv *Server, path, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	return rec
}
