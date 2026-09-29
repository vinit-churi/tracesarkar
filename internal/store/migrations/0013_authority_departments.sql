-- 0013: which department owns which kind of complaint, and who its PIO is.
--
-- This is the table that decides where a citizen's complaint is sent and who
-- an RTI is addressed to. Both are consequential enough that a row without a
-- source is worse than no row: a complaint at the wrong desk is the failure a
-- citizen does not forgive, and an RTI to the wrong PIO is simply returned.
--
-- So `source_ref` and `retrieved_at` are NOT NULL. There is deliberately no
-- way to record "we think it is Roads & Traffic" — either a document says so
-- and is cited, or the platform states that it does not know.

CREATE TABLE IF NOT EXISTS authority_departments (
  id             uuid PRIMARY KEY,
  authority      text NOT NULL DEFAULT 'BMC',
  ward           text,                        -- null = applies to the whole authority
  category       text NOT NULL,               -- the issue category, e.g. road_defect
  department     text NOT NULL,

  -- Designations only. A named individual is personal data with a short shelf
  -- life; the post outlives the person and is what a complaint should name.
  officer_designation text,
  pio_designation     text,
  pio_address         text,

  -- Provenance. Not optional: see above.
  source_ref     text NOT NULL,
  retrieved_at   timestamptz NOT NULL,
  archive_key    text,
  note           text,

  effective_from date NOT NULL DEFAULT CURRENT_DATE,
  created_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT authority_departments_unique UNIQUE (authority, ward, category, effective_from)
);

CREATE INDEX IF NOT EXISTS authority_departments_lookup_idx
    ON authority_departments (authority, category, ward, effective_from DESC);
