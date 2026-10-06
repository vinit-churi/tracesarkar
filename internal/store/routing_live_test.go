package store_test

import "testing"

// The desk and the clock, from the rows seeded by migrations 0015 and 0019.
// A verdict that names neither is an observation; this is what makes it an
// action.
func TestLiveRoutingFindsTheDeskAndTheClock(t *testing.T) {
	ctx, _, db := liveDB(t)

	t.Run("waste in the pilot ward", func(t *testing.T) {
		r, ok, err := db.RoutingFor(ctx, "BMC", "R/C", "waste")
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
		r, ok, err := db.RoutingFor(ctx, "BMC", "R/C", "water_drainage")
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

	t.Run("a ward with nothing recorded is not an error", func(t *testing.T) {
		_, ok, err := db.RoutingFor(ctx, "BMC", "R/N", "waste")
		if err != nil {
			t.Fatalf("RoutingFor: %v", err)
		}
		if ok {
			t.Error("R/N has no waste department; it must not be found")
		}
	})
}
