package notify

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebhookPostsTitleAndBody(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := NewWebhook(srv.URL, srv.Client())
	err := n.Send(context.Background(), Message{
		Title: "2 changes in bmc_roads_api",
		Lines: []string{"Road A|R/S: status In Progress -> Completed"},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if !strings.Contains(gotBody, "2 changes in bmc_roads_api") {
		t.Errorf("title missing from payload: %s", gotBody)
	}
	if !strings.Contains(gotBody, "status In Progress -> Completed") {
		t.Errorf("body missing from payload: %s", gotBody)
	}
}

func TestWebhookReportsFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := NewWebhook(srv.URL, srv.Client())
	if err := n.Send(context.Background(), Message{Title: "x"}); err == nil {
		t.Fatal("a failing webhook must report an error")
	}
}

func TestLoggerNotifierNeverFails(t *testing.T) {
	n := NewLogger(nil)
	if err := n.Send(context.Background(), Message{Title: "hello", Lines: []string{"a"}}); err != nil {
		t.Fatalf("logging notifier should not fail: %v", err)
	}
}

func TestNewFromConfigFallsBackToLoggingWithoutAWebhook(t *testing.T) {
	n := New("", nil, nil)
	if _, ok := n.(*Logger); !ok {
		t.Errorf("expected the logging notifier, got %T", n)
	}
}

func TestMessageTruncatesVeryLongChangeLists(t *testing.T) {
	lines := make([]string, 500)
	for i := range lines {
		lines[i] = "a change"
	}
	m := Message{Title: "many", Lines: lines}

	text := m.Text(10)

	if strings.Count(text, "a change") > 10 {
		t.Errorf("expected at most 10 lines, got %d", strings.Count(text, "a change"))
	}
	if !strings.Contains(text, "490 more") {
		t.Errorf("the remainder should be counted: %s", text)
	}
}
