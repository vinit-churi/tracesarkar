package jurisdiction

import (
	"strings"
	"testing"
)

func TestAPointWellInsideAWardIsRoutedConfidently(t *testing.T) {
	got := Resolve(8, []WardCandidate{
		{Ward: "R/C", Authority: "BMC", Inside: true, DistanceToEdgeM: 900},
	})

	if got.Confidence != High {
		t.Errorf("confidence: got %q, want %q", got.Confidence, High)
	}
	if got.Ward != "R/C" || got.Authority != "BMC" {
		t.Errorf("routed to %s / %s", got.Authority, got.Ward)
	}
	if got.NeedsQuestion {
		t.Error("a point 900 m inside a ward needs no question")
	}
}

func TestAPointNearAWardLineAsksBeforeRouting(t *testing.T) {
	// 6 m from the boundary with a 10 m fix: the report could belong to
	// either ward, and sending it to the wrong desk is the failure a citizen
	// does not forgive.
	got := Resolve(10, []WardCandidate{
		{Ward: "R/C", Authority: "BMC", Inside: true, DistanceToEdgeM: 6},
		{Ward: "R/S", Authority: "BMC", Inside: false, DistanceToEdgeM: 6},
	})

	if got.Confidence != CheckThis {
		t.Errorf("confidence: got %q, want %q", got.Confidence, CheckThis)
	}
	if !got.NeedsQuestion {
		t.Error("a point on a ward line is exactly when to ask the one question")
	}
	if len(got.Alternatives) == 0 || got.Alternatives[0].Ward != "R/S" {
		t.Errorf("the neighbouring ward must be offered: %+v", got.Alternatives)
	}
	if !strings.Contains(got.Basis, "R/S") {
		t.Errorf("the basis must name the other ward: %q", got.Basis)
	}
}

func TestAPointInNoWardSaysSoRatherThanGuessing(t *testing.T) {
	got := Resolve(8, nil)

	if got.Confidence != None {
		t.Errorf("confidence: got %q, want %q", got.Confidence, None)
	}
	if got.Ward != "" {
		t.Errorf("no ward must mean no ward, got %q", got.Ward)
	}
	// Outside BMC is a real answer — Greater Mumbai is not the whole MMR.
	if !strings.Contains(strings.ToLower(got.Basis), "outside") {
		t.Errorf("the basis should say the point is outside the area we hold: %q", got.Basis)
	}
	if !got.NeedsQuestion {
		t.Error("an unroutable point needs a human")
	}
}

func TestAPoorFixNearALineCannotBeConfident(t *testing.T) {
	got := Resolve(80, []WardCandidate{
		{Ward: "R/C", Authority: "BMC", Inside: true, DistanceToEdgeM: 40},
	})

	if got.Confidence == High {
		t.Errorf("a 80 m fix 40 m from a ward line is not confident: %+v", got)
	}
}

func TestAPoorFixDeepInsideAWardIsStillFine(t *testing.T) {
	// Accuracy only matters relative to the boundary. A 60 m fix in the middle
	// of a ward still lands in that ward.
	got := Resolve(60, []WardCandidate{
		{Ward: "R/C", Authority: "BMC", Inside: true, DistanceToEdgeM: 2000},
	})

	if got.Confidence != High {
		t.Errorf("confidence: got %q, want %q", got.Confidence, High)
	}
}

func TestOverlappingPolygonsAreTreatedAsDoubt(t *testing.T) {
	// Two wards both claiming to contain the point means the boundary set is
	// wrong. Picking one silently would hide a data problem behind a routed
	// complaint.
	got := Resolve(8, []WardCandidate{
		{Ward: "R/C", Authority: "BMC", Inside: true, DistanceToEdgeM: 500},
		{Ward: "R/S", Authority: "BMC", Inside: true, DistanceToEdgeM: 400},
	})

	if got.Confidence != CheckThis {
		t.Errorf("confidence: got %q, want %q", got.Confidence, CheckThis)
	}
}

func TestTheBasisNeverBlamesAnyone(t *testing.T) {
	// Hard rule 3, on the routing surface too.
	forbidden := []string{"failed", "negligent", "responsible for", "at fault", "ignored"}
	for _, r := range []Resolution{
		Resolve(8, []WardCandidate{{Ward: "R/C", Authority: "BMC", Inside: true, DistanceToEdgeM: 900}}),
		Resolve(10, []WardCandidate{{Ward: "R/C", Authority: "BMC", Inside: true, DistanceToEdgeM: 6}}),
		Resolve(8, nil),
	} {
		lower := strings.ToLower(r.Basis)
		for _, w := range forbidden {
			if strings.Contains(lower, w) {
				t.Errorf("basis asserts fault (%q): %q", w, r.Basis)
			}
		}
	}
}
