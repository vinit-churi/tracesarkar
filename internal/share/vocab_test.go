package share

import (
	"strings"
	"testing"
)

// A citation is the one span in a post that passes through every language
// unchanged, which means the platform does not know how it ends. Some end in a
// year, some in a document number whose last character is a full stop
// ("No.Ch.E./25MGC/Bridges dated 28.06.2019"). Putting words after it therefore
// produces either a run-on or a double stop, and the reader cannot tell where
// the quoted reference stopped and the platform started speaking.
//
// So the rule is structural rather than per-language: in every vocabulary file,
// {citation} is the last thing in its sentence.
func TestACitationIsTheLastThingItsSentenceSays(t *testing.T) {
	for _, l := range Languages() {
		v, err := vocabulary(l)
		if err != nil {
			t.Fatalf("vocabulary(%q): %v", l, err)
		}
		for key, s := range v.Sentences {
			i := strings.Index(s, "{citation}")
			if i < 0 {
				continue
			}
			if tail := strings.TrimSpace(s[i+len("{citation}"):]); tail != "" {
				t.Errorf("%s/%s says %q after the citation:\n%s", l, key, tail, s)
			}
		}
	}
}

// Every language offers the same sentences, or a post in one language silently
// drops a fact another language states. A missing key falls back to English
// inside vocab.sentence, which hides the gap instead of failing.
func TestEveryLanguageCarriesTheSameSentences(t *testing.T) {
	en, err := vocabulary(EN)
	if err != nil {
		t.Fatalf("vocabulary(en): %v", err)
	}
	for _, l := range Languages() {
		if l == EN {
			continue
		}
		v, err := vocabulary(l)
		if err != nil {
			t.Fatalf("vocabulary(%q): %v", l, err)
		}
		for key := range en.Sentences {
			if strings.TrimSpace(v.Sentences[key]) == "" {
				t.Errorf("%s has no sentence for %q", l, key)
			}
		}
	}
}
