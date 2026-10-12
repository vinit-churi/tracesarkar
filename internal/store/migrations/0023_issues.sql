-- 0023: issues, the thing reports are sightings of.
--
-- A report is one person's photograph at one moment. An issue is the problem
-- itself, which several people may photograph and which persists until it is
-- fixed. Keeping them apart is what lets a second sighting corroborate the
-- first instead of inflating a count, and it is why the glossary refuses to
-- let the two words be used interchangeably.
CREATE TABLE IF NOT EXISTS issues (
  id             uuid PRIMARY KEY,
  authority      text,
  ward           text,
  category       text NOT NULL,
  subcategory    text,
  -- Where the problem is, as the mean of the sightings that formed it. A
  -- second report moves it slightly; that is the point, it is an estimate
  -- that improves.
  centroid       geography(Point, 4326) NOT NULL,
  status         text NOT NULL DEFAULT 'open',
  -- NotifiedAt is when the authority was told, and it is the only thing that
  -- starts a clock. Null until somebody lodges a complaint: photographing a
  -- problem here tells the corporation nothing (D085).
  notified_at    timestamptz,
  reference      text,
  first_seen_at  timestamptz NOT NULL,
  last_seen_at   timestamptz NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT issues_status_known CHECK (status IN
    ('open', 'claimed_resolved', 'citizen_confirmed', 'reopened'))
);

CREATE INDEX IF NOT EXISTS issues_centroid_idx ON issues USING GIST (centroid);
CREATE INDEX IF NOT EXISTS issues_open_idx ON issues (category, status);

-- Which reports are sightings of which issue. A report belongs to at most one
-- issue; an issue has at least one report.
CREATE TABLE IF NOT EXISTS issue_reports (
  issue_id    uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  report_id   uuid NOT NULL UNIQUE,
  -- How far this sighting was from the issue when it joined, and why. A merge
  -- is an assertion that two photographs are of the same thing in the world,
  -- so it is recorded in a form somebody can audit rather than inferred later.
  distance_m  double precision,
  basis       text,
  joined_at   timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (issue_id, report_id)
);

-- The history, append-only and hash-chained. Each row carries the hash of the
-- row before it, so an edit, a deletion or a reordering anywhere in an issue's
-- history is detectable from its end — including by us, which is the point.
CREATE TABLE IF NOT EXISTS issue_events (
  id         uuid PRIMARY KEY,
  issue_id   uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  seq        integer NOT NULL,
  kind       text NOT NULL,
  at         timestamptz NOT NULL,
  detail     text,
  actor      text,
  prev_hash  text,
  hash       text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (issue_id, seq)
);

CREATE INDEX IF NOT EXISTS issue_events_issue_idx ON issue_events (issue_id, seq);

-- Append-only in fact, not only in intention. A history that can be edited in
-- place is not evidence, and a rule that lives in the application is a rule
-- that is one forgotten code path from being broken.
CREATE OR REPLACE FUNCTION issue_events_are_append_only() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'issue_events is append-only: an issue''s history is evidence and is never edited or deleted';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS issue_events_no_update ON issue_events;
CREATE TRIGGER issue_events_no_update
  BEFORE UPDATE OR DELETE ON issue_events
  FOR EACH ROW EXECUTE FUNCTION issue_events_are_append_only();
