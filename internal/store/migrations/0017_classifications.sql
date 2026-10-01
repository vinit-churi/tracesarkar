-- 0017: what the classifier concluded about a capture.
--
-- Kept beside the report rather than on it, because a report has exactly one
-- photograph set and may be classified more than once — a prompt revision, a
-- model change, or a citizen correcting it. The history is the point: when an
-- accuracy number moves, the question is always "what changed", and that is
-- unanswerable if each run overwrites the last.
--
-- Every row records which model and which prompt version produced it (ADR
-- 0006). A classification whose provenance is unknown cannot be included in
-- an eval, because there is no telling what it is evidence of.

CREATE TABLE IF NOT EXISTS classifications (
  id             uuid PRIMARY KEY,
  report_id      uuid NOT NULL REFERENCES reports(id) ON DELETE CASCADE,

  -- What the model said.
  category       text,
  subcategory    text,
  severity       text,
  hazard_to_life boolean,
  water_present  boolean,
  people_present boolean,
  image_quality  text,
  is_civic_issue boolean,
  confidence     numeric(4,3),
  rationale      text,

  -- What the platform decided, after its own rules. This is what a citizen
  -- sees, and it is not always what the model said.
  outcome            text NOT NULL,
  overridden         boolean NOT NULL DEFAULT false,
  override_reason    text,
  redaction_required boolean NOT NULL DEFAULT false,
  rejection_reason   text,

  -- Provenance.
  model          text NOT NULL,
  prompt_version text NOT NULL,
  latency_ms     integer,

  -- A failed attempt is a row too. A classification that silently never ran
  -- is indistinguishable from one that ran and found nothing.
  error          text,

  created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS classifications_report_idx
    ON classifications (report_id, created_at DESC);

-- Finding work: captures with no successful classification yet.
CREATE INDEX IF NOT EXISTS classifications_pending_idx
    ON classifications (created_at) WHERE error IS NOT NULL;
