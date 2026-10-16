-- 0040: drainage stops going to the water supply counter.
--
-- `water_drainage` has had one desk for everything since migration 0019, and
-- that desk is Water Works — the **water supply** department under the
-- Hydraulic Engineer. Migration 0030 wrote the problem down at the time and
-- could not fix it: "this desk is water *supply*, so burst pipes and leaks
-- belong to it while waterlogging, blocked drains and sewage overflow belong
-- to the maintenance desk… Resolving that needs the ward drainage handbooks."
-- Two documents resolve it now, and migration 0035's `subcategory` is the
-- mechanism that lets the answer be recorded without throwing away the half
-- that was right.
--
--   **The ward**, for what drains off a street. The Assistant Engineer
--   (Maintenance & Repair) handbooks list among the duties of the post:
--   32. "Desilting of minor nallas and road side SWD", 33. "To minimize
--   flooding spots during monsoon", 34. "To provide missing manholes and
--   dhapas", and in the prose, desilting "nallas and road side S.W.D to
--   minimize the flooding spots either departmentally or through the
--   e-tendering". Verbatim in A Ward and in P/North.
--
--   **The Chief Engineer (Storm Water Drains)**, for the network. Manual 1 of
--   its own Section 4 handbook gives its functions as planning and maintenance
--   of the SWD network in City, Eastern and Western Suburbs, desilting of the
--   SWD network, **desilting of major nallas**, improvement of the existing
--   system, and BRIMSTOWAD.
--
--   **Water Works** keeps the category, because the category-wide row is what
--   a subcategory nobody has split falls back to, and burst pipelines and
--   leakage — the two things in this category that are about supply — are
--   exactly what it is right for.
--
-- The major/minor nalla split is real and is not visible in a photograph, the
-- same shape of problem as the bridge span below. A silted nallah is sent to
-- the Storm Water Drains department because a nalla large enough to photograph
-- and worry about is usually on its network, and the ward's own desilting duty
-- is named in the note so the other reading is not lost.
--
-- Ward to division comes from the Bridges handbook's Manual 16, which groups
-- every ward under one of the department's three divisional mailboxes —
-- `_sct` and `_nct` for City, `_sws` and `_nws` for the Western Suburbs,
-- `_ses` and `_nes` for the Eastern Suburbs. The Storm Water Drains department
-- is organised on the same three divisions, each with its own office.
--
-- No personal mobile numbers. The SWD directory prints one against every post;
-- the office landline and address are what a complaint needs, and they outlive
-- the person.

INSERT INTO authority_departments
  (id, authority, ward, category, subcategory, department, officer_designation,
   pio_designation, pio_address, office_phone, escalates_to,
   source_ref, retrieved_at, archive_key, note, effective_from)
SELECT
  gen_random_uuid(), d.authority, d.ward, 'water_drainage', s.subcategory,
  d.department,
  d.officer_designation,
  d.pio_designation,
  d.pio_address,
  d.office_phone,
  d.escalates_to,
  'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
    || 'Wards/Assistant%20Commissioner%20(A-Ward)/RTI%20Manuals/'
    || 'RTI%2017%20Manual%20AE%20MAINT%20A%20Ward.pdf',
  '2026-10-16T00:00:00Z',
  'archive/bmc_portal_rti_manuals/2026/10/16/'
    || '66c55b7bb057ec3723e80ea83bd06679c8f3773a1102f6f8476d8ee2e78006e7.pdf',
  'Duties 32 to 34 of the ward Assistant Engineer (Maintenance & Repair): '
    || 'desilting of minor nallas and road side SWD, minimising flooding spots '
    || 'during monsoon, and providing missing manholes and dhapas. The same '
    || 'numbered duties appear in the P/North handbook. The Chief Engineer '
    || '(Storm Water Drains) holds the network and the major nallas, which is a '
    || 'separate desk; this one is the street.',
  '2026-10-16'
FROM authority_departments d
CROSS JOIN (VALUES ('waterlogging'), ('blocked drain')) AS s(subcategory)
WHERE d.authority = 'BMC'
  AND d.category = 'road_defect'
  AND d.officer_designation LIKE 'Assistant Engineer (Maintenance & Repair)%';

-- A silted nallah, and a bridge of six metres or less, both belong to the
-- Storm Water Drains department, by division.
INSERT INTO authority_departments
  (id, authority, ward, category, subcategory, department, officer_designation,
   pio_designation, pio_address, office_phone, escalates_to,
   source_ref, retrieved_at, archive_key, note, effective_from)
SELECT
  gen_random_uuid(), 'BMC', w.ward, j.category, j.subcategory,
  j.department_prefix || ' — ' || w.division,
  'Deputy Chief Engineer (Storm Water Drains) ' || w.division,
  'Deputy Chief Engineer (Storm Water Drains) ' || w.division,
  o.office,
  o.phone,
  'Chief Engineer (Storm Water Drains), Engineering Services & Projects '
    || 'Office Building, 2nd Floor, 546 N.M. Joshi Marg, Bakri Adda, '
    || 'Byculla (West), Mumbai 400011',
  j.source_ref,
  '2026-10-16T00:00:00Z',
  j.archive_key,
  j.note,
  '2026-10-16'
