package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vinit-churi/tracesarkar/internal/redact"
)

// This file holds the adversarial redaction set: photographs a person has
// marked up with every face and plate they can see.
//
// It is the only thing that can turn the detector's behaviour into a number.
// Run over the captures that existed when it shipped, the detector reached a
// verdict on all of them and missed nothing — because not one of them
// contained a face or a plate. Recall was not good or bad, it was unmeasured,
// and ADR 0008 has required this set since before the detector was written.

// PendingTruth is a photograph nobody has marked up yet.
type PendingTruth struct {
	MediaID     string
	ReportID    string
	ArchiveKey  string
	ContentType string
	CapturedAt  time.Time
}

// RedactionTruth is one person's judgement about one photograph.
type RedactionTruth struct {
	MediaID  string
	ReportID string
	// Boxes is every face and plate visible in the photograph. Empty is a
	// judgement: somebody looked and there is nothing to find.
	Boxes []redact.Box
	// PeoplePresent is whether anybody is visible at all, which a face box
	// does not cover — a person photographed from behind is present with no
	// face to mark, and that combination is what the gate holds an image on.
	PeoplePresent bool
	Conditions    []string
	Notes         string
	AnnotatedBy   string
}

// TruthCase is an annotated photograph as the scorer reads it.
type TruthCase struct {
	MediaID       string
	ReportID      string
	ArchiveKey    string
	Boxes         []redact.Box
	PeoplePresent bool
	Conditions    []string
	Notes         string
}

