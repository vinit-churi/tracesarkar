-- 0035: a desk may own part of a category, not all of it.
--
-- Migration 0014 established that a category can have more than one owning
-- department, and left the choice between them to the caller. That was enough
-- while the split was by road class, where a photograph cannot tell you which
-- side it falls on anyway. `structural` breaks it, because the split is
-- visible in the photograph:
--
--   cracked bridge/FOB   Chief Engineer (Bridges) — the structural audit and
--                        the repairs of every MCGM bridge above 6 m of span
--   distressed building  the ward's Assistant Engineer (Building & Factory),
--   collapsed wall       Designated Officer under MMC Act s.354 for a ruinous
--   unsafe scaffolding   or dangerous structure
--
-- Returning whichever row happened to be inserted first would send half of
-- these to an office that cannot act. So the table learns what the deadline
-- machinery already knew (D138): an obligation, and now a desk, is recorded
-- with the scope it actually has.
--
-- `subcategory` is null for a desk that owns the whole category, which is
-- every row seeded so far. A row naming one is narrower and wins the lookup.
-- The value is a subcategory from internal/classify/taxonomy.go, checked by a
-- live test rather than a foreign key, because the taxonomy is code and a
-- constraint here would make adding a label a migration.

ALTER TABLE authority_departments ADD COLUMN IF NOT EXISTS subcategory text;

-- The old constraint treated two rows differing only by subcategory as a
-- duplicate. Postgres also treats NULLs as distinct in a UNIQUE constraint, so
-- a plain four-column key would stop catching real duplicates among the
-- category-wide rows. Two partial indexes keep both halves honest.
ALTER TABLE authority_departments
  DROP CONSTRAINT IF EXISTS authority_departments_unique;

CREATE UNIQUE INDEX IF NOT EXISTS authority_departments_category_wide_idx
    ON authority_departments (authority, ward, category, department, effective_from)
 WHERE subcategory IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS authority_departments_subcategory_idx
    ON authority_departments (authority, ward, category, subcategory, department,
                              effective_from)
 WHERE subcategory IS NOT NULL;

-- The lookup reads category first and narrows, so the index leads with it.
CREATE INDEX IF NOT EXISTS authority_departments_subcategory_lookup_idx
    ON authority_departments (authority, category, ward, subcategory,
                              effective_from DESC);
