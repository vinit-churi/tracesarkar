package store

import (
	"context"
	"fmt"
	"time"
)

// System 8's first surface: the numbers an operator currently reads out of
// logs.
//
// Everything here is a count. Deliberately: an operator console that shows
// rows invites moderating from a dashboard, and this platform's moderation
// surfaces already exist and are deliberately narrow — `/review/` for an
// attribution verdict, `/label/` for the evaluation set, `/truth/` for the
// redaction set. What is missing is not another queue. It is the answer to
// "is collection healthy, is anything stuck, and what is the ratio" without
// grepping a log.

// Queue is work waiting, with the thing that would clear it.
type Queue struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	// Clears names what moves the number, so a count nobody can act on is
	// not mistaken for one somebody should.
	Clears string `json:"clears"`
}

// IssueCount is the issues in one status.
type IssueCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// Health is the whole operator view.
type Health struct {
	At      time.Time        `json:"at"`
	Sources []SnapshotStatus `json:"-"`
	Runs    []RunRow         `json:"-"`
	Queues  []Queue          `json:"queues"`
	Issues  []IssueCount     `json:"issues"`
	// Captures is every stored report, so a queue of zero can be told from a
	// platform with nothing in it.
	Captures int `json:"captures"`
}

// ClaimedVersusConfirmed is the ratio this platform exists to publish
// ([D007](../../docs/00-overview/05-decision-log.md)): how many issues an
// authority says it attended, against how many a resident photographed fixed.
//
// Returned as two counts and never as a single percentage. A percentage of
// small numbers is a lie with a decimal point in it, and at this stage the
// numbers are small.
func (h Health) ClaimedVersusConfirmed() (claimed, confirmed int) {
	for _, c := range h.Issues {
		switch c.Status {
		case "claimed_resolved":
			claimed = c.Count
		case "citizen_confirmed":
			confirmed = c.Count
		}
	}
	return claimed, confirmed
}

