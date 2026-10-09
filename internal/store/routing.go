package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Routing is the desk a problem belongs to, and what obliges it.
type Routing struct {
	Department       string
	Officer          string
	Office           string
	OfficePhone      string
	OfficeEmail      string
	OfficeHours      string
	VisitingHours    string
	EscalatesTo      string
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
		       COALESCE(pio_address, ''), COALESCE(office_phone, ''),
		       COALESCE(office_email, ''), COALESCE(office_hours, ''),
		       COALESCE(visiting_hours, ''), COALESCE(escalates_to, ''),
		       source_ref
		  FROM authority_departments
		 WHERE authority = $1 AND ward = $2 AND category = $3
		 ORDER BY effective_from DESC, created_at
		 LIMIT 1`, authority, ward, category).
		Scan(&r.Department, &r.Officer, &r.Office, &r.OfficePhone,
			&r.OfficeEmail, &r.OfficeHours, &r.VisitingHours, &r.EscalatesTo,
			&r.SourceRef)
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

// Desk is one department's published responsibility in a ward.
type Desk struct {
	Category         string
	Department       string
	Officer          string
	Office           string
	OfficePhone      string
	OfficeEmail      string
	OfficeHours      string
	VisitingHours    string
	EscalatesTo      string
	SourceRef        string
	RetrievedAt      time.Time
	DeadlineHours    int
	DeadlineCitation string
}

// WardProfile is everything the platform can say about a ward without any
// citizen having reported anything.
//
// It holds published facts about offices and obligations and nothing else: no
// captures, no coordinates, no person. That is what makes it safe to serve
// without a sign-in, and it is most of what the platform knows.
type WardProfile struct {
	Authority string
	Ward      string
	Desks     []Desk
	// Segments and Metres describe how much of the ward's road network has
	// contract geometry behind it. The gap matters as much as the number:
	// where there is none, the platform cannot tell a road BMC has not
	// published from a road BMC does not own.
	Segments int
	Metres   float64
}

// WardProfileFor assembles it.
func (d *DB) WardProfileFor(ctx context.Context, authority, ward string) (WardProfile, error) {
	p := WardProfile{Authority: authority, Ward: ward}

	rows, err := d.pool.Query(ctx, `
		SELECT category, department, COALESCE(officer_designation, ''),
		       COALESCE(pio_address, ''), COALESCE(office_phone, ''),
		       COALESCE(office_email, ''), COALESCE(office_hours, ''),
		       COALESCE(visiting_hours, ''), COALESCE(escalates_to, ''),
		       source_ref, retrieved_at
		  FROM authority_departments
		 WHERE authority = $1 AND ward = $2
		 ORDER BY category, department`, authority, ward)
	if err != nil {
		return p, fmt.Errorf("ward profile: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var k Desk
		if err := rows.Scan(&k.Category, &k.Department, &k.Officer, &k.Office,
			&k.OfficePhone, &k.OfficeEmail, &k.OfficeHours, &k.VisitingHours,
			&k.EscalatesTo, &k.SourceRef, &k.RetrievedAt); err != nil {
			return p, fmt.Errorf("scan desk: %w", err)
		}
		p.Desks = append(p.Desks, k)
	}
	if err := rows.Err(); err != nil {
		return p, err
	}

	// The deadline for each desk, from the same table the product reads.
	for i := range p.Desks {
		key, ok := deadlineKeys[p.Desks[i].Category]
		if !ok {
			continue
		}
		var value, citation string
		err := d.pool.QueryRow(ctx, `
			SELECT value::text, citation FROM legal_constants
			 WHERE key = $1 AND effective_from <= now()::date
			   AND (effective_to IS NULL OR effective_to >= now()::date)
			 ORDER BY effective_from DESC LIMIT 1`, key).Scan(&value, &citation)
		if err != nil {
			continue
		}
		if hours, convErr := strconv.Atoi(strings.Trim(value, `"`)); convErr == nil {
			p.Desks[i].DeadlineHours = hours
			p.Desks[i].DeadlineCitation = citation
		}
	}

	// How much road has a contract behind it, inside this ward's boundary.
	err = d.pool.QueryRow(ctx, `
		SELECT count(*), COALESCE(sum(ST_Length(s.geom)), 0)
		  FROM road_segments s
		  JOIN ward_boundaries w ON w.authority = $1 AND w.ward = $2
		 WHERE ST_Intersects(s.geom, w.geom)`, authority, ward).
		Scan(&p.Segments, &p.Metres)
	if err != nil {
		return p, fmt.Errorf("ward coverage: %w", err)
	}
	return p, nil
}
