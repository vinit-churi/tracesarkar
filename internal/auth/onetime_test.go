package auth

import (
	"strings"
	"testing"
	"time"
)

// The token goes in a link in an email. It is the only thing standing between
// a stranger and somebody's account, so it has to be unguessable and it has
// to be unguessable without anybody remembering to make it so.
func TestAOneTimeTokenIsLongAndNeverRepeats(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		ot, err := NewOneTime()
		if err != nil {
			t.Fatalf("NewOneTime: %v", err)
		}
		if len(ot.Token) < 40 {
			t.Fatalf("token is %d characters, too short to be unguessable", len(ot.Token))
		}
		// It travels in a URL and gets copied out of mail clients that break
		// lines and strip characters.
		if strings.ContainsAny(ot.Token, "+/=&?# \n") {
			t.Fatalf("token %q is not safe in a URL", ot.Token)
		}
		if seen[ot.Token] {
			t.Fatalf("token repeated after %d draws", i)
		}
		seen[ot.Token] = true
	}
}

// The database holds the hash. A reset link that leaks from a backup is a
// password reset for everybody who had one pending.
func TestOnlyTheHashIsEverStored(t *testing.T) {
	ot, err := NewOneTime()
	if err != nil {
		t.Fatalf("NewOneTime: %v", err)
	}
	if len(ot.Hash) == 0 {
		t.Fatal("no hash was produced")
	}
	if strings.Contains(string(ot.Hash), ot.Token) {
		t.Error("the hash contains the token")
	}
	if !OneTimeMatches(ot.Hash, ot.Token) {
		t.Error("a token does not match its own hash")
	}
	other, _ := NewOneTime()
	if OneTimeMatches(ot.Hash, other.Token) {
		t.Error("a different token matched")
	}
	if OneTimeMatches(ot.Hash, "") {
		t.Error("an empty token matched")
	}
	if OneTimeMatches(nil, ot.Token) {
		t.Error("a token matched an absent hash")
	}
	if got := HashOneTime(ot.Token); string(got) != string(ot.Hash) {
		t.Error("HashOneTime does not reproduce the stored hash")
	}
}

// A reset window measured in days is a window an attacker has days to use. A
// verification window measured in minutes is a user who has to ask twice.
// They are different problems and must not share a number.
func TestEachPurposeHasItsOwnWindow(t *testing.T) {
	if !PasswordReset.Valid() || !EmailVerification.Valid() {
		t.Fatal("a purpose the platform issues reports itself invalid")
	}
	if Purpose("do_anything").Valid() {
		t.Error("an unknown purpose was accepted")
	}
	reset, verify := PasswordReset.TTL(), EmailVerification.TTL()
	if reset <= 0 || verify <= 0 {
		t.Fatalf("a purpose has no window: reset %v, verify %v", reset, verify)
	}
	if reset > 2*time.Hour {
		t.Errorf("the password reset window is %v; a stolen link should not be "+
			"useful for that long", reset)
	}
	if verify <= reset {
		t.Errorf("verification (%v) is no longer than reset (%v); verifying an "+
			"address is not urgent and resetting a password is", verify, reset)
	}
	if Purpose("do_anything").TTL() != 0 {
		t.Error("an unknown purpose has a usable window")
	}
}
