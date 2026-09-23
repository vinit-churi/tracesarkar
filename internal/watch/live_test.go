package watch

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/ingest"
)

// The defect-liability GR of 14 January 2019, whose transcription is in
// docs/01-research/sources/2026-09-18-pwd-dlp-gr-2019.md. Its OCR text is clean
// Devanagari, unlike the PDF's own text layer.
const dlpGRIdentifier = "in.gov.maharashtra.gr.201901141237173518"

func TestLiveGRTextIsFetchableAndMatchable(t *testing.T) {
	if os.Getenv("TRACESARKAR_LIVE") != "1" {
		t.Skip("set TRACESARKAR_LIVE=1 to run live mirror tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	f := ingest.NewFetcher(nil, ingest.DefaultUserAgent)

	metaRes, err := f.Fetch(ctx, metadataURL(dlpGRIdentifier))
	if err != nil {
		t.Fatalf("metadata: %v", err)
	}
	md, err := parseMetadata(metaRes.Body)
	if err != nil {
		t.Fatalf("parse metadata: %v", err)
	}
	if md.TextName == "" {
		t.Fatal("this item should have extracted text")
	}

	// The download URL redirects to a storage node; the fetcher must follow it.
	textRes, err := f.Fetch(ctx, downloadURL(dlpGRIdentifier, md.TextName))
	if err != nil {
		t.Fatalf("download text: %v", err)
	}
	if len(textRes.Body) < 1000 {
		t.Fatalf("suspiciously short text: %d bytes", len(textRes.Body))
	}

	hits := matchKeywords(string(textRes.Body), DefaultKeywords)
	if len(hits) == 0 {
		t.Fatalf("expected the defect-liability term to match; got none")
	}
	var found bool
	for _, h := range hits {
		if h == "दोष दायित्व" || h == "defect liability" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a defect-liability hit, got %v", hits)
	}
}
