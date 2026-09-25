package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// crypto_SHA256 is named so tests can reference the same hash the verifier uses.
const crypto_SHA256 = crypto.SHA256

// GoogleCertsURL publishes the public keys Google signs ID tokens with.
const GoogleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"

var errNoSuchKey = errors.New("no such signing key")

// googleIssuers are the two spellings Google uses.
var googleIssuers = map[string]bool{
	"accounts.google.com":         true,
	"https://accounts.google.com": true,
}

// Identity is who Google says the person is.
type Identity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// KeyFunc returns the public key for a key id.
type KeyFunc func(ctx context.Context, kid string) (*rsa.PublicKey, error)

// GoogleVerifier checks Google ID tokens: the signature against Google's
// published keys, and then that the token was minted for this application.
type GoogleVerifier struct {
	clientID string
	keys     KeyFunc
}

// NewGoogleVerifier returns a verifier. An empty client id leaves Google
// sign-in switched off rather than accepting whatever arrives.
func NewGoogleVerifier(clientID string, keys KeyFunc) *GoogleVerifier {
	return &GoogleVerifier{clientID: clientID, keys: keys}
}

// Unconfigured explains why Google sign-in is unavailable, for an error a
// person can act on.
func (v *GoogleVerifier) Unconfigured() string {
	if v.clientID == "" {
		return "Google sign-in is not configured: set GOOGLE_CLIENT_ID"
	}
	return ""
}

// Verify checks an ID token and returns the identity inside it.
func (v *GoogleVerifier) Verify(ctx context.Context, token string) (Identity, error) {
	if reason := v.Unconfigured(); reason != "" {
		return Identity{}, errors.New(reason)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[2] == "" {
		return Identity{}, errors.New("token is malformed")
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	rawHeader, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || json.Unmarshal(rawHeader, &header) != nil {
		return Identity{}, errors.New("token header is not readable")
	}
	// Only RS256. Accepting "none", or an HMAC algorithm with the public key as
	// the secret, is the classic way these verifiers are broken.
	if header.Alg != "RS256" {
		return Identity{}, fmt.Errorf("unexpected signing algorithm %q", header.Alg)
	}

	key, err := v.keys(ctx, header.Kid)
	if err != nil {
		return Identity{}, fmt.Errorf("signing key: %w", err)
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Identity{}, errors.New("token signature is malformed")
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto_SHA256, digest[:], signature); err != nil {
		return Identity{}, errors.New("token signature does not match Google's key")
	}

	var claims struct {
		Issuer        string `json:"iss"`
		Audience      string `json:"aud"`
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Expiry        int64  `json:"exp"`
	}
	rawClaims, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || json.Unmarshal(rawClaims, &claims) != nil {
		return Identity{}, errors.New("token payload is not readable")
	}

	if !googleIssuers[claims.Issuer] {
		return Identity{}, fmt.Errorf("token was issued by %q, not Google", claims.Issuer)
	}
	// Without this check any Google user of any application could sign in here.
	if claims.Audience != v.clientID {
		return Identity{}, errors.New("token was issued for a different application")
	}
	if claims.Expiry > 0 && time.Now().Unix() >= claims.Expiry {
		return Identity{}, errors.New("token has expired")
	}
	if claims.Subject == "" {
		return Identity{}, errors.New("token names no subject")
	}

	return Identity{
		Subject:       claims.Subject,
		Email:         NormaliseEmail(claims.Email),
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
	}, nil
}

// GoogleKeys fetches and caches Google's signing keys.
type GoogleKeys struct {
	url    string
	http   *http.Client
	mu     sync.Mutex
	keys   map[string]*rsa.PublicKey
	loaded time.Time
	ttl    time.Duration
}

// NewGoogleKeys returns a cache over Google's published certificates.
func NewGoogleKeys(client *http.Client) *GoogleKeys {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &GoogleKeys{url: GoogleCertsURL, http: client, ttl: time.Hour}
}

// Key returns the public key for a key id, refreshing the cache when the id is
// unknown — Google rotates keys, and a rotation should not mean an outage.
func (g *GoogleKeys) Key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	g.mu.Lock()
	cached, ok := g.keys[kid]
	fresh := time.Since(g.loaded) < g.ttl
	g.mu.Unlock()
	if ok && fresh {
		return cached, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.url, nil)
	if err != nil {
		return nil, fmt.Errorf("build key request: %w", err)
	}
	resp, err := g.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch Google keys: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("fetch Google keys: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read Google keys: %w", err)
	}
	keys, err := parseJWKS(body)
	if err != nil {
		return nil, err
	}

	g.mu.Lock()
	g.keys = keys
	g.loaded = time.Now()
	g.mu.Unlock()

	key, ok := keys[kid]
	if !ok {
		return nil, errNoSuchKey
	}
	return key, nil
}

func parseJWKS(body []byte) (map[string]*rsa.PublicKey, error) {
	var doc struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("parse key set: %w", err)
	}

	out := map[string]*rsa.PublicKey{}
	for _, k := range doc.Keys {
		if k.Kty != "RSA" || k.Kid == "" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		out[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(new(big.Int).SetBytes(eBytes).Int64()),
		}
	}
	if len(out) == 0 {
		return nil, errors.New("key set contains no usable RSA keys")
	}
	return out, nil
}
