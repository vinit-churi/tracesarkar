package auth

import (
	"testing"
	"time"
)

func testIssuer(t *testing.T) *Issuer {
	t.Helper()
	iss, err := NewIssuer("a-secret-at-least-32-bytes-long!!", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return iss
}

func TestIssueThenVerify(t *testing.T) {
	iss := testIssuer(t)

	token, err := iss.Issue("account-123", "person@example.com")
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	claims, err := iss.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.AccountID != "account-123" {
		t.Errorf("account: got %q", claims.AccountID)
	}
	if claims.Email != "person@example.com" {
		t.Errorf("email: got %q", claims.Email)
	}
}

func TestVerifyRejectsATamperedToken(t *testing.T) {
	iss := testIssuer(t)
	token, err := iss.Issue("account-123", "person@example.com")
	if err != nil {
		t.Fatal(err)
	}

	// Flip a character in the payload; the signature must no longer match.
	tampered := token[:len(token)-4] + "AAAA"
	if _, err := iss.Verify(tampered); err == nil {
		t.Fatal("a tampered token must not verify")
	}
}

func TestVerifyRejectsATokenFromAnotherSecret(t *testing.T) {
	mine := testIssuer(t)
	theirs, err := NewIssuer("a-different-secret-also-32-bytes!", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	token, err := theirs.Issue("account-123", "x@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mine.Verify(token); err == nil {
		t.Fatal("a token signed with another secret must not verify")
	}
}

func TestVerifyRejectsAnExpiredToken(t *testing.T) {
	iss, err := NewIssuer("a-secret-at-least-32-bytes-long!!", -time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, err := iss.Issue("account-123", "x@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := iss.Verify(token); err == nil {
		t.Fatal("an expired token must not verify")
	}
}

func TestShortSecretsAreRefused(t *testing.T) {
	if _, err := NewIssuer("too-short", time.Hour); err == nil {
		t.Fatal("a short signing secret must be refused, not silently accepted")
	}
}
