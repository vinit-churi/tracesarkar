-- 0029: the ward maintenance desk for the other twenty-three wards.
--
-- Until now one ward could produce a complaint addressed to a named office.
-- Boundaries covered all twenty-four, works covered all twenty-four, and the
-- thing that decides where a complaint goes covered one — so a resident
-- standing in Andheri got a resolved ward, a matched contract, and nowhere to
-- send it.
--
-- These rows come from the same document migration 0015 cited for R/C: BMC's
-- consolidated list of Public Information Officers and First Appellate
-- Authorities, updated to 01.01.2025, already archived in R2 on 29 September
-- 2026. Every ward carries the same post with the same remit, worded
-- identically in all twenty-four blocks:
--
--   सहायक अभियंता (परिरक्षण) — "परिरक्षणाबाबतच्या सर्व बाबी उदा. विभागातील रस्ते,
--   पर्जन्य जलवाहिन्या, मलनिसारण, महापालिका इमारती व अन्य स्थावर मालमत्ता,
--   किरकोळ कामे कंत्राटदार, नगरसेवक निधी इ."
--
--   Assistant Engineer (Maintenance) — all maintenance matters, e.g. ward
--   roads, storm-water drains, sewerage, municipal buildings and other
--   immovable property, minor-works contractors, corporator funds.
--
-- R/C is deliberately absent: it already has two researched road_defect rows
-- from migration 0015, and a row dated later would silently win over them.
--
-- Only `road_defect` is seeded. The same remit covers storm-water and
-- sewerage, but `water_drainage` in this platform also holds burst pipes and
-- leakage, which belong to the Water Works desk — a split this document does
-- not settle, and guessing it would send a complaint to an office that cannot
-- act on it.
--
-- Telephone numbers are recorded only where the scan is unambiguous, and
-- e-mail addresses are omitted entirely. Migration 0015 left R/C's road
-- telephones null for the same reason: a number guessed at is rung, nobody
-- answers, and the citizen concludes the platform is wrong about everything
-- else. The office and the designation are what a complaint actually needs.
--
-- Designations only, never names. The source names the individual holding each
-- post; the register's blocklist forbids storing that, and a post outlives
-- whoever holds it.

INSERT INTO authority_departments
  (id, authority, ward, category, department, officer_designation,
   pio_designation, pio_address, office_phone, escalates_to,
   source_ref, retrieved_at, archive_key, note, effective_from)
