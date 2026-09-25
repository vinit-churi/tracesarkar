package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/auth"
)

type fakeAccounts struct {
	byEmail  map[string]Account
	created  []string
	nextID   string
	takenErr bool
}

func newFakeAccounts() *fakeAccounts {
	return &fakeAccounts{byEmail: map[string]Account{}, nextID: "account-new"}
}

func (f *fakeAccounts) CreateEmailAccount(_ context.Context, email string, hash []byte, name string) (string, error) {
	key := auth.NormaliseEmail(email)
	if _, exists := f.byEmail[key]; exists || f.takenErr {
		return "", ErrEmailTaken
	}
	f.byEmail[key] = Account{ID: f.nextID, Email: key, PasswordHash: hash, DisplayName: name}
	f.created = append(f.created, f.nextID)
	return f.nextID, nil
}

func (f *fakeAccounts) FindAccountByEmail(_ context.Context, email string) (Account, error) {
	a, ok := f.byEmail[auth.NormaliseEmail(email)]
	if !ok {
		return Account{}, ErrNoAccount
	}
	return a, nil
}

func (f *fakeAccounts) UpsertGoogleAccount(_ context.Context, id auth.Identity) (string, bool, error) {
	return "account-google", true, nil
}

func (f *fakeAccounts) GetAccount(_ context.Context, id string) (AccountInfo, error) {
	for _, a := range f.byEmail {
		if a.ID == id {
			return AccountInfo{ID: a.ID, Email: a.Email, DisplayName: a.DisplayName,
				AuthProviders: []string{"password"}}, nil
		}
	}
	return AccountInfo{}, ErrNoAccount
}

