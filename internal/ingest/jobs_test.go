package ingest

import (
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/sources"
)

func TestBuildJobsUsesTheRegisterBlocklist(t *testing.T) {
	reg, err := sources.Load("../../data/sources.yaml")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	jobs, skipped := BuildJobs(reg, sources.TierPersonal)

	var roads *Job
	for i := range jobs {
		if jobs[i].SourceID == "bmc_roads_api" {
			roads = &jobs[i]
		}
	}
	if roads == nil {
		t.Fatalf("bmc_roads_api should be schedulable at the personal tier; skipped: %v", skipped)
	}
	if len(roads.Blocklist) == 0 {
		t.Error("the job must carry the register's blocklist")
	}
	found := false
	for _, f := range roads.Blocklist {
		if strings.EqualFold(f, "contractorRepMobile") {
			found = true
		}
	}
	if !found {
		t.Errorf("contractorRepMobile must be stripped: %v", roads.Blocklist)
	}
	if len(roads.Endpoints) == 0 {
		t.Error("the job needs endpoints")
	}
	for _, ep := range roads.Endpoints {
		if ep.Parse == nil {
			t.Errorf("endpoint %s has no parser", ep.Name)
		}
		if !strings.HasPrefix(ep.URL, "https://") {
			t.Errorf("endpoint %s should be https: %s", ep.Name, ep.URL)
		}
	}
}

func TestBuildJobsSkipsSourcesThePolicyForbids(t *testing.T) {
	reg, err := sources.Load("../../data/sources.yaml")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	jobs, skipped := BuildJobs(reg, sources.TierPersonal)

	for _, j := range jobs {
		if j.SourceID == "mahatenders" {
			t.Fatal("mahatenders must never be scheduled (D027)")
		}
	}
	if reason, ok := skipped["mahatenders"]; !ok || reason == "" {
		t.Errorf("a skipped source must be reported with a reason: %v", skipped)
	}
}

func TestBuildJobsAtPublicTierNeedsReviewedTerms(t *testing.T) {
	reg, err := sources.Load("../../data/sources.yaml")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	jobs, skipped := BuildJobs(reg, sources.TierPublic)

	for _, j := range jobs {
		if j.SourceID == "bmc_roads_api" {
			t.Error("terms for the roads API are unreviewed; it must not run at the public tier")
		}
	}
	if _, ok := skipped["bmc_roads_api"]; !ok {
		t.Error("the skip should be explained")
	}
}
