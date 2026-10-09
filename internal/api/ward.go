package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Wards is the published-facts half of the platform: what each office is
// responsible for and what obliges it. Separate from Labels because this is
// the only thing served without a sign-in.
type WardProfiles interface {
	WardProfileFor(ctx context.Context, authority, ward string) (store.WardProfile, error)
}

// publicDesk is deliberately narrow. Everything on it is a published fact
// about an office — a designation, an address the body prints in its own
// handbook, a deadline with its citation. Nothing here came from a citizen,
// which is what makes it safe without a sign-in.
type publicDesk struct {
	Category         string `json:"category"`
	Department       string `json:"department"`
	Officer          string `json:"officer,omitempty"`
	Office           string `json:"office,omitempty"`
	Phone            string `json:"phone,omitempty"`
	Email            string `json:"email,omitempty"`
	Hours            string `json:"hours,omitempty"`
	VisitingHours    string `json:"visiting_hours,omitempty"`
	EscalatesTo      string `json:"escalates_to,omitempty"`
	DeadlineHours    int    `json:"deadline_hours,omitempty"`
	DeadlineCitation string `json:"deadline_citation,omitempty"`
	Source           string `json:"source"`
	RetrievedAt      string `json:"retrieved_at"`
}

// handleWardProfile serves what is known about one ward, to anybody.
func (s *Server) handleWardProfile(w http.ResponseWriter, r *http.Request) {
	if s.wards == nil {
		writeError(w, http.StatusNotImplemented, "ward profiles are not configured on this server")
		return
	}

	authority := strings.ToUpper(strings.TrimSpace(r.PathValue("authority")))
	// Ward codes carry a slash — R/C — so the path segment arrives escaped and
	// is read back rather than split on.
	ward := strings.ToUpper(strings.TrimSpace(r.PathValue("ward")))
	if authority == "" || ward == "" {
		writeError(w, http.StatusBadRequest, "name an authority and a ward")
		return
	}

	profile, err := s.wards.WardProfileFor(r.Context(), authority, ward)
	if err != nil {
		s.log.Error("could not read ward profile", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read that ward")
		return
	}

	desks := make([]publicDesk, 0, len(profile.Desks))
	for _, d := range profile.Desks {
		desks = append(desks, publicDesk{
			Category: d.Category, Department: d.Department,
			Officer: d.Officer, Office: d.Office,
			Phone: d.OfficePhone, Email: d.OfficeEmail,
			Hours: d.OfficeHours, VisitingHours: d.VisitingHours,
			EscalatesTo:      d.EscalatesTo,
			DeadlineHours:    d.DeadlineHours,
			DeadlineCitation: d.DeadlineCitation,
			Source:           d.SourceRef,
			RetrievedAt:      d.RetrievedAt.UTC().Format("2 January 2006"),
		})
	}

	// Cacheable, because none of it is personal and none of it changes often.
	// A public endpoint is also a load surface.
	w.Header().Set("Cache-Control", "public, max-age=900")
	writeJSON(w, http.StatusOK, map[string]any{
		"authority": profile.Authority,
		"ward":      profile.Ward,
		"desks":     desks,
		"coverage": map[string]any{
			"segments": profile.Segments,
			"km":       profile.Metres / 1000,
		},
		// Stated with the data, not left to the page, so any consumer of this
		// endpoint carries the limit along with the facts.
		"limits": []string{
			"Departments and deadlines are recorded for this ward only.",
			"Contract geometry covers part of the road network, not all of it. " +
				"Where there is none, this platform cannot tell a road the " +
				"corporation has not published from a road it does not own — " +
				"a stretch may belong to MMRDA, MSRDC, PWD or NHAI instead.",
			"Nothing here is filed with anyone. Every complaint or application " +
				"the platform produces is a draft a person reviews and sends.",
		},
	})
}
