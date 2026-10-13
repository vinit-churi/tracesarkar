package publish

import (
	"strings"
	"testing"
	"time"
)

func capture() Capture {
	return Capture{
		ReportID: "r1", Lat: lat, Lon: lon,
		Category: "road_defect", Subcategory: "pothole",
		MediaKey: "media/r1/a.jpg",
	}
}

// Hard rule 5, made structural. The rule is not "redact well", which depends
// on a detector nobody can prove correct; it is that no public derivative
// exists until redaction has run. A missing redactor then produces an absence,
// not a leak.
func TestNoImageIsPublishedWithoutARedaction(t *testing.T) {
	d := Derive(capture(), nil)

	if d.ImageKey != "" {
		t.Errorf("an image was published with no redaction behind it: %q", d.ImageKey)
	}
	if !strings.Contains(strings.ToLower(d.Withheld), "redact") {
		t.Errorf("the absence is not explained: %q", d.Withheld)
	}
	// Everything that is not the photograph still publishes. An issue with no
	// picture is still worth showing.
	if d.Category != "road_defect" || d.Lat == 0 {
		t.Error("the rest of the derivative was withheld along with the image")
	}
}

// The classifier's opinion that nobody is in the frame does not substitute for
// redaction. It is a judgement that can be wrong, and the rule is not
// conditional on it.
func TestAClassifierSayingNoPeopleIsNotARedaction(t *testing.T) {
	c := capture()
	c.PeoplePresent = false

	if d := Derive(c, nil); d.ImageKey != "" {
		t.Error("an image was published because a model said nobody was in it")
	}
}

func TestARedactedImageIsPublished(t *testing.T) {
	r := &Redaction{
		RanAt: time.Now().UTC(), Tool: "test-redactor",
		OutputKey: "public/r1/redacted.jpg", Regions: 2, Published: true,
	}
	d := Derive(capture(), r)

	if d.ImageKey != "public/r1/redacted.jpg" {
		t.Errorf("the redacted image was not published: %q", d.ImageKey)
	}
	if d.Withheld != "" {
		t.Errorf("a published image still reports a withholding: %q", d.Withheld)
	}
	// Never the original. The archive key is the unredacted photograph.
	if strings.Contains(d.ImageKey, "media/") {
		t.Error("the derivative points at the original capture")
	}
}

// A redaction that produced nothing is not a redaction.
func TestAnIncompleteRedactionPublishesNoImage(t *testing.T) {
	for _, r := range []*Redaction{
		{RanAt: time.Now(), Tool: "t", Published: true},                    // no output
		{Tool: "t", OutputKey: "public/r1/x.jpg", Published: true},         // never ran
		{RanAt: time.Now(), OutputKey: "public/r1/x.jpg", Published: true}, // no tool recorded
		{
			// Everything ran, everything is recorded, and the scan itself
			// said no. A redacted copy exists for the platform's own use;
			// that is not the same as permission to show it.
			RanAt: time.Now(), Tool: "t", OutputKey: "public/r1/x.jpg",
		},
	} {
		if d := Derive(capture(), r); d.ImageKey != "" {
			t.Errorf("published on an incomplete redaction: %+v", r)
		}
	}
}

// Hard rule 4 travels with it: what goes out is a cell, never the position
// the complaint is addressed from.
func TestTheDerivativeCarriesOnlyACoarsenedPosition(t *testing.T) {
	c := capture()
	d := Derive(c, nil)

	if d.Lat == c.Lat || d.Lon == c.Lon {
		t.Error("the exact position reached the derivative")
	}
	if d.Lat != Coarsen(c.Lat, c.Lon).Lat {
		t.Error("the position was not snapped to the grid")
	}
	if d.CellMetres < 100 {
		t.Error("the derivative does not say how vague its position is")
	}
}
