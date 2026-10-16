-- 0033: the street furniture desk — a dead streetlight now has somewhere to go.
--
-- `street_furniture` has been in the taxonomy since v1 and has never routed.
-- internal/classify/taxonomy.go states the rule it was breaking in its own
-- comment: "A category exists here only if something downstream can route it.
-- Adding one without a department mapping produces a report that is classified
-- and then goes nowhere." A citizen could photograph a streetlight that is
-- out, have it classified correctly, and be told nothing.
--
-- The reason it sat unrouted was a real doubt, not an oversight. Mumbai's
-- distribution licensees — Adani Electricity in the suburbs, Tata Power, BEST
-- in the island city — each publish a street light contact, and the CONNECT
-- 2026 ward directory prints Adani's by name. That reads like the licensee
-- owning the lamp. Two documents settle it the other way:
--
--   1. The ward Assistant Engineer (Maintenance & Repair) handbooks under
--      s.4(1)(b) of the RTI Act list, in the duties of the post, item 38:
--      "To provide street lighting in area as per LUX value (illumination) of
--      Reliance energy and TATA power". The licensee sets the illumination
--      standard the ward works to. The duty is the ward's. The same numbered
--      item, worded identically, appears in A Ward (island city) and P/North
--      (western suburbs, Adani's licence area) — one city ward and one suburban
--      ward agreeing is what makes it safe to read as the standard duty list
--      rather than one ward's local arrangement.
--
--   2. BMC's own English Diary 2026 counts street lights as municipal assets,
--      division-wise: City 42,421, Western Suburb 87,347, Eastern Suburb
--      11,377, total 1,41,115 — and counts them under the Roads & Traffic
--      statistics, not under a licensee's.
--
-- The negative evidence matters as much. The Chief Engineer (Mechanical &
-- Electrical) Section 4 handbook — manual 2, the department's own list of
-- twenty-two functions — does **not** claim street lighting. It claims
-- "प्रखर झोताच्या उंच दिव्यांची तरतूद व देखभाल" (provision and maintenance of
-- high-mast lamps) and the illumination of gardens and playgrounds, and
-- nothing else that lights a street. A citywide M&E desk was the obvious
-- guess and would have been wrong for the ordinary lamp post.
--
-- So the desk is the ward's own maintenance office: the same post that already
-- holds minor roads in migration 0029, and whose handbook states its mission as
-- "To Repair & Maintain Minor Roads, Footpath & Municipal Properties" — which
-- is most of what this category contains.
--
-- These rows are derived from the road_defect maintenance rows rather than
-- retyped, so the office, address, telephone and appellate authority cannot
-- drift between two desks that are the same desk. The office details are the
-- consolidated PIO list's (migrations 0015 and 0029, archived); the duty is
-- the handbooks'. source_ref carries the handbook, because the claim this row
-- makes is the duty — the address is only how to reach it. archive_key is
-- deliberately null: the handbooks are read from the live portal and are not
-- in R2 yet, and pointing at the PIO list's archive key would cite a document
-- that says nothing about street lighting.
--
-- One honest limit, recorded rather than hidden. `street_furniture` also holds
-- "damaged signage", and road signage may belong to the Chief Engineer (Roads
-- & Traffic) rather than the ward. Routing is keyed by category, so one desk
-- answers for all five subcategories. Streetlights, footpaths and railings are
-- squarely this post's; signage is the one that may need splitting when a
-- Traffic department handbook settles it.

INSERT INTO authority_departments
  (id, authority, ward, category, department, officer_designation,
   pio_designation, pio_address, office_phone, escalates_to,
   source_ref, retrieved_at, archive_key, note, effective_from)
SELECT
  gen_random_uuid(),
  d.authority,
  d.ward,
  'street_furniture',
  d.department,
  d.officer_designation,
  d.pio_designation,
  d.pio_address,
  d.office_phone,
  -- Every ward but R/C already carries its appellate authority from the PIO
  -- list. R/C's road rows came from migration 0015, which did not record it;
  -- the same list's R/Central maintenance block names कार्यकारी अभियंता,
  -- आर/मध्य विभाग — Executive Engineer, R/Central Ward.
  COALESCE(d.escalates_to, 'Executive Engineer, R/Central Ward'),
  'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
    || 'Wards/Assistant%20Commissioner%20(A-Ward)/RTI%20Manuals/'
    || 'RTI%2017%20Manual%20AE%20MAINT%20A%20Ward.pdf',
  '2026-10-16T00:00:00Z',
  NULL,
  'Street lighting is duty 38 of the ward Assistant Engineer (Maintenance & '
    || 'Repair) under RTI s.4(1)(b): "To provide street lighting in area as per '
    || 'LUX value (illumination) of Reliance energy and TATA power" — the '
    || 'licensee sets the standard, the ward does the work. The same item '
    || 'appears in the P/North handbook (https://portal.mcgm.gov.in/irj/go/km/'
    || 'docs/documents/MCGM%20Department%20List/Wards/Assistant%20Commissioner'
    || '%20(P-North)/RTI%20Manuals/PDFMannualAEMAINTPnorthyear201314.pdf). The '
    || 'Chief Engineer (M&E) handbook claims high-mast lamps and garden '
    || 'illumination only, not street lighting. Office, address and appellate '
    || 'authority are this ward''s maintenance desk ([migration 0029]).',
  '2026-10-16'
FROM authority_departments d
WHERE d.authority = 'BMC'
  AND d.category = 'road_defect'
  AND d.officer_designation LIKE 'Assistant Engineer (Maintenance & Repair)%';

-- A ward with no maintenance desk would silently get no street furniture desk
-- either, and the failure would show up as a citizen with nowhere to send a
-- photograph rather than as a failed migration. Twenty-four or nothing.
DO $$
DECLARE n integer;
BEGIN
  SELECT count(*) INTO n
    FROM authority_departments
   WHERE authority = 'BMC' AND category = 'street_furniture';
  IF n <> 24 THEN
    RAISE EXCEPTION 'street furniture seeded % ward(s), expected 24', n;
  END IF;
END $$;
