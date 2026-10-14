package api

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	"net/http"
	"strconv"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/evidence"
	"github.com/vinit-churi/tracesarkar/internal/publish"
	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Evidence compiles one issue into the record behind the pack.
type Evidence interface {
	EvidenceFor(ctx context.Context, issueID, accountID string) (evidence.Record, []store.EvidencePhoto, bool, error)
}

// handleEvidencePack serves the compiled record as a PDF.
//
// This is the one share surface that is not public, and deliberately so: the
// pack carries the exact position and the complaint reference, because it
// exists to be attached to a complaint and a complaint needs both. Everything
// in it already belongs to the person asking for it.
//
// Scoped by sighting rather than by ownership — an issue belongs to nobody,
// and the people who photographed it are the ones who can act on it — and a
// stranger is told the issue does not exist rather than that they may not have
// it, because which issues exist is itself information.
func (s *Server) handleEvidencePack(w http.ResponseWriter, r *http.Request) {
	if s.evidence == nil {
		writeError(w, http.StatusNotImplemented, "evidence packs are not configured on this server")
		return
	}

	issueID := r.PathValue("id")
	record, photos, ok, err := s.evidence.EvidenceFor(r.Context(), issueID, s.reporter(r))
	if err != nil {
		s.log.Error("could not compile an evidence pack", "issue_id", issueID, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not compile that record")
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "no such issue")
		return
	}

	body, err := evidence.Build(record, s.packPhotos(r, photos))
	if err != nil {
		s.log.Error("could not build an evidence pack", "issue_id", issueID, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not compile that record")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	// A name somebody can find again in a downloads folder, and one that says
	// what the file is rather than naming a random id first.
	w.Header().Set("Content-Disposition",
		`attachment; filename="tracesarkar-record-`+evidence.Stamp(record)+`.pdf"`)
	// Private: it carries the exact position and the reference.
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// packPhotos fetches the redacted copies the pack will carry.
//
// A photograph that will not load or will not decode costs the picture and not
// the pack: the compiled record is worth having without it, and failing the
// whole document because one object is unreadable is the wrong trade for
// somebody standing in a ward office.
func (s *Server) packPhotos(r *http.Request, refs []store.EvidencePhoto) []evidence.Photo {
	if s.blobs == nil {
		return nil
	}
	var out []evidence.Photo
	for _, ref := range refs {
		if ref.Key == "" || !ref.Published {
			continue
		}
		raw, err := s.blobs.Get(r.Context(), ref.Key)
		if err != nil {
			s.log.Error("could not read a redacted copy for a pack", "error", err.Error())
			continue
		}
		// Sized for the page before it is embedded, so the pack stays small
		// enough to email from a ward office.
		sized, err := publish.ForWeb(raw)
		if err != nil {
			s.log.Error("could not size a photograph for a pack", "error", err.Error())
			continue
		}
		cfg, _, err := image.DecodeConfig(bytes.NewReader(sized))
		if err != nil {
			s.log.Error("could not read a photograph for a pack", "error", err.Error())
			continue
		}
		p := evidence.Photo{
			JPEG: sized, Width: cfg.Width, Height: cfg.Height,
			AccuracyM: ref.AccuracyM,
			Redaction: redactionWords(ref),
		}
		if ref.TakenAt != nil {
			if t, err := time.Parse(time.RFC3339, *ref.TakenAt); err == nil {
				p.TakenAt = t
			}
		}
		out = append(out, p)
	}
	return out
}

func itoa(n int) string { return strconv.Itoa(n) }

func redactionWords(ref store.EvidencePhoto) string {
	if ref.Tool == "" {
		return ""
	}
	switch ref.Regions {
	case 0:
		return ref.Tool + ", which found nothing to obscure"
	case 1:
		return ref.Tool + ", 1 region obscured"
	}
	return ref.Tool + ", " + itoa(ref.Regions) + " regions obscured"
}
