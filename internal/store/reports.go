package store

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// NewReport is a capture as it arrives: a photograph's coordinates, its
// accuracy, and when the device says it was taken. Nothing is enriched yet.
type NewReport struct {
	AccountID   string
	Lat         float64
	Lon         float64
	AccuracyM   float64
	HeadingDeg  *float64
	CapturedAt  time.Time
	Description string
	Lang        string
	Idempotency string
	DeviceHash  string // hex, optional
}

// NewMedia is one image belonging to a report.
type NewMedia struct {
	ReportID    string
	Role        string // close | wide | noticeboard
	ArchiveKey  string
	ContentType string
	Bytes       int64
	SHA256      string // hex
	Width       int
	Height      int
	CapturedAt  *time.Time
}

// ReportLabel is the human judgement that makes a report usable as evaluation
// data. Without it a photograph is a photograph, not ground truth.
type ReportLabel struct {
	ReportID        string
	FrameType       string
	Label           string
	Conditions      []string
	WardGroundTruth string
	Notes           string
	LabelledBy      string
}

// Media as stored.
type Media struct {
	ID          string
	Role        string
	ArchiveKey  string
	ContentType string
	Bytes       int64
}

// Report is a stored capture with whatever is attached to it so far.
type Report struct {
	ID               string
	AccountID        string
	Status           string
	Lat              float64
	Lon              float64
	AccuracyM        float64
	CapturedAt       time.Time
	ReceivedAt       time.Time
	ClockSkewSeconds int
	Description      string
	Category         string
	Subcategory      string
	HazardToLife     bool
	Media            []Media
	Label            *ReportLabel
}

// EnsureAccount returns the account for a handle, creating it if needed. Phone
// verification arrives with v0.1; Phase 0 uses named local accounts so that
// field-kit captures carry an owner.
func (d *DB) EnsureAccount(ctx context.Context, handle string) (string, error) {
	var id string
	err := d.pool.QueryRow(ctx, `
		INSERT INTO accounts (id, display_handle)
		VALUES (gen_random_uuid(), $1)
		ON CONFLICT (phone_hash) DO NOTHING
		RETURNING id::text`, handle).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = d.pool.QueryRow(ctx,
			`SELECT id::text FROM accounts WHERE display_handle = $1 LIMIT 1`, handle).Scan(&id)
	}
	if err != nil {
		return "", fmt.Errorf("ensure account: %w", err)
	}
	return id, nil
}

// SaveReport persists a capture. It returns the report's id and whether this
// call created it: a retried upload returns the original rather than a
// duplicate. Nothing here enriches anything — that is the point (hard rule 7).
func (d *DB) SaveReport(ctx context.Context, in NewReport) (id string, created bool, err error) {
	if in.Lat < -90 || in.Lat > 90 || in.Lon < -180 || in.Lon > 180 {
		return "", false, fmt.Errorf("coordinates out of range: %v, %v", in.Lat, in.Lon)
	}
	if in.CapturedAt.IsZero() {
		return "", false, errors.New("captured_at is required")
	}

	var deviceHash []byte
	if in.DeviceHash != "" {
		decoded, decodeErr := hex.DecodeString(in.DeviceHash)
		if decodeErr != nil {
			return "", false, fmt.Errorf("device hash: %w", decodeErr)
		}
		deviceHash = decoded
	}

	// Accuracy is a fraction of a metre and the confidence gate reads it, so it
	// must not be coerced to an integer: NULLIF($n, 0) would do exactly that.
	var accuracy *float64
	if in.AccuracyM > 0 {
		accuracy = &in.AccuracyM
	}

	err = d.pool.QueryRow(ctx, `
		INSERT INTO reports (id, account_id, location, location_accuracy_m, heading_deg,
		                     captured_at, description, description_lang, idempotency_key,
		                     device_hash)
		VALUES (gen_random_uuid(), NULLIF($1,'')::uuid,
		        ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography,
		        $4::numeric, $5::numeric, $6, NULLIF($7,''), NULLIF($8,''), NULLIF($9,''), $10)
		RETURNING id::text`,
		in.AccountID, in.Lon, in.Lat, accuracy, in.HeadingDeg,
		in.CapturedAt, in.Description, in.Lang, in.Idempotency, deviceHash).Scan(&id)
	if err == nil {
		return id, true, nil
	}

	// A retry of the same upload: return the report already stored.
	if in.Idempotency != "" && isUniqueViolation(err) {
		lookupErr := d.pool.QueryRow(ctx, `
			SELECT id::text FROM reports
			WHERE account_id IS NOT DISTINCT FROM NULLIF($1,'')::uuid AND idempotency_key = $2`,
			in.AccountID, in.Idempotency).Scan(&id)
		if lookupErr == nil {
			return id, false, nil
		}
		return "", false, fmt.Errorf("save report (retry lookup): %w", lookupErr)
	}
	return "", false, fmt.Errorf("save report: %w", err)
}

