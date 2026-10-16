// Command api serves the platform's HTTP surface.
//
// At Phase 0 this is personal-tier only: a health check, the capture endpoint,
// and the field kit that posts to it. v0.1 opens the same endpoint to the
// public behind phone verification.
//
//	api serve [--addr :8080]
//
// Configuration comes from .env or the environment; see .env.example. The
// bearer token is API_TOKEN, and the field kit asks for it once.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/api"
	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/auth"
	"github.com/vinit-churi/tracesarkar/internal/config"
	"github.com/vinit-churi/tracesarkar/internal/mail"
	"github.com/vinit-churi/tracesarkar/internal/polish"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var err error
	switch command {
	case "serve":
		err = serve(ctx, os.Args[min(2, len(os.Args)):])
	case "health":
		// Used by the container's HEALTHCHECK; the runtime image has no curl.
		err = healthCheck(ctx, envOr("HEALTH_URL", "http://127.0.0.1:8080"))
	case "-h", "--help", "help":
		usage()
		return
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		slog.Error("api failed", "error", err.Error())
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `tracesarkar api — the HTTP surface

  api serve [--addr :8080]
  api health                 ask a running server whether it is well

Configuration comes from .env or the environment:
  R2_*, POSTGRESQL_CONNECTION, POSTGRES_CA_PATH, TRACESARKAR_TIER,
  API_TOKEN       the bearer token the field kit sends
  API_ACCOUNT     the handle captures are attributed to (default: field-kit)
  AUTH_SECRET     signing secret for session tokens (32+ bytes, required)
  GOOGLE_CLIENT_ID  enables Google sign-in when set
  ALLOWED_ORIGINS   comma-separated browser origins for the web client
`)
}

