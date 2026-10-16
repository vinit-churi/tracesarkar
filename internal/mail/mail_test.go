package mail

import (
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/auth"
)

func TestALinkCarriesThePurposeAndTheToken(t *testing.T) {
	cases := []struct {
		name    string
		base    string
		purpose auth.Purpose
		token   string
		wantErr bool
		// contains is checked against the finished URL.
		contains []string
	}{
		{
			name: "a reset link", base: "https://tracesarkar.org",
			purpose: auth.PasswordReset, token: "abc-123_xyz",
			contains: []string{"https://tracesarkar.org/", "abc-123_xyz"},
		},
		{
			name: "a verification link", base: "https://tracesarkar.org/",
			purpose: auth.EmailVerification, token: "tok",
			contains: []string{"https://tracesarkar.org/", "tok"},
		},
		{
			name: "no base URL", base: "", purpose: auth.PasswordReset,
			token: "tok", wantErr: true,
		},
		{
			name: "no token", base: "https://tracesarkar.org",
			purpose: auth.PasswordReset, token: "", wantErr: true,
		},
		{
			name: "a purpose the platform does not issue", base: "https://tracesarkar.org",
			purpose: auth.Purpose("anything"), token: "tok", wantErr: true,
		},
		{
			// A base URL from configuration is not a place to accept
			// javascript: or a mailto:.
			name: "a base URL that is not http", base: "javascript:alert(1)",
			purpose: auth.PasswordReset, token: "tok", wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Link(c.base, c.purpose, c.token)
			if c.wantErr {
				if err == nil {
					t.Fatalf("want an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Link: %v", err)
			}
			for _, want := range c.contains {
				if !strings.Contains(got, want) {
					t.Errorf("%q does not contain %q", got, want)
				}
			}
			// One slash, not two, however the base was written.
			if strings.Contains(strings.TrimPrefix(got, "https://"), "//") {
				t.Errorf("%q has a doubled slash", got)
			}
		})
	}
}

// Two things the copy must do, because a reset mail is the one message a
// stranger can cause to be sent to somebody else's inbox.
func TestTheCopySaysWhatToDoIfYouDidNotAskForIt(t *testing.T) {
	for _, purpose := range []auth.Purpose{auth.PasswordReset, auth.EmailVerification} {
		m, err := Compose(purpose, "person@example.org", "https://tracesarkar.org/x?t=tok")
		if err != nil {
			t.Fatalf("Compose(%s): %v", purpose, err)
		}
		if m.To != "person@example.org" {
			t.Errorf("To is %q", m.To)
		}
		if strings.TrimSpace(m.Subject) == "" {
			t.Error("no subject")
		}
		if !strings.Contains(m.Body, "https://tracesarkar.org/x?t=tok") {
			t.Errorf("the body does not carry the link:\n%s", m.Body)
		}
		// Somebody who did not ask has to be told that ignoring it is safe
		// and that nothing has changed yet.
		lower := strings.ToLower(m.Body)
		if !strings.Contains(lower, "did not") {
			t.Errorf("%s: the body does not address somebody who did not ask:\n%s",
				purpose, m.Body)
		}
		// The window has to be stated, or the link looks broken when it
		// expires.
		if !strings.Contains(lower, "hour") {
			t.Errorf("%s: the body does not say how long the link lasts:\n%s",
				purpose, m.Body)
		}
		// Never an accusation and never a password.
		if strings.Contains(lower, "your password is") {
			t.Error("the body appears to carry a password")
		}
	}
	if _, err := Compose(auth.Purpose("anything"), "a@b.c", "https://x/y"); err == nil {
		t.Error("composed a message for a purpose the platform does not issue")
	}
}

// The reset mail and the verification mail must not read the same: somebody
// who gets the wrong one needs to be able to tell.
func TestTheTwoMessagesAreDistinguishable(t *testing.T) {
	reset, err := Compose(auth.PasswordReset, "a@b.c", "https://x/y")
	if err != nil {
		t.Fatal(err)
	}
	verify, err := Compose(auth.EmailVerification, "a@b.c", "https://x/y")
	if err != nil {
		t.Fatal(err)
	}
	if reset.Subject == verify.Subject || reset.Body == verify.Body {
		t.Error("the two messages read the same")
	}
}
