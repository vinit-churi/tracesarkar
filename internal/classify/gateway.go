package classify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Classifier turns a photograph into a routable result.
//
// One method, so a second provider is a second implementation rather than a
// rewrite — and so the eval harness can measure any of them identically.
type Classifier interface {
	Classify(ctx context.Context, image []byte, localityHint string) (Result, error)
}

// PolicyError is a refusal to call a model at all, distinguishable from a
// provider failure because the two need different responses: a policy refusal
// is a configuration mistake, not something to retry.
type PolicyError struct{ Err error }

func (e *PolicyError) Error() string { return e.Err.Error() }
func (e *PolicyError) Unwrap() error { return e.Err }

// GatewayOptions configures a client for an OpenAI-compatible gateway.
type GatewayOptions struct {
	BaseURL string
	APIKey  string
	Model   string
	// Policy is what the gateway reports about this model. DataKind is whose
	// data this client is allowed to send. Together they stop a model being
	// pointed at data it must not see.
	Policy   ModelPolicy
	DataKind DataKind

	PromptName  string
	Timeout     time.Duration
	MaxAttempts int
	HTTPClient  *http.Client
}

// Gateway calls a model through an OpenAI-compatible endpoint.
type Gateway struct {
	opts GatewayOptions
}

// NewGateway builds a client. Defaults are deliberately conservative: a
// classification that hangs is a report that never gets routed.
func NewGateway(o GatewayOptions) *Gateway {
	if o.Timeout == 0 {
		o.Timeout = 45 * time.Second
	}
	if o.MaxAttempts == 0 {
		o.MaxAttempts = 3
	}
	if o.PromptName == "" {
		o.PromptName = "road_defect_v1"
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{}
	}
	if o.DataKind == "" {
		o.DataKind = CitizenPhotograph // the safe default
	}
	return &Gateway{opts: o}
}

// Classify sends one photograph and returns the model's structured answer.
//
// The locality hint is deliberately coarse — a ward or suburb name. It helps
// the model read signage; precise coordinates are personal data and never
// leave this service (hard rule 4).
func (g *Gateway) Classify(ctx context.Context, image []byte, localityHint string) (Result, error) {
	// Before anything is built, let alone sent.
	if g.opts.Policy.ID != "" {
		if err := g.opts.Policy.AllowedFor(g.opts.DataKind); err != nil {
			return Result{}, &PolicyError{Err: err}
		}
	}

	prompt, err := LoadPrompt(g.opts.PromptName)
	if err != nil {
		return Result{}, err
	}

	userContent := []any{
		map[string]any{
			"type": "image_url",
			"image_url": map[string]any{
				"url": "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(image),
			},
		},
	}
	if localityHint != "" {
		userContent = append(userContent, map[string]any{
			"type": "text",
			"text": "This was photographed in " + localityHint + ".",
		})
	}

	body := map[string]any{
		"model": g.opts.Model,
		"messages": []any{
			map[string]any{"role": "system", "content": prompt.WithTaxonomy()},
			map[string]any{"role": "user", "content": userContent},
		},
		// Not optional. Parsing free text into a classification fails
		// silently, and a silently wrong category routes a complaint to a
		// desk that cannot act on it.
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "civic_classification",
				"strict": true,
				"schema": ResultSchema(),
			},
		},
		"max_tokens": 600,
	}

	var lastErr error
	for attempt := 1; attempt <= g.opts.MaxAttempts; attempt++ {
		res, err := g.call(ctx, body)
		if err == nil {
			return res, nil
		}
		lastErr = err
		if attempt < g.opts.MaxAttempts {
			select {
			case <-ctx.Done():
				return Result{}, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}
	}
	return Result{}, fmt.Errorf("classify after %d attempts: %w", g.opts.MaxAttempts, lastErr)
}

func (g *Gateway) call(ctx context.Context, body map[string]any) (Result, error) {
	ctx, cancel := context.WithTimeout(ctx, g.opts.Timeout)
	defer cancel()

	payload, err := json.Marshal(body)
	if err != nil {
		return Result{}, fmt.Errorf("encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.opts.BaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return Result{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.opts.APIKey)

	res, err := g.opts.HTTPClient.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("call model: %w", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return Result{}, fmt.Errorf("read response: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		// The body may carry the provider's own message; it must not carry
		// anything of ours, so it is safe to surface.
		return Result{}, fmt.Errorf("model returned %d: %s", res.StatusCode, truncate(string(raw), 200))
	}

	var envelope struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return Result{}, fmt.Errorf("decode response: %w", err)
	}
	if len(envelope.Choices) == 0 {
		return Result{}, fmt.Errorf("model returned no choices")
	}

	var out Result
	if err := json.Unmarshal([]byte(envelope.Choices[0].Message.Content), &out); err != nil {
		return Result{}, fmt.Errorf(
			"model did not return the schema — refusing to guess at prose: %w", err)
	}
	return out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
