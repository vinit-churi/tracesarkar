-- The GR watcher records which watched terms a resolution's text mentions.
-- A keyword hit is a lead for a human to read, never a legal constant, so it is
-- kept separate from the classification column.

ALTER TABLE gr_items ADD COLUMN IF NOT EXISTS keywords text[] NOT NULL DEFAULT '{}';
ALTER TABLE gr_items ADD COLUMN IF NOT EXISTS first_seen_at timestamptz NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS gr_items_keywords_idx ON gr_items USING gin (keywords);
