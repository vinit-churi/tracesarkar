package action

import (
	"strings"
	"testing"
)

func facts() Facts {
	return Facts{
		Category:         "waste",
		Subcategory:      "illegal dumping",
		Ward:             "R/C",
		Authority:        "BMC",
		Where:            "S.V. Road, Borivali West",
		Lat:              19.23409,
		Lon:              72.84419,
		Department:       "Solid Waste Management — R/Central Ward",
		Officer:          "Assistant Engineer (SWM) R/Central",
		Office:           "Chandavarkar Road, Borivali (West), Mumbai 400092",
		DeadlineHours:    24,
		DeadlineCitation: "R/Central SWM RTI handbook, §4(1)(b)(iii); MMC Act 1888 s.365(a)",
	}
}

func TestDraftNamesTheDeskAndSaysWhatToDo(t *testing.T) {
	d := Draft(facts())

	// The desk, not just the body. "Complain to BMC" is what every predecessor
	// platform says and it is why they go nowhere.
	if !strings.Contains(d.Text, "Assistant Engineer (SWM) R/Central") {
		t.Error("the draft does not name the officer")
	}
	if !strings.Contains(d.Text, "S.V. Road, Borivali West") {
		t.Error("the draft does not say where")
	}
	// Coordinates, because a street name alone does not find a heap of rubbish.
	if !strings.Contains(d.Text, "19.23409") {
		t.Error("the draft does not carry the position")
	}

	if len(d.Channels) == 0 {
		t.Fatal("a draft with nowhere to send it is not an action")
	}
	var named bool
	for _, c := range d.Channels {
		if strings.Contains(c.Name, "MARG") || strings.Contains(c.Name, "1916") {
			named = true
		}
	}
	if !named {
		t.Error("none of the channels is one BMC actually runs")
	}
}

// Hard rule 1. Nothing here files anything, and nothing may suggest it does.
func TestADraftNeverClaimsToFileAnything(t *testing.T) {
	d := Draft(facts())
	whole := d.Text + " " + d.WhatHappensNext
	for _, c := range d.Channels {
		whole += " " + c.Name + " " + c.How
	}

	for _, forbidden := range []string{
		"we will file", "we have filed", "submitted on your behalf",
		"we will submit", "automatically",
	} {
		if strings.Contains(strings.ToLower(whole), forbidden) {
			t.Errorf("the draft says %q", forbidden)
		}
	}
	if !strings.Contains(strings.ToLower(d.WhatHappensNext), "you") {
		t.Error("it must be clear the person sends it, not the platform")
	}
}

// Hard rule 3. It states what is there and asks for the duty to be performed.
// It does not conclude that anyone is at fault.
func TestADraftNeverAssertsWrongdoing(t *testing.T) {
	d := Draft(facts())
	lower := strings.ToLower(d.Text)
	for _, forbidden := range []string{
		"negligen", "corrupt", "failed to", "guilty", "deliberate", "scam",
	} {
		if strings.Contains(lower, forbidden) {
			t.Errorf("the draft asserts wrongdoing: %q", forbidden)
		}
	}
}

// D085: the clock runs from notification, so the draft must not imply one is
// already running.
func TestTheDeadlineIsDescribedAsStartingOnLodging(t *testing.T) {
	d := Draft(facts())
	next := strings.ToLower(d.WhatHappensNext)

	if !strings.Contains(next, "24 hours") {
		t.Error("the deadline is not stated")
	}
	if !strings.Contains(next, "once") && !strings.Contains(next, "from the time") {
		t.Errorf("the deadline does not say what starts it: %q", d.WhatHappensNext)
	}
	// Hard rule 2: a legal constant shown without its source is not evidence.
	if !strings.Contains(d.DeadlineCitation, "MMC Act") {
		t.Error("the deadline is shown without its citation")
	}
	// And the reference number is the thing everything later depends on.
	if !strings.Contains(next, "reference") {
		t.Error("it does not ask for the reference number")
	}
}

// A category with no department recorded has no next step, and saying nothing
// is better than inventing a desk.
func TestNoDepartmentMeansNoDraft(t *testing.T) {
	f := facts()
	f.Department = ""
	f.Officer = ""

	d := Draft(f)
	if d.Available {
		t.Error("a draft was offered with no department behind it")
	}
	if d.Text != "" {
		t.Error("text was produced for a desk that is not known")
	}
	if !strings.Contains(strings.ToLower(d.WhatHappensNext), "do not") &&
		!strings.Contains(strings.ToLower(d.WhatHappensNext), "not yet") {
		t.Errorf("it does not say why there is nothing to do: %q", d.WhatHappensNext)
	}
}

// Without a deadline on record the draft still works; it just does not claim
// one. The water-supply handbook publishes no repair timeline (D098).
func TestADraftWithoutADeadlineDoesNotInventOne(t *testing.T) {
	f := facts()
	f.DeadlineHours = 0
	f.DeadlineCitation = ""

	d := Draft(f)
	if !d.Available {
		t.Fatal("a known desk with no published deadline is still actionable")
	}
	if strings.Contains(d.WhatHappensNext, "48") || strings.Contains(d.WhatHappensNext, "24 hours") {
		t.Error("it borrowed a deadline from somewhere else")
	}
	if !strings.Contains(strings.ToLower(d.WhatHappensNext), "no published") {
		t.Errorf("it does not say the deadline is unknown: %q", d.WhatHappensNext)
	}
}
