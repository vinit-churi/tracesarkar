package api

import (
	"bytes"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/vinit-churi/tracesarkar/internal/share"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// unfurlMark is where the per-issue metadata goes. It is a comment in the
// source file so that the page remains a working page on its own, served by
// anything, with no template step.
const unfurlMark = "<!-- unfurl:"

// issuePageHandler serves the permalink shell with metadata for that issue.
//
// [D130] made the shell the same bytes for every id, so that a crawler reached
// nothing before the gate had been consulted. An unfurl needs per-issue
// metadata, so the gate is now consulted *earlier* rather than skipped: the
// issue is looked up before a byte is written, an unknown id is a 404 here
// rather than a sentence the page's own JavaScript draws, and the image is
// named only when publish.Derive says there is one.
//
// That matters more here than on the page itself. An unfurl image is fetched
// by WhatsApp or X and cached on their servers, out of reach of any later
// correction — so a photograph the detector held must not be offered to them,
// not even once.
func (s *Server) issuePageHandler(shell []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Without a database the shell alone is a working page, exactly as
		// every other optional surface degrades.
		if s.publicIssues == nil {
			writePage(w, shell, http.StatusOK)
			return
		}

		v, found, err := s.publicIssues.PublicIssue(r.Context(), r.PathValue("id"))
		if err != nil {
			s.log.Error("could not read an issue for its page", "error", err.Error())
			// The shell still works: its own fetch will report the failure in
			// words. Metadata is what is lost, which is the right thing to
			// lose.
			writePage(w, shell, http.StatusOK)
			return
		}
		if !found {
			writePage(w, bytes.Replace(shell, []byte(unfurlMark), []byte(noSuchIssueMeta+"\n"+unfurlMark), 1),
				http.StatusNotFound)
			return
		}

		writePage(w, injectUnfurl(shell, s.unfurlFor(r, v)), http.StatusOK)
	})
}

func writePage(w http.ResponseWriter, body []byte, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Shorter than the shell's old ten minutes, because the page now carries
	// facts that change: a status, a sighting count, a clock.
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

const noSuchIssueMeta = `<meta property="og:title" content="No such issue · TraceSarkar">`

// unfurl is the metadata block for one issue.
type unfurl struct {
	Title       string
	Description string
	URL         string
	// Image is empty when no photograph passed the gate. The card type steps
	// down with it, because a large card with no image renders a blank panel.
	Image string
}

func (s *Server) unfurlFor(r *http.Request, v store.PublicIssueView) unfurl {
	f, d := s.factsOnly(r, v)

	u := unfurl{
		Title: share.Title(f) + " · TraceSarkar",
		URL:   s.permalink(r, v.IssueID),
	}
	if d.ImageKey != "" {
		u.Image = s.origin(r) + "/v1/public/issue/" + v.IssueID + "/card"
	}
	u.Description = unfurlDescription(v, f)
	return u
}

// unfurlDescription is the sentence most people who see this link will read,
// and the one they will not click through to correct.
//
// It states what the platform knows and stops. The status wording is the
// page's own, so an authority's claim and a resident's photograph never read
// the same (G14), and the deadline appears only with the document that creates
// it — the same rule the page and the post follow.
func unfurlDescription(v store.PublicIssueView, f share.Facts) string {
	parts := []string{sentence(label(v.Status))}
	if v.Sightings > 1 {
		parts = append(parts, fmt.Sprintf("Seen %d times.", v.Sightings))
	}
	if !f.Lodged {
		parts = append(parts, "Nobody has recorded telling the authority yet.")
	} else if f.DeadlineHours > 0 && f.DeadlineShortRef != "" {
		if f.DeadlineState == "elapsed" {
			parts = append(parts, fmt.Sprintf("The %d hours to attend it (%s) have passed.",
				f.DeadlineHours, f.DeadlineShortRef))
		} else {
			parts = append(parts, fmt.Sprintf("%d hours to attend it (%s).",
				f.DeadlineHours, f.DeadlineShortRef))
		}
	} else {
		parts = append(parts, "The authority has been told.")
	}
	parts = append(parts, "Every fact here carries the document it came from.")

	out := strings.Join(parts, " ")
	// The gate the generated post already passes, applied to the one line a
	// third party will cache.
	if vs := share.Check(out); len(vs) > 0 {
		return "A civic problem in the Mumbai Metropolitan Region, with the " +
			"documents behind every fact on the page."
	}
	return out
}

// sentence ends a status label so it can be joined with others.
//
// The labels are written for a chip on a page, where a full stop would be
// wrong; here they are the first sentence of a paragraph, and "Open Nobody has
// recorded telling the authority yet" is what joining them without one reads
// like.
func sentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasSuffix(s, ".") || strings.HasSuffix(s, "!") ||
		strings.HasSuffix(s, "?") {
		return s
	}
	return s + "."
}

// injectUnfurl writes the block into the shell.
//
// Every value comes from a database this process does not control the contents
// of, so each is HTML-escaped: an unescaped quote would close the attribute
// and the rest would be markup.
func injectUnfurl(shell []byte, u unfurl) []byte {
	var b strings.Builder
	meta := func(attr, key, val string) {
		if val == "" {
			return
		}
		fmt.Fprintf(&b, "<meta %s=%q content=%q>\n", attr, key, html.EscapeString(val))
	}
	meta("property", "og:type", "article")
	meta("property", "og:site_name", "TraceSarkar")
	meta("property", "og:title", u.Title)
	meta("property", "og:description", u.Description)
	meta("property", "og:url", u.URL)
	meta("property", "og:image", u.Image)
	if u.Image != "" {
		meta("property", "og:image:alt",
			"The photograph, with faces and number plates obscured, and the facts printed under it")
		meta("name", "twitter:card", "summary_large_image")
	} else {
		meta("name", "twitter:card", "summary")
	}
	meta("name", "twitter:title", u.Title)
	meta("name", "twitter:description", u.Description)

	out := bytes.Replace(shell, []byte(unfurlMark), []byte(b.String()+unfurlMark), 1)
	// The browser tab, for the reader who actually opens it. The page's own
	// script sets this again once the facts load; this is what a crawler and a
	// slow connection see.
	return bytes.Replace(out,
		[]byte("<title>A civic issue · TraceSarkar</title>"),
		[]byte("<title>"+html.EscapeString(u.Title)+"</title>"), 1)
}

// origin is the scheme and host this request arrived on. An unfurl image has
// to be an absolute URL: the crawler fetching it has no page to resolve a
// relative path against.
func (s *Server) origin(r *http.Request) string {
	return strings.TrimSuffix(s.permalink(r, ""), "/issue/")
}
