package issue

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Event is one thing that happened to an issue.
//
// The history is the evidence. "Reported on the 2nd, the corporation said
// fixed on the 5th, photographed unfixed on the 9th" is the claim this
// platform exists to be able to make, and a claim like that is only worth
// anything if the record behind it cannot be quietly revised — including by
// us. Each event carries the hash of the one before it, so an edit, a
// deletion or a reordering anywhere in the history is detectable from the
// end of it.
type Event struct {
	Kind   string    `json:"kind"`
	At     time.Time `json:"at"`
	Detail string    `json:"detail,omitempty"`
	Actor  string    `json:"actor,omitempty"`

	PrevHash string `json:"prev_hash,omitempty"`
	Hash     string `json:"hash"`
}

// digest is what the hash is taken over: the event's own content and the
// hash before it. Times are rendered to the second in UTC so a chain written
// on one machine verifies on another.
func digest(e Event, prev string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		prev,
		e.Kind,
		e.At.UTC().Format(time.RFC3339),
		e.Detail,
		e.Actor,
	}, "\x1f")))
	return hex.EncodeToString(sum[:])
}

// Chain stamps a sequence of events, oldest first.
func Chain(events []Event) []Event {
	out := make([]Event, 0, len(events))
	prev := ""
	for _, e := range events {
		e.PrevHash = prev
		e.Hash = digest(e, prev)
		out = append(out, e)
		prev = e.Hash
	}
	return out
}

// Append adds one event to an existing chain without disturbing it.
//
// Nothing already written is recomputed. A chain that had to be rewritten on
// every append would prove nothing, because rewriting is exactly the thing it
// exists to make detectable.
func Append(chain []Event, e Event) []Event {
	prev := ""
	if n := len(chain); n > 0 {
		prev = chain[n-1].Hash
	}
	e.PrevHash = prev
	e.Hash = digest(e, prev)
	return append(append([]Event(nil), chain...), e)
}

// Verify reports whether a history is the one that was written.
//
// It names the event that broke, because a failure nobody can locate is a
// failure nobody can act on.
func Verify(chain []Event) error {
	prev := ""
	for i, e := range chain {
		if e.PrevHash != prev {
			return fmt.Errorf(
				"event %d (%s) follows a different history than the one recorded: "+
					"an event before it was changed, removed or reordered",
				i+1, e.Kind)
		}
		if want := digest(e, prev); e.Hash != want {
			return fmt.Errorf("event %d (%s) has been altered since it was recorded",
				i+1, e.Kind)
		}
		prev = e.Hash
	}
	return nil
}
