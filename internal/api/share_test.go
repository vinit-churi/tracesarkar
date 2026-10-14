package api

import (
	"context"
	"encoding/json"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/share"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

type fakeHandles struct {
	handles []share.Handle
	err     error
}

func (f *fakeHandles) Handles(context.Context, string, string) ([]share.Handle, error) {
	return f.handles, f.err
}

func bmcHandle() share.Handle {
	return share.Handle{
		Platform: "x", Handle: "@mybmc", Kind: share.Institutional,
		SourceRef:   "https://example.test/bmc",
		RetrievedAt: time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC),
	}
}

func shareServer(t *testing.T, f *fakePublicIssues, h Handles) *Server {
	t.Helper()
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		PublicIssues: f, Blobs: &fakeStore{}, Handles: h,
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func getKit(t *testing.T, f *fakePublicIssues, h Handles) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	rec := httptest.NewRecorder()
	// No Authorization header: a share kit is for a public issue, and asking a
	// resident to sign in before they can send a post defeats the purpose.
	shareServer(t, f, h).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/v1/public/issue/"+f.view.IssueID+"/share", nil))
	var body map[string]any
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("body is not JSON: %v\n%s", err, rec.Body.String())
		}
	}
	return rec, body
}

func TestTheKitOffersEveryLanguageAndChannel(t *testing.T) {
	rec, body := getKit(t, &fakePublicIssues{view: anIssue(), found: true},
		&fakeHandles{handles: []share.Handle{bmcHandle()}})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	posts, ok := body["posts"].(map[string]any)
	if !ok {
		t.Fatalf("no posts in the kit: %s", rec.Body.String())
	}
	for _, want := range []string{"en.whatsapp", "en.x", "mr.whatsapp", "mr.x", "hi.whatsapp", "hi.x"} {
		p, ok := posts[want].(map[string]any)
		if !ok {
			t.Errorf("no post for %s", want)
			continue
		}
		if s, _ := p["text"].(string); strings.TrimSpace(s) == "" {
			t.Errorf("%s has empty text", want)
		}
	}
}

// Hard rule 4, at the surface. The kit is built from the same coarsened view
// the public page uses, so an exact coordinate cannot reach a post.
func TestTheKitNeverCarriesTheExactPosition(t *testing.T) {
	rec, _ := getKit(t, &fakePublicIssues{view: anIssue(), found: true}, &fakeHandles{})
	raw := rec.Body.String()
	for _, exact := range []string{"19.2309441", "72.8567123", "19.23094", "72.85671"} {
		if strings.Contains(raw, exact) {
			t.Errorf("the exact position reached the share kit: %s", exact)
		}
	}
}

// The reference number belongs to the person who lodged the complaint and
// leads back to their details on the authority's own portal.
func TestTheKitNeverCarriesTheComplaintReference(t *testing.T) {
	rec, _ := getKit(t, &fakePublicIssues{view: anIssue(), found: true}, &fakeHandles{})
	if strings.Contains(rec.Body.String(), reference) {
		t.Error("the complaint reference reached the share kit")
	}
}

// A registry that is down must not silently produce an untagged post that
// reads as finished — but it must not fail the whole kit either. The post is
// served without a handle and the kit says so.
func TestAHandleRegistryFailureCostsTheHandleNotTheKit(t *testing.T) {
	rec, body := getKit(t, &fakePublicIssues{view: anIssue(), found: true},
		&fakeHandles{err: context.DeadlineExceeded})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "@") {
		t.Error("a post was tagged although the registry could not be read")
	}
	if body["handles_unavailable"] != true {
		t.Error("the kit does not say the registry was unavailable")
	}
}

func TestAnUnknownIssueHasNoKit(t *testing.T) {
	// A real-looking id, because an empty one never reaches the handler: the
	// router collapses the doubled slash and redirects instead.
	rec, _ := getKit(t, &fakePublicIssues{
		view:  store.PublicIssueView{IssueID: "3f1a0f9c-0000-4000-8000-0000000000ff"},
		found: false,
	}, &fakeHandles{})
	if rec.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404", rec.Code)
	}
}

// The card is drawn on the redacted copy and on nothing else. If the gate
// withheld the photograph there is no card, even though the object exists and
// this server could read it.
func TestNoCardIsDrawnOnAPhotographTheGateWithheld(t *testing.T) {
	v := anIssue()
	v.Redaction.Published = false
	f := &fakePublicIssues{view: v, found: true}

	rec := httptest.NewRecorder()
	shareServer(t, f, &fakeHandles{}).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/v1/public/issue/"+v.IssueID+"/card", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404 for a held photograph", rec.Code)
	}
}

