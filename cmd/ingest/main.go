// Command ingest runs the Phase 0 collectors.
//
//	ingest migrate            apply database migrations
//	ingest run [source ...]   snapshot every schedulable source, or those named
//	ingest status             per-endpoint collection health
//	ingest changes [--since]  recent changes to published works data
//	ingest watch              archive newly published Government Resolutions
//	ingest all                snapshot then watch, for a single scheduled trigger
//
// Exposure tier comes from TRACESARKAR_TIER and defaults to personal (D044).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/config"
	"github.com/vinit-churi/tracesarkar/internal/ingest"
	"github.com/vinit-churi/tracesarkar/internal/notify"
	"github.com/vinit-churi/tracesarkar/internal/sources"
	"github.com/vinit-churi/tracesarkar/internal/store"
	"github.com/vinit-churi/tracesarkar/internal/watch"
	"github.com/vinit-churi/tracesarkar/internal/works"
)

// version is stamped by the release build.
var version = "dev"

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
	case "watch":
		err = runWatch(ctx, os.Args[2:])
	case "all":
		err = runAll(ctx, os.Args[2:])
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
  ingest watch [--limit n]   archive newly published Government Resolutions
  ingest all                 snapshot, then watch — one nightly invocation

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

// withRunLog records that a command ran, and how it ended, so that a night which
// fails on a machine that then powers itself off is still diagnosable.
func withRunLog(ctx context.Context, db *store.DB, cfg config.Config, command string,
	body func() (map[string]any, error)) error {

	host, _ := os.Hostname()
	if host == "" {
		host = "unknown"
	}
	runID, startErr := db.StartRun(ctx, store.RunStart{
		Command: command, Host: host, Tier: cfg.Tier, Version: version,
	})
	if startErr != nil {
		// Losing the record is not a reason to skip the collection.
		slog.Warn("could not record the start of this run", "error", startErr.Error())
	}

	detail, err := body()

	if runID != "" {
		message := ""
		if err != nil {
			message = err.Error()
		}
		// A fresh context: the run's own may already be cancelled.
		finishCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if finishErr := db.FinishRun(finishCtx, runID, store.RunFinish{
			OK: err == nil, Error: message, Detail: detail,
		}); finishErr != nil {
			slog.Warn("could not record the end of this run", "error", finishErr.Error())
		}
	}
	return err
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

	// Setup is inside the run record too, so a failure before the first fetch is
	// still visible afterwards.
	return withRunLog(ctx, db, cfg, "run", func() (map[string]any, error) {
		reg, err := sources.Load(*registerPath)
		if err != nil {
			return nil, err
		}
		if err := db.SyncRegister(ctx, reg); err != nil {
			return nil, err
		}

		arch, err := archive.New(archive.Options{
			Endpoint:  cfg.R2.Endpoint,
			Bucket:    cfg.R2.Bucket,
			Region:    cfg.R2.Region,
			AccessKey: cfg.R2.AccessKey,
			Secret:    cfg.R2.Secret,
		})
		if err != nil {
			return nil, err
		}

		alerts := notify.New(cfg.NotifyWebhook, nil, slog.Default())

		jobs, skipped := ingest.BuildJobs(reg, cfg.Tier)
		for id, reason := range skipped {
			slog.Info("source not scheduled", "source", id, "reason", reason, "tier", cfg.Tier)
		}

		return snapshotJobs(ctx, db, cfg, arch, alerts, jobs, only)
	})
}

