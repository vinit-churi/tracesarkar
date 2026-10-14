package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Record is everything the pack states about one issue.
//
// Every pointer field is nil when the platform does not hold that fact, and
// nil is printed as a stated absence rather than a blank: "no contract is
// attributed to this location" is a fact a reader can act on, an empty row is
// a reader wondering whether the document is broken.
type Record struct {
	IssueID     string
	Category    string
	Subcategory string
	Status      string
	Sightings   int

	Authority string
	Ward      string
	WardBasis string

	// Lat and Lon are exact. This document is for the person who reported the
	// problem, to attach to their own complaint, and a complaint needs the
	// place. It is labelled as exact wherever it appears so it is never
	// mistaken for the coarsened cell the public page publishes.
	Lat, Lon  float64
	AccuracyM float64

	FirstSeen time.Time
	LastSeen  time.Time

	LodgedOn  *time.Time
	Reference string

	Desk *Desk

	DeadlineHours    int
	DeadlineCitation string

	Contract *Contract

	Events []Event
}

// Desk is the office the complaint belongs to.
type Desk struct {
	Department  string
	Officer     string
	Address     string
	Phone       string
	Email       string
	SourceRef   string
	RetrievedAt time.Time
}

// Contract is a published contract attribution.
type Contract struct {
	WorkCode    string
	Contractor  string
	Confidence  string
	DistanceM   float64
	Basis       string
	SourceRef   string
	RetrievedAt time.Time
}

// Event is one entry in the hash-chained history.
type Event struct {
	Seq    int
	Kind   string
	At     time.Time
	Detail string
	Hash   string
}

// Photo is one capture, already redacted.
//
// Redacted and not the original: this pack exists to be handed to somebody
// else, which is exactly the case redaction is for. The archival original
// stays break-glass.
type Photo struct {
	JPEG          []byte
	Width, Height int
	TakenAt       time.Time
	AccuracyM     float64
	Redaction     string
}

func day(t time.Time) string { return t.UTC().Format("2 January 2006") }

// stop ends a sentence without doubling a full stop that is already there.
// The basis strings come from other packages and some of them punctuate.
func stop(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasSuffix(s, ".") || strings.HasSuffix(s, "?") ||
		strings.HasSuffix(s, "!") {
		return s
	}
	return s + "."
}

func moment(t time.Time) string { return t.UTC().Format("2 January 2006, 15:04") + " UTC" }

