package polish

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// The measurement ADR 0020 asked for before any model could be let near this
// surface: "Revisit only with an eval set that measures invented facts, not
// tone."
//
// So that is what this measures, and nothing else. There is no score for
// whether the Marathi reads better — that judgement belongs to a native
// speaker and is not something the platform can award itself. What it can
// measure is how often the model, given sentences whose every fact is
// provable, hands back sentences that are not the same facts.

// Case is one post the polisher is measured on.
type Case struct {
	Name  string
	Facts share.Facts
	Lang  share.Lang
	Chan  share.Channel
}

// Refusal is one case the guard rejected, kept whole so a failure can be read
// rather than inferred from a count.
type Refusal struct {
	Case     string         `json:"case"`
	Lang     share.Lang     `json:"lang"`
	Baseline string         `json:"baseline"`
	Faults   []Fault        `json:"faults"`
	Got      map[int]string `json:"got"`
}

// Accepted is one case the guard passed and the model actually changed.
//
// Kept because the harness deliberately refuses to score whether the Marathi
// reads better — that judgement belongs to a native speaker — and a
// measurement that reports only refusals says nothing about what the model
// produced when it worked. These are what that person reads.
//
// Unchanged posts are not kept: "the model returned the template" is a
// number, not something worth anybody's reading time.
type Accepted struct {
	Case     string     `json:"case"`
	Lang     share.Lang `json:"lang"`
	Baseline string     `json:"baseline"`
	Polished string     `json:"polished"`
}

// Score is the measurement.
type Score struct {
	Posts     int            `json:"posts"`
	Published int            `json:"published"`
	Refused   int            `json:"refused"`
	Unchanged int            `json:"unchanged"`
	Errors    int            `json:"errors"`
	ByKind    map[string]int `json:"by_kind"`
	// Accepted holds every published post the model actually reworded.
	Accepted []Accepted `json:"accepted"`
}

// factKinds are the faults that mean the model changed what the post claims,
// as against the faults that mean it misbehaved in some other way. The
// distinction is the one ADR 0020 drew: a long line or a lost script is a
// quality problem, a changed number is the thing that must never ship.
var factKinds = map[string]bool{
	"span_lost":      true,
	"number_changed": true,
	"token_invented": true,
	"tone":           true,
}

// InventedFacts counts the posts refused because the model changed a claim.
func (s Score) InventedFacts() int {
	n := 0
	for kind, c := range s.ByKind {
		if factKinds[kind] {
			n += c
		}
	}
	return n
}

