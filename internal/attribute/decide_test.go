package attribute

import (
	"strings"
	"testing"
)

func candidates(distances ...float64) []Candidate {
	out := make([]Candidate, 0, len(distances))
	names := []string{"Yogi Nagar Road", "Jain Derasar Road", "Rokadia X Lane", "Link Road"}
	for i, d := range distances {
		out = append(out, Candidate{
			WorkID:         "work-" + string(rune('a'+i)),
			WorkCode:       "W-41" + string(rune('0'+i)),
			ContractorName: "M/s Example Infracon Pvt. Ltd",
			LocationName:   names[i%len(names)],
			DistanceM:      d,
		})
	}
	return out
}

func TestNothingNearbyIsStatedNotGuessed(t *testing.T) {
	got := Decide(6, nil)

	if got.Confidence != None {
		t.Errorf("confidence: got %q, want %q", got.Confidence, None)
	}
	if got.Match != nil {
		t.Error("no candidates must mean no match")
	}
	// The coverage-honesty line (K7): absence is stated, never silent.
	if !strings.Contains(strings.ToLower(got.Basis), "no road") {
		t.Errorf("the absence must be stated plainly: %q", got.Basis)
	}
}

func TestOneRoadClearlyNearestIsHighConfidence(t *testing.T) {
	got := Decide(6, candidates(3.2))

	if got.Confidence != High {
		t.Errorf("confidence: got %q, want %q", got.Confidence, High)
	}
	if got.Match == nil || got.Match.LocationName != "Yogi Nagar Road" {
		t.Fatalf("match: %+v", got.Match)
	}
}

func TestTwoRoadsTooCloseToSeparateNeedsAHumanLook(t *testing.T) {
	// A junction: the point is 4 m from one road and 6 m from another. With a
	// 6 m GPS fix those are not distinguishable, and picking one silently is
	// how a defect gets attributed to the wrong contractor.
	got := Decide(6, candidates(4.0, 6.0))

	if got.Confidence != CheckThis {
		t.Errorf("confidence: got %q, want %q", got.Confidence, CheckThis)
	}
	if len(got.Alternatives) == 0 {
		t.Error("the other candidate must be offered, not discarded")
	}
	if !strings.Contains(got.Basis, "Jain Derasar Road") {
		t.Errorf("the basis must name what it could not separate: %q", got.Basis)
	}
}

func TestAClearGapSeparatesTwoCandidates(t *testing.T) {
	// Same junction, but the second road is 40 m away: no ambiguity.
	got := Decide(6, candidates(3.0, 43.0))

	if got.Confidence != High {
		t.Errorf("confidence: got %q, want %q", got.Confidence, High)
	}
}

func TestAPoorGPSFixCanNeverProduceHighConfidence(t *testing.T) {
	// Above 50 m the capture screen already tells the reporter to step into
	// the open. An attribution built on it must not claim certainty.
	got := Decide(80, candidates(5.0))

	if got.Confidence == High {
		t.Errorf("an 80 m fix must not yield high confidence: %+v", got)
	}
	if !strings.Contains(strings.ToLower(got.Basis), "gps") &&
		!strings.Contains(strings.ToLower(got.Basis), "accur") {
		t.Errorf("the basis must say the fix is why: %q", got.Basis)
	}
}

func TestAnUnknownAccuracyIsTreatedAsUnknownNotAsPerfect(t *testing.T) {
	got := Decide(0, candidates(2.0))

	if got.Confidence == High {
		t.Error("a missing accuracy must not be read as a perfect fix")
	}
}

func TestTheBasisIsPlainLanguageWithTheNumbersInIt(t *testing.T) {
	got := Decide(6, candidates(3.2))

	// MatchBasis (screen spec 3.4): the join in plain language — matched
	// geometry, buffer distance, confidence band, and what would change it.
	// Never a raw score alone.
	for _, want := range []string{"3.2", "Yogi Nagar Road"} {
		if !strings.Contains(got.Basis, want) {
			t.Errorf("basis is missing %q: %q", want, got.Basis)
		}
	}
	if got.WouldChange == "" {
		t.Error("the basis must say what would change the answer")
	}
}