// AddReportMedia attaches an image to a report.
func (d *DB) AddReportMedia(ctx context.Context, in NewMedia) error {
	sum, err := hex.DecodeString(in.SHA256)
	if err != nil {
		return fmt.Errorf("media sha256: %w", err)
	}
	_, err = d.pool.Exec(ctx, `
		INSERT INTO report_media (id, report_id, role, archive_key, content_type, bytes,
		                          sha256, width, height, captured_at)
		VALUES (gen_random_uuid(), $1::uuid, $2, $3, NULLIF($4,''), NULLIF($5,0)::bigint,
		        $6, NULLIF($7,0), NULLIF($8,0), $9)`,
		in.ReportID, in.Role, in.ArchiveKey, in.ContentType, in.Bytes,
		sum, in.Width, in.Height, in.CapturedAt)
	if err != nil {
		return fmt.Errorf("add media: %w", err)
	}
	return nil
}

// SaveReportLabel records the human judgement for an evaluation capture.
func (d *DB) SaveReportLabel(ctx context.Context, in ReportLabel) error {
	conditions := in.Conditions
	if conditions == nil {
		conditions = []string{}
	}
	_, err := d.pool.Exec(ctx, `
		INSERT INTO report_labels (report_id, frame_type, label, conditions,
		                           ward_ground_truth, notes, labelled_by)
		VALUES ($1::uuid, $2, $3, $4, NULLIF($5,''), NULLIF($6,''), NULLIF($7,'')::uuid)
		ON CONFLICT (report_id) DO UPDATE SET
		  frame_type = EXCLUDED.frame_type, label = EXCLUDED.label,
		  conditions = EXCLUDED.conditions,
		  ward_ground_truth = EXCLUDED.ward_ground_truth,
		  notes = EXCLUDED.notes, labelled_at = now()`,
		in.ReportID, in.FrameType, in.Label, conditions,
		in.WardGroundTruth, in.Notes, in.LabelledBy)
	if err != nil {
		return fmt.Errorf("save label: %w", err)
	}
	return nil
}

// GetReport returns a report with its media and label.
func (d *DB) GetReport(ctx context.Context, id string) (Report, error) {
	var r Report
	var accountID, description, category, subcategory *string
	err := d.pool.QueryRow(ctx, `
		SELECT id::text, account_id::text, status::text,
		       ST_Y(location::geometry), ST_X(location::geometry),
		       COALESCE(location_accuracy_m, 0), captured_at, received_at, clock_skew_s,
		       description, category::text, subcategory, hazard_to_life
		FROM reports WHERE id = $1::uuid`, id).Scan(
		&r.ID, &accountID, &r.Status, &r.Lat, &r.Lon, &r.AccuracyM,
		&r.CapturedAt, &r.ReceivedAt, &r.ClockSkewSeconds,
		&description, &category, &subcategory, &r.HazardToLife)
	if err != nil {
		return Report{}, fmt.Errorf("get report: %w", err)
	}
	r.AccountID = derefString(accountID)
	r.Description = derefString(description)
	r.Category = derefString(category)
	r.Subcategory = derefString(subcategory)

	mediaRows, err := d.pool.Query(ctx, `
		SELECT id::text, role, archive_key, COALESCE(content_type,''), COALESCE(bytes,0)
		FROM report_media WHERE report_id = $1::uuid ORDER BY created_at`, id)
	if err != nil {
		return Report{}, fmt.Errorf("get report media: %w", err)
	}
	defer mediaRows.Close()
	for mediaRows.Next() {
		var m Media
		if err := mediaRows.Scan(&m.ID, &m.Role, &m.ArchiveKey, &m.ContentType, &m.Bytes); err != nil {
			return Report{}, fmt.Errorf("scan media: %w", err)
		}
		r.Media = append(r.Media, m)
	}
	if err := mediaRows.Err(); err != nil {
		return Report{}, fmt.Errorf("read media: %w", err)
	}

	var label ReportLabel
	var ward, notes *string
	err = d.pool.QueryRow(ctx, `
		SELECT report_id::text, frame_type, label, conditions, ward_ground_truth, notes
		FROM report_labels WHERE report_id = $1::uuid`, id).Scan(
		&label.ReportID, &label.FrameType, &label.Label, &label.Conditions, &ward, &notes)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// Unlabelled is normal: most reports are not evaluation data.
	case err != nil:
		return Report{}, fmt.Errorf("get label: %w", err)
	default:
		label.WardGroundTruth = derefString(ward)
		label.Notes = derefString(notes)
		r.Label = &label
	}
	return r, nil
}

// LabelCounts reports how many labelled captures exist per label, which is how
// progress towards the evaluation set is measured.
func (d *DB) LabelCounts(ctx context.Context) (map[string]int, error) {
	rows, err := d.pool.Query(ctx, `SELECT label, count(*) FROM report_labels GROUP BY label`)
	if err != nil {
		return nil, fmt.Errorf("label counts: %w", err)
	}
	defer rows.Close()

	out := map[string]int{}
	for rows.Next() {
		var label string
		var n int
		if err := rows.Scan(&label, &n); err != nil {
			return nil, fmt.Errorf("scan label count: %w", err)
		}
		out[label] = n
	}
	return out, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
