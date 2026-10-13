package publish

import "time"

// Capture is what the platform holds about one photograph, internally.
type Capture struct {
	ReportID      string
	Lat, Lon      float64
	Category      string
	Subcategory   string
	MediaKey      string
	PeoplePresent bool
}

// Redaction is the record that faces and plates were obscured, and that the
// result was written somewhere of its own.
//
// It exists as a record rather than a flag because hard rule 5 is about what
// was *done*, not about what was believed: a boolean saying "safe" can be set
// by anything, while a record naming the tool, the time and the output can be
// checked against the file it claims to have produced.
type Redaction struct {
	RanAt     time.Time
	Tool      string
	OutputKey string
	Regions   int
}

// complete reports whether a redaction actually happened. A record missing any
// of its parts describes an intention, not an act.
func (r *Redaction) complete() bool {
	return r != nil && !r.RanAt.IsZero() && r.Tool != "" && r.OutputKey != ""
}

// Derivative is what a public surface may show.
type Derivative struct {
	ReportID    string
	Category    string
	Subcategory string

	// The position, snapped to a cell. Never the capture's own (hard rule 4).
	Lat, Lon   float64
	CellMetres float64

	// ImageKey is empty unless a redacted image exists. Withheld then says why.
	ImageKey string
	Withheld string
}

// Derive builds what may be published from what is held.
//
// The image is the point of this function. Hard rule 5 says redaction happens
// before the public derivative is written, and the way to honour a rule like
// that is to make it structural rather than conditional: there is no path here
// that emits an image without a completed redaction behind it. A detector that
// is missing, broken or not yet built therefore produces a derivative with no
// picture — an absence, which is visible and safe — rather than an unredacted
// photograph, which is neither.
//
// The classifier's view that nobody is in the frame is deliberately not
// consulted. It is a judgement that can be wrong, and the rule does not depend
// on it being right.
func Derive(c Capture, r *Redaction) Derivative {
	at := Coarsen(c.Lat, c.Lon)

	d := Derivative{
		ReportID:    c.ReportID,
		Category:    c.Category,
		Subcategory: c.Subcategory,
		Lat:         at.Lat,
		Lon:         at.Lon,
		CellMetres:  at.CellMetres,
	}

	if !r.complete() {
		d.Withheld = "The photograph is not shown: it has not been through " +
			"redaction, and no image is published before faces and number " +
			"plates have been obscured."
		return d
	}
	d.ImageKey = r.OutputKey
	return d
}
