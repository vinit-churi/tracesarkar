// Package attribute decides which road contract covers a reported point.
//
// This is the platform's central claim, so the rules here are deliberately
// conservative. The expensive error is not "no match" — it is a confident
// match to the wrong road, which attributes a defect to a contractor who never
// worked on it. Everything below prefers to ask a question rather than guess.
//
// The decision is pure: it takes candidates and their distances and returns a
// band. Nothing here touches the database, so every rule is testable.
package attribute

import (
	"fmt"
	"strings"
	"time"
)

// Confidence is the band shown to a person. Bands, not raw scores: a
// percentage invites false precision about a judgement this coarse.
type Confidence string

const (
	High      Confidence = "high"
	CheckThis Confidence = "check_this"
	None      Confidence = "none"
)

// Candidate is a road segment near the reported point.
type Candidate struct {
	WorkID         string     `json:"work_id"`
	WorkCode       string     `json:"work_code"`
	ContractorName string     `json:"contractor_name"`
	LocationName   string     `json:"location_name"`
	DistanceM      float64    `json:"distance_m"`
	EndDate        *time.Time `json:"end_date"`
}

// Attribution is the answer, with the reasoning attached.
type Attribution struct {
	Confidence    Confidence  `json:"confidence"`
	Match         *Candidate  `json:"match"`
	Alternatives  []Candidate `json:"alternatives"`
	SearchRadiusM float64     `json:"search_radius_m"`
	// Basis is the join in plain language (screen spec 3.4) — never a raw
	// score alone, and never an assertion of wrongdoing (hard rule 3).
	Basis       string `json:"basis"`
	WouldChange string `json:"would_change"`
}

const (
	// A carriageway is around 10–15 m wide, so a point genuinely on a road can
	// sit this far from the centreline the geometry records.
	roadHalfWidthM = 15.0

	// Above this, the capture screen already tells the reporter the fix is
	// weak. An attribution built on one cannot claim certainty.
	weakFixM = 50.0

	// The smallest gap that separates two roads when the fix is perfect.
	minSeparationM = 5.0
)

// SearchRadius is how far from the point to look, widened by how uncertain the
// point is.
func SearchRadius(accuracyM float64) float64 {
	if accuracyM <= 0 {
		// Unknown accuracy: search the width of a road and no further. A wider
		// search would invent candidates we have no basis to rank.
		return roadHalfWidthM
	}
	return roadHalfWidthM + accuracyM
}

// Decide picks the covering road, or declines to.
func Decide(accuracyM float64, candidates []Candidate) Attribution {
	a := Attribution{SearchRadiusM: SearchRadius(accuracyM)}

	if len(candidates) == 0 {
		a.Confidence = None
		a.Basis = fmt.Sprintf(
			"No road work with published geometry lies within %.0f m of this point. "+
				"That does not mean the road is uncovered — it means we have no "+
				"geometry for it yet.", a.SearchRadiusM)
		a.WouldChange = "Geometry published for this stretch, or a more precise position, " +
			"would let us answer."
		return a
	}

	nearest := candidates[0]
	a.Match = &nearest
	if len(candidates) > 1 {
		a.Alternatives = candidates[1:]
	}

	// How far apart two candidates must be before the nearer one is genuinely
	// nearer rather than nearer by less than the position's own error.
	separation := minSeparationM
	if accuracyM > separation {
		separation = accuracyM
	}

	// The rival that matters is the nearest candidate that would produce a
	// *different* answer. BMC splits one street across several work records,
	// so two candidates at the same distance are often the same road written
	// twice — one answer, not a junction. Treating that as ambiguous would
	// spend the single question the capture flow is allowed on a choice
	// between identical options.
	ambiguous := false
	var rival Candidate
	for _, c := range candidates[1:] {
		if sameAnswer(nearest, c) {
			continue
		}
		rival = c
		ambiguous = c.DistanceM-nearest.DistanceM < separation
		break
	}

	switch {
	case accuracyM <= 0:
		a.Confidence = CheckThis
		a.Basis = fmt.Sprintf(
			"The nearest road with published geometry is %s, %.1f m away. "+
				"The position carries no accuracy reading, so how far it may be out "+
				"is unknown.", describe(nearest), nearest.DistanceM)
		a.WouldChange = "A position with a known accuracy would settle this."

	case accuracyM > weakFixM:
		a.Confidence = CheckThis
		a.Basis = fmt.Sprintf(
			"The nearest road with published geometry is %s, %.1f m away. "+
				"The GPS fix is accurate only to about %.0f m, which is wider than "+
				"the road, so the point could sit on a neighbouring stretch.",
			describe(nearest), nearest.DistanceM, accuracyM)
		a.WouldChange = "A capture taken in the open, with a fix under 15 m, would settle this."

	case ambiguous:
		a.Confidence = CheckThis
		a.Basis = fmt.Sprintf(
			"Two roads are about equally close: %s at %.1f m and %s at %.1f m. "+
				"With a fix accurate to %.0f m they cannot be told apart.",
			describe(nearest), nearest.DistanceM,
			describe(rival), rival.DistanceM, accuracyM)
		a.WouldChange = "Confirming which road the problem is on would settle this."

	default:
		a.Confidence = High
		a.Basis = fmt.Sprintf(
			"The point lies %.1f m from %s, within a %.0f m search based on a fix "+
				"accurate to about %.0f m.",
			nearest.DistanceM, describe(nearest), a.SearchRadiusM, accuracyM)
		if rival.WorkID != "" {
			a.Basis += fmt.Sprintf(" The nearest road under a different contract is %.1f m away.",
				rival.DistanceM)
		}
		a.WouldChange = "A correction to the road geometry, or a more precise position, " +
			"would change this."
	}

	return a
}

// sameAnswer reports whether two candidates would attribute the defect
// identically. If they name the same contract package and the same company,
// which of them is nearer changes nothing a person would see.
func sameAnswer(a, b Candidate) bool {
	return a.WorkCode == b.WorkCode &&
		strings.EqualFold(strings.TrimSpace(a.ContractorName), strings.TrimSpace(b.ContractorName))
}

// describe names a candidate the way a person would: by street, falling back to
// the contract package when BMC published no name for the stretch.
func describe(c Candidate) string {
	if name := strings.TrimSpace(c.LocationName); name != "" {
		return name
	}
	if c.WorkCode != "" {
		return "the stretch under " + c.WorkCode
	}
	return "an unnamed stretch"
}
