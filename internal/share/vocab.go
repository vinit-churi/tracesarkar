package share

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// The vocabulary is data, not string literals in Go.
//
// docs/02-product/10-accessibility-and-vernacular.md requires that copy in a
// language be reviewed by a native speaker before release, and that machine
// translation never reach legal text. Neither is enforceable against literals
// scattered through a package: a reviewer cannot read Go to check Marathi. So
// each language is one file with its own review state, and the file is what a
// reviewer is handed.
//
//go:embed i18n/share.*.json
var vocabFiles embed.FS

type vocab struct {
	Lang       Lang              `json:"lang"`
	Review     string            `json:"review"`
	ReviewedBy string            `json:"reviewed_by"`
	ReviewedOn string            `json:"reviewed_on"`
	Note       string            `json:"note"`
	WardWord   string            `json:"ward_word"`
	Months     []string          `json:"months"`
	Sentences  map[string]string `json:"sentences"`
	Categories map[string]string `json:"categories"`
	Things     map[string]string `json:"things"`
}

// state maps the file's own word to the state a caller sees. Anything that is
// not explicitly reviewed is a draft: the default has to be the cautious one,
// or a new file ships as reviewed by omission.
func (v *vocab) state() ReviewState {
	if v.Review == "reviewed" {
		return Reviewed
	}
	return Draft
}

// date formats a day in this language's month names, with Latin digits in
// every language on purpose: a legal citation or a deadline checked against
// the source document is easier to verify with the digits the document uses.
func (v *vocab) date(t time.Time) string {
	t = t.UTC()
	m := t.Format("Jan")
	if len(v.Months) == 12 {
		m = v.Months[int(t.Month())-1]
	}
	return fmt.Sprintf("%d %s %d", t.Day(), m, t.Year())
}

// thing names what was photographed, as specifically as this language can.
//
// The fallback chain is deliberate and visible: the subcategory's own word,
// then the category's word, then English. It never guesses at a translation,
// and every step down is recorded so a gap in the vocabulary shows up in the
// response instead of reading as a finished sentence.
func (v *vocab) thing(f Facts) (string, []string) {
	if w, ok := v.Things[strings.ToLower(strings.TrimSpace(f.Subcategory))]; ok && w != "" {
		return w, nil
	}
	if w, ok := v.Categories[strings.TrimSpace(f.Category)]; ok && w != "" {
		return w, []string{"no " + string(v.Lang) + " word for " + f.Subcategory +
			"; used the category instead"}
	}
	en, err := vocabulary(EN)
	if err == nil && v.Lang != EN {
		if w, ok := en.Things[strings.ToLower(strings.TrimSpace(f.Subcategory))]; ok && w != "" {
			return w, []string{"no " + string(v.Lang) + " word for " + f.Subcategory +
				"; used English"}
		}
	}
	// Nothing in any vocabulary knows this word. The subcategory itself is the
	// honest answer, and Check still runs over the finished string.
	if s := strings.TrimSpace(f.Subcategory); s != "" {
		return s, []string{"no word for " + s + " in any vocabulary"}
	}
	return "", []string{"the capture has no subcategory"}
}

func (v *vocab) sentence(key string) string { return v.Sentences[key] }

var (
	vocabOnce sync.Once
	vocabs    map[Lang]*vocab
	vocabErr  error
)

func vocabulary(l Lang) (*vocab, error) {
	vocabOnce.Do(loadVocabs)
	if vocabErr != nil {
		return nil, vocabErr
	}
	v, ok := vocabs[l]
	if !ok {
		return nil, fmt.Errorf("share: no vocabulary for language %q", l)
	}
	return v, nil
}

func loadVocabs() {
	entries, err := vocabFiles.ReadDir("i18n")
	if err != nil {
		vocabErr = fmt.Errorf("share: read vocabularies: %w", err)
		return
	}
	vocabs = map[Lang]*vocab{}
	for _, e := range entries {
		body, err := vocabFiles.ReadFile("i18n/" + e.Name())
		if err != nil {
			vocabErr = fmt.Errorf("share: read %s: %w", e.Name(), err)
			return
		}
		var v vocab
		if err := json.Unmarshal(body, &v); err != nil {
			vocabErr = fmt.Errorf("share: parse %s: %w", e.Name(), err)
			return
		}
		if v.Lang == "" {
			vocabErr = fmt.Errorf("share: %s names no language", e.Name())
			return
		}
		vocabs[v.Lang] = &v
	}
	// A language offered by Languages with no file behind it would fail at
	// request time on a live server. Fail at startup instead.
	for _, l := range Languages() {
		if _, ok := vocabs[l]; !ok {
			vocabErr = fmt.Errorf("share: language %q is offered but has no vocabulary file", l)
			return
		}
	}
}

// Review reports how far each language's copy has been checked, so the
// native-speaker review is a visible task rather than an assumption.
func Review() map[Lang]ReviewState {
	out := map[Lang]ReviewState{}
	for _, l := range Languages() {
		if v, err := vocabulary(l); err == nil {
			out[l] = v.state()
		}
	}
	return out
}
