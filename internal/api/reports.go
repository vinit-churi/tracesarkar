package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// captureMeta is the JSON part of a capture, as specified in
// docs/03-architecture/03-api-design.md.
type captureMeta struct {
	Location struct {
		Lat        float64  `json:"lat"`
		Lon        float64  `json:"lon"`
		AccuracyM  float64  `json:"accuracy_m"`
		HeadingDeg *float64 `json:"heading_deg"`
	} `json:"location"`
	CapturedAt      string `json:"captured_at"`
	Description     string `json:"description"`
	DescriptionLang string `json:"description_lang"`
	DeviceID        string `json:"device_id"`

	// Label is sent by the field kit, which captures evaluation data rather
	// than citizen reports. Ordinary captures omit it.
	Label *struct {
		FrameType       string   `json:"frame_type"`
		Label           string   `json:"label"`
		Conditions      []string `json:"conditions"`
		WardGroundTruth string   `json:"ward_ground_truth"`
		Notes           string   `json:"notes"`
	} `json:"label"`
}

// handlePostReport stores a capture and returns 202. The report is durable
// before this returns; classification, jurisdiction and attribution happen
// afterwards, because a capture must never be lost waiting for them.
func (s *Server) handlePostReport(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.maxSize)

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if ok := asMaxBytes(err, &tooLarge); ok {
			writeError(w, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("upload exceeds %d bytes", s.maxSize))
			return
		}
		writeError(w, http.StatusBadRequest, "could not read the upload: "+err.Error())
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	var meta captureMeta
	rawMeta := r.FormValue("meta")
	if rawMeta == "" {
		writeError(w, http.StatusBadRequest, "meta is required")
		return
	}
	if err := json.Unmarshal([]byte(rawMeta), &meta); err != nil {
		writeError(w, http.StatusBadRequest, "meta is not valid JSON: "+err.Error())
		return
	}

	if meta.Location.Lat == 0 && meta.Location.Lon == 0 {
		writeError(w, http.StatusBadRequest,
			"location is required: a capture without coordinates cannot be routed")
		return
	}
	capturedAt, err := time.Parse(time.RFC3339, meta.CapturedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest,
			"captured_at must be an RFC 3339 timestamp: "+err.Error())
		return
	}

	files := r.MultipartForm.File["media"]
	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "at least one photograph is required")
		return
	}
	roles := r.MultipartForm.Value["role"]

	reportID, created, err := s.reports.SaveReport(r.Context(), NewReport{
		AccountID:   s.reporter(r),
		Lat:         meta.Location.Lat,
		Lon:         meta.Location.Lon,
		AccuracyM:   meta.Location.AccuracyM,
		HeadingDeg:  meta.Location.HeadingDeg,
		CapturedAt:  capturedAt,
		Description: meta.Description,
		Lang:        meta.DescriptionLang,
		Idempotency: r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		s.log.Error("could not save report", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not store the report")
		return
	}

	// The photographs are stored before replying. A 202 promises the capture is
	// safe; it would be a lie if the image were still only in memory.
	for i, fh := range files {
		role := "close"
		if i < len(roles) && roles[i] != "" {
			role = roles[i]
		}

		file, err := fh.Open()
		if err != nil {
			s.log.Error("could not read upload", "error", err.Error())
			writeError(w, http.StatusInternalServerError, "could not read the photograph")
			return
		}
		body, err := io.ReadAll(file)
		_ = file.Close()
		if err != nil {
			s.log.Error("could not read upload", "error", err.Error())
			writeError(w, http.StatusInternalServerError, "could not read the photograph")
			return
		}

		sum := sha256.Sum256(body)
		digest := hex.EncodeToString(sum[:])
		contentType := fh.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}
		key := mediaKey(reportID, digest, fh.Filename)

		if err := s.media.Put(r.Context(), key, body, contentType); err != nil {
			s.log.Error("could not store media", "error", err.Error(), "report_id", reportID)
			writeError(w, http.StatusInternalServerError,
				"could not store the photograph; please retry")
			return
		}
		if err := s.reports.AddReportMedia(r.Context(), NewMedia{
			ReportID:    reportID,
			Role:        role,
			ArchiveKey:  key,
			ContentType: contentType,
			Bytes:       int64(len(body)),
			SHA256:      digest,
		}); err != nil {
			s.log.Error("could not record media", "error", err.Error(), "report_id", reportID)
			writeError(w, http.StatusInternalServerError, "could not record the photograph")
			return
		}
	}

	if meta.Label != nil {
		if err := s.reports.SaveReportLabel(r.Context(), ReportLabel{
			ReportID:        reportID,
			FrameType:       meta.Label.FrameType,
			Label:           meta.Label.Label,
			Conditions:      meta.Label.Conditions,
			WardGroundTruth: meta.Label.WardGroundTruth,
			Notes:           meta.Label.Notes,
			LabelledBy:      s.reporter(r),
		}); err != nil {
			// The capture is already safe; a label that failed to save is worth
			// logging, not worth rejecting the report over.
			s.log.Warn("could not save label", "error", err.Error(), "report_id", reportID)
		}
	}

	s.log.Info("report stored",
		"report_id", reportID, "created", created, "media", len(files),
		"accuracy_m", meta.Location.AccuracyM)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"report_id":          reportID,
		"status":             "pending",
		"created":            created,
		"media":              len(files),
		"estimated_ready_ms": 4000,
		"poll":               "/v1/reports/" + reportID,
	})
}

// reporter is the account a capture belongs to. A signed-in person is
// attributed to themselves; the field kit presents the server's own token,
// which carries no identity, so the server's account is the only honest
// answer. Getting this wrong would file every citizen's report under one
// account, and no report could be traced back to who took the photograph.
func (s *Server) reporter(r *http.Request) string {
	if claims, ok := claimsFrom(r.Context()); ok && claims.AccountID != "" {
		return claims.AccountID
	}
	return s.account
}

// mediaKey addresses a photograph by report and content hash, so the same image
// uploaded twice lands in the same place.
func mediaKey(reportID, digest, filename string) string {
	ext := ".jpg"
	if i := strings.LastIndex(filename, "."); i >= 0 && len(filename)-i <= 5 {
		ext = strings.ToLower(filename[i:])
	}
	return fmt.Sprintf("media/%s/%s%s", reportID, digest, ext)
}

// asMaxBytes reports whether err is the body-size error, without depending on
// the error's text.
func asMaxBytes(err error, target **http.MaxBytesError) bool {
	for err != nil {
		if e, ok := err.(*http.MaxBytesError); ok {
			*target = e
			return true
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}
