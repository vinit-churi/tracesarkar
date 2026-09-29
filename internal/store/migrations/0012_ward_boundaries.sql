-- 0012: ward boundaries, from BMC's own geo/getwardlayer.
--
-- PostGIS is authoritative for containment (CLAUDE.md). This is the table that
-- decides which authority a report belongs to, so it carries its vintage: the
-- 2025 delimitation moved roughly a quarter of BMC's ward lines, which makes
-- "which boundary set was this decided against" a question with legal weight
-- rather than a curiosity.

CREATE TABLE IF NOT EXISTS ward_boundaries (
  id           uuid PRIMARY KEY,
  -- Canonical form, per the glossary: R/C, not RC. BMC's own API returns both
  -- spellings from different endpoints.
  ward         text NOT NULL,
  authority    text NOT NULL DEFAULT 'BMC',
  geom         geography(MultiPolygon, 4326) NOT NULL,

  -- Provenance, because a routing decision has to be explicable later.
  source_id    text,
  retrieved_at timestamptz,
  archive_key  text,
  vintage      text,

  loaded_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT ward_boundaries_unique UNIQUE (authority, ward)
);

CREATE INDEX IF NOT EXISTS ward_boundaries_geom_idx ON ward_boundaries USING GIST (geom);
