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
	"syscall"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/api"
	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/config"
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

Configuration comes from .env or the environment:
  R2_*, POSTGRESQL_CONNECTION, POSTGRES_CA_PATH, TRACESARKAR_TIER,
  API_TOKEN     the bearer token the field kit sends
  API_ACCOUNT   the handle captures are attributed to (default: field-kit)
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

	pool, err := store.Connect(ctx, cfg.Postgres.URL, cfg.Postgres.CAPath)
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

	srv, err := api.New(api.Options{
		Reports: db,
		Media:   blobs,
		Token:   token,
		Account: accountID,
		Log:     slog.Default(),
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

	slog.Info("listening", "addr", *addr, "tier", cfg.Tier, "account", handle)
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
