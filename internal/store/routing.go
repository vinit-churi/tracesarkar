package store

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/classify"
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
// deadlineKeys maps a whole category to the constant that binds it. A entry
// here says the deadline applies to everything in the category.
var deadlineKeys = map[string]string{
	"waste": "waste.refuse_removal_hours",
}

// subcategoryDeadlineKeys maps one specific thing to a constant that binds
// only that thing.
//
// The Bombay High Court's 48-hour duty is the reason this map exists.
// Paragraph 70(ix) reads "All potholes, once brought to the notice of the
// concerned Corporation or Authority, shall be attended to forthwith and, in
// any event, within forty-eight hours". It says potholes. It does not say road
// defects, and a faded lane marking is not covered by it.
//
// Keying the order on the category made the platform quote a pothole order
// beside a missing manhole cover, which overstates the law in the one
// direction that matters: the share kit prints the quoted paragraph next to
// the photograph, and a forwarded post is not correctable.
var subcategoryDeadlineKeys = map[string]string{
	"pothole": "pothole.attend_hours.court",
}

// boundDeadline is a constant and the thing inside a category it binds.
// appliesTo is empty when it binds the whole category.
type boundDeadline struct {
	key       string
	appliesTo string
}

// categoryScopes names, in words a resident would read, what a subcategory
// constant actually covers. A scope with no wording here is not published on a
// surface that lists categories, because "48 hours" beside a category it does
// not bind is the overstatement D138 removed.
var categoryScopes = map[string]string{"pothole": "potholes"}

// deadlinesWithin lists the obligations that bind anything inside a category.
//
// The ward page lists desks by category while the court's forty-eight hours
// binds potholes, so scoping the constant correctly dropped the obligation
// from every ward page — the page asked by category and found nothing.
// Dropping it is wrong, since the obligation is real and a resident should see
// it; attaching it to the category is the overstatement that was just removed.
// So it is returned with the scope it has, and the surface states both.
func deadlinesWithin(category string) []boundDeadline {
	category = strings.TrimSpace(category)
	var out []boundDeadline
	if k, ok := deadlineKeys[category]; ok {
		out = append(out, boundDeadline{key: k})
	}
	// Deterministic order: the taxonomy is small and a map range would reorder
	// a public page between requests.
	subs := make([]string, 0, len(subcategoryDeadlineKeys))
	for sub := range subcategoryDeadlineKeys {
		subs = append(subs, sub)
	}
	sort.Strings(subs)
	for _, sub := range subs {
		if classify.CategoryOf(sub) != category {
			continue
		}
		scope, ok := categoryScopes[sub]
		if !ok {
			continue
		}
		out = append(out, boundDeadline{key: subcategoryDeadlineKeys[sub], appliesTo: scope})
	}
	return out
}

// deadlineKeyFor picks the narrowest constant that actually binds this thing,
// and returns empty when none does. Empty means no clock is shown at all —
// an obligation we cannot cite is not an obligation with a default.
func deadlineKeyFor(category, subcategory string) string {
	if k, ok := subcategoryDeadlineKeys[strings.ToLower(strings.TrimSpace(subcategory))]; ok {
		return k
	}
	return deadlineKeys[strings.TrimSpace(category)]
}

// RoutingFor returns the department responsible for a category in a ward, with
// the deadline that binds it.
//
// Several departments can share a category — a road defect in R/C has two,
// split by road class and by season — so the first by effective date is
// returned and the rest are the caller's problem only when it has to choose.
// Not found is not an error: most wards have no department recorded, and that
// is a state the product states plainly rather than hides.
func (d *DB) RoutingFor(ctx context.Context, authority, ward, category, subcategory string) (Routing, bool, error) {
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

	key := deadlineKeyFor(category, subcategory)
	if key == "" {
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
	// DeadlineAppliesTo names what inside the category the obligation binds,
	// empty when it binds the whole category. A desk is listed by category and
	// the court's forty-eight hours binds potholes, so the scope has to travel
	// with the number or the page overstates the law.
	DeadlineAppliesTo string
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
	//
	// A desk is listed by category and an obligation may bind only part of one,
	// so the scope travels with the number. "48 hours" beside a category the
	// order does not cover is the overstatement D138 removed; dropping it
	// altogether loses an obligation a resident should see.
	for i := range p.Desks {
		for _, bound := range deadlinesWithin(p.Desks[i].Category) {
			var value, citation string
			err := d.pool.QueryRow(ctx, `
				SELECT value::text, citation FROM legal_constants
				 WHERE key = $1 AND effective_from <= now()::date
				   AND (effective_to IS NULL OR effective_to >= now()::date)
				 ORDER BY effective_from DESC LIMIT 1`, bound.key).Scan(&value, &citation)
			if err != nil {
				continue
			}
			hours, convErr := strconv.Atoi(strings.Trim(value, `"`))
			if convErr != nil || citation == "" {
				continue
			}
			// The first obligation found wins the desk's own fields, which is
			// what a surface with one slot renders.
			if p.Desks[i].DeadlineHours == 0 {
				p.Desks[i].DeadlineHours = hours
				p.Desks[i].DeadlineCitation = citation
				p.Desks[i].DeadlineAppliesTo = bound.appliesTo
			}
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
