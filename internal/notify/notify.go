// Package notify delivers alerts about collection: changes observed, runs that
// failed, and nothing else. At the personal tier there is no volume budget, but
// messages are still bundled per source per run.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Message is one alert.
type Message struct {
	Title string
	Lines []string
}

// Text renders the message, keeping at most max lines and counting the rest.
func (m Message) Text(max int) string {
	var b strings.Builder
	b.WriteString(m.Title)
	shown := m.Lines
	if max > 0 && len(shown) > max {
		shown = shown[:max]
	}
	for _, line := range shown {
		b.WriteString("\n")
		b.WriteString(line)
	}
	if remaining := len(m.Lines) - len(shown); remaining > 0 {
		fmt.Fprintf(&b, "\n… and %d more", remaining)
	}
	return b.String()
}

// Notifier delivers messages.
type Notifier interface {
	Send(ctx context.Context, m Message) error
}

// Logger writes alerts to the structured log. It is the fallback when no
// channel is configured, and it never fails.
type Logger struct{ log *slog.Logger }

// NewLogger returns a logging notifier.
func NewLogger(log *slog.Logger) *Logger {
	if log == nil {
		log = slog.Default()
	}
	return &Logger{log: log}
}

// Send logs the message.
func (l *Logger) Send(_ context.Context, m Message) error {
	l.log.Info("alert", "title", m.Title, "detail", m.Text(20))
	return nil
}

// Webhook posts alerts as JSON, which suits ntfy, Slack-style hooks and most
// self-hosted receivers.
type Webhook struct {
	url  string
	http *http.Client
}

// NewWebhook returns a webhook notifier.
func NewWebhook(url string, client *http.Client) *Webhook {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &Webhook{url: url, http: client}
}

// Send posts the message.
func (w *Webhook) Send(ctx context.Context, m Message) error {
	var payload bytes.Buffer
	enc := json.NewEncoder(&payload)
	// Keep arrows and quotes readable; these payloads are read by humans.
	enc.SetEscapeHTML(false)
	if err := enc.Encode(map[string]any{
		"title":   m.Title,
		"message": m.Text(20),
		"text":    m.Text(20),
	}); err != nil {
		return fmt.Errorf("encode alert: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(payload.Bytes()))
	if err != nil {
		return fmt.Errorf("build alert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.http.Do(req)
	if err != nil {
		return fmt.Errorf("send alert: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return fmt.Errorf("alert endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}
	return nil
}

// New returns a webhook notifier when a URL is configured, and the logging
// notifier otherwise.
func New(webhookURL string, client *http.Client, log *slog.Logger) Notifier {
	if strings.TrimSpace(webhookURL) == "" {
		return NewLogger(log)
	}
	return NewWebhook(webhookURL, client)
}
