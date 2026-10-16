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

// `structural` is one category and two authorities, which is why routing had
// to learn about subcategories. BMC's revised bridge circular of 28 June 2019
// gives the Chief Engineer (Bridges) the structural audit and the repairs of
// every MCGM bridge above six metres of span; the ward's own Assistant
// Engineer (Building & Factory) is the Designated Officer under MMC Act s.354
// for a ruinous or dangerous structure. A photograph of a cracked footover
// bridge and a photograph of a leaning building are the same category and
// belong to different desks.
func TestLiveAStructuralReportSplitsByWhatItIs(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		bridge, found, err := db.RoutingFor(ctx, "BMC", ward, "structural", "cracked bridge/FOB")
		if err != nil {
			t.Fatalf("%s: RoutingFor bridge: %v", ward, err)
		}
		if !found {
			t.Errorf("%s: a cracked bridge has no desk", ward)
		} else {
			if !strings.Contains(bridge.Department, "Bridges") {
				t.Errorf("%s: a cracked bridge went to %q", ward, bridge.Department)
			}
			if strings.TrimSpace(bridge.SourceRef) == "" {
				t.Errorf("%s: the bridge desk has no source reference", ward)
			}
		}

		building, found, err := db.RoutingFor(ctx, "BMC", ward, "structural", "distressed building")
		if err != nil {
			t.Fatalf("%s: RoutingFor building: %v", ward, err)
		}
		if !found {
			t.Errorf("%s: a distressed building has no desk", ward)
			continue
		}
		if strings.Contains(building.Department, "Bridges") {
			t.Errorf("%s: a building went to the bridge department", ward)
		}
		if !strings.Contains(building.Officer, "Building & Factory") {
			t.Errorf("%s: expected the ward Building & Factory post, got %q",
				ward, building.Officer)
		}
		// A subcategory with no row of its own falls back to the category's
		// desk rather than to nothing.
		wall, found, err := db.RoutingFor(ctx, "BMC", ward, "structural", "collapsed wall")
		if err != nil || !found {
			t.Errorf("%s: a collapsed wall has no desk (%v)", ward, err)
			continue
		}
		if wall.Department != building.Department {
			t.Errorf("%s: a collapsed wall went to %q, not the ward desk %q",
				ward, wall.Department, building.Department)
		}
	}
}

// Tree felling is the one thing inside `environment` that BMC owns. The Garden
// Department's own Section 4 handbook states the mission — "To Control illegal
// tree cutting" — and lists, under TREE AUTHORITY, taking action against
// illegal tree cutting, removal of dead and dangerous trees, and inspection of
// trees on complaint. The rest of the category is somebody else's: a mangrove
// is the Forest Department's, an effluent discharge the pollution board's. So
// the desk is recorded against the subcategory, and the category stays empty.
func TestLiveTreeFellingRoutesAndTheRestOfEnvironmentDoesNot(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		tree, found, err := db.RoutingFor(ctx, "BMC", ward, "environment", "tree felling")
		if err != nil {
			t.Fatalf("%s: RoutingFor: %v", ward, err)
		}
		if !found {
			t.Errorf("%s: tree felling has no desk", ward)
			continue
		}
		if !strings.Contains(tree.Officer, "Gardens") {
			t.Errorf("%s: expected the ward Gardens post, got %q", ward, tree.Officer)
		}
		if strings.TrimSpace(tree.SourceRef) == "" {
			t.Errorf("%s: the tree desk has no source reference", ward)
		}
		if strings.TrimSpace(tree.EscalatesTo) == "" {
			t.Errorf("%s: the tree desk has no office above it", ward)
		}
		// Nothing else in the category may inherit it.
		for _, sub := range []string{"mangrove destruction", "effluent discharge", "air pollution source"} {
			if _, found, err := db.RoutingFor(ctx, "BMC", ward, "environment", sub); err == nil && found {
				t.Errorf("%s: %q was routed to a desk BMC has not claimed", ward, sub)
			}
		}
	}
}

// The archive, not the parsed row, is the evidentiary record. A desk that
// cites only a live URL cites a path on a portal that rewrites its own paths
// and replaces documents in place — so in a year the row would say where it
// came from and nobody could produce it. Every routing row must carry an
// archive key as well as a source.
func TestLiveEveryDeskCitesAnArchivedDocument(t *testing.T) {
	ctx, pool, _ := liveDB(t)

	rows, err := pool.Query(ctx, `
		SELECT ward, category, COALESCE(subcategory, ''), department
		  FROM authority_departments
		 WHERE archive_key IS NULL OR archive_key = ''
		 ORDER BY category, ward`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()
	var unarchived int
	for rows.Next() {
		var ward, category, sub, dept string
		if err := rows.Scan(&ward, &category, &sub, &dept); err != nil {
			t.Fatalf("scan: %v", err)
		}
		unarchived++
		if unarchived <= 5 {
			t.Errorf("%s/%s%s: %q cites a live URL with nothing archived behind it",
				ward, category, map[bool]string{true: " (" + sub + ")"}[sub != ""], dept)
		}
	}
	if unarchived > 5 {
		t.Errorf("and %d more rows with no archive key", unarchived-5)
	}
}
