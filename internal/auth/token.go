package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// MinSecretBytes is the shortest signing secret worth having. Anything less is
// brute-forceable, and a configuration mistake here is silent.
const MinSecretBytes = 32

// Claims are what a session token carries. Deliberately little: an account id
// and an email, nothing a client should not already know.
type Claims struct {
	AccountID string `json:"sub"`
	Email     string `json:"email,omitempty"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Issuer    string `json:"iss"`
}

// Issuer mints and verifies session tokens. The format is a JWT with HMAC-SHA256,
// which every client library understands.
type Issuer struct {
	secret []byte
	ttl    time.Duration
}

// NewIssuer refuses a secret too short to be safe.
func NewIssuer(secret string, ttl time.Duration) (*Issuer, error) {
	if len(secret) < MinSecretBytes {
		return nil, fmt.Errorf("signing secret must be at least %d bytes", MinSecretBytes)
	}
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &Issuer{secret: []byte(secret), ttl: ttl}, nil
}

const tokenIssuer = "tracesarkar"

// Issue returns a signed token for an account.
func (i *Issuer) Issue(accountID, email string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		AccountID: accountID,
		Email:     email,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(i.ttl).Unix(),
		Issuer:    tokenIssuer,
	}

	header := base64url([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("encode claims: %w", err)
	}
	signingInput := header + "." + base64url(payload)
	return signingInput + "." + base64url(i.sign(signingInput)), nil
}

// Verify checks the signature and the expiry, and returns the claims.
func (i *Issuer) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("token is malformed")
	}
	signingInput := parts[0] + "." + parts[1]

	expected := i.sign(signingInput)
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Claims{}, errors.New("token signature is malformed")
	}
	if !hmac.Equal(expected, actual) {
		return Claims{}, errors.New("token signature does not match")
	}

	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, errors.New("token payload is malformed")
	}
	var claims Claims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return Claims{}, errors.New("token payload is not readable")
	}
	if claims.ExpiresAt > 0 && time.Now().UTC().Unix() >= claims.ExpiresAt {
		return Claims{}, errors.New("token has expired")
	}
	if claims.AccountID == "" {
		return Claims{}, errors.New("token names no account")
	}
	return claims, nil
}

func (i *Issuer) sign(input string) []byte {
	mac := hmac.New(sha256.New, i.secret)
	mac.Write([]byte(input))
	return mac.Sum(nil)
}

func base64url(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
