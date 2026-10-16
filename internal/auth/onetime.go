package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"
)

// Purpose is what a one-time credential may be used for.
//
// It is part of the stored row and part of the lookup, so a credential minted
// to verify an address cannot be presented to reset a password. Without it the
// two would be the same secret with two meanings, and the weaker window would
// set the strength of both.
type Purpose string

const (
	// PasswordReset lets somebody who cannot sign in set a new password.
	PasswordReset Purpose = "password_reset"
	// EmailVerification proves the address on an account reaches the person
	// holding it.
	EmailVerification Purpose = "email_verification"
)

// oneTimeBytes is 256 bits of entropy. The token is the whole credential —
// there is no second factor behind it and no rate limit that can save an
// account from a guessable one.
const oneTimeBytes = 32

// OneTime is a single-use credential delivered out of band.
//
// The two fields go to different places and that separation is the point:
// Token is put in a link and never persisted, Hash is persisted and never
// sent. A database that leaks cannot be used to reset anybody's password.
type OneTime struct {
	// Token goes in the link. URL-safe, unpadded, so it survives a mail
	// client that wraps lines and a user who copies it by hand.
	Token string
	// Hash is what the row holds.
	Hash []byte
}

// NewOneTime mints a credential.
func NewOneTime() (OneTime, error) {
	raw := make([]byte, oneTimeBytes)
	if _, err := rand.Read(raw); err != nil {
		return OneTime{}, fmt.Errorf("read random bytes: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return OneTime{Token: token, Hash: HashOneTime(token)}, nil
}

// HashOneTime is what gets stored.
//
// SHA-256 rather than bcrypt, deliberately. bcrypt's cost exists to slow down
// guessing a secret a person chose, and this secret is 256 random bits:
// nothing can guess it, so the work factor would buy latency and no safety.
// What matters instead is that the stored form is one-way and that the
// comparison is constant-time.
func HashOneTime(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// OneTimeMatches compares a presented token against a stored hash.
func OneTimeMatches(hash []byte, token string) bool {
	if len(hash) != sha256.Size || token == "" {
		return false
	}
	return subtle.ConstantTimeCompare(hash, HashOneTime(token)) == 1
}

// TTL is how long a credential for this purpose stays usable.
//
// The two windows are different because the two risks are. A reset link is a
// live key to an account, so it expires inside an hour. A verification link
// proves an address and grants nothing, so making it short only means the
// person has to ask twice.
func (p Purpose) TTL() time.Duration {
	switch p {
	case PasswordReset:
		return time.Hour
	case EmailVerification:
		return 24 * time.Hour
	default:
		return 0
	}
}

// Valid reports whether this is a purpose the platform issues. Checked before
// a row is written, so a typo in a caller cannot create a credential whose
// window is zero and whose meaning is nothing.
func (p Purpose) Valid() bool { return p.TTL() > 0 }
