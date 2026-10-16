-- 0036: the two desks behind `structural`.
--
-- `structural` holds a cracked bridge or footover bridge, a distressed
-- building, unsafe scaffolding and a collapsed wall. One category, two
-- authorities, and unlike the road-class split of migration 0014 this one is
-- visible in the photograph — so migration 0035 taught the table to record a
-- desk with the scope it has, and these rows use it.
--
-- **The ward, for anything that is a structure.** The ward's Assistant
-- Engineer (Building & Factory) is the Designated Officer appointed by the
-- Municipal Commissioner under MMC Act sections 351, 352, 352A and **354 —
-- "to require removal etc. of ruinous or dangerous structures and to
-- temporarily secure etc."** The A-ward handbook for the post lists among its
-- duties "Inspection of private dilapidated buildings and taking legal actions
-- for repairs or demolition", serving notices for pulling down dangerous parts
-- of a building, and attending collapses in an emergency. It names the office
-- it answers to: "Reporting to which office — Assistant Commissioner, A Ward",
-- with the Ward Executive Engineer for technical matters.
--
-- **The Bridge department, for a bridge.** BMC's revised circular
-- No.Ch.E./25MGC/Bridges dated 28.06.2019, "Revised guidelines for safety
-- measures and maintenance of Bridge", divides the work by span:
--
--   above 6.00 m span          Chief Engineer (Bridges) — structural audit,
--                              major and minor repairs, and the proposal for
--                              demolition and reconstruction
--   up to and including 6.00 m Chief Engineer (Storm Water Drains)
--   on H.E. land               Hydraulic Engineer
--   private layouts            the co-operative housing society or plot owner
--
-- and gives the ward Assistant Commissioner the housekeeping on any of them:
-- cleaning, rain-water outlets, removing encroachments and hoardings, and
-- carrying out a closure or demolition when the Bridge department directs it.
--
-- Span is the discriminator and a resident cannot measure it from a
-- photograph. The rows below send a cracked bridge to the Bridge department,
-- which is right for every footover bridge and every flyover and wrong only
-- for a culvert of six metres or less — a case the note names, so the Storm
-- Water Drains desk can be added when its own document is read rather than
-- guessed at now. The department's Manual 17 gives the inventory this is
-- about: 452 MCGM bridges, of which 67 are footover bridges, 40 are footover
-- bridges at railway level, 47 are rail over-bridges and 23 are skywalks.
--
-- Manual 16 names an Assistant Engineer (Bridges) for every ward with a role
-- mailbox, grouped — five wards share the City mailbox, four the north-city
-- one, six the south-west-suburb one, and so on. The grouping is reproduced
-- exactly; no mailbox is attributed to a ward the table does not cover.
--
-- Escalation for the bridge rows is the Chief Engineer (Bridges), not the
-- Executive Engineer immediately above the Assistant Engineer. The handbook's
-- appellate table lists Executive Engineers by division — "(Western Sub.)
-- R/South, P/North", "(Western Sub.) R/Centre, R/North" — and those divisions
-- do not cover all twenty-four wards between them. Naming the head of the
-- department is a rung too high; naming the wrong Executive Engineer would be
-- an appeal that comes back.

-- The ward's Building & Factory desk. Derived from the maintenance rows rather
-- than retyped: the handbook gives the same ward-office address for both posts
-- ("A ward office, 134-E, S.B.S. Road, Fort, Mumbai 400 001"), and two desks
-- in one building should not be able to drift apart in this table.
INSERT INTO authority_departments
  (id, authority, ward, category, subcategory, department, officer_designation,
   pio_designation, pio_address, office_phone, escalates_to,
   source_ref, retrieved_at, archive_key, note, effective_from)
SELECT
  gen_random_uuid(), d.authority, d.ward, 'structural', NULL,
  replace(d.department, 'Maintenance & Repair', 'Building & Factory'),
  replace(d.officer_designation, '(Maintenance & Repair)', '(Building & Factory)'),
  replace(d.officer_designation, '(Maintenance & Repair)', '(Building & Factory)'),
  d.pio_address,
  d.office_phone,
  'Assistant Commissioner, ' || d.ward || ' Ward',
  'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
    || 'Wards/Assistant%20Commissioner%20(A-Ward)/RTI%20Manuals/'
    || 'RTI%2017%20Manual%20A.E.%20B.%20&%20F.%20for%20A%20ward.pdf',
  '2026-10-16T00:00:00Z',
  NULL,
  'Designated Officer under MMC Act ss.351, 352, 352A and 354 — removal of '
    || 'ruinous or dangerous structures. The post''s Section 4(1)(b) handbook '
    || 'lists inspection of private dilapidated buildings, notices to pull down '
    || 'dangerous parts, and attending collapses in an emergency, and names the '
    || 'Assistant Commissioner as the office it reports to. Office and telephone '
    || 'are this ward''s, from the consolidated PIO list ([migration 0029]); the '
    || 'handbook gives the same ward-office address for this post.',
  '2026-10-16'
FROM authority_departments d
WHERE d.authority = 'BMC'
  AND d.category = 'road_defect'
  AND d.officer_designation LIKE 'Assistant Engineer (Maintenance & Repair)%';

-- The Bridge department, for a bridge.
INSERT INTO authority_departments
  (id, authority, ward, category, subcategory, department, officer_designation,
   pio_designation, pio_address, office_phone, office_email, escalates_to,
   source_ref, retrieved_at, archive_key, note, effective_from)
