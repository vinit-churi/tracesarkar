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
			Name: "classify",
			Run: func(ctx context.Context) error {
				return runClassify(ctx, []string{"-limit", strconv.Itoa(*classifyLimit)})
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
