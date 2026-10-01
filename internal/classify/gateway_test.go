package classify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// A minimal OpenAI-compatible reply carrying our schema in the message content.
func gatewayReply(t *testing.T, r Result) string {
	t.Helper()
	payload, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(map[string]any{
		"choices": []any{map[string]any{
			"message": map[string]any{"content": string(payload)},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestTheGatewaySendsTheImageAndTheSchema(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(gatewayReply(t, Result{
			Category: "road_defect", Subcategory: "pothole", Severity: "high",
			IsCivicIssue: true, ImageQuality: "good", Confidence: 0.9,
		})))
	}))
	defer srv.Close()

	c := NewGateway(GatewayOptions{
		BaseURL: srv.URL, APIKey: "test", Model: "some/model",
		HTTPClient: srv.Client(),
	})

	res, err := c.Classify(context.Background(), []byte("jpeg-bytes"), "Borivali West")
	if err != nil {
		t.Fatalf("Classify: %v", err)
	}
	if res.Subcategory != "pothole" {
		t.Errorf("result not parsed: %+v", res)
	}

	// The schema must be sent, or the model is free to return prose and the
	// parse becomes a guess.
	if got["response_format"] == nil {
		t.Error("no response_format sent; structured output is not optional here")
	}
	raw, _ := json.Marshal(got)
	if !strings.Contains(string(raw), "image_url") {
		t.Error("the image never reached the request")
	}
	// The locality hint helps read signage; precise coordinates must not be
	// sent to a third party at all.
	if !strings.Contains(string(raw), "Borivali West") {
		t.Error("the coarse locality hint should be included")
	}
}

func TestPreciseCoordinatesAreNeverSent(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 1<<20)
		n, _ := r.Body.Read(b)
		body = string(b[:n])
		_, _ = w.Write([]byte(gatewayReply(t, Result{
			Category: "road_defect", Subcategory: "pothole",
			IsCivicIssue: true, ImageQuality: "good", Confidence: 0.9,
		})))
	}))
	defer srv.Close()

	c := NewGateway(GatewayOptions{BaseURL: srv.URL, APIKey: "k", Model: "m", HTTPClient: srv.Client()})
	_, _ = c.Classify(context.Background(), []byte("jpeg"), "Borivali West")

	// A latitude would look like this. Hard rule 4: exact position stays ours.
	for _, leak := range []string{"19.22", "72.85", "accuracy_m"} {
		if strings.Contains(body, leak) {
			t.Errorf("the request leaked %q to the provider", leak)
		}
	}
}

func TestAGatewayErrorIsReportedNotSwallowed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer srv.Close()

	c := NewGateway(GatewayOptions{BaseURL: srv.URL, APIKey: "k", Model: "m",
		HTTPClient: srv.Client(), MaxAttempts: 1})

	if _, err := c.Classify(context.Background(), []byte("jpeg"), ""); err == nil {
		t.Fatal("a 429 must surface; a swallowed error becomes an unclassified report nobody notices")
	}
}

