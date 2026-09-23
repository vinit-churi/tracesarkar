// Command ingest runs the Phase 0 collectors.
//
//	ingest migrate            apply database migrations
//	ingest run [source ...]   snapshot every schedulable source, or those named
//	ingest status             per-endpoint collection health
//	ingest changes [--since]  recent changes to published works data
//
// Exposure tier comes from TRACESARKAR_TIER and defaults to personal (D044).
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/config"
	"github.com/vinit-churi/tracesarkar/internal/ingest"
	"github.com/vinit-churi/tracesarkar/internal/notify"
	"github.com/vinit-churi/tracesarkar/internal/sources"
	"github.com/vinit-churi/tracesarkar/internal/store"
	"github.com/vinit-churi/tracesarkar/internal/works"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var err error
	switch os.Args[1] {
	case "migrate":
		err = runMigrate(ctx)
	case "run":
		err = runSnapshot(ctx, os.Args[2:])
	case "status":
		err = runStatus(ctx)
	case "changes":
		err = runChanges(ctx, os.Args[2:])
	case "-h", "--help", "help":
		usage()
		return
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		log.Error("command failed", "command", os.Args[1], "error", err.Error())
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `tracesarkar ingest — Phase 0 collectors

  ingest migrate             apply database migrations
  ingest run [source ...]    snapshot schedulable sources
  ingest status              per-endpoint collection health
  ingest changes [--since d] recent changes (default 7 days)

Configuration comes from .env or the environment:
  R2_BUCKET_URL, R2_BUCKET_NAME, R2_ACCESS_KEY, R2_SECRET_ACCESS_KEY,
  POSTGRESQL_CONNECTION, POSTGRES_CA_PATH, NOTIFY_WEBHOOK_URL, TRACESARKAR_TIER
`)
}

func envFile() string {
	if v := os.Getenv("TRACESARKAR_ENV_FILE"); v != "" {
		return v
	}
	return ".env"
}

func load(ctx context.Context) (config.Config, *store.DB, func(), error) {
	cfg, err := config.Load(envFile())
	if err != nil {
		return config.Config{}, nil, func() {}, err
	}
	pool, err := store.Connect(ctx, cfg.Postgres.URL, cfg.Postgres.CAPath)
	if err != nil {
		return config.Config{}, nil, func() {}, err
	}
	return cfg, store.NewDB(pool), pool.Close, nil
}

func runMigrate(ctx context.Context) error {
	cfg, err := config.Load(envFile())
	if err != nil {
		return err
	}
	pool, err := store.Connect(ctx, cfg.Postgres.URL, cfg.Postgres.CAPath)
	if err != nil {
		return err
	}
	defer pool.Close()

	applied, err := store.Migrate(ctx, pool)
	if err != nil {
		return err
	}
	if len(applied) == 0 {
		slog.Info("schema is up to date")
		return nil
	}
	for _, m := range applied {
		slog.Info("migration applied", "version", m.Version, "name", m.Name)
	}
	return nil
}

func runSnapshot(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	registerPath := fs.String("register", "data/sources.yaml", "path to the source register")
	if err := fs.Parse(args); err != nil {
		return err
	}
	only := map[string]bool{}
	for _, name := range fs.Args() {
		only[name] = true
	}

	cfg, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	reg, err := sources.Load(*registerPath)
	if err != nil {
		return err
	}
	if err := db.SyncRegister(ctx, reg); err != nil {
		return err
	}

	arch, err := archive.New(archive.Options{
		Endpoint:  cfg.R2.Endpoint,
		Bucket:    cfg.R2.Bucket,
		Region:    cfg.R2.Region,
		AccessKey: cfg.R2.AccessKey,
		Secret:    cfg.R2.Secret,
	})
	if err != nil {
		return err
	}

	alerts := notify.New(cfg.NotifyWebhook, nil, slog.Default())
	runner := &ingest.Runner{
		Fetcher: ingest.NewFetcher(nil, ingest.DefaultUserAgent),
		Archive: arch,
		Store:   db,
		Log:     slog.Default(),
	}

	jobs, skipped := ingest.BuildJobs(reg, cfg.Tier)
	for id, reason := range skipped {
		slog.Info("source not scheduled", "source", id, "reason", reason, "tier", cfg.Tier)
	}

	var failures int
	for _, job := range jobs {
		if len(only) > 0 && !only[job.SourceID] {
			continue
		}
		started := time.Now().UTC()
		summary, runErr := runner.Run(ctx, job)

		message := ""
		if runErr != nil {
			message = runErr.Error()
			failures++
			// Log the detail here: the summary alone cannot be debugged.
			slog.Error("run reported errors", "source", job.SourceID, "error", message)
		}
		if err := db.MarkSourceRun(ctx, job.SourceID, started, message); err != nil {
			slog.Warn("could not update source status", "source", job.SourceID, "error", err.Error())
		}

		slog.Info("run finished",
			"source", job.SourceID,
			"changed_endpoints", summary.Changed,
			"unchanged_endpoints", summary.Unchanged,
			"failed_endpoints", summary.Failed,
			"changes", len(summary.Changes))

		if err := sendSummary(ctx, alerts, job.SourceID, summary, runErr); err != nil {
			slog.Warn("alert delivery failed", "source", job.SourceID, "error", err.Error())
		}
	}

	if failures > 0 {
		return fmt.Errorf("%d source(s) reported errors", failures)
	}
	return nil
}

