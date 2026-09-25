package auth

import (
	"strings"
	"testing"
)

func TestHashPasswordThenVerify(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if strings.Contains(string(hash), "correct horse") {
		t.Fatal("the hash must not contain the password")
	}
	if !VerifyPassword(hash, "correct horse battery staple") {
		t.Error("the right password should verify")
	}
	if VerifyPassword(hash, "wrong password") {
		t.Error("the wrong password must not verify")
	}
}

func TestHashesAreSaltedSoTwoUsersWithOnePasswordDiffer(t *testing.T) {
	a, err := HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	b, err := HashPassword("same-password")
	if err != nil {
		t.Fatal(err)
	}
	if string(a) == string(b) {
		t.Error("identical passwords must not produce identical hashes")
	}
}

func TestPasswordPolicy(t *testing.T) {
	tests := []struct {
		name     string
		password string
		ok       bool
	}{
		{"long enough", "a-reasonable-passphrase", true},
		{"exactly the minimum", strings.Repeat("x", 10), true},
		{"too short", "short", false},
		{"empty", "", false},
		{"whitespace only", "          ", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordPolicy(tt.password)
			if tt.ok && err != nil {
				t.Errorf("expected acceptable, got %v", err)
			}
			if !tt.ok && err == nil {
				t.Error("expected rejection")
			}
		})
	}
}

func TestNormaliseEmail(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Person@Example.COM", "person@example.com"},
		{"  spaced@example.com  ", "spaced@example.com"},
		{"already@lower.in", "already@lower.in"},
	}
	for _, tt := range tests {
		if got := NormaliseEmail(tt.in); got != tt.want {
			t.Errorf("NormaliseEmail(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestValidEmail(t *testing.T) {
	for _, good := range []string{"a@b.co", "person.name+tag@example.org"} {
		if err := CheckEmail(good); err != nil {
			t.Errorf("%q should be acceptable: %v", good, err)
		}
	}
	for _, bad := range []string{"", "not-an-email", "@example.com", "person@"} {
		if err := CheckEmail(bad); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}
