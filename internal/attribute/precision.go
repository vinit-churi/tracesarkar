package attribute

// Summary is how a review queue stands.
//
// The number that matters is precision over what a person actually judged.
// A join that answers confidently and wrongly is the failure this whole
// system is built to avoid, so the denominator is judged answers, never the
// size of the queue.
type Summary struct {
	Total   int `json:"total"`
	Correct int `json:"correct"`
	Wrong   int `json:"wrong"`
	Unsure  int `json:"unsure"`
	Pending int `json:"pending"`
}

// Reviewed is how many a person has looked at.
func (s Summary) Reviewed() int { return s.Correct + s.Wrong + s.Unsure }

// judged is how many produced an opinion about the answer. "Unsure" is an
// opinion about the evidence, not about the join, so it is excluded rather
// than counted as a success.
func (s Summary) judged() int { return s.Correct + s.Wrong }

// HasPrecision reports whether anyone has judged enough to quote a number.
func (s Summary) HasPrecision() bool { return s.judged() > 0 }

// Precision is the share of judged answers that were right.
func (s Summary) Precision() float64 {
	if !s.HasPrecision() {
		return 0
	}
	return float64(s.Correct) / float64(s.judged())
}

// NeedsMoreEvidence reports whether the queue has settled the question.
//
// It has not if anything is still unreviewed, or if a large share of what was
// reviewed could not be decided — a pile of "unsure" means the evidence is
// too thin to conclude anything, which is itself worth knowing.
func (s Summary) NeedsMoreEvidence() bool {
	if s.Pending > 0 {
		return true
	}
	reviewed := s.Reviewed()
	if reviewed == 0 {
		return true
	}
	return float64(s.Unsure)/float64(reviewed) > 0.2
}
