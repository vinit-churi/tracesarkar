package mail

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// render turns a Message into the bytes SMTP carries.
//
// Written by hand rather than with a library, because the whole of it is six
// headers and a base64 body, and the one thing that actually has to be right
// is the part a library would not protect: a header value is terminated by a
// newline, so a newline inside a value ends that header and begins another.
// An address or a subject arriving from the database is untrusted for exactly
// that reason — `To: a@b.c\r\nBcc: everyone` is a mail sent to everyone — and
// this function refuses rather than strips, because stripping silently sends
// a mail somebody did not intend.
func render(from, fromName string, m Message) ([]byte, error) {
	for name, v := range map[string]string{"From": from, "To": m.To, "Subject": m.Subject,
		"the sender name": fromName} {
		if strings.ContainsAny(v, "\r\n") {
			return nil, fmt.Errorf("mail: %s contains a newline and cannot be a header", name)
		}
	}
	if strings.TrimSpace(m.To) == "" {
		return nil, fmt.Errorf("mail: no recipient")
	}
	if _, err := mail.ParseAddress(m.To); err != nil {
		return nil, fmt.Errorf("mail: recipient %q is not an address: %w", m.To, err)
	}

	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return nil, fmt.Errorf("mail: read random bytes: %w", err)
	}
	host := from
	if _, h, ok := strings.Cut(from, "@"); ok {
		host = h
	}

	fromHeader := from
	if strings.TrimSpace(fromName) != "" {
		fromHeader = mime.QEncoding.Encode("utf-8", fromName) + " <" + from + ">"
	}

	var b strings.Builder
	w := func(k, v string) { b.WriteString(k + ": " + v + "\r\n") }
	w("From", fromHeader)
	w("To", m.To)
	w("Subject", mime.QEncoding.Encode("utf-8", m.Subject))
	w("Date", time.Now().UTC().Format(time.RFC1123Z))
	w("Message-ID", "<"+hex.EncodeToString(id)+"@"+host+">")
	w("MIME-Version", "1.0")
	w("Content-Type", "text/plain; charset=utf-8")
	// Base64 rather than 8bit: the body is Devanagari as often as not, and a
	// server that does not advertise 8BITMIME will mangle it.
	w("Content-Transfer-Encoding", "base64")
	// So that an out-of-office reply does not arrive at a noreply address,
	// and so mail filters can tell this is not a person writing.
	w("Auto-Submitted", "auto-generated")
	b.WriteString("\r\n")

	encoded := base64.StdEncoding.EncodeToString([]byte(m.Body))
	for len(encoded) > 76 {
		b.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	if encoded != "" {
		b.WriteString(encoded + "\r\n")
	}
	return []byte(b.String()), nil
}

// LogSender writes the message to the log instead of sending it.
//
// It exists so the personal tier works with no mail provider at all: under
// [ADR 0017] the only account is the operator's own, and the operator can read
// their own reset link out of the log. The cost is real and is not hidden — a
// one-time credential in a log file is a credential anybody who can read logs
// can use — so every line says so, and nothing selects this sender by
// default.
//
// [ADR 0017]: ../../docs/04-adr/0017-email-and-google-identity-before-public-tier.md
type LogSender struct{ log *slog.Logger }

// NewLogSender returns a sender that logs.
func NewLogSender(log *slog.Logger) *LogSender {
	if log == nil {
		log = slog.Default()
	}
	return &LogSender{log: log}
}

func (l *LogSender) Describe() string {
	return "the log (no mail is sent; one-time credentials are written to the log)"
}

func (l *LogSender) Send(_ context.Context, m Message) error {
	l.log.Warn("writing a one-time credential to the log instead of sending mail",
		"to", m.To, "subject", m.Subject, "body", m.Body,
		"why", "no mail route is configured (see MAIL_* in .env.example)")
	return nil
}

// SMTPOptions configures a real mail route.
type SMTPOptions struct {
	Host     string
	Port     int
	Username string
	Password string
	// From is the envelope and header sender. It has to be an address the
	// domain's SPF record authorises, or the mail is spam.
	From     string
	FromName string
	// TLS is "starttls" (the default, port 587) or "implicit" (port 465).
	// There is no "none": a password and a one-time credential do not travel
	// in the clear.
	TLS     string
	Timeout time.Duration
}

// SMTPSender delivers through an SMTP server.
type SMTPSender struct{ opts SMTPOptions }

// NewSMTPSender validates the configuration. Every check here is a mistake
// that would otherwise be discovered by a person who never received a reset
// mail and had no way to find out why.
func NewSMTPSender(o SMTPOptions) (*SMTPSender, error) {
	if strings.TrimSpace(o.Host) == "" {
		return nil, fmt.Errorf("mail: no SMTP host")
	}
	if strings.TrimSpace(o.From) == "" {
		return nil, fmt.Errorf("mail: no sender address")
	}
	if _, err := mail.ParseAddress(o.From); err != nil {
		return nil, fmt.Errorf("mail: sender %q is not an address: %w", o.From, err)
	}
	switch o.TLS {
	case "":
		o.TLS = "starttls"
	case "starttls", "implicit":
	default:
		return nil, fmt.Errorf("mail: unknown TLS mode %q; use starttls or implicit", o.TLS)
	}
	if o.Port == 0 {
		o.Port = 587
		if o.TLS == "implicit" {
			o.Port = 465
		}
	}
	if o.Timeout == 0 {
		o.Timeout = 20 * time.Second
	}
	return &SMTPSender{opts: o}, nil
}

func (s *SMTPSender) Describe() string {
	return fmt.Sprintf("smtp %s:%d (%s)", s.opts.Host, s.opts.Port, s.opts.TLS)
}

// Send delivers one message.
//
// net/smtp rather than a dependency: this is one connection, one AUTH, one
// recipient. What it does not do for us is the deadline, so the connection is
// dialled with one and the context is honoured by closing underneath.
func (s *SMTPSender) Send(ctx context.Context, m Message) error {
	body, err := render(s.opts.From, s.opts.FromName, m)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(s.opts.Host, fmt.Sprint(s.opts.Port))
	dialer := &net.Dialer{Timeout: s.opts.Timeout}
	tlsConf := &tls.Config{ServerName: s.opts.Host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	if s.opts.TLS == "implicit" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConf)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("mail: dial %s: %w", addr, err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(s.opts.Timeout))
	}

	c, err := smtp.NewClient(conn, s.opts.Host)
	if err != nil {
		return fmt.Errorf("mail: SMTP handshake: %w", err)
	}
	defer c.Close()

	if s.opts.TLS == "starttls" {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			// Not a thing to shrug at: the password and the credential in the
			// body would both cross the network in the clear.
			return fmt.Errorf("mail: %s does not offer STARTTLS", addr)
		}
		if err := c.StartTLS(tlsConf); err != nil {
			return fmt.Errorf("mail: STARTTLS: %w", err)
		}
	}
	if s.opts.Username != "" {
		if err := c.Auth(smtp.PlainAuth("", s.opts.Username, s.opts.Password, s.opts.Host)); err != nil {
			return fmt.Errorf("mail: authenticate: %w", err)
		}
	}
	if err := c.Mail(s.opts.From); err != nil {
		return fmt.Errorf("mail: MAIL FROM: %w", err)
	}
	if err := c.Rcpt(m.To); err != nil {
		return fmt.Errorf("mail: RCPT TO: %w", err)
	}
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("mail: DATA: %w", err)
	}
	if _, err := wc.Write(body); err != nil {
		return fmt.Errorf("mail: write body: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("mail: finish body: %w", err)
	}
	return c.Quit()
}
