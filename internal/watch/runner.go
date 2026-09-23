package watch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/ingest"
)

// Archive is the object-storage surface the watcher needs.
type Archive interface {
	Put(ctx context.Context, key string, body []byte, contentType string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// GRRecord is one Government Resolution as stored.
type GRRecord struct {
	Sanketank  string
	Title      string
	IssuedOn   *time.Time
	DocumentID string
	Keywords   []string
}

// GRStore persists what the watcher finds.
type GRStore interface {
	KnownGRs(ctx context.Context, sanketanks []string) (map[string]bool, error)
	SaveGR(ctx context.Context, rec GRRecord) error
	SaveRawDocument(ctx context.Context, doc ingest.RawDocument) (string, error)
}

// Hit is a resolution whose text mentions a watched term.
type Hit struct {
	Sanketank string
	Title     string
	IssuedOn  *time.Time
	Keywords  []string
}

// Result summarises a run.
type Result struct {
	Seen int
	New  int
	Hits []Hit
}

// GRWatcher archives newly published Government Resolutions.
type GRWatcher struct {
	Fetcher *ingest.Fetcher
	Archive Archive
	Store   GRStore
	Log     *slog.Logger

	// Limit caps how many new resolutions one run downloads.
	Limit int
	// Rows is how many recent items to examine per run.
	Rows     int
	Keywords []string

	SearchURL    string
	MetadataBase string
	DownloadBase string
}

// NewGRWatcher returns a watcher with the mirror's defaults.
func NewGRWatcher(f *ingest.Fetcher, arch Archive, store GRStore) *GRWatcher {
	return &GRWatcher{
		Fetcher:      f,
		Archive:      arch,
		Store:        store,
		Limit:        25,
		Rows:         100,
		Keywords:     DefaultKeywords,
		SearchURL:    searchEndpoint,
		MetadataBase: metadataBase,
		DownloadBase: downloadBase,
	}
}

func (w *GRWatcher) log() *slog.Logger {
	if w.Log != nil {
		return w.Log
	}
	return slog.Default()
}

// Run looks at the most recently added resolutions, archives the ones not seen
// before, and reports those mentioning a watched term. One bad item does not
// stop the rest.
func (w *GRWatcher) Run(ctx context.Context) (Result, error) {
	var result Result

	searchRes, err := w.Fetcher.Fetch(ctx, w.searchURL())
	if err != nil {
		return result, fmt.Errorf("gr search: %w", err)
	}
	items, err := parseSearchResponse(searchRes.Body)
	if err != nil {
		return result, err
	}
	result.Seen = len(items)
	if len(items) == 0 {
		return result, nil
	}

	sanketanks := make([]string, 0, len(items))
	for _, item := range items {
		sanketanks = append(sanketanks, item.Sanketank)
	}
	known, err := w.Store.KnownGRs(ctx, sanketanks)
	if err != nil {
		return result, fmt.Errorf("known resolutions: %w", err)
	}

	var errs []error
	for _, item := range items {
		if known[item.Sanketank] {
			continue
		}
		if w.Limit > 0 && result.New >= w.Limit {
			w.log().Info("gr watcher reached its per-run limit", "limit", w.Limit)
			break
		}
		hit, err := w.collect(ctx, item)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", item.Sanketank, err))
			continue
		}
		result.New++
		if hit != nil {
			result.Hits = append(result.Hits, *hit)
		}
	}
	return result, errors.Join(errs...)
}

func (w *GRWatcher) collect(ctx context.Context, item Item) (*Hit, error) {
	metaRes, err := w.Fetcher.Fetch(ctx, w.MetadataBase+item.Identifier)
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}
	md, err := parseMetadata(metaRes.Body)
	if err != nil {
		return nil, err
	}

	pdfRes, err := w.Fetcher.Fetch(ctx, w.DownloadBase+item.Identifier+"/"+md.PDFName)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}

	key := archive.KeyFor("maha_gr", pdfRes.RetrievedAt, pdfRes.SHA256, ".pdf")
	exists, err := w.Archive.Exists(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("archive lookup: %w", err)
	}
	if !exists {
		if err := w.Archive.Put(ctx, key, pdfRes.Body, "application/pdf"); err != nil {
			return nil, fmt.Errorf("archive write: %w", err)
		}
	}

	docID, err := w.Store.SaveRawDocument(ctx, ingest.RawDocument{
		SourceID:    "maha_gr_archive",
		Endpoint:    "gr_pdf",
		URL:         w.DownloadBase + item.Identifier + "/" + md.PDFName,
		ArchiveKey:  key,
		SHA256:      pdfRes.SHA256,
		ContentType: "application/pdf",
		Bytes:       int64(len(pdfRes.Body)),
		RetrievedAt: pdfRes.RetrievedAt,
		CapturedBy:  "ingester:maha_gr_archive",
	})
	if err != nil {
		return nil, fmt.Errorf("save document: %w", err)
	}

	title := md.Title
	if strings.TrimSpace(title) == "" {
		title = item.Title
	}

	// The mirror's extracted text is enough to flag a resolution for reading.
	// It is never enough to encode a legal constant.
	var keywords []string
	if md.TextName != "" {
		if textRes, err := w.Fetcher.Fetch(ctx, w.DownloadBase+item.Identifier+"/"+md.TextName); err == nil {
			keywords = matchKeywords(string(textRes.Body), w.Keywords)
		} else {
			w.log().Debug("no extracted text", "sanketank", item.Sanketank, "error", err.Error())
		}
	}

	record := GRRecord{
		Sanketank:  item.Sanketank,
		Title:      title,
		IssuedOn:   md.IssuedOn,
		DocumentID: docID,
		Keywords:   keywords,
	}
	if err := w.Store.SaveGR(ctx, record); err != nil {
		return nil, fmt.Errorf("save resolution: %w", err)
	}

	w.log().Info("resolution archived",
		"sanketank", item.Sanketank, "issued_on", md.IssuedOn, "keywords", keywords)

	if len(keywords) == 0 {
		return nil, nil
	}
	return &Hit{
		Sanketank: item.Sanketank,
		Title:     title,
		IssuedOn:  md.IssuedOn,
		Keywords:  keywords,
	}, nil
}

func (w *GRWatcher) searchURL() string {
	rows := w.Rows
	if rows <= 0 {
		rows = 100
	}
	if w.SearchURL == searchEndpoint {
		return searchURL(rows)
	}
	// A test or mirror override keeps the same query shape.
	return fmt.Sprintf("%s?q=identifier%%3A%s*&fl%%5B%%5D=identifier&fl%%5B%%5D=addeddate"+
		"&fl%%5B%%5D=title&sort%%5B%%5D=addeddate+desc&rows=%d&page=1&output=json",
		w.SearchURL, GRIdentifierPrefix, rows)
}