func TestTheBasisNeverAssertsWrongdoing(t *testing.T) {
	// Hard rule 3. The platform states facts and joins; it never concludes.
	forbidden := []string{"responsible for", "at fault", "failed", "negligent",
		"liable", "caused", "guilty"}

	for _, a := range []Attribution{
		Decide(6, candidates(3.2)),
		Decide(6, candidates(4.0, 6.0)),
		Decide(80, candidates(5.0)),
		Decide(6, nil),
	} {
		text := strings.ToLower(a.Basis + " " + a.WouldChange)
		for _, word := range forbidden {
			if strings.Contains(text, word) {
				t.Errorf("basis asserts wrongdoing (%q): %q", word, a.Basis)
			}
		}
	}
}

func TestSearchRadiusGrowsWithUncertainty(t *testing.T) {
	tight := SearchRadius(5)
	loose := SearchRadius(40)

	if !(loose > tight) {
		t.Errorf("a worse fix must search wider: %v vs %v", tight, loose)
	}
	if tight < 10 {
		t.Errorf("the radius must cover the width of a road, got %v", tight)
	}
}

func TestTwoSegmentsOfTheSameRoadAreNotAnAmbiguity(t *testing.T) {
	// BMC splits a street across several work records. Two candidates at the
	// same distance, under the same package and the same contractor, are one
	// answer written twice — not a junction. Asking someone to choose between
	// identical answers wastes the single question the flow is allowed.
	same := []Candidate{
		{WorkID: "a", WorkCode: "W-447", ContractorName: "M/s. BSCPL Infrastructure Ltd",
			LocationName: "RSC-53, Sector-5", DistanceM: 0.0},
		{WorkID: "b", WorkCode: "W-447", ContractorName: "M/s. BSCPL Infrastructure Ltd",
			LocationName: "RSC-53, Sector-5", DistanceM: 0.0},
	}

	got := Decide(6, same)

	if got.Confidence != High {
		t.Errorf("confidence: got %q, want %q — the answer is the same either way",
			got.Confidence, High)
	}
	if strings.Contains(got.Basis, "equally close") {
		t.Errorf("this is not an ambiguity: %q", got.Basis)
	}
}

func TestTwoDifferentContractorsAtTheSameDistanceStayAmbiguous(t *testing.T) {
	// The case that matters: same distance, different companies. Picking one
	// silently would name the wrong party.
	rival := []Candidate{
		{WorkID: "a", WorkCode: "W-447", ContractorName: "M/s. BSCPL Infrastructure Ltd",
			LocationName: "Link Road", DistanceM: 3.0},
		{WorkID: "b", WorkCode: "W-415", ContractorName: "M/s Other Infracon Pvt. Ltd",
			LocationName: "Service Road", DistanceM: 4.0},
	}

	got := Decide(6, rival)

	if got.Confidence != CheckThis {
		t.Errorf("confidence: got %q, want %q", got.Confidence, CheckThis)
	}
}

func TestSameRoadDifferentContractorIsStillAmbiguous(t *testing.T) {
	// Same street name, different package: a resurfacing boundary. Which
	// contractor covers the defect is a real question.
	got := Decide(6, []Candidate{
		{WorkID: "a", WorkCode: "W-447", ContractorName: "M/s. BSCPL Infrastructure Ltd",
			LocationName: "Link Road", DistanceM: 2.0},
		{WorkID: "b", WorkCode: "W-415", ContractorName: "M/s Other Infracon Pvt. Ltd",
			LocationName: "Link Road", DistanceM: 3.0},
	})

	if got.Confidence != CheckThis {
		t.Errorf("confidence: got %q, want %q", got.Confidence, CheckThis)
	}
}
