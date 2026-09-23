package watch

import (
	"testing"
	"time"
)

const searchResponse = `{"responseHeader":{"status":0},"response":{"numFound":172649,"start":0,"docs":[
 {"identifier":"in.gov.maharashtra.gr.202609222028573048","addeddate":"2026-09-23T01:55:41Z","title":"Maharashtra GR: #202609222028573048"},
 {"identifier":"in.gov.maharashtra.gr.202609221226002554","addeddate":"2026-09-23T01:45:18Z","title":"Maharashtra GR: #202609221226002554"}
]}}`

func TestParseSearchResponseReadsIdentifiers(t *testing.T) {
	items, err := parseSearchResponse([]byte(searchResponse))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Sanketank != "202609222028573048" {
		t.Errorf("the sanketank is the identifier suffix: got %q", items[0].Sanketank)
	}
	if items[0].Identifier != "in.gov.maharashtra.gr.202609222028573048" {
		t.Errorf("identifier: got %q", items[0].Identifier)
	}
}

func TestParseSearchResponseRejectsGarbage(t *testing.T) {
	if _, err := parseSearchResponse([]byte("<html>")); err == nil {
		t.Error("expected an error for a non-JSON response")
	}
}

const metadataResponse = `{"dir":"/4/items/x","server":"ia800904.us.archive.org",
 "metadata":{"title":"Maharashtra GR: #201901141237173518","date":"14-01-2019",
             "addeddate":"2024-02-04 16:50:04"},
 "files":[{"name":"201901141237173518.epub","format":"EPUB"},
          {"name":"201901141237173518.pdf","format":"Text PDF","size":"3220662"},
          {"name":"201901141237173518_djvu.txt","format":"DjVuTXT","size":"15691"}]}`

func TestParseMetadataFindsThePDFAndText(t *testing.T) {
	md, err := parseMetadata([]byte(metadataResponse))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if md.PDFName != "201901141237173518.pdf" {
		t.Errorf("pdf: got %q", md.PDFName)
	}
	if md.TextName != "201901141237173518_djvu.txt" {
		t.Errorf("text: got %q", md.TextName)
	}
	if md.Title != "Maharashtra GR: #201901141237173518" {
		t.Errorf("title: got %q", md.Title)
	}
	want := time.Date(2019, 1, 14, 0, 0, 0, 0, time.UTC)
	if md.IssuedOn == nil || !md.IssuedOn.Equal(want) {
		t.Errorf("issued on: got %v, want %v", md.IssuedOn, want)
	}
}

func TestParseMetadataToleratesAMissingDate(t *testing.T) {
	md, err := parseMetadata([]byte(`{"metadata":{"title":"t"},"files":[{"name":"a.pdf","format":"Text PDF"}]}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if md.IssuedOn != nil {
		t.Errorf("expected no date, got %v", md.IssuedOn)
	}
}

func TestParseMetadataErrorsWithoutAPDF(t *testing.T) {
	_, err := parseMetadata([]byte(`{"metadata":{},"files":[{"name":"a.epub","format":"EPUB"}]}`))
	if err == nil {
		t.Error("an item with no PDF cannot be archived as a GR")
	}
}

func TestDownloadURL(t *testing.T) {
	got := downloadURL("in.gov.maharashtra.gr.123", "123.pdf")
	want := "https://archive.org/download/in.gov.maharashtra.gr.123/123.pdf"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMatchKeywordsFindsEnglishAndMarathiTerms(t *testing.T) {
	text := `शासन निर्णय: माहिती अधिकार अधिनियम शुल्क सुधारणा.
	         This resolution revises the fee payable with an application.`

	hits := matchKeywords(text, []string{"माहिती अधिकार", "defect liability", "fee"})

	if len(hits) != 2 {
		t.Fatalf("got %v, want the Marathi RTI term and 'fee'", hits)
	}
	if hits[0] != "माहिती अधिकार" || hits[1] != "fee" {
		t.Errorf("hits should keep watch-list order: %v", hits)
	}
}

func TestMatchKeywordsIsCaseInsensitive(t *testing.T) {
	if hits := matchKeywords("Defect Liability Period", []string{"defect liability"}); len(hits) != 1 {
		t.Errorf("expected a case-insensitive match, got %v", hits)
	}
}

func TestMatchKeywordsReturnsNothingWhenAbsent(t *testing.T) {
	if hits := matchKeywords("unrelated text", []string{"pothole"}); len(hits) != 0 {
		t.Errorf("got %v", hits)
	}
}

func TestDefaultKeywordsCoverTheOpenQuestions(t *testing.T) {
	// The watcher exists to catch the documents the roadmap is blocked on.
	joined := ""
	for _, k := range DefaultKeywords {
		joined += k + "|"
	}
	for _, needed := range []string{"माहिती अधिकार", "दोष दायित्व", "खड्डे"} {
		if !contains(joined, needed) {
			t.Errorf("watch list is missing %q", needed)
		}
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
