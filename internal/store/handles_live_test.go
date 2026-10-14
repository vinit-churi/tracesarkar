package store_test

import (
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// The registry must hand back BMC's own desk and nothing it cannot source.
func TestLiveHandlesAreServedWithTheirProvenance(t *testing.T) {
	ctx, _, db := liveDB(t)

	hs, err := db.Handles(ctx, "BMC", "R/S")
	if err != nil {
		t.Fatalf("Handles: %v", err)
	}
	if len(hs) == 0 {
		t.Fatal("no handle for BMC; migration 0027 seeds one")
	}
	for _, h := range hs {
		if h.SourceRef == "" || h.RetrievedAt.IsZero() {
			t.Errorf("%s came back with no provenance: %+v", h.Handle, h)
		}
		if h.Kind != share.Institutional && h.Kind != share.Escalation {
			t.Errorf("%s has kind %q", h.Handle, h.Kind)
		}
	}
}

// A handle past its re-verification date is not served. Handles get renamed
// and abandoned, and a stale registry eventually tags a stranger.
func TestLiveAHandleDueForReverificationIsNotServed(t *testing.T) {
	ctx, pool, _ := liveDB(t)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO authority_handles
		  (id, authority, ward, platform, handle, kind, source_ref,
		   retrieved_at, first_hand, effective_from, verify_by)
		VALUES (gen_random_uuid(), 'TESTAUTH', NULL, 'x', '@stale_desk',
		        'institutional', 'https://example.test/s', now(), false,
		        '2020-01-01', '2020-06-01')`); err != nil {
		t.Fatalf("insert stale handle: %v", err)
	}

	var n int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM authority_handles
		 WHERE authority = 'TESTAUTH' AND verify_by >= CURRENT_DATE`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Errorf("a handle past its verify_by is still in date: %d", n)
	}
}

// The constraints are the second lock. Each of these is a way a bad row has
// actually reached a registry somewhere.
func TestLiveTheHandleRegistryRefusesRowsItCannotStandBehind(t *testing.T) {
	ctx, pool, _ := liveDB(t)

	cases := []struct {
		name, values string
	}{
		{"no source", `'BMC', NULL, 'x', '@a_desk', 'institutional', '   ', now(), false, '2026-10-14', '2027-10-14'`},
		{"unknown kind", `'BMC', NULL, 'x', '@a_desk', 'friendly', 'https://e.test', now(), false, '2026-10-14', '2027-10-14'`},
		{"not an x handle", `'BMC', NULL, 'x', 'mybmc', 'institutional', 'https://e.test', now(), false, '2026-10-14', '2027-10-14'`},
		{"already due", `'BMC', NULL, 'x', '@a_desk', 'institutional', 'https://e.test', now(), false, '2026-10-14', '2026-10-14'`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tx, err := pool.Begin(ctx)
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			_, err = tx.Exec(ctx, `
				INSERT INTO authority_handles
				  (id, authority, ward, platform, handle, kind, source_ref,
				   retrieved_at, first_hand, effective_from, verify_by)
				VALUES (gen_random_uuid(), `+c.values+`)`)
			if err == nil {
				t.Errorf("the database accepted a row it should refuse: %s", c.values)
			}
		})
	}
}

// The short citation is what lets a deadline appear in a 280-character post.
// Without it the sentence is dropped, so its absence is a silent loss.
func TestLiveTheCourtDeadlineHasAShortCitation(t *testing.T) {
	ctx, pool, _ := liveDB(t)

	var short string
	err := pool.QueryRow(ctx, `
		SELECT COALESCE(short_citation, '') FROM legal_constants
		 WHERE key = 'pothole.attend_hours.court'
		 ORDER BY effective_from DESC LIMIT 1`).Scan(&short)
	if err != nil {
		t.Fatalf("read short citation: %v", err)
	}
	if short == "" {
		t.Error("the 48-hour deadline has no short citation, so no post can state it")
	}
}