func authServer(t *testing.T, accounts *fakeAccounts) *Server {
	t.Helper()
	issuer, err := auth.NewIssuer("a-test-signing-secret-32-bytes!!!", 0)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{}, Token: "test-token",
		Account: "account-1", Accounts: accounts, Issuer: issuer,
		AllowedOrigins: []string{"http://localhost:5000"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func postJSON(t *testing.T, srv *Server, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	return rec
}

func TestRegisterCreatesAnAccountAndReturnsAToken(t *testing.T) {
	accounts := newFakeAccounts()
	srv := authServer(t, accounts)

	rec := postJSON(t, srv, "/v1/auth/register",
		`{"email":"Person@Example.com","password":"a-reasonable-passphrase","name":"A Person"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d, want 201: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token   string `json:"token"`
		Account struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"account"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Token == "" {
		t.Error("a token should be returned so the client is signed in already")
	}
	if resp.Account.Email != "person@example.com" {
		t.Errorf("email should be normalised: %q", resp.Account.Email)
	}
	if strings.Contains(rec.Body.String(), "passphrase") {
		t.Error("the response must never echo the password")
	}
}

func TestRegisterRefusesAWeakPassword(t *testing.T) {
	srv := authServer(t, newFakeAccounts())
	rec := postJSON(t, srv, "/v1/auth/register", `{"email":"a@b.co","password":"short"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
}

func TestRegisterRefusesABadEmail(t *testing.T) {
	srv := authServer(t, newFakeAccounts())
	rec := postJSON(t, srv, "/v1/auth/register", `{"email":"nope","password":"a-reasonable-passphrase"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
}

func TestRegisterReportsADuplicateAddress(t *testing.T) {
	accounts := newFakeAccounts()
	srv := authServer(t, accounts)
	body := `{"email":"a@b.co","password":"a-reasonable-passphrase"}`

	if rec := postJSON(t, srv, "/v1/auth/register", body); rec.Code != http.StatusCreated {
		t.Fatalf("first: %d", rec.Code)
	}
	rec := postJSON(t, srv, "/v1/auth/register", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("got %d, want 409", rec.Code)
	}
}

func TestLoginReturnsATokenForTheRightPassword(t *testing.T) {
	accounts := newFakeAccounts()
	srv := authServer(t, accounts)
	postJSON(t, srv, "/v1/auth/register", `{"email":"a@b.co","password":"a-reasonable-passphrase"}`)

	rec := postJSON(t, srv, "/v1/auth/login", `{"email":"A@B.co","password":"a-reasonable-passphrase"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "token") {
		t.Error("expected a token")
	}
}

func TestLoginRejectsTheWrongPassword(t *testing.T) {
	accounts := newFakeAccounts()
	srv := authServer(t, accounts)
	postJSON(t, srv, "/v1/auth/register", `{"email":"a@b.co","password":"a-reasonable-passphrase"}`)

	rec := postJSON(t, srv, "/v1/auth/login", `{"email":"a@b.co","password":"not-the-password"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestLoginGivesTheSameAnswerForAnUnknownAddress(t *testing.T) {
	srv := authServer(t, newFakeAccounts())

	rec := postJSON(t, srv, "/v1/auth/login", `{"email":"nobody@example.com","password":"a-reasonable-passphrase"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
	// The response must not reveal whether the address is registered.
	if strings.Contains(strings.ToLower(rec.Body.String()), "no such") ||
		strings.Contains(strings.ToLower(rec.Body.String()), "not found") {
		t.Errorf("the reply distinguishes unknown accounts from wrong passwords: %s", rec.Body.String())
	}
}

func TestMeReturnsTheSignedInAccount(t *testing.T) {
	accounts := newFakeAccounts()
	srv := authServer(t, accounts)
	rec := postJSON(t, srv, "/v1/auth/register", `{"email":"a@b.co","password":"a-reasonable-passphrase"}`)
	var reg struct {
		Token string `json:"token"`
	}
	json.Unmarshal(rec.Body.Bytes(), &reg)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+reg.Token)
	me := httptest.NewRecorder()
	srv.Handler().ServeHTTP(me, req)

	if me.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", me.Code, me.Body.String())
	}
	if !strings.Contains(me.Body.String(), "a@b.co") {
		t.Errorf("expected the account: %s", me.Body.String())
	}
}

func TestMeRejectsAnInvalidToken(t *testing.T) {
	srv := authServer(t, newFakeAccounts())
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestAUserTokenIsAcceptedOnTheCaptureEndpoint(t *testing.T) {
	accounts := newFakeAccounts()
	srv := authServer(t, accounts)
	rec := postJSON(t, srv, "/v1/auth/register", `{"email":"a@b.co","password":"a-reasonable-passphrase"}`)
	var reg struct {
		Token string `json:"token"`
	}
	json.Unmarshal(rec.Body.Bytes(), &reg)

	req := captureRequest(t, validMeta, map[string][]byte{"close": []byte("jpeg")})
	req.Header.Set("Authorization", "Bearer "+reg.Token)
	upload := httptest.NewRecorder()
	srv.Handler().ServeHTTP(upload, req)

	if upload.Code != http.StatusAccepted {
		t.Fatalf("a signed-in user should be able to upload: %d %s", upload.Code, upload.Body.String())
	}
}

func TestBrowserPreflightIsAnswered(t *testing.T) {
	srv := authServer(t, newFakeAccounts())

	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:5000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5000" {
		t.Errorf("the allowed origin should be echoed: %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestAnUnknownOriginIsNotGrantedAccess(t *testing.T) {
	srv := authServer(t, newFakeAccounts())

	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/login", nil)
	req.Header.Set("Origin", "https://somewhere-else.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("an unlisted origin must not be granted access")
	}
}

func TestGoogleSignInIsRefusedCleanlyWhenUnconfigured(t *testing.T) {
	srv := authServer(t, newFakeAccounts())
	rec := postJSON(t, srv, "/v1/auth/google", `{"id_token":"whatever"}`)
	if rec.Code == http.StatusOK {
		t.Fatal("without a client id, Google sign-in must not succeed")
	}
	if !strings.Contains(rec.Body.String(), "GOOGLE_CLIENT_ID") {
		t.Errorf("the error should say what is missing: %s", rec.Body.String())
	}
}

var _ = errors.Is

func TestMeUsesJSONFieldNamesAClientCanRead(t *testing.T) {
	accounts := newFakeAccounts()
	srv := authServer(t, accounts)
	rec := postJSON(t, srv, "/v1/auth/register", `{"email":"a@b.co","password":"a-reasonable-passphrase"}`)
	var reg struct {
		Token string `json:"token"`
	}
	json.Unmarshal(rec.Body.Bytes(), &reg)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+reg.Token)
	me := httptest.NewRecorder()
	srv.Handler().ServeHTTP(me, req)

	body := me.Body.String()
	for _, want := range []string{`"id"`, `"email"`, `"auth_providers"`} {
		if !strings.Contains(body, want) {
			t.Errorf("response should carry %s: %s", want, body)
		}
	}
	// Go field names leaking into the wire format is a client-breaking detail.
	if strings.Contains(body, `"DisplayName"`) || strings.Contains(body, `"AuthProviders"`) {
		t.Errorf("Go field names leaked into JSON: %s", body)
	}
}