func serve(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", envOr("API_ADDR", ":8080"), "address to listen on")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(envFile())
	if err != nil {
		return err
	}
	token := os.Getenv("API_TOKEN")
	if token == "" {
		return errors.New("API_TOKEN is not set; refusing to serve an unauthenticated capture endpoint")
	}

	caPath, err := cfg.Postgres.CAFile()
	if err != nil {
		return err
	}
	pool, err := store.Connect(ctx, cfg.Postgres.URL, caPath)
	if err != nil {
		return err
	}
	defer pool.Close()
	db := store.NewDB(pool)

	if _, err := store.Migrate(ctx, pool); err != nil {
		return err
	}

	handle := envOr("API_ACCOUNT", "field-kit")
	accountID, err := db.EnsureAccount(ctx, handle)
	if err != nil {
		return err
	}

	blobs, err := archive.New(archive.Options{
		Endpoint:  cfg.R2.Endpoint,
		Bucket:    cfg.R2.Bucket,
		Region:    cfg.R2.Region,
		AccessKey: cfg.R2.AccessKey,
		Secret:    cfg.R2.Secret,
	})
	if err != nil {
		return err
	}

	// Sessions. The secret is required: without it the server would either run
	// unauthenticated or invent a secret that changes on every restart, signing
	// people out silently.
	secret := os.Getenv("AUTH_SECRET")
	if secret == "" {
		return errors.New("AUTH_SECRET is not set; generate one with: openssl rand -base64 48")
	}
	issuer, err := auth.NewIssuer(secret, 30*24*time.Hour)
	if err != nil {
		return err
	}

	// Google sign-in is optional: without a client id the endpoint refuses
	// cleanly and email/password still works.
	var google *auth.GoogleVerifier
	if clientID := os.Getenv("GOOGLE_CLIENT_ID"); clientID != "" {
		google = auth.NewGoogleVerifier(clientID, auth.NewGoogleKeys(nil).Key)
		slog.Info("google sign-in enabled")
	} else {
		google = auth.NewGoogleVerifier("", nil)
		slog.Info("google sign-in disabled", "reason", "GOOGLE_CLIENT_ID is not set")
	}

	// The share kit's rewording model. Off unless SHARE_POLISH is set, and a
	// hard startup error when it is set without a model or an endpoint —
	// there is deliberately no default model here. A model chosen by default
	// is a model nobody decided on, and CLAUDE.md requires that choice to be
	// a recorded decision rather than an implementation detail.
	var polisher polish.Polisher
	if os.Getenv("SHARE_POLISH") != "" {
		baseURL, model := os.Getenv("POLISH_BASE_URL"), os.Getenv("POLISH_MODEL")
		if baseURL == "" || model == "" {
			return errors.New("SHARE_POLISH is set but POLISH_BASE_URL or POLISH_MODEL is not; " +
				"name the model explicitly so the choice is on the record")
		}
		g, err := polish.NewGateway(polish.GatewayOptions{
			BaseURL:    baseURL,
			APIKey:     os.Getenv("POLISH_API_KEY"),
			Model:      model,
			PromptName: envOr("POLISH_PROMPT", "share_polish_v1"),
			SchemaMode: polish.SchemaMode(envOr("POLISH_SCHEMA_MODE", string(polish.SchemaJSON))),
		})
		if err != nil {
			return err
		}
		// A share kit is a public GET already cached for five minutes at the
		// edge; paying for one answer on every request is the cheapest
		// mistake available.
		polisher = polish.Cached(g, 30*time.Minute, 1024)
		slog.Info("share kit rewording enabled", "model", model,
			"prompt", envOr("POLISH_PROMPT", "share_polish_v1"))
	} else {
		slog.Info("share kit rewording disabled", "reason", "SHARE_POLISH is not set")
	}

	// How a reset or verification link gets to a person. Three states, and
	// the default is the honest one: with no MAIL_MODE the endpoints refuse
	// rather than mint a credential nobody will ever receive.
	var sender mail.Sender
	switch mode := os.Getenv("MAIL_MODE"); mode {
	case "smtp":
		port, err := strconv.Atoi(envOr("SMTP_PORT", "0"))
		if err != nil {
			return fmt.Errorf("SMTP_PORT is not a number: %w", err)
		}
		sender, err = mail.NewSMTPSender(mail.SMTPOptions{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     port,
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("MAIL_FROM"),
			FromName: envOr("MAIL_FROM_NAME", "TraceSarkar"),
			TLS:      os.Getenv("SMTP_TLS"),
		})
		if err != nil {
			return err
		}
	case "log":
		// The personal tier: under ADR 0017 the only account is the
		// operator's own, and the operator can read their own link out of the
		// log. It writes a one-time credential to the log, which is why it
		// has to be asked for by name.
		sender = mail.NewLogSender(slog.Default())
		slog.Warn("mail is going to the log, not to a mailbox",
			"why", "MAIL_MODE=log", "cost", "one-time credentials appear in the log")
	case "":
	default:
		return fmt.Errorf("MAIL_MODE is %q; use smtp, log, or leave it unset", mode)
	}

	// Where this service lives, as a fact rather than as whatever hostname
	// answered a request. It is the base of every permalink burned into a
	// shared card, so without it the platform generates links from the
	// request's Host header — which a caller of a public endpoint chooses.
	publicBaseURL := strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL"))
	if publicBaseURL == "" {
		slog.Warn("PUBLIC_BASE_URL is not set",
			"effect", "permalinks follow the request's Host header",
			"cost", "a caller can choose the link in a generated post, and a "+
				"forwarded post dies if this service's hostname changes")
	}
	if sender != nil {
		if publicBaseURL == "" {
			return errors.New("MAIL_MODE is set but PUBLIC_BASE_URL is not; a reset " +
				"link cannot be built without knowing where the client lives")
		}
		// Fail at startup rather than on the first reset somebody needs.
		if _, err := mail.Link(publicBaseURL, auth.PasswordReset, "startup-check"); err != nil {
			return err
		}
		slog.Info("password resets and email verification enabled",
			"mail", sender.Describe(), "links", publicBaseURL)
	} else {
		slog.Info("password resets and email verification disabled",
			"reason", "MAIL_MODE is not set")
	}

	origins := splitList(envOr("ALLOWED_ORIGINS",
		"http://localhost:8080,http://localhost:5000,http://127.0.0.1:5000"))

	srv, err := api.New(api.Options{
		Reports:        db,
		Media:          blobs,
		Accounts:       db,
		Reviews:        db,
		Labels:         db,
		Truth:          db,
		Wards:          db,
		Issues:         db,
		PublicIssues:   db,
		Handles:        db,
		Evidence:       db,
		Polish:         polisher,
		Credentials:    db,
		Mail:           sender,
		PublicBaseURL:  publicBaseURL,
		Blobs:          blobs,
		Details:        db,
		Issuer:         issuer,
		Google:         google,
		AllowedOrigins: origins,
		Token:          token,
		Account:        accountID,
		Log:            slog.Default(),
	})
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		// Uploads come from phones on Indian mobile networks: generous.
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		IdleTimeout:  2 * time.Minute,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	slog.Info("listening", "addr", *addr, "tier", cfg.Tier, "account", handle,
		"origins", origins, "google", google.Unconfigured() == "")
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	slog.Info("stopped")
	return nil
}

func envFile() string {
	if v := os.Getenv("TRACESARKAR_ENV_FILE"); v != "" {
		return v
	}
	return ".env"
}

// splitList turns a comma-separated setting into a slice.
func splitList(v string) []string {
	var out []string
	for _, part := range strings.Split(v, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
