-- 0041: sewage overflow, the last thing in `water_drainage` at the wrong desk.
--
-- Migration 0040 moved waterlogging, blocked drains and silted nallahs off the
-- water supply counter and left one behind. The Chief Engineer (Sewerage
-- Operation) handbook, Manual 1, is unambiguous about whose it is:
--
--   "Sewerage Operation department mainly carries out work of operation and
--   maintenance of Sewerage System in Municipal limits… operates and maintains
--   the Sewerage Network, Sewage Pumping Stations, Sewage Treatment Plants
--   (STP) and Waste Water Treatment Facilities (WWTF)… in seven Sewerage Zones
--   of Mumbai City."
--
-- and, against the Executive Engineer (Mechanical) for main and pipe main
-- sewers, "Overall Control & Monitoring of Smooth Functioning of Sewerage
-- Network… by way of systematic cleaning & attending day-to-day complaints
-- related to sewer lines". A sewer backing up into a street is a day-to-day
-- complaint related to a sewer line.
--
-- One citywide desk rather than three divisional ones, and that is a
-- limitation written down rather than papered over. The department works in
-- three divisions — City from Colaba to Sion on the east and Mahim on the
-- west, Western from Bandra to Dahisar, Eastern from Kurla to Mulund — and its
-- directory gives the City office at the Dadar pumping station and the Western
-- one at Versova. The Eastern office does not appear in the English directory
-- at all, and seeding two divisions out of three would leave six wards
-- pointing at an office chosen by elimination. The head of department's
-- address, landline and role mailbox are complete and correct for every ward,
-- so that is what the rows carry until the Marathi directory or a later
-- revision supplies the third.
--
-- `escalates_to` is null, for the reason migration 0030 gave: the handbook
-- does not state what sits above the Chief Engineer, and an appeal sent to the
-- wrong office comes back. The ward's part is in the note instead, because it
-- is real and a resident may be sent there first: ward offices raise proposals
-- for repairs to pipe sewer lines and street connections, which this
-- department verifies and sanctions.
--
-- No personal numbers. The directory prints a mobile against several posts and
-- a private gmail address against one; none of that belongs in this table.

INSERT INTO authority_departments
  (id, authority, ward, category, subcategory, department, officer_designation,
   pio_designation, pio_address, office_phone, office_email,
   source_ref, retrieved_at, archive_key, note, effective_from)
SELECT
  gen_random_uuid(), 'BMC', w.ward, 'water_drainage', 'sewage overflow',
  'Chief Engineer (Sewerage Operation)',
  'Chief Engineer (Sewerage Operation)',
  'Chief Engineer (Sewerage Operation)',
  'Engineering Hub Building, First Floor, Dr. E. Moses Road, Worli, '
    || 'Mumbai 400018',
  '24958001',
  'che.so@mcgm.gov.in',
  'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
    || 'Chief%20Engineer%20(Sewerage%20Operations)/RTI%20Manuals/'
    || 'ChEngineer_S0_RTI_E02.pdf',
  '2026-10-16T00:00:00Z',
  'archive/bmc_portal_rti_manuals/2026/10/16/'
    || '1a53fd9f3829039a00bd3af9fefebf83f306227085ffdf330765a50aee0bb9dc.pdf',
  'Manual 1 of the department''s Section 4 handbook: operation and maintenance '
    || 'of the sewerage network, pumping stations and treatment plants across '
    || 'seven sewerage zones, and — against the Executive Engineer (Mechanical) '
    || 'for main sewers — "attending day-to-day complaints related to sewer '
    || 'lines". The department works in three divisions and this row names the '
    || 'head of department, because the Eastern Suburbs divisional office is '
    || 'not in the English directory and two out of three would leave six wards '
    || 'guessing. The ward office has a part in this: it raises proposals for '
    || 'repairs to pipe sewer lines and street connections, which this '
    || 'department verifies and sanctions.',
  '2026-10-16'
FROM (VALUES
  ('A'), ('B'), ('C'), ('D'), ('E'), ('F/S'), ('F/N'), ('G/S'), ('G/N'),
  ('H/E'), ('H/W'), ('K/E'), ('K/W'), ('P/S'), ('P/N'), ('R/S'), ('R/C'),
  ('R/N'), ('L'), ('M/E'), ('M/W'), ('N'), ('S'), ('T')
) AS w(ward);

DO $$
DECLARE n integer;
BEGIN
  SELECT count(*) INTO n FROM authority_departments
   WHERE authority = 'BMC' AND category = 'water_drainage'
     AND subcategory = 'sewage overflow';
  IF n <> 24 THEN
    RAISE EXCEPTION 'sewage overflow seeded % ward(s), expected 24', n;
  END IF;
  SELECT count(*) INTO n FROM authority_departments
   WHERE archive_key IS NULL OR archive_key = '';
  IF n <> 0 THEN
    RAISE EXCEPTION '% row(s) cite no archived document', n;
  END IF;
END $$;
