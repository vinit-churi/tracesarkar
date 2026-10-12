package issue

import (
	"strings"
	"testing"
	"time"
)

func at(min int) time.Time {
	return time.Date(2026, 10, 12, 9, min, 0, 0, time.UTC)
}

// An issue's history is the evidence. "Reported on the 2nd, the corporation
// said fixed on the 5th, photographed unfixed on the 9th" is the claim this
// platform exists to be able to make — so the history has to be something
// nobody can quietly revise, including us.
func TestTheChainDetectsAnAlteredHistory(t *testing.T) {
	events := []Event{
		{Kind: "reported", At: at(0), Detail: "first sighting"},
		{Kind: "notified", At: at(10), Detail: "MARG-2026-1012-001"},
		{Kind: "claimed_resolved", At: at(20), Detail: "authority marked it done"},
	}
	chained := Chain(events)

	if err := Verify(chained); err != nil {
		t.Fatalf("an untouched chain did not verify: %v", err)
	}

	// Change what the authority is recorded as having said.
	tampered := append([]Event(nil), chained...)
	tampered[1].Detail = "MARG-2026-1012-999"
	if err := Verify(tampered); err == nil {
		t.Error("an edited event passed verification")
	}

	// Remove the inconvenient middle of the story.
	dropped := []Event{chained[0], chained[2]}
	if err := Verify(dropped); err == nil {
		t.Error("a removed event passed verification")
	}

	// Reorder, so the claim of a fix appears to precede the report.
	swapped := []Event{chained[0], chained[2], chained[1]}
	if err := Verify(swapped); err == nil {
		t.Error("a reordered chain passed verification")
	}
}

func TestTheFirstEventAnchorsTheChain(t *testing.T) {
	chained := Chain([]Event{{Kind: "reported", At: at(0)}})
	if chained[0].PrevHash != "" {
		t.Error("the first event points at something before it")
	}
	if chained[0].Hash == "" {
		t.Error("the first event has no hash")
	}
	if err := Verify(chained); err != nil {
		t.Errorf("a single-event chain did not verify: %v", err)
	}
}

// Appending must not disturb what is already there, or the chain would have
// to be rewritten on every event and would prove nothing.
func TestAppendingLeavesTheEarlierEventsAlone(t *testing.T) {
	first := Chain([]Event{{Kind: "reported", At: at(0)}})
	grown := Append(first, Event{Kind: "notified", At: at(10)})

	if grown[0].Hash != first[0].Hash {
		t.Error("appending changed the hash of an earlier event")
	}
	if grown[1].PrevHash != first[0].Hash {
		t.Error("the new event does not point at the one before it")
	}
	if err := Verify(grown); err != nil {
		t.Errorf("the grown chain did not verify: %v", err)
	}
}

func TestVerifySaysWhichEventBroke(t *testing.T) {
	chained := Chain([]Event{
		{Kind: "reported", At: at(0)},
		{Kind: "notified", At: at(10)},
		{Kind: "claimed_resolved", At: at(20)},
	})
	chained[2].Detail = "edited"

	err := Verify(chained)
	if err == nil {
		t.Fatal("no error")
	}
	// A failure nobody can locate is a failure nobody can act on.
	if !strings.Contains(err.Error(), "3") && !strings.Contains(err.Error(), "claimed_resolved") {
		t.Errorf("the error does not say which event broke: %v", err)
	}
}

func TestAnEmptyChainIsValid(t *testing.T) {
	if err := Verify(nil); err != nil {
		t.Errorf("an issue with no events yet is not a tampered one: %v", err)
	}
}
