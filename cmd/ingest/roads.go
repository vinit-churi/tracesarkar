package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/vinit-churi/tracesarkar/internal/roads"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// segmentSink adapts the store to the loader's interface.
type segmentSink struct{ db *store.DB }

func (s segmentSink) SaveSegment(ctx context.Context, seg roads.Segment) error {
	return s.db.SaveRoadSegment(ctx, store.NewRoadSegment{
		WorkID:       seg.WorkID,
		WorkCode:     seg.WorkCode,
		Ward:         seg.Ward,
		LocationName: seg.LocationName,
		GeoJSON:      seg.GeoJSON(),
		StartDate:    seg.StartDate,
		EndDate:      seg.EndDate,
	})
}

// runRoads projects the collected works archive into queryable road geometry.
//
// It reads from works, which keeps BMC's JSON exactly as served, and writes
// road_segments, which is derived and can be rebuilt at any time.
func runRoads(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("roads", flag.ExitOnError)
	ward := fs.String("ward", "", "restrict to one ward, e.g. R/C")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()
	_ = cfg

	stored, err := db.WorkRecords(ctx)
	if err != nil {
		return err
	}
	records := make([]roads.WorkRecord, 0, len(stored))
	for _, w := range stored {
		records = append(records, roads.WorkRecord{WorkID: w.ID, Record: w.Current})
	}
	slog.Info("projecting road geometry", "works", len(records), "ward", *ward)

	result, err := roads.Load(ctx, records, segmentSink{db}, *ward)
	if err != nil {
		return err
	}

	slog.Info("road geometry loaded",
		"seen", result.Seen, "loaded", result.Loaded,
		"no_geometry", result.NoGeometry, "other_ward", result.OtherWard,
		"rejected", result.Rejected)

	counts, err := db.CountRoadSegments(ctx)
	if err != nil {
		return err
	}
	wards := make([]string, 0, len(counts))
	for w := range counts {
		wards = append(wards, w)
	}
	sort.Strings(wards)
	fmt.Fprintf(os.Stdout, "\nroad segments now loaded:\n")
	for _, w := range wards {
		fmt.Fprintf(os.Stdout, "  %-8s %d\n", w, counts[w])
	}
	return nil
}

