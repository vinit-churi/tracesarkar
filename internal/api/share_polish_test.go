package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/polish"
	"github.com/vinit-churi/tracesarkar/internal/share"
)

// upper rewords every line by shouting it: a change the guard has no reason
// to refuse, so it proves the polished text reaches the response.
type upper struct {
	err   error
	fault bool
}

func (u *upper) Describe() (string, string) { return "test-model", "test-prompt@v1" }

func (u *upper) Reword(_ context.Context, _ share.Lang, _ share.Channel, lines []polish.Segment) (map[int]string, error) {
	if u.err != nil {
		return nil, u.err
	}
	out := map[int]string{}
	for _, l := range lines {
		if u.fault {
			// A number the platform never recorded.
			out[l.Index] = l.Text + " 9999"
			continue
		}
		out[l.Index] = l.Text
	}
	// Reword one line visibly, without touching a number or a span.
	for _, l := range lines {
		if strings.HasPrefix(l.Text, "Still open") {
			out[l.Index] = "This is still open."
		}
	}
	return out, nil
}

func polishServer(t *testing.T, f *fakePublicIssues, p polish.Polisher) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		PublicIssues: f, Blobs: &fakeStore{},
		Handles: &fakeHandles{handles: []share.Handle{bmcHandle()}},
		Polish:  p,
		Token:   "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func polishedKit(t *testing.T, p polish.Polisher) map[string]any {
	t.Helper()
	f := &fakePublicIssues{view: anIssue(), found: true}
	rec := httptest.NewRecorder()
	polishServer(t, f, p).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/v1/public/issue/"+f.view.IssueID+"/share", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	return body
}

func post(t *testing.T, body map[string]any, key string) map[string]any {
	t.Helper()
	posts, ok := body["posts"].(map[string]any)
	if !ok {
		t.Fatal("no posts in the kit")
	}
	p, ok := posts[key].(map[string]any)
	if !ok {
		t.Fatalf("no %s post", key)
	}
	return p
}

// Without a polisher the kit is exactly what it was, and every post says its
// words came from the template.
func TestWithoutAPolisherEveryPostSaysItIsTemplateWorded(t *testing.T) {
	body := polishedKit(t, nil)
	for _, key := range []string{"en.whatsapp", "mr.whatsapp", "hi.x"} {
		if got := post(t, body, key)["wording"]; got != string(share.WordedByTemplate) {
			t.Errorf("%s wording is %v, want %q", key, got, share.WordedByTemplate)
		}
	}
	if _, ok := body["polish"]; ok {
		t.Error("the response reports polish that is not configured")
	}
}

func TestAPolishedPostIsLabelledAndNamesTheModel(t *testing.T) {
	body := polishedKit(t, &upper{})
	p := post(t, body, "en.whatsapp")
	if got := p["wording"]; got != string(share.WordedByModel) {
		t.Errorf("wording is %v, want %q", got, share.WordedByModel)
	}
	if text, _ := p["text"].(string); !strings.Contains(text, "This is still open.") {
		t.Errorf("the reworded line is not in the post:\n%v", p["text"])
	}
	meta, ok := body["polish"].(map[string]any)
	if !ok {
		t.Fatal("the response does not say a model was used")
	}
	if meta["model"] != "test-model" || meta["prompt"] != "test-prompt@v1" {
		t.Errorf("polish metadata is %v", meta)
	}
}

// Polishing must never change the review state — in particular it must never
// make an unreviewed language look reviewed.
//
// Compared before against after rather than against a fixed expectation. The
// claim is about what polishing does to the field, not about which languages
// are drafts, and Marathi and Hindi stopped being drafts on 16 October 2026
// ([D175](../../docs/00-overview/05-decision-log.md)).
func TestPolishingDoesNotChangeTheReviewState(t *testing.T) {
	plain := polishedKit(t, nil)
	polished := polishedKit(t, &upper{})

	for _, lang := range share.Languages() {
		for _, ch := range share.Channels() {
			key := string(lang) + "." + string(ch)
			before, after := post(t, plain, key), post(t, polished, key)
			if before["review"] != after["review"] {
				t.Errorf("%s review went from %v to %v when a model reworded it",
					key, before["review"], after["review"])
			}
			// And the two really are different posts, or this proves nothing.
			if after["wording"] != string(share.WordedByModel) {
				t.Errorf("%s was not reworded, so the comparison is vacuous", key)
			}
		}
	}
	if a, b := plain["draft_languages"], polished["draft_languages"]; fmt.Sprint(a) != fmt.Sprint(b) {
		t.Errorf("draft_languages went from %v to %v when a model reworded the posts", a, b)
	}
}

// A model that invents a number costs nothing: the deterministic post is
// served and the kit still answers.
func TestAnInventedNumberFallsBackToTheTemplate(t *testing.T) {
	body := polishedKit(t, &upper{fault: true})
	for _, key := range []string{"en.whatsapp", "mr.x"} {
		p := post(t, body, key)
		if got := p["wording"]; got != string(share.WordedByTemplate) {
			t.Errorf("%s wording is %v, want %q", key, got, share.WordedByTemplate)
		}
		if text, _ := p["text"].(string); strings.Contains(text, "9999") {
			t.Errorf("%s published an invented number:\n%s", key, text)
		}
	}
}

// A provider that is down is not an outage of the share kit.
func TestAModelOutageStillReturnsAKit(t *testing.T) {
	body := polishedKit(t, &upper{err: errors.New("gateway timeout")})
	p := post(t, body, "en.whatsapp")
	if p["wording"] != string(share.WordedByTemplate) {
		t.Errorf("wording is %v", p["wording"])
	}
	if text, _ := p["text"].(string); text == "" {
		t.Error("the kit came back with no text")
	}
}

// The reader is about to send this under their own name. The page has to say
// who chose the words — and it must not let "a model reworded this" and "a
// native speaker has read this" read as the same statement.
func TestTheIssuePageDistinguishesAModelsWordingFromAReview(t *testing.T) {
	rec := httptest.NewRecorder()
	f := &fakePublicIssues{found: true, view: anIssue()}
	polishServer(t, f, nil).Handler().
		ServeHTTP(rec, httptest.NewRequest("GET", "/issue/"+f.view.IssueID+"/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	page := rec.Body.String()

	for _, promise := range []string{
		// The draft notice still turns on review, not on wording.
		`post.review === "draft"`,
		"has not yet been read by a native speaker",
		// And a model's wording is disclosed on its own terms.
		`post.wording === "model"`,
		"reworded by a language model",
		"checked against the",
	} {
		if !strings.Contains(page, promise) {
			t.Errorf("the issue page does not carry %q", promise)
		}
	}
}
