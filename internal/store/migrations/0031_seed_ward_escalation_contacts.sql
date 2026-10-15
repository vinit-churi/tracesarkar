-- 0031: how to reach the office a complaint escalates to.
--
-- Migrations 0015, 0029 and 0030 gave every ward a desk and every desk a next
-- rung: `escalates_to` already says "Executive Engineer, A Ward" or "Assistant
-- Commissioner, R/Central Ward". It names the office and gives no way to reach
-- it. A citizen whose 48 hours have run out can read who is next and still has
-- nowhere to send the second letter.
--
-- BMC publishes the contact for that rung, attached to the post and not to the
-- person: a role mailbox per ward (`ac.<ward>@mcgm.gov.in`) and per zone
-- (`dmc.z<n>@mcgm.gov.in`), plus the ward control room and the ward office
-- address. None of it is personal data and all of it survives a transfer.
--
-- Source: CONNECT - 2026, BMC's disaster-management contact directory,
-- section 14 "Ward Wise Information", pages 73-337. Note that the file is
-- published under a 2021 filename; the document itself is titled and dated
-- 2026 on every page, and it is recorded here by its contents.
--
-- Three things this migration deliberately does NOT do.
--
-- 1. No officers' names and no personal mobile numbers. The source lists both
--    for every post. The register's blocklist forbids storing them, several
--    entries already carry 2026 retirement dates, and a post outlives whoever
--    holds it. Only the role mailbox, the office landline and the control room
--    are taken.
--
-- 2. K/E and P/N get no address. Their pages are laid out in two columns and
--    the source labels those wards with names BMC does not use officially —
--    'K/South'/'K/North' and 'P/East'/'P/West' — so the two addresses on each
--    page cannot be attributed to a ward with confidence from the text. The
--    mailbox and control room for both wards are unambiguous and are seeded.
--
-- 3. Two further mailboxes appear in the source and are NOT seeded:
--    `ac.kn@mcgm.gov.in` and `ac.pe@mcgm.gov.in`. They belong to the same
--    alternative ward naming described above. Each ward here carries the
--    mailbox whose local part matches its official code; sending a citizen's
--    escalation to a mailbox nobody reads is worse than offering none.
--
-- `dmcz4@mcgm.gov.in` is recorded without the dot its six siblings carry.
-- That is how the source prints it, and it is left verbatim rather than
-- normalised into an address that may not exist.

CREATE TABLE IF NOT EXISTS ward_escalation_contacts (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  authority     text NOT NULL,
  ward          text,
  tier          text NOT NULL CHECK (tier IN ('ward', 'zone')),
  designation   text NOT NULL,
  office_email  text,
  control_room  text,
  office_address text,
  zone          text NOT NULL,
  source_ref    text NOT NULL,
  retrieved_at  timestamptz NOT NULL,
  effective_from date NOT NULL,
  CONSTRAINT ward_tier_has_ward CHECK (tier <> 'ward' OR ward IS NOT NULL),
  CONSTRAINT zone_tier_has_no_ward CHECK (tier <> 'zone' OR ward IS NULL)
);

COMMENT ON TABLE ward_escalation_contacts IS
  'The office a ward desk escalates to, with a way to reach it. Role mailboxes only, never an individual.';
COMMENT ON COLUMN ward_escalation_contacts.office_email IS
  'Role mailbox exactly as published, including irregular spellings.';
COMMENT ON COLUMN ward_escalation_contacts.zone IS
  'BMC administrative zone. Each ward belongs to exactly one; each zone has one Deputy Municipal Commissioner.';

CREATE UNIQUE INDEX IF NOT EXISTS ward_escalation_ward_uniq
  ON ward_escalation_contacts (authority, ward) WHERE tier = 'ward';
CREATE UNIQUE INDEX IF NOT EXISTS ward_escalation_zone_uniq
  ON ward_escalation_contacts (authority, zone) WHERE tier = 'zone';

