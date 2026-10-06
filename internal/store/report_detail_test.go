package store

import (
	"testing"

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
