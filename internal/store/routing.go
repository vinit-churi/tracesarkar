package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Routing is the desk a problem belongs to, and what obliges it.
type Routing struct {
	Department       string
	Officer          string
	Office           string
	SourceRef        string
	DeadlineHours    int
	DeadlineCitation string
}

// deadlineKeys maps a category to the legal constant that governs it.
//
// Keyed here rather than in the database because the mapping is a product
// decision, not a fact about the world: which of several published clocks the
// platform shows for a category is ours to choose, while the clocks themselves
// are sourced.
var deadlineKeys = map[string]string{
	"road_defect": "pothole.attend_hours.court",
	"waste":       "waste.refuse_removal_hours",
}

// RoutingFor returns the department responsible for a category in a ward, with
// the deadline that binds it.
//
// Several departments can share a category — a road defect in R/C has two,
// split by road class and by season — so the first by effective date is
// returned and the rest are the caller's problem only when it has to choose.
// Not found is not an error: most wards have no department recorded, and that
// is a state the product states plainly rather than hides.
func (d *DB) RoutingFor(ctx context.Context, authority, ward, category string) (Routing, bool, error) {
	var r Routing
	err := d.pool.QueryRow(ctx, `
		SELECT department, COALESCE(officer_designation, ''),
		       COALESCE(pio_address, ''), source_ref
		  FROM authority_departments
		 WHERE authority = $1 AND ward = $2 AND category = $3
		 ORDER BY effective_from DESC, created_at
		 LIMIT 1`, authority, ward, category).
		Scan(&r.Department, &r.Officer, &r.Office, &r.SourceRef)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return Routing{}, false, nil
		}
		return Routing{}, false, fmt.Errorf("routing for %s/%s: %w", ward, category, err)
	}

	key, ok := deadlineKeys[category]
	if !ok {
		// No deadline mapped is not a failure. Water supply publishes none.
		return r, true, nil
	}
	var value, citation string
	err = d.pool.QueryRow(ctx, `
		SELECT value::text, citation FROM legal_constants
		 WHERE key = $1 AND effective_from <= now()::date
		   AND (effective_to IS NULL OR effective_to >= now()::date)
		 ORDER BY effective_from DESC LIMIT 1`, key).Scan(&value, &citation)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return r, true, nil
		}
		return Routing{}, false, fmt.Errorf("deadline %s: %w", key, err)
	}
	if hours, convErr := strconv.Atoi(strings.Trim(value, `"`)); convErr == nil {
		r.DeadlineHours = hours
		r.DeadlineCitation = citation
	}
	return r, true, nil
}

// Wards lists the wards the platform holds a boundary for, sorted.
//
// Offered rather than typed, because a ward typed by hand is a ward that can
// be misspelled, and a misspelled ground truth scores a correct answer as
// wrong. It is also the honest edge of the platform: a place not on this list
// is one where nothing can be resolved, and the person standing there should
// be told that rather than left guessing which code to enter.
func (d *DB) Wards(ctx context.Context) ([]string, error) {
	rows, err := d.pool.Query(ctx,
		`SELECT DISTINCT ward FROM ward_boundaries WHERE ward <> '' ORDER BY ward`)
	if err != nil {
		return nil, fmt.Errorf("wards: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var w string
		if err := rows.Scan(&w); err != nil {
			return nil, fmt.Errorf("scan ward: %w", err)
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
