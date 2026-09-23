// Package ingest fetches source data politely, archives what is new, and
// records every attempt.
package ingest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// DefaultUserAgent identifies the crawler honestly, as the politeness policy in
// docs/03-architecture/07-ingestion-and-scrapers.md requires.
const DefaultUserAgent = "TraceSarkar/0.1 (+https://github.com/vinit-churi/tracesarkar; civic accountability research)"

// HTTPError is a non-2xx response.
type HTTPError struct {
	StatusCode int
	URL        string
	Snippet    string
}

func (e *HTTPError) Error() string {
	if e.Snippet != "" {
		return fmt.Sprintf("http %d from %s: %s", e.StatusCode, e.URL, e.Snippet)
	}
	return fmt.Sprintf("http %d from %s", e.StatusCode, e.URL)
}

// Result is one successful fetch.
type Result struct {
	Body        []byte
	SHA256      string
	StatusCode  int
	ContentType string
	Duration    time.Duration
	Attempts    int
	RetrievedAt time.Time
}

// Fetcher performs polite, retrying HTTP GETs.
type Fetcher struct {
	HTTP      *http.Client
	UserAgent string
	Retries   int
	Backoff   time.Duration
}

// NewFetcher returns a Fetcher with the politeness defaults.
func NewFetcher(client *http.Client, userAgent string) *Fetcher {
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second}
	}
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}
	return &Fetcher{HTTP: client, UserAgent: userAgent, Retries: 3, Backoff: 2 * time.Second}
}

// Fetch retrieves a URL. Server errors are retried with jittered backoff;
// client errors are not, because they will not fix themselves.
func (f *Fetcher) Fetch(ctx context.Context, url string) (Result, error) {
	retries := f.Retries
	if retries < 1 {
		retries = 1
	}

	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		res, err := f.once(ctx, url)
		if err == nil {
			res.Attempts = attempt
			return res, nil
		}
		lastErr = err

		var httpErr *HTTPError
		if ok := asHTTPError(err, &httpErr); ok && httpErr.StatusCode < 500 && httpErr.StatusCode != http.StatusTooManyRequests {
			return Result{}, err
		}
		if attempt == retries {
			break
		}
		wait := time.Duration(attempt) * f.Backoff
		wait += time.Duration(rand.Int63n(int64(wait/2 + 1)))
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		case <-time.After(wait):
		}
	}
	return Result{}, fmt.Errorf("after %d attempts: %w", retries, lastErr)
}

func (f *Fetcher) once(ctx context.Context, url string) (Result, error) {
	started := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", f.UserAgent)
	req.Header.Set("Accept", "application/json, application/geo+json, */*")

	resp, err := f.HTTP.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("get %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("read %s: %w", url, err)
	}
	if resp.StatusCode/100 != 2 {
		return Result{}, &HTTPError{
			StatusCode: resp.StatusCode,
			URL:        url,
			Snippet:    strings.TrimSpace(string(body[:min(len(body), 200)])),
		}
	}
	if len(body) == 0 {
		return Result{}, fmt.Errorf("get %s: empty response body", url)
	}

	sum := sha256.Sum256(body)
	return Result{
		Body:        body,
		SHA256:      hex.EncodeToString(sum[:]),
		StatusCode:  resp.StatusCode,
		ContentType: strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0]),
		Duration:    time.Since(started),
		RetrievedAt: started.UTC(),
	}, nil
}

func asHTTPError(err error, target **HTTPError) bool {
	for err != nil {
		if e, ok := err.(*HTTPError); ok {
			*target = e
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