func snapshotJobs(ctx context.Context, db *store.DB, cfg config.Config, arch *archive.Client,
	alerts notify.Notifier, jobs []ingest.Job, only map[string]bool) (map[string]any, error) {

	runner := &ingest.Runner{
		Fetcher: ingest.NewFetcher(nil, ingest.DefaultUserAgent),
		Archive: arch,
		Store:   db,
		Log:     slog.Default(),
	}

	detail := map[string]any{}
	var changedTotal, unchangedTotal, failedTotal, changeCount int

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

		changedTotal += summary.Changed
		unchangedTotal += summary.Unchanged
		failedTotal += summary.Failed
		changeCount += len(summary.Changes)

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

	detail["changed_endpoints"] = changedTotal
	detail["unchanged_endpoints"] = unchangedTotal
	detail["failed_endpoints"] = failedTotal
	detail["changes"] = changeCount

	if failures > 0 {
		return detail, fmt.Errorf("%d source(s) reported errors", failures)
	}
	return detail, nil
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

func runWatch(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("watch", flag.ExitOnError)
	limit := fs.Int("limit", 25, "how many new resolutions to fetch in one run")
	rows := fs.Int("rows", 100, "how many recent items to examine")
	registerPath := fs.String("register", "data/sources.yaml", "path to the source register")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	// Everything after this point is inside the run record, including the setup
	// that can fail: a missing register used to leave no trace at all.
	var result watch.Result
	runErr := withRunLog(ctx, db, cfg, "watch", func() (map[string]any, error) {
		arch, err := archive.New(archive.Options{
			Endpoint:  cfg.R2.Endpoint,
			Bucket:    cfg.R2.Bucket,
			Region:    cfg.R2.Region,
			AccessKey: cfg.R2.AccessKey,
			Secret:    cfg.R2.Secret,
		})
		if err != nil {
			return nil, err
		}

		reg, err := sources.Load(*registerPath)
		if err != nil {
			return nil, err
		}
		src, ok := reg.Get("maha_gr_archive")
		if !ok {
			return nil, fmt.Errorf("source maha_gr_archive is not in %s", *registerPath)
		}

		watcher := watch.NewGRWatcher(ingest.NewFetcher(nil, ingest.DefaultUserAgent), arch, db)
		watcher.Limit = *limit
		watcher.Rows = *rows
		watcher.Log = slog.Default()
		watcher.Source = src
		watcher.Tier = cfg.Tier

		result, err = watcher.Run(ctx)
		return map[string]any{
			"examined": result.Seen, "new": result.New,
			"flagged": len(result.Hits), "skipped": result.SkippedReason,
		}, err
	})

	if result.SkippedReason != "" {
		slog.Info("watch not scheduled", "source", "maha_gr_archive",
			"reason", result.SkippedReason, "tier", cfg.Tier)
		return nil
	}
	slog.Info("watch finished",
		"source", "maha_gr_archive",
		"examined", result.Seen, "new", result.New, "flagged", len(result.Hits))

	if len(result.Hits) > 0 {
		alerts := notify.New(cfg.NotifyWebhook, nil, slog.Default())
		lines := make([]string, 0, len(result.Hits))
		for _, h := range result.Hits {
			issued := "date unknown"
			if h.IssuedOn != nil {
				issued = h.IssuedOn.Format("2 Jan 2006")
			}
			lines = append(lines, fmt.Sprintf("GR %s (%s): %s", h.Sanketank, issued, strings.Join(h.Keywords, ", ")))
		}
		if err := alerts.Send(ctx, notify.Message{
			Title: fmt.Sprintf("%d resolution(s) mention a watched term", len(result.Hits)),
			Lines: lines,
		}); err != nil {
			slog.Warn("alert delivery failed", "error", err.Error())
		}
	}
	return runErr
}

// runAll is what the nightly trigger calls: both collectors in one invocation,
// so the schedule needs no argument overrides and the invoking identity needs
// nothing beyond permission to start the job.
//
// The watch runs even when the snapshot failed: they read different sources,
// and one being unreachable is no reason to skip the other.
func runAll(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("all", flag.ExitOnError)
	registerPath := fs.String("register", "data/sources.yaml", "path to the source register")
	if err := fs.Parse(args); err != nil {
		return err
	}
	// Each subcommand takes its own flags, so pass only what both understand.
	shared := []string{"--register", *registerPath}

	var errs []error
	if err := runSnapshot(ctx, shared); err != nil {
		errs = append(errs, fmt.Errorf("snapshot: %w", err))
	}
	if err := runWatch(ctx, shared); err != nil {
		errs = append(errs, fmt.Errorf("watch: %w", err))
	}
	return errors.Join(errs...)
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

	return printRecentRuns(ctx, db, 8)
}

// printRecentRuns is how a night that failed on a machine that has since powered
// itself off is diagnosed.
func printRecentRuns(ctx context.Context, db *store.DB, limit int) error {
	runs, err := db.RecentRuns(ctx, limit)
	if err != nil {
		return err
	}
	if len(runs) == 0 {
		return nil
	}

	fmt.Printf("\nrecent runs\n")
	fmt.Printf("%-19s %-8s %-20s %-9s %s\n", "STARTED (UTC)", "COMMAND", "HOST", "OUTCOME", "DETAIL")
	for _, r := range runs {
		outcome := "running"
		switch {
		case r.FinishedAt == nil && time.Since(r.StartedAt) > time.Hour:
			outcome = "no return"
		case r.FinishedAt == nil:
			outcome = "running"
		case r.OK:
			outcome = "ok"
		default:
			outcome = "FAILED"
		}
		detail := r.Error
		if detail == "" {
			detail = summariseDetail(r.Detail)
		}
		if len(detail) > 64 {
			detail = detail[:61] + "..."
		}
		fmt.Printf("%-19s %-8s %-20s %-9s %s\n",
			r.StartedAt.UTC().Format("2006-01-02 15:04"), truncate(r.Command, 8),
			truncate(r.Host, 20), outcome, detail)
	}
	return nil
}

func summariseDetail(detail map[string]any) string {
	if len(detail) == 0 {
		return ""
	}
	keys := make([]string, 0, len(detail))
	for k := range detail {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := detail[k]
		if v == nil || v == "" {
			continue
		}
		if f, ok := v.(float64); ok {
			if f == 0 {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s=%g", k, f))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, " ")
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
