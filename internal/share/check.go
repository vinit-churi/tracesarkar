package share

import (
	"fmt"
	"regexp"
	"strings"
)

// Violation is one thing the text may not say.
type Violation struct {
	Kind  string `json:"kind"`
	Term  string `json:"term"`
	Where string `json:"where"`
}

func (v Violation) String() string {
	return fmt.Sprintf("%s: %q", v.Kind, v.Term)
}

// The prohibited-pattern list (share kit §3). It runs on the finished string
// rather than being trusted to the templates, because the templates are not
// the only author: the subcategory comes from a classifier, the authority name
// from a database, and the user may edit the text before it is sent.
//
// Three kinds, because they fail for different reasons:
//
//   - wrongdoing — hard rule 3. The platform states facts and joins. It does
//     not conclude that anybody is corrupt, and a post that does is a legal
//     exposure as well as a broken promise.
//   - insult — an insult gets the post dismissed and the platform with it.
//   - political — a party reference turns a civic fact into a partisan claim
//     and loses half the audience on contact.
//
// Terms are matched on word boundaries. Substring matching here would refuse
// "scammed" and "assessment" alike, and a validator that refuses real posts
// gets switched off.
var prohibited = []struct {
	kind  string
	terms []string
}{
	{"wrongdoing", []string{
		"corrupt", "corruption", "bribe", "bribes", "bribery", "kickback",
		"kickbacks", "scam", "scams", "fraud", "fraudulent", "embezzle",
		"embezzled", "embezzlement", "loot", "looted", "stolen", "theft",
		"guilty", "criminal", "nexus", "collusion", "colluded", "mafia",
	}},
	{"insult", []string{
		"useless", "idiot", "idiots", "stupid", "incompetent", "pathetic",
		"shameless", "disgrace", "disgraceful", "joke", "clowns", "lazy",
		"nikamma", "chor",
	}},
	{"political", []string{
		"bjp", "congress", "shiv sena", "sena", "ncp", "mns", "aimim",
		"aap", "udhhav", "uddhav", "fadnavis", "thackeray",
	}},
}

var prohibitedRes = func() []struct {
	kind string
	re   *regexp.Regexp
} {
	var out []struct {
		kind string
		re   *regexp.Regexp
	}
	for _, g := range prohibited {
		quoted := make([]string, 0, len(g.terms))
		for _, t := range g.terms {
			quoted = append(quoted, regexp.QuoteMeta(t))
		}
		// \b on both ends: a term inside a longer word is a different word.
		re := regexp.MustCompile(`(?i)\b(` + strings.Join(quoted, "|") + `)\b`)
		out = append(out, struct {
			kind string
			re   *regexp.Regexp
		}{g.kind, re})
	}
	return out
}()

// Check reports everything in the text that may not be published.
//
// It is deliberately not a score and not a model call: a post is either clear
// of the list or it is not, and the answer has to be the same every time for
// the same string.
func Check(text string) []Violation {
	var out []Violation
	for _, g := range prohibitedRes {
		for _, m := range g.re.FindAllString(text, -1) {
			out = append(out, Violation{Kind: g.kind, Term: strings.ToLower(m), Where: text})
		}
	}
	return out
}
