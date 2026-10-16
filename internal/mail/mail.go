// Package mail delivers the two messages an account needs: a password reset
// link and an address-verification link.
//
// It is plain text only. No HTML part, no tracking pixel, no open or click
// tracking, no third-party template service. A reset mail is read once by one
// person, and the only thing it has to do is carry a link that works in every
// client including the ones that strip markup. There is also nothing here
// worth measuring: knowing who opened a password reset is knowing something
// about somebody's account that this platform has no use for.
package mail

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/auth"
)

// Message is one mail.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender delivers a message, or says why it could not.
type Sender interface {
	Send(ctx context.Context, m Message) error
	// Describe names the route, for a startup log line that makes it obvious
	// whether real mail is going out.
	Describe() string
}

// paths are where each link lands in the client. The purpose is in the path
// rather than in a query parameter so that a link pasted into the wrong form
// cannot be mistaken for the other kind.
var paths = map[auth.Purpose]string{
	auth.PasswordReset:     "account/reset",
	auth.EmailVerification: "account/verify",
}

// Link builds the URL a person follows, from the public base URL of the
// client and the one-time token.
//
// The base URL comes from configuration, which is why the scheme is checked:
// a misconfigured base must fail here rather than put `javascript:` in front
// of a credential in somebody's inbox.
func Link(base string, purpose auth.Purpose, token string) (string, error) {
	path, ok := paths[purpose]
	if !ok || !purpose.Valid() {
		return "", fmt.Errorf("mail: no link for purpose %q", purpose)
	}
	if strings.TrimSpace(token) == "" {
		return "", fmt.Errorf("mail: no token to put in a %s link", purpose)
	}
	base = strings.TrimSpace(base)
	if base == "" {
		return "", fmt.Errorf("mail: no public base URL is configured, so a %s "+
			"link cannot be built", purpose)
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mail: base URL %q is not a URL: %w", base, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("mail: base URL %q is not http or https", base)
	}
	if u.Host == "" {
		return "", fmt.Errorf("mail: base URL %q names no host", base)
	}
	u.Path = "/" + strings.TrimPrefix(strings.TrimSuffix(u.Path, "/")+"/"+path, "/")
	u.RawQuery = url.Values{"token": []string{token}}.Encode()
	return u.String(), nil
}

// Compose writes the message for a purpose.
//
// The copy is written here rather than in a template file because it is two
// messages and both of them are load-bearing: a reset mail is the one message
// a stranger can cause to be sent to somebody else's inbox, so it has to tell
// an unexpecting reader that ignoring it is safe and that nothing has changed
// yet. A test asserts both messages say so.
//
// English only, and that is a recorded limitation rather than an oversight. A
// reset is requested by somebody who cannot sign in, so the platform does not
// know their language at the moment it has to write to them. Carrying a
// language preference on the account would fix it and is not built.
func Compose(purpose auth.Purpose, to, link string) (Message, error) {
	window := humanWindow(purpose.TTL())
	switch purpose {
	case auth.PasswordReset:
		return Message{
			To:      to,
			Subject: "Reset your TraceSarkar password",
			Body: "Somebody asked to reset the password for the TraceSarkar account\n" +
				"registered to this address. To set a new one, open this link:\n\n" +
				link + "\n\n" +
				"The link works once and expires in " + window + ".\n\n" +
				"If you did not ask for this, you can ignore this message. Nothing\n" +
				"has changed, your password still works, and nobody has been given\n" +
				"access to your account or your reports.\n\n" +
				"TraceSarkar never asks for your password by email, and never asks\n" +
				"you to reply to a message with it.\n",
		}, nil
	case auth.EmailVerification:
		return Message{
			To:      to,
			Subject: "Confirm your email address for TraceSarkar",
			Body: "To confirm that this address reaches you, open this link:\n\n" +
				link + "\n\n" +
				"The link works once and expires in " + window + ".\n\n" +
				"If you did not ask for this, you can ignore this message. An\n" +
				"unconfirmed address changes nothing about an account that already\n" +
				"exists, and confirming is not a sign-in.\n",
		}, nil
	default:
		return Message{}, fmt.Errorf("mail: no message for purpose %q", purpose)
	}
}

// humanWindow writes a duration the way the mail should read it. Deliberately
// coarse: "about a day" is what a reader needs, and "24h0m0s" is not.
func humanWindow(d time.Duration) string {
	switch {
	case d <= 0:
		return "a short time"
	case d < time.Hour:
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	case d < 2*time.Hour:
		return "one hour"
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours", int(d.Hours()))
	default:
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
}
