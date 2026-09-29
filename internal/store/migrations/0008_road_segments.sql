-- 0008: road geometry, projected out of the collected works records.
--
-- The works table keeps BMC's JSON exactly as it was served, because the
-- archive is the evidentiary record. This table is the queryable projection of
-- the one field the attribution join needs: the shape of the road a contract
-- covers.
--
-- It is derived, not authoritative. It can be dropped and rebuilt from works at
-- any time, and a reload must never be able to produce duplicates, so the
-- identity is the work it came from.

CREATE TABLE IF NOT EXISTS road_segments (
  id            uuid PRIMARY KEY,
  work_id       uuid NOT NULL REFERENCES works(id) ON DELETE CASCADE,
  work_code     text NOT NULL,
  ward          text NOT NULL,
  location_name text,

  -- geography, not geometry: the join asks "within N metres of this point",
  -- and geography answers in metres without picking a projection. PostGIS is
  -- authoritative for containment and distance (CLAUDE.md); H3 is an index.
  geom          geography(MultiLineString, 4326) NOT NULL,

  -- The dates that bound a defect liability period. dlpPeriod is null
  -- throughout BMC's API, so the completion date is what we actually have.
  start_date    timestamptz,
  end_date      timestamptz,

  loaded_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT road_segments_work_unique UNIQUE (work_id)
);

-- The index the join lives on.
CREATE INDEX IF NOT EXISTS road_segments_geom_idx ON road_segments USING GIST (geom);
CREATE INDEX IF NOT EXISTS road_segments_ward_idx ON road_segments (ward);
CREATE INDEX IF NOT EXISTS road_segments_work_code_idx ON road_segments (work_code);
