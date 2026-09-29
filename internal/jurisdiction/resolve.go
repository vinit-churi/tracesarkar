package jurisdiction

import (
	"fmt"
	"strings"
)

// Confidence bands, matching the attribution surface so a citizen sees one
// vocabulary across the product.
type Confidence string

const (
	High      Confidence = "high"
	CheckThis Confidence = "check_this"
	None      Confidence = "none"
)

// WardCandidate is a ward the point falls in, or near.
type WardCandidate struct {
	Ward      string `json:"ward"`
	Authority string `json:"authority"`
	Inside    bool   `json:"inside"`
	// DistanceToEdgeM is the distance to the ward's boundary line — not to the
	// polygon, which is zero anywhere inside it. Nearness to the *edge* is
	// what decides whether the answer is safe.
	DistanceToEdgeM float64 `json:"distance_to_edge_m"`
}

// Resolution is which authority owns the point, and how sure we are.
type Resolution struct {
	Ward          string          `json:"ward"`
	Authority     string          `json:"authority"`
	Confidence    Confidence      `json:"confidence"`
	Alternatives  []WardCandidate `json:"alternatives"`
	Basis         string          `json:"basis"`
	NeedsQuestion bool            `json:"needs_question"`
}

// Resolve decides which ward owns a point.
//
// The rule is not "which polygon contains it" — PostGIS answers that. It is
// "is that answer safe given how uncertain the position is". A point six
// metres inside a ward line, reported by a phone accurate to ten, is not
// reliably in either ward, and a complaint sent to the wrong desk is the
// failure a citizen does not forgive.
func Resolve(accuracyM float64, candidates []WardCandidate) Resolution {
	var inside []WardCandidate
	for _, c := range candidates {
		if c.Inside {
			inside = append(inside, c)
		}
	}

	switch {
	case len(inside) == 0:
		r := Resolution{Confidence: None, NeedsQuestion: true}
		r.Basis = "This point is outside every ward boundary we hold. " +
			"Greater Mumbai is not the whole metropolitan region, so it may " +
			"belong to another corporation or to a state authority."
		if len(candidates) > 0 {
			r.Alternatives = candidates
			r.Basis += fmt.Sprintf(" The nearest we hold is %s, %.0f m away.",
				candidates[0].Ward, candidates[0].DistanceToEdgeM)
		}
		return r

	case len(inside) > 1:
		// Two polygons claiming one point means the boundary set is wrong.
		// Choosing one would hide a data problem behind a routed complaint.
		r := Resolution{
			Ward: inside[0].Ward, Authority: inside[0].Authority,
			Confidence: CheckThis, Alternatives: inside[1:], NeedsQuestion: true,
		}
		r.Basis = fmt.Sprintf(
			"More than one ward boundary contains this point — %s. "+
				"The boundary data disagrees with itself here.",
			strings.Join(wardNames(inside), " and "))
		return r
	}

	owner := inside[0]
	r := Resolution{Ward: owner.Ward, Authority: owner.Authority}

	// Neighbours whose line runs within the position's own error.
	var near []WardCandidate
	for _, c := range candidates {
		if !c.Inside && c.DistanceToEdgeM <= accuracyM {
			near = append(near, c)
		}
	}

	// A margin: the boundary must be further away than the position could be
	// wrong, with a floor so a perfect fix still leaves a little room.
	margin := accuracyM
	if margin < 10 {
		margin = 10
	}

	switch {
	case owner.DistanceToEdgeM > margin && len(near) == 0:
		r.Confidence = High
		r.Basis = fmt.Sprintf(
			"The point is inside %s %s, %.0f m from the nearest ward line — "+
				"further than the %.0f m the position could be out.",
			owner.Authority, owner.Ward, owner.DistanceToEdgeM, accuracyM)

	default:
		r.Confidence = CheckThis
		r.NeedsQuestion = true
		r.Alternatives = near
		if len(near) > 0 {
			r.Basis = fmt.Sprintf(
				"The point falls in %s %s, but the line with %s runs %.0f m away "+
					"and the position may be out by %.0f m. It could belong to either.",
				owner.Authority, owner.Ward, strings.Join(wardNames(near), " or "),
				near[0].DistanceToEdgeM, accuracyM)
		} else {
			r.Basis = fmt.Sprintf(
				"The point falls in %s %s, but only %.0f m from a ward line, "+
					"and the position may be out by %.0f m.",
				owner.Authority, owner.Ward, owner.DistanceToEdgeM, accuracyM)
		}
	}
	return r
}

func wardNames(cs []WardCandidate) []string {
	out := make([]string, 0, len(cs))
	for _, c := range cs {
		out = append(out, c.Ward)
	}
	return out
}
