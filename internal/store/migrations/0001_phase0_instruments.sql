-- Phase 0 instruments.
-- See docs/03-architecture/02-data-model.md §9A and
--     docs/05-delivery/07-phase-0-instruments.md
--
-- Forward-only. PostGIS is expected but not required for these tables; the
-- geometry column is added in a later migration once the extension is enabled.

CREATE TABLE IF NOT EXISTS sources (
  id                text PRIMARY KEY,
  name              text NOT NULL,
  url               text,
  category          text NOT NULL,
  acquisition       text NOT NULL,
  cadence           text,
  licence           text,
  terms_reviewed_on date,
  terms_reviewed_by text,
  robots_ok         boolean,
  status            text NOT NULL DEFAULT 'planned',
  blocklist         text[] NOT NULL DEFAULT '{}',
  last_success_at   timestamptz,
  last_error        text,
  updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS raw_documents (
  id            uuid PRIMARY KEY,
  source_id     text NOT NULL REFERENCES sources(id),
  url           text,
  archive_key   text NOT NULL,
  sha256        bytea NOT NULL,
  content_type  text,
  bytes         bigint,
  retrieved_at  timestamptz NOT NULL,
  parsed_at     timestamptz,
  parse_status  text,
  captured_by   text NOT NULL,
  CONSTRAINT raw_documents_sha256_key UNIQUE (sha256)
);
CREATE INDEX IF NOT EXISTS raw_documents_source_retrieved_idx
  ON raw_documents (source_id, retrieved_at DESC);

-- One row per fetch attempt, whatever the outcome.
CREATE TABLE IF NOT EXISTS fetch_log (
  id              uuid PRIMARY KEY,
  source_id       text NOT NULL REFERENCES sources(id),
  endpoint        text NOT NULL,
  requested_at    timestamptz NOT NULL,
  status_code     int,
  bytes           bigint,
  sha256          bytea,
  raw_document_id uuid REFERENCES raw_documents(id),
  unchanged       boolean NOT NULL DEFAULT false,
  duration_ms     int,
  error           text
);
CREATE INDEX IF NOT EXISTS fetch_log_source_endpoint_idx
  ON fetch_log (source_id, endpoint, requested_at DESC);

-- Latest known state of each record in a published works dataset, after the
-- blocklist strip.
CREATE TABLE IF NOT EXISTS works (
  id              uuid PRIMARY KEY,
  source_id       text NOT NULL REFERENCES sources(id),
  endpoint        text NOT NULL,
  natural_key     text NOT NULL,
  current         jsonb NOT NULL,
  first_seen_at   timestamptz NOT NULL,
  last_seen_at    timestamptz NOT NULL,
  vanished_at     timestamptz,
  raw_document_id uuid NOT NULL REFERENCES raw_documents(id),
  CONSTRAINT works_natural_key_unique UNIQUE (source_id, endpoint, natural_key)
);
CREATE INDEX IF NOT EXISTS works_last_seen_idx ON works (source_id, last_seen_at DESC);

-- Every observed change, with both values and both source documents.
CREATE TABLE IF NOT EXISTS work_changes (
  id                   uuid PRIMARY KEY,
  work_id              uuid NOT NULL REFERENCES works(id) ON DELETE CASCADE,
  kind                 text NOT NULL CHECK (kind IN ('added','changed','vanished','reappeared')),
  field                text,
  old_value            jsonb,
  new_value            jsonb,
  prev_raw_document_id uuid REFERENCES raw_documents(id),
  raw_document_id      uuid NOT NULL REFERENCES raw_documents(id),
  observed_at          timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS work_changes_observed_idx ON work_changes (observed_at DESC);
CREATE INDEX IF NOT EXISTS work_changes_work_idx ON work_changes (work_id, observed_at DESC);

-- Government resolutions seen by the watcher. A classification is a lead for a
-- human to read, never a legal_constants value.
CREATE TABLE IF NOT EXISTS gr_items (
  sanketank       text PRIMARY KEY,
  issued_on       date,
  title           text,
  raw_document_id uuid REFERENCES raw_documents(id),
  classification  jsonb,
  prompt_version  text,
  model           text,
  reviewed_by     text,
  reviewed_at     timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS gr_items_issued_idx ON gr_items (issued_on DESC);

CREATE TABLE IF NOT EXISTS court_items (
  id              uuid PRIMARY KEY,
  source_id       text NOT NULL REFERENCES sources(id),
  cnr             text,
  case_number     text,
  decided_on      date,
  matched_on      text NOT NULL,
  raw_document_id uuid REFERENCES raw_documents(id),
  created_at      timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT court_items_source_cnr_unique UNIQUE (source_id, cnr)
);

-- Legal constants. Nothing is hard-coded in Go; every value cites its source.
CREATE TABLE IF NOT EXISTS legal_constants (
  id             uuid PRIMARY KEY,
  key            text NOT NULL,
  value          jsonb NOT NULL,
  citation       text NOT NULL,
  citation_url   text,
  verification   text NOT NULL CHECK (verification IN ('verified','secondary','unverified')),
  effective_from date NOT NULL,
  effective_to   date,
  created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS legal_constants_key_idx ON legal_constants (key, effective_from DESC);
