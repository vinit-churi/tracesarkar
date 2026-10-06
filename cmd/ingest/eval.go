package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/classify"
)

// runEval scores the classifier against the photographs a person labelled.
//
// This is what makes system 3's bar checkable. Everything else about the
// classifier can be argued; this produces a number, and the failures it prints
// are the only part anyone can act on.
func runEval(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	limit := fs.Int("limit", 1000, "how many labelled captures to score")
	model := fs.String("model", envOr("CLASSIFY_MODEL", ""), "model id")
	out := fs.String("out", "", "write the report as JSON to this path")
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
		return fmt.Errorf("CLASSIFY_API_KEY is not set; evaluation needs a provider")
	}
	if *model == "" {
		return fmt.Errorf("CLASSIFY_MODEL is not set")
	}

	entries, err := db.EvalSet(ctx, *limit)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		// Refused rather than reported as a perfect score over nothing.
		return fmt.Errorf("no labelled captures yet; label some at /label/ first")
	}

	blobs, err := archive.New(archive.Options{
		Endpoint: cfg.R2.Endpoint, Bucket: cfg.R2.Bucket, Region: cfg.R2.Region,
		AccessKey: cfg.R2.AccessKey, Secret: cfg.R2.Secret,
	})
	if err != nil {
		return err
	}

	cases := make([]classify.EvalCase, 0, len(entries))
	var missing int
	for _, e := range entries {
		image, err := blobs.Get(ctx, e.ArchiveKey)
		if err != nil {
			// A label whose photograph has gone is not a case. Counted and
			// named rather than silently dropped, because a shrinking eval set
			// flatters every number computed from it.
			missing++
			fmt.Fprintf(os.Stderr, "skipping %s: %v\n", e.ReportID, err)
			continue
		}
		cases = append(cases, classify.EvalCase{
			ID:    e.ReportID,
			Image: image,
			// Ground truth is the label. Category and hazard are read off the
			// same taxonomy the guardrails use, so they cannot drift from it.
			WantCategory:    classify.CategoryOf(e.Label),
			WantSubcategory: e.Label,
			WantHazard:      classify.IsHazard(e.Label),
			Note:            note(e.Conditions, e.Notes),
		})
	}
	if len(cases) == 0 {
		return fmt.Errorf("every labelled capture is missing its photograph")
	}

	coverage, err := db.CoverageByWard(ctx)
	if err != nil {
		return err
	}
	// Scored against everything routable anywhere: the question here is whether
	// the model reads the photograph, not whether that ward has a department.
	all := classify.Coverage{}
	for _, categories := range coverage {
		for _, c := range categories {
			all[c] = true
		}
	}

	prompt, err := classify.LoadPrompt("road_defect_v1")
	if err != nil {
		return err
	}
	mode := classify.SchemaJSON
	if os.Getenv("CLASSIFY_SCHEMA_MODE") == string(classify.SchemaJSONObject) {
		mode = classify.SchemaJSONObject
	}
	kind := classify.CitizenPhotograph
	if os.Getenv("CLASSIFY_DATA_KIND") == string(classify.OwnTestData) {
		kind = classify.OwnTestData
	}

	client := classify.NewGateway(classify.GatewayOptions{
		BaseURL:    envOr("CLASSIFY_BASE_URL", "https://api.deepseek.com/v1"),
		APIKey:     key,
		Model:      *model,
		SchemaMode: mode,
		DataKind:   kind,
		Timeout:    120 * time.Second,
	})

	report, err := classify.RunEval(ctx, client, cases, all)
	if err != nil {
		return err
	}
	report.Model = *model
	report.Prompt = prompt.Version

	printEval(report, missing)

	if *out != "" {
		body, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(*out, body, 0o644); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
		fmt.Fprintf(os.Stdout, "\nwritten to %s\n", *out)
	}
	return nil
}

// note folds the conditions and the labeller's words into one line. Conditions
// are usually what explains a miss: a model that fails at night and succeeds at
// noon has a lighting problem, not an accuracy problem.
func note(conditions []string, notes string) string {
	var s string
	for i, c := range conditions {
		if i > 0 {
			s += ", "
		}
		s += c
	}
	if notes != "" {
		if s != "" {
			s += " — "
		}
		s += notes
	}
	return s
}

func printEval(r classify.Report, missing int) {
	fmt.Fprintf(os.Stdout, "model      %s\nprompt     %s\ncases      %d\n",
		r.Model, r.Prompt, r.Cases)
	if missing > 0 {
		fmt.Fprintf(os.Stdout, "skipped    %d (photograph missing)\n", missing)
	}
	fmt.Fprintf(os.Stdout, "category   %.1f%%  (bar: 90%%)\n", r.CategoryAccuracy()*100)
	fmt.Fprintf(os.Stdout, "subcategory %.1f%%\n", r.SubcategoryAccuracy()*100)

	// Printed as a count, not a percentage. "99.2% recall" reads as a pass;
	// "2 hazards missed" reads as two photographs of something that could hurt
	// someone, routed as ordinary.
	fmt.Fprintf(os.Stdout, "hazards missed %d\n", r.HazardsMissed)
	if r.Overrides > 0 {
		fmt.Fprintf(os.Stdout, "rescued by the hazard rule %d\n", r.Overrides)
	}

	if len(r.Misses) == 0 {
		return
	}
	fmt.Fprintf(os.Stdout, "\nmisses — read these, the number alone cannot be acted on\n")
	for _, m := range r.Misses {
		fmt.Fprintf(os.Stdout, "  %s  %s: want %q got %q", m.ID[:8], m.Field, m.Want, m.Got)
		if m.Note != "" {
			fmt.Fprintf(os.Stdout, "  [%s]", m.Note)
		}
		fmt.Fprintln(os.Stdout)
		if m.Rationale != "" {
			fmt.Fprintf(os.Stdout, "        %s\n", m.Rationale)
		}
	}
}