// OperatorHealth gathers it in one round trip per question.
func (d *DB) OperatorHealth(ctx context.Context, runLimit int) (Health, error) {
	h := Health{At: time.Now().UTC()}

	sources, err := d.Status(ctx)
	if err != nil {
		return h, fmt.Errorf("operator health: sources: %w", err)
	}
	h.Sources = sources

	if runLimit > 0 {
		runs, err := d.RecentRuns(ctx, runLimit)
		if err != nil {
			return h, fmt.Errorf("operator health: runs: %w", err)
		}
		h.Runs = runs
	}

	// Every status the table allows, including the ones at zero. A status
	// missing from a dashboard reads as "none" and is indistinguishable from
	// "we stopped counting it".
	h.Issues = []IssueCount{
		{Status: "open"}, {Status: "claimed_resolved"},
		{Status: "citizen_confirmed"}, {Status: "reopened"},
	}
	rows, err := d.pool.Query(ctx, `SELECT status, count(*) FROM issues GROUP BY status`)
	if err != nil {
		return h, fmt.Errorf("operator health: issues: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return h, fmt.Errorf("operator health: scan issue count: %w", err)
		}
		for i := range h.Issues {
			if h.Issues[i].Status == status {
				h.Issues[i].Count = n
			}
		}
	}
	if err := rows.Err(); err != nil {
		return h, fmt.Errorf("operator health: issues: %w", err)
	}

	// Each queue is the same shape as the stage that clears it, so a number
	// here and the sweep's own behaviour cannot drift apart.
	queues := []struct {
		name, clears, query string
	}{
		// The same condition PendingEnrichment uses, so the number and the
		// stage that clears it cannot drift apart: a capture is enriched when
		// it has both a jurisdiction and an attribution row with no error.
		{"captures not enriched", "the sweep's enrich stage", `
			SELECT count(*) FROM reports r
			 WHERE NOT EXISTS (SELECT 1 FROM report_jurisdiction j
			                    WHERE j.report_id = r.id AND j.error IS NULL)
			    OR NOT EXISTS (SELECT 1 FROM report_attribution a
			                    WHERE a.report_id = r.id AND a.error IS NULL)`},
		{"media not scanned for faces or plates", "the sweep's redact stage", `
			SELECT count(*) FROM report_media WHERE redaction_verdict IS NULL`},
		{"media held by the redaction gate", "a person, or a better detector", `
			SELECT count(*) FROM report_media WHERE redaction_verdict = 'held'`},
		{"captures not classified", "the sweep's classify stage", `
			SELECT count(*) FROM reports r
			 WHERE NOT EXISTS (SELECT 1 FROM classifications c WHERE c.report_id = r.id)`},
		// Three separate numbers, because the first draft of this dashboard
		// reported one — "captures not attached to an issue: 3" — while the
		// sweep's own issues stage printed "nothing to attach". Both were
		// right. The stage attaches a capture only when its latest
		// classification was `accepted` or `not_yet_covered`; those three
		// were `needs_confirmation`, so the platform had asked the citizen
		// which problem it was and nobody had answered. A dashboard that
		// shows work for a job that will never pick it up trains an operator
		// to ignore the number, which is the opposite of the point.
		//
		// So: what the sweep will take, what is waiting on a person who is
		// not the operator, and what the platform has already decided about.
		{"captures not attached to an issue", "the sweep's issues stage", `
			SELECT count(*) FROM reports r
			  JOIN LATERAL (
			    SELECT outcome, category FROM classifications
			     WHERE report_id = r.id AND error IS NULL
			     ORDER BY created_at DESC LIMIT 1
			  ) c ON true
			 WHERE NOT EXISTS (SELECT 1 FROM issue_reports ir WHERE ir.report_id = r.id)
			   AND COALESCE(c.category,'') <> ''
			   AND c.outcome IN ('accepted', 'not_yet_covered')`},
		{"captures waiting for the citizen to say which problem this is",
			"the person who sent it, and no job can do it for them", `
			SELECT count(*) FROM reports r
			  JOIN LATERAL (
			    SELECT outcome FROM classifications
			     WHERE report_id = r.id AND error IS NULL
			     ORDER BY created_at DESC LIMIT 1
			  ) c ON true
			 WHERE NOT EXISTS (SELECT 1 FROM issue_reports ir WHERE ir.report_id = r.id)
			   AND c.outcome = 'needs_confirmation'`},
		{"captures the platform declined, as unreadable or out of scope",
			"nothing — this is a decision, not a backlog", `
			SELECT count(*) FROM reports r
			  JOIN LATERAL (
			    SELECT outcome FROM classifications
			     WHERE report_id = r.id AND error IS NULL
			     ORDER BY created_at DESC LIMIT 1
			  ) c ON true
			 WHERE NOT EXISTS (SELECT 1 FROM issue_reports ir WHERE ir.report_id = r.id)
			   AND c.outcome IN ('needs_retake', 'out_of_scope')`},
		{"attribution answers awaiting a verdict", "a person, at /review/", `
			SELECT count(*) FROM attribution_reviews WHERE verdict IS NULL`},
		{"one-time links still live", "their own expiry, then the sweep", `
			SELECT count(*) FROM auth_tokens
			 WHERE used_at IS NULL AND expires_at > now()`},
	}
	for _, q := range queues {
		var n int
		if err := d.pool.QueryRow(ctx, q.query).Scan(&n); err != nil {
			// A queue that cannot be counted is reported as such rather than
			// as zero: a dashboard that silently shows nothing is worse than
			// no dashboard.
			return h, fmt.Errorf("operator health: %s: %w", q.name, err)
		}
		h.Queues = append(h.Queues, Queue{Name: q.name, Count: n, Clears: q.clears})
	}

	if err := d.pool.QueryRow(ctx, `SELECT count(*) FROM reports`).Scan(&h.Captures); err != nil {
		return h, fmt.Errorf("operator health: captures: %w", err)
	}
	return h, nil
}
