package issue

import (
	"fmt"
	"time"
)

// ClockState is where an obligation stands.
type ClockState string

const (
	// NotStarted: nobody has told the authority, so nothing is running. The
	// commonest state, and the one most easily misrepresented.
	NotStarted ClockState = "not_started"
	Running    ClockState = "running"
	Elapsed    ClockState = "elapsed"
	// NoDeadline: a desk is known and no deadline is published for it. An
	// obligation without a number is not an obligation with a default.
	NoDeadline ClockState = "no_deadline"
)

// Clock is an obligation and when it was triggered.
type Clock struct {
	NotifiedAt *time.Time
	Hours      int
}

// Reading is the clock at a moment.
type Reading struct {
	State ClockState
	Left  time.Duration
	Over  time.Duration
	Says  string
}

func NewClock(notifiedAt *time.Time, hours int) Clock {
	return Clock{NotifiedAt: notifiedAt, Hours: hours}
}

// At reads the clock.
//
// The rule it exists to enforce: the deadline runs from the moment the
// authority was told, never from the photograph. Photographing something here
// notifies nobody, and a platform that starts a legal clock on its own capture
// is telling the citizen something untrue about where they stand.
func (c Clock) At(now time.Time) Reading {
	if c.Hours <= 0 {
		return Reading{State: NoDeadline, Says: "No deadline is published for " +
			"this kind of problem, so there is no clock to run."}
	}
	if c.NotifiedAt == nil {
		return Reading{State: NotStarted, Says: fmt.Sprintf(
			"Nothing is running yet. The %d hours begin when you lodge the "+
				"complaint with the authority — photographing it here tells "+
				"them nothing.", c.Hours)}
	}

	due := c.NotifiedAt.Add(time.Duration(c.Hours) * time.Hour)
	if now.After(due) {
		over := now.Sub(due)
		return Reading{State: Elapsed, Over: over, Says: fmt.Sprintf(
			"The %d hours passed %s ago.", c.Hours, humanDuration(over))}
	}
	left := due.Sub(now)
	return Reading{State: Running, Left: left, Says: fmt.Sprintf(
		"%s left of the %d hours.", humanDuration(left), c.Hours)}
}

func humanDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "under a minute"
	case d < time.Hour:
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%d hours", int(d.Hours()))
	default:
		return fmt.Sprintf("%d days", int(d.Hours()/24))
	}
}
