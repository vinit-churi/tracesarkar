// Package polish puts a model in the share path without letting it author a
// fact.
//
// [ADR 0020] kept every model out of `internal/share` and gave the reason:
// the share kit is the first surface that composes prose about a named party
// which will be forwarded beyond any page the platform can correct, and a
// fluent generator is good at exactly the thing that must never happen. That
// reasoning is still right about *authorship*. It was too strong about
// *wording*: the Marathi and Hindi templates are stilted, nobody who grew up
// speaking either has read them, and a stilted post in the reader's own
// language is a post that does not get sent.
//
// So the model is given the narrowest job the problem admits. It never sees a
// fact set and never writes a post. It receives the sentences
// `internal/share` already composed, one numbered line at a time, and returns
// the same lines in the same order, reworded. Everything that carries
// provenance — the permalink, the hashtag line — is never sent at all, and
// every line that comes back is checked span by span against the line that
// went out ([Verify]). A single failure anywhere discards the whole candidate
// and the deterministic post is served.
//
// What that buys, precisely: the model may improve a sentence and cannot
// introduce, remove, soften or sharpen a claim. The bound is still
// structural; it has just moved from "no model" to "the model's output must
// be provably the same facts".
//
// [ADR 0020]: ../../docs/04-adr/0020-share-text-is-generated-not-written-by-a-model.md
package polish

import (
	"strings"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// Segment is one line of a composed post, and whether a model may touch it.
//
// The split is not cosmetic. A permalink and a hashtag run are tokens other
// people resolve and search for, not prose: there is no better way to word
// them, and any rewording is damage. They are marked literal, never sent, and
// put back by this package rather than by the model.
type Segment struct {
	Index int
	Text  string
	Prose bool
}

// Split divides a composed post into the lines a model may reword and the
// lines it never sees.
//
// The literal lines are recognised by equality with the fact that produced
// them, not by looking for a "http" prefix or a leading "#". A post is
// forwarded as text and a user may edit it, so a line that merely looks like
// a permalink is prose; only the line that *is* this issue's permalink is
// held back.
func Split(f share.Facts, p share.Post) []Segment {
	tags := strings.Join(p.Hashtags, " ")
	lines := strings.Split(p.Text, "\n")
	segs := make([]Segment, 0, len(lines))
	for i, text := range lines {
		literal := (f.Permalink != "" && text == f.Permalink) ||
			(tags != "" && text == tags)
		segs = append(segs, Segment{Index: i, Text: text, Prose: !literal})
	}
	return segs
}

// Prose returns only the segments a model is allowed to reword.
func Prose(segs []Segment) []Segment {
	out := make([]Segment, 0, len(segs))
	for _, s := range segs {
		if s.Prose && strings.TrimSpace(s.Text) != "" {
			out = append(out, s)
		}
	}
	return out
}

// Reassemble puts the reworded lines back in their original places, taking
// the literal lines from the baseline rather than from the model.
//
// A line the model did not return keeps the platform's own wording, so a
// partial answer degrades line by line instead of failing the post. A line
// the model returned for a literal segment is ignored outright — that is the
// one place where ignoring the model is safer than reporting a fault, because
// the fault would be reported about text that is already correct.
func Reassemble(segs []Segment, got map[int]string) string {
	out := make([]string, 0, len(segs))
	for _, s := range segs {
		text := s.Text
		if s.Prose {
			if v, ok := got[s.Index]; ok && strings.TrimSpace(v) != "" {
				text = strings.TrimSpace(v)
			}
		}
		out = append(out, text)
	}
	return strings.Join(out, "\n")
}
