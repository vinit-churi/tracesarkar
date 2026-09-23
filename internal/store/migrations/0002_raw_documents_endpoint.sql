-- The snapshotter tracks the last body per endpoint, not just per source, so a
-- source with several endpoints does not compare one against another.

ALTER TABLE raw_documents ADD COLUMN IF NOT EXISTS endpoint text;
CREATE INDEX IF NOT EXISTS raw_documents_source_endpoint_idx
  ON raw_documents (source_id, endpoint, retrieved_at DESC);
