package polish

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// The measurement ADR 0020 made the condition for letting a model near this
// surface. It is guarded by TRACESARKAR_LIVE because it spends money and
// needs a provider, and it reports rather than merely passing: the numbers
// are the point, and they belong in a decision-log entry whenever the model
// or the prompt changes.
//
//	POLISH_BASE_URL=… POLISH_API_KEY=… POLISH_MODEL=… \
//	  TRACESARKAR_LIVE=1 go test ./internal/polish/ -run TestLive -v
func TestLiveTheModelIsMeasuredBeforeItIsTrusted(t *testing.T) {
	if os.Getenv("TRACESARKAR_LIVE") == "" {
		t.Skip("set TRACESARKAR_LIVE=1 to measure a real provider")
	}
	baseURL, model := os.Getenv("POLISH_BASE_URL"), os.Getenv("POLISH_MODEL")
	if baseURL == "" || model == "" {
		t.Skip("set POLISH_BASE_URL and POLISH_MODEL to measure a real provider")
	}

	g, err := NewGateway(GatewayOptions{
		BaseURL:    baseURL,
		APIKey:     os.Getenv("POLISH_API_KEY"),
		Model:      model,
		SchemaMode: SchemaMode(envOr("POLISH_SCHEMA_MODE", string(SchemaJSON))),
		// Generous here, unlike in production: this is a measurement, not a
		// page load, and a timeout would be recorded as a model failure.
		Timeout:     60 * time.Second,
		MaxAttempts: 2,
	})
	if err != nil {
		t.Fatalf("NewGateway: %v", err)
	}

	// Cached, exactly as production wraps it. The re-verification pass below
	// asks the same questions again, and paying twice to learn the same thing
	// is the mistake the cache exists to prevent.
	p := Cached(g, time.Hour, 4096)

	cases := Corpus()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
	defer cancel()

	score, refusals, err := Measure(ctx, p, cases)
	if err != nil {
		t.Fatalf("Measure: %v", err)
	}

	t.Logf("model %s, prompt %s\n%s", model, func() string {
		_, p := g.Describe()
		return p
	}(), score.String())
	t.Logf("invented facts: %d of %d posts", score.InventedFacts(), score.Posts)

	for _, r := range refusals {
		t.Logf("\n--- refused: %s / %s ---\nbaseline:\n%s\nfaults: %v",
			r.Case, r.Lang, r.Baseline, r.Faults)
	}

	// The invariant, independent of Apply: anything that was published must
	// still verify clean when checked again from scratch. A bug in Apply that
	// published a faulty candidate would be invisible to every other test.
	for _, c := range cases {
		base, err := share.Compose(c.Facts, c.Lang, c.Chan)
		if err != nil {
			t.Fatalf("compose %s: %v", c.Name, err)
		}
		res := Apply(ctx, p, c.Facts, base)
		if res.Post.Wording != share.WordedByModel {
			continue
		}
		segs := Split(c.Facts, base)
		if fs := Verify(c.Facts, base, segs, res.Got); len(fs) > 0 {
			t.Errorf("%s/%s/%s was published with faults %v:\n%s",
				c.Name, c.Lang, c.Chan, fs, res.Post.Text)
		}
		if res.Post.Review != base.Review {
			t.Errorf("%s/%s changed review state from %q to %q",
				c.Name, c.Lang, base.Review, res.Post.Review)
		}
	}

	// A model that is refused most of the time is not worth the call. This is
	// the only threshold, it is deliberately low, and it is about whether the
	// feature earns its cost rather than about whether the guard works.
	if min := score.Posts / 3; score.Published < min {
		t.Errorf("only %d of %d posts were publishable (floor %d); this model "+
			"or this prompt is not worth the call", score.Published, score.Posts, min)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
