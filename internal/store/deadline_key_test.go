package store

import "testing"

// The 48-hour duty is pothole-specific. Paragraph 70(ix) of the Bombay High
// Court's order reads "All potholes, once brought to the notice of the
// concerned Corporation or Authority, shall be attended to forthwith and, in
// any event, within forty-eight hours" — it does not say "all road defects".
//
// Attaching it to a faded road marking overstates the law, and the share kit
// quotes the paragraph next to whatever it is attached to, so a wrong mapping
// becomes a public post quoting a court order about potholes beside a
// photograph of something else.
func TestTheCourtsPotholeDeadlineBindsPotholesOnly(t *testing.T) {
	cases := []struct {
		category, subcategory, want string
	}{
		{"road_defect", "pothole", "pothole.attend_hours.court"},
		{"road_defect", "Pothole", "pothole.attend_hours.court"},
		{"road_defect", "missing manhole cover", ""},
		{"road_defect", "open manhole", ""},
		{"road_defect", "missing/faded markings", ""},
		{"road_defect", "damaged speed breaker", ""},
		{"road_defect", "", ""},
		{"waste", "overflowing bin", "waste.refuse_removal_hours"},
		{"waste", "", "waste.refuse_removal_hours"},
		{"water_drainage", "blocked drain", ""},
	}
	for _, c := range cases {
		got := deadlineKeyFor(c.category, c.subcategory)
		if got != c.want {
			t.Errorf("deadlineKeyFor(%q, %q) = %q, want %q",
				c.category, c.subcategory, got, c.want)
		}
	}
}

// The ward page lists desks by category, but the court's forty-eight hours
// binds a subcategory. Scoping the constant correctly (D138) silently dropped
// the obligation from every ward page, because the page looked it up by
// category and found nothing.
//
// Dropping it is wrong: the obligation is real and a resident reading a ward
// page should see it. Attaching it to the whole category is also wrong. So it
// is shown with the scope it actually has.
func TestACategoryReportsTheObligationsBoundWithinIt(t *testing.T) {
	cases := []struct {
		category string
		wantKey  string
		wantFor  string
	}{
		{"road_defect", "pothole.attend_hours.court", "potholes"},
		{"waste", "waste.refuse_removal_hours", ""},
		{"water_drainage", "", ""},
		{"street_furniture", "", ""},
	}
	for _, c := range cases {
		got := deadlinesWithin(c.category)
		if c.wantKey == "" {
			if len(got) != 0 {
				t.Errorf("%s: invented %d obligation(s): %+v", c.category, len(got), got)
			}
			continue
		}
		if len(got) != 1 {
			t.Errorf("%s: got %d obligations, want 1", c.category, len(got))
			continue
		}
		if got[0].key != c.wantKey {
			t.Errorf("%s: key %q, want %q", c.category, got[0].key, c.wantKey)
		}
		if got[0].appliesTo != c.wantFor {
			t.Errorf("%s: scope %q, want %q", c.category, got[0].appliesTo, c.wantFor)
		}
	}
}
