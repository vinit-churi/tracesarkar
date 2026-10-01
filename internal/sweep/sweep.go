// Package sweep runs the stages that turn a stored capture into an answer.
//
// Capture and enrichment are deliberately separate: the photograph is persisted
// before anything is attempted on it, and every stage after that is retryable
// (hard rule 7). This package is what makes "retryable" true in practice — it
// runs each stage on a schedule, independently, and lets a stage fail without
// taking the others down with it.
package sweep

import (
	"context"
	"log/slog"
	"time"
)

// A Stage is one pass over the work waiting to be done. It is expected to
// process a bounded batch and return; the next run picks up the remainder.
type Stage struct {
	Name string
	Run  func(context.Context) error
}

// Result is what one stage did.
type Result struct {
	Name string
	Err  error
	Took time.Duration
}

// Run executes every stage in order and returns one Result each.
//
// A stage that fails does not stop the ones after it. The stages are
// independent — a classification provider being down must not stop jurisdiction
// from resolving — and anything missed is picked up by the next run, because
// each stage selects its own outstanding work rather than being handed it.
//
// A cancelled context stops the remaining stages instead of starting them.
// Cloud Run sends SIGTERM before it stops a job, and beginning a paid vision
// call into a shutdown wastes money and leaves a half-finished pass behind.
func Run(ctx context.Context, stages []Stage) []Result {
	results := make([]Result, 0, len(stages))

	for _, s := range stages {
		if err := ctx.Err(); err != nil {
			results = append(results, Result{Name: s.Name, Err: err})
			continue
		}

		started := time.Now()
		err := s.Run(ctx)
		took := time.Since(started)

		if err != nil {
			slog.ErrorContext(ctx, "sweep stage failed",
				"stage", s.Name, "took_ms", took.Milliseconds(), "error", err)
		} else {
			slog.InfoContext(ctx, "sweep stage done",
				"stage", s.Name, "took_ms", took.Milliseconds())
		}
		results = append(results, Result{Name: s.Name, Err: err, Took: took})
	}

	return results
}

// Failed reports whether any stage failed. The job's exit code is what the
// scheduler alerts on, so this is the one question the caller has to answer.
func Failed(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}
