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
