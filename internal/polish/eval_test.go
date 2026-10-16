package polish

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// fn is a Polisher built from a function, for cases where the answer has to
// depend on what was sent.
type fn struct {
	f    func([]Segment) map[int]string
	name string
}

func (p *fn) Reword(_ context.Context, _ share.Lang, _ share.Channel, lines []Segment) (map[int]string, error) {
	return p.f(lines), nil
}
func (p *fn) Describe() (string, string) { return p.name, "test" }

// Every case in the corpus has to compose, or the measurement silently skips
// the states that matter most.
func TestTheCorpusComposesInEveryLanguageAndChannel(t *testing.T) {
	cases := Corpus()
	if len(cases) == 0 {
		t.Fatal("the corpus is empty")
	}
	langs, chans, cats := map[share.Lang]int{}, map[share.Channel]int{}, map[string]int{}
	for _, c := range cases {
		if _, err := share.Compose(c.Facts, c.Lang, c.Chan); err != nil {
			t.Errorf("%s/%s/%s does not compose: %v", c.Name, c.Lang, c.Chan, err)
		}
		langs[c.Lang]++
		chans[c.Chan]++
		cats[c.Facts.Category]++
	}
	for _, l := range share.Languages() {
		if langs[l] == 0 {
			t.Errorf("the corpus has no %s cases", l)
		}
	}
	for _, ch := range share.Channels() {
		if chans[ch] == 0 {
			t.Errorf("the corpus has no %s cases", ch)
		}
	}
	// All six routable categories, so a measurement is not taken on potholes
	// alone.
	for _, cat := range []string{"road_defect", "waste", "water_drainage",
		"street_furniture", "structural", "environment"} {
		if cats[cat] == 0 {
			t.Errorf("the corpus has no %s cases", cat)
		}
	}
}

// The corpus has to contain the states that remove a sentence, because those
// are where a model is most likely to put one back.
func TestTheCorpusCoversThePostsThatSayLess(t *testing.T) {
	var noClock, unlodged, noHandle, single int
	for _, c := range Corpus() {
		switch {
		case c.Facts.DeadlineCitation == "" && c.Facts.DeadlineShortRef == "":
			noClock++
		case !c.Facts.Lodged:
			unlodged++
		}
		if len(c.Facts.Handles) == 0 {
			noHandle++
		}
		if c.Facts.Sightings <= 1 {
			single++
		}
	}
	for name, n := range map[string]int{
		"no deadline": noClock, "never told": unlodged,
		"no handle": noHandle, "one sighting": single,
	} {
		if n == 0 {
			t.Errorf("the corpus has no %q case", name)
		}
	}
}

func TestMeasurePublishesAnEchoAndCountsItUnchanged(t *testing.T) {
	echoAll := &fn{name: "echo", f: func(lines []Segment) map[int]string {
		out := map[int]string{}
		for _, l := range lines {
			out[l.Index] = l.Text
		}
		return out
	}}
	cases := Corpus()
	s, refusals, err := Measure(context.Background(), echoAll, cases)
	if err != nil {
		t.Fatalf("Measure: %v", err)
	}
	if s.Posts != len(cases) || s.Published != len(cases) {
		t.Errorf("%d of %d published", s.Published, s.Posts)
	}
	if s.Unchanged != len(cases) {
		t.Errorf("%d unchanged, want all %d", s.Unchanged, len(cases))
	}
	if len(refusals) != 0 || s.InventedFacts() != 0 {
		t.Errorf("an echo was refused: %v", refusals)
	}
}

// The measurement the ADR asked for: a model that changes a number is caught
// on every post, and the count is attributed to an invented fact rather than
// to tone.
func TestMeasureCountsAnInventedFactOnEveryPost(t *testing.T) {
	bump := regexp.MustCompile(`\d+`)
	liar := &fn{name: "liar", f: func(lines []Segment) map[int]string {
		out := map[int]string{}
		for _, l := range lines {
			out[l.Index] = bump.ReplaceAllString(l.Text, "999")
		}
		return out
	}}
	cases := Corpus()
	s, refusals, err := Measure(context.Background(), liar, cases)
	if err != nil {
		t.Fatalf("Measure: %v", err)
	}
	if s.Published != 0 {
		t.Errorf("%d posts with invented numbers were published", s.Published)
	}
	if s.ByKind["number_changed"] == 0 {
		t.Errorf("no number_changed faults recorded: %v", s.ByKind)
	}
	if s.InventedFacts() == 0 {
		t.Error("InventedFacts is zero on a corpus of invented facts")
	}
	// A refusal has to be readable, not just counted.
	if len(refusals) == 0 || refusals[0].Baseline == "" || len(refusals[0].Faults) == 0 {
		t.Error("a refusal was recorded without enough to read it")
	}
	if len(refusals[0].Got) == 0 {
		t.Error("a refusal did not keep what the model actually said")
	}
}

