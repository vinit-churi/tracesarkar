package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/vinit-churi/tracesarkar/internal/sweep"
)

// runSweep turns stored captures into answers, on a schedule.
//
// Capture persists the photograph and returns; everything after is done here.
// Without it a report sits unenriched until someone runs a command by hand,
// which is fine for six captures and useless on a day of fieldwork.
func runSweep(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("sweep", flag.ExitOnError)
	enrichLimit := fs.Int("enrich-limit", 200, "captures to enrich in this pass")
	classifyLimit := fs.Int("classify-limit", 100, "captures to classify in this pass")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Each stage reuses the existing subcommand rather than a second copy of
	// its logic, so a fix to either is a fix to both.
	results := sweep.Run(ctx, []sweep.Stage{
		{
			Name: "enrich",
			Run: func(ctx context.Context) error {
				return runEnrich(ctx, []string{"-limit", strconv.Itoa(*enrichLimit)})
			},
		},
		{
			// Before classification, not after. Classification is a third
			// party, and ADR 0008 says an un-redacted photograph of an
			// identifiable person must not reach one — so the mosaic is
			// applied first and the classifier reads that copy. A held
			// verdict withholds the picture from the public, not from the
			// reporter: the redacted copy is written either way, so an
			// answer still comes back.
			Name: "redact",
			Run: func(ctx context.Context) error {
				return runRedact(ctx, []string{"--limit", strconv.Itoa(*classifyLimit)})
			},
		},
		{
			Name: "classify",
			Run: func(ctx context.Context) error {
				return runClassify(ctx, []string{"-limit", strconv.Itoa(*classifyLimit)})
			},
		},
		{
			// Last of the capture stages, because the category decides which
			// issues a sighting could belong to.
			Name: "issues",
			Run: func(ctx context.Context) error {
				return runIssues(ctx, []string{"-limit", strconv.Itoa(*enrichLimit)})
			},
		},
		{
			// Housekeeping, and it belongs here rather than on a timer of its
			// own: a spent password-reset row has no reason to outlive the
			// week in which somebody might ask about it, and a job that
			// already runs is cheaper than a job that has to be scheduled,
			// watched and alerted on. The query matches nothing most passes.
			Name: "credentials",
			Run: func(ctx context.Context) error {
				return runCredentials(ctx, nil)
			},
		},
	})

	for _, r := range results {
		status := "ok"
		if r.Err != nil {
			status = r.Err.Error()
		}
		fmt.Fprintf(os.Stdout, "%s\t%dms\t%s\n", r.Name, r.Took.Milliseconds(), status)
	}

	if sweep.Failed(results) {
		// Non-zero so the scheduler's failure alert fires. The stages that did
		// work have already committed it; the next run retries the rest.
		return fmt.Errorf("sweep: one or more stages failed")
	}
	return nil
}
