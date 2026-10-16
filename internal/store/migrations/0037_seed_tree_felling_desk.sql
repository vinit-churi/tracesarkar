-- 0037: the Tree Authority, and the last category that routed nowhere.
--
-- `environment` holds six things and BMC owns one of them. A mangrove is the
-- Forest Department's and the coastal zone authority's; an effluent discharge
-- is the Maharashtra Pollution Control Board's; an air pollution source is
-- shared between the board and the corporation on a basis no document here
-- settles. Seeding the category to the Garden Department would have misrouted
-- five subcategories out of six — the error migration 0035 exists to prevent.
--
-- So one row per ward, scoped to `tree felling`, and the category itself stays
-- empty. A citizen photographing a mangrove being filled still gets the honest
-- answer, which is that this platform does not yet know where that goes.
--
-- The claim is the Garden Department's own. Manual 1 of its Section 4(1)(b)
-- handbook states the mission of the department in three lines — "To Protect
-- Open spaces, To Control illegal tree cutting, Protection and Preservation of
-- trees" — and lists under TREE AUTHORITY: planting new trees; removal of dead
-- and dangerous trees and branches; pruning; **taking action against illegal
-- tree cutting**; inspection of trees in public and private premises as per
-- complaints received; supervision of tree cutting carried out under
-- permission. Manual 4 gives the statute: the Mumbai Municipal Corporation Act
-- 1888 and the **Maharashtra (Urban Areas) Protection and Preservation of
-- Trees Act, 1975**, as amended to 1 January 2016.
--
-- Manual 16 names an Assistant Superintendent of Gardens for each of the
-- twenty-four wards, sitting in the ward and answering to that ward's
-- Assistant Commissioner as First Appellate Authority — the same post
-- migration 0031 seeded as the ward rung of the escalation ladder, which is
-- why the two agree.
--
-- Three deliberate gaps in the mailbox column, all of them the document's:
--
--   * Eight wards — C, D, E, F/South, F/North, H/East, S and R/Central — have
--     no e-mail printed against the post at all. Sixteen addresses exist in
--     the whole manual; these are not among them.
--   * B Ward's row prints `asg01a.garden@mcgm.gov.in`, which is A Ward's. It
--     is in B's own cell in the scan, so it is BMC's text rather than a
--     parsing artefact, and it is recorded here in the note rather than in the
--     field: a tree complaint for Dongri arriving at the Fort officer is a
--     misroute this platform would have caused.
--   * No office address is stored. The manual prints one per ward, and the
--     two-column scan drops characters when it is read as text — "Shrikant
--     Paleka", "Room No-407, Aazad road" against a ward in the wrong
--     direction. A truncated address is a wrong address, and the role mailbox
--     is the channel that actually works.
--
-- Note also what the handbook does not give: Manual 4, "norms set by it for
-- discharge of its function", names two statutes and no time limit. That is
-- the fifth BMC department in a row to publish none.

INSERT INTO authority_departments
  (id, authority, ward, category, subcategory, department, officer_designation,
   pio_designation, office_email, escalates_to,
   source_ref, retrieved_at, archive_key, note, effective_from)
SELECT
  gen_random_uuid(), 'BMC', g.ward, 'environment', 'tree felling',
  'Garden Department (Tree Authority) — ' || g.ward || ' Ward',
  'Assistant Superintendent of Gardens, ' || g.ward || ' Ward',
  'Assistant Superintendent of Gardens, ' || g.ward || ' Ward',
  g.email,
  'Assistant Commissioner, ' || g.ward || ' Ward',
  'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
    || 'Garden%20Department/Garden_Department_E17.pdf',
  '2026-10-16T00:00:00Z',
  NULL,
  'Tree Authority under the Maharashtra (Urban Areas) Protection and '
    || 'Preservation of Trees Act, 1975 (amended to 01.01.2016) and MMC Act 1888. '
    || 'The department''s Manual 1 gives its mission as controlling illegal tree '
    || 'cutting and lists action against illegal tree cutting, removal of dead '
    || 'and dangerous trees and branches, pruning, and inspection of trees on '
    || 'complaint. First Appellate Authority is this ward''s Assistant '
    || 'Commissioner, per Manual 16.' || g.caveat,
  '2026-10-16'
