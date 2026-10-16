package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"log/slog"
	"strings"
	"testing"
)

// A header is terminated by a newline, so a newline inside a value ends the
// header and starts another one. An address or a subject that reaches this
// package from a database is untrusted for exactly that reason: a `To` of
// "a@b.c\r\nBcc: everyone@example.org" is a mail sent to everyone.
func TestAHeaderValueCanNeverCarryANewline(t *testing.T) {
	for _, bad := range []string{
		"a@b.c\r\nBcc: everyone@example.org",
		"a@b.c\nBcc: everyone@example.org",
		"a@b.c\rX: y",
	} {
		if _, err := render("from@tracesarkar.org", "", Message{
			To: bad, Subject: "s", Body: "b",
		}); err == nil {
			t.Errorf("rendered a message with To %q", bad)
		}
		if _, err := render("from@tracesarkar.org", "", Message{
			To: "a@b.c", Subject: bad, Body: "b",
		}); err == nil {
			t.Errorf("rendered a message with a subject carrying a newline")
		}
		if _, err := render(bad, "", Message{To: "a@b.c", Subject: "s", Body: "b"}); err == nil {
			t.Errorf("rendered a message with From %q", bad)
		}
	}
}

func TestARenderedMessageIsAWellFormedMail(t *testing.T) {
	out, err := render("noreply@tracesarkar.org", "TraceSarkar", Message{
		To:      "person@example.org",
		Subject: "Reset your TraceSarkar password",
		Body:    "Open this link:\n\nhttps://tracesarkar.org/account/reset?token=abc\n",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	s := string(out)

	// Every line ends CRLF, which SMTP requires and a bare \n breaks.
	if bytes.Contains(bytes.ReplaceAll(out, []byte("\r\n"), nil), []byte("\n")) {
		t.Error("the message contains a bare newline")
	}
	for _, want := range []string{
		"From: TraceSarkar <noreply@tracesarkar.org>",
		"To: person@example.org",
		"Subject: Reset your TraceSarkar password",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"Content-Transfer-Encoding: base64",
		"Date: ",
		"Message-ID: <",
		// An automated mail that does not say so gets auto-replies and
		// out-of-office messages back at a noreply address.
		"Auto-Submitted: auto-generated",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("the message has no %q:\n%s", want, s)
		}
	}

	// The body round-trips, which a hand-rolled encoder is exactly the thing
	// to get wrong.
	_, rest, ok := strings.Cut(s, "\r\n\r\n")
	if !ok {
		t.Fatal("no blank line between the headers and the body")
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(rest, "\r\n", ""))
	if err != nil {
		t.Fatalf("the body is not valid base64: %v", err)
	}
	if !strings.Contains(string(decoded), "https://tracesarkar.org/account/reset?token=abc") {
		t.Errorf("the decoded body lost the link:\n%s", decoded)
	}
}

// Devanagari in a subject or a body has to survive, or the vernacular copy
// this platform is built around cannot be mailed.
func TestAMessageSurvivesDevanagari(t *testing.T) {
	out, err := render("noreply@tracesarkar.org", "ट्रेससरकार", Message{
		To: "person@example.org", Subject: "तुमचा पासवर्ड", Body: "हा दुवा उघडा.\n",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	s := string(out)
	// A non-ASCII header is encoded rather than sent raw, which some servers
	// reject outright.
	if strings.Contains(s, "तुमचा पासवर्ड") {
		t.Error("a non-ASCII subject was sent unencoded")
	}
	if !strings.Contains(s, "=?utf-8?") && !strings.Contains(s, "=?UTF-8?") {
		t.Errorf("the subject is not MIME-encoded:\n%s", s)
	}
	_, rest, _ := strings.Cut(s, "\r\n\r\n")
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(rest, "\r\n", ""))
	if err != nil {
		t.Fatalf("the body is not valid base64: %v", err)
	}
	if !strings.Contains(string(decoded), "हा दुवा उघडा.") {
		t.Errorf("the Devanagari body did not survive: %q", decoded)
	}
}

// The log sender exists so the personal tier works with no mail provider at
// all. It writes a credential to the logs, which is a real cost, so it has to
// say so every single time rather than looking like ordinary mail.
func TestTheLogSenderSaysItIsWritingACredentialToTheLogs(t *testing.T) {
	var buf bytes.Buffer
	s := NewLogSender(slog.New(slog.NewTextHandler(&buf, nil)))
	if err := s.Send(context.Background(), Message{
		To: "person@example.org", Subject: "Reset", Body: "open https://x/y?token=abc",
	}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "https://x/y?token=abc") {
		t.Errorf("the link is not in the log, so the sender is useless:\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "credential") {
		t.Errorf("the log line does not say it is a credential:\n%s", out)
	}
	if !strings.Contains(s.Describe(), "log") {
		t.Errorf("Describe is %q, which does not warn a reader of a startup log", s.Describe())
	}
}

func TestAnSMTPSenderRefusesAnIncompleteConfiguration(t *testing.T) {
	cases := []struct {
		name string
		o    SMTPOptions
	}{
		{"no host", SMTPOptions{From: "a@b.c"}},
		{"no from", SMTPOptions{Host: "smtp.example.org"}},
		{"a from that is not an address", SMTPOptions{Host: "smtp.example.org", From: "nobody"}},
		{"an unknown TLS mode", SMTPOptions{
			Host: "smtp.example.org", From: "a@b.c", TLS: "sideways"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := NewSMTPSender(c.o); err == nil {
				t.Error("an incomplete configuration was accepted")
			}
		})
	}
	s, err := NewSMTPSender(SMTPOptions{
		Host: "smtp.example.org", From: "noreply@tracesarkar.org", FromName: "TraceSarkar",
	})
	if err != nil {
		t.Fatalf("NewSMTPSender: %v", err)
	}
	if !strings.Contains(s.Describe(), "smtp.example.org") {
		t.Errorf("Describe is %q and does not name the host", s.Describe())
	}
}
