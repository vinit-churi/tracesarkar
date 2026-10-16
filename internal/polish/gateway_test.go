package polish

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// reply wraps content in the provider envelope every OpenAI-compatible
// endpoint returns.
func reply(content string) string {
	b, _ := json.Marshal(map[string]any{
		"choices": []any{map[string]any{
			"finish_reason": "stop",
			"message":       map[string]any{"content": content},
		}},
	})
	return string(b)
}

func serve(t *testing.T, h http.HandlerFunc) (*Gateway, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	g, err := NewGateway(GatewayOptions{
		BaseURL: srv.URL, APIKey: "k", Model: "test-model",
		MaxAttempts: 1, Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewGateway: %v", err)
	}
	return g, srv
}

func TestTheGatewayAsksForStructuredOutputAndNothingElse(t *testing.T) {
	var body map[string]any
	g, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("request is not JSON: %v", err)
		}
		io.WriteString(w, reply(`{"lines":[{"i":0,"text":"ok"}]}`))
	})

	got, err := g.Reword(context.Background(), share.MR, share.WhatsApp,
		[]Segment{{Index: 0, Text: "something", Prose: true}})
	if err != nil {
		t.Fatalf("Reword: %v", err)
	}
	if got[0] != "ok" {
		t.Errorf("got %q", got[0])
	}

	rf, _ := body["response_format"].(map[string]any)
	if rf["type"] != "json_schema" {
		t.Errorf("response_format is %v, want json_schema", rf["type"])
	}
	msgs, _ := body["messages"].([]any)
	if len(msgs) != 2 {
		t.Fatalf("sent %d messages", len(msgs))
	}
	system, _ := msgs[0].(map[string]any)["content"].(string)
	// The register block for the requested language has to be in the prompt,
	// or the prompt is generic and the Marathi comes back generic.
	if !strings.Contains(system, "विभाग") || !strings.Contains(system, "प्रभाग") {
		t.Error("the Marathi register block is not in the system prompt")
	}
	if !strings.Contains(system, "Marathi") {
		t.Error("the system prompt does not name the language")
	}
}

func TestTheXWindowReachesThePrompt(t *testing.T) {
	var system string
	g, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct{ Content string } `json:"messages"`
		}
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &body)
		if len(body.Messages) > 0 {
			system = body.Messages[0].Content
		}
		io.WriteString(w, reply(`{"lines":[{"i":0,"text":"ok"}]}`))
	})
	if _, err := g.Reword(context.Background(), share.EN, share.X,
		[]Segment{{Index: 0, Text: "x", Prose: true}}); err != nil {
		t.Fatalf("Reword: %v", err)
	}
	if !strings.Contains(system, "280") {
		t.Error("the X window is not stated in the prompt")
	}
}

func TestAProviderFailureIsAnError(t *testing.T) {
	cases := []struct {
		name string
		h    http.HandlerFunc
	}{
		{"a 500", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "upstream exploded")
		}},
		{"no choices", func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, `{"choices":[]}`)
		}},
		{"empty content", func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, reply(""))
		}},
		{"content that is not JSON", func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, reply("Sure! Here are your lines."))
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g, _ := serve(t, c.h)
			if _, err := g.Reword(context.Background(), share.EN, share.WhatsApp,
				[]Segment{{Index: 0, Text: "x", Prose: true}}); err == nil {
				t.Error("want an error")
			}
		})
	}
}

// Some providers wrap JSON in a fence even when told not to. That is a
// provider quirk, not a refusal, and treating it as one would switch the
// feature off for everyone on that provider.
func TestAFencedAnswerIsStillAnAnswer(t *testing.T) {
	g, _ := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, reply("```json\n{\"lines\":[{\"i\":3,\"text\":\"fenced\"}]}\n```"))
	})
	got, err := g.Reword(context.Background(), share.EN, share.WhatsApp,
		[]Segment{{Index: 3, Text: "x", Prose: true}})
	if err != nil {
		t.Fatalf("Reword: %v", err)
	}
	if got[3] != "fenced" {
		t.Errorf("got %v", got)
	}
}

// A provider that is down must not be asked once per page load for as long as
// it stays down.
func TestTheBreakerStopsAskingAProviderThatIsDown(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	g, err := NewGateway(GatewayOptions{
		BaseURL: srv.URL, APIKey: "k", Model: "m", MaxAttempts: 1,
		BreakerThreshold: 3, BreakerCooldown: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewGateway: %v", err)
	}
	seg := []Segment{{Index: 0, Text: "x", Prose: true}}
	for i := 0; i < 10; i++ {
		_, err := g.Reword(context.Background(), share.EN, share.WhatsApp, seg)
		if err == nil {
			t.Fatal("want an error")
		}
		if i >= 3 && !errors.Is(err, ErrBreakerOpen) {
			t.Fatalf("call %d: want the breaker open, got %v", i, err)
		}
	}
	if n := atomic.LoadInt32(&calls); n != 3 {
		t.Errorf("the provider was called %d times, want 3", n)
	}
}

// A share kit is a public GET already cached at the edge. Paying for one
// answer many times is the cheapest mistake to avoid.
func TestTheCacheAsksOnceForTheSameLines(t *testing.T) {
	s := &stub{reply: map[int]string{0: "reworded"}}
	c := Cached(s, time.Minute, 8)
	seg := []Segment{{Index: 0, Text: "same line", Prose: true}}
	for i := 0; i < 5; i++ {
		got, err := c.Reword(context.Background(), share.MR, share.WhatsApp, seg)
		if err != nil {
			t.Fatalf("Reword: %v", err)
		}
		if got[0] != "reworded" {
			t.Fatalf("got %v", got)
		}
	}
	if s.calls != 1 {
		t.Errorf("the model was asked %d times, want 1", s.calls)
	}
	// A different language is a different question.
	if _, err := c.Reword(context.Background(), share.HI, share.WhatsApp, seg); err != nil {
		t.Fatalf("Reword: %v", err)
	}
	if s.calls != 2 {
		t.Errorf("a different language reused the answer (%d calls)", s.calls)
	}
}

// Caching a failure would keep serving the template for the whole TTL after a
// blip.
func TestTheCacheDoesNotRememberAFailure(t *testing.T) {
	s := &stub{err: errors.New("blip")}
	c := Cached(s, time.Minute, 8)
	seg := []Segment{{Index: 0, Text: "x", Prose: true}}
	for i := 0; i < 3; i++ {
		if _, err := c.Reword(context.Background(), share.EN, share.WhatsApp, seg); err == nil {
			t.Fatal("want an error")
		}
	}
	if s.calls != 3 {
		t.Errorf("the failure was cached: %d calls", s.calls)
	}
}

func TestAGatewayWithNoModelOrURLIsAStartupError(t *testing.T) {
	if _, err := NewGateway(GatewayOptions{Model: "m"}); err == nil {
		t.Error("a gateway with no base URL was accepted")
	}
	if _, err := NewGateway(GatewayOptions{BaseURL: "https://x"}); err == nil {
		t.Error("a gateway with no model was accepted")
	}
	if _, err := NewGateway(GatewayOptions{
		BaseURL: "https://x", Model: "m", PromptName: "no_such_prompt",
	}); err == nil {
		t.Error("a gateway with a missing prompt file was accepted")
	}
}
