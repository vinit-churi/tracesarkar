package main

import "testing"

// A manifest is edited by hand, so it has to tolerate what a person leaves in
// it: comments explaining why a document is on the list, blank lines between
// groups, and trailing spaces from a copied URL.
func TestParseManifest(t *testing.T) {
	in := `
# BMC Section 4 handbooks
https://portal.mcgm.gov.in/a.pdf

  https://portal.mcgm.gov.in/b%20c.pdf  
# a comment after a url
https://portal.mcgm.gov.in/a.pdf
`
	got := parseManifest(in)
	want := []string{
		"https://portal.mcgm.gov.in/a.pdf",
		"https://portal.mcgm.gov.in/b%20c.pdf",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d url(s) %v, want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("url %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// The archive key ends in the extension so that a human opening the bucket can
// tell a PDF from a scan of one. The URL is the better source for it: BMC's
// portal serves several of these as application/octet-stream.
func TestExtensionFor(t *testing.T) {
	cases := []struct {
		url, contentType, want string
	}{
		{"https://x/RTI%2017%20Manual.pdf", "application/octet-stream", ".pdf"},
		{"https://x/a.PDF", "", ".pdf"},
		{"https://x/noext", "application/pdf", ".pdf"},
		{"https://x/noext", "text/html; charset=utf-8", ".html"},
		{"https://x/data.json?q=1", "", ".json"},
		{"https://x/noext", "", ".bin"},
	}
	for _, c := range cases {
		if got := extensionFor(c.url, c.contentType); got != c.want {
			t.Errorf("extensionFor(%q, %q) = %q, want %q",
				c.url, c.contentType, got, c.want)
		}
	}
}
