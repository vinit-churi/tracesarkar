// Package issue decides whether a new report is a new problem or another
// sighting of one already known.
//
// This is the difference between a platform that accumulates complaints and
// one that accumulates evidence. Two people photographing the same pothole
// must produce one issue with two reports behind it, not two issues — counting
// the same hole twice is how every predecessor ended up with a map full of
// noise and nothing anybody acted on.
//
// Merging is a claim that two photographs are of the same thing in the world,
// so every decision carries its reasoning and the distance it turned on.
package issue

import "fmt"

// Action is what to do with a report.
type Action string

const (
	// Create: nothing already known explains it.
	Create Action = "create"
	// Join: another sighting of an open issue.
	Join Action = "join"
	// Reopen: the authority said this was fixed and here it is again. The
	// most valuable event the platform can record.
	Reopen Action = "reopen"
)

// Candidate is an existing issue near the new report.
type Candidate struct {
	IssueID   string
	Category  string
	DistanceM float64
	Status    string
}

// Decision is what to do, and why.
type Decision struct {
	Action  Action
	IssueID string
	// DistanceM is how far the joined issue was. Kept so a merge can be
	// audited later: it is an assertion about the world, not bookkeeping.
	DistanceM float64
	Basis     string
}

// minMergeM is how close two sightings must be before a good fix will merge
// them.
//
// Eight metres is about the length of a car. Two potholes further apart than
// that are two potholes a crew would be sent to separately, and merging them
// hides one — which is the worse error, because an issue that exists in the
// data conceals the problem it absorbed.
const minMergeM = 8

// Decide chooses between a new issue and another sighting of a known one.
//
// The radius comes from the fix, not from a constant: two reports twenty
// metres apart on a forty-metre fix may be one pothole seen from either end of
// the error, while the same twenty metres on a six-metre fix is two potholes.
func Decide(accuracyM float64, category string, near []Candidate) Decision {
	if accuracyM <= 0 {
		// No accuracy means no distance the position can be trusted within,
		// so there is no honest radius to merge inside.
		return Decision{Action: Create, Basis: "The position carries no accuracy " +
			"reading, so it cannot be judged to be the same place as anything else."}
	}

	radius := accuracyM
	if radius < minMergeM {
		radius = minMergeM
	}

	var best *Candidate
	for i := range near {
		c := near[i]
		// A different kind of problem in the same place is a different
		// problem, owned by a different department.
		if c.Category != category {
			continue
		}
		// A location whose fix a citizen photographed starts again: keeping
		// the two apart is what makes "fixed twice in a year" visible.
		if c.Status == "citizen_confirmed" {
			continue
		}
		if c.DistanceM > radius {
			continue
		}
		if best == nil || c.DistanceM < best.DistanceM {
			best = &near[i]
		}
	}

	if best == nil {
		return Decision{Action: Create, Basis: fmt.Sprintf(
			"No open %s issue lies within %.0f m, the distance this position "+
				"could be out by.", readable(category), radius)}
	}

	if best.Status == "claimed_resolved" {
		return Decision{
			Action: Reopen, IssueID: best.IssueID, DistanceM: best.DistanceM,
			Basis: fmt.Sprintf(
				"An issue %.0f m away was marked resolved and the problem is "+
					"reported again. Recorded as the same issue reopening, not "+
					"as a new one.", best.DistanceM),
		}
	}

	return Decision{
		Action: Join, IssueID: best.IssueID, DistanceM: best.DistanceM,
		Basis: fmt.Sprintf(
			"An open %s issue lies %.0f m away, inside the %.0f m this position "+
				"could be out by, so this is recorded as another sighting of it "+
				"rather than a second problem.",
			readable(category), best.DistanceM, radius),
	}
}

func readable(category string) string {
	switch category {
	case "road_defect":
		return "road defect"
	case "water_drainage":
		return "water or drainage"
	case "street_furniture":
		return "street furniture"
	default:
		return category
	}
}
