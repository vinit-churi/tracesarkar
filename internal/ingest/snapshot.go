package ingest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/works"
)

// Archive is the subset of the object store the runner needs.
type Archive interface {
	Put(ctx context.Context, key string, body []byte, contentType string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// FetchRecord is one row of fetch_log: every attempt, whatever the outcome.
type FetchRecord struct {
	SourceID    string
	Endpoint    string
	RequestedAt time.Time
	StatusCode  int
	Bytes       int64
	SHA256      string
	DocumentID  string
	Unchanged   bool
	DurationMS  int
	Error       string
}

// RawDocument is an archived artefact.
type RawDocument struct {
	SourceID    string
	Endpoint    string
	URL         string
	ArchiveKey  string
	SHA256      string
	ContentType string
	Bytes       int64
	RetrievedAt time.Time
	CapturedBy  string
}

// SnapshotWrite is the parsed result of one endpoint, with the changes observed
// against the previous snapshot.
type SnapshotWrite struct {
	SourceID   string
	Endpoint   string
	DocumentID string
	ObservedAt time.Time
	Records    []works.Record
	Changes    []works.Change
}

// Store is the database surface the runner needs.
type Store interface {
	LastDocumentSHA(ctx context.Context, sourceID, endpoint string) (string, error)
	RecordFetch(ctx context.Context, rec FetchRecord) error
	SaveRawDocument(ctx context.Context, doc RawDocument) (string, error)
	LoadWorks(ctx context.Context, sourceID, endpoint string) (map[string]works.Record, error)
	ApplySnapshot(ctx context.Context, in SnapshotWrite) error
	// MarkDocumentParsed records that a document was parsed and applied. Until
	// it is called, the bytes count as unseen and the next run retries them.
	MarkDocumentParsed(ctx context.Context, documentID string) error
}

// Endpoint is one URL within a source, with the parser for its shape.
type Endpoint struct {
	Name string
	URL  string
	// Alternate URLs are tried, in order, when the primary is not found. The
	// storm-water API needs this: its path carries the desilting season, and on
	// 1 January the new season is not published yet.
	Alternate []string
	Parse     works.Parser
	Ext       string
}

// Job is everything needed to snapshot one source.
type Job struct {
	SourceID  string
	Blocklist []string
	Endpoints []Endpoint
}

// Summary reports what a run did, for the alert channel.
type Summary struct {
	SourceID  string
	Changed   int
	Unchanged int
	Failed    int
	Changes   []works.Change
}

// Runner executes jobs.
type Runner struct {
	Fetcher *Fetcher
	Archive Archive
	Store   Store
	Log     *slog.Logger
	Now     func() time.Time
}

func (r *Runner) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now().UTC()
}

func (r *Runner) log() *slog.Logger {
	if r.Log != nil {
		return r.Log
	}
	return slog.Default()
}

// Run snapshots every endpoint of a job. One failing endpoint does not stop the
// others; the errors are joined and returned.
func (r *Runner) Run(ctx context.Context, job Job) (Summary, error) {
	summary := Summary{SourceID: job.SourceID}
	var errs []error

	for _, ep := range job.Endpoints {
		changes, unchanged, err := r.runEndpoint(ctx, job, ep)
		switch {
		case err != nil:
			summary.Failed++
			errs = append(errs, fmt.Errorf("%s/%s: %w", job.SourceID, ep.Name, err))
		case unchanged:
			summary.Unchanged++
		default:
			summary.Changed++
			summary.Changes = append(summary.Changes, changes...)
		}
	}
	return summary, errors.Join(errs...)
}

// fetchWithFallback tries the endpoint's URL, then its alternates if the
// primary is simply not there. Any other failure is returned as is: a 500 or a
// timeout says nothing about which URL is correct.
func (r *Runner) fetchWithFallback(ctx context.Context, ep Endpoint) (Result, string, error) {
	res, err := r.Fetcher.Fetch(ctx, ep.URL)
	if err == nil {
		return res, ep.URL, nil
	}

	var httpErr *HTTPError
	if !asHTTPError(err, &httpErr) || httpErr.StatusCode != http.StatusNotFound {
		return Result{}, ep.URL, err
	}

	for _, alt := range ep.Alternate {
		altRes, altErr := r.Fetcher.Fetch(ctx, alt)
		if altErr == nil {
			r.log().Info("primary URL is absent; using an alternate",
				"endpoint", ep.Name, "primary", ep.URL, "used", alt)
			return altRes, alt, nil
		}
	}
	return Result{}, ep.URL, err
}

func (r *Runner) runEndpoint(ctx context.Context, job Job, ep Endpoint) (changes []works.Change, unchanged bool, err error) {
	requestedAt := r.now()
	res, usedURL, fetchErr := r.fetchWithFallback(ctx, ep)

	if fetchErr != nil {
		record := FetchRecord{
			SourceID:    job.SourceID,
			Endpoint:    ep.Name,
			RequestedAt: requestedAt,
			Error:       fetchErr.Error(),
		}
		var httpErr *HTTPError
		if asHTTPError(fetchErr, &httpErr) {
			record.StatusCode = httpErr.StatusCode
		}
		if recErr := r.Store.RecordFetch(ctx, record); recErr != nil {
			return nil, false, errors.Join(fetchErr, recErr)
		}
		return nil, false, fetchErr
	}

	record := FetchRecord{
		SourceID:    job.SourceID,
		Endpoint:    ep.Name,
		RequestedAt: requestedAt,
		StatusCode:  res.StatusCode,
		Bytes:       int64(len(res.Body)),
		SHA256:      res.SHA256,
		DurationMS:  int(res.Duration.Milliseconds()),
	}

	lastSHA, err := r.Store.LastDocumentSHA(ctx, job.SourceID, ep.Name)
	if err != nil {
		return nil, false, fmt.Errorf("read last document: %w", err)
	}
	if lastSHA != "" && lastSHA == res.SHA256 {
		record.Unchanged = true
		if err := r.Store.RecordFetch(ctx, record); err != nil {
			return nil, false, fmt.Errorf("record fetch: %w", err)
		}
		r.log().Debug("unchanged", "source", job.SourceID, "endpoint", ep.Name)
		return nil, true, nil
	}

	// New bytes: archive first, then parse. The archive is the evidentiary
	// record, so nothing is parsed that was not stored.
	key := archive.KeyFor(job.SourceID, res.RetrievedAt, res.SHA256, ep.Ext)
	exists, err := r.Archive.Exists(ctx, key)
	if err != nil {
		return nil, false, fmt.Errorf("archive lookup: %w", err)
	}
	if !exists {
		contentType := res.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		if err := r.Archive.Put(ctx, key, res.Body, contentType); err != nil {
			return nil, false, fmt.Errorf("archive write: %w", err)
		}
	}

	docID, err := r.Store.SaveRawDocument(ctx, RawDocument{
		SourceID:    job.SourceID,
		Endpoint:    ep.Name,
		URL:         usedURL,
		ArchiveKey:  key,
		SHA256:      res.SHA256,
		ContentType: res.ContentType,
		Bytes:       int64(len(res.Body)),
		RetrievedAt: res.RetrievedAt,
		CapturedBy:  "ingester:" + job.SourceID,
	})
	if err != nil {
		return nil, false, fmt.Errorf("save raw document: %w", err)
	}
	record.DocumentID = docID
	if err := r.Store.RecordFetch(ctx, record); err != nil {
		return nil, false, fmt.Errorf("record fetch: %w", err)
	}

	parsed, err := ep.Parse(res.Body, job.Blocklist)
	if err != nil {
		return nil, false, fmt.Errorf("parse: %w", err)
	}

	previous, err := r.Store.LoadWorks(ctx, job.SourceID, ep.Name)
	if err != nil {
		return nil, false, fmt.Errorf("load previous snapshot: %w", err)
	}

	changes = works.Diff(previous, parsed)
	if err := r.Store.ApplySnapshot(ctx, SnapshotWrite{
		SourceID:   job.SourceID,
		Endpoint:   ep.Name,
		DocumentID: docID,
		ObservedAt: res.RetrievedAt,
		Records:    parsed,
		Changes:    changes,
	}); err != nil {
		return nil, false, fmt.Errorf("apply snapshot: %w", err)
	}
	if err := r.Store.MarkDocumentParsed(ctx, docID); err != nil {
		return nil, false, fmt.Errorf("mark document parsed: %w", err)
	}

	r.log().Info("snapshot",
		"source", job.SourceID, "endpoint", ep.Name,
		"records", len(parsed), "changes", len(changes))
	return changes, false, nil
}