SELECT
  gen_random_uuid(), 'BMC', m.ward, 'structural', 'cracked bridge/FOB',
  'Chief Engineer (Bridges)',
  'Assistant Engineer (Bridges), ' || m.ward || ' Ward',
  'Assistant Engineer (Bridges), ' || m.ward || ' Ward',
  'Chief Engineer (Bridges), Engineering Hub Building, Ground Floor, '
    || 'near Worli Naka, Dr. E. Moses Road, Worli, Mumbai 400018',
  '24958001',
  m.email,
  'Chief Engineer (Bridges)',
  'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
    || 'Chief%20Engineer%20(Bridges)/Docs/'
    || 'Circular%20(Revised)%20Maint%20of%20Bridge.pdf',
  '2026-10-16T00:00:00Z',
  NULL,
  'Circular No.Ch.E./25MGC/Bridges dated 28.06.2019: the Chief Engineer '
    || '(Bridges) carries out the structural audit of bridges above 6.00 m span '
    || 'and the major and minor repairs that follow from it. A bridge of 6.00 m '
    || 'span or less belongs to the Chief Engineer (Storm Water Drains) under '
    || 'the same circular, and one on private land to the society that owns it — '
    || 'neither is recorded yet, so a small culvert will be sent here wrongly. '
    || 'Cleaning, rain-water outlets, encroachments and hoardings on any bridge '
    || 'are the ward Assistant Commissioner''s under clause 2. The officer and '
    || 'the mailbox are from Manual 16 of the department''s Section 4 handbook, '
    || 'which groups wards: ' || m.grouping || '.',
  '2026-10-16'
FROM (VALUES
  ('A',   'ae01abe_sct.br@mcgm.gov.in',     'A, B, C, D, E Ward'),
  ('B',   'ae01abe_sct.br@mcgm.gov.in',     'A, B, C, D, E Ward'),
  ('C',   'ae01abe_sct.br@mcgm.gov.in',     'A, B, C, D, E Ward'),
  ('D',   'ae01abe_sct.br@mcgm.gov.in',     'A, B, C, D, E Ward'),
  ('E',   'ae01abe_sct.br@mcgm.gov.in',     'A, B, C, D, E Ward'),
  ('F/N', 'ae01fngn_nct.br@mcgm.gov.in',    'F/North and G/North Ward, sharing a mailbox with G/South and F/South Ward'),
  ('F/S', 'ae01fngn_nct.br@mcgm.gov.in',    'G/South and F/South Ward, sharing a mailbox with F/North and G/North Ward'),
  ('G/N', 'ae01fngn_nct.br@mcgm.gov.in',    'F/North and G/North Ward, sharing a mailbox with G/South and F/South Ward'),
  ('G/S', 'ae01fngn_nct.br@mcgm.gov.in',    'G/South and F/South Ward, sharing a mailbox with F/North and G/North Ward'),
  ('H/E', 'ae01hehw_sws.br@mcgm.gov.in',    'H/East, H/West and K/East Ward'),
  ('H/W', 'ae01hehw_sws.br@mcgm.gov.in',    'H/East, H/West and K/East Ward'),
  ('K/E', 'ae01hehw_sws.br@mcgm.gov.in',    'H/East, H/West and K/East Ward'),
  ('K/W', 'ae01hehw_sws.br@mcgm.gov.in',    'K/West, P/North and P/South Ward, sharing a mailbox with H/East, H/West and K/East Ward'),
  ('P/N', 'ae01hehw_sws.br@mcgm.gov.in',    'K/West, P/North and P/South Ward, sharing a mailbox with H/East, H/West and K/East Ward'),
  ('P/S', 'ae01hehw_sws.br@mcgm.gov.in',    'K/West, P/North and P/South Ward, sharing a mailbox with H/East, H/West and K/East Ward'),
  ('R/C', 'ae01rsrcrn_nws.br@mcgm.gov.in',  'R/Centre West, R/Centre East, R/North and R/South Ward'),
  ('R/N', 'ae01rsrcrn_nws.br@mcgm.gov.in',  'R/Centre West, R/Centre East, R/North and R/South Ward'),
  ('R/S', 'ae01rsrcrn_nws.br@mcgm.gov.in',  'R/Centre West, R/Centre East, R/North and R/South Ward'),
  ('L',   'ae01lses.br@mcgm.gov.in',        'L Ward'),
  ('M/E', 'ae01lses.br@mcgm.gov.in',        'M/East and M/West Ward'),
  ('M/W', 'ae01lses.br@mcgm.gov.in',        'M/East and M/West Ward'),
  ('N',   'ae01n_nes.br@mcgm.gov.in',       'N Ward'),
  ('S',   'ae01n_nes.br@mcgm.gov.in',       'S Ward'),
  ('T',   'ae01n_nes.br@mcgm.gov.in',       'T Ward')
) AS m(ward, email, grouping);

-- Twenty-four of each, or the migration is wrong about something.
DO $$
DECLARE wide integer; bridges integer;
BEGIN
  SELECT count(*) INTO wide FROM authority_departments
   WHERE authority = 'BMC' AND category = 'structural' AND subcategory IS NULL;
  SELECT count(*) INTO bridges FROM authority_departments
   WHERE authority = 'BMC' AND category = 'structural'
     AND subcategory = 'cracked bridge/FOB';
  IF wide <> 24 OR bridges <> 24 THEN
    RAISE EXCEPTION 'structural seeded % ward desks and % bridge desks, expected 24 of each',
      wide, bridges;
  END IF;
END $$;
