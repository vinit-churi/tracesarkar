-- 0010: carry the contractor and its provenance onto the projected segment.
--
-- The attribution names a company. Hard rule 2 forbids displaying a fact about
-- a named party without a source reference and a retrieval time, so the
-- projection carries them alongside the name rather than leaving the UI to
-- find them later. A segment whose contractor is unsourced renders no
-- contractor at all.

ALTER TABLE road_segments ADD COLUMN IF NOT EXISTS contractor_name text;
ALTER TABLE road_segments ADD COLUMN IF NOT EXISTS source_id       text;
ALTER TABLE road_segments ADD COLUMN IF NOT EXISTS retrieved_at    timestamptz;
ALTER TABLE road_segments ADD COLUMN IF NOT EXISTS archive_key     text;

CREATE INDEX IF NOT EXISTS road_segments_contractor_idx
    ON road_segments (contractor_name) WHERE contractor_name IS NOT NULL;
