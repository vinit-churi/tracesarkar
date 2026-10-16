package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/ingest"
)

// browserUA is not decoration. BMC's SAP portal answers a request without a
// browser User-Agent with HTTP 200 and the body "Could not open iView", which
// reads as a live-but-empty page rather than a refusal (D148). Every fetch of
// a portal document has to carry one.
const browserUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// runArchiveDocs stores documents that were read by hand.
//
// Most of what this platform cites is fetched by a collector on a schedule.
// The Section 4 handbooks, citizen charters and departmental circulars behind
// the routing tables are not: they were found by reading, one at a time, and
// until they are in the bucket the rows that cite them point at a live URL on
// a portal that rewrites its own paths. The archive, not the parsed row, is
// the evidentiary record, so this closes the gap in one command and prints
// the key each migration should cite.
func runArchiveDocs(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("archive-docs", flag.ExitOnError)
	sourceID := fs.String("source", "", "source register id, e.g. bmc_portal_rti_manuals")
	manifest := fs.String("manifest", "", "file of URLs, one per line; # comments allowed")
	dryRun := fs.Bool("dry-run", false, "fetch and hash, but write nothing")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*sourceID) == "" {
		return fmt.Errorf("archive-docs: -source is required, and must exist in the source register")
	}

	urls := fs.Args()
	if *manifest != "" {
		body, err := os.ReadFile(*manifest)
		if err != nil {
			return fmt.Errorf("read manifest: %w", err)
		}
		urls = append(urls, parseManifest(string(body))...)
	}
	if len(urls) == 0 {
		return fmt.Errorf("archive-docs: nothing to archive")
	}

	cfg, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	blobs, err := archive.New(archive.Options{
		Endpoint: cfg.R2.Endpoint, Bucket: cfg.R2.Bucket, Region: cfg.R2.Region,
		AccessKey: cfg.R2.AccessKey, Secret: cfg.R2.Secret,
	})
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 3 * time.Minute}
	var stored, already, failed int
	for _, u := range urls {
		key, existed, err := archiveOne(ctx, client, blobs, db, *sourceID, u, *dryRun)
		switch {
		case err != nil:
			failed++
			fmt.Printf("FAILED\t%s\t%v\n", u, err)
		case existed:
			already++
			fmt.Printf("have\t%s\t%s\n", key, u)
		default:
			stored++
			fmt.Printf("stored\t%s\t%s\n", key, u)
		}
	}
	fmt.Printf("\n%d stored, %d already archived, %d failed\n", stored, already, failed)
	if failed > 0 {
		return fmt.Errorf("archive-docs: %d document(s) could not be archived", failed)
	}
	return nil
}

// archiveOne fetches, hashes, stores and records a single document. It returns
// the key and whether those exact bytes were already in the archive.
//
// Addressing by content hash is what makes a re-run safe: the same document
// fetched twice lands on the same key and the second write is skipped, while a
// document the portal has edited gets a new key and the old evidence survives.
func archiveOne(ctx context.Context, client *http.Client, blobs *archive.Client,
	db archiveStore, sourceID, rawURL string, dryRun bool) (string, bool, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", false, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("User-Agent", browserUA)
	res, err := client.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("fetch: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("fetch returned %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 64<<20))
	if err != nil {
		return "", false, fmt.Errorf("read: %w", err)
	}
	if len(body) == 0 {
		return "", false, fmt.Errorf("fetch returned an empty body")
	}
	// The portal's refusal is a 200. Catching it here keeps an error page from
	// being archived as though it were the handbook.
	if isPortalRefusal(body) {
		return "", false, fmt.Errorf("portal returned its iView error page, not the document")
	}

	sum := sha256.Sum256(body)
	digest := hex.EncodeToString(sum[:])
	contentType := res.Header.Get("Content-Type")
	key := archive.KeyFor(sourceID, time.Now(), digest, extensionFor(rawURL, contentType))

	if existing, found, err := db.RawDocumentKeyByHash(ctx, digest); err != nil {
		return "", false, err
	} else if found {
		return existing, true, nil
	}
	if dryRun {
		return key, false, nil
	}

	if err := blobs.Put(ctx, key, body, strings.TrimSpace(contentType)); err != nil {
		return "", false, err
	}
	if _, err := db.SaveRawDocument(ctx, ingest.RawDocument{
		SourceID:    sourceID,
		URL:         rawURL,
		ArchiveKey:  key,
		SHA256:      digest,
		ContentType: contentType,
		Bytes:       int64(len(body)),
		RetrievedAt: time.Now().UTC(),
		CapturedBy:  "ingest archive-docs",
	}); err != nil {
		return "", false, err
	}
	return key, false, nil
}

// archiveStore is the slice of the store this command needs, so the command
// can be read without the whole database surface in view.
type archiveStore interface {
	RawDocumentKeyByHash(ctx context.Context, sha256hex string) (string, bool, error)
	SaveRawDocument(ctx context.Context, doc ingest.RawDocument) (string, error)
}

// parseManifest reads a hand-edited list of URLs.
func parseManifest(body string) []string {
	var out []string
	seen := map[string]bool{}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if seen[line] {
			continue
		}
		seen[line] = true
		out = append(out, line)
	}
	return out
}

// extensionFor picks the extension the archive key ends in.
//
// The URL wins over the Content-Type because BMC's portal serves several of
// these as application/octet-stream, and a bucket full of `.bin` is a bucket
// nobody opens.
func extensionFor(rawURL, contentType string) string {
	if u, err := url.Parse(rawURL); err == nil {
		if ext := strings.ToLower(path.Ext(u.Path)); ext != "" && len(ext) <= 6 {
			return ext
		}
	}
	if mt, _, err := mime.ParseMediaType(contentType); err == nil {
		switch mt {
		case "application/pdf":
			return ".pdf"
		case "text/html":
			return ".html"
		case "application/json", "application/geo+json":
			return ".json"
		case "text/csv":
			return ".csv"
		}
	}
	return ".bin"
}

// isPortalRefusal reports whether a 200 is really BMC's iView error page.
func isPortalRefusal(body []byte) bool {
	if len(body) > 1<<20 {
		return false
	}
	s := string(body)
	return strings.Contains(s, "Could not open iView") ||
		strings.Contains(s, "The iView has timed out")
}
