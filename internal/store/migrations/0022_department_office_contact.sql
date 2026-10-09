-- 0022: how to actually reach a desk.
--
-- The platform records the post rather than the person holding it, because a
-- name has a short shelf life and a complaint should address the office. That
-- was right and it was also incomplete: a designation with no telephone, no
-- email and no visiting hours tells a citizen who is responsible and gives
-- them no way to act on it.
--
-- BMC publishes all of it in the same Section 4(1)(b) handbooks, attached to
-- the post and not to the individual — an office landline, a role mailbox, the
-- hours the office is open, the hours a member of the public may walk in, and
-- the office this one reports to. None of it is personal data, and all of it
-- survives a transfer.
--
-- That last column is the escalation. "Reporting to which office" is the next
-- rung, published by the body itself, and it does not go stale when somebody
-- moves on.
ALTER TABLE authority_departments
  ADD COLUMN IF NOT EXISTS office_phone  text,
  ADD COLUMN IF NOT EXISTS office_email  text,
  ADD COLUMN IF NOT EXISTS office_hours  text,
  ADD COLUMN IF NOT EXISTS visiting_hours text,
  ADD COLUMN IF NOT EXISTS escalates_to  text;

COMMENT ON COLUMN authority_departments.office_phone IS
  'Office landline as published. Never an individual''s mobile.';
COMMENT ON COLUMN authority_departments.office_email IS
  'Role mailbox as published. Never an individual''s address.';
COMMENT ON COLUMN authority_departments.escalates_to IS
  'The office this one reports to, from the handbook''s own "reporting to" row.';

-- Verified first-hand on 9 October 2026 from the copies already archived in
-- R2 on 3 October, not from the live site.

UPDATE authority_departments SET
  office_phone   = '022-28946000',
  office_email   = 'ae01swm.rc@mcgm.gov.in',
  office_hours   = 'Monday to Friday 8.00 a.m.–12.00 noon and 2.30–5.30 p.m.; Saturday 8.00–11.30 a.m.',
  visiting_hours = 'Chowky 6.30 a.m. – 1.15 p.m.',
  escalates_to   = 'Assistant Commissioner, R/Central Ward'
 WHERE authority = 'BMC' AND ward = 'R/C' AND category = 'waste';

UPDATE authority_departments SET
  office_phone   = '022-28901344 extn. 227',
  office_email   = 'aewwrc@gmail.com',
  office_hours   = 'Monday to Friday 8.00 a.m.–12.00 noon and 2.30–5.30 p.m.; Saturday 8.00–11.30 a.m.',
  visiting_hours = '3.00–5.00 p.m., Monday to Friday',
  escalates_to   = 'Assistant Commissioner, R/Central Ward; then Dy. Hydraulic Engineer (Western Suburb), K/West ward office, 4th Floor, Paliram Road, Andheri (W), Mumbai 400058'
 WHERE authority = 'BMC' AND ward = 'R/C' AND category = 'water_drainage';

-- The two road rows are deliberately left null. Their handbook has not been
-- re-read for these fields, and an office telephone guessed at is worse than
-- one absent: a citizen rings it, nobody answers, and they conclude the
-- platform is wrong about everything else too.
