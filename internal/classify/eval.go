package classify

import (
	"context"
	"fmt"
	"strings"
)

// EvalCase is one labelled photograph: what it is, as a person judged it.
type EvalCase struct {
	ID              string
	Image           []byte
	WantCategory    string
	WantSubcategory string
	WantHazard      bool
	// Note carries anything the labeller wrote — "taken at dusk", "under
	// water" — which is usually what explains a miss.
	Note string
}

// Miss is one case the classifier got wrong, kept so the failures can be read
// rather than summarised. An accuracy number alone cannot be acted on.
type Miss struct {
	ID        string `json:"id"`
	Field     string `json:"field"`
	Want      string `json:"want"`
	Got       string `json:"got"`
	Rationale string `json:"rationale"`
	Note      string `json:"note,omitempty"`
}

// Report is what a run of the eval set produced.
type Report struct {
	Model        string `json:"model"`
	Prompt       string `json:"prompt_version"`
	Cases        int    `json:"cases"`
	CategoryHits int    `json:"category_hits"`
	SubHits      int    `json:"subcategory_hits"`
	// HazardsMissed counts photographs a person marked dangerous that the
	// platform would have routed as ordinary. Counted separately because
	// averaging it into one accuracy figure hides the error that can hurt
	// somebody.
	HazardsMissed int `json:"hazards_missed"`
	// Overrides counts the times the known-hazard rule rescued the model.
	Overrides int    `json:"overrides"`
	Misses    []Miss `json:"misses"`
}

// CategoryAccuracy is the share routed to the right department.
func (r Report) CategoryAccuracy() float64 {
	if r.Cases == 0 {
		return 0
	}
	return float64(r.CategoryHits) / float64(r.Cases)
}

// SubcategoryAccuracy is the share described correctly.
func (r Report) SubcategoryAccuracy() float64 {
	if r.Cases == 0 {
		return 0
	}
	return float64(r.SubHits) / float64(r.Cases)
}

// RunEval scores a classifier over labelled photographs.
//
// It scores the platform's *decision*, not the raw model output — the
// guardrails are part of what a citizen experiences, so a hazard the override
// rescued is not a missed hazard, and a category the platform would refuse is
// not a hit.
func RunEval(ctx context.Context, c Classifier, cases []EvalCase, covered Coverage) (Report, error) {
	if len(cases) == 0 {
		// "100% over zero photographs" is the most dangerous number a report
		// can carry.
		return Report{}, fmt.Errorf("the eval set is empty; there is nothing to measure")
	}

	report := Report{Cases: len(cases)}

	for _, tc := range cases {
		raw, err := c.Classify(ctx, tc.Image, "")
		if err != nil {
			// A provider outage scored as 0% accuracy sends someone chasing a
			// prompt problem that does not exist.
			return Report{}, fmt.Errorf("case %s: %w", tc.ID, err)
		}

		d := Apply(raw, covered)
		if d.Overridden {
			report.Overrides++
		}

		switch {
		case strings.EqualFold(d.Category, tc.WantCategory):
			report.CategoryHits++
		default:
			report.Misses = append(report.Misses, Miss{
				ID: tc.ID, Field: "category",
				Want: tc.WantCategory, Got: d.Category,
				Rationale: d.Rationale, Note: tc.Note,
			})
		}

		if strings.EqualFold(d.Subcategory, tc.WantSubcategory) {
			report.SubHits++
		} else if strings.EqualFold(d.Category, tc.WantCategory) {
			// Only worth reporting separately when the department was right;
			// otherwise the category miss already explains it.
			report.Misses = append(report.Misses, Miss{
				ID: tc.ID, Field: "subcategory",
				Want: tc.WantSubcategory, Got: d.Subcategory,
				Rationale: d.Rationale, Note: tc.Note,
			})
		}

		if tc.WantHazard && !d.HazardToLife {
			report.HazardsMissed++
			report.Misses = append(report.Misses, Miss{
				ID: tc.ID, Field: "hazard_to_life",
				Want: "true", Got: "false",
				Rationale: d.Rationale, Note: tc.Note,
			})
		}
	}

	return report, nil
}
