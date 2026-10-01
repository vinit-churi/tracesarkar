package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/attribute"
	"github.com/vinit-churi/tracesarkar/internal/jurisdiction"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// runEnrich answers the two spatial questions about each capture: which
// authority owns the spot, and which contract covers it.
//
// Both run after capture, never during, and each is recorded separately. A
// report with a ward and no contract is a complete answer — about half of
// Borivali has no contract data, and saying so is the honest result rather
// than a failure.
func runEnrich(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("enrich", flag.ExitOnError)
	limit := fs.Int("limit", 50, "how many captures to enrich in this pass")
	if err := fs.Parse(args); err != nil {
		return err
	}

	_, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	points, err := db.UnenrichedReports(ctx, *limit)
	if err != nil {
		return err
	}
	if len(points) == 0 {
		fmt.Fprintln(os.Stdout, "nothing to enrich")
		return nil
	}

	var routed, attributed, uncovered int
	for _, p := range points {
		// Jurisdiction.
		hits, err := db.WardsNear(ctx, p.Lat, p.Lon, 400)
		if err != nil {
			_ = db.SaveJurisdiction(ctx, p.ReportID, jurisdiction.Resolution{
				Confidence: jurisdiction.None}, "", err.Error())
		} else {
			candidates := make([]jurisdiction.WardCandidate, 0, len(hits))
			vintage := ""
			for _, h := range hits {
				candidates = append(candidates, jurisdiction.WardCandidate{
					Ward: h.Ward, Authority: h.Authority,
					Inside: h.Inside, DistanceToEdgeM: h.DistanceToEdgeM,
				})
				if h.Inside && h.Vintage != "" {
					vintage = h.Vintage
				}
			}
			res := jurisdiction.Resolve(p.AccuracyM, candidates)
			if err := db.SaveJurisdiction(ctx, p.ReportID, res, vintage, ""); err != nil {
				return err
			}
			if res.Ward != "" {
				routed++
			}
		}

		// Attribution.
		matches, err := db.NearestRoadSegments(ctx, p.Lat, p.Lon,
			attribute.SearchRadius(p.AccuracyM), 8)
		if err != nil {
			_ = db.SaveAttribution(ctx, p.ReportID, attribute.Attribution{
				Confidence: attribute.None}, "", nil, err.Error())
			continue
		}
		a := attribute.Decide(p.AccuracyM, candidatesFrom(matches))

		// Provenance travels with the contractor name (hard rule 2).
		var source string
		var retrieved *time.Time
		if a.Match != nil {
			for _, m := range matches {
				if m.WorkID == a.Match.WorkID {
					source, retrieved = m.SourceID, m.RetrievedAt
					break
				}
			}
			attributed++
		} else {
			uncovered++
		}
		if err := db.SaveAttribution(ctx, p.ReportID, a, source, retrieved, ""); err != nil {
			return err
		}
	}

	slog.Info("enrichment pass", "captures", len(points),
		"routed", routed, "attributed", attributed, "no_contract", uncovered)

	fmt.Fprintf(os.Stdout, "\n%d capture(s): %d routed to a ward, %d matched to a contract, "+
		"%d with no contract data for that stretch\n", len(points), routed, attributed, uncovered)
	return nil
}

var _ = store.ReportPoint{}
