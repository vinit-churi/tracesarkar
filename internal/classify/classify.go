// Package classify turns a photograph into a routable category.
//
// The model's answer is never the final answer. Everything here exists because
// a classification decides which department receives a complaint, and a
// confident wrong answer sends a citizen's report to a desk that cannot act on
// it. So the model proposes; these rules dispose.
package classify

import (
	"fmt"
	"strings"
)

// Outcome is what the platform does with a classification.
type Outcome string

const (
	Accepted          Outcome = "accepted"           // route it
	NeedsConfirmation Outcome = "needs_confirmation" // ask the citizen which
	NeedsRetake       Outcome = "needs_retake"       // the photograph cannot be read
	OutOfScope        Outcome = "out_of_scope"       // not public infrastructure
	// NotYetCovered is a real civic problem the platform recognises and has
	// nowhere to send. It exists because the alternative — accepting it —
	// classifies a report and then drops it, and the citizen never hears
	// anything again.
	NotYetCovered Outcome = "not_yet_covered"
)

// Coverage is the set of categories the platform can actually route, which is
// the set that has a row in authority_departments. It is passed in rather than
// hard-coded here so that adding a department makes its category routable
// without anyone remembering to edit this package.
type Coverage map[string]bool

// CoverageFor builds a coverage set.
func CoverageFor(categories ...string) Coverage {
	c := make(Coverage, len(categories))
	for _, k := range categories {
		c[strings.ToLower(strings.TrimSpace(k))] = true
	}
	return c
}

// Covers reports whether a category can be routed.
func (c Coverage) Covers(category string) bool {
	return c[strings.ToLower(strings.TrimSpace(category))]
}

// Result is what the model returns, before any rule is applied. The shape is
// guaranteed by structured output; the *vocabulary* is not, which is why the
// taxonomy is checked below.
type Result struct {
	Category      string   `json:"category"`
	Subcategory   string   `json:"subcategory"`
	Severity      string   `json:"severity"`
	HazardToLife  bool     `json:"hazard_to_life"`
	SurfaceType   string   `json:"surface_type"`
	WaterPresent  bool     `json:"water_present"`
	PeoplePresent bool     `json:"people_present"`
	ImageQuality  string   `json:"image_quality"`
	IsCivicIssue  bool     `json:"is_civic_issue"`
	Confidence    float64  `json:"confidence"`
	Alternatives  []string `json:"alternatives"`
	Rationale     string   `json:"rationale"`
}

// Decision is the Result after the platform's own rules have been applied.
type Decision struct {
	Result
	Outcome Outcome `json:"outcome"`
	// Overridden records that a rule changed what the model said. Never
	// silent: an override is a disagreement worth being able to audit.
	Overridden        bool   `json:"overridden"`
	OverrideReason    string `json:"override_reason,omitempty"`
	RedactionRequired bool   `json:"redaction_required"`
	RejectionReason   string `json:"rejection_reason,omitempty"`
}

// Below this, the model is guessing and the citizen should choose.
const confidenceFloor = 0.65

// hazardous subcategories are lethal often enough that a model's "no" is not
// good enough. The list is a floor, not a ceiling — the model may still flag
// something not on it.
var hazardous = map[string]bool{
	"open manhole":          true,
	"missing manhole cover": true,
	"live wire":             true,
	"exposed electrical":    true,
	"collapsed wall":        true,
	"cracked bridge/fob":    true,
	"distressed building":   true,
	"unsafe scaffolding":    true,
	"sewage overflow":       true,
	"open drain":            true,
}

// Apply runs the platform's rules over a model result.
func Apply(r Result, covered Coverage) Decision {
	d := Decision{Result: r}

	// A photograph with people in it can never produce a public derivative
	// until redaction has run (hard rule 5).
	d.RedactionRequired = r.PeoplePresent

	// The hazard override comes first: it must survive every other outcome,
	// because an unreadable photograph of an open manhole is still an open
	// manhole.
	if !r.HazardToLife && hazardous[strings.ToLower(strings.TrimSpace(r.Subcategory))] {
		d.HazardToLife = true
		d.Overridden = true
		d.OverrideReason = fmt.Sprintf(
			"%q is on the known-hazard list, so it is treated as a hazard "+
				"whatever the classifier concluded", r.Subcategory)
	}

	switch {
	case r.ImageQuality == "unusable":
		d.Outcome = NeedsRetake

	case !r.IsCivicIssue:
		d.Outcome = OutOfScope
		d.RejectionReason = rejectionFor(r.Category)

	case !KnownCategory(r.Category):
		// Structured output guarantees a string, not a meaningful one. A
		// category with no department behind it cannot be routed.
		d.Outcome = NeedsConfirmation
		d.Alternatives = TopCategories()

	case r.Confidence < confidenceFloor:
		d.Outcome = NeedsConfirmation
		if len(d.Alternatives) == 0 {
			d.Alternatives = TopCategories()
		}

	case !covered.Covers(r.Category):
		// Recognised, and nowhere to send it.
		d.Outcome = NotYetCovered
		d.RejectionReason = notYetCoveredFor(r.Category, r.Subcategory)

	default:
		d.Outcome = Accepted
	}

	return d
}

// notYetCoveredFor tells the citizen what the platform recognised and that it
// does not cover it yet. It names the thing, because "not supported" tells
// them nothing, and it promises nothing, because the platform does not act on
// what it cannot route.
func notYetCoveredFor(category, subcategory string) string {
	what := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(category)), "_", " ")
	if s := strings.TrimSpace(subcategory); s != "" {
		what = strings.ToLower(s)
	}
	return fmt.Sprintf(
		"This looks like %s. TraceSarkar covers road defects in this ward so far, "+
			"and does not yet have the department contacts to take %s anywhere. "+
			"Your photograph is kept, and you can delete it.", what, what)
}

// rejectionFor explains what the platform thinks it is looking at, in the
// citizen's terms. Specific, never generic, and never a rebuke — the person
// took the trouble to photograph something.
func rejectionFor(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "private_property", "private_building":
		return "This looks like a private building or premises. " +
			"TraceSarkar covers public infrastructure — roads, footpaths, drains, " +
			"street lighting and public land."
	case "vehicle":
		return "This looks like a vehicle rather than public infrastructure. " +
			"Parking and traffic offences go to the traffic police, not the corporation."
	default:
		return "This does not look like public infrastructure. " +
			"TraceSarkar covers roads, footpaths, drains, street lighting and public land. " +
			"If this is public, say so and it will be looked at."
	}
}
