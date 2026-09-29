package roads

import (
	"context"
	"fmt"
	"log/slog"
)

// Sink stores segments. The loader depends on this rather than on the database
// so the counting logic can be tested without one.
type Sink interface {
	SaveSegment(ctx context.Context, s Segment) error
}

// WorkRecord is one row of the works archive: the stored JSON, plus the
// archive row it came from, which the JSON itself does not carry.
type WorkRecord struct {
	WorkID string
	Record map[string]any
}

// Result is what a load run did. Every input is accounted for in exactly one
// of these counts: a work that silently vanishes from a load is how coverage
// rots without anyone noticing.
type Result struct {
	Seen       int
	Loaded     int
	NoGeometry int // BMC published no shape for this work
	OtherWard  int // outside the ward being loaded
	Rejected   int // a shape that looks wrong — not the same as a missing one
}

// Load projects works records into road geometry.
//
// ward, when non-empty, restricts the load to that ward.
func Load(ctx context.Context, records []WorkRecord, sink Sink, ward string) (Result, error) {
	var r Result

	for _, in := range records {
		r.Seen++
		record := in.Record

		if ward != "" && wardOf(record) != ward {
			r.OtherWard++
			continue
		}

		seg, err := SegmentFrom(record)
		switch {
		case ErrNoGeometry(err):
			// Expected and common: BMC has not published a shape for this
			// work. Counted, not logged per row.
			r.NoGeometry++
			continue
		case err != nil:
			// Something is wrong with the source rather than simply absent —
			// a coordinate outside the region, most likely swapped. Loud.
			r.Rejected++
			slog.Warn("road geometry refused", "error", err.Error())
			continue
		}

		seg.WorkID = in.WorkID
		if err := sink.SaveSegment(ctx, seg); err != nil {
			// A storage failure is not a data-quality skip. Carrying on would
			// report a clean run over a half-loaded table.
			return r, fmt.Errorf("load %s: %w", seg.WorkCode, err)
		}
		r.Loaded++
	}

	return r, nil
}

func wardOf(record map[string]any) string { return ward(record) }
