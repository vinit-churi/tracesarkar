# Data model

PostgreSQL 16 + PostGIS. Written as SQL because ambiguity in a schema is expensive later.

**Conventions:** UUIDv7 primary keys · `timestamptz` in UTC · money as `bigint` paise ·
`geography(Point, 4326)` for points, `geography(MultiPolygon, 4326)` for areas · native enums for
status · append-only timelines.

---

## 1. Extensions and enums

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE EXTENSION IF NOT EXISTS h3;          -- optional; see ADR 0004

CREATE TYPE report_status AS ENUM ('pending','classified','merged','rejected');

CREATE TYPE issue_status AS ENUM (
  'unverified','verified','routed','acknowledged','sla_breached',
  'claimed_resolved','citizen_confirmed','reopened','escalated','closed'
);

CREATE TYPE issue_category AS ENUM (
  'road_defect','waste','water_drainage','street_furniture',
  'structural','environment','transit','illegal_construction','other'
);

CREATE TYPE severity AS ENUM ('low','medium','high','critical');

CREATE TYPE authority_kind AS ENUM (
  'municipal_corporation','municipal_council','parastatal','state_department',
  'railway','police','regulator','panchayat'
);

CREATE TYPE contract_stage AS ENUM (
  'tender','awarded','work_order','in_progress','completed','terminated','unknown'
);

CREATE TYPE escalation_kind AS ENUM (
  'rti','rti_first_appeal','rti_second_appeal','rts_appeal',
  'ngt_oa','ngt_compensation','lokayukta','consumer_ejagriti',
  'hc_compensation_claim','contractor_liability_notice'
);

CREATE TYPE escalation_status AS ENUM (
  'draft','ready','filed','acknowledged','replied','appealed','disposed','withdrawn'
);
```

---

## 2. Identity and trust

```sql
CREATE TABLE accounts (
  id                uuid PRIMARY KEY,
  handle            text UNIQUE NOT NULL,          -- pseudonymous, user-chosen
  phone_hash        bytea UNIQUE NOT NULL,         -- HMAC(phone, pepper); plaintext never stored
  phone_verified_at timestamptz,
  display_language  text NOT NULL DEFAULT 'en',    -- en | mr | hi | gu
  trust_score       numeric(4,3) NOT NULL DEFAULT 0.300,
  trust_updated_at  timestamptz NOT NULL DEFAULT now(),
  role              text NOT NULL DEFAULT 'citizen', -- citizen | moderator | admin | api
  status            text NOT NULL DEFAULT 'active',  -- active | limited | suspended
  created_at        timestamptz NOT NULL DEFAULT now(),
  deleted_at        timestamptz,
  CONSTRAINT trust_range CHECK (trust_score BETWEEN 0 AND 1)
);

-- Sybil / astroturf signals. Deliberately separate so it can be purged independently.
CREATE TABLE account_signals (
  account_id     uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  device_hash    bytea,
  ip_prefix_hash bytea,
  first_seen     timestamptz NOT NULL DEFAULT now(),
  last_seen      timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (account_id, device_hash, ip_prefix_hash)
);

CREATE TABLE trust_events (
  id          uuid PRIMARY KEY,
  account_id  uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  kind        text NOT NULL,        -- verified_report | confirmed_fix | rejected_report | ...
  delta       numeric(4,3) NOT NULL,
  ref_type    text, ref_id uuid,
  reason      text,
  created_at  timestamptz NOT NULL DEFAULT now()
);
```

---

## 3. Geography and authorities

```sql
CREATE TABLE authorities (
  id            uuid PRIMARY KEY,
  code          text UNIQUE NOT NULL,           -- BMC, MMRDA, WR, PWD…
  name          text NOT NULL,
  name_mr       text,
  kind          authority_kind NOT NULL,
  parent_id     uuid REFERENCES authorities(id),
  website       text,
  created_at    timestamptz NOT NULL DEFAULT now()
);

