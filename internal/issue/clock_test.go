package issue

import (
	"testing"
	"time"
)

// D085, the rule that is easiest to get wrong and most damaging when it is:
// the clock runs from the moment the authority is told, never from the
// photograph. Photographing something on this platform notifies nobody.
func TestTheClock(t *testing.T) {
	reported := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	notified := reported.Add(3 * time.Hour)

	tests := []struct {
		name     string
		notified *time.Time
		hours    int
		now      time.Time
		state    ClockState
		left     time.Duration
	}{
		{
			// Nobody has been told, so nothing is running. The commonest
			// state, and the one a citizen most needs stated plainly.
			name: "not yet lodged", notified: nil, hours: 48,
			now: reported.Add(10 * time.Hour), state: NotStarted,
		},
		{
			name: "lodged and running", notified: &notified, hours: 48,
			now:   notified.Add(10 * time.Hour),
			state: Running, left: 38 * time.Hour,
		},
		{
			name: "lodged and elapsed", notified: &notified, hours: 48,
			now:   notified.Add(50 * time.Hour),
			state: Elapsed,
		},
		{
			// Water supply publishes no repair deadline. An obligation with
			// no number is not an obligation with a default.
			name: "no deadline published", notified: &notified, hours: 0,
			now: notified.Add(100 * time.Hour), state: NoDeadline,
		},
		{
			// The boundary belongs to the citizen: at exactly the deadline
			// it has not yet been breached.
			name: "exactly at the deadline", notified: &notified, hours: 48,
			now:   notified.Add(48 * time.Hour),
			state: Running, left: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClock(tt.notified, tt.hours).At(tt.now)
			if c.State != tt.state {
				t.Fatalf("state = %q, want %q", c.State, tt.state)
			}
			if tt.state == Running && c.Left != tt.left {
				t.Errorf("left = %v, want %v", c.Left, tt.left)
			}
			if c.Says == "" {
				t.Error("the clock says nothing")
			}
		})
	}
}

// A clock that has not started must not imply one has, and must say what
// would start it.
func TestAnUnstartedClockSaysWhatWouldStartIt(t *testing.T) {
	c := NewClock(nil, 48).At(time.Now())
	says := c.Says
	for _, want := range []string{"48", "lodge"} {
		if !contains(says, want) {
			t.Errorf("the clock does not mention %q: %q", want, says)
		}
	}
	for _, forbidden := range []string{"remaining", "left", "overdue"} {
		if contains(says, forbidden) {
			t.Errorf("an unstarted clock implies it is running: %q", says)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if equalFold(s[i:i+len(sub)], sub) {
				return true
			}
		}
		return false
	})()
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		x, y := a[i], b[i]
		if 'A' <= x && x <= 'Z' {
			x += 32
		}
		if 'A' <= y && y <= 'Z' {
			y += 32
		}
		if x != y {
			return false
		}
	}
	return true
}
