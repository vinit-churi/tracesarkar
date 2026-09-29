-- 0011: the review queue for attribution answers.
--
-- System 1's done bar is a precision number a human checked, not a number the
-- join computed about itself. Each row is one point, the answer the join gave
-- for it, and — once a person has looked — whether that answer was right.
--
-- Two kinds of point. A `probe` is generated near known geometry, offset by a
-- few metres to imitate a defect at the road edge seen through a phone's GPS;
-- it has an expected answer, so disagreements surface without anyone clicking.
-- A `report` is a real capture, which has no expected answer at all — only a
-- person can say whether the road named is the road in the photograph.

CREATE TABLE IF NOT EXISTS attribution_reviews (
  id                  uuid PRIMARY KEY,
  kind                text NOT NULL CHECK (kind IN ('probe','report')),
  report_id           uuid REFERENCES reports(id) ON DELETE CASCADE,

  lat                 double precision NOT NULL,
  lon                 double precision NOT NULL,
  accuracy_m          numeric(7,1) NOT NULL,

  -- What the join answered.
  confidence          text NOT NULL CHECK (confidence IN ('high','check_this','none')),
  matched_work_id     uuid REFERENCES works(id) ON DELETE SET NULL,
  matched_work_code   text,
  matched_contractor  text,
  matched_location    text,
  distance_m          double precision,
  basis               text NOT NULL,

  -- For probes: the segment the point was generated from.
  expected_work_id    uuid REFERENCES works(id) ON DELETE SET NULL,

  -- What a person said. Null until reviewed.
  verdict             text CHECK (verdict IN ('correct','wrong','unsure')),
  verdict_note        text,
  reviewed_at         timestamptz,
  reviewed_by         uuid REFERENCES accounts(id) ON DELETE SET NULL,

  created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS attribution_reviews_pending_idx
    ON attribution_reviews (created_at) WHERE verdict IS NULL;
CREATE INDEX IF NOT EXISTS attribution_reviews_verdict_idx
    ON attribution_reviews (verdict);
