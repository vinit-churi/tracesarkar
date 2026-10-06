-- 0020: when the photograph says it was taken.
--
-- A capture carries two times: the one the client reports, and the one the
-- camera wrote into the image. They agree to within seconds when someone
-- photographs a problem and sends it — measured on the first real captures,
-- three seconds apart.
--
-- They disagree by hours when someone picks an older photograph out of their
-- gallery. That matters, because the position attached to a report is the
-- device's position *now*. Pick a photograph taken outside the office this
-- morning and send it from your desk, and the report claims the problem is at
-- your desk — and attribution will name whichever contract covers the desk.
--
-- Recording the photograph's own timestamp is what makes that detectable. It
-- is nullable because plenty of images carry no EXIF at all, and absence is an
-- ordinary outcome rather than a fault.
ALTER TABLE report_media ADD COLUMN IF NOT EXISTS exif_taken_at timestamptz;

COMMENT ON COLUMN report_media.exif_taken_at IS
  'When the image says it was taken, read from its own EXIF. Null when absent.';
