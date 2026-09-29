package sources_test

import (
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/sources"
)

// The blocklist for BMC's roads API once named contractorRepName and
// contractorRepMobile — fields the API does not return. It therefore stripped
// nothing, and 2,387 personal mobile numbers were stored before anyone looked.
// The register has to name the fields the source actually returns.
func TestRoadsAPIBlocklistNamesTheFieldsThatCarryPersonalData(t *testing.T) {
	reg, err := sources.Load("../../data/sources.yaml")
	if err != nil {
		t.Fatalf("load register: %v", err)
	}

	src, ok := reg.Get("bmc_roads_api")
	if !ok {
		t.Fatal("bmc_roads_api is not in the register")
	}
	blocklist := src.Blocklist
	if len(blocklist) == 0 {
		t.Fatal("bmc_roads_api has no blocklist")
	}
	joined := strings.ToLower(strings.Join(blocklist, ","))

	// Verified present in the live response, 29 September 2026.
	for _, field := range []string{"qmarepname", "qmarepmobile"} {
		if !strings.Contains(joined, field) {
			t.Errorf("blocklist does not strip %q: %v", field, blocklist)
		}
	}
}

// BMC's prabhag layer carries the personal names of junior engineers and
// councillors, on a layer dated 2016. Probed 29 September 2026: populated on
// 7 of 227 prabhags, so it is useless as a mapping and personal data anyway.
// Blocked before an ingester exists, rather than after 2,387 rows were stored.
func TestProhibitedPersonFieldsOnTheProbedLayersAreBlocked(t *testing.T) {
	reg, err := sources.Load("../../data/sources.yaml")
	if err != nil {
		t.Fatalf("load register: %v", err)
	}
	src, ok := reg.Get("bmc_roads_api")
	if !ok {
		t.Fatal("bmc_roads_api is not in the register")
	}
	joined := strings.ToLower(strings.Join(src.Blocklist, ","))

	for _, field := range []string{"councillor", "jr_engg"} {
		if !strings.Contains(joined, field) {
			t.Errorf("blocklist does not strip %q: %v", field, src.Blocklist)
		}
	}
}
