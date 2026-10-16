package polish

import (
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

func facts() share.Facts {
	lodged := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	return share.Facts{
		IssueID:          "018f0000-0000-7000-8000-000000000001",
		Permalink:        "https://tracesarkar.org/i/abc",
		Authority:        "BMC",
		Ward:             "K/E",
		Category:         "road_defect",
		Subcategory:      "pothole",
		Status:           "open",
		FirstSeen:        time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
		Sightings:        4,
		Lodged:           true,
		LodgedOn:         &lodged,
		DeadlineHours:    48,
		DeadlineState:    "elapsed",
		DeadlineShortRef: "HC order",
		DeadlineCitation: "Bombay High Court order in PIL 85 of 2013",
		Handles: []share.Handle{{
			Platform: "x", Handle: "@mybmc", Kind: share.Institutional,
			SourceRef: "https://portal.mcgm.gov.in/", RetrievedAt: lodged, FirstHand: true,
		}},
	}
}

func baseline(t *testing.T, lang share.Lang, ch share.Channel) share.Post {
	t.Helper()
	p, err := share.Compose(facts(), lang, ch)
	if err != nil {
		t.Fatalf("Compose(%s,%s): %v", lang, ch, err)
	}
	return p
}

// The permalink and the hashtag run are tokens other people resolve and
// search for. There is no better wording for them, so they are never sent.
func TestThePermalinkAndTheHashtagsAreNeverSentToAModel(t *testing.T) {
	f := facts()
	p := baseline(t, share.EN, share.WhatsApp)
	segs := Split(f, p)
	if len(segs) == 0 {
		t.Fatal("Split returned nothing")
	}
	for _, s := range Prose(segs) {
		if strings.Contains(s.Text, f.Permalink) {
			t.Errorf("line %d would send the permalink: %q", s.Index, s.Text)
		}
		if strings.Contains(s.Text, "#") {
			t.Errorf("line %d would send a hashtag: %q", s.Index, s.Text)
		}
	}
	var literal int
	for _, s := range segs {
		if !s.Prose {
			literal++
		}
	}
	if literal != 2 {
		t.Errorf("marked %d lines literal, want the permalink and the hashtag run", literal)
	}
}

// Reassembling with no change at all must reproduce the post byte for byte,
// or the polish path quietly rewrites posts it was not asked to touch.
func TestReassemblingNothingReproducesTheBaselineExactly(t *testing.T) {
	for _, lang := range share.Languages() {
		for _, ch := range share.Channels() {
			p := baseline(t, lang, ch)
			segs := Split(facts(), p)
			if got := Reassemble(segs, nil); got != p.Text {
				t.Errorf("%s/%s round trip changed the post:\nwant %q\ngot  %q",
					lang, ch, p.Text, got)
			}
		}
	}
}

// A literal line is put back from the baseline, not from the model, so a
// model that returns one cannot have it published.
func TestALiteralLineIsTakenFromTheBaselineNotTheModel(t *testing.T) {
	f := facts()
	p := baseline(t, share.EN, share.WhatsApp)
	segs := Split(f, p)
	var literal int
	for _, s := range segs {
		if !s.Prose {
			literal = s.Index
		}
	}
	out := Reassemble(segs, map[int]string{literal: "https://evil.example/phish"})
	if strings.Contains(out, "evil.example") {
		t.Errorf("a model-supplied literal line reached the post:\n%s", out)
	}
}
