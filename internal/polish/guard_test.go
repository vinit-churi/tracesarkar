package polish

import (
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// asSent is the model's answer when it changed nothing: every prose line
// returned exactly as it went out.
func asSent(segs []Segment) map[int]string {
	got := map[int]string{}
	for _, s := range Prose(segs) {
		got[s.Index] = s.Text
	}
	return got
}

func kinds(fs []Fault) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, f.Kind)
	}
	return out
}

func hasKind(fs []Fault, kind string) bool {
	for _, f := range fs {
		if f.Kind == kind {
			return true
		}
	}
	return false
}

// The guard must not refuse an honest answer, or it gets switched off.
func TestAnUnchangedAnswerIsClean(t *testing.T) {
	for _, lang := range share.Languages() {
		for _, ch := range share.Channels() {
			p := baseline(t, lang, ch)
			segs := Split(facts(), p)
			if fs := Verify(facts(), p, segs, asSent(segs)); len(fs) > 0 {
				t.Errorf("%s/%s: unchanged answer reported %v", lang, ch, kinds(fs))
			}
		}
	}
}

// A reworded line that keeps every span and every number is what the whole
// path exists to allow.
func TestAnHonestRewordingIsAccepted(t *testing.T) {
	p := baseline(t, share.EN, share.WhatsApp)
	segs := Split(facts(), p)
	got := asSent(segs)
	for i, text := range got {
		if strings.HasPrefix(text, "Pothole in") {
			got[i] = "There is a pothole on a road in K/E ward, BMC."
		}
		if strings.HasPrefix(text, "Seen 4 times") {
			got[i] = "It has been reported 4 times since 2 Sep 2026."
		}
	}
	if fs := Verify(facts(), p, segs, got); len(fs) > 0 {
		t.Errorf("an honest rewording was refused: %v", fs)
	}
}

func TestTheGuardCatchesWhatMustNeverReachAPost(t *testing.T) {
	// Each case rewrites the line that starts with `on` and expects `want`.
	cases := []struct {
		name string
		lang share.Lang
		ch   share.Channel
		on   string
		to   string
		want string
	}{
		{
			// The number is the obligation. 48 is a court order; 72 is a lie.
			name: "a changed number",
			lang: share.EN, ch: share.WhatsApp,
			on: "48 hours", to: "72 hours to attend. That time has passed — Bombay High Court order in PIL 85 of 2013",
			want: "number_changed",
		},
		{
			name: "a dropped number",
			lang: share.EN, ch: share.WhatsApp,
			on: "Seen 4 times", to: "It has been seen repeatedly since 2 Sep 2026.",
			want: "number_changed",
		},
		{
			// A citation that has been reworded is no longer a citation.
			name: "a reworded citation",
			lang: share.EN, ch: share.WhatsApp,
			on: "48 hours", to: "48 hours to attend. That time has passed — a Bombay High Court order of 2013",
			want: "span_lost",
		},
		{
			name: "a dropped handle",
			lang: share.EN, ch: share.WhatsApp,
			on: "@mybmc", to: "This is your ward.",
			want: "span_lost",
		},
		{
			name: "an invented handle",
			lang: share.EN, ch: share.WhatsApp,
			on: "@mybmc", to: "@mybmc @CMOMaharashtra — this is your ward.",
			want: "token_invented",
		},
		{
			name: "an invented link",
			lang: share.EN, ch: share.WhatsApp,
			on: "Still open", to: "Still open. See https://example.org/evidence",
			want: "token_invented",
		},
		{
			// Hard rule 3, enforced on the model's words with the same list
			// that runs on the platform's own.
			name: "an accusation",
			lang: share.EN, ch: share.WhatsApp,
			on: "Still open", to: "Still open — this is pure corruption.",
			want: "tone",
		},
		{
			name: "a line answered in the wrong language",
			lang: share.MR, ch: share.WhatsApp,
			on: "अद्याप", to: "This is still open.",
			want: "script_lost",
		},
		{
			name: "a line that grew into commentary",
			lang: share.EN, ch: share.WhatsApp,
			on: "Still open", to: "Still open, and residents here have been waiting a very " +
				"long time for somebody at the corporation to take the smallest amount of " +
				"interest in a road that everyone who lives on it uses every single day.",
			want: "line_too_long",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := baseline(t, c.lang, c.ch)
			segs := Split(facts(), p)
			got := asSent(segs)
			var hit bool
			for i, text := range got {
				if strings.Contains(text, c.on) {
					got[i], hit = c.to, true
					break
				}
			}
			if !hit {
				t.Fatalf("no baseline line contains %q; post is:\n%s", c.on, p.Text)
			}
			fs := Verify(facts(), p, segs, got)
			if !hasKind(fs, c.want) {
				t.Errorf("want a %q fault, got %v", c.want, kinds(fs))
			}
		})
	}
}

// A half-polished post reads worse than an unpolished one: two registers in
// one message look like two people wrote it.
func TestAPartialAnswerIsRefused(t *testing.T) {
	p := baseline(t, share.MR, share.WhatsApp)
	segs := Split(facts(), p)
	got := asSent(segs)
	for i := range got {
		delete(got, i)
		break
	}
	if fs := Verify(facts(), p, segs, got); !hasKind(fs, "line_missing") {
		t.Errorf("want a line_missing fault, got %v", kinds(fs))
	}
}

// A line the model was never given is a model that stopped following the
// contract, and nothing it returned can be trusted after that.
func TestALineThatWasNeverSentIsRefused(t *testing.T) {
	p := baseline(t, share.EN, share.WhatsApp)
	segs := Split(facts(), p)
	got := asSent(segs)
	got[len(segs)+7] = "Also, vote them out."
	if fs := Verify(facts(), p, segs, got); !hasKind(fs, "line_unknown") {
		t.Errorf("want a line_unknown fault, got %v", kinds(fs))
	}
	// The permalink's own index was never sent either.
	var literal = -1
	for _, s := range segs {
		if !s.Prose {
			literal = s.Index
		}
	}
	got = asSent(segs)
	got[literal] = "https://evil.example/phish"
	if fs := Verify(facts(), p, segs, got); !hasKind(fs, "line_unknown") {
		t.Errorf("a literal index came back and was not refused: %v", kinds(fs))
	}
}

// X rejects the post rather than truncating it, and a truncated citation is a
// wrong citation.
func TestAPolishedPostMustStillFitTheWindow(t *testing.T) {
	p := baseline(t, share.EN, share.X)
	segs := Split(facts(), p)
	got := asSent(segs)
	for i, text := range got {
		if strings.HasPrefix(text, "Pothole") {
			got[i] = text + " " + strings.Repeat("padding ", 40)
			break
		}
	}
	fs := Verify(facts(), p, segs, got)
	if !hasKind(fs, "too_long_for_channel") && !hasKind(fs, "line_too_long") {
		t.Errorf("an over-length X post was not refused: %v", kinds(fs))
	}
}

// English copy with Devanagari in it means the model answered the wrong
// prompt, and a mixed-script post reads as broken.
func TestEnglishMustComeBackInEnglish(t *testing.T) {
	p := baseline(t, share.EN, share.WhatsApp)
	segs := Split(facts(), p)
	got := asSent(segs)
	for i, text := range got {
		if strings.HasPrefix(text, "Still open") {
			got[i] = "अद्याप प्रलंबित."
			break
		}
	}
	if fs := Verify(facts(), p, segs, got); !hasKind(fs, "script_lost") {
		t.Errorf("want a script_lost fault, got %v", kinds(fs))
	}
}