func TestProseInsteadOfJSONIsAnErrorRatherThanAGuess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"It looks like a pothole to me."}}]}`))
	}))
	defer srv.Close()

	c := NewGateway(GatewayOptions{BaseURL: srv.URL, APIKey: "k", Model: "m", HTTPClient: srv.Client()})

	if _, err := c.Classify(context.Background(), []byte("jpeg"), ""); err == nil {
		t.Fatal("free text must not be parsed into a classification")
	}
}

func TestItGivesUpRatherThanHangingForever(t *testing.T) {
	// The handler must be released by the test, not by the client giving up:
	// the server does not observe a client-side timeout promptly, so waiting
	// on the request context here deadlocks srv.Close().
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release) // runs first: lets the handler return before Close waits

	c := NewGateway(GatewayOptions{BaseURL: srv.URL, APIKey: "k", Model: "m",
		HTTPClient: srv.Client(), Timeout: 150 * time.Millisecond, MaxAttempts: 1})

	start := time.Now()
	_, err := c.Classify(context.Background(), []byte("jpeg"), "")
	if err == nil {
		t.Fatal("a stalled provider must fail the call")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("gave up too late: %s", elapsed)
	}
}

func TestARefusedModelNeverReachesTheNetwork(t *testing.T) {
	// The policy guard is the point: a model that must not see this data
	// should fail before a request is built, not after it is sent.
	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
	}))
	defer srv.Close()

	c := NewGateway(GatewayOptions{
		BaseURL: srv.URL, APIKey: "k", Model: "stealth/pixel-canary",
		HTTPClient: srv.Client(),
		Policy: ModelPolicy{ID: "stealth/pixel-canary", NoTraining: "none",
			ZDR: "none", StructuredOutput: true, Vision: true},
		DataKind: CitizenPhotograph,
	})

	_, err := c.Classify(context.Background(), []byte("jpeg"), "")
	if err == nil {
		t.Fatal("a policy-refused model must not be called")
	}
	if called {
		t.Error("the photograph was sent to a provider the policy refuses")
	}
	var pe *PolicyError
	if !errors.As(err, &pe) {
		t.Errorf("the caller should be able to tell a policy refusal apart: %T", err)
	}
}

func TestJSONObjectModeSendsTheShapeInThePromptInstead(t *testing.T) {
	// Some providers accept only the weaker json_object mode, which
	// guarantees parseable JSON but enforces no schema. The shape then has to
	// reach the model through the prompt, and be validated here afterwards.
	var sent map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&sent)
		_, _ = w.Write([]byte(gatewayReply(t, Result{
			Category: "road_defect", Subcategory: "pothole",
			IsCivicIssue: true, ImageQuality: "good", Confidence: 0.9,
		})))
	}))
	defer srv.Close()

	c := NewGateway(GatewayOptions{
		BaseURL: srv.URL, APIKey: "k", Model: "deepseek-flash",
		HTTPClient: srv.Client(), SchemaMode: SchemaJSONObject,
	})
	if _, err := c.Classify(context.Background(), []byte("jpeg"), ""); err != nil {
		t.Fatalf("Classify: %v", err)
	}

	rf, _ := sent["response_format"].(map[string]any)
	if rf == nil || rf["type"] != "json_object" {
		t.Errorf("response_format: got %v, want json_object", sent["response_format"])
	}
	if rf["json_schema"] != nil {
		t.Error("json_object mode must not send a json_schema block; providers reject it")
	}
	// The shape has to be described somewhere the model will read it.
	raw, _ := json.Marshal(sent)
	for _, want := range []string{"is_civic_issue", "image_quality", "confidence"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("json_object mode did not describe %q to the model", want)
		}
	}
}

func TestAReplyMissingRequiredFieldsIsRejected(t *testing.T) {
	// Without schema enforcement the provider can return any JSON at all.
	// Accepting a reply with no category would route a complaint nowhere.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"rationale\":\"something\"}"}}]}`))
	}))
	defer srv.Close()

	c := NewGateway(GatewayOptions{BaseURL: srv.URL, APIKey: "k", Model: "m",
		HTTPClient: srv.Client(), SchemaMode: SchemaJSONObject, MaxAttempts: 1})

	_, err := c.Classify(context.Background(), []byte("jpeg"), "")
	if err == nil {
		t.Fatal("a reply with no category must be refused")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "category") {
		t.Errorf("the error should name what was missing: %v", err)
	}
}

func TestReasoningModelsGetEnoughRoomToAnswer(t *testing.T) {
	// A reasoning model spends max_tokens on reasoning first. Measured
	// against deepseek-flash: 577 reasoning tokens before a 177-token answer.
	// A 600-token budget returns finish_reason "length" and empty content —
	// which looks like a broken classifier rather than a budget mistake.
	var sent map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&sent)
		_, _ = w.Write([]byte(gatewayReply(t, Result{
			Category: "road_defect", Subcategory: "pothole",
			IsCivicIssue: true, ImageQuality: "good", Confidence: 0.9,
		})))
	}))
	defer srv.Close()

	c := NewGateway(GatewayOptions{BaseURL: srv.URL, APIKey: "k", Model: "m", HTTPClient: srv.Client()})
	if _, err := c.Classify(context.Background(), []byte("jpeg"), ""); err != nil {
		t.Fatal(err)
	}

	budget, _ := sent["max_tokens"].(float64)
	if budget < 2000 {
		t.Errorf("max_tokens is %v; a reasoning model will spend that before answering", budget)
	}
}

func TestAnEmptyAnswerIsReportedAsSuch(t *testing.T) {
	// finish_reason "length" with empty content is the reasoning-budget
	// failure. It must not surface as an unhelpful JSON parse error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"finish_reason":"length","message":{"content":""}}]}`))
	}))
	defer srv.Close()

	c := NewGateway(GatewayOptions{BaseURL: srv.URL, APIKey: "k", Model: "m",
		HTTPClient: srv.Client(), MaxAttempts: 1})

	_, err := c.Classify(context.Background(), []byte("jpeg"), "")
	if err == nil {
		t.Fatal("an empty answer must be an error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "token") {
		t.Errorf("the error should point at the token budget: %v", err)
	}
}
