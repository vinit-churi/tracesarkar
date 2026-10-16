-- 0046: the question the platform asks and then forgets it asked.
--
-- `internal/classify` has three paths to `needs_confirmation`, and each one
-- computes the vocabulary to offer: the model named a subcategory the
-- platform cannot route, or a category with no department behind it, or it
-- scored below the confidence floor. In every case `Decision.Alternatives`
-- holds the list a citizen should be shown.
--
-- Nothing persisted it. `SaveClassification` wrote nineteen columns and not
-- that one, so the question was computed, discarded, and never asked. The
-- capture then sits for ever: the sweep's issues stage attaches a capture
-- only when its latest classification was `accepted` or `not_yet_covered`,
-- so a `needs_confirmation` capture is skipped on every pass, correctly, and
-- waits for an answer that no surface could collect. Three real captures
-- were in that state when the operator console first counted them
-- (D178) — the oldest from 5 October.
--
-- Hard rule 7 is "never lose a captured report". A capture nobody can finish
-- is lost in the way that matters: the photograph is safe and the report is
-- never made.
--
-- Two columns.
--
-- `alternatives` is what was offered, stored rather than recomputed, because
-- the taxonomy changes. A citizen answering a question asked in September
-- must be answering from the list they were shown, and a list regenerated
-- today could contain options that did not exist then — or be missing the one
-- they would have picked.
--
-- `source` is who answered. A reporter-confirmed row is the right answer by
-- definition, not a prediction, and anything reading this table has to be
-- able to tell the difference.
--
-- Stated precisely, because it is easy to overstate: **the evaluation set is
-- not affected today**. `ingest eval` grades the classifier by re-running it
-- live against the archived image and comparing to a label a person recorded
-- at `/label/`; it never reads this table, so a reporter row could not have
-- inflated a score. What it would have corrupted is every *other* reader —
-- `LatestClassification` and the sweep's own queries return the newest row
-- regardless of who wrote it, so without this column a citizen's answer is
-- indistinguishable from a model's output to any future caller, including a
-- later eval that decides to score stored rows instead of re-running.
-- Default 'model', so every existing row is correctly attributed.

ALTER TABLE classifications ADD COLUMN IF NOT EXISTS alternatives jsonb;
ALTER TABLE classifications ADD COLUMN IF NOT EXISTS source text NOT NULL DEFAULT 'model';

ALTER TABLE classifications DROP CONSTRAINT IF EXISTS classifications_source_known;
ALTER TABLE classifications ADD CONSTRAINT classifications_source_known
  CHECK (source IN ('model', 'reporter'));

COMMENT ON COLUMN classifications.alternatives IS
  'The options offered to the reporter when the outcome was '
  'needs_confirmation. Stored rather than recomputed: the taxonomy changes, '
  'and an answer must come from the list that was actually shown.';
COMMENT ON COLUMN classifications.source IS
  'model or reporter. A reporter''s answer is the right answer by '
  'definition and must never be scored as a prediction the classifier made.';

-- The lookup is "my captures still waiting on me", so it is by account and by
-- recency. The account lives on reports, so this index serves the join's
-- inner half: the latest classification per report.
CREATE INDEX IF NOT EXISTS classifications_latest_idx
  ON classifications (report_id, created_at DESC) WHERE error IS NULL;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM classifications WHERE source NOT IN ('model','reporter')) THEN
    RAISE EXCEPTION 'a classification has an unknown source';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
     WHERE table_name = 'classifications' AND column_name = 'alternatives') THEN
    RAISE EXCEPTION 'classifications.alternatives was not created';
  END IF;
END $$;