// MediaAwaitingTruth returns the account's own photographs with no judgement
// recorded, oldest first.
//
// Scoped to one account for the same reason the labelling queue is: the
// photographs are somebody's, serving another person's capture is refused
// elsewhere, and the ids and timestamps are a disclosure on their own.
func (d *DB) MediaAwaitingTruth(ctx context.Context, accountID string, limit int) ([]PendingTruth, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := d.pool.Query(ctx, `
		SELECT m.id::text, m.report_id::text, m.archive_key,
		       COALESCE(m.content_type, ''), r.captured_at
		  FROM report_media m
		  JOIN reports r ON r.id = m.report_id
		 WHERE r.account_id::text = $1
		   AND NOT EXISTS (
		    SELECT 1 FROM redaction_truth t WHERE t.media_id = m.id)
		 ORDER BY r.captured_at
		 LIMIT $2`, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("find media awaiting ground truth: %w", err)
	}
	defer rows.Close()

	var out []PendingTruth
	for rows.Next() {
		var p PendingTruth
		if err := rows.Scan(&p.MediaID, &p.ReportID, &p.ArchiveKey,
			&p.ContentType, &p.CapturedAt); err != nil {
			return nil, fmt.Errorf("scan media awaiting ground truth: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SaveRedactionTruth records a judgement, replacing any earlier one for the
// same photograph. Replacing rather than appending because two sets of ground
// truth for one photograph would be counted twice by everything downstream.
//
// It refuses what cannot be scored instead of storing it. A box reaching past
// the edge of the image, a condition outside the vocabulary, or a face marked
// up with nobody present would each distort a figure quietly, long after the
// annotation that caused it was forgotten.
func (d *DB) SaveRedactionTruth(ctx context.Context, in RedactionTruth) error {
	if in.MediaID == "" || in.ReportID == "" {
		return fmt.Errorf("save redaction truth: the judgement names no photograph")
	}
	for _, b := range in.Boxes {
		if err := b.Valid(); err != nil {
			return fmt.Errorf("save redaction truth for %s: %w", in.MediaID, err)
		}
		if b.Kind == redact.Face && !in.PeoplePresent {
			return fmt.Errorf("save redaction truth for %s: a face is marked up but "+
				"the judgement says nobody is present", in.MediaID)
		}
	}
	for _, c := range in.Conditions {
		if !redact.KnownCondition(c) {
			return fmt.Errorf("save redaction truth for %s: %q is not a condition the "+
				"breakdown knows", in.MediaID, c)
		}
	}

	boxes := in.Boxes
	if boxes == nil {
		boxes = []redact.Box{}
	}
	body, err := json.Marshal(boxes)
	if err != nil {
		return fmt.Errorf("save redaction truth for %s: %w", in.MediaID, err)
	}
	conditions := in.Conditions
	if conditions == nil {
		conditions = []string{}
	}
	var by *string
	if in.AnnotatedBy != "" {
		by = &in.AnnotatedBy
	}

	_, err = d.pool.Exec(ctx, `
		INSERT INTO redaction_truth
		  (media_id, report_id, boxes, people_present, conditions, notes, annotated_by, annotated_at)
		VALUES ($1::uuid, $2::uuid, $3::jsonb, $4, $5, $6, $7::uuid, now())
		ON CONFLICT (media_id) DO UPDATE SET
		  boxes          = EXCLUDED.boxes,
		  people_present = EXCLUDED.people_present,
		  conditions     = EXCLUDED.conditions,
		  notes          = EXCLUDED.notes,
		  annotated_by   = EXCLUDED.annotated_by,
		  annotated_at   = now()`,
		in.MediaID, in.ReportID, body, in.PeoplePresent, conditions, in.Notes, by)
	if err != nil {
		return fmt.Errorf("save redaction truth for %s: %w", in.MediaID, err)
	}
	return nil
}

// RedactionTruthSet returns every annotated photograph, oldest judgement
// first, for the scorer to run the detectors over.
//
// Not scoped to an account. The set is the platform's measurement of its own
// detector, and a figure computed from one person's subset of it would be a
// different figure depending on who ran the command.
func (d *DB) RedactionTruthSet(ctx context.Context, limit int) ([]TruthCase, error) {
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}
	rows, err := d.pool.Query(ctx, `
		SELECT t.media_id::text, t.report_id::text, m.archive_key,
		       t.boxes, t.people_present, t.conditions, t.notes
		  FROM redaction_truth t
		  JOIN report_media m ON m.id = t.media_id
		 ORDER BY t.annotated_at
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("read the redaction truth set: %w", err)
	}
	defer rows.Close()

	var out []TruthCase
	for rows.Next() {
		var c TruthCase
		var body []byte
		if err := rows.Scan(&c.MediaID, &c.ReportID, &c.ArchiveKey, &body,
			&c.PeoplePresent, &c.Conditions, &c.Notes); err != nil {
			return nil, fmt.Errorf("scan redaction truth: %w", err)
		}
		if err := json.Unmarshal(body, &c.Boxes); err != nil {
			return nil, fmt.Errorf("read boxes for %s: %w", c.MediaID, err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// MediaRef is where one photograph is held and who it belongs to.
type MediaRef struct {
	Key         string
	ContentType string
	Owner       string
	ReportID    string
}

// MediaByID looks up one photograph by its own id.
//
// By media id rather than by report id because a report may carry more than
// one photograph, and ground truth belongs to the photograph that was marked
// up rather than to whichever of them happens to be first.
//
// An id that is not a uuid is "not found" rather than an error: it is a
// mistyped URL, not a fault.
func (d *DB) MediaByID(ctx context.Context, mediaID string) (MediaRef, error) {
	var m MediaRef
	err := d.pool.QueryRow(ctx, `
		SELECT m.archive_key, COALESCE(m.content_type, ''),
		       r.account_id::text, m.report_id::text
		  FROM report_media m
		  JOIN reports r ON r.id = m.report_id
		 WHERE m.id = $1::uuid`, mediaID).
		Scan(&m.Key, &m.ContentType, &m.Owner, &m.ReportID)
	switch {
	case errors.Is(err, pgx.ErrNoRows), err != nil && isBadUUID(err):
		return MediaRef{}, ErrNotFound
	case err != nil:
		return MediaRef{}, fmt.Errorf("media by id: %w", err)
	}
	return m, nil
}
