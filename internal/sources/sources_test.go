package sources

import (
	"os"
	"path/filepath"
	"testing"
)

func writeRegister(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sources.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadReadsRegisterFields(t *testing.T) {
	path := writeRegister(t, `
sources:
  - id: bmc_roads_api
    name: BMC roads dashboard API
    url: https://roads.mcgm.gov.in:3000/api/
    category: works
    acquisition: api
    cadence: daily
    licence: unverified
    terms_reviewed_on: null
    robots_ok: null
    status: candidate
    blocklist: contractorRepName, contractorRepMobile
    notes: >
      Two thousand works.
`)

	reg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	src, ok := reg.Get("bmc_roads_api")
	if !ok {
		t.Fatal("source not found")
	}
	if src.Name != "BMC roads dashboard API" || src.Category != "works" || src.Acquisition != "api" {
		t.Errorf("unexpected fields: %+v", src)
	}
	if src.Status != "candidate" {
		t.Errorf("status: got %q", src.Status)
	}
	if got := src.Blocklist; len(got) != 2 || got[0] != "contractorRepName" || got[1] != "contractorRepMobile" {
		t.Errorf("blocklist: got %v", got)
	}
}

func TestLoadAcceptsBlocklistAsList(t *testing.T) {
	path := writeRegister(t, `
sources:
  - id: s
    name: n
    category: works
    acquisition: api
    status: candidate
    blocklist:
      - VehicleNo
      - SlipNo
`)
	reg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	src, _ := reg.Get("s")
	if len(src.Blocklist) != 2 || src.Blocklist[1] != "SlipNo" {
		t.Errorf("blocklist: got %v", src.Blocklist)
	}
}

func TestLoadRejectsDuplicateIDs(t *testing.T) {
	path := writeRegister(t, `
sources:
  - id: dup
    name: a
    category: works
    acquisition: api
    status: candidate
  - id: dup
    name: b
    category: works
    acquisition: api
    status: candidate
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected an error for duplicate ids")
	}
}

func TestMayRunPolicy(t *testing.T) {
	reviewed := "2026-09-01"
	no := false
	tests := []struct {
		name   string
		src    Source
		tier   string
		wantOK bool
	}{
		{
			name:   "personal tier may run a candidate whose terms are unreviewed",
			src:    Source{ID: "a", Status: "candidate", Acquisition: "api"},
			tier:   TierPersonal,
			wantOK: true,
		},
		{
			name:   "public tier may not run until terms are reviewed",
			src:    Source{ID: "a", Status: "candidate", Acquisition: "api"},
			tier:   TierPublic,
			wantOK: false,
		},
		{
			name:   "public tier may run a reviewed source",
			src:    Source{ID: "a", Status: "active", Acquisition: "api", TermsReviewedOn: &reviewed},
			tier:   TierPublic,
			wantOK: true,
		},
		{
			name:   "blocked never runs, at any tier",
			src:    Source{ID: "m", Status: "blocked", Acquisition: "manual", RobotsOK: &no},
			tier:   TierPersonal,
			wantOK: false,
		},
		{
			name:   "robots.txt disallow never runs",
			src:    Source{ID: "a", Status: "candidate", Acquisition: "api", RobotsOK: &no},
			tier:   TierPersonal,
			wantOK: false,
		},
		{
			name:   "manual acquisition is not scheduled",
			src:    Source{ID: "a", Status: "candidate", Acquisition: "manual"},
			tier:   TierPersonal,
			wantOK: false,
		},
		{
			name:   "retired never runs",
			src:    Source{ID: "a", Status: "retired", Acquisition: "api"},
			tier:   TierPersonal,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, reason := MayRun(tt.src, tt.tier)
			if ok != tt.wantOK {
				t.Fatalf("got ok=%v (%s), want %v", ok, reason, tt.wantOK)
			}
			if !ok && reason == "" {
				t.Error("a refusal must explain itself")
			}
		})
	}
}

func TestMayPublishRequiresReviewedTerms(t *testing.T) {
	reviewed := "2026-09-01"
	if ok, _ := MayPublish(Source{Status: "candidate"}); ok {
		t.Error("unreviewed terms must not reach a public surface")
	}
	if ok, reason := MayPublish(Source{Status: "active", TermsReviewedOn: &reviewed}); !ok {
		t.Errorf("reviewed source should be publishable: %s", reason)
	}
}

// The register that ships in the repository must itself be valid.
func TestRepositoryRegisterLoads(t *testing.T) {
	reg, err := Load("../../data/sources.yaml")
	if err != nil {
		t.Fatalf("load repository register: %v", err)
	}
	if len(reg.All()) < 10 {
		t.Fatalf("expected the full register, got %d sources", len(reg.All()))
	}
	roads, ok := reg.Get("bmc_roads_api")
	if !ok {
		t.Fatal("bmc_roads_api missing from the register")
	}
	if len(roads.Blocklist) == 0 {
		t.Error("bmc_roads_api must carry a personal-data blocklist")
	}
	if maha, ok := reg.Get("mahatenders"); !ok {
		t.Error("mahatenders missing")
	} else if ok, _ := MayRun(maha, TierPersonal); ok {
		t.Error("mahatenders must never be schedulable (D027)")
	}
}
