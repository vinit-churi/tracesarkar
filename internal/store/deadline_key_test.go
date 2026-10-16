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
	// A category can bind more than one obligation, and `waste` now does: the
	// charter's twelve hours for the category and its forty-eight for a dead
	// animal. The page states each with the scope it has.
	cases := []struct {
		category string
		want     []boundDeadline
	}{
		{"road_defect", []boundDeadline{
			{key: "pothole.attend_hours.court", appliesTo: "potholes"},
		}},
		{"waste", []boundDeadline{
			{key: "waste.refuse_removal_hours"},
			{key: "waste.dead_animal_removal_hours", appliesTo: "a dead animal"},
		}},
		{"water_drainage", nil},
		{"street_furniture", nil},
	}
	for _, c := range cases {
		got := deadlinesWithin(c.category)
		if len(got) != len(c.want) {
			t.Errorf("%s: got %d obligation(s), want %d: %+v",
				c.category, len(got), len(c.want), got)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("%s[%d]: got %+v, want %+v", c.category, i, got[i], c.want[i])
			}
		}
	}
}
