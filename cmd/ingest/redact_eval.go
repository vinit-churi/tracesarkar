package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sort"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/redact"
)

// runRedactEval scores the face and plate detectors against the photographs a
// person has marked up.
//
// This is the measurement ADR 0008 required before launch and the detector
// shipped without. Run over the captures that existed then, redaction reached a
// verdict on every one and missed nothing — because not one contained a face or
// a plate. The number that mattered was not good or bad, it did not exist.
//
// Nothing here is scored against the redacted derivative. The detectors run on
// the archival original, which is where a face still is, and the archival
// original never leaves the machine.
func runRedactEval(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("redact-eval", flag.ExitOnError)
	limit := fs.Int("limit", 1000, "how many annotated photographs to score")
	out := fs.String("out", "", "write the report as JSON to this path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	annotated, err := db.RedactionTruthSet(ctx, *limit)
	if err != nil {
		return err
	}
	if len(annotated) == 0 {
		// Refused rather than reported as a clean run over nothing. A harness
		// that prints a report with no cases in it is how a detector comes to
		// be described as measured when it is not.
		return fmt.Errorf("no photograph has been marked up yet; the set is built at /truth/")
	}

	blobs, err := archive.New(archive.Options{
		Endpoint: cfg.R2.Endpoint, Bucket: cfg.R2.Bucket, Region: cfg.R2.Region,
		AccessKey: cfg.R2.AccessKey, Secret: cfg.R2.Secret,
	})
	if err != nil {
		return err
	}

	faces, err := redact.NewFaces()
	if err != nil {
		return err
	}
	scanner := redact.NewScanner(faces, redact.NewPlates())

	cases := make([]redact.Case, 0, len(annotated))
	var missing int
	for _, a := range annotated {
		body, err := blobs.Get(ctx, a.ArchiveKey)
		if err != nil {
			// Named and counted rather than dropped. A shrinking eval set
			// flatters every figure computed from it.
			missing++
			fmt.Fprintf(os.Stderr, "skipping %s: %v\n", a.ReportID, err)
			continue
		}
		img, _, err := image.Decode(bytes.NewReader(body))
		if err != nil {
			missing++
			fmt.Fprintf(os.Stderr, "skipping %s: %v\n", a.ReportID, err)
			continue
		}
		cases = append(cases, redact.Case{
			ID:            a.ReportID,
			Image:         img,
			Truth:         a.Boxes,
			Conditions:    a.Conditions,
			PeoplePresent: a.PeoplePresent,
		})
	}
	if len(cases) == 0 {
		return fmt.Errorf("every annotated photograph is missing its file")
	}

	started := time.Now()
	report := redact.Score(ctx, scanner, cases)
	printRedactEval(report, missing, time.Since(started))

	if *out != "" {
		body, err := json.MarshalIndent(redactEvalJSON(report), "", "  ")
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

func printRedactEval(r redact.Report, missing int, took time.Duration) {
	fmt.Fprintf(os.Stdout, "tools      %s\ncases      %d\n", r.Tools, r.Cases)
	if missing > 0 {
		fmt.Fprintf(os.Stdout, "skipped    %d (file missing or unreadable)\n", missing)
	}
	fmt.Fprintf(os.Stdout, "held       %d\ntook       %s\n\n", r.Held, took.Round(time.Millisecond))

	for _, kind := range []redact.Kind{redact.Face, redact.Plate} {
		t := r.Kinds[kind]
		fmt.Fprintf(os.Stdout, "%-6s  present %-3d covered %-3d  recall %s   found %-3d spurious %-3d  precision %s\n",
			kind, t.Truth, t.Covered, share(t.Recall()), t.Found, t.Spurious, share(t.Precision()))
	}

	// The headline, printed last so it is the line left on the screen, and as a
	// count rather than a percentage: "98% recall" reads as a pass, "one face
	// published" reads as what it is.
	fmt.Fprintf(os.Stdout, "\nfaces and plates that would have been published: %d\n", r.Exposed())

	if len(r.ByCondition) > 0 {
		fmt.Fprintf(os.Stdout, "\nby condition — this is what says whether a miss is a lighting problem\n")
		names := make([]string, 0, len(r.ByCondition))
		for name := range r.ByCondition {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			t := r.ByCondition[name]
			fmt.Fprintf(os.Stdout, "  %-24s present %-3d covered %-3d recall %s\n",
				name, t.Truth, t.Covered, share(t.Recall()))
		}
	}

	if len(r.Misses) == 0 {
		return
	}
	fmt.Fprintf(os.Stdout, "\nmisses — read these, the number alone cannot be acted on\n")
	for _, m := range r.Misses {
		status := "held"
		if m.Published {
			status = "PUBLISHED"
		}
		fmt.Fprintf(os.Stdout, "  %s  %-5s %.0f%% covered  %s", short(m.ID), m.Kind, m.Covered*100, status)
		if len(m.Conditions) > 0 {
			fmt.Fprintf(os.Stdout, "  [%s]", join(m.Conditions))
		}
		fmt.Fprintln(os.Stdout)
	}
}

// share prints a figure that does not exist as a dash rather than as a number.
// Recall over a set with nothing to find is not 100%, and printing it as 100%
// is the mistake this whole harness was built to stop making.
func share(v float64, ok bool) string {
	if !ok {
		return "  —  "
	}
	return fmt.Sprintf("%5.1f%%", v*100)
}

func short(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func join(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}

// redactEvalJSON is the report as it is written to disk: the same figures, with
// the unmeasured ones absent rather than zero, so a file cannot be read as a
// recall of nought.
func redactEvalJSON(r redact.Report) map[string]any {
	kinds := map[string]any{}
	for kind, t := range r.Kinds {
		entry := map[string]any{
			"present": t.Truth, "covered": t.Covered,
			"found": t.Found, "spurious": t.Spurious,
		}
		if v, ok := t.Recall(); ok {
			entry["recall"] = v
		}
		if v, ok := t.Precision(); ok {
			entry["precision"] = v
		}
		kinds[string(kind)] = entry
	}
	conditions := map[string]any{}
	for name, t := range r.ByCondition {
		entry := map[string]any{"present": t.Truth, "covered": t.Covered}
		if v, ok := t.Recall(); ok {
			entry["recall"] = v
		}
		conditions[name] = entry
	}
	return map[string]any{
		"tools":        r.Tools,
		"cases":        r.Cases,
		"held":         r.Held,
		"exposed":      r.Exposed(),
		"kinds":        kinds,
		"by_condition": conditions,
		"misses":       r.Misses,
		"scored_at":    time.Now().UTC().Format(time.RFC3339),
	}
}
