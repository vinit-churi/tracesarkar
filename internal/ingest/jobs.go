package ingest

import (
	"fmt"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/sources"
	"github.com/vinit-churi/tracesarkar/internal/works"
)

// endpointsFor describes the endpoints of each source the snapshotter knows how
// to read. Adding a source here without an entry in the register does nothing:
// BuildJobs only returns jobs the register's policy allows.
func endpointsFor(sourceID string, now time.Time) []Endpoint {
	switch sourceID {
	case "bmc_roads_api":
		const base = "https://roads.mcgm.gov.in:3000/api/"
		return []Endpoint{
			{
				Name:  "publicdashboard",
				URL:   base + "publicdashboard/",
				Parse: works.ParseRoadsDashboard,
				Ext:   ".json",
			},
			{
				Name:  "roadlayer",
				URL:   base + "publicdashboard/getpublicdashboardwithstartedafter01oct2025roadlayer",
				Parse: works.ParseRoadsGeometry,
				Ext:   ".geojson",
			},
			{
				Name:  "ward",
				URL:   base + "ward/",
				Parse: works.ParseWardMaster,
				Ext:   ".json",
			},
		}
	case "bmc_swd_api":
		// The path carries the desilting season. From January the runner also
		// probes the next season, so a new season is captured on its first day.
		base := fmt.Sprintf("https://swd.mcgm.gov.in/swdwebapi%d/", seasonYear(now))
		return []Endpoint{
			{
				Name:  "progresscard",
				URL:   base + "report.svc/report/getprogresscard",
				Parse: works.ParseSWDProgressCard,
				Ext:   ".json",
			},
		}
	default:
		return nil
	}
}

// seasonYear returns the desilting season a date belongs to. The season is
// named for the calendar year it runs in.
func seasonYear(now time.Time) int { return now.UTC().Year() }

// BuildJobs turns the register into runnable jobs for the given tier. Sources
// the policy forbids are returned in the skipped map with the reason.
func BuildJobs(reg *sources.Register, tier string) ([]Job, map[string]string) {
	return buildJobsAt(reg, tier, time.Now().UTC())
}

func buildJobsAt(reg *sources.Register, tier string, now time.Time) ([]Job, map[string]string) {
	var jobs []Job
	skipped := map[string]string{}

	for _, src := range reg.All() {
		endpoints := endpointsFor(src.ID, now)
		if len(endpoints) == 0 {
			continue // no collector implemented for this source yet
		}
		if ok, reason := sources.MayRun(src, tier); !ok {
			skipped[src.ID] = reason
			continue
		}
		jobs = append(jobs, Job{
			SourceID:  src.ID,
			Blocklist: src.Blocklist,
			Endpoints: endpoints,
		})
	}

	// Sources with no collector but a policy objection are worth reporting too,
	// so an operator sees why nothing runs for them.
	for _, src := range reg.All() {
		if _, already := skipped[src.ID]; already {
			continue
		}
		if len(endpointsFor(src.ID, now)) > 0 {
			continue
		}
		if ok, reason := sources.MayRun(src, tier); !ok {
			skipped[src.ID] = reason
		}
	}
	return jobs, skipped
}
