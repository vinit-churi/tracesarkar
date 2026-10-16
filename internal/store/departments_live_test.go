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
// the clock with its citation. Twelve hours, from BMC's own Citizen Charter of
// 7 December 2020 — which supersedes the 24 hours the 2006 R/Central handbook
// gave, being both newer and citywide.
func TestLiveTheWasteDeadlineReachesEveryWard(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		r, found, err := db.RoutingFor(ctx, "BMC", ward, "waste", "overflowing bin")
		if err != nil || !found {
			continue
		}
		if r.DeadlineHours != 12 || r.DeadlineCitation == "" {
			t.Errorf("%s: waste deadline is %d hours, citation %q",
				ward, r.DeadlineHours, r.DeadlineCitation)
		}
	}
}

// The same charter gives a dead animal its own clock — 48 hours, four times
// the twelve the rest of the category gets — so the narrower constant must win
// for that subcategory and leave the others alone. This is D138's rule applied
// a second time: a clock is shown with the scope the document gives it.
func TestLiveADeadAnimalCarriesItsOwnClock(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		dead, found, err := db.RoutingFor(ctx, "BMC", ward, "waste", "dead animal")
		if err != nil || !found {
			continue
		}
		if dead.DeadlineHours != 48 || dead.DeadlineCitation == "" {
			t.Errorf("%s: a dead animal has %d hours and citation %q",
				ward, dead.DeadlineHours, dead.DeadlineCitation)
		}
		bin, found, err := db.RoutingFor(ctx, "BMC", ward, "waste", "overflowing bin")
		if err != nil || !found {
			continue
		}
		if bin.DeadlineHours != 12 {
			t.Errorf("%s: the dead-animal clock leaked onto a bin: %d hours",
				ward, bin.DeadlineHours)
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

// A streetlight that is out is `street_furniture` in the taxonomy, and the
// taxonomy's own rule is that a category exists only if something downstream
// can route it. Until these rows existed a citizen could photograph a dead
// streetlight, have it classified correctly, and be told nothing about where
// it goes.
func TestLiveEveryWardRoutesStreetFurniture(t *testing.T) {
	ctx, _, db := liveDB(t)

	var missing []string
	for _, ward := range jurisdiction.KnownWards() {
		r, found, err := db.RoutingFor(ctx, "BMC", ward, "street_furniture", "streetlight out")
		if err != nil {
			t.Fatalf("%s: RoutingFor: %v", ward, err)
		}
		if !found || strings.TrimSpace(r.Department) == "" {
			missing = append(missing, ward)
			continue
		}
		if strings.TrimSpace(r.SourceRef) == "" {
			t.Errorf("%s: desk %q has no source reference", ward, r.Department)
		}
		if strings.TrimSpace(r.Officer) == "" {
			t.Errorf("%s: no officer designation to address a complaint to", ward)
		}
		if strings.TrimSpace(r.EscalatesTo) == "" {
			t.Errorf("%s: street furniture desk has no office above it", ward)
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d ward(s) resolve but have no street furniture desk: %s",
			len(missing), strings.Join(missing, " "))
	}
}

// The distribution licensee lights the lamp; it does not own the complaint.
// Adani Electricity, Tata Power and BEST supply the power and publish the LUX
// standard the ward works to — the ward's own Assistant Engineer (Maintenance
// & Repair) is the office a resident writes to. Naming a licensee here would
// send a citizen to a company that cannot act on a municipal asset.
func TestLiveTheStreetLightDeskIsMunicipalNotTheLicensee(t *testing.T) {
	ctx, _, db := liveDB(t)

	licensees := []string{"Adani", "Tata Power", "Reliance", "BEST"}
	for _, ward := range jurisdiction.KnownWards() {
		r, found, err := db.RoutingFor(ctx, "BMC", ward, "street_furniture", "streetlight out")
		if err != nil || !found {
			continue
		}
		for _, l := range licensees {
			if strings.Contains(r.Department, l) || strings.Contains(r.Officer, l) {
				t.Errorf("%s: street light routed to a licensee: %q / %q",
					ward, r.Department, r.Officer)
			}
		}
		if !strings.Contains(r.Officer, "Maintenance") {
			t.Errorf("%s: expected the ward maintenance post, got %q", ward, r.Officer)
		}
	}
}
