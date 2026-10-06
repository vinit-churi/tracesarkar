package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/classify"
)

// runClassify classifies captures that have not been classified yet.
//
// Separate from capture on purpose: a photograph is durable before anything is
// attempted on it, so this can fail, be retried, or be run against a different
// model later without a citizen ever losing a report.
func runClassify(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("classify", flag.ExitOnError)
	limit := fs.Int("limit", 25, "how many captures to classify in this pass")
	model := fs.String("model", envOr("CLASSIFY_MODEL", ""), "model id")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	key := os.Getenv("CLASSIFY_API_KEY")
	if key == "" {
		return fmt.Errorf("CLASSIFY_API_KEY is not set; classification needs a provider")
	}
	if *model == "" {
		return fmt.Errorf("CLASSIFY_MODEL is not set")
	}

	mode := classify.SchemaJSON
	if os.Getenv("CLASSIFY_SCHEMA_MODE") == string(classify.SchemaJSONObject) {
		mode = classify.SchemaJSONObject
	}
	kind := classify.CitizenPhotograph
	if os.Getenv("CLASSIFY_DATA_KIND") == string(classify.OwnTestData) {
		kind = classify.OwnTestData
	}

	prompt, err := classify.LoadPrompt("road_defect_v1")
	if err != nil {
		return err
	}

	blobs, err := archive.New(archive.Options{
		Endpoint: cfg.R2.Endpoint, Bucket: cfg.R2.Bucket, Region: cfg.R2.Region,
		AccessKey: cfg.R2.AccessKey, Secret: cfg.R2.Secret,
	})
	if err != nil {
		return err
	}

	// Coverage comes from the database, so a category becomes routable in a
	// ward the moment a department is recorded for it there.
	byWard, err := db.CoverageByWard(ctx)
	if err != nil {
		return err
	}
	coverage := make(map[string]classify.Coverage, len(byWard))
	for ward, categories := range byWard {
		coverage[ward] = classify.CoverageFor(categories...)
	}
	if len(coverage) == 0 {
		slog.Warn("no department mappings; every classification will be not_yet_covered")
	}

	client := classify.NewGateway(classify.GatewayOptions{
		BaseURL:    envOr("CLASSIFY_BASE_URL", "https://api.deepseek.com/v1"),
		APIKey:     key,
		Model:      *model,
		SchemaMode: mode,
		DataKind:   kind,
		Timeout:    120 * time.Second,
	})

	done, err := classify.Run(ctx, classify.RunnerOptions{
		Store: db, Blobs: blobs, Classifier: client,
		CoverageByWard: coverage,
		Model:          *model, Prompt: prompt.Version,
		Limit: *limit, Log: slog.Default(),
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "\nclassified %d capture(s) with %s\n", done, *model)
	return nil
}

// envOr reads a setting with a fallback.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
