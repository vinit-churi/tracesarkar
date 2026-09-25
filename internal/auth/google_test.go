package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"
)

// signedGoogleToken mints an RS256 token the way Google would, so the verifier
// can be tested without the network.
func signedGoogleToken(t *testing.T, key *rsa.PrivateKey, kid string, claims map[string]any) string {
	t.Helper()
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT", "kid": kid})
	payload, _ := json.Marshal(claims)
	input := base64url(header) + "." + base64url(payload)

	digest := sha256.Sum256([]byte(input))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto_SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return input + "." + base64url(sig)
}

func testKeySet(t *testing.T, kid string, key *rsa.PrivateKey) KeyFunc {
	t.Helper()
	return func(_ context.Context, wantKid string) (*rsa.PublicKey, error) {
		if wantKid != kid {
			return nil, errNoSuchKey
		}
		return &key.PublicKey, nil
	}
}

func validClaims() map[string]any {
	return map[string]any{
		"iss":            "https://accounts.google.com",
		"aud":            "client-id-123.apps.googleusercontent.com",
		"sub":            "google-subject-9876",
		"email":          "Person@Example.com",
		"email_verified": true,
		"name":           "A Person",
		"exp":            time.Now().Add(time.Hour).Unix(),
		"iat":            time.Now().Add(-time.Minute).Unix(),
	}
}

func TestGoogleVerifierAcceptsAValidToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewGoogleVerifier("client-id-123.apps.googleusercontent.com", testKeySet(t, "kid-1", key))

	token := signedGoogleToken(t, key, "kid-1", validClaims())
	identity, err := v.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}

	if identity.Subject != "google-subject-9876" {
		t.Errorf("subject: got %q", identity.Subject)
	}
	// The email is normalised here so callers cannot create two accounts that
	// differ only in case.
	if identity.Email != "person@example.com" {
		t.Errorf("email should be normalised: got %q", identity.Email)
	}
	if !identity.EmailVerified {
		t.Error("email_verified should carry through")
	}
	if identity.Name != "A Person" {
		t.Errorf("name: got %q", identity.Name)
	}
}

func TestGoogleVerifierRejectsAnotherApplicationsToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewGoogleVerifier("our-client-id", testKeySet(t, "kid-1", key))

	claims := validClaims() // aud is a different client
	token := signedGoogleToken(t, key, "kid-1", claims)

	if _, err := v.Verify(context.Background(), token); err == nil {
		t.Fatal("a token minted for another application must be refused")
	}
}

func TestGoogleVerifierRejectsAnotherIssuer(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewGoogleVerifier("client-id-123.apps.googleusercontent.com", testKeySet(t, "kid-1", key))

	claims := validClaims()
	claims["iss"] = "https://evil.example.com"
	token := signedGoogleToken(t, key, "kid-1", claims)

	if _, err := v.Verify(context.Background(), token); err == nil {
		t.Fatal("a token from another issuer must be refused")
	}
}

func TestGoogleVerifierRejectsAnExpiredToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewGoogleVerifier("client-id-123.apps.googleusercontent.com", testKeySet(t, "kid-1", key))

	claims := validClaims()
	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	token := signedGoogleToken(t, key, "kid-1", claims)

	if _, err := v.Verify(context.Background(), token); err == nil {
		t.Fatal("an expired token must be refused")
	}
}

func TestGoogleVerifierRejectsATokenSignedByTheWrongKey(t *testing.T) {
	realKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	attackerKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewGoogleVerifier("client-id-123.apps.googleusercontent.com", testKeySet(t, "kid-1", realKey))

	token := signedGoogleToken(t, attackerKey, "kid-1", validClaims())
	if _, err := v.Verify(context.Background(), token); err == nil {
		t.Fatal("a token signed by another key must be refused")
	}
}

func TestGoogleVerifierRejectsAnUnsignedToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewGoogleVerifier("client-id-123.apps.googleusercontent.com", testKeySet(t, "kid-1", key))

	// The classic attack: alg "none" with the signature removed.
	header, _ := json.Marshal(map[string]string{"alg": "none", "typ": "JWT", "kid": "kid-1"})
	payload, _ := json.Marshal(validClaims())
	token := base64url(header) + "." + base64url(payload) + "."

	if _, err := v.Verify(context.Background(), token); err == nil {
		t.Fatal(`a token with alg "none" must be refused`)
	}
}

func TestParseJWKSExtractsRSAKeys(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	n := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes())
	doc := `{"keys":[{"kty":"RSA","kid":"abc","use":"sig","alg":"RS256","n":"` + n + `","e":"` + e + `"}]}`

	keys, err := parseJWKS([]byte(doc))
	if err != nil {
		t.Fatalf("parseJWKS: %v", err)
	}
	got, ok := keys["abc"]
	if !ok {
		t.Fatal("key abc missing")
	}
	if got.N.Cmp(key.PublicKey.N) != 0 || got.E != key.PublicKey.E {
		t.Error("the parsed key does not match the original")
	}
}

func TestGoogleVerifierNeedsAClientID(t *testing.T) {
	v := NewGoogleVerifier("", nil)
	if _, err := v.Verify(context.Background(), "anything"); err == nil {
		t.Fatal("without a configured client id, Google sign-in must refuse rather than accept anything")
	}
	if !strings.Contains(v.Unconfigured(), "GOOGLE_CLIENT_ID") {
		t.Errorf("the reason should name the setting: %q", v.Unconfigured())
	}
}