VALUES
  (gen_random_uuid(), 'BMC', 'A', 'road_defect',
   'A Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), A Ward',
   'Assistant Engineer (Maintenance), A Ward',
   '134 E, Shahid Bhagat Singh Marg, near Reserve Bank of India, Fort, Mumbai 400001',
   '22661353',
   'Executive Engineer, A Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, A Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'B', 'road_defect',
   'B Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), B Ward',
   'Assistant Engineer (Maintenance), B Ward',
   'Ramchandra Bhatt Marg, opposite J.J. Hospital, Mumbai 400009',
   NULL,
   'Executive Engineer, B Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, B Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'C', 'road_defect',
   'C Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), C Ward',
   'Assistant Engineer (Maintenance), C Ward',
   '76, Shrikant Palekar Marg, near Chandanwadi crematorium, Mumbai 400002',
   '22014022',
   'Executive Engineer, C Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, C Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'D', 'road_defect',
   'D Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), D Ward',
   'Assistant Engineer (Maintenance), D Ward',
   'Jobanputra Compound, Nana Chowk, Grant Road (West), Mumbai 400007',
   '23861426',
   'Executive Engineer, D Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, D Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'E', 'road_defect',
   'E Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), E Ward',
   'Assistant Engineer (Maintenance), E Ward',
   '10, Shaikh Hafizuddin Marg, Byculla, Mumbai 400008',
   '23081471',
   'Executive Engineer, E Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, E Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'F/S', 'road_defect',
   'F/S Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), F/S Ward',
   'Assistant Engineer (Maintenance), F/S Ward',
   '2nd Floor, Room No. 19, Dr. B.A. Road, Parel, Mumbai 400012',
   '24134560',
   'Executive Engineer, F/S Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, F/S Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'F/N', 'road_defect',
   'F/N Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), F/N Ward',
   'Assistant Engineer (Maintenance), F/N Ward',
   '96, Bhau Daji Road, Matunga, Mumbai 400019',
   '24024353',
   'Executive Engineer, F/N Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, F/N Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'G/S', 'road_defect',
   'G/S Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), G/S Ward',
   'Assistant Engineer (Maintenance), G/S Ward',
   'N.M. Joshi Marg, Mumbai 400013',
   '24305031',
   'Executive Engineer, G/S Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, G/S Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'G/N', 'road_defect',
   'G/N Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), G/N Ward',
   'Assistant Engineer (Maintenance), G/N Ward',
   'Harishchandra Yewale Marg, behind Plaza Cinema, Dadar, Mumbai 400028',
   '24212776',
   'Executive Engineer, G/N Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, G/N Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'H/E', 'road_defect',
   'H/E Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), H/E Ward',
   'Assistant Engineer (Maintenance), H/E Ward',
   'Prabhat Colony, T.P.S. 5, Santacruz (East), Mumbai 400055',
   '26182217',
   'Executive Engineer, H/E Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, H/E Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'H/W', 'road_defect',
   'H/W Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), H/W Ward',
   'Assistant Engineer (Maintenance), H/W Ward',
   'St Martins Road, Bandra (West), Mumbai 400050',
   '26422311',
   'Executive Engineer, H/W Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, H/W Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'K/E', 'road_defect',
   'K/E Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), K/E Ward',
   'Assistant Engineer (Maintenance), K/E Ward',
   'Azad Road, Gundavali, Andheri (East), Mumbai 400069',
   '26837104',
   'Executive Engineer, K/E Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, K/E Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'K/W', 'road_defect',
   'K/W Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), K/W Ward',
   'Assistant Engineer (Maintenance), K/W Ward',
   'Ward Office Building, Paliram Road, beside Swami Vivekanand Road, opposite Andheri Station, Andheri (West), Mumbai 400058',
   '26239131',
   'Executive Engineer, K/W Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, K/W Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'L', 'road_defect',
   'L Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), L Ward',
   'Assistant Engineer (Maintenance), L Ward',
   'Laxmanrao Yadav Mandai Building, S.G. Barve Marg, Kurla, Mumbai 400070',
   '26505103',
   'Executive Engineer, L Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, L Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'M/E', 'road_defect',
   'M/E Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), M/E Ward',
   'Assistant Engineer (Maintenance), M/E Ward',
   'Municipal Office Building, Late Madhukar Tukaram Kadam Marg near Hirakeri Road junction, Deonar Municipal Colony, Govandi, Mumbai 400043',
   NULL,
   'Executive Engineer, M/E Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, M/E Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'M/W', 'road_defect',
   'M/W Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), M/W Ward',
   'Assistant Engineer (Maintenance), M/W Ward',
   'Sharadbhau Acharya Marg, Chembur, Mumbai 400071',
   '25283900',
   'Executive Engineer, M/W Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, M/W Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'N', 'road_defect',
   'N Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), N Ward',
   'Assistant Engineer (Maintenance), N Ward',
   'Ward Office Building, Jawahar Road, Ghatkopar (East), Mumbai 400077',
   '25094161',
   'Executive Engineer, N Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, N Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'P/S', 'road_defect',
   'P/S Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), P/S Ward',
   'Assistant Engineer (Maintenance), P/S Ward',
   'Municipal Office Building, C.T.S. No. 746, Hillage Pahadi, Swami Vivekanand Marg, Goregaon (West), Mumbai 400104',
   '28721186',
   'Executive Engineer, P/S Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, P/S Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'P/N', 'road_defect',
   'P/N Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), P/N Ward',
   'Assistant Engineer (Maintenance), P/N Ward',
   'Ward Office Building, Liberty Garden, Mamlatdar Wadi, Malad (West), Mumbai 400064',
   '28824245',
   'Executive Engineer, P/N Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, P/N Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'R/S', 'road_defect',
   'R/S Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), R/S Ward',
   'Assistant Engineer (Maintenance), R/S Ward',
   'M.G. Chhed Road No. 2, near S.V.P. swimming pool, Kandivali (West), Mumbai 400067',
   '28054784',
   'Executive Engineer, R/S Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, R/S Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'R/N', 'road_defect',
   'R/N Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), R/N Ward',
   'Assistant Engineer (Maintenance), R/N Ward',
   'Below Sangeetkar Sudhir Phadke flyover, Jayvant Sawant Marg, Dahisar (West), Mumbai 400068',
   '28936000',
   'Executive Engineer, R/N Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, R/N Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'S', 'road_defect',
   'S Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), S Ward',
   'Assistant Engineer (Maintenance), S Ward',
   'L.B. Shastri Marg, Bhandup (West), Mumbai 400078',
   '25947570',
   'Executive Engineer, S Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, S Ward.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'T', 'road_defect',
   'T Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair), T Ward',
   'Assistant Engineer (Maintenance), T Ward',
   'Lala Devidayal Marg, Mulund (West), Mumbai 400080',
   '25645291',
   'Executive Engineer, T Ward',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Ward roads maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025. First Appellate Authority: Executive Engineer, T Ward.',
   '2026-09-29')
ON CONFLICT DO NOTHING;
