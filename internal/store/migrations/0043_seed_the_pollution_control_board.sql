-- 0043: the first desks that are not the corporation's.
--
-- An effluent discharge and an air pollution source are the Maharashtra
-- Pollution Control Board's. The Board is constituted under section 4 of the
-- Water (Prevention and Control of Pollution) Act 1974 and is the state board
-- under section 5 of the Air (Prevention and Control of Pollution) Act 1981 —
-- its own site states both — so neither thing is BMC's in any sense, and
-- migration 0042's `covers_authority` is what lets the platform say so
-- without pretending the office belongs to the corporation.
--
-- The Board publishes one Regional Office for Mumbai, whose jurisdiction it
-- states as "Mumbai Municipal Corporation Area", and four Sub-Regional
-- Offices under it, each with a ward list. The lists are reproduced exactly:
--
--   S.R.O. Mumbai-I    "Mumbai Island , Ward No. A.B.C D. FF(South) F(North)
--                       G(South) and G(North)."
--   S.R.O. Mumbai-II   "Part Of Mumbai Suburb, Ward No. M(East) M(West),
--                       H(East) H(West) and L."
--   S.R.O. Mumbai-III  "Part Of Mumbai Suburb, Ward No. K(East) K(West), S, N,
--                       and (P (South)."
--   S.R.O. Mumbai-IV   "Suburb of Mumbai, Ward No. P(North), R(North),
--                       R(South) and T."
--
-- Twenty-two wards. **E Ward and R/Central appear in none of the four lists.**
-- That is the Board's gap, not a transcription slip, and it is not filled by
-- inference — E would "obviously" be Mumbai-I and R/Central "obviously"
-- Mumbai-IV, and obviously is not a source. Those two wards are given the
-- Regional Office instead, whose stated jurisdiction is the whole Mumbai
-- Municipal Corporation Area and therefore covers them on the Board's own
-- wording.
--
-- All five offices share one address and one pair of telephone numbers on the
-- Board's page; the mailboxes differ, and the mailbox is what a complaint
-- actually travels on. The Board's older jurisdiction PDF gives a different
-- address at M.G. Road and an older revision of this page gives a different
-- floor; neither is used, and the discrepancy is recorded here so that a
-- future reader finds it already noticed rather than appearing to correct us.
--
-- `escalates_to` names the Regional Office above each Sub-Regional one, which
-- the page's own structure establishes. Above the Regional Office the Board's
-- own hierarchy is not stated on the page, so the chain stops there rather
-- than guessing at a Member Secretary.
--
-- What is deliberately **not** seeded: mangrove destruction, creek dumping and
-- encroachment on a water body. Mangroves are the Forest Department's Mangrove
-- Cell and the coastal zone authority's, and dumping in a creek is shared
-- between the Board and the corporation on a basis no document in hand
-- settles. Three subcategories of `environment` still answer "not known",
-- which is the honest state.

INSERT INTO authority_departments
  (id, authority, covers_authority, ward, category, subcategory, department,
   officer_designation, pio_designation, pio_address, office_phone,
   office_email, escalates_to,
   source_ref, retrieved_at, archive_key, note, effective_from)
SELECT
  gen_random_uuid(), 'MPCB', 'BMC', m.ward, 'environment', s.subcategory,
  m.office,
  m.officer,
  m.officer,
  'Kalpataru Point, 2nd floor, opposite MovieMax Theatre, Sion (East), '
    || 'Mumbai 400022',
  m.phone,
  m.email,
  m.escalates_to,
  'https://mpcb.gov.in/en/about-us/offices/mumbai',
  '2026-10-16T00:00:00Z',
  'archive/mpcb/2026/10/16/'
    || 'd2175c0523fd9dd7406b0b0ca255269dd989fd95df90deb6e6bd57c5eaf616eb.html',
  s.note || ' ' || m.note,
  '2026-10-16'
FROM (VALUES
  ('A',   'S.R.O. Mumbai-I',  'Sub-Regional Officer, Mumbai-I',  '24016239', 'sromumbai1@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Mumbai Island , Ward No. A.B.C D. FF(South) F(North) G(South) and G(North)."'),
  ('B',   'S.R.O. Mumbai-I',  'Sub-Regional Officer, Mumbai-I',  '24016239', 'sromumbai1@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Mumbai Island , Ward No. A.B.C D. FF(South) F(North) G(South) and G(North)."'),
  ('C',   'S.R.O. Mumbai-I',  'Sub-Regional Officer, Mumbai-I',  '24016239', 'sromumbai1@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Mumbai Island , Ward No. A.B.C D. FF(South) F(North) G(South) and G(North)."'),
  ('D',   'S.R.O. Mumbai-I',  'Sub-Regional Officer, Mumbai-I',  '24016239', 'sromumbai1@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Mumbai Island , Ward No. A.B.C D. FF(South) F(North) G(South) and G(North)."'),
  ('F/S', 'S.R.O. Mumbai-I',  'Sub-Regional Officer, Mumbai-I',  '24016239', 'sromumbai1@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Mumbai Island , Ward No. A.B.C D. FF(South) F(North) G(South) and G(North)."'),
  ('F/N', 'S.R.O. Mumbai-I',  'Sub-Regional Officer, Mumbai-I',  '24016239', 'sromumbai1@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Mumbai Island , Ward No. A.B.C D. FF(South) F(North) G(South) and G(North)."'),
  ('G/S', 'S.R.O. Mumbai-I',  'Sub-Regional Officer, Mumbai-I',  '24016239', 'sromumbai1@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Mumbai Island , Ward No. A.B.C D. FF(South) F(North) G(South) and G(North)."'),
  ('G/N', 'S.R.O. Mumbai-I',  'Sub-Regional Officer, Mumbai-I',  '24016239', 'sromumbai1@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Mumbai Island , Ward No. A.B.C D. FF(South) F(North) G(South) and G(North)."'),
  ('M/E', 'S.R.O. Mumbai-II', 'Sub-Regional Officer, Mumbai-II', '24015269', 'sromumbai2@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. M(East) M(West), H(East) H(West) and L."'),
  ('M/W', 'S.R.O. Mumbai-II', 'Sub-Regional Officer, Mumbai-II', '24015269', 'sromumbai2@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. M(East) M(West), H(East) H(West) and L."'),
  ('H/E', 'S.R.O. Mumbai-II', 'Sub-Regional Officer, Mumbai-II', '24015269', 'sromumbai2@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. M(East) M(West), H(East) H(West) and L."'),
  ('H/W', 'S.R.O. Mumbai-II', 'Sub-Regional Officer, Mumbai-II', '24015269', 'sromumbai2@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. M(East) M(West), H(East) H(West) and L."'),
  ('L',   'S.R.O. Mumbai-II', 'Sub-Regional Officer, Mumbai-II', '24015269', 'sromumbai2@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. M(East) M(West), H(East) H(West) and L."'),
  ('K/E', 'S.R.O. Mumbai-III','Sub-Regional Officer, Mumbai-III','24016239', 'sromumbai3@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. K(East) K(West), S, N, and (P (South)."'),
  ('K/W', 'S.R.O. Mumbai-III','Sub-Regional Officer, Mumbai-III','24016239', 'sromumbai3@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. K(East) K(West), S, N, and (P (South)."'),
  ('S',   'S.R.O. Mumbai-III','Sub-Regional Officer, Mumbai-III','24016239', 'sromumbai3@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. K(East) K(West), S, N, and (P (South)."'),
  ('N',   'S.R.O. Mumbai-III','Sub-Regional Officer, Mumbai-III','24016239', 'sromumbai3@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. K(East) K(West), S, N, and (P (South)."'),
  ('P/S', 'S.R.O. Mumbai-III','Sub-Regional Officer, Mumbai-III','24016239', 'sromumbai3@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Part Of Mumbai Suburb, Ward No. K(East) K(West), S, N, and (P (South)."'),
  ('P/N', 'S.R.O. Mumbai-IV', 'Sub-Regional Officer, Mumbai-IV', '24016239', 'sromumbai4@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Suburb of Mumbai, Ward No. P(North), R(North), R(South) and T."'),
  ('R/N', 'S.R.O. Mumbai-IV', 'Sub-Regional Officer, Mumbai-IV', '24016239', 'sromumbai4@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Suburb of Mumbai, Ward No. P(North), R(North), R(South) and T."'),
  ('R/S', 'S.R.O. Mumbai-IV', 'Sub-Regional Officer, Mumbai-IV', '24016239', 'sromumbai4@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Suburb of Mumbai, Ward No. P(North), R(North), R(South) and T."'),
  ('T',   'S.R.O. Mumbai-IV', 'Sub-Regional Officer, Mumbai-IV', '24016239', 'sromumbai4@mpcb.gov.in', 'Regional Officer, Mumbai', 'Jurisdiction as published: "Suburb of Mumbai, Ward No. P(North), R(North), R(South) and T."'),
  -- The two wards the Board's own sub-regional lists leave out.
  ('E',   'Regional Office, Mumbai', 'Regional Officer, Mumbai', '24016239', 'romumbai@mpcb.gov.in', NULL, 'This ward appears in none of the four Sub-Regional Office jurisdiction lists the Board publishes. The Regional Office is named instead, whose stated jurisdiction is the whole "Mumbai Municipal Corporation Area"; which Sub-Regional Office holds it is not published and is not guessed here.'),
  ('R/C', 'Regional Office, Mumbai', 'Regional Officer, Mumbai', '24016239', 'romumbai@mpcb.gov.in', NULL, 'This ward appears in none of the four Sub-Regional Office jurisdiction lists the Board publishes. The Regional Office is named instead, whose stated jurisdiction is the whole "Mumbai Municipal Corporation Area"; which Sub-Regional Office holds it is not published and is not guessed here.')
) AS m(ward, office, officer, phone, email, escalates_to, note)
CROSS JOIN (VALUES
  ('effluent discharge',
   'The Maharashtra Pollution Control Board is constituted under section 4 of '
     || 'the Water (Prevention and Control of Pollution) Act 1974 and consents '
     || 'to, and acts on, discharges of effluent.'),
  ('air pollution source',
   'The Maharashtra Pollution Control Board is the state board under section 5 '
     || 'of the Air (Prevention and Control of Pollution) Act 1981; the whole '
     || 'state has been an air pollution control area since 06.11.1996.')
) AS s(subcategory, note);

DO $$
DECLARE n integer;
BEGIN
  SELECT count(*) INTO n FROM authority_departments
   WHERE authority = 'MPCB' AND covers_authority = 'BMC';
  IF n <> 48 THEN
    RAISE EXCEPTION 'pollution board seeded % row(s), expected 48', n;
  END IF;
  -- Every BMC ward must reach the board, including the two its own lists omit.
  SELECT count(DISTINCT ward) INTO n FROM authority_departments
   WHERE authority = 'MPCB';
  IF n <> 24 THEN
    RAISE EXCEPTION 'pollution board covers % ward(s), expected 24', n;
  END IF;
  SELECT count(*) INTO n FROM authority_departments
   WHERE archive_key IS NULL OR archive_key = '';
  IF n <> 0 THEN
    RAISE EXCEPTION '% row(s) cite no archived document', n;
  END IF;
END $$;
