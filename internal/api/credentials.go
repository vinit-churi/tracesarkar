package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/mail"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// ErrNoCredential is a one-time link that cannot be used.
var ErrNoCredential = store.ErrNoCredential

// Credentials is the persistence the reset and verification endpoints need.
//
// Separate from Accounts rather than bolted onto it, so a deployment without
// a mail route simply does not implement it and the endpoints refuse cleanly
// instead of half-working.
type Credentials interface {
	CreateAuthToken(ctx context.Context, accountID string, purpose auth.Purpose, hash []byte, expiresAt time.Time) error
	ConsumeAuthToken(ctx context.Context, purpose auth.Purpose, hash []byte) (accountID string, err error)
	SetPassword(ctx context.Context, accountID string, hash []byte) (version int64, err error)
	RevokeSessions(ctx context.Context, accountID string) (version int64, err error)
	TokenVersion(ctx context.Context, accountID string) (int64, error)
	MarkEmailVerified(ctx context.Context, accountID string) error
}

// handleForgotPassword mints a reset link and mails it.
//
// It answers 202 and the same body whatever happens: for a registered
// address, an unregistered one, an account with no password at all, and a
// mail route that is down. The reason is not politeness. On a civic platform
// "is this person registered here" is worth something to the wrong people,
// and a forgotten-password form is the easiest place in any product to ask
// the question a thousand times.
func (s *Server) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if !s.allowAuthAttempt(w, r) {
		return
	}
	if s.accounts == nil || s.credentials == nil || s.mail == nil {
		writeError(w, http.StatusNotImplemented,
			"password resets are not configured on this server")
		return
	}

	var in struct{ Email string }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not valid JSON")
		return
	}

	// The one identical answer, written once and returned on every path
	// below, so a later edit cannot make one branch distinguishable.
	accepted := func() {
		writeJSON(w, http.StatusAccepted, map[string]any{
			"status": "accepted",
			"note": "If that address has an account here, a link to set a new " +
				"password is on its way. The link works once and expires within " +
				"the hour.",
		})
	}

	if err := auth.CheckEmail(in.Email); err != nil {
		// Not even a malformed address gets a different answer: "that is not
		// an address" and "that address is not registered" are two steps of
		// the same probe.
		accepted()
		return
	}

	account, err := s.accounts.FindAccountByEmail(r.Context(), in.Email)
	if err != nil && !errors.Is(err, ErrNoAccount) {
		s.log.Error("could not read account for a reset", "error", err.Error())
		accepted()
		return
	}
	if errors.Is(err, ErrNoAccount) {
		accepted()
		return
	}

	s.issueCredential(r.Context(), auth.PasswordReset, account.ID, account.Email)
	accepted()
}

// handleResetPassword spends a reset link and sets the password.
func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if !s.allowAuthAttempt(w, r) {
		return
	}
	if s.credentials == nil {
		writeError(w, http.StatusNotImplemented,
			"password resets are not configured on this server")
		return
	}

	var in struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not valid JSON")
		return
	}
	if strings.TrimSpace(in.Token) == "" {
		writeError(w, http.StatusBadRequest, "that link is not valid")
		return
	}
	// The password policy is checked before the link is spent. A person who
	// typed something too short must get another go with the same link; the
	// alternative is a dead link and a second mail for a typo.
	if err := auth.CheckPasswordPolicy(in.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		s.log.Error("could not hash password", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not set the password")
		return
	}

	accountID, err := s.credentials.ConsumeAuthToken(r.Context(),
		auth.PasswordReset, auth.HashOneTime(in.Token))
	if errors.Is(err, ErrNoCredential) {
		writeError(w, http.StatusBadRequest,
			"that link is not valid. It may have been used already or expired; ask for a new one")
		return
	}
	if err != nil {
		s.log.Error("could not spend a reset link", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not set the password")
		return
	}

	if _, err := s.credentials.SetPassword(r.Context(), accountID, hash); err != nil {
		s.log.Error("could not set password", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not set the password")
		return
	}
	s.log.Info("password was reset", "account_id", accountID)

	// No session comes back, and that is deliberate. Somebody who has just
	// proved control of a mailbox has not proved they know the password they
	// set; handing them a session would make the mailbox the credential and
	// make a reset link a sign-in link.
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "password_set",
		"note": "Your password has been changed and every signed-in device has " +
			"been signed out. Sign in with the new password.",
	})
}

