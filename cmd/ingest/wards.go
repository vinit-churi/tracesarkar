package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/jurisdiction"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

const wardLayerURL = "https://roads.mcgm.gov.in:3000/api/geo/getwardlayer"

// runWards fetches BMC's ward boundary layer, archives it, and loads it.
//
// The archived document is the evidentiary record, not the parsed polygon: a
// routing decision may have to be explained years later, and "this is the
// boundary set BMC published on that date" is the only defensible answer.
func runWards(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("wards", flag.ExitOnError)
	url := fs.String("url", wardLayerURL, "ward layer endpoint")
	vintage := fs.String("vintage", "", "which delimitation this is, e.g. 2025")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, db, closeDB, err := load(ctx)
	if err != nil {
		return err
	}
	defer closeDB()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *url, nil)
	if err != nil {
		return fmt.Errorf("ward layer request: %w", err)
	}
	client := &http.Client{Timeout: 90 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		// BMC answers Indian addresses only (ADR 0015). From anywhere else
		// this is a connection reset, not a 403.
		return fmt.Errorf("fetch ward layer (reachable only from an Indian address): %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("ward layer returned %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 32<<20))
	if err != nil {
		return fmt.Errorf("read ward layer: %w", err)
	}

	boundaries, err := jurisdiction.ParseWardLayer(body)
	if err != nil {
		return err
	}

	// Archive before storing the projection.
	sum := sha256.Sum256(body)
	digest := hex.EncodeToString(sum[:])
	key := fmt.Sprintf("archive/bmc_roads_api/wardlayer/%s/%s.json",
		time.Now().UTC().Format("2006-01-02"), digest[:16])

	blobs, err := archive.New(archive.Options{
		Endpoint: cfg.R2.Endpoint, Bucket: cfg.R2.Bucket, Region: cfg.R2.Region,
		AccessKey: cfg.R2.AccessKey, Secret: cfg.R2.Secret,
	})
	if err != nil {
		return err
	}
	if err := blobs.Put(ctx, key, body, "application/geo+json"); err != nil {
		return fmt.Errorf("archive ward layer: %w", err)
	}

	retrieved := time.Now().UTC()
	for _, b := range boundaries {
		if err := db.SaveWardBoundary(ctx, store.NewWardBoundary{
			Ward:        b.Ward,
			Authority:   "BMC",
			GeoJSON:     b.GeoJSON,
			SourceID:    "bmc_roads_api",
			RetrievedAt: &retrieved,
			ArchiveKey:  key,
			Vintage:     *vintage,
		}); err != nil {
			return err
		}
	}

	total, err := db.CountWardBoundaries(ctx)
	if err != nil {
		return err
	}
	slog.Info("ward boundaries loaded", "wards", len(boundaries),
		"bytes", len(body), "archive_key", key)
	fmt.Fprintf(os.Stdout, "\n%d ward boundaries stored (%d in the table)\n",
		len(boundaries), total)
	return nil
}