// Stamp is a digest of the facts in the record.
//
// It is printed on the pack so that two people can agree they are holding the
// same record, and it is computed from the facts rather than from the rendered
// bytes so that a layout change does not invalidate every stamp ever quoted.
func Stamp(r Record) string {
	h := sha256.New()
	for _, line := range Lines(r) {
		h.Write([]byte(line))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// block is one labelled section of the document.
type block struct {
	heading string
	lines   []line
}

type line struct {
	text  string
	style face
	gap   bool
}

func body(s string) line { return line{text: s, style: bodyFace} }
func bold(s string) line { return line{text: s, style: boldFace} }
func space() line        { return line{gap: true} }

// Lines is the whole document as plain text, in reading order.
//
// It exists so that what the pack says can be tested, hashed and read aloud
// without rendering a single glyph — the tone check and the stamp both run
// over this, not over the PDF.
func Lines(r Record) []string {
	var out []string
	for _, b := range blocks(r) {
		if b.heading != "" {
			out = append(out, b.heading)
		}
		for _, l := range b.lines {
			if !l.gap && strings.TrimSpace(l.text) != "" {
				out = append(out, l.text)
			}
		}
	}
	return out
}

func blocks(r Record) []block {
	return []block{
		cover(r),
		where(r),
		whose(r),
		obligation(r),
		attribution(r),
		history(r),
		sources(r),
		disclaimer(),
	}
}

func cover(r Record) block {
	thing := strings.TrimSpace(r.Subcategory)
	if thing == "" {
		thing = strings.ReplaceAll(r.Category, "_", " ")
	}
	b := block{heading: "The issue"}
	b.lines = append(b.lines,
		bold(strings.ToUpper(thing[:1])+thing[1:]),
		body("Issue "+r.IssueID),
		body("Status: "+statusWords(r.Status)),
	)
	if r.Sightings > 1 {
		b.lines = append(b.lines, body(fmt.Sprintf(
			"Seen %d times. One problem reported more than once, not %d separate complaints.",
			r.Sightings, r.Sightings)))
	}
	if !r.FirstSeen.IsZero() {
		span := "First photographed " + day(r.FirstSeen)
		if !r.LastSeen.IsZero() && !r.LastSeen.Equal(r.FirstSeen) {
			span += ", most recently " + day(r.LastSeen)
		}
		b.lines = append(b.lines, body(span+"."))
	}
	return b
}

// statusWords never lets an authority's claim and a resident's photograph read
// the same (G14).
func statusWords(s string) string {
	switch s {
	case "open":
		return "open"
	case "claimed_resolved":
		return "the authority has recorded this as attended; no resident has photographed it fixed"
	case "citizen_confirmed":
		return "a resident has photographed it fixed"
	case "reopened":
		return "seen again after being recorded as attended"
	}
	return "not published"
}

func where(r Record) block {
	b := block{heading: "Where"}
	if r.Lat == 0 && r.Lon == 0 {
		b.lines = append(b.lines, body("No position was recorded with this report."))
		return b
	}
	b.lines = append(b.lines,
		bold(fmt.Sprintf("%.7f, %.7f", r.Lat, r.Lon)),
		body("This is the position the reporting device recorded, to the precision it "+
			"recorded it. It is not the coarsened cell the public page publishes."),
	)
	if r.AccuracyM > 0 {
		b.lines = append(b.lines, body(fmt.Sprintf(
			"The device reported an accuracy of about %.0f m at the moment of capture.", r.AccuracyM)))
	}
	return b
}

func whose(r Record) block {
	b := block{heading: "Whose it is"}
	if r.Ward != "" || r.Authority != "" {
		b.lines = append(b.lines, bold(strings.TrimSpace(r.Authority+" · "+r.Ward+" ward")))
	}
	if r.WardBasis != "" {
		b.lines = append(b.lines, body(stop(r.WardBasis)))
	}
	// Hard rule 2: a desk is a named office, and it is named only with the
	// document that says so.
	if r.Desk == nil || strings.TrimSpace(r.Desk.SourceRef) == "" {
		b.lines = append(b.lines, space(), body(
			"No department is recorded for this category in this ward, so this pack "+
				"does not name one. A complaint addressed to a desk that cannot act "+
				"on it is returned."))
		return b
	}
	d := r.Desk
	b.lines = append(b.lines, space(), bold(d.Department))
	if d.Officer != "" {
		b.lines = append(b.lines, body("Addressed to: "+d.Officer+
			" (a post, not a person — the post outlives whoever holds it)."))
	}
	for _, field := range [][2]string{
		{"Office", d.Address}, {"Telephone", d.Phone}, {"Email", d.Email},
	} {
		if strings.TrimSpace(field[1]) != "" {
			b.lines = append(b.lines, body(field[0]+": "+field[1]))
		}
	}
	b.lines = append(b.lines, body("Source: "+d.SourceRef+
		", retrieved "+day(d.RetrievedAt)+"."))
	return b
}

func obligation(r Record) block {
	b := block{heading: "What binds it, and when the clock started"}
	if r.LodgedOn == nil {
		b.lines = append(b.lines, body(
			"Nobody has recorded telling the authority about this, so no deadline is "+
				"running. Photographing a problem notifies nobody: a deadline begins "+
				"when a complaint is lodged."))
	} else {
		b.lines = append(b.lines, bold("The authority was told on "+day(*r.LodgedOn)+"."))
		if strings.TrimSpace(r.Reference) != "" {
			b.lines = append(b.lines, body("Reference given by the authority: "+r.Reference))
		}
	}
	// A deadline with no citation is an assertion, and this platform does not
	// make those — least of all in a document meant for a filing.
	if r.DeadlineHours > 0 && strings.TrimSpace(r.DeadlineCitation) != "" {
		b.lines = append(b.lines, space(),
			bold(strconv.Itoa(r.DeadlineHours)+" hours to attend it."),
			body(r.DeadlineCitation))
	} else {
		b.lines = append(b.lines, space(), body(
			"No deadline is published for this kind of problem in a document this "+
				"platform holds, so none is stated here."))
	}
	return b
}

func attribution(r Record) block {
	b := block{heading: "The contract covering this location"}
	c := r.Contract
	if c == nil || strings.TrimSpace(c.SourceRef) == "" {
		b.lines = append(b.lines, body(
			"No contract is attributed to this location. Either no published work "+
				"covers it, or the match was not strong enough to publish. A "+
				"contractor named on a weak match is the error this platform is "+
				"built to avoid, so nothing is named here."))
		return b
	}
	if c.WorkCode != "" {
		b.lines = append(b.lines, bold("Work "+c.WorkCode))
	}
	if c.Contractor != "" {
		b.lines = append(b.lines, body("Contractor of record: "+c.Contractor))
	}
	if c.Basis != "" {
		b.lines = append(b.lines, body(stop(c.Basis)))
	}
	if c.DistanceM > 0 {
		b.lines = append(b.lines, body(fmt.Sprintf(
			"Distance from the reported point to this work: %.1f m.", c.DistanceM)))
	}
	if c.Confidence != "" {
		b.lines = append(b.lines, body("Match confidence, as a band rather than a score: "+
			strings.ReplaceAll(c.Confidence, "_", " ")+"."))
	}
	b.lines = append(b.lines, body("Source: "+c.SourceRef+", retrieved "+day(c.RetrievedAt)+"."))
	return b
}

func history(r Record) block {
	b := block{heading: "History"}
	if len(r.Events) == 0 {
		b.lines = append(b.lines, body("No events are recorded against this issue."))
		return b
	}
	for _, e := range r.Events {
		b.lines = append(b.lines, bold(fmt.Sprintf("%d. %s — %s",
			e.Seq, strings.ReplaceAll(e.Kind, "_", " "), moment(e.At))))
		if strings.TrimSpace(e.Detail) != "" {
			b.lines = append(b.lines, body(e.Detail))
		}
		if strings.TrimSpace(e.Hash) != "" {
			b.lines = append(b.lines, body("chain "+e.Hash))
		}
	}
	b.lines = append(b.lines, space(), body(
		"Each entry carries the hash of the one before it, so an entry changed, "+
			"removed or reordered after the fact is detectable from the end of the "+
			"chain. The chain is never recomputed when an entry is appended."))
	return b
}

func sources(r Record) block {
	b := block{heading: "Where each fact came from"}
	var refs []string
	if r.Desk != nil && strings.TrimSpace(r.Desk.SourceRef) != "" {
		refs = append(refs, "The department and the office it sits in: "+
			r.Desk.SourceRef+", retrieved "+day(r.Desk.RetrievedAt)+".")
	}
	if r.DeadlineHours > 0 && strings.TrimSpace(r.DeadlineCitation) != "" {
		refs = append(refs, "The "+strconv.Itoa(r.DeadlineHours)+
			"-hour obligation: "+r.DeadlineCitation)
	}
	if r.Contract != nil && strings.TrimSpace(r.Contract.SourceRef) != "" {
		refs = append(refs, "The contract covering this location: "+
			r.Contract.SourceRef+", retrieved "+day(r.Contract.RetrievedAt)+".")
	}
	if len(refs) == 0 {
		b.lines = append(b.lines, body(
			"This pack states no fact about a named party, so there is no source to list."))
		return b
	}
	for _, s := range refs {
		b.lines = append(b.lines, body(s))
	}
	return b
}

func disclaimer() block {
	return block{
		heading: "What this document is",
		lines: []line{
			body("This is a compiled record of what one platform holds about one reported " +
				"problem. It is not legal advice, and it is not a finding about anybody's " +
				"conduct."),
			body("It states facts and the joins between them — this location, this ward, " +
				"this department, this obligation — and it does not conclude that anybody " +
				"has done anything wrong."),
			body("Nothing in this document was filed with any authority by the platform. " +
				"Every complaint, application and appeal is submitted by the person who " +
				"chooses to submit it."),
			body("The same issue always compiles to the same document, so the stamp on the " +
				"first page identifies the record rather than the moment it was printed."),
		},
	}
}
