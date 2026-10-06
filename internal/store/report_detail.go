package store

import "github.com/vinit-churi/tracesarkar/internal/classify"

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

	ContractorName string  `json:"contractor_name,omitempty"`
	WorkCode       string  `json:"work_code,omitempty"`
	RoadName       string  `json:"road_name,omitempty"`
	DistanceM      float64 `json:"distance_m,omitempty"`
	ContractBasis  string  `json:"contract_basis,omitempty"`
	ContractSource string  `json:"contract_source,omitempty"`
	AttrConfidence string  `json:"attribution_confidence,omitempty"`
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
	if contractGoverns[category] {
		return d
	}

	d.ContractorName = ""
	d.WorkCode = ""
	d.RoadName = ""
	d.DistanceM = 0
	d.ContractBasis = ""
	d.ContractSource = ""
	d.AttrConfidence = ""
	return d
}
