package polish

import (
	"context"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// Polisher rewords the prose lines of one post.
//
// One method, so a second provider is a second implementation rather than a
// rewrite, and so the eval harness can measure any of them identically — the
// shape `internal/classify` settled on for the same reason.
type Polisher interface {
	Reword(ctx context.Context, lang share.Lang, ch share.Channel, lines []Segment) (map[int]string, error)
	// Describe names the model and prompt version behind an answer, for the
	// log line that records why a post reads the way it does.
	Describe() (model, prompt string)
}

// Result is what Apply decided, including when it decided to change nothing.
type Result struct {
	Post share.Post
	// Faults is empty only when the model's answer was published. It is kept
	// on a refusal because the refusal is the interesting event: it is the
	// measurement the eval set is built from.
	Faults []Fault
	// Got is the model's raw answer, kept so a refusal can be read rather
	// than inferred from a count — the eval set is built out of these.
	Got    map[int]string
	Model  string
	Prompt string
}

// Apply rewords a composed post, or returns it untouched.
//
// It cannot fail. Every path that is not a clean, verified answer returns the
// deterministic post the platform composed, because a share that does not
// happen is the one outcome worse than a stilted one. The caller logs
// Result.Faults and serves Result.Post either way.
func Apply(ctx context.Context, p Polisher, f share.Facts, base share.Post) Result {
	if p == nil {
		return Result{Post: base}
	}
	model, prompt := p.Describe()
	out := Result{Post: base, Model: model, Prompt: prompt}

	segs := Split(f, base)
	sent := Prose(segs)
	if len(sent) == 0 {
		return out
	}

	got, err := p.Reword(ctx, base.Lang, base.Channel, sent)
	if err != nil {
		out.Faults = []Fault{{"model_error", -1, err.Error()}}
		return out
	}
	out.Got = got
	if fs := Verify(f, base, segs, got); len(fs) > 0 {
		out.Faults = fs
		return out
	}

	// Only the text changes. Hashtags are derived from the fact set, Sources
	// are the provenance of facts nobody reworded, Dropped records what the
	// channel window cost, and Review is a statement about a native speaker
	// that a model cannot make true.
	out.Post.Text = Reassemble(segs, got)
	out.Post.Wording = share.WordedByModel
	return out
}
