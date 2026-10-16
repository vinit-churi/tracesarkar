package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

// The permalink is the one line of a generated post that a reader follows to
// check everything else, and it is burned into an image and forwarded beyond
// any page this platform can correct.
//
// It was built from r.Host, which is a header. On a public, unauthenticated
// GET that means a caller chooses it: send Host: evil.example and the
// platform generates a post saying "every fact above is on this page" above a
// link to a site the caller controls, ready to forward. It also means a post
// forwarded today dies if the service's hostname ever changes, because
// nothing can reach a WhatsApp forward to fix it.
func TestThePermalinkComesFromConfigurationAndNotFromAHeader(t *testing.T) {
	f := &fakePublicIssues{view: anIssue(), found: true}
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		PublicIssues: f, Blobs: &fakeStore{},
		Handles:       &fakeHandles{handles: []share.Handle{bmcHandle()}},
		PublicBaseURL: "https://tracesarkar.org",
		Token:         "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, spoofed := range []string{"evil.example", "evil.example:8443", "tracesarkar.org.evil.example"} {
		t.Run(spoofed, func(t *testing.T) {
			req := httptest.NewRequest("GET",
				"/v1/public/issue/"+f.view.IssueID+"/share", nil)
			req.Host = spoofed
			req.Header.Set("X-Forwarded-Proto", "https")
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), spoofed) {
				t.Errorf("the kit carries the spoofed host %q:\n%s", spoofed, rec.Body.String())
			}

			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			posts := body["posts"].(map[string]any)
			text := posts["en.whatsapp"].(map[string]any)["text"].(string)
			want := "https://tracesarkar.org/issue/" + f.view.IssueID
			if !strings.Contains(text, want) {
				t.Errorf("the post does not carry %q:\n%s", want, text)
			}
		})
	}
}

// The unfurl is fetched by a third party and cached out of reach of any later
// correction, so og:url has to be the canonical one too.
func TestTheUnfurlCarriesTheConfiguredURL(t *testing.T) {
	f := &fakePublicIssues{view: anIssue(), found: true}
	srv, err := New(Options{
		Reports: &fakeReports{}, Media: &fakeBlobs{},
		PublicIssues: f, Blobs: &fakeStore{},
		PublicBaseURL: "https://tracesarkar.org/",
		Token:         "test-token", Account: "field-kit",
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/issue/"+f.view.IssueID+"/", nil)
	req.Host = "evil.example"
	req.Header.Set("User-Agent", "WhatsApp/2.23")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	page := rec.Body.String()

	if strings.Contains(page, "evil.example") {
		t.Errorf("the unfurl carries the spoofed host:\n%s", page)
	}
	if !strings.Contains(page, "https://tracesarkar.org/issue/"+f.view.IssueID) {
		t.Errorf("og:url is not the configured URL:\n%s", page)
	}
	// A trailing slash in configuration must not double up.
	if strings.Contains(page, "tracesarkar.org//") {
		t.Errorf("the configured trailing slash doubled:\n%s", page)
	}
}

// Without configuration the request host is still used, because a local run
// has no canonical URL and refusing would break development. The startup log
// says what that costs.
func TestWithoutConfigurationThePermalinkFallsBackToTheRequestHost(t *testing.T) {
	f := &fakePublicIssues{view: anIssue(), found: true}
	srv := shareServer(t, f, &fakeHandles{})
	req := httptest.NewRequest("GET", "/v1/public/issue/"+f.view.IssueID+"/share", nil)
	req.Host = "localhost:8080"
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "http://localhost:8080/issue/") {
		t.Errorf("the local permalink is not the request host:\n%s", rec.Body.String())
	}
}

// A misconfigured base must fail at startup, not put a broken link in front
// of every reader.
func TestABaseURLThatIsNotAURLIsAStartupError(t *testing.T) {
	for _, bad := range []string{"javascript:alert(1)", "tracesarkar.org", "ftp://x/y", "https://"} {
		if _, err := New(Options{
			Reports: &fakeReports{}, Media: &fakeBlobs{},
			Token: "t", Account: "a", PublicBaseURL: bad,
		}); err == nil {
			t.Errorf("PublicBaseURL %q was accepted", bad)
		}
	}
}
