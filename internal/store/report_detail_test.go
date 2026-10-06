package store

import (
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/classify"
)

// A road contract covers a stretch of road. It does not make its contractor
// answerable for a sack of rubbish someone left there.
//
// The first real capture of a waste problem came back naming the road
// contractor at high confidence, because enrichment attributes every capture
// regardless of what is in the photograph. The join was right and the
// presentation was not: putting a company's name beside a garbage report says
// they are responsible for it, which is an assertion the platform does not get
// to make (hard rule 3), about a named party (hard rule 2).
func TestContractIsOnlyShownWhereTheContractGovernsIt(t *testing.T) {
	contract := func() ReportDetail {
		return ReportDetail{
			Ward: "R/C", Authority: "BMC",
			ContractorName: "M/s Example Infracon Pvt. Ltd",
			WorkCode:       "W-415", RoadName: "S.V. Road",
			DistanceM: 9.4, AttrConfidence: "high",
			ContractBasis: "The point lies 9.4 m from …",
		}
	}
	classified := func(category string) *classify.Decision {
		return &classify.Decision{
			Result:  classify.Result{Category: category, Subcategory: "x"},
			Outcome: classify.Accepted,
		}
	}

	tests := []struct {
		name     string
		detail   ReportDetail
		wantName bool
	}{
		{
			name: "a road defect shows the contractor",
			detail: func() ReportDetail {
				d := contract()
				d.Classification = classified("road_defect")
				return d
			}(),
			wantName: true,
		},
		{
			name: "waste does not",
			detail: func() ReportDetail {
				d := contract()
				d.Classification = classified("waste")
				return d
			}(),
		},
		{
			name: "water drainage does not",
			detail: func() ReportDetail {
				d := contract()
				d.Classification = classified("water_drainage")
				return d
			}(),
		},
		{
			// Enrichment runs before classification. Until we know what is in
			// the photograph we cannot know whose contract is relevant, and
			// naming someone on the strength of not knowing is the worst case.
			name:   "an unclassified capture does not",
			detail: contract(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.detail.ForCitizen()

			if tt.wantName {
				if got.ContractorName == "" {
					t.Error("a road defect must still show who holds the contract")
				}
				if got.ContractBasis == "" {
					t.Error("a named party must keep its basis (hard rule 2)")
				}
				return
			}

			// Nothing that names or points at the contractor survives.
			if got.ContractorName != "" || got.WorkCode != "" ||
				got.RoadName != "" || got.ContractBasis != "" ||
				got.AttrConfidence != "" || got.DistanceM != 0 {
				t.Errorf("contract details leaked to a %s report: %+v",
					tt.name, got)
			}
			// The things that are still true and still useful stay.
			if got.Ward != "R/C" || got.Authority != "BMC" {
				t.Error("suppressing the contract must not suppress the ward")
			}
		})
	}
}

// A photograph picked out of a gallery hours after it was taken carries a
// position that is wherever the device was when it was sent — which may be an
// office on the other side of the city. The report must say so, because
// everything downstream treats that position as where the problem is, and
// attribution will name whichever contract covers it.
func TestAStalePhotographSaysThePositionIsNotWhereItWasTaken(t *testing.T) {
	sent := time.Date(2026, 10, 6, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		takenAt   *time.Time
		wantStale bool
	}{
		{
			name:    "no timestamp in the image is not a claim either way",
			takenAt: nil,
		},
		{
			// Measured on the first real captures: three seconds apart.
			name:    "taken and sent together",
			takenAt: ptr(sent.Add(-3 * time.Second)),
		},
		{
			name:    "a few minutes is still the same walk",
			takenAt: ptr(sent.Add(-4 * time.Minute)),
		},
		{
			name:      "taken this morning, sent from the office",
			takenAt:   ptr(sent.Add(-5 * time.Hour)),
			wantStale: true,
		},
		{
			// A camera clock set wrong runs ahead. Unknown, not trustworthy.
			name:      "taken in the future",
			takenAt:   ptr(sent.Add(3 * time.Hour)),
			wantStale: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ReportDetail{CreatedAt: sent.Format(time.RFC3339), ExifTakenAt: tt.takenAt}
			if got := d.PositionIsStale(); got != tt.wantStale {
				t.Errorf("PositionIsStale() = %v, want %v", got, tt.wantStale)
			}
		})
	}
}

func ptr(t time.Time) *time.Time { return &t }

// A road defect normally shows its contractor. One whose position is stale
// does not: the contract was matched against wherever the device was when the
// photograph was sent, which is not where the problem is.
func TestAStalePositionWithholdsTheContractEvenForARoadDefect(t *testing.T) {
	sent := time.Date(2026, 10, 6, 11, 0, 0, 0, time.UTC)
	taken := sent.Add(-5 * time.Hour)

	d := ReportDetail{
		CreatedAt: sent.Format(time.RFC3339),
		Ward:      "R/C", Authority: "BMC",
		ExifTakenAt:    &taken,
		ContractorName: "M/s Example Infracon Pvt. Ltd",
		Classification: &classify.Decision{
			Result:  classify.Result{Category: "road_defect", Subcategory: "pothole"},
			Outcome: classify.Accepted,
		},
	}

	if got := d.ForCitizen(); got.ContractorName != "" {
		t.Errorf("named a contractor from a position five hours out of date: %q",
			got.ContractorName)
	}
	// The ward is still worth stating — it is wrong for a different reason and
	// the person can correct it.
	if d.ForCitizen().Ward != "R/C" {
		t.Error("the ward should survive")
	}
}
