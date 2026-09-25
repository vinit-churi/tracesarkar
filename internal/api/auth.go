package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Vocabulary shared with the store, so handlers need no translation layer.
type (
	Account     = store.AccountAuth
	AccountInfo = store.AccountInfo
)

// Errors the handlers distinguish.
var (
	ErrEmailTaken = store.ErrEmailTaken
	ErrNoAccount  = store.ErrNoAccount
)

// Accounts is the identity persistence the auth endpoints need.
type Accounts interface {
	CreateEmailAccount(ctx context.Context, email string, passwordHash []byte, displayName string) (string, error)
	FindAccountByEmail(ctx context.Context, email string) (Account, error)
	UpsertGoogleAccount(ctx context.Context, identity auth.Identity) (id string, created bool, err error)
	GetAccount(ctx context.Context, id string) (AccountInfo, error)
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if s.accounts == nil || s.issuer == nil {
		writeError(w, http.StatusNotImplemented, "accounts are not configured on this server")
		return
	}
	var in credentials
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not valid JSON")
		return
	}
	if err := auth.CheckEmail(in.Email); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := auth.CheckPasswordPolicy(in.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		s.log.Error("could not hash password", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not create the account")
		return
	}

	id, err := s.accounts.CreateEmailAccount(r.Context(), in.Email, hash, strings.TrimSpace(in.Name))
	switch {
	case errors.Is(err, ErrEmailTaken):
		writeError(w, http.StatusConflict, "that email address is already registered")
		return
	case err != nil:
		s.log.Error("could not create account", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not create the account")
		return
	}

	s.issueSession(w, http.StatusCreated, id, auth.NormaliseEmail(in.Email), strings.TrimSpace(in.Name))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if s.accounts == nil || s.issuer == nil {
		writeError(w, http.StatusNotImplemented, "accounts are not configured on this server")
		return
	}
	var in credentials
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not valid JSON")
		return
	}

	account, err := s.accounts.FindAccountByEmail(r.Context(), in.Email)
	if err != nil && !errors.Is(err, ErrNoAccount) {
		s.log.Error("could not read account", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not sign in")
		return
	}

	// One answer for both "no such address" and "wrong password": telling them
	// apart hands an attacker a list of who is registered here, and on a civic
	// platform that list is worth something.
	if errors.Is(err, ErrNoAccount) || !auth.VerifyPassword(account.PasswordHash, in.Password) {
		writeError(w, http.StatusUnauthorized, "email or password is incorrect")
		return
	}

	s.issueSession(w, http.StatusOK, account.ID, account.Email, account.DisplayName)
}

func (s *Server) handleGoogle(w http.ResponseWriter, r *http.Request) {
	if s.accounts == nil || s.issuer == nil {
		writeError(w, http.StatusNotImplemented, "accounts are not configured on this server")
		return
	}
	if s.google == nil {
		writeError(w, http.StatusNotImplemented, "Google sign-in is not configured: set GOOGLE_CLIENT_ID")
		return
	}
	if reason := s.google.Unconfigured(); reason != "" {
		writeError(w, http.StatusNotImplemented, reason)
		return
	}

	var in struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not valid JSON")
		return
	}
	if strings.TrimSpace(in.IDToken) == "" {
		writeError(w, http.StatusBadRequest, "id_token is required")
		return
	}

	identity, err := s.google.Verify(r.Context(), in.IDToken)
	if err != nil {
		s.log.Warn("google sign-in refused", "error", err.Error())
		writeError(w, http.StatusUnauthorized, "that Google sign-in could not be verified")
		return
	}

	id, created, err := s.accounts.UpsertGoogleAccount(r.Context(), identity)
	if err != nil {
		s.log.Error("could not link google account", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not sign in")
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	s.issueSession(w, status, id, identity.Email, identity.Name)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "sign in first")
		return
	}
	if s.accounts == nil {
		writeError(w, http.StatusNotImplemented, "accounts are not configured on this server")
		return
	}

	info, err := s.accounts.GetAccount(r.Context(), claims.AccountID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "sign in first")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account": info})
}

func (s *Server) issueSession(w http.ResponseWriter, status int, accountID, email, name string) {
	token, err := s.issuer.Issue(accountID, email)
	if err != nil {
		s.log.Error("could not issue token", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not complete sign-in")
		return
	}
	writeJSON(w, status, map[string]any{
		"token": token,
		"account": map[string]any{
			"id":    accountID,
			"email": email,
			"name":  name,
		},
	})
}
