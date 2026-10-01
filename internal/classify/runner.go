package classify

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Pending is a stored capture waiting to be classified.
type Pending struct {
	ReportID   string
	ArchiveKey string
}

// Attempt is one classification, successful or not.
type Attempt struct {
	ReportID string
	Decision Decision
	Model    string
	Prompt   string
	Latency  time.Duration
	Err      string
}

// Store is the persistence the runner needs.
type Store interface {
	Pending(ctx context.Context, limit int) ([]Pending, error)
	Save(ctx context.Context, a Attempt) error
}

// Blobs reads stored photographs.
type Blobs interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

// RunnerOptions configures a pass over the queue.
type RunnerOptions struct {
	Store      Store
	Blobs      Blobs
	Classifier Classifier
	Coverage   Coverage
	Model      string
	Prompt     string
	Limit      int
	Log        *slog.Logger
}

// Run classifies the captures waiting for it, and reports how many succeeded.
//
// It runs *after* capture, never during: a capture is durable before anything
// is attempted on it (hard rule 7), so this is free to fail and be retried.
func Run(ctx context.Context, o RunnerOptions) (int, error) {
	log := o.Log
	if log == nil {
		log = slog.Default()
	}
	if o.Limit == 0 {
		o.Limit = 25
	}

	pending, err := o.Store.Pending(ctx, o.Limit)
	if err != nil {
		return 0, fmt.Errorf("find pending captures: %w", err)
	}

	var done int
	for _, p := range pending {
		attempt := Attempt{ReportID: p.ReportID, Model: o.Model, Prompt: o.Prompt}
		started := time.Now()

		image, err := o.Blobs.Get(ctx, p.ArchiveKey)
		if err == nil {
			var raw Result
			raw, err = o.Classifier.Classify(ctx, image, "")
			if err == nil {
				attempt.Decision = Apply(raw, o.Coverage)
				done++
			}
		}
		attempt.Latency = time.Since(started)
		if attempt.Latency <= 0 {
			attempt.Latency = time.Microsecond
		}
		if err != nil {
			// One capture that cannot be classified must not block the queue
			// behind it. Nothing captured is ever lost, and a stuck queue
			// loses everything after the stuck item.
			attempt.Err = err.Error()
			log.Warn("classification failed", "report_id", p.ReportID, "error", err.Error())
		}

		// Recorded either way: a classification that silently never ran looks
		// exactly like one that ran and found nothing.
		if saveErr := o.Store.Save(ctx, attempt); saveErr != nil {
			return done, fmt.Errorf("record attempt for %s: %w", p.ReportID, saveErr)
		}
	}

	if len(pending) > 0 {
		log.Info("classification pass", "seen", len(pending), "classified", done,
			"model", o.Model, "prompt", o.Prompt)
	}

	// A pass where nothing at all succeeded is reported, not just logged. One
	// unreadable photograph among good ones is ordinary; a whole batch failing
	// identically is an expired key, a retired model name or a provider that is
	// down, and that must reach whoever is watching rather than sit in a log
	// nobody reads during a day of fieldwork.
	//
	// A single capture failing on its own is not enough to tell the two apart,
	// so it is left to the retry cap and the recorded attempt.
	if len(pending) > 1 && done == 0 {
		return done, fmt.Errorf("classify: all %d captures in this pass failed", len(pending))
	}
	return done, nil
}
