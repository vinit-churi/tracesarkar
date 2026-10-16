-- 0038: the first statutory service deadlines, and the fee to fell a tree.
--
-- Every clock the platform has held until now comes from a court order, a
-- tender clause or a departmental charter. These come from a **statute**: the
-- Garden Department's notification of its services under the Maharashtra Right
-- to Public Services Act 2015, issued under circular MOM/1361 dated
-- 29.09.2016. Under that Act a notified service has a fixed time limit, a
-- designated officer, and two levels of appeal, all named in the notification
-- itself.
--
--   Service                                 Limit     Designated officer
--   Letting a ground on hire                 7 days   Divisional Assistant
--                                                     Superintendent of Gardens
--   वृक्षांची छाटणी — pruning a tree           7 days   Divisional Assistant
--                                                     Superintendent of Gardens
--   वृक्षांची कापणी / पुनर्रोपण —                45 days   Zonal Deputy
--   felling or transplanting                          Superintendent of Gardens
--
-- Fees for the third, per tree: a processing fee of Rs. 6,000 and a deposit of
-- Rs. 6,000 to fell, or a deposit of Rs. 10,000 to transplant. Recorded in
-- paise as integers, never as rupees and never as a float.
--
-- **What these are not.** Forty-five days is the time the Tree Authority has
-- to decide an application to fell a tree. It is not the time it has to answer
-- a citizen who photographs a tree being felled without permission, and
-- attaching it to the `tree felling` subcategory would tell a resident that
-- the corporation has forty-five days to act on their report. It does not say
-- that. So none of these is wired into `deadlineKeys` or
-- `subcategoryDeadlineKeys`, and a live test asserts that a report of illegal
-- felling still carries no clock.
--
-- What they are for is the sentence the platform can honestly put beside such
-- a report: felling lawfully requires the Tree Authority's permission, that
-- permission costs a published fee per tree, and the Authority is bound by
-- statute to decide within forty-five days. That is a fact about the process,
-- not an accusation about anybody, and it is the kind of join this platform
-- exists to make. No surface renders it yet — `store.LegalConstant` is how it
-- will, and is why these rows are reachable from code rather than inert.
--
-- The 7-day ground-hire limit is not seeded: nothing in the taxonomy is about
-- hiring a ground, and a constant nothing can ever show is a row that drifts
-- (the reasoning migration 0034 used to leave open defecation out).

INSERT INTO legal_constants
  (id, key, value, citation, citation_url, verification, short_citation,
   effective_from)
