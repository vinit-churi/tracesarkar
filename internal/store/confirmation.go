package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/vinit-churi/tracesarkar/internal/classify"
)

// Errors a caller is expected to handle.
var (
	// ErrNotAwaitingConfirmation covers four cases on purpose: no such
	// capture, somebody else's capture, a capture the platform never asked
	// about, and one already answered. Telling them apart would let anybody
	// enumerate other people's captures by watching which id gives which
	// error.
	ErrNotAwaitingConfirmation = errors.New("that capture is not waiting for an answer")
	// ErrNotAnOfferedOption is a chosen value that was not on the list.
	ErrNotAnOfferedOption = errors.New("that is not one of the options offered")
)

// AwaitingConfirmation is a capture the platform asked about and nobody has
// answered.
type AwaitingConfirmation struct {
	ReportID    string   `json:"report_id"`
	CapturedAt  string   `json:"captured_at"`
	Ward        string   `json:"ward"`
	Authority   string   `json:"authority"`
	Category    string   `json:"category"`
	Subcategory string   `json:"subcategory"`
	Rationale   string   `json:"rationale"`
	Confidence  float64  `json:"confidence"`
	Options     []string `json:"options"`
	// AsksForCategory is whether the options are categories rather than
	// subcategories, because the two questions read differently to a person:
	// "which of these is it" against "what kind of problem is this".
	AsksForCategory bool   `json:"asks_for_category"`
	MediaKey        string `json:"-"`
}

// CapturesAwaitingConfirmation lists this account's unanswered questions,
// oldest first.
//
// Scoped to one account and never to all of them. A capture carries a
// position accurate to metres and a photograph that has not necessarily been
// redacted yet; the only person who may see the list is the one who took it.
func (d *DB) CapturesAwaitingConfirmation(ctx context.Context, accountID string, limit int) ([]AwaitingConfirmation, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := d.pool.Query(ctx, `
		SELECT r.id::text,
		       to_char(r.captured_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
		       COALESCE(j.ward,''), COALESCE(j.authority,''),
		       COALESCE(c.category,''), COALESCE(c.subcategory,''),
		       COALESCE(c.rationale,''), COALESCE(c.confidence,0)::float8,
		       c.alternatives
		  FROM reports r
		  JOIN LATERAL (
		    SELECT category, subcategory, rationale, confidence, outcome, alternatives
		      FROM classifications
		     WHERE report_id = r.id AND error IS NULL
		     ORDER BY created_at DESC LIMIT 1
		  ) c ON true
		  LEFT JOIN report_jurisdiction j ON j.report_id = r.id AND j.error IS NULL
		 WHERE r.account_id = $1::uuid
		   AND c.outcome = 'needs_confirmation'
		   AND NOT EXISTS (SELECT 1 FROM issue_reports ir WHERE ir.report_id = r.id)
		 ORDER BY r.captured_at
		 LIMIT $2`, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("captures awaiting confirmation: %w", err)
	}
	defer rows.Close()

	var out []AwaitingConfirmation
	for rows.Next() {
		var a AwaitingConfirmation
		var raw []byte
		if err := rows.Scan(&a.ReportID, &a.CapturedAt, &a.Ward, &a.Authority,
			&a.Category, &a.Subcategory, &a.Rationale, &a.Confidence, &raw); err != nil {
			return nil, fmt.Errorf("scan awaiting confirmation: %w", err)
		}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &a.Options); err != nil {
				return nil, fmt.Errorf("decode options for %s: %w", a.ReportID, err)
			}
		}
		// A question asked before migration 0046 has no stored options, and
		// the taxonomy is the only honest fallback: better the current list
		// than no question at all for a capture that is otherwise stuck for
		// ever.
		if len(a.Options) == 0 {
			if classify.KnownCategory(a.Category) {
				a.Options = classify.Subcategories(a.Category)
			} else {
				a.Options = classify.TopCategories()
			}
		}
		a.AsksForCategory = !classify.KnownCategory(a.Category) ||
			len(a.Options) > 0 && classify.KnownCategory(a.Options[0])
		out = append(out, a)
	}
	return out, rows.Err()
}