// handleRequestEmailVerification mails a link proving the address reaches the
// person holding the account. Authenticated: there is no reason for a
// stranger to be able to send mail to somebody else's address.
func (s *Server) handleRequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "sign in first")
		return
	}
	if s.accounts == nil || s.credentials == nil || s.mail == nil {
		writeError(w, http.StatusNotImplemented,
			"email verification is not configured on this server")
		return
	}
	if !s.allowAuthAttempt(w, r) {
		return
	}

	info, err := s.accounts.GetAccount(r.Context(), claims.AccountID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "sign in first")
		return
	}
	if strings.TrimSpace(info.Email) == "" {
		writeError(w, http.StatusBadRequest, "this account has no email address to confirm")
		return
	}
	if info.EmailVerified {
		writeJSON(w, http.StatusAccepted, map[string]any{
			"status": "already_verified",
			"note":   "This address is already confirmed.",
		})
		return
	}

	s.issueCredential(r.Context(), auth.EmailVerification, info.ID, info.Email)
	writeJSON(w, http.StatusAccepted, map[string]any{
		"status": "accepted",
		"note":   "A link to confirm this address is on its way.",
	})
}

// handleVerifyEmail spends a verification link.
//
// Unauthenticated on purpose: the link is opened in whatever browser the mail
// client hands it to, which is often not the one holding the session. The
// token is the proof, and it proves only that somebody reads that mailbox —
// which is exactly the claim being recorded.
func (s *Server) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	if !s.allowAuthAttempt(w, r) {
		return
	}
	if s.credentials == nil {
		writeError(w, http.StatusNotImplemented,
			"email verification is not configured on this server")
		return
	}

	var in struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not valid JSON")
		return
	}
	if strings.TrimSpace(in.Token) == "" {
		writeError(w, http.StatusBadRequest, "that link is not valid")
		return
	}

	accountID, err := s.credentials.ConsumeAuthToken(r.Context(),
		auth.EmailVerification, auth.HashOneTime(in.Token))
	if errors.Is(err, ErrNoCredential) {
		writeError(w, http.StatusBadRequest,
			"that link is not valid. It may have been used already or expired; ask for a new one")
		return
	}
	if err != nil {
		s.log.Error("could not spend a verification link", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not confirm the address")
		return
	}
	if err := s.credentials.MarkEmailVerified(r.Context(), accountID); err != nil {
		s.log.Error("could not mark an address verified", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not confirm the address")
		return
	}
	s.log.Info("email address confirmed", "account_id", accountID)

	// No session, for the same reason as the reset: confirming an address is
	// not signing in, and the person may be on a borrowed device.
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "verified",
		"note":   "This address is confirmed. You can close this page.",
	})
}

// handleRevokeSessions is "sign out everywhere".
//
// It bumps the account's token version, which invalidates every token in
// existence including the one that asked. The password is untouched: this is
// not a lockout, it is a person who left a session open on a device they no
// longer have.
func (s *Server) handleRevokeSessions(w http.ResponseWriter, r *http.Request) {
	claims, ok := claimsFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "sign in first")
		return
	}
	if s.credentials == nil {
		writeError(w, http.StatusNotImplemented,
			"session revocation is not configured on this server")
		return
	}

	if _, err := s.credentials.RevokeSessions(r.Context(), claims.AccountID); err != nil {
		s.log.Error("could not revoke sessions", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not sign out your devices")
		return
	}
	s.log.Info("sessions revoked", "account_id", claims.AccountID)
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "revoked",
		"note": "Every signed-in device has been signed out, including this one. " +
			"Your password has not changed.",
	})
}

// issueCredential mints a one-time credential, records its hash and mails the
// link.
//
// Every failure is logged and swallowed. The caller has already decided to
// answer 202 whatever happens, and it must keep answering 202: a 500 here
// would be a different answer for a registered address, which is the one
// thing handleForgotPassword exists to avoid.
func (s *Server) issueCredential(ctx context.Context, purpose auth.Purpose, accountID, email string) {
	ot, err := auth.NewOneTime()
	if err != nil {
		s.log.Error("could not mint a one-time credential",
			"purpose", string(purpose), "error", err.Error())
		return
	}
	link, err := mail.Link(s.publicBaseURL, purpose, ot.Token)
	if err != nil {
		s.log.Error("could not build a link", "purpose", string(purpose), "error", err.Error())
		return
	}
	msg, err := mail.Compose(purpose, email, link)
	if err != nil {
		s.log.Error("could not compose mail", "purpose", string(purpose), "error", err.Error())
		return
	}

	// The row before the mail. A credential that exists and was never
	// delivered is a link nobody can use; a mail that was delivered for a
	// credential that was never recorded is a link that looks broken.
	if err := s.credentials.CreateAuthToken(ctx, accountID, purpose, ot.Hash,
		time.Now().UTC().Add(purpose.TTL())); err != nil {
		s.log.Error("could not record a one-time credential",
			"purpose", string(purpose), "account_id", accountID, "error", err.Error())
		return
	}

	// Never the link and never the token in this log line: that is the whole
	// credential, and a log is not where it belongs — except through
	// mail.LogSender, which is selected deliberately and says so.
	if err := s.mail.Send(ctx, msg); err != nil {
		s.log.Error("could not send mail", "purpose", string(purpose),
			"account_id", accountID, "error", err.Error())
		return
	}
	s.log.Info("sent a one-time link", "purpose", string(purpose), "account_id", accountID)
}