FROM (VALUES
  ('A',   'City'),             ('B',   'City'),
  ('C',   'City'),             ('D',   'City'),
  ('E',   'City'),             ('F/S', 'City'),
  ('F/N', 'City'),             ('G/S', 'City'),
  ('G/N', 'City'),
  ('H/E', 'Western Suburbs'),  ('H/W', 'Western Suburbs'),
  ('K/E', 'Western Suburbs'),  ('K/W', 'Western Suburbs'),
  ('P/S', 'Western Suburbs'),  ('P/N', 'Western Suburbs'),
  ('R/S', 'Western Suburbs'),  ('R/C', 'Western Suburbs'),
  ('R/N', 'Western Suburbs'),
  ('L',   'Eastern Suburbs'),  ('M/E', 'Eastern Suburbs'),
  ('M/W', 'Eastern Suburbs'),  ('N',   'Eastern Suburbs'),
  ('S',   'Eastern Suburbs'),  ('T',   'Eastern Suburbs')
) AS w(ward, division)
JOIN (VALUES
  ('City',
   'Engineering Services & Projects Office Building, 2nd Floor, 546 N.M. Joshi '
     || 'Marg, Bakri Adda, Byculla (West), Mumbai 400011', '23015817'),
  ('Western Suburbs',
   'K/West Ward Office Building, Terrace Floor, S.V. Road, Andheri (West), '
     || 'Mumbai 400056', '26253594'),
  ('Eastern Suburbs',
   'Pant Nagar Garage Building, 5th Floor, behind the BEST depot, '
     || 'Ghatkopar (East), Mumbai 400077', '25163067')
) AS o(division, office, phone) ON o.division = w.division
CROSS JOIN (VALUES
  ('water_drainage', 'silted nallah', 'Chief Engineer (Storm Water Drains)',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
     || 'Chief%20Engineer%20(Storm%20Water%20Drains)/RTI%20Manuals/'
     || 'Ch%20Engineer_SWD_RTI_E02.pdf',
   'archive/bmc_portal_rti_manuals/2026/10/16/'
     || 'f6b413b5de1a5b1b13b0097524232e2dccf18adc7583e8578ebc430ad6bd964b.pdf',
   'Manual 1 of the department''s Section 4 handbook: planning and maintenance '
     || 'of the SWD network in City, Eastern and Western Suburbs, desilting of '
     || 'the SWD network, and desilting of major nallas. A minor nalla and the '
     || 'road-side drain are the ward''s, under duty 32 of its Assistant '
     || 'Engineer (Maintenance & Repair); the division is not visible in a '
     || 'photograph, so both desks are listed for this ward.'),
  ('structural', 'cracked bridge/FOB',
   'Chief Engineer (Storm Water Drains), spans up to 6 m',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
     || 'Chief%20Engineer%20(Bridges)/Docs/'
     || 'Circular%20(Revised)%20Maint%20of%20Bridge.pdf',
   'archive/bmc_portal_rti_manuals/2026/10/16/'
     || 'aa0db59ce1d11bdde51db6da013c6c0d6f50927d4b92a7a3d82158701fb790b0.pdf',
   'Circular No.Ch.E./25MGC/Bridges dated 28.06.2019, clause 4: the inventory '
     || 'and the major and minor repairs of every bridge or culvert up to and '
     || 'including 6.00 m span belong to the Chief Engineer (S.W.D.). Above '
     || 'that span it is the Bridge department, which holds every footover '
     || 'bridge, flyover and skywalk in the city.')
) AS j(category, subcategory, department_prefix, source_ref, archive_key, note);

-- The Bridge department's rows now say which spans they own, so the two sit
-- side by side on a ward page and a person at a small culvert can read that
-- the other desk is theirs.
UPDATE authority_departments
   SET department = 'Chief Engineer (Bridges), spans over 6 m'
 WHERE authority = 'BMC' AND category = 'structural'
   AND subcategory = 'cracked bridge/FOB'
   AND department = 'Chief Engineer (Bridges)';

DO $$
DECLARE n integer;
BEGIN
  SELECT count(*) INTO n FROM authority_departments
   WHERE authority = 'BMC' AND category = 'water_drainage'
     AND subcategory IN ('waterlogging', 'blocked drain', 'silted nallah');
  IF n <> 72 THEN
    RAISE EXCEPTION 'drainage subcategories seeded % row(s), expected 72', n;
  END IF;
  SELECT count(*) INTO n FROM authority_departments
   WHERE authority = 'BMC' AND category = 'structural'
     AND subcategory = 'cracked bridge/FOB';
  IF n <> 48 THEN
    RAISE EXCEPTION 'bridge desks are now % row(s), expected 48', n;
  END IF;
  SELECT count(*) INTO n FROM authority_departments
   WHERE archive_key IS NULL OR archive_key = '';
  IF n <> 0 THEN
    RAISE EXCEPTION '% row(s) cite no archived document', n;
  END IF;
END $$;
