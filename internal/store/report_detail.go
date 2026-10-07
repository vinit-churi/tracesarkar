package store

import (
	"time"

	"github.com/vinit-churi/tracesarkar/internal/action"
	"github.com/vinit-churi/tracesarkar/internal/classify"
)

// ReportDetail is everything the platform knows about one capture.
//
// Parts may be absent. A capture whose enrichment has not run is still a
// capture, and "we are still working on it" is a better answer than an error —
// the photograph is safe either way.
type ReportDetail struct {
	ID        string `json:"id"`
	AccountID string `json:"-"` // never rendered; used to decide who may read this
	Status    string `json:"status"`
	CreatedAt string `json:"created_at,omitempty"`

	Lat       float64 `json:"lat,omitempty"`
	Lon       float64 `json:"lon,omitempty"`
	AccuracyM float64 `json:"accuracy_m,omitempty"`

	Classification *classify.Decision `json:"classification,omitempty"`

	Ward           string `json:"ward,omitempty"`
	Authority      string `json:"authority,omitempty"`
	WardConfidence string `json:"ward_confidence,omitempty"`
	WardBasis      string `json:"ward_basis,omitempty"`
	NeedsQuestion  bool   `json:"needs_question,omitempty"`

	// ExifTakenAt is when the image says it was taken, where it says so.
	ExifTakenAt *time.Time `json:"exif_taken_at,omitempty"`

	// NextStep is what to do about this capture. Absent until the platform
	// knows the category and the ward, because a next step addressed to the
	// wrong desk is worse than none.
	NextStep *action.Next `json:"next_step,omitempty"`

	ContractorName string  `json:"contractor_name,omitempty"`
	WorkCode       string  `json:"work_code,omitempty"`
	RoadName       string  `json:"road_name,omitempty"`
	DistanceM      float64 `json:"distance_m,omitempty"`
	ContractBasis  string  `json:"contract_basis,omitempty"`
	ContractSource string  `json:"contract_source,omitempty"`
	AttrConfidence string  `json:"attribution_confidence,omitempty"`

	// ContractRetrievedAt is when the source naming the contractor was
	// fetched. Hard rule 2: no source and retrieval time, no name.
	ContractRetrievedAt *time.Time `json:"contract_retrieved_at,omitempty"`
}

// staleAfter is how far the photograph's own timestamp may be from the moment
// the report was sent before the position stops meaning "where this is".
//
// Fifteen minutes is a walk, not a journey: the first real captures were three
// seconds apart, and a capture queued offline and sent when signal returns is
// still minutes, not hours.
const staleAfter = 15 * time.Minute

// PositionIsStale reports whether the photograph was taken far enough from the
// moment it was sent that the position attached to it is not where it was
// taken.
//
// The position on a report is the device's position when it was sent. Pick a
// photograph out of a gallery at your desk and the report claims the problem is
// at your desk — and attribution will name whichever contract covers the desk.
// False when the image carries no timestamp: that is not a claim either way.
func (d ReportDetail) PositionIsStale() bool {
	if d.ExifTakenAt == nil || d.CreatedAt == "" {
		return false
	}
	sent, err := time.Parse(time.RFC3339, d.CreatedAt)
	if err != nil {
		return false
	}
	gap := sent.Sub(*d.ExifTakenAt)
	if gap < 0 {
		// A camera clock running ahead is not trustworthy either.
		gap = -gap
	}
	return gap > staleAfter
}

// contractGoverns is the set of categories a road works contract actually
// answers for. A contract covers a stretch of road; it does not make its
// contractor responsible for everything that happens on it.
var contractGoverns = map[string]bool{"road_defect": true}

// ForCitizen returns the detail with anything the platform has no standing to
// claim removed.
//
// Enrichment attributes every capture, because "which contract covers this
// point" is a true and useful question whatever the photograph shows. But the
// answer is only *about* the problem when the contract governs that kind of
// problem. The first real waste capture came back naming a road contractor at
// high confidence, and a company's name beside a garbage report reads as an
// accusation — hard rule 3 forbids the assertion, hard rule 2 governs the
// name.
//
// So the attribution stays in the record and leaves the reply. An unclassified
// capture is treated the same way: until we know what is in the photograph we
// cannot know whose contract is relevant, and naming someone on the strength of
// not knowing is the worst case of all.
func (d ReportDetail) ForCitizen() ReportDetail {
	category := ""
	if d.Classification != nil {
		category = d.Classification.Category
	}
	// A stale position was not where the photograph was taken, so the contract
	// matched against it covers the wrong place. Naming its holder would be
	// worse than naming nobody.
	//
	// And a name is only ever shown with where it came from and when we
	// fetched it (hard rule 2). Missing either, the name does not render.
	sourced := d.ContractSource != "" && d.ContractRetrievedAt != nil
	if contractGoverns[category] && !d.PositionIsStale() && sourced {
		return d
	}

	d.ContractorName = ""
	d.WorkCode = ""
	d.RoadName = ""
	d.DistanceM = 0
	d.ContractBasis = ""
	d.ContractSource = ""
	d.ContractRetrievedAt = nil
	d.AttrConfidence = ""
	return d
}
