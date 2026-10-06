package exif_test

import (
	"os"
	"testing"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/exif"
)

// Synthetic bytes prove the parser's logic; a real photograph proves the
// assumption behind it — that the phones this platform is used from write a
// timestamp at all, and that it agrees with what the client reported.
//
// Run with a path to a capture:
//
//	TRACESARKAR_EXIF_SAMPLE=/tmp/real.jpg go test ./internal/exif/
func TestRealCaptureCarriesItsOwnTimestamp(t *testing.T) {
	path := os.Getenv("TRACESARKAR_EXIF_SAMPLE")
	if path == "" {
		t.Skip("set TRACESARKAR_EXIF_SAMPLE to a stored capture to run this")
	}
	image, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ist := time.FixedZone("IST", 5*3600+1800)
	taken, ok := exif.TakenAt(image, ist)
	if !ok {
		t.Fatal("a real capture carries no timestamp; the gap check cannot work")
	}
	t.Logf("taken at %s", taken.Format(time.RFC3339))

	if taken.Year() < 2020 || taken.After(time.Now().Add(24*time.Hour)) {
		t.Errorf("implausible timestamp: %s", taken)
	}
}
