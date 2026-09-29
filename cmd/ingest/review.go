package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/vinit-churi/tracesarkar/internal/attribute"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// candidatesFrom adapts the store's matches to the decision's input.
func candidatesFrom(ms []store.RoadMatch) []attribute.Candidate {
	out := make([]attribute.Candidate, 0, len(ms))
	for _, m := range ms {
		out = append(out, attribute.Candidate{
			WorkID:         m.WorkID,
			WorkCode:       m.WorkCode,
			ContractorName: m.ContractorName,
			LocationName:   m.LocationName,
			DistanceM:      m.DistanceM,
			EndDate:        m.EndDate,
		})
	}
	return out
}

// runReview fills the review queue with attributions for a person to judge.
func runReview(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("review", flag.ExitOnError)
	ward := fs.String("ward", "R/C", "ward to generate probes in")
	count := fs.Int("count", 50, "how many probes to generate")
	accuracy := fs.Float64("accuracy", 8, "GPS accuracy to simulate, in metres")
	minOff := fs.Float64("min-offset", 3, "closest a probe sits to the centreline")
	maxOff := fs.Float64("max-offset", 12, "furthest a probe sits from the centreline")
	if err := fs.Parse(args); err != nil {
		return err
	}

	_, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	probes, err := db.GenerateProbes(ctx, *ward, *count, *minOff, *maxOff)
	if err != nil {
		return err
	}
	if len(probes) == 0 {
		return fmt.Errorf("no road geometry loaded for ward %s; run: ingest roads --ward %s", *ward, *ward)
	}

	bands := map[attribute.Confidence]int{}
	for _, p := range probes {
		matches, err := db.NearestRoadSegments(ctx, p.Lat, p.Lon,
			attribute.SearchRadius(*accuracy), 8)
		if err != nil {
			return err
		}
		a := attribute.Decide(*accuracy, candidatesFrom(matches))
		bands[a.Confidence]++

		item := store.NewReviewItem{
			Kind:           "probe",
			Lat:            p.Lat,
			Lon:            p.Lon,
			AccuracyM:      *accuracy,
			Confidence:     string(a.Confidence),
			Basis:          a.Basis,
			ExpectedWorkID: &p.ExpectedWorkID,
		}
		if a.Match != nil {
			workID := a.Match.WorkID
			distance := a.Match.DistanceM
			item.MatchedWorkID = &workID
			item.MatchedWorkCode = a.Match.WorkCode
			item.MatchedContractor = a.Match.ContractorName
			item.MatchedLocation = a.Match.LocationName
			item.DistanceM = &distance
		}
		if err := db.SaveReviewItem(ctx, item); err != nil {
			return err
		}
	}

	slog.Info("review queue filled", "probes", len(probes), "ward", *ward,
		"accuracy_m", *accuracy,
		"high", bands[attribute.High], "check_this", bands[attribute.CheckThis],
		"none", bands[attribute.None])

	summary, err := db.ReviewSummary(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "\nreview queue: %d total, %d waiting for a verdict\n",
		summary.Total, summary.Pending)
	if summary.HasPrecision() {
		fmt.Fprintf(os.Stdout, "precision so far: %.1f%% over %d judged\n",
			100*summary.Precision(), summary.Correct+summary.Wrong)
	}
	return nil
}
