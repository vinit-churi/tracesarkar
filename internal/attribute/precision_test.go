package attribute

import "testing"

func TestPrecisionCountsOnlyWhatWasActuallyJudged(t *testing.T) {
	// Unreviewed rows are not evidence. Counting them as either right or
	// wrong would let the number drift with the size of the queue rather than
	// with the quality of the join.
	s := Summary{Total: 100, Correct: 40, Wrong: 10, Unsure: 5, Pending: 45}

	if got := s.Precision(); got != 0.8 {
		t.Errorf("precision: got %v, want 0.8 (40 of 50 judged)", got)
	}
	if got := s.Reviewed(); got != 55 {
		t.Errorf("reviewed: got %d, want 55", got)
	}
}

func TestPrecisionIsUndefinedBeforeAnyoneHasJudged(t *testing.T) {
	s := Summary{Total: 50, Pending: 50}

	if s.HasPrecision() {
		t.Error("precision must not be reported before anything is judged")
	}
	if got := s.Precision(); got != 0 {
		t.Errorf("got %v, want 0", got)
	}
}

func TestUnsureCountsAsNeitherRightNorWrong(t *testing.T) {
	// "Unsure" is a real answer about the *evidence*, not about the join. It
	// must not be quietly scored as a success.
	s := Summary{Total: 10, Correct: 5, Wrong: 0, Unsure: 5}

	if got := s.Precision(); got != 1.0 {
		t.Errorf("precision: got %v, want 1.0", got)
	}
	if !s.NeedsMoreEvidence() {
		t.Error("a queue that is half unsure has not settled the question")
	}
}

func TestAQueueWithNoDoubtIsSettled(t *testing.T) {
	s := Summary{Total: 50, Correct: 48, Wrong: 2}

	if s.NeedsMoreEvidence() {
		t.Error("a fully judged queue with little doubt is settled")
	}
}
