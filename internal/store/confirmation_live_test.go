package store_test

import (
	"errors"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/classify"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// The three captures the operator console found: the platform asked which
// problem they were, and until now no surface could collect the answer, so
// they sat unattached from 5 October. Hard rule 7 is "never lose a captured
// report", and a capture nobody can finish is lost in the way that matters.
func TestLiveTheQuestionsThePlatformAskedAreReadable(t *testing.T) {
	ctx, pool, db := liveDB(t)

	var account string
	err := pool.QueryRow(ctx, `
		SELECT r.account_id::text FROM reports r
		  JOIN LATERAL (
		    SELECT outcome FROM classifications
		     WHERE report_id = r.id AND error IS NULL
		     ORDER BY created_at DESC LIMIT 1
		  ) c ON true
		 WHERE c.outcome = 'needs_confirmation' AND r.account_id IS NOT NULL
		 LIMIT 1`).Scan(&account)
	if err != nil {
		t.Skipf("no capture is awaiting confirmation: %v", err)
	}

	pending, err := db.CapturesAwaitingConfirmation(ctx, account, 50)
	if err != nil {
		t.Fatalf("CapturesAwaitingConfirmation: %v", err)
	}
	if len(pending) == 0 {
		t.Fatal("the account owns an unanswered capture and the list is empty")
	}
	for _, p := range pending {
		t.Logf("%s  %s  %s/%s  conf=%.2f  asks_for_category=%v  %d options",
			p.ReportID[:8], p.CapturedAt[:10], p.Category, p.Subcategory,
			p.Confidence, p.AsksForCategory, len(p.Options))
		if len(p.Options) == 0 {
			t.Errorf("%s is a question with no options, which nobody can answer", p.ReportID)
		}
		// Every option has to be something the platform can act on, or the
		// answer leaves the capture exactly where it was.
		for _, o := range p.Options {
			if p.AsksForCategory {
				if !classify.KnownCategory(o) {
					t.Errorf("%s offers category %q, which is not in the taxonomy",
						p.ReportID, o)
				}
				continue
			}
			if !classify.KnownSubcategory(p.Category, o) {
				t.Errorf("%s offers %q, which is not a subcategory of %q",
					p.ReportID, o, p.Category)
			}
		}
	}
}

// One account must never see another's captures: a capture carries a position
// accurate to metres and a photograph that may not be redacted yet.
func TestLiveTheQuestionListIsScopedToOneAccount(t *testing.T) {
	ctx, _, db := liveDB(t)
	// A well-formed uuid that owns nothing.
	pending, err := db.CapturesAwaitingConfirmation(ctx,
		"00000000-0000-0000-0000-000000000000", 50)
	if err != nil {
		t.Fatalf("CapturesAwaitingConfirmation: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("an account that owns nothing was shown %d captures", len(pending))
	}
}

// Four different refusals, one error. Telling them apart would let anybody
// enumerate other people's captures by watching which id gives which answer.
func TestLiveConfirmingRefusesIdentically(t *testing.T) {
	ctx, pool, db := liveDB(t)

	var reportID, account string
	err := pool.QueryRow(ctx, `
		SELECT r.id::text, COALESCE(r.account_id::text,'')
		  FROM reports r
		  JOIN LATERAL (
		    SELECT outcome FROM classifications
		     WHERE report_id = r.id AND error IS NULL
		     ORDER BY created_at DESC LIMIT 1
		  ) c ON true
		 WHERE c.outcome = 'needs_confirmation' LIMIT 1`).Scan(&reportID, &account)
	if err != nil {
		t.Skipf("no capture is awaiting confirmation: %v", err)
	}

	cases := []struct {
		name, report, acct, chosen string
		want                       error
	}{
		{"somebody else's capture", reportID,
			"00000000-0000-0000-0000-000000000000", "pothole",
			store.ErrNotAwaitingConfirmation},
		{"a capture that does not exist", "00000000-0000-0000-0000-000000000001",
			account, "pothole", store.ErrNotAwaitingConfirmation},
		{"an option that was not offered", reportID, account,
			"a problem nobody listed", store.ErrNotAnOfferedOption},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := db.ConfirmClassification(ctx, c.report, c.acct, c.chosen); !errors.Is(err, c.want) {
				t.Errorf("got %v, want %v", err, c.want)
			}
		})
	}
}
