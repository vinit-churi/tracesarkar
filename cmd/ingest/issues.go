package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/vinit-churi/tracesarkar/internal/issue"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// runIssues turns classified captures into issues, or into further sightings
// of issues already known.
//
// Runs after classification because the category decides which issues a
// sighting could belong to: merging on position alone would put a pothole and
// a sack of rubbish in the same ticket, bound for a department that can act on
// one of them.
func runIssues(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("issues", flag.ExitOnError)
	limit := fs.Int("limit", 100, "how many captures to attach in this pass")
	if err := fs.Parse(args); err != nil {
		return err
	}

	_, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	pending, err := db.UnattachedReports(ctx, *limit)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		fmt.Fprintln(os.Stdout, "nothing to attach")
		return nil
	}

	var created, joined, reopened int
	for _, u := range pending {
		// Searched well beyond the merge radius so the decision sees what it
		// is rejecting, not only what it accepts.
		near, err := db.IssuesNear(ctx, u.Lat, u.Lon, 150)
		if err != nil {
			return err
		}
		candidates := make([]issue.Candidate, 0, len(near))
		for _, n := range near {
			candidates = append(candidates, issue.Candidate{
				IssueID: n.IssueID, Category: n.Category,
				Status: n.Status, DistanceM: n.DistanceM,
			})
		}

		decision := issue.Decide(u.AccuracyM, u.Category, candidates)
		issueID, err := db.AttachReport(ctx, store.Attachment{
			ReportID: u.ReportID, Authority: u.Authority, Ward: u.Ward,
			Category: u.Category, Subcategory: u.Subcategory,
			Lat: u.Lat, Lon: u.Lon, SeenAt: u.SeenAt, Actor: u.AccountID,
			Decision: decision,
		})
		if err != nil {
			return err
		}

		switch decision.Action {
		case issue.Create:
			created++
		case issue.Join:
			joined++
		case issue.Reopen:
			reopened++
		}
		slog.Info("report attached", "report_id", u.ReportID,
			"issue_id", issueID, "action", decision.Action,
			"distance_m", decision.DistanceM)
	}

	fmt.Fprintf(os.Stdout, "%d new issue(s), %d corroboration(s), %d reopened\n",
		created, joined, reopened)
	return nil
}