// sendSummary alerts on changes and on failures, and stays quiet otherwise.
func sendSummary(ctx context.Context, alerts notify.Notifier, sourceID string, s ingest.Summary, runErr error) error {
	if runErr == nil && len(s.Changes) == 0 {
		return nil
	}
	title := fmt.Sprintf("%s: %d change(s)", sourceID, len(s.Changes))
	if runErr != nil {
		title = fmt.Sprintf("%s: %d change(s), %d endpoint(s) failed", sourceID, len(s.Changes), s.Failed)
	}

	lines := make([]string, 0, len(s.Changes)+1)
	for _, c := range s.Changes {
		lines = append(lines, describeChange(c))
	}
	if runErr != nil {
		lines = append(lines, "error: "+runErr.Error())
	}
	return alerts.Send(ctx, notify.Message{Title: title, Lines: lines})
}

func describeChange(c works.Change) string {
	switch c.Kind {
	case works.KindChanged:
		return fmt.Sprintf("%s: %s %v → %v", c.NaturalKey, c.Field, c.Old, c.New)
	default:
		return fmt.Sprintf("%s: %s", c.NaturalKey, c.Kind)
	}
}

func runStatus(ctx context.Context) error {
	_, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	rows, err := db.Status(ctx)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Println("no fetches recorded yet")
		return nil
	}

	fmt.Printf("%-16s %-18s %-22s %8s %8s %8s\n", "SOURCE", "ENDPOINT", "LAST FETCH (UTC)", "RECORDS", "RUNS", "FAILS")
	for _, r := range rows {
		last := "—"
		if r.LastFetchAt != nil {
			last = r.LastFetchAt.UTC().Format("2006-01-02 15:04")
		}
		fmt.Printf("%-16s %-18s %-22s %8d %8d %8d\n",
			truncate(r.SourceID, 16), truncate(r.Endpoint, 18), last, r.Records, r.Attempts, r.Failures)
	}
	return nil
}

func runChanges(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("changes", flag.ExitOnError)
	days := fs.Int("since", 7, "days of history to show")
	limit := fs.Int("limit", 100, "maximum rows")
	if err := fs.Parse(args); err != nil {
		return err
	}

	_, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	rows, err := db.ChangesSince(ctx, time.Now().UTC().AddDate(0, 0, -*days), *limit)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Printf("no changes in the last %d day(s)\n", *days)
		return nil
	}
	for _, r := range rows {
		when := r.ObservedAt.UTC().Format("2006-01-02 15:04")
		switch r.Kind {
		case works.KindChanged:
			fmt.Printf("%s  %s/%s  %s\n    %s: %s → %s\n",
				when, r.SourceID, r.Endpoint, r.NaturalKey, r.Field, trimJSON(r.Old), trimJSON(r.New))
		default:
			fmt.Printf("%s  %s/%s  %s  [%s]\n", when, r.SourceID, r.Endpoint, r.NaturalKey, r.Kind)
		}
	}
	return nil
}

func trimJSON(s string) string {
	return strings.Trim(s, `"`)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