func TestTheCardIsServedAsAnImage(t *testing.T) {
	f := &fakePublicIssues{view: anIssue(), found: true}
	rec := httptest.NewRecorder()
	shareServer(t, f, &fakeHandles{}).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/v1/public/issue/"+f.view.IssueID+"/card", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("Content-Type is %q, want image/jpeg", ct)
	}
	img, err := jpeg.Decode(strings.NewReader(rec.Body.String()))
	if err != nil {
		t.Fatalf("the card is not a decodable jpeg: %v", err)
	}
	// The fact block is added below the photograph, so the card is always
	// taller than the 32x24 fixture it was drawn on.
	if b := img.Bounds(); b.Dy() <= 24 {
		t.Errorf("the card is %dx%d; the fact block is missing", b.Dx(), b.Dy())
	}
	_ = image.Rect
}

// Every fact a post states has to be reproducible from the page it links to.
func TestEveryPostLinksToThePageThatCarriesItsSources(t *testing.T) {
	_, body := getKit(t, &fakePublicIssues{view: anIssue(), found: true},
		&fakeHandles{handles: []share.Handle{bmcHandle()}})
	posts := body["posts"].(map[string]any)
	for name, raw := range posts {
		p := raw.(map[string]any)
		text, _ := p["text"].(string)
		if !strings.Contains(text, "/issue/"+anIssue().IssueID) {
			t.Errorf("%s does not link to the issue page:\n%s", name, text)
		}
	}
}

// An unreviewed translation is offered but never silently: the client has to
// be able to mark it so a user can read it before sending.
func TestTheKitSaysWhichLanguagesAreStillDrafts(t *testing.T) {
	_, body := getKit(t, &fakePublicIssues{view: anIssue(), found: true}, &fakeHandles{})
	posts := body["posts"].(map[string]any)
	if got := posts["en.x"].(map[string]any)["review"]; got != "reviewed" {
		t.Errorf("English is %v, want reviewed", got)
	}
	if got := posts["mr.x"].(map[string]any)["review"]; got != "draft" {
		t.Errorf("Marathi is %v, want draft", got)
	}
}

// 501 rather than a stack trace, matching every other optional surface.
func TestTheKitSaysSoWhenItIsNotConfigured(t *testing.T) {
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		Token: "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest("GET",
		"/v1/public/issue/3f1a0f9c-0000-4000-8000-000000000001/share", nil))
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("status %d, want 501", rec.Code)
	}
}

func TestTheCardIsOfferedInTheShapeTheChannelWants(t *testing.T) {
	f := &fakePublicIssues{view: anIssue(), found: true}
	srv := shareServer(t, f, &fakeHandles{})

	var feedH, storyH int
	for _, c := range []struct {
		query    string
		portrait bool
	}{
		{"", false},
		{"?shape=feed", false},
		{"?shape=story", true},
	} {
		rec := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rec, httptest.NewRequest("GET",
			"/v1/public/issue/"+f.view.IssueID+"/card"+c.query, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%q: status %d: %s", c.query, rec.Code, rec.Body.String())
		}
		img, err := jpeg.Decode(strings.NewReader(rec.Body.String()))
		if err != nil {
			t.Fatalf("%q: not a jpeg: %v", c.query, err)
		}
		b := img.Bounds()
		if c.portrait {
			storyH = b.Dy()
			if b.Dy() <= b.Dx() {
				t.Errorf("the story card is %dx%d, which is not portrait", b.Dx(), b.Dy())
			}
		} else {
			feedH = b.Dy()
		}
	}
	if feedH == 0 || storyH == 0 {
		t.Fatal("one of the shapes produced nothing")
	}
}

// A shape with no layout is a 400, not the default card: somebody asking for a
// frame is going to post the result into it.
func TestAnUnknownCardShapeIsRefused(t *testing.T) {
	f := &fakePublicIssues{view: anIssue(), found: true}
	rec := httptest.NewRecorder()
	shareServer(t, f, &fakeHandles{}).Handler().ServeHTTP(rec,
		httptest.NewRequest("GET", "/v1/public/issue/"+f.view.IssueID+"/card?shape=square", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status %d, want 400", rec.Code)
	}
}
