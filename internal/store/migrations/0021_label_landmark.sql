-- 0021: where the surveyor says they are, in words they actually know.
--
-- The ward is what the jurisdiction resolver computes from a GPS fix, and
-- `ward_ground_truth` exists to score that computation. Which means it cannot
-- be filled in from the same GPS fix — a measurement that marks its own
-- homework measures nothing.
--
-- But almost nobody standing on a street in Mumbai knows which lettered ward
-- they are in. Asking for one produces a guess, and a guessed ground truth is
-- worse than none: it scores a correct answer as wrong and nobody can tell
-- which points are sound.
--
-- A street name or a landmark is something the person does know, and it
-- establishes the ward independently of the GPS — from BMC's own ward maps, at
-- a desk, later. So the field work records what the surveyor can actually
-- state, and the ward is resolved from it afterwards.
ALTER TABLE report_labels ADD COLUMN IF NOT EXISTS landmark text;

COMMENT ON COLUMN report_labels.landmark IS
  'Street or landmark as the surveyor stated it. Resolves the ward independently of the GPS fix.';