INSERT INTO ward_escalation_contacts
  (authority, ward, tier, designation, office_email, control_room,
   office_address, zone, source_ref, retrieved_at, effective_from)
VALUES
  ('BMC', 'A', 'ward', 'Assistant Commissioner, A Ward',
   'ac.a@mcgm.gov.in', '2262 4000, 22607062, 9820574323', '134/E, Shahid Bhagat Singh Marg, Near Reserve Bank of India, Mumbai – 400 001',
   'Zone 1', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'B', 'ward', 'Assistant Commissioner, B Ward',
   'ac.b@mcgm.gov.in', '23794000', '121, Ramchandra Bhat Marg, Babula Tank, Cross Lane, Opp J J Hospital, Mumbai – 400009',
   'Zone 1', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'C', 'ward', 'Assistant Commissioner, C Ward',
   'ac.c@mcgm.gov.in', '2201 4000', '76, Shrikant Palekar Marg, Chandanwadi,Marin Lines, Mumbai - 400 002',
   'Zone 1', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'D', 'ward', 'Assistant Commissioner, D Ward',
   'ac.d@mcgm.gov.in', '2386 4000', 'Jobanputra Compound, Nana Chowk, Grant Road (W), Mumbai - 400 007',
   'Zone 1', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'E', 'ward', 'Assistant Commissioner, E Ward',
   'ac.e@mcgm.gov.in', '23014000 , 9930762131', 'New Mhada Mill Compound, Rambhau Bhogale Marg, Ghodapdeo, Byculla (E), Mumbai - 400 008',
   'Zone 1', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'F/S', 'ward', 'Assistant Commissioner, F/S Ward',
   'ac.fs@mcgm.gov.in', '2410 3000, 24166093/3200/27009321290855', 'Junction of Dr Ambedkar Marg and Jagannath Bhatankar Marg Junction, Parel T T, Mumbai – 400012',
   'Zone 2', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'F/N', 'ward', 'Assistant Commissioner, F/N Ward',
   'ac.fn@mcgm.gov.in', '2408 4000', 'Plot No 96, Bhau Daji Road, King’s Circle, Matunga (E), Mumbai – 400019',
   'Zone 2', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'G/S', 'ward', 'Assistant Commissioner, G/S Ward',
   'ac.gs@mcgm.gov.in', '65193111', 'Dhanmill Naka, N M Joshi Marg, Mumbai - 400 013',
   'Zone 2', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'G/N', 'ward', 'Assistant Commissioner, G/N Ward',
   'ac.gn@mcgm.gov.in', '69517888', 'Harishchandra Yelve Marg, Behind Plaza Cinema, Dadar (W), Mumbai - 400 028',
   'Zone 2', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'H/E', 'ward', 'Assistant Commissioner, H/E Ward',
   'ac.heast@mcgm.gov.in', '2611 4000', 'Plot No 137, TPS IV, Road No 2, Prabhat Colony, Santacruz (E), Mumbai - 400 055',
   'Zone 3', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'H/W', 'ward', 'Assistant Commissioner, H/W Ward',
   'ac.hw@mcgm.gov.in', '2644 4000', '2nd Hasnabad Lane, Near Khar Education Society, Khar (W) Mumbai- 400052',
   'Zone 3', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'K/E', 'ward', 'Assistant Commissioner, K/E Ward',
   'ac.ke@mcgm.gov.in', '26847000', NULL,
   'Zone 3', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'K/W', 'ward', 'Assistant Commissioner, K/W Ward',
   'ac.kw@mcgm.gov.in', '2623 4000, 8928443696', 'Paliram Road, Off S V Road, Opp Andheri Station, Andheri W, Mumbai - 400 058',
   'Zone 4', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'P/S', 'ward', 'Assistant Commissioner, P/S Ward',
   'ac.ps@mcgm.gov.in', '28727000', 'CTS No 746, Village Pahadi, S V Road, Goregaon (W), Mumbai – 400104',
   'Zone 4', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'P/N', 'ward', 'Assistant Commissioner, P/N Ward',
   'ac.pn@mcgm.gov.in', '9326564842', NULL,
   'Zone 4', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'L', 'ward', 'Assistant Commissioner, L Ward',
   'ac.l@mcgm.gov.in', '26505109, 8652750405', 'Laxmanrao Yadav Market Building, S G Barve Marg, Kurla West, Mumbai - 400070',
   'Zone 5', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'M/E', 'ward', 'Assistant Commissioner, M/E Ward',
   'ac.me@mcgm.gov.in', '25558789', 'Madhukar Tukaram Marg & Peryphery Road Junction, Deonar Colony, Govandi, Mumbai – 400043',
   'Zone 5', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'M/W', 'ward', 'Assistant Commissioner, M/W Ward',
   'ac.mw@mcgm.gov.in', '25284000', 'Sharadbhau Acharya Marg, Near Natraj Cinema, Chembur, Mumbai - 400 071',
   'Zone 5', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'N', 'ward', 'Assistant Commissioner, N Ward',
   'ac.n@mcgm.gov.in', '2501 3000', 'Jawahar Road, Ghatkopar (E), Mumbai – 400 077',
   'Zone 6', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'S', 'ward', 'Assistant Commissioner, S Ward',
   'ac.s@mcgm.gov.in', '69365002', 'Lal Bahadur Shastri Marg, Mangatram Petrol Pump, Bhandup (W), Mumbai – 400 078',
   'Zone 6', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'T', 'ward', 'Assistant Commissioner, T Ward',
   'ac.t@mcgm.gov.in', '25694000', 'Lala Devidayal Marg, Near Panch Rasta, Mulund (W), Mumbai – 400080',
   'Zone 6', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'R/S', 'ward', 'Assistant Commissioner, R/S Ward',
   'ac.rs@mcgm.gov.in', '28054788', 'Mahatma Gandhi Cross Road No 2, Near Sardar Vallabbhai Patel Swimming Pool, Kandivali (W), Mumbai- 400067',
   'Zone 7', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'R/C', 'ward', 'Assistant Commissioner, R/C Ward',
   'ac.rc@mcgm.gov.in', '2893 1188', 'Chandavarkar Lane, Near Railway Station, Borivali West, Mumbai- 400 092',
   'Zone 7', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', 'R/N', 'ward', 'Assistant Commissioner, R/N Ward',
   'ac.rn@mcgm.gov.in', '28936000', 'CTS No 921/7, Dahisar Gaon, Rustumji Colony, Ranganath Keskar marg, Dahisar West, Mumbai-400 068',
   'Zone 7', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', NULL, 'zone', 'Deputy Municipal Commissioner, Zone 1',
   'dmc.z1@mcgm.gov.in', NULL, NULL,
   'Zone 1', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', NULL, 'zone', 'Deputy Municipal Commissioner, Zone 2',
   'dmc.z2@mcgm.gov.in', NULL, NULL,
   'Zone 2', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', NULL, 'zone', 'Deputy Municipal Commissioner, Zone 3',
   'dmc.z3@mcgm.gov.in', NULL, NULL,
   'Zone 3', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', NULL, 'zone', 'Deputy Municipal Commissioner, Zone 4',
   'dmcz4@mcgm.gov.in', NULL, NULL,
   'Zone 4', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', NULL, 'zone', 'Deputy Municipal Commissioner, Zone 5',
   'dmc.z5@mcgm.gov.in', NULL, NULL,
   'Zone 5', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', NULL, 'zone', 'Deputy Municipal Commissioner, Zone 6',
   'dmc.z6@mcgm.gov.in', NULL, NULL,
   'Zone 6', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15'),
  ('BMC', NULL, 'zone', 'Deputy Municipal Commissioner, Zone 7',
   'dmc.z7@mcgm.gov.in', NULL, NULL,
   'Zone 7', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBMCDairyBooks', '2026-10-15T00:00:00Z', '2026-10-15')
ON CONFLICT DO NOTHING;
