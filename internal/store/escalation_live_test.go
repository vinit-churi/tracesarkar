package store_test

import (
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/jurisdiction"
)

// Every one of the twenty-four wards must have somewhere to send the second
// letter. A ward with a desk and no escalation contact is the state this
// migration exists to end.
func TestLiveEveryWardHasAnEscalationContact(t *testing.T) {
	ctx, _, db := liveDB(t)

	wards := []string{
		"A", "B", "C", "D", "E", "F/N", "F/S", "G/N", "G/S", "H/E", "H/W",
		"K/E", "K/W", "L", "M/E", "M/W", "N", "P/N", "P/S", "R/C", "R/N",
		"R/S", "S", "T",
	}
	for _, w := range wards {
		c, err := db.EscalationFor(ctx, "BMC", w)
		if err != nil {
			t.Errorf("EscalationFor(%s): %v", w, err)
			continue
		}
		if c.Ward.Designation == "" {
			t.Errorf("ward %s has no Assistant Commissioner row", w)
		}
		if c.Ward.OfficeEmail == "" {
			t.Errorf("ward %s has no role mailbox to write to", w)
		}
		if c.Zone.Designation == "" {
			t.Errorf("ward %s resolves to no zone", w)
		}
	}
}

// A mailbox is served only as published. The source prints one of the seven
// zone addresses without the dot its siblings carry, and normalising it would
// invent an address that may not exist.
func TestLiveZoneMailboxIsServedVerbatim(t *testing.T) {
	ctx, _, db := liveDB(t)

	c, err := db.EscalationFor(ctx, "BMC", "K/W")
	if err != nil {
		t.Fatalf("EscalationFor: %v", err)
	}
	if c.Zone.OfficeEmail != "dmcz4@mcgm.gov.in" {
		t.Errorf("zone 4 mailbox = %q, want it exactly as published", c.Zone.OfficeEmail)
	}
}

// Provenance is not optional. A contact displayed without a source and a
// retrieval date is a fact about a named office the platform cannot stand up.
func TestLiveEscalationContactsCarryProvenance(t *testing.T) {
	ctx, _, db := liveDB(t)

	c, err := db.EscalationFor(ctx, "BMC", "R/C")
	if err != nil {
		t.Fatalf("EscalationFor: %v", err)
	}
	for _, g := range []struct {
		what string
		ref  string
		zero bool
	}{
		{"ward", c.Ward.SourceRef, c.Ward.RetrievedAt.IsZero()},
		{"zone", c.Zone.SourceRef, c.Zone.RetrievedAt.IsZero()},
	} {
		if g.ref == "" || g.zero {
			t.Errorf("%s contact came back with no provenance", g.what)
		}
	}
}

// K/East and P/North were seeded with a mailbox and no address, because the
// consolidated PIO list gave those two wards two-column pages headed with ward
// names BMC does not use officially. A second document closed the gap: Manual
// 16 of the Garden Department's handbook prints the Assistant Commissioner's
// address for every ward in its own column (migration 0037). Every ward now
// has both, and a refusal that was right at the time is a refusal that could
// be lifted by better evidence rather than by a guess.
func TestLiveEveryWardHasBothAMailboxAndAnAddress(t *testing.T) {
	ctx, _, db := liveDB(t)

	for _, w := range jurisdiction.KnownWards() {
		c, err := db.EscalationFor(ctx, "BMC", w)
		if err != nil {
			t.Fatalf("EscalationFor(%s): %v", w, err)
		}
		if c.Ward.OfficeEmail == "" {
			t.Errorf("ward %s has no mailbox", w)
		}
		if c.Ward.OfficeAddress == "" {
			t.Errorf("ward %s has no office address", w)
		}
		if c.Ward.SourceRef == "" {
			t.Errorf("ward %s has an address with no document behind it", w)
		}
	}
}

// The register records where a department's handbook is published. It must
// never be read as a claim about who answers for a complaint.
func TestLiveDepartmentRegisterIsSourced(t *testing.T) {
	ctx, _, db := liveDB(t)

	ds, err := db.Departments(ctx)
	if err != nil {
		t.Fatalf("Departments: %v", err)
	}
	if len(ds) < 80 {
		t.Fatalf("register holds %d departments; migration 0032 seeds 86", len(ds))
	}
	var bridges bool
	for _, d := range ds {
		if d.SourceRef == "" {
			t.Errorf("%s has no source reference", d.NameEN)
		}
		if d.RTIURL == "" && d.DocsURL == "" {
			t.Errorf("%s has neither handbook nor docs link", d.NameEN)
		}
		if strings.Contains(d.NameEN, "Bridges") {
			bridges = true
			if d.RTIURL == "" {
				t.Error("Bridges has no handbook link; it is one of the three this unblocks")
			}
		}
	}
	if !bridges {
		t.Error("no Bridges department in the register")
	}
}

// An escalation line is printed on the public ward page under a heading that
// carries the ward's own code, and the one thing it has to be is unambiguous
// about which office to write to. Two spellings of one ward read as two
// wards: R/C carried both "R/C Ward" and "R/Central Ward" until migration
// 0045, and a resident reading the second under a heading saying the first
// had no way to tell they were the same office.
//
// The check is "names its own ward" rather than "names no other", because a
// row may legitimately carry another ward's office *address* — the Dy.
// Hydraulic Engineer for the Western Suburbs sits in the K/West ward office.
func TestLiveAnEscalationNamesItsOwnWardAndOnlyOneSpellingOfIt(t *testing.T) {
	ctx, pool, _ := liveDB(t)

	rows, err := pool.Query(ctx, `
		SELECT ward, category, COALESCE(subcategory,''), escalates_to
		  FROM authority_departments
		 WHERE escalates_to ILIKE '%ward%'
		   AND escalates_to NOT LIKE '%' || ward || ' Ward%'
		   AND escalates_to NOT LIKE '%' || ward || ' ward%'
		 ORDER BY ward, category`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var bad int
	for rows.Next() {
		var ward, category, subcategory, escalation string
		if err := rows.Scan(&ward, &category, &subcategory, &escalation); err != nil {
			t.Fatalf("scan: %v", err)
		}
		bad++
		t.Errorf("%s/%s%s escalates to a line that does not name %s Ward:\n  %s",
			ward, category, map[bool]string{true: " (" + subcategory + ")"}[subcategory != ""],
			ward, escalation)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	if bad == 0 {
		t.Log("every escalation that names a ward names its own")
	}
}
