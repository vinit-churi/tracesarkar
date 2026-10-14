-- 0026: what a person can see in a photograph, so the detector can be graded.
--
-- ADR 0008 required an adversarial evaluation set before launch — faces at
-- distance, partial occlusion, reflections, plates at angle, night, and Indian
-- plate formats. The detector shipped without one, and so its recall was
-- unknown: every capture it had been run over happened to contain no face and
-- no plate, which means it had never had the chance to miss anything.
--
-- This table is that set. One row per photograph a person has marked up, with
-- every face and plate they could see, in fractions of the image so the
-- annotation survives the photograph being resized.
--
-- An empty box list is a judgement, not a blank: it says a person looked and
-- there is nothing here to find. That distinction is the whole point of the
-- row existing, because a set of photographs with nothing in them can only
-- measure false positives, and reporting recall over it would flatter the
-- detector exactly where it matters most.

CREATE TABLE IF NOT EXISTS redaction_truth (
  media_id       uuid PRIMARY KEY REFERENCES report_media(id) ON DELETE CASCADE,
  report_id      uuid NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
  boxes          jsonb NOT NULL DEFAULT '[]'::jsonb,
  -- Whether a person is visible at all, which is not the same as a face being
  -- visible: somebody photographed from behind is present with no face to
  -- mark. The gate holds an image on that contradiction, so the set has to be
  -- able to state it.
  people_present boolean NOT NULL DEFAULT false,
  conditions     text[] NOT NULL DEFAULT '{}',
  notes          text NOT NULL DEFAULT '',
  annotated_by   uuid REFERENCES accounts(id) ON DELETE SET NULL,
  annotated_at   timestamptz NOT NULL DEFAULT now()
);

-- A box list that is not a list would break the scorer at the point it is read
-- rather than at the point it was written.
ALTER TABLE redaction_truth
  DROP CONSTRAINT IF EXISTS redaction_truth_boxes_is_a_list;
ALTER TABLE redaction_truth
  ADD CONSTRAINT redaction_truth_boxes_is_a_list CHECK (jsonb_typeof(boxes) = 'array');

-- A face marked up means a person is present. Recording the opposite would
-- make the contradiction the gate relies on untestable. Written with the
-- containment operator rather than a subquery, which a CHECK cannot hold.
ALTER TABLE redaction_truth
  DROP CONSTRAINT IF EXISTS redaction_truth_a_face_means_a_person;
ALTER TABLE redaction_truth
  ADD CONSTRAINT redaction_truth_a_face_means_a_person CHECK (
    people_present OR NOT (boxes @> '[{"kind": "face"}]'::jsonb)
  );

CREATE INDEX IF NOT EXISTS redaction_truth_annotated_at_idx
    ON redaction_truth (annotated_at);
