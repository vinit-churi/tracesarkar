package share

import (
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/classify"
)

// The one test in this package that reaches for another, and it earns it: the
// classifier's taxonomy is the authority for what a photograph can be
// classified as, and the vocabulary is what a post calls it. If they
// disagree, a post names the *category* where it should name the thing —
// "Environment in H/W ward" for a felled tree, which is a sentence that says
// nothing. The fallback is recorded in Post.Fallbacks so it is never silent,
// but a recorded uselessness is still useless.
//
// English only. Marathi and Hindi are a review task, not a code task, and a
// test that demanded them would either be permanently red or would invite
// somebody to invent a Marathi word to make it pass.
func TestEnglishHasAWordForEverySubcategoryTheClassifierCanReturn(t *testing.T) {
	v, err := vocabulary(EN)
	if err != nil {
		t.Fatalf("vocabulary(en): %v", err)
	}
	for _, cat := range classify.TopCategories() {
		for _, sub := range classify.Subcategories(cat) {
			if w := v.Things[sub]; w == "" {
				t.Errorf("no English word for %q (%s); a post would say %q instead",
					sub, cat, v.Categories[cat])
			}
		}
	}
}

// Every category needs a word too, because that is what a subcategory nobody
// has translated falls back to.
func TestEveryLanguageHasAWordForEveryCategory(t *testing.T) {
	for _, lang := range Languages() {
		v, err := vocabulary(lang)
		if err != nil {
			t.Fatalf("vocabulary(%q): %v", lang, err)
		}
		for _, cat := range classify.TopCategories() {
			if v.Categories[cat] == "" {
				t.Errorf("%s has no word for the category %q", lang, cat)
			}
		}
	}
}
