package polish

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// SchemaMode is how the provider is asked to return structured data. The same
// two modes internal/classify needs, for the same reason: some providers
// reject a named schema outright and the shape then has to reach the model
// through the prompt and be validated here.
type SchemaMode string

const (
	SchemaJSON       SchemaMode = "json_schema"
	SchemaJSONObject SchemaMode = "json_object"
)

// GatewayOptions configures a client for an OpenAI-compatible endpoint.
type GatewayOptions struct {
	BaseURL string
	APIKey  string
	Model   string

	PromptName string
	SchemaMode SchemaMode
	// Timeout is deliberately short. This call sits on a page load, not on a
	// background enrichment: a share kit that arrives late is a share that
	// does not happen, and the deterministic post is already correct. Spending
	// a classification's 45-second budget here would be the wrong trade.
	Timeout     time.Duration
	MaxAttempts int
	// Breaker stops a provider that is down from being asked once per request
	// for as long as it stays down.
	BreakerThreshold int
	BreakerCooldown  time.Duration
	HTTPClient       *http.Client
}

// Gateway rewords lines through an OpenAI-compatible endpoint.
//
// What crosses the wire is worth stating precisely: the prose lines of a post
// composed from store.PublicIssueView, which is the coarsened, reference-free
// view the public permalink is built from. No coordinate, no reporter, no
// media URL, no complaint reference — and not the permalink either, so the
// provider is never told which issue the sentences belong to.
type Gateway struct {
	opts GatewayOptions
	pr   Prompt

	mu       sync.Mutex
	failures int
	openedAt time.Time
}

// NewGateway builds a client and loads its prompt, so a missing or malformed
// prompt file is a startup error rather than a surprise on the first share.
func NewGateway(o GatewayOptions) (*Gateway, error) {
	if strings.TrimSpace(o.BaseURL) == "" {
		return nil, fmt.Errorf("polish: no base URL")
	}
	if strings.TrimSpace(o.Model) == "" {
		return nil, fmt.Errorf("polish: no model")
	}
	if o.PromptName == "" {
		o.PromptName = "share_polish_v1"
	}
	if o.SchemaMode == "" {
		o.SchemaMode = SchemaJSON
	}
	if o.Timeout == 0 {
		o.Timeout = 8 * time.Second
	}
	if o.MaxAttempts == 0 {
		o.MaxAttempts = 2
	}
	if o.BreakerThreshold == 0 {
		o.BreakerThreshold = 5
	}
	if o.BreakerCooldown == 0 {
		o.BreakerCooldown = 2 * time.Minute
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{}
	}
	pr, err := LoadPrompt(o.PromptName)
	if err != nil {
		return nil, err
	}
	return &Gateway{opts: o, pr: pr}, nil
}

// Describe names the model and the prompt version behind an answer.
func (g *Gateway) Describe() (string, string) {
	return g.opts.Model, g.pr.Name + "@" + g.pr.Version
}

// ErrBreakerOpen is returned while the provider is being left alone.
var ErrBreakerOpen = fmt.Errorf("polish: provider is in cooldown after repeated failures")

// Reword sends the prose lines and returns the model's rewording.
func (g *Gateway) Reword(ctx context.Context, lang share.Lang, ch share.Channel, lines []Segment) (map[int]string, error) {
	if len(lines) == 0 {
		return nil, nil
	}
	if g.breakerOpen() {
		return nil, ErrBreakerOpen
	}

	type outLine struct {
		I    int    `json:"i"`
		Text string `json:"text"`
	}
	send := make([]outLine, 0, len(lines))
	for _, l := range lines {
		send = append(send, outLine{l.Index, l.Text})
	}
	payload, err := json.MarshalIndent(map[string]any{"lines": send}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode lines: %w", err)
	}

	system := g.pr.WithLanguage(lang, ch)
	var responseFormat map[string]any
	switch g.opts.SchemaMode {
	case SchemaJSONObject:
		responseFormat = map[string]any{"type": "json_object"}
		system += "\n\n## Reply format\n\nReply with JSON only, matching exactly:\n\n" +
			"{\"lines\":[{\"i\":<the same number you were given>,\"text\":\"<the reworded line>\"}]}\n"
	default:
		responseFormat = map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "share_rewording",
				"strict": true,
				"schema": rewordingSchema(),
			},
		}
	}

	body := map[string]any{
		"model": g.opts.Model,
		"messages": []any{
			map[string]any{"role": "system", "content": system},
			map[string]any{"role": "user", "content": string(payload)},
		},
		// Structured output is not optional. Parsing free text back into
		// numbered lines fails silently, and a silently mis-numbered line puts
		// one fact's wording on another fact.
		"response_format": responseFormat,
		// Generous for the size of the answer, because a reasoning model
		// spends this budget thinking first and a budget that runs out
		// returns finish_reason "length" with empty content — which reads as
		// a broken provider rather than a budget mistake (the same trap
		// internal/classify documents).
		"max_tokens": 4000,
		// The job is to reword a fixed set of sentences, not to explore. A low
		// temperature also makes the eval set mean something between runs.
		"temperature": 0.2,
	}

	var lastErr error
	for attempt := 1; attempt <= g.opts.MaxAttempts; attempt++ {
		got, err := g.call(ctx, body)
		if err == nil {
			g.succeeded()
			return got, nil
		}
		lastErr = err
		if attempt < g.opts.MaxAttempts {
			select {
			case <-ctx.Done():
				g.failed()
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * 250 * time.Millisecond):
			}
		}
	}
	g.failed()
	return nil, fmt.Errorf("polish after %d attempts: %w", g.opts.MaxAttempts, lastErr)
}

func rewordingSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"lines": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"i":    map[string]any{"type": "integer"},
						"text": map[string]any{"type": "string"},
					},
					"required":             []string{"i", "text"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"lines"},
		"additionalProperties": false,
	}
}

func (g *Gateway) call(ctx context.Context, body map[string]any) (map[int]string, error) {
	ctx, cancel := context.WithTimeout(ctx, g.opts.Timeout)
	defer cancel()

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(g.opts.BaseURL, "/")+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.opts.APIKey)

	res, err := g.opts.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call model: %w", err)
	}
	defer res.Body.Close()

	out, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("model returned %d: %s", res.StatusCode, clip(string(out), 200))
	}

	var envelope struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(out, &envelope); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(envelope.Choices) == 0 {
		return nil, fmt.Errorf("model returned no choices")
	}
	choice := envelope.Choices[0]
	content := strings.TrimSpace(choice.Message.Content)
	if content == "" {
		return nil, fmt.Errorf("model returned empty content (finish_reason %q)", choice.FinishReason)
	}

	var answer struct {
		Lines []struct {
			I    int    `json:"i"`
			Text string `json:"text"`
		} `json:"lines"`
	}
	if err := json.Unmarshal([]byte(strip(content)), &answer); err != nil {
		return nil, fmt.Errorf("decode rewording: %w", err)
	}
	got := map[int]string{}
	for _, l := range answer.Lines {
		got[l.I] = l.Text
	}
	return got, nil
}

// strip removes a fenced code block, which some providers wrap JSON in even
// when asked not to.
func strip(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[i+1:]
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "```"))
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func (g *Gateway) breakerOpen() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.openedAt.IsZero() {
		return false
	}
	if time.Since(g.openedAt) < g.opts.BreakerCooldown {
		return true
	}
	// One request is let through to find out whether it is back.
	g.openedAt, g.failures = time.Time{}, g.opts.BreakerThreshold-1
	return false
}

func (g *Gateway) failed() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.failures++
	if g.failures >= g.opts.BreakerThreshold {
		g.openedAt = time.Now()
	}
}

func (g *Gateway) succeeded() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.failures, g.openedAt = 0, time.Time{}
}

// Cached wraps a Polisher in a small in-process cache.
//
// Two reasons, and the second is the important one. A share kit is a public
// GET that is already cached for five minutes at the edge, so a model call per
// request would be paid many times for one answer. And an answer that is
// stable for the life of the cache means two people opening the same issue
// read the same post, which is the determinism ADR 0020 valued and the thing a
// model most easily takes away.
func Cached(p Polisher, ttl time.Duration, max int) Polisher {
	if p == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if max <= 0 {
		max = 512
	}
	return &cache{inner: p, ttl: ttl, max: max, entries: map[string]entry{}}
}

type entry struct {
	got map[int]string
	at  time.Time
}

type cache struct {
	inner Polisher
	ttl   time.Duration
	max   int

	mu      sync.Mutex
	entries map[string]entry
}

func (c *cache) Describe() (string, string) { return c.inner.Describe() }

func (c *cache) Reword(ctx context.Context, lang share.Lang, ch share.Channel, lines []Segment) (map[int]string, error) {
	k := cacheKey(lang, ch, lines)

	c.mu.Lock()
	if e, ok := c.entries[k]; ok && time.Since(e.at) < c.ttl {
		c.mu.Unlock()
		return clone(e.got), nil
	}
	c.mu.Unlock()

	got, err := c.inner.Reword(ctx, lang, ch, lines)
	if err != nil {
		// Failures are not cached: a provider that recovers in a second
		// should be used in a second.
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.max {
		// Crude, and correct for the purpose: the cache is a cost control, not
		// a store, so emptying it loses nothing but money.
		c.entries = map[string]entry{}
	}
	c.entries[k] = entry{got: clone(got), at: time.Now()}
	return clone(got), nil
}

// cacheKey is the hash of exactly what would be sent, so a changed sentence,
// a changed language or a changed channel is a different question.
func cacheKey(lang share.Lang, ch share.Channel, lines []Segment) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s", lang, ch)
	for _, l := range lines {
		fmt.Fprintf(h, "|%d\x00%s", l.Index, l.Text)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func clone(m map[int]string) map[int]string {
	if m == nil {
		return nil
	}
	out := make(map[int]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
