package issue

import "testing"

func candidate(id string, category string, distance float64, status string) Candidate {
	return Candidate{IssueID: id, Category: category, DistanceM: distance, Status: status}
}

func TestDecide(t *testing.T) {
	tests := []struct {
		name      string
		accuracyM float64
		category  string
		near      []Candidate
		want      Action
		wantIssue string
	}{
		{
			// The whole point of system 4. Two people photographing the same
			// pothole must not produce two issues — counting the same hole
			// twice is the failure mode of every platform that came before.
			name:      "a second report of the same defect joins the first",
			accuracyM: 6, category: "road_defect",
			near: []Candidate{candidate("i1", "road_defect", 4, "open")},
			want: Join, wantIssue: "i1",
		},
		{
			// Rubbish dumped beside a pothole is a different problem, owned
			// by a different department. Merging them would send one
			// complaint to a desk that cannot act on half of it.
			name:      "a different kind of problem in the same place is its own issue",
			accuracyM: 6, category: "waste",
			near: []Candidate{candidate("i1", "road_defect", 2, "open")},
			want: Create,
		},
		{
			name:      "nothing nearby is a new issue",
			accuracyM: 6, category: "road_defect",
			near: nil,
			want: Create,
		},
		{
			// Further apart than the fix could explain. Two potholes thirty
			// metres apart on a six-metre fix are two potholes, and merging
			// them hides one.
			name:      "a defect beyond what the fix could explain is its own issue",
			accuracyM: 6, category: "road_defect",
			near: []Candidate{candidate("i1", "road_defect", 30, "open")},
			want: Create,
		},
		{
			// A loose fix widens what might be the same thing. Two reports
			// twenty metres apart on a forty-metre fix may well be one
			// pothole seen from both ends of the error.
			name:      "a loose fix widens what counts as the same place",
			accuracyM: 40, category: "road_defect",
			near: []Candidate{candidate("i1", "road_defect", 20, "open")},
			want: Join, wantIssue: "i1",
		},
		{
			// The nearest wins, so a report does not attach to whichever
			// issue happened to be created first.
			name:      "the nearest of several is the one joined",
			accuracyM: 10, category: "road_defect",
			near: []Candidate{
				candidate("far", "road_defect", 9, "open"),
				candidate("near", "road_defect", 2, "open"),
			},
			want: Join, wantIssue: "near",
		},
		{
			// The authority said it was fixed and here it is again. That is
			// the most valuable event this platform can record, and it must
			// not look like a new problem — the history is the point.
			name:      "a defect reported again after the authority claimed a fix reopens it",
			accuracyM: 6, category: "road_defect",
			near: []Candidate{candidate("i1", "road_defect", 3, "claimed_resolved")},
			want: Reopen, wantIssue: "i1",
		},
		{
			// A citizen photographed the fix. A later report at the same spot
			// is a new failure, not a continuation of the old one — keeping
			// them separate is what makes "fixed twice in a year" visible.
			name:      "a defect at a confirmed-fixed location is a new issue",
			accuracyM: 6, category: "road_defect",
			near: []Candidate{candidate("i1", "road_defect", 3, "citizen_confirmed")},
			want: Create,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Decide(tt.accuracyM, tt.category, tt.near)
			if d.Action != tt.want {
				t.Fatalf("action = %q, want %q (%s)", d.Action, tt.want, d.Basis)
			}
			if tt.wantIssue != "" && d.IssueID != tt.wantIssue {
				t.Errorf("joined %q, want %q", d.IssueID, tt.wantIssue)
			}
			// Every decision has to be explicable afterwards: merging two
			// reports is a claim that they are the same thing in the world.
			if d.Basis == "" {
				t.Error("the decision carries no basis")
			}
		})
	}
}

// A capture with no accuracy reading cannot be merged on distance, because
// there is no distance it can be trusted within.
func TestAnUnknownFixDoesNotMerge(t *testing.T) {
	d := Decide(0, "road_defect", []Candidate{candidate("i1", "road_defect", 2, "open")})
	if d.Action != Create {
		t.Errorf("merged on an unknown fix: %q — %s", d.Action, d.Basis)
	}
}