VALUES
  (gen_random_uuid(), 'tree.pruning_decision_days', '7'::jsonb,
   'BMC Garden Department, services notified under the Maharashtra Right to '
     || 'Public Services Act 2015 (circular MOM/1361 dated 29.09.2016), item 2, '
     || '"वृक्षांची छाटणी करणे" (pruning of trees): 7 दिवस. Designated officer the '
     || 'Divisional Assistant Superintendent of Gardens; first appeal the Zonal '
     || 'Deputy Superintendent of Gardens; second appeal the Superintendent of '
     || 'Gardens or the Divisional Assistant Commissioner.',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
     || 'Gardens%20and%20Tree%20Authority/Docs/'
     || '%E0%A4%B2%E0%A5%8B%E0%A4%95%E0%A4%B8%E0%A5%87%E0%A4%B5%E0%A4%BE%20'
     || '%E0%A4%B9%E0%A4%95%E0%A5%8D%E0%A4%95%20%E0%A4%85%E0%A4%A7%E0%A4%BF'
     || '%E0%A4%A8%E0%A4%BF%E0%A4%AF%E0%A4%AE%20%E0%A5%A8%E0%A5%A6%E0%A5%A7'
     || '%E0%A5%AB%20%E0%A4%85%E0%A4%82%E0%A4%A4%E0%A4%B0%E0%A5%8D%E0%A4%97'
     || '%E0%A4%A4%20%E0%A4%B8%E0%A5%87%E0%A4%B5%E0%A4%BE.pdf',
   'verified', 'MRTPS Act 2015 notification, Gardens, item 2', '2016-09-29'),

  (gen_random_uuid(), 'tree.felling_permission_days', '45'::jsonb,
   'BMC Garden Department, services notified under the Maharashtra Right to '
     || 'Public Services Act 2015 (circular MOM/1361 dated 29.09.2016), item 3, '
     || '"वृक्षांची कापणी / वृक्ष पुनर्रोपित करणे" (felling or transplanting a tree): '
     || '45 दिवस. This is the time to decide an application for permission, not '
     || 'the time to answer a report of felling without it. Designated officer '
     || 'the Zonal Deputy Superintendent of Gardens; first appeal the '
     || 'Superintendent of Gardens; second appeal the Deputy Municipal '
     || 'Commissioner.',
   NULL, 'verified', 'MRTPS Act 2015 notification, Gardens, item 3', '2016-09-29'),

  (gen_random_uuid(), 'tree.felling_processing_fee_paise', '600000'::jsonb,
   'BMC Garden Department, services notified under the Maharashtra Right to '
     || 'Public Services Act 2015 (circular MOM/1361 dated 29.09.2016), item 3: '
     || 'प्रक्रिया शुल्क रु. ६०००/- प्रती वृक्ष — a processing fee of Rs. 6,000 per tree.',
   NULL, 'verified', 'MRTPS Act 2015 notification, Gardens, item 3', '2016-09-29'),

  (gen_random_uuid(), 'tree.felling_deposit_paise', '600000'::jsonb,
   'BMC Garden Department, services notified under the Maharashtra Right to '
     || 'Public Services Act 2015 (circular MOM/1361 dated 29.09.2016), item 3: '
     || 'अनामत रक्कम रु. ६०००/- प्रती वृक्ष — a deposit of Rs. 6,000 per tree to fell.',
   NULL, 'verified', 'MRTPS Act 2015 notification, Gardens, item 3', '2016-09-29'),

  (gen_random_uuid(), 'tree.transplant_deposit_paise', '1000000'::jsonb,
   'BMC Garden Department, services notified under the Maharashtra Right to '
     || 'Public Services Act 2015 (circular MOM/1361 dated 29.09.2016), item 3: '
     || 'अनामत रक्कम रु. १००००/- प्रती वृक्ष — a deposit of Rs. 10,000 per tree to '
     || 'transplant.',
   NULL, 'verified', 'MRTPS Act 2015 notification, Gardens, item 3', '2016-09-29');

-- The garden desks escalate twice, and the two ladders are not the same one.
-- Migration 0037 recorded the RTI appellate authority from Manual 16 — the
-- ward's Assistant Commissioner, which is who an information request is
-- appealed to. The Right to Public Services notification gives a different
-- ladder for the service itself, and a resident chasing a pruning request
-- needs that one. Both are stated, in the order a person would use them.
UPDATE authority_departments
   SET escalates_to = escalates_to
         || ' (appeal under the RTI Act); for the service itself, the Zonal '
         || 'Deputy Superintendent of Gardens, then the Superintendent of '
         || 'Gardens (Right to Public Services Act 2015 notification)',
       note = note
         || ' Pruning is a notified service with a statutory limit of 7 days '
         || 'and felling permission one of 45 days; neither is a deadline for '
         || 'answering a report of illegal felling (migration 0038).'
 WHERE authority = 'BMC' AND category = 'environment'
   AND subcategory = 'tree felling'
   AND escalates_to NOT LIKE '%Right to Public Services%';

-- R/Central's two road rows came from migration 0015 and never carried an
-- appellate authority, while the other twenty-three wards got theirs in 0029.
-- The same consolidated PIO list has it: the R/मध्य block's maintenance row
-- names कार्यकारी अभियंता, आर/मध्य विभाग — Executive Engineer, R/Central Ward.
UPDATE authority_departments
   SET escalates_to = 'Executive Engineer, R/Central Ward',
       note = COALESCE(note || ' ', '')
         || 'First Appellate Authority added 16.10.2026 from the same '
         || 'consolidated PIO list the row cites.'
 WHERE authority = 'BMC' AND ward = 'R/C' AND category = 'road_defect'
   AND officer_designation LIKE 'Assistant Engineer (Maintenance & Repair)%'
   AND escalates_to IS NULL;

DO $$
DECLARE n integer;
BEGIN
  SELECT count(*) INTO n FROM legal_constants WHERE key LIKE 'tree.%';
  IF n <> 5 THEN
    RAISE EXCEPTION 'expected 5 tree constants, found %', n;
  END IF;
  SELECT count(*) INTO n FROM authority_departments
   WHERE authority = 'BMC' AND category = 'environment'
     AND subcategory = 'tree felling'
     AND escalates_to NOT LIKE '%Right to Public Services%';
  IF n <> 0 THEN
    RAISE EXCEPTION '% garden desk(s) still name only the RTI ladder', n;
  END IF;
END $$;
