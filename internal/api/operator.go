package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/vinit-churi/tracesarkar/internal/store"
)

// Operator is the collection and queue health system 8 needs.
type Operator interface {
	OperatorHealth(ctx context.Context, runLimit int) (store.Health, error)
	OperatorTrend(ctx context.Context, days int) (store.Trend, error)
}

// handleOperatorHealth answers the three questions an operator currently asks
// a log: is collection running, is anything stuck, and what is the ratio.
//
// Authenticated, and not public. Not because the numbers are secret — the
// claimed-versus-confirmed ratio is meant to be published
// ([D007](../../docs/00-overview/05-decision-log.md)) — but because a count of
// what is stuck is an operational signal and the published version of that
// ratio should come from a surface designed for it, with the denominator and
// the caveats beside it, rather than from a dashboard built for one person.
func (s *Server) handleOperatorHealth(w http.ResponseWriter, r *http.Request) {
	if s.operator == nil {
		writeError(w, http.StatusNotImplemented, "operator health is not configured on this server")
		return
	}

	runs := 10
	if v := r.URL.Query().Get("runs"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > 100 {
			writeError(w, http.StatusBadRequest, "runs must be a number between 0 and 100")
			return
		}
		runs = n
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	h, err := s.operator.OperatorHealth(ctx, runs)
	if err != nil {
		// Named rather than swallowed: a dashboard that silently shows zero
		// is worse than one that says it could not count.
		s.log.Error("could not read operator health", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read collection health")
		return
	}

	sources := make([]map[string]any, 0, len(h.Sources))
	for _, src := range h.Sources {
		sources = append(sources, map[string]any{
			"source": src.SourceID, "endpoint": src.Endpoint,
			"last_fetch": src.LastFetchAt, "last_change": src.LastChange,
			"records": src.Records, "attempts": src.Attempts, "failures": src.Failures,
		})
	}

	runRows := make([]map[string]any, 0, len(h.Runs))
	for _, run := range h.Runs {
		row := map[string]any{
			"command": run.Command, "started_at": run.StartedAt,
			"finished_at": run.FinishedAt, "ok": run.OK,
		}
		// A run with no finish time never returned, which is a different
		// failure from one that returned an error and must not read the same.
		if run.FinishedAt == nil {
			row["never_returned"] = true
		}
		if run.Error != "" {
			row["error"] = run.Error
		}
		runRows = append(runRows, row)
	}

	claimed, confirmed := h.ClaimedVersusConfirmed()
	writeJSON(w, http.StatusOK, map[string]any{
		"at":       h.At,
		"captures": h.Captures,
		"queues":   h.Queues,
		"issues":   h.Issues,
		"sources":  sources,
		"runs":     runRows,
		// The two numbers, never a percentage. A percentage of small numbers
		// is a lie with a decimal point in it (D007).
		"claimed_resolved":  claimed,
		"citizen_confirmed": confirmed,
		"note": "Every queue names what clears it. A queue whose answer is a " +
			"person, or nothing, is not a backlog.",
	})
}

// handleOperatorTrend answers the question the point-in-time counts cannot:
// is this getting worse.
//
// A separate call from the health endpoint rather than a field on it. The
// health numbers are a handful of counts and are wanted on every page load;
// this one scans a month of two tables and is wanted when somebody is
// actually asking about a trend. Folding them together would make the common
// request pay for the rare one.
func (s *Server) handleOperatorTrend(w http.ResponseWriter, r *http.Request) {
	if s.operator == nil {
		writeError(w, http.StatusNotImplemented, "operator health is not configured on this server")
		return
	}

	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 180 {
			writeError(w, http.StatusBadRequest, "days must be a number between 1 and 180")
			return
		}
		days = n
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	t, err := s.operator.OperatorTrend(ctx, days)
	if err != nil {
		s.log.Error("could not read the operator trend", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "could not read collection history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"days":     t.Days,
		"fetches":  t.Fetches,
		"runs":     t.Runs,
		"captures": t.Captures,
		// Named rather than left for somebody to discover by looking for a
		// series that is not there. A queue is counted by asking what is
		// unfinished now, nothing has ever recorded that answer, and a
		// series that looked like depth but was really arrivals would be the
		// most misleading thing on the page.
		"no_history_for": "queue depth. A queue is counted by asking the database " +
			"what is unfinished right now, and that answer has never been recorded, " +
			"so \"the backlog is growing\" cannot be answered from this data. " +
			"Captures and issues per day are arrivals, not depth.",
	})
}
