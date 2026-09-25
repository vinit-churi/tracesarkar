// Package auth handles identity: passwords, session tokens, and Google
// sign-in.
//
// ADR 0011 chose phone-only identity for citizen reporting. Email and Google
// are additional routes for people who hold an account rather than file a
// report — see ADR 0017.
package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// MinPasswordLength is deliberately a length rule and nothing else. Composition
// rules ("one symbol, one digit") push people towards predictable passwords and
// buy little; length is what helps.
const MinPasswordLength = 10

// HashPassword returns a bcrypt hash. The cost is the library default, which
// tracks hardware over time.
func HashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	return hash, nil
}

// VerifyPassword reports whether the password matches the hash. It takes the
// same time whether the account exists or not, because bcrypt comparison is
// constant-time for a given cost.
func VerifyPassword(hash []byte, password string) bool {
	return bcrypt.CompareHashAndPassword(hash, []byte(password)) == nil
}

// CheckPasswordPolicy rejects passwords that are too short to be worth hashing.
func CheckPasswordPolicy(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("a password is required")
	}
	if len(password) < MinPasswordLength {
		return fmt.Errorf("a password needs at least %d characters", MinPasswordLength)
	}
	return nil
}

// NormaliseEmail lowercases and trims, so one person cannot hold two accounts
// that differ only in case.
func NormaliseEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// CheckEmail validates the address shape. It does not confirm the address
// exists; verification is a separate step.
func CheckEmail(email string) error {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return errors.New("an email address is required")
	}
	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return fmt.Errorf("that does not look like an email address")
	}
	at := strings.LastIndex(addr.Address, "@")
	if at <= 0 || at == len(addr.Address)-1 || !strings.Contains(addr.Address[at:], ".") {
		return errors.New("that does not look like an email address")
	}
	return nil
}
