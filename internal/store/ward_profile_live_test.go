package store_test

import "testing"

// Everything the platform can say about a ward before anybody reports
// anything. It is most of what the project knows and, until now, none of it
// was visible without signing in.
func TestLiveWardProfileIsPublishableOnItsOwn(t *testing.T) {
	ctx, _, db := liveDB(t)

	p, err := db.WardProfileFor(ctx, "BMC", "R/C")
	if err != nil {
		t.Fatalf("WardProfileFor: %v", err)
	}

	if len(p.Desks) < 3 {
		t.Fatalf("expected the seeded desks, got %d", len(p.Desks))
	}
	var withDeadline, withContact int
	for _, d := range p.Desks {
		// Hard rule 2: a published fact about a named party carries its
		// source and when it was read.
		if d.SourceRef == "" || d.RetrievedAt.IsZero() {
			t.Errorf("%s has no source or no retrieval time", d.Category)
		}
		if d.DeadlineHours > 0 {
			withDeadline++
			if d.DeadlineCitation == "" {
				t.Errorf("%s shows a deadline with no citation", d.Category)
			}
		}
		if d.OfficePhone != "" || d.OfficeEmail != "" {
			withContact++
		}
	}
	if withDeadline == 0 {
		t.Error("no desk carries a deadline; the page would state no obligation")
	}
	if withContact == 0 {
		t.Error("no desk carries a contact; the page would name a desk nobody can reach")
	}

	// The coverage figure is the honest half of the page.
	if p.Segments == 0 {
		t.Error("no road segments resolved inside the ward boundary")
	}
	t.Logf("R/C: %d desks, %d road segments, %.0f km with contract geometry",
		len(p.Desks), p.Segments, p.Metres/1000)
}

// A ward with nothing recorded is an ordinary answer, not an error. Most wards
// are in that state and the page has to say so rather than fail.
func TestLiveAWardWithNoDesksStillProfiles(t *testing.T) {
	ctx, _, db := liveDB(t)

	p, err := db.WardProfileFor(ctx, "BMC", "R/N")
	if err != nil {
		t.Fatalf("WardProfileFor: %v", err)
	}
	if len(p.Desks) != 0 {
		t.Errorf("R/N has no departments recorded; got %d", len(p.Desks))
	}
	// The boundary exists, so the road count still means something.
	if p.Segments == 0 {
		t.Error("R/N has a boundary and should still resolve road segments")
	}
}
