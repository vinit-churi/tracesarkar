package polish

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// Fault is one reason a candidate may not be published. Line is the index of
// the offending line, or -1 for a fault about the whole post.
type Fault struct {
	Kind   string `json:"kind"`
	Line   int    `json:"line"`
	Detail string `json:"detail"`
}

func (f Fault) String() string {
	if f.Line < 0 {
		return fmt.Sprintf("%s: %s", f.Kind, f.Detail)
	}
	return fmt.Sprintf("line %d %s: %s", f.Line, f.Kind, f.Detail)
}

// A reworded sentence is longer than a template in some languages and shorter
// in others, so the cap is generous. What it stops is not a long sentence: it
// is a model that answered the prompt with a paragraph of its own opinion.
const (
	growthFactor = 2
	growthSlack  = 20
)

var digits = regexp.MustCompile(`[\p{Nd}]+`)

// Verify reports every reason the model's lines may not replace the baseline.
//
// An empty slice is the only result that publishes. There is no score and no
// threshold: a post either says the same facts as the post the platform can
// prove, or it is discarded and the deterministic text is served instead.
//
// The checks are deliberately all of the cheap, structural kind. Not one of
// them asks a model whether another model behaved, because that is the same
// trust problem one layer further down.
func Verify(f share.Facts, base share.Post, segs []Segment, got map[int]string) []Fault {
	var fs []Fault

	sent := map[int]Segment{}
	for _, s := range Prose(segs) {
		sent[s.Index] = s
	}

	// A line that was never sent is a model that has stopped following the
	// contract, and nothing else it returned can be trusted after that. The
	// permalink's own index is in here too: it is never sent, so an answer
	// for it is an answer to a question nobody asked.
	for _, i := range sortedKeys(got) {
		if _, ok := sent[i]; !ok {
			fs = append(fs, Fault{"line_unknown", i,
				"returned a line that was never sent"})
		}
	}

	for _, i := range sortedKeys(sent) {
		cand, ok := got[i]
		if !ok || strings.TrimSpace(cand) == "" {
			// Two registers in one message look like two people wrote it, so
			// a half-polished post is worse than an unpolished one.
			fs = append(fs, Fault{"line_missing", i, "no rewording returned"})
			continue
		}
		fs = append(fs, verifyLine(f, base.Lang, sent[i].Text, strings.TrimSpace(cand), i)...)
	}

	if base.Channel == share.X {
		if n := len([]rune(Reassemble(segs, got))); n > share.XLimit {
			fs = append(fs, Fault{"too_long_for_channel", -1,
				fmt.Sprintf("%d runes, the window is %d", n, share.XLimit)})
		}
	}
	return fs
}

func verifyLine(f share.Facts, lang share.Lang, was, now string, i int) []Fault {
	var fs []Fault

	// Every span that carries provenance survives byte for byte, as often as
	// it appeared. A reworded citation is not a citation, and a handle the
	// model dropped is a post addressed to nobody.
	for _, span := range protectedSpans(f) {
		before := strings.Count(was, span)
		if before == 0 {
			continue
		}
		if after := strings.Count(now, span); after != before {
			fs = append(fs, Fault{"span_lost", i, fmt.Sprintf(
				"%q appeared %d time(s) and now appears %d", span, before, after)})
		}
	}

	// Numbers are the obligation. 48 hours is a court order; 72 is a lie, and
	// a dropped count is a fact the reader no longer has.
	if a, b := digits.FindAllString(was, -1), digits.FindAllString(now, -1); !sameMultiset(a, b) {
		fs = append(fs, Fault{"number_changed", i,
			fmt.Sprintf("numbers were %v and are now %v", a, b)})
	}

	// A handle, hashtag or link the platform did not put there is a fact
	// about a named party with no source behind it (hard rule 2).
	was2, now2 := addressable(was), addressable(now)
	for _, tok := range sortedStrings(now2) {
		if !was2[tok] {
			fs = append(fs, Fault{"token_invented", i,
				fmt.Sprintf("%q is not in the line that was sent", tok)})
		}
	}

	// A line answered in the wrong script is a model that answered the wrong
	// prompt, and a half-English Marathi post reads as broken.
	switch lang {
	case share.EN:
		if hasDevanagari(now) {
			fs = append(fs, Fault{"script_lost", i,
				"English copy came back carrying Devanagari"})
		}
	default:
		if hasDevanagari(was) && !hasDevanagari(now) {
			fs = append(fs, Fault{"script_lost", i,
				fmt.Sprintf("%s copy came back with no Devanagari", lang)})
		}
	}

	if cap := growthFactor*len([]rune(was)) + growthSlack; len([]rune(now)) > cap {
		fs = append(fs, Fault{"line_too_long", i, fmt.Sprintf(
			"%d runes against a cap of %d", len([]rune(now)), cap)})
	}

	// Hard rule 3, enforced on the model's words with the same list that runs
	// on the platform's own.
	for _, v := range share.Check(now) {
		fs = append(fs, Fault{"tone", i, v.String()})
	}
	return fs
}

// protectedSpans is every string the model may not alter, taken from the fact
// set rather than parsed out of the text. Parsing would mean guessing which
// run of characters was a citation, and a guess is what this package exists
// to avoid.
func protectedSpans(f share.Facts) []string {
	var out []string
	add := func(s string) {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	add(f.DeadlineCitation)
	add(f.DeadlineShortRef)
	add(f.Authority)
	add(f.Ward)
	add(f.Permalink)
	for _, h := range f.Handles {
		add(h.Handle)
	}
	// Longest first, so a fault names the citation rather than the year
	// inside it.
	sort.Slice(out, func(i, j int) bool { return len(out[i]) > len(out[j]) })
	return out
}

// addressable collects the tokens that resolve to somebody or something
// outside the post: an account, a searchable tag, a link.
func addressable(s string) map[string]bool {
	out := map[string]bool{}
	for _, tok := range strings.Fields(s) {
		t := strings.Trim(tok, ".,;:!?()[]{}\"'—–।")
		switch {
		case strings.HasPrefix(t, "@"), strings.HasPrefix(t, "#"),
			strings.HasPrefix(t, "http://"), strings.HasPrefix(t, "https://"),
			strings.Contains(t, "@") && strings.Contains(t, "."):
			out[t] = true
		}
	}
	return out
}

func hasDevanagari(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Devanagari, r) {
			return true
		}
	}
	return false
}

func sameMultiset(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func sortedKeys[V any](m map[int]V) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func sortedStrings(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
