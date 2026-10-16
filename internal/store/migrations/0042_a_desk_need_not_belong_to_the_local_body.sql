-- 0042: a desk need not belong to the local body.
--
-- Every one of the 289 routing rows says `authority = 'BMC'`, and the lookup
-- filters on it, because until now every desk the platform knew was the
-- corporation's. `environment` ends that. A mangrove is the Forest
-- Department's, an effluent discharge is the Maharashtra Pollution Control
-- Board's — the Board is constituted under section 4 of the Water (Prevention
-- and Control of Pollution) Act 1974 and is the state board under section 5 of
-- the Air (Prevention and Control of Pollution) Act 1981 — and neither is BMC
-- in any sense. A platform that can only name the corporation would have to
-- answer "not known" for things it knows perfectly well.
--
-- The column that makes this work is `covers_authority`: **who the desk serves**,
-- as against `authority`, **who the desk belongs to**. A Pollution Control Board
-- office covering Mumbai carries `authority = 'MPCB'` and
-- `covers_authority = 'BMC'`, so a report resolved to a BMC ward finds it while
-- a report resolved to another corporation's ward with the same letter does
-- not. Null means the two are the same, which is every row seeded so far, so
-- nothing has to be rewritten.
--
-- The lookup is ordered so the narrower answer wins: a desk that names this
-- subcategory beats a desk that owns the whole category, whoever either
-- belongs to. That is the same rule migration 0035 introduced, now applied
-- across authorities too.

ALTER TABLE authority_departments ADD COLUMN IF NOT EXISTS covers_authority text;

COMMENT ON COLUMN authority_departments.covers_authority IS
  'Whose area this desk serves, when that differs from who it belongs to. '
  'Null means the desk belongs to the same body it serves.';

DROP INDEX IF EXISTS authority_departments_subcategory_lookup_idx;
CREATE INDEX IF NOT EXISTS authority_departments_serving_lookup_idx
    ON authority_departments
       (COALESCE(covers_authority, authority), category, ward, subcategory,
        effective_from DESC);

-- The two uniqueness indexes keyed on `authority`, which is still right: two
-- rows from different bodies for the same ward and subcategory are not
-- duplicates, they are the split the table now has to express.

-- The pollution control board, as a source. Two of its pages were read by
-- hand on 16 October 2026 — the office list and the home page's statement of
-- the Board's statutory mandate. Nothing is scraped and nothing is on a
-- schedule; `acquisition` says so.
INSERT INTO sources (id, name, url, category, acquisition, cadence, status,
                     licence, updated_at)
VALUES ('mpcb',
        'Maharashtra Pollution Control Board',
        'https://mpcb.gov.in/',
        'environment',
        'manual',
        'on demand',
        'active',
        NULL,
        now())
ON CONFLICT (id) DO NOTHING;
