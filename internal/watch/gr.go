// Package watch follows document sources for the records the roadmap is
// blocked on: new Government Resolutions, and orders in the pothole PIL.
//
// A watcher archives what it finds and flags the items whose text mentions a
// watched term. It never turns a document into a legal constant: that needs a
// human to read the primary text.
package watch

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// GRSource is the Maharashtra GR mirror on the Internet Archive. The state
// portal's own search is captcha-gated; the mirror's is a plain GET.
const (
	GRIdentifierPrefix = "in.gov.maharashtra.gr."
	searchEndpoint     = "https://archive.org/advancedsearch.php"
	downloadBase       = "https://archive.org/download/"
	metadataBase       = "https://archive.org/metadata/"
)

// DefaultKeywords are the terms worth an alert. They cover the open questions:
// the RTI fee (Q2), defect liability periods (Q8), and the pothole directions
// (Q28), in Marathi and English.
var DefaultKeywords = []string{
	"माहिती अधिकार", // Right to Information
	"दोष दायित्व",   // defect liability
	"खड्डे",         // potholes
	"लोकसेवा हक्क",  // Right to Public Services
	"defect liability",
	"right to information",
	"pothole",
}

// Item is one GR seen in the mirror.
type Item struct {
	Identifier string
	Sanketank  string
	Title      string
	AddedAt    time.Time
}

// Metadata describes an item's files.
type Metadata struct {
	Title    string
	IssuedOn *time.Time
	PDFName  string
	TextName string
}

func parseSearchResponse(body []byte) ([]Item, error) {
	var payload struct {
		Response struct {
			NumFound int `json:"numFound"`
			Docs     []struct {
				Identifier string `json:"identifier"`
				Title      string `json:"title"`
				AddedDate  string `json:"addeddate"`
			} `json:"docs"`
		} `json:"response"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("gr search: decode: %w", err)
	}

	items := make([]Item, 0, len(payload.Response.Docs))
	for _, doc := range payload.Response.Docs {
		if !strings.HasPrefix(doc.Identifier, GRIdentifierPrefix) {
			continue
		}
		item := Item{
			Identifier: doc.Identifier,
			Sanketank:  strings.TrimPrefix(doc.Identifier, GRIdentifierPrefix),
			Title:      doc.Title,
		}
		if t, err := time.Parse(time.RFC3339, doc.AddedDate); err == nil {
			item.AddedAt = t.UTC()
		}
		items = append(items, item)
	}
	return items, nil
}

func parseMetadata(body []byte) (Metadata, error) {
	var payload struct {
		Metadata struct {
			Title string `json:"title"`
			Date  string `json:"date"`
		} `json:"metadata"`
		Files []struct {
			Name   string `json:"name"`
			Format string `json:"format"`
		} `json:"files"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Metadata{}, fmt.Errorf("gr metadata: decode: %w", err)
	}

	md := Metadata{Title: payload.Metadata.Title}
	for _, f := range payload.Files {
		switch {
		case strings.HasSuffix(strings.ToLower(f.Name), ".pdf") && md.PDFName == "":
			md.PDFName = f.Name
		case strings.HasSuffix(strings.ToLower(f.Name), "_djvu.txt") && md.TextName == "":
			md.TextName = f.Name
		}
	}
	if md.PDFName == "" {
		return Metadata{}, fmt.Errorf("gr metadata: item has no PDF")
	}
	// The mirror records the GR's own date as dd-mm-yyyy.
	if d := strings.TrimSpace(payload.Metadata.Date); d != "" {
		if t, err := time.Parse("02-01-2006", d); err == nil {
			utc := t.UTC()
			md.IssuedOn = &utc
		}
	}
	return md, nil
}

func downloadURL(identifier, file string) string {
	return downloadBase + identifier + "/" + file
}

func metadataURL(identifier string) string {
	return metadataBase + identifier
}

// searchURL asks for the most recently added items, newest first.
func searchURL(rows int) string {
	return fmt.Sprintf(
		"%s?q=identifier%%3A%s*&fl%%5B%%5D=identifier&fl%%5B%%5D=addeddate&fl%%5B%%5D=title"+
			"&sort%%5B%%5D=addeddate+desc&rows=%d&page=1&output=json",
		searchEndpoint, strings.ReplaceAll(GRIdentifierPrefix, ".", "."), rows)
}

// matchKeywords returns the watched terms present in the text, in watch-list
// order. Matching is case-insensitive.
func matchKeywords(text string, keywords []string) []string {
	lower := strings.ToLower(text)
	var hits []string
	for _, k := range keywords {
		if strings.Contains(lower, strings.ToLower(k)) {
			hits = append(hits, k)
		}
	}
	return hits
}
