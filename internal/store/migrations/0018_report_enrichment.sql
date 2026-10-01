-- 0018: where a capture is, and which contract covers it.
--
-- Three systems answer three questions about one photograph, and each is
-- recorded separately because each can succeed or fail on its own. A report
-- with a ward but no contract is the common case — about half of Borivali has
-- no contract data — and that is a normal answer, not a partial failure.
--
-- Like classifications, these are append-only. Re-running against corrected
-- boundaries or newly published geometry adds a row; it does not erase what
-- the platform told a citizen last week.

CREATE TABLE IF NOT EXISTS report_jurisdiction (
  id              uuid PRIMARY KEY,
  report_id       uuid NOT NULL REFERENCES reports(id) ON DELETE CASCADE,

  ward            text,
  authority       text,
  confidence      text NOT NULL,
  needs_question  boolean NOT NULL DEFAULT false,
  basis           text NOT NULL,
  alternatives    jsonb,

  -- Which boundary set decided this. The 2025 delimitation moved about a
  -- quarter of BMC's ward lines, so "which boundaries" is a question with
  -- legal weight rather than a curiosity.
  boundary_vintage text,

  error           text,
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS report_jurisdiction_report_idx
    ON report_jurisdiction (report_id, created_at DESC);

CREATE TABLE IF NOT EXISTS report_attribution (
  id              uuid PRIMARY KEY,
  report_id       uuid NOT NULL REFERENCES reports(id) ON DELETE CASCADE,

  confidence      text NOT NULL,
  matched_work_id uuid REFERENCES works(id) ON DELETE SET NULL,
  work_code       text,
  contractor_name text,
  location_name   text,
  distance_m      double precision,
  search_radius_m double precision,
  basis           text NOT NULL,
  would_change    text,
  alternatives    jsonb,

  -- Provenance for the contractor name. Hard rule 2: a fact about a named
  -- party renders only with a source and a retrieval time.
  source_id       text,
  retrieved_at    timestamptz,

  error           text,
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS report_attribution_report_idx
    ON report_attribution (report_id, created_at DESC);