-- Time-versioned: CIDCO→NMMC handovers, ward redistricting, etc.
CREATE TABLE authority_boundaries (
  id            uuid PRIMARY KEY,
  authority_id  uuid NOT NULL REFERENCES authorities(id),
  geom          geography(MultiPolygon,4326) NOT NULL,
  source_id     text NOT NULL,                  -- source register key
  source_version text,
  valid_from    date NOT NULL,
  valid_to      date,                           -- NULL = current
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON authority_boundaries USING gist (geom);
CREATE INDEX ON authority_boundaries (authority_id, valid_from, valid_to);

CREATE TABLE wards (
  id            uuid PRIMARY KEY,
  authority_id  uuid NOT NULL REFERENCES authorities(id),
  code          text NOT NULL,                  -- 'R/S', 'P/N'
  name          text NOT NULL,
  name_mr       text,
  geom          geography(MultiPolygon,4326) NOT NULL,
  source_id     text NOT NULL,
  valid_from    date NOT NULL,
  valid_to      date,
  UNIQUE (authority_id, code, valid_from)
);
CREATE INDEX ON wards USING gist (geom);

-- Overrides ward containment. The single most important correctness layer.
CREATE TABLE road_ownership (
  id            uuid PRIMARY KEY,
  authority_id  uuid NOT NULL REFERENCES authorities(id),
  road_name     text,
  geom          geography(MultiLineString,4326) NOT NULL,
  buffer_m      int NOT NULL DEFAULT 15,
  source_id     text NOT NULL,                  -- usually an RTI reply
  confidence    numeric(3,2) NOT NULL DEFAULT 0.90,
  valid_from    date NOT NULL,
  valid_to      date
);
CREATE INDEX ON road_ownership USING gist (geom);

-- Flyovers, viaducts, sea link: disambiguation zones
CREATE TABLE elevated_structures (
  id            uuid PRIMARY KEY,
  authority_id  uuid NOT NULL REFERENCES authorities(id),
  name          text NOT NULL,
  geom          geography(MultiPolygon,4326) NOT NULL,
  ground_authority_id uuid REFERENCES authorities(id)
);
CREATE INDEX ON elevated_structures USING gist (geom);

CREATE TABLE railway_premises (
  id            uuid PRIMARY KEY,
  authority_id  uuid NOT NULL REFERENCES authorities(id),  -- WR / CR
  station_name  text NOT NULL,
  station_code  text,
  geom          geography(MultiPolygon,4326) NOT NULL
);
CREATE INDEX ON railway_premises USING gist (geom);

-- (authority, category) -> department, SLA, escalation ladder, filing channel
CREATE TABLE authority_departments (
  id            uuid PRIMARY KEY,
  authority_id  uuid NOT NULL REFERENCES authorities(id),
  category      issue_category NOT NULL,
  department    text NOT NULL,
  sla_hours     int,
  sla_source    text,          -- 'bombay_hc_2025-10-13' | 'citizen_charter_2024' | ...
  escalation_ladder jsonb NOT NULL DEFAULT '[]',
  filing_channel jsonb NOT NULL DEFAULT '{}',   -- adapter config
  pio           jsonb,                          -- name, designation, address
  first_appellate_authority jsonb,
  UNIQUE (authority_id, category)
);
```

---

## 4. Assets

```sql
CREATE TABLE assets (
  id            uuid PRIMARY KEY,
  kind          text NOT NULL,        -- road_segment | footpath | manhole | drain | streetlight | fob
  name          text,
  authority_id  uuid REFERENCES authorities(id),
  ward_id       uuid REFERENCES wards(id),
  geom          geography(Geometry,4326) NOT NULL,
  attributes    jsonb NOT NULL DEFAULT '{}',
  source_id     text,
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON assets USING gist (geom);
CREATE INDEX ON assets (kind, authority_id);
```

Assets are optional in v0.1 (issues can exist without one) and become the backbone from v0.3, when
repeat-failure detection (`H7`) needs a stable object to attach history to.

---

## 5. Reports and issues

```sql
CREATE TABLE reports (
  id                uuid PRIMARY KEY,
  account_id        uuid REFERENCES accounts(id),
  issue_id          uuid,                       -- FK added after issues; set on merge
  status            report_status NOT NULL DEFAULT 'pending',

  location          geography(Point,4326) NOT NULL,
  location_accuracy_m numeric(6,1),
  heading_deg       numeric(5,2),
  captured_at       timestamptz NOT NULL,       -- device clock
  received_at       timestamptz NOT NULL DEFAULT now(),
  clock_skew_s      int GENERATED ALWAYS AS
                      (EXTRACT(EPOCH FROM (received_at - captured_at))::int) STORED,

  description       text,
  description_lang  text,
  audio_key         text,                       -- S3 key for a voice note

  classification    jsonb,                      -- the structured VLM output
  category          issue_category,
  subcategory       text,
  severity          severity,
  hazard_to_life    boolean NOT NULL DEFAULT false,

  image_integrity   jsonb NOT NULL DEFAULT '{}',
  device_hash       bytea,
  ip_prefix_hash    bytea,

  created_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON reports USING gist (location);
CREATE INDEX ON reports (account_id, created_at DESC);
CREATE INDEX ON reports (status) WHERE status = 'pending';

CREATE TABLE report_media (
  id            uuid PRIMARY KEY,
  report_id     uuid NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
  role          text NOT NULL,          -- wide | close | confirmation
  archive_key   text NOT NULL,          -- private original, EXIF intact
  public_key    text,                   -- redacted derivative
  sha256        bytea NOT NULL,
  width         int, height int, bytes bigint,
  exif          jsonb,                  -- retained on the archival copy only
  phash         bytea,                  -- perceptual hash for reuse detection
  redaction     jsonb NOT NULL DEFAULT '{}',  -- {faces: n, plates: n, model_version: ...}
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON report_media (phash);

CREATE TABLE issues (
  id                uuid PRIMARY KEY,
  public_slug       text UNIQUE NOT NULL,
  status            issue_status NOT NULL DEFAULT 'unverified',
  category          issue_category NOT NULL,
  subcategory       text,
  severity          severity NOT NULL,
  hazard_to_life    boolean NOT NULL DEFAULT false,

  location          geography(Point,4326) NOT NULL,   -- centroid of member reports
  location_public   geography(Point,4326) NOT NULL,   -- coarsened for public display
  h3_r9             text,                             -- bucketing / heatmaps
  asset_id          uuid REFERENCES assets(id),

  authority_id      uuid REFERENCES authorities(id),
  ward_id           uuid REFERENCES wards(id),
  department        text,
  jurisdiction      jsonb NOT NULL DEFAULT '{}',      -- full resolution incl. confidence + basis

  report_count      int NOT NULL DEFAULT 1,
  corroboration_count int NOT NULL DEFAULT 0,

  first_reported_at timestamptz NOT NULL,
  routed_at         timestamptz,
  sla_due_at        timestamptz,
  claimed_resolved_at timestamptz,
  confirmed_at      timestamptz,
  closed_at         timestamptz,

  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON issues USING gist (location);
CREATE INDEX ON issues (ward_id, status, first_reported_at DESC);
CREATE INDEX ON issues (h3_r9);
CREATE INDEX ON issues (status, sla_due_at) WHERE status IN ('routed','acknowledged');

ALTER TABLE reports
  ADD CONSTRAINT reports_issue_fk FOREIGN KEY (issue_id) REFERENCES issues(id);

-- Append-only. Never UPDATE, never DELETE.
CREATE TABLE issue_events (
  id            uuid PRIMARY KEY,
  issue_id      uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  seq           bigint NOT NULL,
  kind          text NOT NULL,        -- reported | classified | located | attributed | filed | ...
  actor_type    text NOT NULL,        -- citizen | platform | authority | moderator
  actor_id      uuid,
  payload       jsonb NOT NULL DEFAULT '{}',
  prev_hash     bytea,
  hash          bytea NOT NULL,       -- sha256(prev_hash || canonical(payload))
  occurred_at   timestamptz NOT NULL DEFAULT now(),
  UNIQUE (issue_id, seq)
);
```

The hash chain on `issue_events` is what lets an exported evidence pack claim tamper-evidence.

---

## 6. Procurement

```sql
CREATE TABLE contractors (
  id                uuid PRIMARY KEY,
  legal_name        text NOT NULL,
  normalised_name   text NOT NULL,           -- for fuzzy matching
  cin               text,                    -- MCA Corporate Identity Number
  gstin             text,
  registered_address text,
  status            text,                    -- active | struck_off | ...
  created_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON contractors USING gin (normalised_name gin_trgm_ops);
CREATE UNIQUE INDEX ON contractors (cin) WHERE cin IS NOT NULL;

CREATE TABLE contractor_aliases (
  id            uuid PRIMARY KEY,
  contractor_id uuid NOT NULL REFERENCES contractors(id) ON DELETE CASCADE,
  alias         text NOT NULL,
  source_id     text NOT NULL,
  UNIQUE (contractor_id, alias)
);

CREATE TABLE contractor_directors (
  id            uuid PRIMARY KEY,
  contractor_id uuid NOT NULL REFERENCES contractors(id) ON DELETE CASCADE,
  din           text,                        -- Director Identification Number
  name          text NOT NULL,
  from_date     date, to_date date,
  source_id     text NOT NULL
);
CREATE INDEX ON contractor_directors (din);

CREATE TABLE blacklistings (
  id            uuid PRIMARY KEY,
  contractor_id uuid NOT NULL REFERENCES contractors(id),
  authority_id  uuid NOT NULL REFERENCES authorities(id),
  reason        text,
  from_date     date NOT NULL,
  to_date       date,
  source_id     text NOT NULL,
  source_doc    text NOT NULL,               -- archive key of the notice
  retrieved_at  timestamptz NOT NULL
);

CREATE TABLE contracts (
  id                uuid PRIMARY KEY,
  external_id       text NOT NULL,           -- verbatim tender/contract ID
  source_id         text NOT NULL,           -- mahatenders | bmc_portal | cppp
  authority_id      uuid REFERENCES authorities(id),
  contractor_id     uuid REFERENCES contractors(id),
  title             text NOT NULL,
  description       text,
  stage             contract_stage NOT NULL DEFAULT 'unknown',
  category          issue_category,          -- inferred work type
  value_paise       bigint,
  awarded_on        date,
  work_order_on     date,
  completed_on      date,
  dlp_months        int,
  dlp_ends_on       date,
  ocds              jsonb NOT NULL DEFAULT '{}',   -- OCDS-shaped record
  extraction        jsonb NOT NULL DEFAULT '{}',   -- provenance of extracted fields
  retrieved_at      timestamptz NOT NULL,
  created_at        timestamptz NOT NULL DEFAULT now(),
  UNIQUE (source_id, external_id)
);
CREATE INDEX ON contracts (contractor_id);
CREATE INDEX ON contracts (dlp_ends_on) WHERE dlp_ends_on IS NOT NULL;

-- Where the work physically is. The join key to issues.
CREATE TABLE contract_sites (
  id            uuid PRIMARY KEY,
  contract_id   uuid NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
  geom          geography(Geometry,4326) NOT NULL,
  derivation    text NOT NULL,     -- gazetteer | road_name_match | explicit_coords | ward_polygon
  confidence    numeric(3,2) NOT NULL,
  basis         jsonb NOT NULL DEFAULT '{}',
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON contract_sites USING gist (geom);

CREATE TABLE issue_contract_matches (
  id              uuid PRIMARY KEY,
  issue_id        uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  contract_id     uuid NOT NULL REFERENCES contracts(id),
  confidence      numeric(3,2) NOT NULL,
  basis           jsonb NOT NULL,          -- every signal and its weight
  matcher_version text NOT NULL,
  in_dlp          boolean,
  published       boolean NOT NULL DEFAULT false,  -- above the publication threshold?
  superseded_by   uuid REFERENCES issue_contract_matches(id),
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON issue_contract_matches (issue_id) WHERE superseded_by IS NULL;
```

---

## 7. Escalations and filings

```sql
CREATE TABLE escalations (
  id              uuid PRIMARY KEY,
  issue_id        uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  account_id      uuid NOT NULL REFERENCES accounts(id),
  kind            escalation_kind NOT NULL,
  status          escalation_status NOT NULL DEFAULT 'draft',
  template_version text NOT NULL,
  generated       jsonb NOT NULL,          -- structured content of the instrument
  rendered_key    text,                    -- S3 key of the PDF/DOCX
  addressee       jsonb NOT NULL,          -- PIO / registry / commission details
  limitation_ends_on date,                 -- NGT 6 months, RTI appeal 30 days, …
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON escalations (limitation_ends_on) WHERE status IN ('draft','ready');

CREATE TABLE filings (
  id              uuid PRIMARY KEY,
  escalation_id   uuid REFERENCES escalations(id),
  issue_id        uuid REFERENCES issues(id),  -- null when the instrument concerns something else
  subject         jsonb,                   -- {kind: work|contract|building|other, ref: …} when issue_id is null
  account_id      uuid NOT NULL REFERENCES accounts(id),
  channel         text NOT NULL,           -- mybmc | aaple_sarkar | swachhata | railmadad | rti_online | ngt | ejagriti
  external_ref    text,                    -- the official ticket / registration number
  filed_at        timestamptz NOT NULL,
  filed_by        text NOT NULL,           -- 'user' | 'adapter'
  response        jsonb NOT NULL DEFAULT '{}',
  last_checked_at timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON filings (issue_id);
CREATE UNIQUE INDEX ON filings (channel, external_ref) WHERE external_ref IS NOT NULL;
```

---

## 8. Context feeds

```sql
CREATE TABLE observations (
  id            uuid PRIMARY KEY,
  kind          text NOT NULL,             -- rainfall | water_level | aqi | temperature
  source_id     text NOT NULL,
  station       text,
  location      geography(Point,4326),
  observed_at   timestamptz NOT NULL,
  value         numeric,
  unit          text,
  raw           jsonb NOT NULL DEFAULT '{}'
);
SELECT create_hypertable('observations','observed_at');   -- if TimescaleDB is adopted; see ADR 0009
CREATE INDEX ON observations USING gist (location);
CREATE INDEX ON observations (kind, observed_at DESC);

CREATE TABLE news_items (
  id            uuid PRIMARY KEY,
  source_id     text NOT NULL,
  url           text UNIQUE NOT NULL,
  headline      text NOT NULL,
  published_at  timestamptz,
  language      text,
  extract       text,                       -- short extract only; never full text
  entities      jsonb NOT NULL DEFAULT '{}',-- extracted places, authorities, contractors
  location      geography(Point,4326),
  retrieved_at  timestamptz NOT NULL
);
CREATE INDEX ON news_items USING gist (location);

CREATE TABLE issue_context (
  issue_id      uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  kind          text NOT NULL,              -- news | weather | works_notice | history
  ref_type      text, ref_id uuid,
  payload       jsonb NOT NULL DEFAULT '{}',
  relevance     numeric(3,2),
  created_at    timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (issue_id, kind, ref_id)
);
```

---

## 9. Configuration, sources, and audit

```sql
-- Legal constants live here, never as literals in code.
CREATE TABLE legal_constants (
  id            uuid PRIMARY KEY,
  key           text NOT NULL,          -- sla.road_defect.MH | compensation.death.MH | rti.fee.MH
  value         jsonb NOT NULL,
  citation      text NOT NULL,          -- the judgment / rule / notification
  citation_url  text,
  verification  text NOT NULL             -- verified | secondary | unverified; public surfaces need 'verified'
                CHECK (verification IN ('verified','secondary','unverified')),
  effective_from date NOT NULL,
  effective_to  date,
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON legal_constants (key, effective_from DESC);

CREATE TABLE sources (
  id            text PRIMARY KEY,        -- 'mahatenders'
  name          text NOT NULL,
  url           text,
  category      text NOT NULL,
  acquisition   text NOT NULL,           -- scrape | api | rti | manual
  cadence       text,
  licence       text,
  terms_reviewed_on date,
  terms_reviewed_by text,
  robots_ok     boolean,
  status        text NOT NULL DEFAULT 'planned',   -- candidate | planned | reviewing | active | blocked | retired
  blocklist     text[] NOT NULL DEFAULT '{}',      -- fields dropped by the parser at every tier
  last_success_at timestamptz,
  last_error    text
);

CREATE TABLE raw_documents (
  id            uuid PRIMARY KEY,
  source_id     text NOT NULL REFERENCES sources(id),
  url           text,
  archive_key   text NOT NULL,           -- S3, write-once
  sha256        bytea NOT NULL,
  content_type  text,
  bytes         bigint,
  retrieved_at  timestamptz NOT NULL,
  parsed_at     timestamptz,
  parse_status  text,
  captured_by   text NOT NULL,           -- 'ingester:<id>' | 'manual:<account_id>'
  UNIQUE (sha256)
);

CREATE TABLE audit_log (
  id            uuid PRIMARY KEY,
  actor_type    text NOT NULL,
  actor_id      uuid,
  action        text NOT NULL,
  target_type   text, target_id uuid,
  reason        text,
  metadata      jsonb NOT NULL DEFAULT '{}',
  occurred_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE corrections (
  id            uuid PRIMARY KEY,
  target_type   text NOT NULL, target_id uuid NOT NULL,
  field         text NOT NULL,
  old_value     jsonb, new_value jsonb,
  reason        text NOT NULL,
  raised_by     text NOT NULL,           -- 'dispute' | 'internal' | 'source_correction'
  dispute_id    uuid,
  published_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE disputes (
  id            uuid PRIMARY KEY,
  target_type   text NOT NULL, target_id uuid NOT NULL,
  claimant      jsonb NOT NULL,          -- name, entity, contact, capacity
  contested_fact text NOT NULL,
  basis         text NOT NULL,
  evidence_keys text[] NOT NULL DEFAULT '{}',
  status        text NOT NULL DEFAULT 'open',   -- open | upheld | rejected | partial
  outcome       text,
  received_at   timestamptz NOT NULL DEFAULT now(),
  decided_at    timestamptz
);
```

---

## 9A. Phase 0 instruments

Tables the [Phase 0](../05-delivery/07-phase-0-instruments.md) tools need. All of them outlive Phase
0: `works` and `work_changes` become the public change log (U14), `clock_instances` becomes the
deadline wallet (U1), and `report_labels` holds the evaluation sets.

```sql
-- One row per fetch attempt, whatever the outcome. The body is archived only when its hash is new.
CREATE TABLE fetch_log (
  id              uuid PRIMARY KEY,
  source_id       text NOT NULL REFERENCES sources(id),
  endpoint        text NOT NULL,
  requested_at    timestamptz NOT NULL,
  status_code     int,
  bytes           bigint,
  sha256          bytea,
  raw_document_id uuid REFERENCES raw_documents(id),   -- set only when the body was new
  error           text
);
CREATE INDEX ON fetch_log (source_id, endpoint, requested_at DESC);

-- Latest known state of each record in a published works dataset, after the blocklist strip.
CREATE TABLE works (
  id              uuid PRIMARY KEY,
  source_id       text NOT NULL REFERENCES sources(id),
  endpoint        text NOT NULL,
  natural_key     text NOT NULL,           -- fixed per endpoint in code; a collision is an error
  current         jsonb NOT NULL,
  geom            geometry(Geometry,4326),
  first_seen_at   timestamptz NOT NULL,
  last_seen_at    timestamptz NOT NULL,
  vanished_at     timestamptz,
  raw_document_id uuid NOT NULL REFERENCES raw_documents(id),
  UNIQUE (source_id, endpoint, natural_key)
);
CREATE INDEX ON works USING gist (geom);

-- Every observed change, with both values and both source documents.
CREATE TABLE work_changes (
  id                   uuid PRIMARY KEY,
  work_id              uuid NOT NULL REFERENCES works(id),
  kind                 text NOT NULL,      -- added | changed | vanished | reappeared
  field                text,               -- null unless kind = 'changed'
  old_value            jsonb,
  new_value            jsonb,
  prev_raw_document_id uuid REFERENCES raw_documents(id),
  raw_document_id      uuid NOT NULL REFERENCES raw_documents(id),
  observed_at          timestamptz NOT NULL
);
CREATE INDEX ON work_changes (observed_at DESC);
CREATE INDEX ON work_changes (work_id, observed_at DESC);

-- Government resolutions seen by the watcher. The classification is a lead for a human,
-- never a legal_constants value.
CREATE TABLE gr_items (
  sanketank       text PRIMARY KEY,
  issued_on       date,
  raw_document_id uuid REFERENCES raw_documents(id),
  classification  jsonb,                   -- department, subject, any period, fee or time limit
  prompt_version  text,
  model           text,
  reviewed_by     uuid REFERENCES accounts(id),
  reviewed_at     timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE court_items (
  id              uuid PRIMARY KEY,
  source_id       text NOT NULL REFERENCES sources(id),
  cnr             text,
  case_number     text,
  decided_on      date,
  matched_on      text NOT NULL,           -- the watch term that matched
  raw_document_id uuid REFERENCES raw_documents(id),
  UNIQUE (source_id, cnr)
);

-- Human labels on field-kit reports: the classification eval set and the jurisdiction golden set.
CREATE TABLE report_labels (
  report_id         uuid PRIMARY KEY REFERENCES reports(id),
  frame_type        text NOT NULL,         -- close | wide | noticeboard
  label             text NOT NULL,         -- taxonomy subcategory, or 'not_civic'
  conditions        text[] NOT NULL DEFAULT '{}',   -- day | night | rain | motion_blur
  ward_ground_truth text,                  -- golden points only
  labelled_by       uuid NOT NULL REFERENCES accounts(id),
  labelled_at       timestamptz NOT NULL DEFAULT now()
);

-- A running clock. It keeps the legal_constants version it started under.
CREATE TABLE clock_instances (
  id          uuid PRIMARY KEY,
  filing_id   uuid REFERENCES filings(id),
  issue_id    uuid REFERENCES issues(id),
  constant_id uuid NOT NULL REFERENCES legal_constants(id),
  starts_at   timestamptz NOT NULL,
  due_at      timestamptz NOT NULL,
  status      text NOT NULL DEFAULT 'running',   -- running | met | missed | superseded
  CHECK (filing_id IS NOT NULL OR issue_id IS NOT NULL)
);
CREATE INDEX ON clock_instances (due_at) WHERE status = 'running';
```

---

## 10. Materialised views

```sql
CREATE MATERIALIZED VIEW ward_stats AS
SELECT
  w.id AS ward_id, w.code, w.authority_id,
  count(*)                                             AS issues_total,
  count(*) FILTER (WHERE i.status = 'sla_breached')     AS breached,
  count(*) FILTER (WHERE i.status = 'claimed_resolved') AS claimed_resolved,
  count(*) FILTER (WHERE i.status = 'citizen_confirmed')AS confirmed,
  count(*) FILTER (WHERE i.status = 'reopened')         AS reopened,
  percentile_cont(0.5) WITHIN GROUP (
    ORDER BY EXTRACT(EPOCH FROM (coalesce(i.closed_at, now()) - i.first_reported_at))/86400
  )                                                     AS median_age_days
FROM wards w LEFT JOIN issues i ON i.ward_id = w.id
WHERE w.valid_to IS NULL
GROUP BY w.id, w.code, w.authority_id;

CREATE MATERIALIZED VIEW contractor_scorecard AS
SELECT
  c.id AS contractor_id,
  count(DISTINCT ct.id)                                       AS contracts_total,
  sum(ct.value_paise)                                         AS value_total_paise,
  count(DISTINCT m.issue_id) FILTER (WHERE m.in_dlp)          AS defects_in_dlp,
  count(DISTINCT m.issue_id)                                  AS defects_total,
  count(DISTINCT b.id)                                        AS blacklistings,
  max(b.from_date)                                            AS last_blacklisted_on
FROM contractors c
LEFT JOIN contracts ct ON ct.contractor_id = c.id
LEFT JOIN issue_contract_matches m
       ON m.contract_id = ct.id AND m.superseded_by IS NULL AND m.published
LEFT JOIN blacklistings b ON b.contractor_id = c.id
GROUP BY c.id;
```

The **grade** derived from `contractor_scorecard` is computed in application code from a published,
versioned formula — not in SQL — so the formula can be shown to users alongside the grade. See
[tender engine](06-tender-engine.md).

---

## 11. Design notes

| Decision | Rationale |
|---|---|
| `reports` and `issues` separate | Deduplication and corroboration are the whole point |
| `location` vs `location_public` | DPDP: public surfaces never see precise reporter coordinates |
| `issue_events` append-only + hash chain | Evidentiary integrity for exports and filings |
| `issue_contract_matches` versioned with `superseded_by` | Attribution changes must leave a trail |
| `legal_constants` table | Judgments and fees change; code must not |
| `raw_documents` unique on sha256 | Deduplicate the archive; the hash is the citation |
| Native enums | Catch invalid states at write time |
| No ORM | The spatial queries are hand-tuned and the schema is stable |
| `clock_skew_s` generated column | A cheap, always-present fraud signal |

## 12. Open questions

- TimescaleDB for `observations`, or plain partitioning? See [ADR 0009](../04-adr/0009-timeseries.md).
- Do assets get created eagerly on first report, or only when a second report lands nearby?
  *Leaning: lazily on the second report, to avoid an asset table full of one-off noise.*
- Retention for `report_media` archival originals — how long, and under what legal-hold rules?
