-- 0034: BMC's own waste clocks, from the Citizen Charter it publishes.
--
-- The two waste constants the platform has shown since migration 0019 come
-- from the R/Central Solid Waste Management RTI handbook: 24 hours, citing a
-- circular of 16 March 2006. They were the best available then. BMC's Solid
-- Waste Management **Citizen Charter of 7 December 2020** is better on both
-- counts a constant is judged by — it is citywide rather than one ward's, and
-- it is fourteen years newer — and it says something different.
--
-- Row 3 of the charter, "अंतिम निर्णयासाठी लागणारा कालावधी" (time required for
-- final decision), against thirteen numbered complaint types:
--
--   I, II, III, IV, V, VI, VII, VIII, IX, XII, XIII — 12 तास   (12 hours)
--   XI  मृत प्राणी            — 48 तास   (48 hours)
--   X   उघड्या जागेवर शौच करणे  — 1 आठवडा  (1 week)
--
-- The eleven at twelve hours include the ones this platform's taxonomy holds:
-- I no sweeping, II dumping rubbish, III a bin not kept clean, IV the refuse
-- vehicle not arriving, V burning waste in the open, XII debris on a public
-- road and XIII household repair debris. So the category clock halves, from
-- 24 hours to 12.
--
-- XI is the reason this is not simply an edit. A dead animal gets 48 hours —
-- four times the rest of the category — and the platform's taxonomy has "dead
-- animal" as a subcategory of `waste`. Showing a resident 12 hours beside a
-- photograph of a dead dog would state an obligation BMC has not taken on,
-- which is the overstatement D138 removed in the other direction. It gets its
-- own constant, and `subcategoryDeadlineKeys` binds it to that subcategory
-- alone.
--
-- X, open defecation at one week, is not seeded: nothing in the taxonomy
-- reaches it, and a constant nothing can show is a row that will drift.
--
-- The old rows are left in place. They were true of their document and the
-- table is keyed by effective date, so the charter wins on every read while
-- the history of what the platform told people stays intact.
--
-- Two further facts from the same page, recorded here because they are the
-- answer to "and if they do not": complaints I–XII must be dealt with free of
-- charge, while XIII — lifting a household's own repair debris — is charged at
-- the ward contract's per-tonne rate plus 15% supervision. The statutory basis
-- is MMC Act 1888 sections 365 to 385, and the charter names two circulars,
-- MOM/1097 of 25.09.2014 and MOM/586 of 24.06.2016.

INSERT INTO legal_constants
  (id, key, value, citation, citation_url, verification, short_citation,
   effective_from)
VALUES
  (gen_random_uuid(), 'waste.refuse_removal_hours', '12'::jsonb,
   'BMC Solid Waste Management Citizen Charter dated 07.12.2020, row 3 '
     || '"अंतिम निर्णयासाठी लागणारा कालावधी": complaint types I–IX, XII and XIII '
     || '— 12 तास (12 hours). The types are, among others, roads not swept, '
     || 'rubbish dumped, a bin not kept clean, the refuse vehicle not arriving, '
     || 'waste burnt in the open, and debris on a public road. Statutory basis '
     || 'MMC Act 1888 ss.365–385; circulars MOM/1097 dated 25.09.2014 and '
     || 'MOM/586 dated 24.06.2016. Supersedes the 24 hours of the 2006 '
     || 'R/Central handbook, being citywide and later.',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
     || 'Citizen%20Charter/Citizen%20Charter-Solid%20Waste%20Management/'
     || 'Citizen%20Charter-SWM%2007.12.2020.pdf',
   'verified', 'BMC SWM Citizen Charter, 07.12.2020', '2020-12-07'),

  (gen_random_uuid(), 'waste.dead_animal_removal_hours', '48'::jsonb,
   'BMC Solid Waste Management Citizen Charter dated 07.12.2020, row 3 '
     || '"अंतिम निर्णयासाठी लागणारा कालावधी": complaint type XI, मृत प्राणी (a dead '
     || 'animal) — 48 तास (48 hours). Four times the twelve hours the rest of '
     || 'the category carries, so it is bound to this subcategory alone. '
     || 'Statutory basis MMC Act 1888 ss.365–385.',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
     || 'Citizen%20Charter/Citizen%20Charter-Solid%20Waste%20Management/'
     || 'Citizen%20Charter-SWM%2007.12.2020.pdf',
   'verified', 'BMC SWM Citizen Charter, 07.12.2020', '2020-12-07');

-- A constant with no short citation is dropped whole from a post (0028), and
-- these two now appear in one. Both carry theirs above; this guards against a
-- later edit quietly removing one.
DO $$
DECLARE n integer;
BEGIN
  SELECT count(*) INTO n FROM legal_constants
   WHERE effective_from = DATE '2020-12-07'
     AND (short_citation IS NULL OR citation_url IS NULL);
  IF n > 0 THEN
    RAISE EXCEPTION 'charter constant without a short citation or a URL';
  END IF;
END $$;
