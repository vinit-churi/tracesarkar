-- 0027: the social accounts a generated post may tag.
--
-- A handle is a fact about a named party, so hard rule 2 applies to it exactly
-- as it applies to a contract value: `source_ref` and `retrieved_at` are NOT
-- NULL and there is deliberately no way to record "we think this is BMC's
-- account". Tagging the wrong account is not a cosmetic error — it points a
-- stream of civic complaints at a stranger.
--
-- Two kinds, and the difference is enforced here rather than in the template:
--
--   * institutional — the authority's own desk. Safe on a first report.
--   * escalation    — a minister or a chief minister's office. Unlocked by an
--                     elapsed deadline and never earlier. Tagging them on
--                     every pothole trains everybody to ignore the platform
--                     (share kit §4).
--
-- `verify_by` exists because handles get renamed, merged and abandoned, and a
-- registry nobody re-checks is a registry that eventually tags somebody else.
-- A row past its date is not served.

CREATE TABLE IF NOT EXISTS authority_handles (
  id             uuid PRIMARY KEY,
  authority      text NOT NULL,
  ward           text,                        -- null = applies to the whole authority
  platform       text NOT NULL,               -- 'x', 'whatsapp', ...
  handle         text NOT NULL,
  kind           text NOT NULL,

  -- Provenance. Not optional: see above.
  source_ref     text NOT NULL,
  retrieved_at   timestamptz NOT NULL,
  -- Whether the authority itself published this handle, as opposed to it
  -- being attested by somebody else. Printed in the source list so a reader
  -- can weigh it; never used to suppress a handle.
  first_hand     boolean NOT NULL DEFAULT false,
  note           text,

  effective_from date NOT NULL DEFAULT CURRENT_DATE,
  verify_by      date NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT authority_handles_kind_is_known
    CHECK (kind IN ('institutional', 'escalation')),
  -- A handle that is not a handle cannot be tagged, and an empty string here
  -- would render as a bare dash in a public post.
  CONSTRAINT authority_handles_x_names_an_account
    CHECK (platform <> 'x' OR handle ~ '^@[A-Za-z0-9_]{1,15}$'),
  CONSTRAINT authority_handles_source_is_not_blank
    CHECK (btrim(source_ref) <> ''),
  -- A row that is already due for re-verification on the day it is written is
  -- a row nobody will come back to.
  CONSTRAINT authority_handles_verify_by_is_in_the_future
    CHECK (verify_by > effective_from),
  CONSTRAINT authority_handles_unique
    UNIQUE (authority, ward, platform, handle, effective_from)
);

CREATE INDEX IF NOT EXISTS authority_handles_lookup_idx
    ON authority_handles (authority, platform, ward, effective_from DESC);

-- BMC's account on X.
--
-- Marked `first_hand = false` on purpose. BMC's own portal does not publish
-- its social handles anywhere we could find on 14 October 2026, so this rests
-- on press coverage that quotes the account directly, including BMC's own
-- rebuttals of fake school-closure posts. That is good enough to tag and not
-- good enough to call first-hand, and the share kit prints the distinction.
--
-- No ward handles are seeded. Twenty-four ward accounts were reported in 2019
-- and the share kit spec says to verify existence before use; none has been.
-- No escalation handles are seeded either — nothing has needed one yet.
INSERT INTO authority_handles
  (id, authority, ward, platform, handle, kind, source_ref, retrieved_at,
   first_hand, note, effective_from, verify_by)
VALUES
  (gen_random_uuid(), 'BMC', NULL, 'x', '@mybmc', 'institutional',
   'https://www.socialsamosa.com/2019/09/mumbaitweets-mumbai-bmc-twitter-strategy '
     || '— and BMC''s own 2026 rebuttals of fake school-closure posts, reported at '
     || 'https://www.freepressjournal.in/mumbai/mumbai-schools-closed-tomorrow-bmc-debunks-fake-messages-claiming-holiday-on-wednesday-august-20',
   '2026-10-14T00:00:00Z', false,
   'Secondary sourcing only: portal.mcgm.gov.in publishes no social links. '
     || 'Re-verify first-hand before the date below.',
   '2026-10-14', '2027-04-14')
ON CONFLICT DO NOTHING;
