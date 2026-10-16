package polish

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// stub is a Polisher that returns whatever it is told to.
type stub struct {
	reply map[int]string
	err   error
	calls int
	// seen records what was actually sent, so a test can assert the model
	// never saw something.
	seen []Segment
}

func (s *stub) Reword(_ context.Context, _ share.Lang, _ share.Channel, lines []Segment) (map[int]string, error) {
	s.calls++
	s.seen = append(s.seen, lines...)
	return s.reply, s.err
}

func (s *stub) Describe() (string, string) { return "stub-model", "stub-prompt" }

// echo returns every line unchanged, which is a legitimate answer.
func echo(segs []Segment) *stub {
	r := map[int]string{}
	for _, s := range Prose(segs) {
		r[s.Index] = s.Text
	}
	return &stub{reply: r}
}

func TestNoPolisherMeansTheDeterministicPost(t *testing.T) {
	p := baseline(t, share.MR, share.WhatsApp)
	got := Apply(context.Background(), nil, facts(), p)
	if got.Post.Text != p.Text {
		t.Error("the post changed with no polisher configured")
	}
	if got.Post.Wording != share.WordedByTemplate {
		t.Errorf("wording is %q, want %q", got.Post.Wording, share.WordedByTemplate)
	}
}

func TestACleanAnswerIsPublishedAndSaysSo(t *testing.T) {
	p := baseline(t, share.MR, share.WhatsApp)
	segs := Split(facts(), p)
	s := echo(segs)
	for i, text := range s.reply {
		if strings.HasPrefix(text, "अद्याप") {
			s.reply[i] = "हे अजून प्रलंबित आहे."
		}
	}
	got := Apply(context.Background(), s, facts(), p)
	if len(got.Faults) > 0 {
		t.Fatalf("a clean answer was refused: %v", got.Faults)
	}
	if !strings.Contains(got.Post.Text, "हे अजून प्रलंबित आहे.") {
		t.Errorf("the reworded line is not in the post:\n%s", got.Post.Text)
	}
	if got.Post.Wording != share.WordedByModel {
		t.Errorf("wording is %q, want %q", got.Post.Wording, share.WordedByModel)
	}
	if got.Model != "stub-model" || got.Prompt != "stub-prompt" {
		t.Errorf("the answer does not name what produced it: %q %q", got.Model, got.Prompt)
	}
}

// The one rule that must not bend: a model is not a native speaker. A
// model-worded Marathi post is still a draft, and anything that flipped that
// would make the review look done when nobody has done it.
func TestAModelDoesNotMakeALanguageReviewed(t *testing.T) {
	for _, lang := range []share.Lang{share.MR, share.HI} {
		p := baseline(t, lang, share.WhatsApp)
		segs := Split(facts(), p)
		got := Apply(context.Background(), echo(segs), facts(), p)
		if got.Post.Review != share.Draft {
			t.Errorf("%s became %q after polishing, want %q",
				lang, got.Post.Review, share.Draft)
		}
	}
}

func TestAFaultMeansTheDeterministicPost(t *testing.T) {
	p := baseline(t, share.EN, share.WhatsApp)
	segs := Split(facts(), p)
	s := echo(segs)
	for i, text := range s.reply {
		if strings.HasPrefix(text, "48 hours") {
			s.reply[i] = "72 hours to attend. That time has passed — Bombay High Court order in PIL 85 of 2013"
		}
	}
	got := Apply(context.Background(), s, facts(), p)
	if got.Post.Text != p.Text {
		t.Errorf("a post with a changed number was published:\n%s", got.Post.Text)
	}
	if len(got.Faults) == 0 {
		t.Error("the refusal recorded no fault, so nothing can be measured")
	}
	if got.Post.Wording != share.WordedByTemplate {
		t.Errorf("wording is %q, want %q", got.Post.Wording, share.WordedByTemplate)
	}
}

func TestAModelFailureIsNotAFailedShare(t *testing.T) {
	p := baseline(t, share.HI, share.X)
	got := Apply(context.Background(), &stub{err: errors.New("gateway timeout")}, facts(), p)
	if got.Post.Text != p.Text {
		t.Error("a model failure changed the post")
	}
	if len(got.Faults) == 0 || got.Faults[0].Kind != "model_error" {
		t.Errorf("want a model_error fault, got %v", got.Faults)
	}
}

// The permalink identifies the issue and the person's report behind it. It is
// never sent, which also means the model is never told which issue this is.
func TestTheModelIsNeverToldWhichIssueThisIs(t *testing.T) {
	f := facts()
	p := baseline(t, share.EN, share.WhatsApp)
	segs := Split(f, p)
	s := echo(segs)
	Apply(context.Background(), s, f, p)
	for _, seg := range s.seen {
		if strings.Contains(seg.Text, f.Permalink) || strings.Contains(seg.Text, f.IssueID) {
			t.Errorf("the model was sent %q", seg.Text)
		}
	}
}
