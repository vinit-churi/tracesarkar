package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/archive"
	"github.com/vinit-churi/tracesarkar/internal/redact"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// derivativeQuality is the JPEG quality of the public derivative. High enough
// that a pothole's edge is still legible, low enough to meet the 150 KB budget
// a public page has on a 2G connection.
const derivativeQuality = 82

// runRedact obscures faces and number plates in captured photographs, and
// records what it managed to do.
//
// It writes the redacted derivative to the private archive rather than to a
// public bucket. The public surface serves it through the API, so the gate that
// refuses an image without a completed redaction record sits in front of every
// read rather than beside one.
func runRedact(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("redact", flag.ExitOnError)
	limit := fs.Int("limit", 50, "how many photographs to scan")
	out := fs.String("out", "", "also write each scanned photograph to this directory, for inspection")
	dry := fs.Bool("dry", false, "scan and report without writing anything")
	if err := fs.Parse(args); err != nil {
		return err
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

	faces, err := redact.NewFaces()
	if err != nil {
		return err
	}
	scanner := redact.NewScanner(faces, redact.NewPlates())

	targets, err := db.MediaToRedact(ctx, *limit)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		slog.Info("nothing to redact")
		return nil
	}

	if *out != "" {
		if err := os.MkdirAll(*out, 0o755); err != nil {
			return fmt.Errorf("make output directory: %w", err)
		}
	}

	var published, held, failed int
	for _, t := range targets {
		verdict, regions, err := redactOne(ctx, db, blobs, scanner, t, *out, *dry)
		switch {
		case err != nil:
			failed++
			slog.Error("redaction failed",
				"report_id", t.ReportID, "attempts", t.Attempts+1, "error", err)
			if !*dry {
				if err := db.BumpRedactionAttempt(ctx, t.MediaID); err != nil {
					slog.Error("could not count the attempt", "report_id", t.ReportID, "error", err)
				}
			}
		case verdict.Publish:
			published++
			slog.Info("redacted", "report_id", t.ReportID, "regions", regions)
		default:
			held++
			slog.Info("held", "report_id", t.ReportID, "reason", verdict.Reason)
		}
	}

	slog.Info("redaction pass complete",
		"scanned", len(targets), "published", published, "held", held, "failed", failed)

	// A pass where nothing at all succeeded is a failure of the pass, not a
	// report of one. The classifier taught this: counting attempts and exiting
	// zero made a broken detector look like an empty queue.
	if published == 0 && held == 0 && len(targets) > 1 {
		return fmt.Errorf("scanned %d photographs and reached no verdict on any of them", len(targets))
	}
	return nil
}

func redactOne(
	ctx context.Context,
	db *store.DB,
	blobs *archive.Client,
	scanner *redact.Scanner,
	t store.RedactionTarget,
	outDir string,
	dry bool,
) (redact.Verdict, int, error) {
	raw, err := blobs.Get(ctx, t.ArchiveKey)
	if err != nil {
		return redact.Verdict{}, 0, fmt.Errorf("fetch photograph: %w", err)
	}

	img, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return redact.Verdict{}, 0, fmt.Errorf("decode %s: %w", t.ContentType, err)
	}

	started := time.Now()
	scan := scanner.Scan(ctx, img)
	verdict := redact.Decide(scan, t.PeoplePresent)

	slog.Info("scanned",
		"report_id", t.ReportID, "format", format,
		"faces", countKind(scan, redact.Face), "plates", countKind(scan, redact.Plate),
		"took_ms", time.Since(started).Milliseconds())

	// The mosaic is applied whatever the verdict. A held photograph is held
	// because something could not be accounted for, not because the regions
	// that were found are safe to leave alone, and the inspection copy has to
	// show what the detectors actually did.
	redacted := redact.Apply(img, scan.Regions)

	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, redacted, &jpeg.Options{Quality: derivativeQuality}); err != nil {
		return redact.Verdict{}, 0, fmt.Errorf("encode derivative: %w", err)
	}

	if outDir != "" {
		name := filepath.Join(outDir, t.ReportID+"-"+t.MediaID[:8]+".jpg")
		if err := os.WriteFile(name, encoded.Bytes(), 0o644); err != nil {
			return redact.Verdict{}, 0, fmt.Errorf("write inspection copy: %w", err)
		}
	}

	if dry {
		return verdict, len(scan.Regions), nil
	}

	result := store.RedactionResult{
		RanAt:     time.Now().UTC(),
		Tool:      scan.Tools(),
		Regions:   len(scan.Regions),
		Published: verdict.Publish,
		Reason:    verdict.Reason,
	}

	// Stored whatever the verdict, because this copy is what the classifier
	// reads: ADR 0008 says an un-redacted photograph of an identifiable person
	// must not go to a third party, and classification is a third party. A held
	// verdict means the copy may not be shown to the public, not that it may
	// not be used.
	key := derivativeKey(t)
	if err := blobs.Put(ctx, key, encoded.Bytes(), "image/jpeg"); err != nil {
		return redact.Verdict{}, 0, fmt.Errorf("store redacted copy: %w", err)
	}
	// Recorded only after the object exists. The other order would leave a
	// record claiming an output object that was never written, which is exactly
	// the claim the record exists to make checkable.
	result.OutputKey = key

	if err := db.RecordRedaction(ctx, t.MediaID, result); err != nil {
		if errors.Is(err, store.ErrMediaGone) {
			// The reporter deleted it while we were scanning. Not a failure.
			slog.Info("photograph was deleted during the scan", "report_id", t.ReportID)
			return verdict, len(scan.Regions), nil
		}
		return redact.Verdict{}, 0, err
	}
	return verdict, len(scan.Regions), nil
}

func derivativeKey(t store.RedactionTarget) string {
	return "derivative/" + t.ReportID + "/" + t.MediaID + ".jpg"
}

func countKind(s redact.Scan, k redact.Kind) int {
	n := 0
	for _, r := range s.Regions {
		if r.Kind == k {
			n++
		}
	}
	return n
}