// ConfirmClassification records the reporter's answer.
//
// It writes a new classification row rather than editing the model's. The
// model's attempt is evidence of what the classifier did and the eval set
// grades against it; overwriting it would destroy the measurement and leave
// no trace that a question was ever asked. The new row carries
// `source = 'reporter'`, so it routes the capture and is excluded from the
// classifier's score.
//
// The chosen value is checked against the options that were **offered**,
// taken from the stored row and not from today's taxonomy, so a person
// answering a question asked in September answers from the list they saw.
func (d *DB) ConfirmClassification(ctx context.Context, reportID, accountID, chosen string) (classify.Decision, error) {
	var dec classify.Decision
	var raw []byte
	err := d.pool.QueryRow(ctx, `
		SELECT COALESCE(c.category,''), COALESCE(c.subcategory,''),
		       COALESCE(c.severity,''), COALESCE(c.hazard_to_life,false),
		       COALESCE(c.water_present,false), COALESCE(c.people_present,false),
		       COALESCE(c.image_quality,''), COALESCE(c.redaction_required,false),
		       c.alternatives
		  FROM reports r
		  JOIN LATERAL (
		    SELECT category, subcategory, severity, hazard_to_life, water_present,
		           people_present, image_quality, redaction_required, outcome,
		           alternatives
		      FROM classifications
		     WHERE report_id = r.id AND error IS NULL
		     ORDER BY created_at DESC LIMIT 1
		  ) c ON true
		 WHERE r.id = $1::uuid AND r.account_id = $2::uuid
		   AND c.outcome = 'needs_confirmation'
		   AND NOT EXISTS (SELECT 1 FROM issue_reports ir WHERE ir.report_id = r.id)`,
		reportID, accountID).Scan(
		&dec.Category, &dec.Subcategory, &dec.Severity, &dec.HazardToLife,
		&dec.WaterPresent, &dec.PeoplePresent, &dec.ImageQuality,
		&dec.RedactionRequired, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return classify.Decision{}, ErrNotAwaitingConfirmation
	}
	if err != nil {
		return classify.Decision{}, fmt.Errorf("read the question for %s: %w", reportID, err)
	}

	var offered []string
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &offered); err != nil {
			return classify.Decision{}, fmt.Errorf("decode options for %s: %w", reportID, err)
		}
	}
	if len(offered) == 0 {
		if classify.KnownCategory(dec.Category) {
			offered = classify.Subcategories(dec.Category)
		} else {
			offered = classify.TopCategories()
		}
	}
	var ok bool
	for _, o := range offered {
		if o == chosen {
			ok = true
			break
		}
	}
	if !ok {
		return classify.Decision{}, ErrNotAnOfferedOption
	}

	// A category answer replaces the category and clears the subcategory the
	// model invented; a subcategory answer keeps the category it belongs to.
	if classify.KnownCategory(chosen) {
		dec.Category, dec.Subcategory = chosen, ""
	} else {
		dec.Subcategory = chosen
	}

	// The decision is re-derived rather than asserted, so a confirmed answer
	// goes through the same coverage and hazard rules as a model's: a
	// category nobody routes yet still comes out as not_yet_covered, and the
	// known-hazard list still overrides.
	dec.IsCivicIssue = true
	dec.Confidence = 1
	dec.Rationale = "Confirmed by the person who took the photograph."
	dec.Alternatives = nil

	// Coverage is per ward — a department recorded for one ward says nothing
	// about the next — so the capture's own ward decides, and a capture whose
	// ward has not resolved yet is not known to be covered at all. A
	// confirmed answer goes through the same rule as a model's: a category
	// nobody routes here still comes out as not_yet_covered rather than being
	// attached to an issue nothing can act on.
	var ward string
	if err := d.pool.QueryRow(ctx, `
		SELECT COALESCE(j.ward,'') FROM report_jurisdiction j
		 WHERE j.report_id = $1::uuid AND j.error IS NULL
		 ORDER BY j.created_at DESC LIMIT 1`, reportID).Scan(&ward); err != nil &&
		!errors.Is(err, pgx.ErrNoRows) {
		return classify.Decision{}, fmt.Errorf("read ward for %s: %w", reportID, err)
	}
	covered, err := d.CoverageForWard(ctx, ward)
	if err != nil {
		return classify.Decision{}, err
	}
	dec.Outcome = classify.Accepted
	if !covered.Covers(dec.Category) {
		dec.Outcome = classify.NotYetCovered
	}

	if err := d.SaveClassification(ctx, NewClassification{
		ReportID: reportID,
		Decision: dec,
		Model:    "none",
		Prompt:   "none",
		Source:   "reporter",
	}); err != nil {
		return classify.Decision{}, err
	}
	return dec, nil
}

// CoverageForWard is the categories a ward can route, as a classify.Coverage.
//
// A thin wrapper on CoverageByWard, and it keeps that method's rule: "covered
// somewhere" is not "covered here", so an unknown or empty ward covers
// nothing rather than everything.
func (d *DB) CoverageForWard(ctx context.Context, ward string) (classify.Coverage, error) {
	if ward == "" {
		return classify.Coverage{}, nil
	}
	byWard, err := d.CoverageByWard(ctx)
	if err != nil {
		return nil, fmt.Errorf("coverage for ward %s: %w", ward, err)
	}
	return classify.CoverageFor(byWard[ward]...), nil
}
