package store_test

import (
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/jurisdiction"
)

// A resolved ward with no desk behind it is a report that goes nowhere: the
// platform names the place, matches the contract, and then cannot say where to
// send the complaint. Every Greater Mumbai ward must now answer for a road
// defect.
func TestLiveEveryWardHasARoadDefectDesk(t *testing.T) {
	ctx, _, db := liveDB(t)

	var missing []string
	for _, ward := range jurisdiction.KnownWards() {
		r, found, err := db.RoutingFor(ctx, "BMC", ward, "road_defect", "pothole")
		if err != nil {
			t.Fatalf("%s: RoutingFor: %v", ward, err)
		}
		if !found || strings.TrimSpace(r.Department) == "" {
			missing = append(missing, ward)
			continue
		}
		// Hard rule 2: a desk named with no document behind it is worse than
		// no desk, because a complaint goes to it.
		if strings.TrimSpace(r.SourceRef) == "" {
			t.Errorf("%s: desk %q has no source reference", ward, r.Department)
		}
		if strings.TrimSpace(r.Officer) == "" {
			t.Errorf("%s: no officer designation to address a complaint to", ward)
		}
		// A post, not a person. The source names individuals; storing one
		// would be personal data with a short shelf life.
		for _, honorific := range []string{"Shri ", "Smt ", "Shri.", "Smt.", "Mr ", "Ms "} {
			if strings.Contains(r.Officer, honorific) {
				t.Errorf("%s: officer designation looks like a person: %q", ward, r.Officer)
			}
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d ward(s) resolve but have no road defect desk: %s",
			len(missing), strings.Join(missing, " "))
	}
}

// Waste and water route everywhere too, each with the document behind it.
func TestLiveEveryWardRoutesWasteAndWater(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, category := range []string{"waste", "water_drainage"} {
		var missing []string
		for _, ward := range jurisdiction.KnownWards() {
			r, found, err := db.RoutingFor(ctx, "BMC", ward, category, "")
			if err != nil {
				t.Fatalf("%s/%s: %v", ward, category, err)
			}
			if !found || strings.TrimSpace(r.Department) == "" {
				missing = append(missing, ward)
				continue
			}
			if strings.TrimSpace(r.SourceRef) == "" {
				t.Errorf("%s/%s: %q has no source reference", ward, category, r.Department)
			}
		}
		if len(missing) > 0 {
			t.Errorf("%s: %d ward(s) have no desk: %s",
				category, len(missing), strings.Join(missing, " "))
		}
	}
}

// Refuse removal binds the whole waste category, so every ward should carry
// the clock with its citation.
func TestLiveTheWasteDeadlineReachesEveryWard(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		r, found, err := db.RoutingFor(ctx, "BMC", ward, "waste", "overflowing bin")
		if err != nil || !found {
			continue
		}
		if r.DeadlineHours != 24 || r.DeadlineCitation == "" {
			t.Errorf("%s: waste deadline is %d hours, citation %q",
				ward, r.DeadlineHours, r.DeadlineCitation)
		}
	}
}

// The pothole deadline binds potholes only, and that must hold in every ward
// now that every ward routes.
func TestLiveTheCourtDeadlineHoldsItsScopeInEveryWard(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		pot, found, err := db.RoutingFor(ctx, "BMC", ward, "road_defect", "pothole")
		if err != nil || !found {
			continue
		}
		if pot.DeadlineHours != 48 || pot.DeadlineCitation == "" {
			t.Errorf("%s: a pothole has %d hours and citation %q",
				ward, pot.DeadlineHours, pot.DeadlineCitation)
		}
		cover, found, err := db.RoutingFor(ctx, "BMC", ward, "road_defect", "missing manhole cover")
		if err != nil || !found {
			continue
		}
		if cover.DeadlineHours != 0 {
			t.Errorf("%s: the pothole order was attached to a manhole cover", ward)
		}
	}
}
