package polish

import (
	"context"
	"regexp"
	"strings"
	"testing"

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
