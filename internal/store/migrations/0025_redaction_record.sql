-- 0025: what redaction actually did to a photograph.
--
-- Hard rule 5 says no raw face or number plate reaches a public derivative.
-- The way to keep a rule like that is to record what was done rather than what
-- was believed: a boolean saying "safe" can be set by anything, while a record
-- naming the tool, the time and the output object can be checked against the
-- file it claims to have produced, and can be re-run when the detector
-- improves (ADR 0008).
--
-- The verdict is stored beside the record because "ran and published" and "ran
-- and held" are both complete outcomes, and only the first of them may put an
-- image on a public page. A row with no verdict has not been looked at.

ALTER TABLE report_media
  ADD COLUMN IF NOT EXISTS redaction_ran_at     timestamptz,
  ADD COLUMN IF NOT EXISTS redaction_tool       text,
  ADD COLUMN IF NOT EXISTS redaction_output_key text,
  ADD COLUMN IF NOT EXISTS redaction_regions    integer,
  ADD COLUMN IF NOT EXISTS redaction_verdict    text,
  ADD COLUMN IF NOT EXISTS redaction_reason     text,
  ADD COLUMN IF NOT EXISTS redaction_attempts   integer NOT NULL DEFAULT 0;

-- A published verdict without an output object is the one combination that
-- would let an unredacted image through, so the database refuses it. The
-- rendering path refuses it too; this is the second lock, on the side that
-- survives a code change.
ALTER TABLE report_media
  DROP CONSTRAINT IF EXISTS report_media_published_needs_an_output;
ALTER TABLE report_media
  ADD CONSTRAINT report_media_published_needs_an_output CHECK (
    redaction_verdict IS DISTINCT FROM 'published'
    OR (redaction_output_key IS NOT NULL
        AND redaction_tool IS NOT NULL
        AND redaction_ran_at IS NOT NULL)
  );

ALTER TABLE report_media
  DROP CONSTRAINT IF EXISTS report_media_redaction_verdict_known;
ALTER TABLE report_media
  ADD CONSTRAINT report_media_redaction_verdict_known CHECK (
    redaction_verdict IS NULL OR redaction_verdict IN ('published', 'held')
  );

-- The sweep asks for photographs nothing has looked at yet.
CREATE INDEX IF NOT EXISTS report_media_unredacted_idx
    ON report_media (created_at)
    WHERE redaction_verdict IS NULL;
