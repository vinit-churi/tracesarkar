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
		// Nothing else in the category may inherit the Tree Authority, and the
		// three subcategories nobody has researched must still answer "not
		// known" rather than borrowing a desk. Effluent and air are no longer
		// among them: they belong to the pollution control board, which is not
		// BMC at all (migration 0043).
		for _, sub := range []string{"mangrove destruction", "creek dumping",
			"encroachment on water body"} {
			if _, found, err := db.RoutingFor(ctx, "BMC", ward, "environment", sub); err == nil && found {
				t.Errorf("%s: %q was routed to a desk nobody has researched", ward, sub)
			}
		}
		for _, sub := range []string{"effluent discharge", "air pollution source"} {
			r, found, err := db.RoutingFor(ctx, "BMC", ward, "environment", sub)
			if err != nil || !found {
				continue
			}
			if r.Authority == "BMC" {
				t.Errorf("%s: %q was claimed for BMC", ward, sub)
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

// `water_drainage` has held one desk for everything since migration 0019, and
// that desk is Water Works — the water *supply* department. It is right for a
// burst main and wrong for everything that drains: BMC's own ward handbooks
// give the ward Assistant Engineer (Maintenance & Repair) "desilting of minor
// nallas and road side SWD" and "to minimise flooding spots during monsoon",
// while the Chief Engineer (Storm Water Drains) handbook claims the SWD
// network and the desilting of **major** nallas. A photograph of a flooded
// lane arriving at the water-supply counter is the misroute this splits.
func TestLiveDrainageDoesNotGoToTheWaterSupplyDesk(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		for _, sub := range []string{"waterlogging", "blocked drain"} {
			r, found, err := db.RoutingFor(ctx, "BMC", ward, "water_drainage", sub)
			if err != nil {
				t.Fatalf("%s/%s: %v", ward, sub, err)
			}
			if !found {
				t.Errorf("%s: %q has no desk", ward, sub)
				continue
			}
			if strings.Contains(r.Department, "Water Works") {
				t.Errorf("%s: %q went to the water supply desk %q", ward, sub, r.Department)
			}
			if !strings.Contains(r.Officer, "Maintenance & Repair") {
				t.Errorf("%s: %q expected the ward maintenance post, got %q",
					ward, sub, r.Officer)
			}
		}

		nallah, found, err := db.RoutingFor(ctx, "BMC", ward, "water_drainage", "silted nallah")
		if err != nil || !found {
			t.Errorf("%s: a silted nallah has no desk (%v)", ward, err)
		} else if !strings.Contains(nallah.Department, "Storm Water Drains") {
			t.Errorf("%s: a silted nallah went to %q", ward, nallah.Department)
		}

		// What the category-wide desk is actually right for must not move.
		for _, sub := range []string{"burst pipeline", "leakage"} {
			r, found, err := db.RoutingFor(ctx, "BMC", ward, "water_drainage", sub)
			if err != nil || !found {
				t.Errorf("%s: %q has no desk (%v)", ward, sub, err)
				continue
			}
			if !strings.Contains(r.Department, "Water Works") {
				t.Errorf("%s: %q should still be the water supply desk, got %q",
					ward, sub, r.Department)
			}
		}
	}
}

// The bridge circular splits at exactly six metres of span, and a resident
// cannot measure span from a photograph. The platform answers with the desk
// that is right for every footover bridge and says, on the card, which side of
// six metres each desk owns — so a person standing at a small culvert can read
// that the other one is theirs rather than being quietly misrouted.
func TestLiveBothBridgeDesksStateTheirSpan(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		p, err := db.WardProfileFor(ctx, "BMC", ward)
		if err != nil {
			t.Fatalf("%s: WardProfileFor: %v", ward, err)
		}
		var bridges, swd int
		for _, d := range p.Desks {
			if d.Subcategory != "cracked bridge/FOB" {
				continue
			}
			if !strings.Contains(d.Department, "6 m") {
				t.Errorf("%s: bridge desk %q does not say which spans it owns",
					ward, d.Department)
			}
			switch {
			case strings.Contains(d.Department, "Bridges"):
				bridges++
			case strings.Contains(d.Department, "Storm Water Drains"):
				swd++
			}
		}
		if bridges != 1 || swd != 1 {
			t.Errorf("%s: %d bridge desk(s) and %d storm-water desk(s), want 1 of each",
				ward, bridges, swd)
		}

		// The one answer a single-desk caller gets stays the Bridge
		// department: it is right for every footover bridge, flyover and
		// skywalk in the city.
		r, found, err := db.RoutingFor(ctx, "BMC", ward, "structural", "cracked bridge/FOB")
		if err != nil || !found {
			t.Errorf("%s: no bridge desk (%v)", ward, err)
		} else if !strings.Contains(r.Department, "Chief Engineer (Bridges)") {
			t.Errorf("%s: the single answer became %q", ward, r.Department)
		}
	}
}

// Sewage overflow was the last thing in `water_drainage` still falling back to
// the water supply desk. The Sewerage Operation department's own Section 4
// handbook says it "operates and maintains the Sewerage Network, Sewage
// Pumping Stations, Sewage Treatment Plants" and attends "day-to-day
// complaints related to sewer lines", which is what a photograph of a sewer
// backing up into a street is about.
func TestLiveSewageOverflowReachesTheSewerageDepartment(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		r, found, err := db.RoutingFor(ctx, "BMC", ward, "water_drainage", "sewage overflow")
		if err != nil {
			t.Fatalf("%s: %v", ward, err)
		}
		if !found {
			t.Errorf("%s: sewage overflow has no desk", ward)
			continue
		}
		if !strings.Contains(r.Department, "Sewerage Operation") {
			t.Errorf("%s: sewage overflow went to %q", ward, r.Department)
		}
		if strings.Contains(r.Department, "Water Works") {
			t.Errorf("%s: sewage overflow is still at the water supply desk", ward)
		}
		if r.OfficeEmail == "" {
			t.Errorf("%s: no mailbox to write to", ward)
		}
	}
}

// Not every desk belongs to the corporation. An effluent discharge is the
// Maharashtra Pollution Control Board's — it is constituted under section 4 of
// the Water (Prevention and Control of Pollution) Act 1974 and is the state
// board under section 5 of the Air Act 1981 — and a platform that can only
// name BMC would have to answer "not known" for something it knows.
func TestLiveAStateBoardCanOwnASubcategory(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, ward := range jurisdiction.KnownWards() {
		for _, sub := range []string{"effluent discharge", "air pollution source"} {
			r, found, err := db.RoutingFor(ctx, "BMC", ward, "environment", sub)
			if err != nil {
				t.Fatalf("%s/%s: %v", ward, sub, err)
			}
			if !found {
				t.Errorf("%s: %q has no desk", ward, sub)
				continue
			}
			if r.Authority != "MPCB" {
				t.Errorf("%s: %q belongs to %q, want MPCB", ward, sub, r.Authority)
			}
			if r.OfficeEmail == "" {
				t.Errorf("%s: %q has no mailbox", ward, sub)
			}
			if strings.TrimSpace(r.SourceRef) == "" {
				t.Errorf("%s: %q has no source", ward, sub)
			}
		}
	}

	// A desk belonging to another body must not leak into the corporation's
	// own categories. Asking BMC for a pothole still gets BMC.
	r, found, err := db.RoutingFor(ctx, "BMC", "R/C", "road_defect", "pothole")
	if err != nil || !found {
		t.Fatalf("RoutingFor pothole: %v (found=%v)", err, found)
	}
	if r.Authority != "BMC" {
		t.Errorf("a pothole is now %q's", r.Authority)
	}
}
