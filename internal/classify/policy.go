package classify

import "fmt"

// DataKind is whose data a call would send to a provider. The distinction
// matters because the platform can consent on its own behalf and cannot
// consent on a stranger's.
type DataKind string

const (
	// OwnTestData is the maintainer's own photographs, taken to build and
	// evaluate the classifier. Theirs to give.
	OwnTestData DataKind = "own_test_data"
	// CitizenPhotograph is a capture from a member of the public. It contains
	// faces and number plates before redaction, belonging to people who were
	// never asked.
	CitizenPhotograph DataKind = "citizen_photograph"
)

// ModelPolicy is what a provider says about a model, as reported by the
// gateway's own model listing rather than by its marketing copy.
type ModelPolicy struct {
	ID string
	// NoTraining and ZDR are "all", "some" or "none" — "some" meaning only
	// some providers behind the model honour it.
	NoTraining       string
	ZDR              string
	StructuredOutput bool
	Vision           bool
}

// AllowedFor reports whether this model may be used on this kind of data.
//
// It exists so that "we will switch to a safer model later" is enforced by the
// code rather than remembered by a person. A model that trains on its inputs
// cannot be asked to forget, so a citizen's deletion request could never be
// honoured — which makes it a promise the platform must not make.
func (m ModelPolicy) AllowedFor(kind DataKind) error {
	if !m.Vision {
		return fmt.Errorf("model %s cannot read images; classification needs vision", m.ID)
	}
	if !m.StructuredOutput {
		return fmt.Errorf(
			"model %s does not support structured output, and parsing free text "+
				"into a classification fails silently", m.ID)
	}

	if kind == CitizenPhotograph {
		// "some" is not a guarantee that can be repeated to the person whose
		// face is in the photograph.
		if m.NoTraining != "all" {
			return fmt.Errorf(
				"model %s may be trained on its inputs (no_training=%q); a citizen "+
					"photograph carries faces and plates that could never be withdrawn",
				m.ID, m.NoTraining)
		}
		if m.ZDR != "all" {
			return fmt.Errorf(
				"model %s does not offer zero data retention (zdr=%q); a deletion "+
					"request could not be honoured", m.ID, m.ZDR)
		}
	}
	return nil
}
