package share

import (
	"strconv"
	"strings"
)

// line is one sentence with a priority. Priority decides what goes when a
// post will not fit: lower numbers are dropped first.
//
// The order encodes an editorial judgement. What the problem is, where it is,
// whose desk it is, and the link are the post; the corroboration count and the
// status gloss are useful but not the argument. The deadline sits above them
// because the obligation is the one fact an authority cannot argue with — and
// it is dropped before the permalink, because a forwarded claim with no page
// behind it is a rumour.
type line struct {
	key  string
	text string
	rank int
}

const (
	rankOptional = 1 // corroboration, status gloss
	rankDeadline = 2
	rankCore     = 3 // what, where, whose desk
	rankLink     = 4 // never dropped
)

func build(f Facts, v *vocab, ch Channel, p *Post) ([]line, []string, []Source) {
	thing, fallbacks := v.thing(f)
	var sources []Source
	var lines []line

	whatKey := "what_where_long"
	if ch == X {
		whatKey = "what_where_short"
	}
	lines = append(lines, line{"what_where", fill(v.sentence(whatKey), map[string]string{
		"thing": thing, "ward": f.Ward, "ward_word": v.WardWord,
		"authority": f.Authority,
	}), rankCore})

	// Whose desk it is. An institutional handle is safe on a first report; an
	// escalation handle is unlocked by an elapsed deadline and never earlier.
	if hs := handlesFor(f, ch); len(hs) > 0 {
		var names []string
		for _, h := range hs {
			names = append(names, h.Handle)
			sources = append(sources, Source{
				Fact: h.Handle + " is " + f.Authority + "'s own account on " + h.Platform,
				Ref:  h.SourceRef, RetrievedAt: h.RetrievedAt, FirstHand: h.FirstHand,
			})
		}
		lines = append(lines, line{"handles", fill(v.sentence("your_ward"), map[string]string{
			"handles": strings.Join(names, " "), "ward_word": v.WardWord,
		}), rankCore})
	}

	if s := v.sentence("status_" + f.Status); s != "" {
		lines = append(lines, line{"status", s, rankOptional})
	}

	// D121: the line appears from the second sighting onward. "Seen 1 times"
	// claims company the issue does not have.
	if f.Sightings > 1 {
		lines = append(lines, line{"seen", fill(v.sentence("seen"), map[string]string{
			"n": strconv.Itoa(f.Sightings), "date": v.date(f.FirstSeen),
		}), rankOptional})
	}

	if f.Lodged && f.LodgedOn != nil {
		lines = append(lines, line{"told_on", fill(v.sentence("told_on"), map[string]string{
			"date": v.date(*f.LodgedOn),
		}), rankOptional})
	} else {
		lines = append(lines, line{"not_told", v.sentence("not_told"), rankOptional})
	}

	if dl, src := deadline(f, v, ch); dl.text != "" {
		lines = append(lines, dl)
		if src != (Source{}) {
			sources = append(sources, src)
		}
	}

	if ch == WhatsApp {
		lines = append(lines, line{"sources_note", v.sentence("sources_note"), rankOptional})
	}
	return lines, fallbacks, sources
}

// deadline states the obligation, or says nothing.
//
// A deadline shown without the document that creates it is an assertion, and
// this platform does not make those (hard rule 6's companion: the citation
// travels with the constant). So a missing citation drops the sentence
// entirely rather than printing the number alone.
func deadline(f Facts, v *vocab, ch Channel) (line, Source) {
	if f.DeadlineHours <= 0 {
		return line{}, Source{}
	}
	short, full := strings.TrimSpace(f.DeadlineShortRef), strings.TrimSpace(f.DeadlineCitation)
	if short == "" && full == "" {
		return line{}, Source{}
	}
	// X gets the short reference and the full text goes on the page. A legal
	// citation truncated to fit a window is a wrong citation, so the short
	// form is a separate recorded field rather than a prefix of the long one.
	ref, key := full, "deadline_running_long"
	if ch == X {
		ref, key = short, "deadline_running_short"
		if ref == "" {
			return line{}, Source{}
		}
	} else if ref == "" {
		return line{}, Source{}
	}

	switch {
	case f.DeadlineState == "elapsed":
		key = strings.Replace(key, "running", "elapsed", 1)
	case !f.Lodged:
		// Nothing is running, so the obligation is stated as a condition. The
		// platform never implies it told anybody.
		if ch == X {
			return line{}, Source{}
		}
		key = "deadline_obligation_long"
	}

	text := fill(v.sentence(key), map[string]string{
		"hours": strconv.Itoa(f.DeadlineHours), "short_ref": short, "citation": full,
	})
	return line{"deadline", text, rankDeadline},
		Source{Fact: strconv.Itoa(f.DeadlineHours) + " hours to attend", Ref: full}
}

// handlesFor filters the registry's answer for this post.
func handlesFor(f Facts, ch Channel) []Handle {
	platform := "x"
	if ch == WhatsApp {
		platform = "x" // the handle is quoted in text, so the X name is the one people know
	}
	var out []Handle
	for _, h := range f.Handles {
		if h.Platform != platform || strings.TrimSpace(h.Handle) == "" {
			continue
		}
		// No source, no handle. There is no "probably official" state.
		if strings.TrimSpace(h.SourceRef) == "" {
			continue
		}
		if h.Kind == Escalation && f.DeadlineState != "elapsed" {
			continue
		}
		out = append(out, h)
	}
	return out
}

func assemble(lines []line, f Facts, tags []string) string {
	var b []string
	for _, l := range lines {
		if strings.TrimSpace(l.text) != "" {
			b = append(b, l.text)
		}
	}
	if f.Permalink != "" {
		b = append(b, f.Permalink)
	}
	if len(tags) > 0 {
		b = append(b, strings.Join(tags, " "))
	}
	return strings.Join(b, "\n")
}

// fit drops sentences, lowest rank first, until the post is inside the window.
//
// It never cuts a sentence in half. A post ending mid-citation is worse than a
// post missing a line, and the missing line is recorded in Dropped so the
// difference between a short post and a lossy one is visible.
func fit(text string, lines []line, f Facts, tags []string, p *Post) string {
	if len([]rune(text)) <= XLimit {
		return text
	}
	kept := make([]line, len(lines))
	copy(kept, lines)
	for rank := rankOptional; rank < rankLink; rank++ {
		for i := 0; i < len(kept); i++ {
			if kept[i].rank != rank {
				continue
			}
			p.Dropped = append(p.Dropped, kept[i].key)
			kept = append(kept[:i], kept[i+1:]...)
			i--
			if text = assemble(kept, f, tags); len([]rune(text)) <= XLimit {
				return text
			}
		}
	}
	return text
}

func fill(tmpl string, vals map[string]string) string {
	if tmpl == "" {
		return ""
	}
	out := tmpl
	for k, v := range vals {
		out = strings.ReplaceAll(out, "{"+k+"}", v)
	}
	return strings.TrimSpace(out)
}

// wardTag turns a ward code into a hashtag. "R/S" is not a tag; "#RSWard" is.
func wardTag(ward string) string {
	s := strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, strings.ToUpper(strings.TrimSpace(ward)))
	if s == "" {
		return ""
	}
	return "#" + s + "Ward"
}
