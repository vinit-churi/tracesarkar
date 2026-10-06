package store

import (
	"context"

	"github.com/vinit-churi/tracesarkar/internal/classify"
)

// Pending adapts the store to the classification runner's interface.
func (d *DB) Pending(ctx context.Context, limit int) ([]classify.Pending, error) {
	captures, err := d.UnclassifiedReports(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]classify.Pending, 0, len(captures))
	for _, c := range captures {
		out = append(out, classify.Pending{
			ReportID: c.ReportID, ArchiveKey: c.ArchiveKey, Ward: c.Ward,
		})
	}
	return out, nil
}

// Save records one classification attempt.
func (d *DB) Save(ctx context.Context, a classify.Attempt) error {
	return d.SaveClassification(ctx, NewClassification{
		ReportID: a.ReportID,
		Decision: a.Decision,
		Model:    a.Model,
		Prompt:   a.Prompt,
		Latency:  a.Latency,
		Err:      a.Err,
	})
}