// A model that adds a sentence of its own is the failure ADR 0020 named. It
// has to be caught by structure, not by the word list.
func TestMeasureCatchesAModelThatAddsAClaim(t *testing.T) {
	windy := &fn{name: "windy", f: func(lines []Segment) map[int]string {
		out := map[int]string{}
		for _, l := range lines {
			out[l.Index] = l.Text
		}
		for _, l := range lines {
			out[l.Index] = l.Text + " This has been ignored for 6 months by the ward office."
			break
		}
		return out
	}}
	s, _, err := Measure(context.Background(), windy, Corpus())
	if err != nil {
		t.Fatalf("Measure: %v", err)
	}
	if s.Published != 0 {
		t.Errorf("%d posts carrying an added claim were published", s.Published)
	}
}

// Score has to read as a report, because that is how it reaches a person.
func TestScoreReadsAsAReport(t *testing.T) {
	s := Score{Posts: 10, Published: 8, Refused: 2, Unchanged: 1,
		ByKind: map[string]int{"number_changed": 1, "line_too_long": 1}}
	out := s.String()
	for _, want := range []string{"10 posts", "8 published", "2 refused",
		"number_changed", "line_too_long"} {
		if !strings.Contains(out, want) {
			t.Errorf("the report does not mention %q:\n%s", want, out)
		}
	}
}

// A slow provider must not make the measurement unusable. The first real run
// of this harness spent a 40-minute window on 21 of 96 cases against a
// reasoning model at roughly two minutes a call, and reported almost nothing
// about the model's fidelity because it never got to ask.
func TestMeasureRunsCasesConcurrently(t *testing.T) {
	const delay = 60 * time.Millisecond
	slow := &fn{name: "slow", f: func(lines []Segment) map[int]string {
		time.Sleep(delay)
		out := map[int]string{}
		for _, l := range lines {
			out[l.Index] = l.Text
		}
		return out
	}}
	cases := Corpus()
	start := time.Now()
	s, _, err := Measure(context.Background(), slow, cases)
	if err != nil {
		t.Fatalf("Measure: %v", err)
	}
	if s.Posts != len(cases) {
		t.Fatalf("measured %d of %d", s.Posts, len(cases))
	}
	// Sequentially this would be len(cases) * delay. Half of that is a
	// generous bar and still impossible without concurrency.
	if took, sequential := time.Since(start), time.Duration(len(cases))*delay; took > sequential/2 {
		t.Errorf("%d cases at %v each took %v; sequential would be %v, so nothing ran in parallel",
			len(cases), delay, took, sequential)
	}
}

// The harness refuses to score whether the Marathi reads better — that
// judgement belongs to a native speaker. What it must do is hand that person
// the text: a measurement that reports only refusals says nothing about what
// the model actually produced when it worked.
func TestMeasureKeepsWhatTheModelActuallyWrote(t *testing.T) {
	reworder := &fn{name: "reworder", f: func(lines []Segment) map[int]string {
		out := map[int]string{}
		for _, l := range lines {
			out[l.Index] = l.Text
		}
		// Change exactly one line, in a way the guard has no reason to
		// refuse. What the change is does not matter here — only that there
		// is one, and that it survives Reassemble, which trims the ends.
		for _, l := range lines {
			if strings.Contains(l.Text, " ") && !strings.Contains(l.Text, "@") &&
				!digits.MatchString(l.Text) {
				out[l.Index] = strings.Replace(l.Text, " ", "  ", 1)
				break
			}
		}
		return out
	}}
	s, _, err := Measure(context.Background(), reworder, Corpus())
	if err != nil {
		t.Fatalf("Measure: %v", err)
	}
	if s.Published == 0 {
		t.Fatal("nothing was published, so there is nothing to read")
	}
	if len(s.Accepted) == 0 {
		t.Error("Measure kept no accepted post, so a native speaker has nothing to read")
	}
	for _, a := range s.Accepted {
		if a.Baseline == "" || a.Polished == "" {
			t.Errorf("%s/%s was kept without both texts", a.Case, a.Lang)
		}
		if a.Baseline == a.Polished {
			t.Errorf("%s/%s was kept although nothing changed; only changes are "+
				"worth a reader's time", a.Case, a.Lang)
		}
	}
}

// A post that exceeds X's window is rejected by X, and share.fit exists to
// prevent that — but fit stops at rankCore and never drops a core line, so
// raising a line's rank can push a post over the limit instead. That is
// exactly what making the resolution status non-droppable risked (D176), and
// it is the kind of thing one hand-written fact set will not catch.
//
// Asserted here rather than in internal/share because the corpus lives here:
// 16 fact sets across every category and status, which is what makes the
// check worth anything.
func TestEveryCorpusPostFitsItsChannel(t *testing.T) {
	for _, c := range Corpus() {
		p, err := share.Compose(c.Facts, c.Lang, c.Chan)
		if err != nil {
			t.Fatalf("compose %s/%s/%s: %v", c.Name, c.Lang, c.Chan, err)
		}
		if c.Chan != share.X {
			continue
		}
		if n := len([]rune(p.Text)); n > share.XLimit {
			t.Errorf("%s/%s is %d runes, over the %d window, having dropped %v:\n%s",
				c.Name, c.Lang, n, share.XLimit, p.Dropped, p.Text)
		}
		// And whatever it dropped, it never dropped the claim about how the
		// issue was resolved.
		for _, key := range p.Dropped {
			if key == "status" && c.Facts.Status != "open" && c.Facts.Status != "" {
				t.Errorf("%s/%s dropped the %q status line", c.Name, c.Lang, c.Facts.Status)
			}
		}
	}
}
