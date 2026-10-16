package store_test

import "testing"

// The desk and the clock, from the rows seeded by migrations 0015 and 0019.
// A verdict that names neither is an observation; this is what makes it an
// action.
func TestLiveRoutingFindsTheDeskAndTheClock(t *testing.T) {
	ctx, _, db := liveDB(t)

	t.Run("waste in the pilot ward", func(t *testing.T) {
		r, ok, err := db.RoutingFor(ctx, "BMC", "R/C", "waste", "")
		if err != nil || !ok {
			t.Fatalf("RoutingFor: %v (found=%v)", err, ok)
		}
		if r.Officer == "" || r.Office == "" {
			t.Errorf("no desk to address: %+v", r)
		}
		if r.DeadlineHours != 24 {
			t.Errorf("deadline = %d, want 24", r.DeadlineHours)
		}
		if r.DeadlineCitation == "" {
			t.Error("a deadline without its citation is not evidence")
		}
	})

	t.Run("water supply has a desk and no published deadline", func(t *testing.T) {
		r, ok, err := db.RoutingFor(ctx, "BMC", "R/C", "water_drainage", "")
		if err != nil || !ok {
			t.Fatalf("RoutingFor: %v (found=%v)", err, ok)
		}
		if r.Department == "" {
			t.Error("the department is known and should be returned")
		}
		// D098: the handbook publishes none, and borrowing one would invent an
		// obligation no document creates.
		if r.DeadlineHours != 0 {
			t.Errorf("invented a deadline of %d hours", r.DeadlineHours)
		}
	})

	// This case used to read "a ward with nothing recorded". No ward is in
	// that state any more — migrations 0029 to 0033 gave all twenty-four a
	// desk for four categories — so the honest version of the same assertion
	// is a category nobody has researched yet. Not found must stay an ordinary
	// answer: the platform says it does not know rather than guessing.
	t.Run("a category with nothing recorded is not an error", func(t *testing.T) {
		_, ok, err := db.RoutingFor(ctx, "BMC", "R/N", "structural", "")
		if err != nil {
			t.Fatalf("RoutingFor: %v", err)
		}
		if ok {
			t.Error("no structural desk is recorded anywhere; it must not be found")
		}
	})
}

// The 48-hour order binds potholes, not every road defect. The platform must
// state the clock for a pothole and state nothing for a missing manhole
// cover, because paragraph 70(ix) says "All potholes" and the share kit
// prints the quoted paragraph beside the photograph.
func TestLiveTheCourtDeadlineIsAttachedOnlyToPotholes(t *testing.T) {
	ctx, _, db := liveDB(t)

	pot, ok, err := db.RoutingFor(ctx, "BMC", "R/C", "road_defect", "pothole")
	if err != nil || !ok {
		t.Fatalf("RoutingFor pothole: %v (found=%v)", err, ok)
	}
	if pot.DeadlineHours != 48 {
		t.Errorf("a pothole has a deadline of %d hours, want 48", pot.DeadlineHours)
	}
	if pot.DeadlineCitation == "" {
		t.Error("a deadline without its citation is not evidence")
	}

	cover, ok, err := db.RoutingFor(ctx, "BMC", "R/C", "road_defect", "missing manhole cover")
	if err != nil || !ok {
		t.Fatalf("RoutingFor manhole cover: %v (found=%v)", err, ok)
	}
	if cover.DeadlineHours != 0 {
		t.Errorf("a pothole order was attached to a missing manhole cover: %d hours",
			cover.DeadlineHours)
	}
}
