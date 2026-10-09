// Package action turns what the platform worked out about a capture into
// something the person can actually do.
//
// This is the difference between a civic app and this one. Knowing that a heap
// of rubbish sits in R/Central and belongs to the Assistant Engineer (SWM) is
// an observation. Knowing which desk to tell, in words that will not be
// refused, and what obliges them once told, is an action.
//
// Nothing here files anything. Hard rule 1 is absolute: the platform generates
// a draft and a person sends it. There is no code path in this package that
// talks to an authority, and the copy never implies there is one.
package action

import (
	"fmt"
	"strings"
)

// Facts is everything already established about a capture. Nothing in this
// package discovers anything; it only arranges what is known.
type Facts struct {
	Category    string
	Subcategory string
	Ward        string
	Authority   string
	Where       string
	Lat, Lon    float64

	// The desk. Empty when no department has been recorded for this ward and
	// category, which is a real and common state.
	Department string
	Officer    string
	Office     string
	// How to reach the desk, as the authority publishes it: an office
	// landline, a role mailbox, the hours it is open and the hours a member of
	// the public may walk in. All attached to the post, none of it an
	// individual's. Empty where the handbook has not been read for them, and
	// empty is then shown as nothing rather than guessed at.
	OfficePhone   string
	OfficeEmail   string
	OfficeHours   string
	VisitingHours string
	// EscalatesTo is the office this one reports to, so the next rung is
	// known before it is needed.
	EscalatesTo string

	// DeadlineHours is zero when no deadline is published — which is the case
	// for water supply, whose handbook sets timelines for connections and none
	// for repairs.
	DeadlineHours    int
	DeadlineCitation string
}

// Channel is one way to reach the authority.
type Channel struct {
	Name string `json:"name"`
	How  string `json:"how"`
	URL  string `json:"url,omitempty"`
}

// Next is what to do about one capture.
type Next struct {
	Available        bool      `json:"available"`
	Text             string    `json:"text,omitempty"`
	Channels         []Channel `json:"channels,omitempty"`
	WhatHappensNext  string    `json:"what_happens_next"`
	DeadlineHours    int       `json:"deadline_hours,omitempty"`
	DeadlineCitation string    `json:"deadline_citation,omitempty"`
}

// Draft composes the complaint and what follows it.
func Draft(f Facts) Next {
	if strings.TrimSpace(f.Department) == "" {
		// No desk means no action. Saying so is better than inventing one:
		// a complaint sent to an office with no duty for this is a month lost.
		return Next{
			WhatHappensNext: "TraceSarkar does not yet have the department " +
				"responsible for this kind of problem in this ward, so there is " +
				"nothing it can tell you to send, and it will not guess at a desk.",
		}
	}

	n := Next{
		Available:        true,
		Channels:         append(officeChannel(f), channelsFor(f.Authority)...),
		DeadlineHours:    f.DeadlineHours,
		DeadlineCitation: f.DeadlineCitation,
	}
	n.Text = complaintText(f)
	n.WhatHappensNext = whatHappensNext(f)
	return n
}

// complaintText states what is there and asks for the duty to be performed.
// It does not characterise anyone's conduct — hard rule 3 — because a
// complaint that accuses is a complaint that gets argued with instead of
// actioned.
func complaintText(f Facts) string {
	what := strings.TrimSpace(f.Subcategory)
	if what == "" {
		what = strings.ReplaceAll(f.Category, "_", " ")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "To: %s\n", f.Officer)
	if f.Office != "" {
		fmt.Fprintf(&b, "%s\n", f.Office)
	}
	fmt.Fprintf(&b, "\nSubject: %s at %s, %s ward\n\n", capitalise(what), f.Where, f.Ward)

	fmt.Fprintf(&b, "There is %s at the location below, in %s ward.\n\n", what, f.Ward)
	fmt.Fprintf(&b, "Location: %s\n", f.Where)
	// The coordinates, because a street name does not find a heap of rubbish
	// and a crew sent to the wrong end of a road reports it as not found.
	fmt.Fprintf(&b, "Coordinates: %.5f, %.5f\n", f.Lat, f.Lon)
	b.WriteString("A photograph is attached.\n\n")

	b.WriteString("I request that it be attended to, and that I be informed of " +
		"the complaint reference number and the action taken.\n")

	return b.String()
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// whatHappensNext explains the clock. It runs from the moment the authority is
// told, never from the capture — photographing something here notifies nobody
// (D085) — so the wording is conditional until a reference number exists.
func whatHappensNext(f Facts) string {
	var b strings.Builder
	b.WriteString("Send this yourself through one of the channels above. " +
		"TraceSarkar does not send it for you and never will.\n\n")

	if f.DeadlineHours > 0 {
		fmt.Fprintf(&b, "Once %s has been told, it has %d hours to attend to this. ",
			f.Authority, f.DeadlineHours)
		b.WriteString("That clock starts from the time you lodge the complaint, " +
			"not from when the photograph was taken.\n\n")
	} else {
		fmt.Fprintf(&b, "There is no published deadline for this kind of problem — "+
			"%s sets timelines for other things and none for this one, so "+
			"TraceSarkar will not quote you a number it cannot source.\n\n",
			f.Authority)
	}

	if f.EscalatesTo != "" {
		fmt.Fprintf(&b, "If nothing happens, the office above this one is %s.\n\n",
			f.EscalatesTo)
	}

	b.WriteString("Keep the complaint reference number. It is what dates the " +
		"clock and what every later step quotes — an escalation without one is " +
		"an assertion, and with one it is a record.")
	return b.String()
}

// officeChannel is the desk's own published contact, listed first because it
// is the one addressed to the post responsible rather than to the corporation
// in general.
//
// Nothing is offered when nothing is published. A telephone number guessed at
// is worse than one absent: it is rung, nobody answers, and the citizen
// concludes the platform is wrong about everything else too.
func officeChannel(f Facts) []Channel {
	if f.OfficePhone == "" && f.OfficeEmail == "" {
		return nil
	}

	var parts []string
	if f.OfficePhone != "" {
		parts = append(parts, f.OfficePhone)
	}
	if f.OfficeEmail != "" {
		parts = append(parts, f.OfficeEmail)
	}
	if f.VisitingHours != "" {
		parts = append(parts, "in person "+f.VisitingHours)
	} else if f.OfficeHours != "" {
		parts = append(parts, "open "+f.OfficeHours)
	}

	return []Channel{{
		Name: "The office directly",
		How: strings.Join(parts, " · ") +
			". Published by the authority against this post, so it outlives " +
			"whoever currently holds it.",
	}}
}

// channelsFor lists the ways this authority accepts a complaint.
//
// Only channels the authority actually runs. The contractual deadline counts
// WhatsApp as valid intimation because BMC's own road tender says so, which is
// why a chat message is listed beside a portal rather than below it.
func channelsFor(authority string) []Channel {
	if !strings.EqualFold(strings.TrimSpace(authority), "BMC") {
		return nil
	}
	return []Channel{
		{
			Name: "MyBMC MARG",
			How:  "BMC's complaint system. Gives a reference number, which is the part that matters.",
			URL:  "https://portal.mcgm.gov.in/",
		},
		{
			Name: "1916",
			How:  "BMC's 24-hour helpline. Ask for the complaint number before you hang up.",
		},
		{
			Name: "WhatsApp",
			How:  "Send the photograph and the location. BMC's own road tender names WhatsApp as valid intimation.",
		},
	}
}
