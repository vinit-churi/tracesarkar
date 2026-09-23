// Package archive stores retrieved artefacts in S3-compatible object storage.
//
// The archive, not the parsed row, is the evidentiary record. Objects are
// addressed by the SHA-256 of their bytes, so a re-fetch of unchanged content
// lands on the same key and nothing is rewritten.
package archive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Options configures a Client.
type Options struct {
	Endpoint  string // https://<account>.r2.cloudflarestorage.com
	Bucket    string
	Region    string // "auto" for R2
	AccessKey string
	Secret    string
	HTTP      *http.Client
}

// Client talks to one bucket.
type Client struct {
	endpoint string
	bucket   string
	region   string
	creds    credentials
	http     *http.Client
}

// New validates options and returns a Client.
func New(opts Options) (*Client, error) {
	if strings.TrimSpace(opts.Endpoint) == "" {
		return nil, errors.New("archive: endpoint is required")
	}
	if strings.TrimSpace(opts.Bucket) == "" {
		return nil, errors.New("archive: bucket is required")
	}
	region := opts.Region
	if region == "" {
		region = "auto"
	}
	httpClient := opts.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 2 * time.Minute}
	}
	return &Client{
		endpoint: strings.TrimSuffix(opts.Endpoint, "/"),
		bucket:   opts.Bucket,
		region:   region,
		creds:    credentials{accessKey: opts.AccessKey, secret: opts.Secret},
		http:     httpClient,
	}, nil
}

// KeyFor returns the object key for an artefact: one directory per source and
// day, named by content hash.
func KeyFor(sourceID string, when time.Time, sha256hex, ext string) string {
	if ext == "" {
		ext = ".bin"
	}
	when = when.UTC()
	return fmt.Sprintf("archive/%s/%04d/%02d/%02d/%s%s",
		sourceID, when.Year(), int(when.Month()), when.Day(), sha256hex, ext)
}

// Sum returns the SHA-256 of b as hex, which is also its archive address.
func Sum(b []byte) string { return sha256Hex(b) }

// Put writes an object. Writing the same key twice with the same bytes is a
// no-op as far as content is concerned, since the key is the content hash.
func (c *Client) Put(ctx context.Context, key string, body []byte, contentType string) error {
	req, err := c.newRequest(ctx, http.MethodPut, key, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("archive: put %s: %w", key, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("archive: put %s: %w", key, statusError(resp))
	}
	return nil
}

// Exists reports whether a key is already stored.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	req, err := c.newRequest(ctx, http.MethodHead, key, nil)
	if err != nil {
		return false, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("archive: head %s: %w", key, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return false, nil
	case resp.StatusCode/100 == 2:
		return true, nil
	default:
		return false, fmt.Errorf("archive: head %s: %w", key, statusError(resp))
	}
}

// Get reads an object back.
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	req, err := c.newRequest(ctx, http.MethodGet, key, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("archive: get %s: %w", key, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("archive: get %s: %w", key, statusError(resp))
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("archive: read %s: %w", key, err)
	}
	return b, nil
}

func (c *Client) newRequest(ctx context.Context, method, key string, body []byte) (*http.Request, error) {
	url := c.endpoint + "/" + c.bucket + "/" + strings.TrimPrefix(key, "/")
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("archive: build request: %w", err)
	}
	if body != nil {
		req.ContentLength = int64(len(body))
	}
	sign(req, c.creds, c.region, "s3", sha256Hex(body), time.Now().UTC())
	return req, nil
}

// statusError summarises a failed response without echoing any credentials.
func statusError(resp *http.Response) error {
	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	text := strings.TrimSpace(string(snippet))
	if text == "" {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return fmt.Errorf("status %d: %s", resp.StatusCode, text)
}