func (s Score) String() string {
	kinds := make([]string, 0, len(s.ByKind))
	for k := range s.ByKind {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	out := fmt.Sprintf("%d posts: %d published (%d reworded, %d returned unchanged), "+
		"%d refused, %d of those a provider failure",
		s.Posts, s.Published, len(s.Accepted), s.Unchanged, s.Refused, s.Errors)
	for _, k := range kinds {
		out += fmt.Sprintf("\n  %-22s %d", k, s.ByKind[k])
	}
	// The number the ADR asked for, named rather than left to be worked out
	// from the list above.
	out += fmt.Sprintf("\n  %-22s %d", "invented a fact", s.InventedFacts())
	return out
}

// MeasureConcurrency is how many cases are in flight at once.
//
// Chosen for the measurement, not for production: a provider worth evaluating
// is one a page load will call once, and the eval calls it ninety-six times.
// The first real run of this harness was sequential and spent its whole
// window on 21 of 96 cases against a model at two minutes a call, reporting
// almost nothing about fidelity because it never got to ask. Low enough not
// to look like an attack on the provider.
const MeasureConcurrency = 6

// Measure runs every case and reports what the guard decided.
//
// Cases are independent — each composes its own baseline and asks its own
// question — so they run concurrently. The result is assembled in the
// corpus's order afterwards, because a measurement that reads differently
// between runs cannot be diffed.
func Measure(ctx context.Context, p Polisher, cases []Case) (Score, []Refusal, error) {
	type outcome struct {
		base share.Post
		res  Result
		err  error
	}
	out := make([]outcome, len(cases))

	var wg sync.WaitGroup
	gate := make(chan struct{}, MeasureConcurrency)
	for i, c := range cases {
		wg.Add(1)
		go func(i int, c Case) {
			defer wg.Done()
			gate <- struct{}{}
			defer func() { <-gate }()

			base, err := share.Compose(c.Facts, c.Lang, c.Chan)
			if err != nil {
				out[i] = outcome{err: fmt.Errorf("compose %s/%s/%s: %w",
					c.Name, c.Lang, c.Chan, err)}
				return
			}
			out[i] = outcome{base: base, res: Apply(ctx, p, c.Facts, base)}
		}(i, c)
	}
	wg.Wait()

	s := Score{ByKind: map[string]int{}}
	var refusals []Refusal
	for i, c := range cases {
		o := out[i]
		if o.err != nil {
			return s, refusals, o.err
		}
		s.Posts++
		name := c.Name + " " + string(c.Chan)

		if len(o.res.Faults) == 0 {
			s.Published++
			if o.res.Post.Text == o.base.Text {
				s.Unchanged++
				continue
			}
			s.Accepted = append(s.Accepted, Accepted{
				Case: name, Lang: c.Lang,
				Baseline: o.base.Text, Polished: o.res.Post.Text,
			})
			continue
		}

		s.Refused++
		for _, f := range o.res.Faults {
			s.ByKind[f.Kind]++
			if f.Kind == "model_error" {
				s.Errors++
			}
		}
		refusals = append(refusals, Refusal{
			Case: name, Lang: c.Lang,
			Baseline: o.base.Text, Faults: o.res.Faults, Got: o.res.Got,
		})
	}
	return s, refusals, nil
}

// Corpus is the fact sets the polisher is measured on: every category, both
// channels, all three languages, and the states that change which sentences
// a post carries at all — lodged and not, a deadline running, elapsed and
// absent, a handle present and missing, one sighting and several.
//
// It is built here rather than read from the database on purpose. A
// measurement that depends on what happens to be in production cannot be
// compared between runs, and the cases that matter most — an unlodged issue,
// a missing citation — are the rare ones.
func Corpus() []Case {
	seen := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	lodged := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	handle := share.Handle{
		Platform: "x", Handle: "@mybmc", Kind: share.Institutional,
		SourceRef: "https://portal.mcgm.gov.in/", RetrievedAt: lodged, FirstHand: true,
	}
	escalation := share.Handle{
		Platform: "x", Handle: "@CMOMaharashtra", Kind: share.Escalation,
		SourceRef: "https://portal.mcgm.gov.in/", RetrievedAt: lodged, FirstHand: false,
	}

	base := func(name, cat, sub, ward, authority string) share.Facts {
		return share.Facts{
			IssueID:   "018f0000-0000-7000-8000-00000000000" + name[:1],
			Permalink: "https://tracesarkar.org/i/" + name,
			Authority: authority, Ward: ward,
			Category: cat, Subcategory: sub,
			FirstSeen: seen, Sightings: 3, Status: "open",
			Lodged: true, LodgedOn: &lodged,
			DeadlineHours: 48, DeadlineState: "elapsed",
			DeadlineShortRef: "HC order",
			DeadlineCitation: "Bombay High Court order in PIL 85 of 2013",
			Handles:          []share.Handle{handle},
		}
	}

	var shapes []struct {
		name string
		f    share.Facts
	}
	add := func(name string, f share.Facts) {
		shapes = append(shapes, struct {
			name string
			f    share.Facts
		}{name, f})
	}

	add("pothole", base("pothole", "road_defect", "pothole", "K/E", "BMC"))
	add("manhole", base("manhole", "road_defect", "open manhole", "A", "BMC"))
	add("garbage", base("garbage", "waste", "uncollected garbage", "R/C", "BMC"))
	add("waterlogging", base("waterlogging", "water_drainage", "waterlogging", "F/N", "BMC"))
	add("streetlight", base("streetlight", "street_furniture", "street light not working", "T", "BMC"))
	add("bridge", base("bridge", "structural", "cracked bridge/FOB", "G/S", "BMC"))

	// A state board, to check that an authority which is not the corporation
	// survives the rewording intact.
	add("effluent", base("effluent", "environment", "effluent discharge", "L", "MPCB"))

	noClock := base("tree", "environment", "tree felling", "H/W", "BMC")
	noClock.DeadlineHours, noClock.DeadlineShortRef, noClock.DeadlineCitation = 0, "", ""
	noClock.DeadlineState = "no_deadline"
	add("tree-no-clock", noClock)

	unlodged := base("unlodged", "road_defect", "pothole", "P/N", "BMC")
	unlodged.Lodged, unlodged.LodgedOn = false, nil
	unlodged.DeadlineState = "not_started"
	add("never-told", unlodged)

	running := base("running", "road_defect", "crack", "N", "BMC")
	running.DeadlineState = "running"
	add("clock-running", running)

	single := base("single", "waste", "illegal dumping", "M/E", "BMC")
	single.Sightings = 1
	add("one-sighting", single)

	noHandle := base("nohandle", "road_defect", "subsidence", "S", "BMC")
	noHandle.Handles = nil
	add("no-handle", noHandle)

	escalated := base("escalated", "road_defect", "pothole", "C", "BMC")
	escalated.Handles = []share.Handle{handle, escalation}
	add("escalation-unlocked", escalated)

	for _, st := range []string{"claimed_resolved", "citizen_confirmed", "reopened"} {
		f := base(st, "road_defect", "pothole", "D", "BMC")
		f.Status = st
		add("status-"+st, f)
	}

	var out []Case
	for _, sh := range shapes {
		for _, lang := range share.Languages() {
			for _, ch := range share.Channels() {
				out = append(out, Case{
					Name: sh.name, Facts: sh.f, Lang: lang, Chan: ch,
				})
			}
		}
	}
	return out
}