FROM (VALUES
  ('A',   'asg01a.garden@mcgm.gov.in',       ''),
  ('B',   NULL,                              ' No mailbox is stored: the manual prints '
                                             || 'asg01a.garden@mcgm.gov.in in B Ward''s own row, '
                                             || 'which is A Ward''s address.'),
  ('C',   NULL,                              ' The manual prints no e-mail for this post.'),
  ('D',   NULL,                              ' The manual prints no e-mail for this post.'),
  ('E',   NULL,                              ' The manual prints no e-mail for this post.'),
  ('F/S', NULL,                              ' The manual prints no e-mail for this post.'),
  ('F/N', NULL,                              ' The manual prints no e-mail for this post.'),
  ('G/S', 'asg01gs.garden@mcgm.gov.in',      ''),
  ('G/N', 'asg01gn.garden@mcgm.gov.in',      ''),
  ('H/E', NULL,                              ' The manual prints no e-mail for this post.'),
  ('H/W', 'asg01hw.garden@mcgm.gov.in',      ''),
  ('K/E', 'asg03z3.garden@mcgm.gov.in',      ''),
  ('K/W', 'asg01z4.garden@mcgm.gov.in',      ''),
  ('P/S', 'asg03z4.garden@mcgm.gov.in',      ''),
  ('P/N', 'asg02z4.garden@mcgm.gov.in',      ''),
  ('L',   'asg01l.garden@mcgm.gov.in',       ''),
  ('M/E', 'asg01meast.garden@mcgm.gov.in',   ''),
  ('M/W', 'asg02mw.garden@mcgm.gov.in',      ''),
  ('N',   'asg01n.garden@mcgm.gov.in',       ''),
  ('S',   NULL,                              ' The manual prints no e-mail for this post.'),
  ('T',   'asgz6.garden@mcgm.gov.in',        ''),
  ('R/S', 'asg02z7.garden@mcgm.gov.in',      ''),
  ('R/C', NULL,                              ' The manual prints no e-mail for this post.'),
  ('R/N', 'asg03z7.garden@mcgm.gov.in',      '')
) AS g(ward, email, caveat);

-- The same manual closes two gaps migration 0031 left open. K/East and
-- P/North were seeded with a mailbox and no office address, because the
-- consolidated PIO list gave those two wards two-column pages headed with
-- ward names BMC does not use officially. Manual 16 of the Garden handbook
-- prints the Assistant Commissioner's address for every ward in its own
-- column, and these two were read from the scan rather than from its text
-- layer, so no character is guessed.
UPDATE ward_escalation_contacts
   SET office_address = 'K/East Ward Municipal Office Building, Room No.103, '
                        || 'Aazad Road, Gundavali, Andheri (East), Mumbai 400069',
       source_ref = source_ref
                    || ' Office address added 16.10.2026 from Garden Department '
                    || 'RTI Manual 16 (Garden_Department_E17.pdf), First Appellate '
                    || 'Authority column.'
 WHERE tier = 'ward' AND ward = 'K/E' AND office_address IS NULL;

UPDATE ward_escalation_contacts
   SET office_address = 'P/North Ward Municipal Office Building, Liberty Garden, '
                        || 'Mamletdar Wadi Marg, Malad (West), Mumbai 400064',
       source_ref = source_ref
                    || ' Office address added 16.10.2026 from Garden Department '
                    || 'RTI Manual 16 (Garden_Department_E17.pdf), First Appellate '
                    || 'Authority column.'
 WHERE tier = 'ward' AND ward = 'P/N' AND office_address IS NULL;

DO $$
DECLARE n integer; noaddr integer;
BEGIN
  SELECT count(*) INTO n FROM authority_departments
   WHERE authority = 'BMC' AND category = 'environment'
     AND subcategory = 'tree felling';
  IF n <> 24 THEN
    RAISE EXCEPTION 'tree felling seeded % ward(s), expected 24', n;
  END IF;
  -- The category itself must stay empty: the other five subcategories are not
  -- BMC's, and a category-wide row would quietly claim them.
  SELECT count(*) INTO n FROM authority_departments
   WHERE authority = 'BMC' AND category = 'environment' AND subcategory IS NULL;
  IF n <> 0 THEN
    RAISE EXCEPTION 'a category-wide environment desk was seeded; % row(s)', n;
  END IF;
  SELECT count(*) INTO noaddr FROM ward_escalation_contacts
   WHERE tier = 'ward' AND office_address IS NULL;
  IF noaddr <> 0 THEN
    RAISE EXCEPTION '% ward(s) still have no office address', noaddr;
  END IF;
END $$;
