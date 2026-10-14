package store_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/redact"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// A photograph with ground truth recorded against it is the unit of the
// adversarial set ADR 0008 required. These tests are the storage half: that an
// unmarked photograph is offered, that a judgement takes it out of the queue,
// and that the database refuses a judgement that could not be scored.
func TestLiveRedactionTruthIsRecordedAndRead(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM redaction_truth WHERE report_id IN
			(SELECT id FROM reports WHERE account_id = $1)`, account)
		_, _ = pool.Exec(c, `DELETE FROM report_media WHERE report_id IN
			(SELECT id FROM reports WHERE account_id = $1)`, account)
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	stamp := time.Now().UTC()
	reportID, _, err := db.SaveReport(ctx, store.NewReport{
		AccountID: account, Lat: 19.2094, Lon: 72.8348, AccuracyM: 6,
		CapturedAt: stamp, Idempotency: "truth-" + stamp.Format("150405.000000000"),
	})
	if err != nil {
		t.Fatalf("SaveReport: %v", err)
	}
	digest := sha256.Sum256([]byte("truth-test-" + reportID))
	if err := db.AddReportMedia(ctx, store.NewMedia{
		ReportID: reportID, Role: "close", ArchiveKey: "media/" + reportID + "/x",
		ContentType: "image/jpeg", Bytes: 1234, SHA256: hex.EncodeToString(digest[:]),
		CapturedAt: &stamp,
	}); err != nil {
		t.Fatalf("AddReportMedia: %v", err)
	}

	queued := func(t *testing.T) (store.PendingTruth, bool) {
		t.Helper()
		pending, err := db.MediaAwaitingTruth(ctx, account, 500)
		if err != nil {
			t.Fatalf("MediaAwaitingTruth: %v", err)
		}
		for _, p := range pending {
			if p.ReportID == reportID {
				return p, true
			}
		}
		return store.PendingTruth{}, false
	}

	p, ok := queued(t)
	if !ok {
		t.Fatal("a photograph nobody has marked up was not offered for annotation")
	}
	if p.ArchiveKey == "" || p.MediaID == "" {
		t.Fatalf("queue item carries nothing to fetch: %+v", p)
	}

	in := store.RedactionTruth{
		MediaID:       p.MediaID,
		ReportID:      reportID,
		Boxes:         []redact.Box{{Kind: redact.Face, MinX: 0.1, MinY: 0.1, MaxX: 0.2, MaxY: 0.25}},
		PeoplePresent: true,
		Conditions:    []string{"night", "distance"},
		Notes:         "two pedestrians, one facing the camera",
		AnnotatedBy:   account,
	}
	if err := db.SaveRedactionTruth(ctx, in); err != nil {
		t.Fatalf("SaveRedactionTruth: %v", err)
	}

	if _, still := queued(t); still {
		t.Error("a photograph already marked up is still being offered")
	}

	set, err := db.RedactionTruthSet(ctx, 500)
	if err != nil {
		t.Fatalf("RedactionTruthSet: %v", err)
	}
	var got *store.TruthCase
	for i := range set {
		if set[i].ReportID == reportID {
			got = &set[i]
		}
	}
	if got == nil {
		t.Fatal("the recorded judgement is not in the set the scorer reads")
	}
	if len(got.Boxes) != 1 || got.Boxes[0].Kind != redact.Face {
		t.Fatalf("boxes came back as %+v", got.Boxes)
	}
	if got.Boxes[0].MaxY != 0.25 {
		t.Errorf("box fractions did not survive the round trip: %+v", got.Boxes[0])
	}
	if !got.PeoplePresent || len(got.Conditions) != 2 || got.ArchiveKey == "" {
		t.Errorf("case came back as %+v", *got)
	}

	// Re-annotating replaces the judgement rather than adding a second one. A
	// photograph with two sets of ground truth would be counted twice.
	in.Boxes = nil
	in.PeoplePresent = false
	in.Conditions = []string{"daylight"}
	if err := db.SaveRedactionTruth(ctx, in); err != nil {
		t.Fatalf("SaveRedactionTruth again: %v", err)
	}
	set, err = db.RedactionTruthSet(ctx, 500)
	if err != nil {
		t.Fatalf("RedactionTruthSet after rework: %v", err)
	}
	seen := 0
	for _, c := range set {
		if c.ReportID == reportID {
			seen++
			if len(c.Boxes) != 0 {
				t.Errorf("the replaced judgement still carries boxes: %+v", c.Boxes)
			}
		}
	}
	if seen != 1 {
		t.Errorf("the photograph appears %d times in the set, want 1", seen)
	}
}

func TestLiveRedactionTruthRefusesWhatCannotBeScored(t *testing.T) {
	ctx, pool, db := liveDB(t)
	account := testAccount(t, ctx, db)
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, _ = pool.Exec(c, `DELETE FROM redaction_truth WHERE report_id IN
			(SELECT id FROM reports WHERE account_id = $1)`, account)
		_, _ = pool.Exec(c, `DELETE FROM report_media WHERE report_id IN
			(SELECT id FROM reports WHERE account_id = $1)`, account)
		_, _ = pool.Exec(c, `DELETE FROM reports WHERE account_id = $1`, account)
		_, _ = pool.Exec(c, `DELETE FROM accounts WHERE id = $1`, account)
	})

	stamp := time.Now().UTC()
	reportID, _, err := db.SaveReport(ctx, store.NewReport{
		AccountID: account, Lat: 19.2094, Lon: 72.8348, AccuracyM: 6,
		CapturedAt: stamp, Idempotency: "truth-bad-" + stamp.Format("150405.000000000"),
	})
	if err != nil {
		t.Fatalf("SaveReport: %v", err)
	}
	digest := sha256.Sum256([]byte("truth-bad-" + reportID))
	if err := db.AddReportMedia(ctx, store.NewMedia{
		ReportID: reportID, Role: "close", ArchiveKey: "media/" + reportID + "/x",
		ContentType: "image/jpeg", Bytes: 1234, SHA256: hex.EncodeToString(digest[:]),
		CapturedAt: &stamp,
	}); err != nil {
		t.Fatalf("AddReportMedia: %v", err)
	}
	pending, err := db.MediaAwaitingTruth(ctx, account, 500)
	if err != nil {
		t.Fatalf("MediaAwaitingTruth: %v", err)
	}
	var mediaID string
	for _, p := range pending {
		if p.ReportID == reportID {
			mediaID = p.MediaID
		}
	}
	if mediaID == "" {
		t.Fatal("no photograph to annotate")
	}

	base := store.RedactionTruth{MediaID: mediaID, ReportID: reportID, AnnotatedBy: account}

	cases := []struct {
		name string
		in   store.RedactionTruth
	}{
		{
			name: "a box off the edge of the photograph",
			in: func() store.RedactionTruth {
				c := base
				c.Boxes = []redact.Box{{Kind: redact.Plate, MinX: 0.9, MinY: 0.9, MaxX: 1.4, MaxY: 1.1}}
				return c
			}(),
		},
		{
			name: "a condition the breakdown does not know",
			in: func() store.RedactionTruth {
				c := base
				c.Conditions = []string{"nite"}
				return c
			}(),
		},
		{
			name: "a face marked up with nobody present",
			in: func() store.RedactionTruth {
				c := base
				c.Boxes = []redact.Box{{Kind: redact.Face, MaxX: 0.2, MaxY: 0.2}}
				c.PeoplePresent = false
				return c
			}(),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := db.SaveRedactionTruth(ctx, c.in); err == nil {
				t.Error("accepted a judgement that cannot be scored")
			}
		})
	}
}

// The Go layer refuses a face marked up with nobody present. This is the
// second lock, on the side that survives a code change: the same row is
// offered straight to the database, inside a transaction that is rolled back.
func TestLiveTheDatabaseItselfRefusesAFaceWithNobodyPresent(t *testing.T) {
	ctx, pool, _ := liveDB(t)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Ids that do not exist: the foreign keys would reject this row anyway,
	// so the test asserts on which constraint complained.
	_, err = tx.Exec(ctx, `
		INSERT INTO redaction_truth (media_id, report_id, boxes, people_present)
		VALUES (gen_random_uuid(), gen_random_uuid(),
		        '[{"kind": "face", "min_x": 0.1, "min_y": 0.1, "max_x": 0.2, "max_y": 0.2}]'::jsonb,
		        false)`)
	if err == nil {
		t.Fatal("the database accepted a face with nobody present")
	}
	if !strings.Contains(err.Error(), "a_face_means_a_person") {
		t.Errorf("rejected for the wrong reason: %v", err)
	}
}

// A box list that is not a list would break the scorer when it is read rather
// than when it is written.
func TestLiveTheDatabaseItselfRefusesBoxesThatAreNotAList(t *testing.T) {
	ctx, pool, _ := liveDB(t)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO redaction_truth (media_id, report_id, boxes, people_present)
		VALUES (gen_random_uuid(), gen_random_uuid(), '{"kind": "face"}'::jsonb, true)`)
	if err == nil {
		t.Fatal("the database accepted a box list that is not a list")
	}
	if !strings.Contains(err.Error(), "boxes_is_a_list") {
		t.Errorf("rejected for the wrong reason: %v", err)
	}
}
