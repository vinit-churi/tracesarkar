-- Reports: the atomic unit of input, and the first thing the platform must
-- never lose (hard rule 7). A report is persisted before any enrichment is
-- attempted, so everything downstream can be retried.
--
-- See docs/03-architecture/02-data-model.md §5. This migration creates the
-- subset v0.1 needs; later milestones extend it rather than rename it.

CREATE EXTENSION IF NOT EXISTS postgis;

DO $$ BEGIN
  CREATE TYPE report_status AS ENUM ('pending', 'enriched', 'merged', 'rejected', 'withdrawn');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE issue_category AS ENUM (
    'road_defect', 'waste', 'water_drainage', 'street_furniture', 'structural',
    'environment', 'transit', 'illegal_construction', 'other');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE severity AS ENUM ('low', 'medium', 'high', 'critical');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- Phone-only identity (ADR 0011). The number itself is never stored, only a
-- peppered hash, so a breach does not hand over a list of reporters.
CREATE TABLE IF NOT EXISTS accounts (
  id             uuid PRIMARY KEY,
  phone_hash     bytea UNIQUE,
  display_handle text,
  tier           text NOT NULL DEFAULT 'personal',
  created_at     timestamptz NOT NULL DEFAULT now(),
  last_seen_at   timestamptz
);

CREATE TABLE IF NOT EXISTS reports (
  id                  uuid PRIMARY KEY,
  account_id          uuid REFERENCES accounts(id),
  issue_id            uuid,                       -- set on merge, in a later milestone
  status              report_status NOT NULL DEFAULT 'pending',

  location            geography(Point, 4326) NOT NULL,
  location_accuracy_m numeric(7,1),
  heading_deg         numeric(5,2),
  captured_at         timestamptz NOT NULL,       -- the device's clock
  received_at         timestamptz NOT NULL DEFAULT now(),
  clock_skew_s        int GENERATED ALWAYS AS
                        (EXTRACT(EPOCH FROM (received_at - captured_at))::int) STORED,

  description         text,
  description_lang    text,

  classification      jsonb,                      -- the structured model output, verbatim
  category            issue_category,
  subcategory         text,
  severity            severity,
  hazard_to_life      boolean NOT NULL DEFAULT false,

  image_integrity     jsonb NOT NULL DEFAULT '{}',
  device_hash         bytea,
  ip_prefix_hash      bytea,
  idempotency_key     text,

  created_at          timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS reports_location_idx ON reports USING gist (location);
CREATE INDEX IF NOT EXISTS reports_account_idx ON reports (account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS reports_pending_idx ON reports (status) WHERE status = 'pending';
-- A retried upload must not create a second report.
CREATE UNIQUE INDEX IF NOT EXISTS reports_idempotency_idx
  ON reports (account_id, idempotency_key) WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS report_media (
  id             uuid PRIMARY KEY,
  report_id      uuid NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
  role           text NOT NULL,          -- close | wide | noticeboard
  archive_key    text NOT NULL,          -- the original, private
  derivative_key text,                   -- redacted, for any public surface
  content_type   text,
  bytes          bigint,
  sha256         bytea NOT NULL,
  width          int,
  height         int,
  captured_at    timestamptz,
  created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS report_media_report_idx ON report_media (report_id);

-- Human labels on field-kit reports: the classification evaluation set and the
-- jurisdiction golden set. Without these, accuracy cannot be claimed.
CREATE TABLE IF NOT EXISTS report_labels (
  report_id         uuid PRIMARY KEY REFERENCES reports(id) ON DELETE CASCADE,
  frame_type        text NOT NULL,
  label             text NOT NULL,
  conditions        text[] NOT NULL DEFAULT '{}',
  ward_ground_truth text,
  notes             text,
  labelled_by       uuid REFERENCES accounts(id),
  labelled_at       timestamptz NOT NULL DEFAULT now()
);
