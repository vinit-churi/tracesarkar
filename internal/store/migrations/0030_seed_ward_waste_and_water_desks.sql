-- 0030: the waste and water desks for the other twenty-three wards.
--
-- From the same document as migration 0029, which was already in the archive:
-- BMC's consolidated PIO and First Appellate Authority list. Every ward block
-- carries three engineering posts, not one —
--
--   सहायक अभियंता (परिरक्षण)   Assistant Engineer (Maintenance)   — seeded in 0029
--   सहायक अभियंता (घ.क.व्य.)    Assistant Engineer (SWM)
--   सहायक अभियंता (जलकामे)     Assistant Engineer (Water Works)   — पाणीपुरवठा, water supply
--
-- What makes this safe to extend to twenty-three wards is that R/C already
-- has both desks, read first-hand from the Solid Waste Management and Water
-- Works handbooks on 3 October (migration 0019), and the designations in those
-- handbooks are the same two this list gives. Two independent documents
-- agreeing on the same post is the evidence; one document generalised on its
-- own would not be.
--
-- One honest limit. This list's कार्यक्षेत्र column describes the information a
-- PIO holds rather than purely what an office operates — for SWM it reads
-- "technical matters and tender information of the ward's SWM department".
-- The handbooks are the stronger evidence of operational responsibility, and
-- they cover R/C only. These rows follow R/C's answer because the post is the
-- same; a ward where that turns out to be wrong is a correction, not a guess
-- left unrecorded.
--
-- `escalates_to` is null throughout. In the maintenance rows the appellate
-- authority sits on the same line and is unambiguous; against the waste and
-- water rows the cell is merged across a block and cannot be read off without
-- guessing which rows it covers. A wrong appellate authority sends an appeal
-- to an office that returns it.
--
-- water_drainage inherits R/C's split, and it is a real one: this desk is
-- water *supply*, so burst pipes and leaks belong to it while waterlogging,
-- blocked drains and sewage overflow belong to the maintenance desk in 0029.
-- The platform holds all of them in one category. Resolving that needs the
-- ward drainage handbooks, not this list.

INSERT INTO authority_departments
  (id, authority, ward, category, department, officer_designation,
   pio_designation, pio_address, office_phone,
   source_ref, retrieved_at, archive_key, note, effective_from)
VALUES
  (gen_random_uuid(), 'BMC', 'A', 'waste',
   'Solid Waste Management — A Ward',
   'Assistant Engineer (SWM), A Ward',
   'Assistant Engineer (SWM), A Ward',
   '134 E, Shahid Bhagat Singh Marg, near Reserve Bank of India, Fort, Mumbai 400001',
   '22661353',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'A', 'water_drainage',
   'Water Works (Hydraulic Engineer) — A Ward',
   'Assistant Engineer (Water Works), A Ward',
   'Assistant Engineer (Water Works), A Ward',
   '134 E, Shahid Bhagat Singh Marg, near Reserve Bank of India, Fort, Mumbai 400001',
   '22661353',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'B', 'waste',
   'Solid Waste Management — B Ward',
   'Assistant Engineer (SWM), B Ward',
   'Assistant Engineer (SWM), B Ward',
   'Ramchandra Bhatt Marg, opposite J.J. Hospital, Mumbai 400009',
   NULL,
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'B', 'water_drainage',
   'Water Works (Hydraulic Engineer) — B Ward',
   'Assistant Engineer (Water Works), B Ward',
   'Assistant Engineer (Water Works), B Ward',
   'Ramchandra Bhatt Marg, opposite J.J. Hospital, Mumbai 400009',
   NULL,
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'C', 'waste',
   'Solid Waste Management — C Ward',
   'Assistant Engineer (SWM), C Ward',
   'Assistant Engineer (SWM), C Ward',
   '76, Shrikant Palekar Marg, near Chandanwadi crematorium, Mumbai 400002',
   '22014022',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'C', 'water_drainage',
   'Water Works (Hydraulic Engineer) — C Ward',
   'Assistant Engineer (Water Works), C Ward',
   'Assistant Engineer (Water Works), C Ward',
   '76, Shrikant Palekar Marg, near Chandanwadi crematorium, Mumbai 400002',
   '22014022',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'D', 'waste',
   'Solid Waste Management — D Ward',
   'Assistant Engineer (SWM), D Ward',
   'Assistant Engineer (SWM), D Ward',
   'Jobanputra Compound, Nana Chowk, Grant Road (West), Mumbai 400007',
   '23861426',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'D', 'water_drainage',
   'Water Works (Hydraulic Engineer) — D Ward',
   'Assistant Engineer (Water Works), D Ward',
   'Assistant Engineer (Water Works), D Ward',
   'Jobanputra Compound, Nana Chowk, Grant Road (West), Mumbai 400007',
   '23861426',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'E', 'waste',
   'Solid Waste Management — E Ward',
   'Assistant Engineer (SWM), E Ward',
   'Assistant Engineer (SWM), E Ward',
   '10, Shaikh Hafizuddin Marg, Byculla, Mumbai 400008',
   '23081471',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'E', 'water_drainage',
   'Water Works (Hydraulic Engineer) — E Ward',
   'Assistant Engineer (Water Works), E Ward',
   'Assistant Engineer (Water Works), E Ward',
   '10, Shaikh Hafizuddin Marg, Byculla, Mumbai 400008',
   '23081471',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'F/S', 'waste',
   'Solid Waste Management — F/S Ward',
   'Assistant Engineer (SWM), F/S Ward',
   'Assistant Engineer (SWM), F/S Ward',
   '2nd Floor, Room No. 19, Dr. B.A. Road, Parel, Mumbai 400012',
   '24134560',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'F/S', 'water_drainage',
   'Water Works (Hydraulic Engineer) — F/S Ward',
   'Assistant Engineer (Water Works), F/S Ward',
   'Assistant Engineer (Water Works), F/S Ward',
   '2nd Floor, Room No. 19, Dr. B.A. Road, Parel, Mumbai 400012',
   '24134560',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'F/N', 'waste',
   'Solid Waste Management — F/N Ward',
   'Assistant Engineer (SWM), F/N Ward',
   'Assistant Engineer (SWM), F/N Ward',
   '96, Bhau Daji Road, Matunga, Mumbai 400019',
   '24024353',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'F/N', 'water_drainage',
   'Water Works (Hydraulic Engineer) — F/N Ward',
   'Assistant Engineer (Water Works), F/N Ward',
   'Assistant Engineer (Water Works), F/N Ward',
   '96, Bhau Daji Road, Matunga, Mumbai 400019',
   '24024353',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'G/S', 'waste',
   'Solid Waste Management — G/S Ward',
   'Assistant Engineer (SWM), G/S Ward',
   'Assistant Engineer (SWM), G/S Ward',
   'N.M. Joshi Marg, Mumbai 400013',
   '24305031',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'G/S', 'water_drainage',
   'Water Works (Hydraulic Engineer) — G/S Ward',
   'Assistant Engineer (Water Works), G/S Ward',
   'Assistant Engineer (Water Works), G/S Ward',
   'N.M. Joshi Marg, Mumbai 400013',
   '24305031',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'G/N', 'waste',
   'Solid Waste Management — G/N Ward',
   'Assistant Engineer (SWM), G/N Ward',
   'Assistant Engineer (SWM), G/N Ward',
   'Harishchandra Yewale Marg, behind Plaza Cinema, Dadar, Mumbai 400028',
   '24212776',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'G/N', 'water_drainage',
   'Water Works (Hydraulic Engineer) — G/N Ward',
   'Assistant Engineer (Water Works), G/N Ward',
   'Assistant Engineer (Water Works), G/N Ward',
   'Harishchandra Yewale Marg, behind Plaza Cinema, Dadar, Mumbai 400028',
   '24212776',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'H/E', 'waste',
   'Solid Waste Management — H/E Ward',
   'Assistant Engineer (SWM), H/E Ward',
   'Assistant Engineer (SWM), H/E Ward',
   'Prabhat Colony, T.P.S. 5, Santacruz (East), Mumbai 400055',
   '26182217',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'H/E', 'water_drainage',
   'Water Works (Hydraulic Engineer) — H/E Ward',
   'Assistant Engineer (Water Works), H/E Ward',
   'Assistant Engineer (Water Works), H/E Ward',
   'Prabhat Colony, T.P.S. 5, Santacruz (East), Mumbai 400055',
   '26182217',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'H/W', 'waste',
   'Solid Waste Management — H/W Ward',
   'Assistant Engineer (SWM), H/W Ward',
   'Assistant Engineer (SWM), H/W Ward',
   'St Martins Road, Bandra (West), Mumbai 400050',
   '26422311',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'H/W', 'water_drainage',
   'Water Works (Hydraulic Engineer) — H/W Ward',
   'Assistant Engineer (Water Works), H/W Ward',
   'Assistant Engineer (Water Works), H/W Ward',
   'St Martins Road, Bandra (West), Mumbai 400050',
   '26422311',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'K/E', 'waste',
   'Solid Waste Management — K/E Ward',
   'Assistant Engineer (SWM), K/E Ward',
   'Assistant Engineer (SWM), K/E Ward',
   'Azad Road, Gundavali, Andheri (East), Mumbai 400069',
   '26837104',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'K/E', 'water_drainage',
   'Water Works (Hydraulic Engineer) — K/E Ward',
   'Assistant Engineer (Water Works), K/E Ward',
   'Assistant Engineer (Water Works), K/E Ward',
   'Azad Road, Gundavali, Andheri (East), Mumbai 400069',
   '26837104',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'K/W', 'waste',
   'Solid Waste Management — K/W Ward',
   'Assistant Engineer (SWM), K/W Ward',
   'Assistant Engineer (SWM), K/W Ward',
   'Ward Office Building, Paliram Road, beside Swami Vivekanand Road, opposite Andheri Station, Andheri (West), Mumbai 400058',
   '26239131',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'K/W', 'water_drainage',
   'Water Works (Hydraulic Engineer) — K/W Ward',
   'Assistant Engineer (Water Works), K/W Ward',
   'Assistant Engineer (Water Works), K/W Ward',
   'Ward Office Building, Paliram Road, beside Swami Vivekanand Road, opposite Andheri Station, Andheri (West), Mumbai 400058',
   '26239131',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'L', 'waste',
   'Solid Waste Management — L Ward',
   'Assistant Engineer (SWM), L Ward',
   'Assistant Engineer (SWM), L Ward',
   'Laxmanrao Yadav Mandai Building, S.G. Barve Marg, Kurla, Mumbai 400070',
   '26505103',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'L', 'water_drainage',
   'Water Works (Hydraulic Engineer) — L Ward',
   'Assistant Engineer (Water Works), L Ward',
   'Assistant Engineer (Water Works), L Ward',
   'Laxmanrao Yadav Mandai Building, S.G. Barve Marg, Kurla, Mumbai 400070',
   '26505103',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'M/E', 'waste',
   'Solid Waste Management — M/E Ward',
   'Assistant Engineer (SWM), M/E Ward',
   'Assistant Engineer (SWM), M/E Ward',
   'Municipal Office Building, Late Madhukar Tukaram Kadam Marg near Hirakeri Road junction, Deonar Municipal Colony, Govandi, Mumbai 400043',
   NULL,
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'M/E', 'water_drainage',
   'Water Works (Hydraulic Engineer) — M/E Ward',
   'Assistant Engineer (Water Works), M/E Ward',
   'Assistant Engineer (Water Works), M/E Ward',
   'Municipal Office Building, Late Madhukar Tukaram Kadam Marg near Hirakeri Road junction, Deonar Municipal Colony, Govandi, Mumbai 400043',
   NULL,
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'M/W', 'waste',
   'Solid Waste Management — M/W Ward',
   'Assistant Engineer (SWM), M/W Ward',
   'Assistant Engineer (SWM), M/W Ward',
   'Sharadbhau Acharya Marg, Chembur, Mumbai 400071',
   '25283900',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'M/W', 'water_drainage',
   'Water Works (Hydraulic Engineer) — M/W Ward',
   'Assistant Engineer (Water Works), M/W Ward',
   'Assistant Engineer (Water Works), M/W Ward',
   'Sharadbhau Acharya Marg, Chembur, Mumbai 400071',
   '25283900',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'N', 'waste',
   'Solid Waste Management — N Ward',
   'Assistant Engineer (SWM), N Ward',
   'Assistant Engineer (SWM), N Ward',
   'Ward Office Building, Jawahar Road, Ghatkopar (East), Mumbai 400077',
   '25094161',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'N', 'water_drainage',
   'Water Works (Hydraulic Engineer) — N Ward',
   'Assistant Engineer (Water Works), N Ward',
   'Assistant Engineer (Water Works), N Ward',
   'Ward Office Building, Jawahar Road, Ghatkopar (East), Mumbai 400077',
   '25094161',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'P/S', 'waste',
   'Solid Waste Management — P/S Ward',
   'Assistant Engineer (SWM), P/S Ward',
   'Assistant Engineer (SWM), P/S Ward',
   'Municipal Office Building, C.T.S. No. 746, Hillage Pahadi, Swami Vivekanand Marg, Goregaon (West), Mumbai 400104',
   '28721186',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'P/S', 'water_drainage',
   'Water Works (Hydraulic Engineer) — P/S Ward',
   'Assistant Engineer (Water Works), P/S Ward',
   'Assistant Engineer (Water Works), P/S Ward',
   'Municipal Office Building, C.T.S. No. 746, Hillage Pahadi, Swami Vivekanand Marg, Goregaon (West), Mumbai 400104',
   '28721186',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'P/N', 'waste',
   'Solid Waste Management — P/N Ward',
   'Assistant Engineer (SWM), P/N Ward',
   'Assistant Engineer (SWM), P/N Ward',
   'Ward Office Building, Liberty Garden, Mamlatdar Wadi, Malad (West), Mumbai 400064',
   '28824245',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'P/N', 'water_drainage',
   'Water Works (Hydraulic Engineer) — P/N Ward',
   'Assistant Engineer (Water Works), P/N Ward',
   'Assistant Engineer (Water Works), P/N Ward',
   'Ward Office Building, Liberty Garden, Mamlatdar Wadi, Malad (West), Mumbai 400064',
   '28824245',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'R/S', 'waste',
   'Solid Waste Management — R/S Ward',
   'Assistant Engineer (SWM), R/S Ward',
   'Assistant Engineer (SWM), R/S Ward',
   'M.G. Chhed Road No. 2, near S.V.P. swimming pool, Kandivali (West), Mumbai 400067',
   '28054784',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'R/S', 'water_drainage',
   'Water Works (Hydraulic Engineer) — R/S Ward',
   'Assistant Engineer (Water Works), R/S Ward',
   'Assistant Engineer (Water Works), R/S Ward',
   'M.G. Chhed Road No. 2, near S.V.P. swimming pool, Kandivali (West), Mumbai 400067',
   '28054784',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'R/N', 'waste',
   'Solid Waste Management — R/N Ward',
   'Assistant Engineer (SWM), R/N Ward',
   'Assistant Engineer (SWM), R/N Ward',
   'Below Sangeetkar Sudhir Phadke flyover, Jayvant Sawant Marg, Dahisar (West), Mumbai 400068',
   '28936000',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'R/N', 'water_drainage',
   'Water Works (Hydraulic Engineer) — R/N Ward',
   'Assistant Engineer (Water Works), R/N Ward',
   'Assistant Engineer (Water Works), R/N Ward',
   'Below Sangeetkar Sudhir Phadke flyover, Jayvant Sawant Marg, Dahisar (West), Mumbai 400068',
   '28936000',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'S', 'waste',
   'Solid Waste Management — S Ward',
   'Assistant Engineer (SWM), S Ward',
   'Assistant Engineer (SWM), S Ward',
   'L.B. Shastri Marg, Bhandup (West), Mumbai 400078',
   '25947570',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'S', 'water_drainage',
   'Water Works (Hydraulic Engineer) — S Ward',
   'Assistant Engineer (Water Works), S Ward',
   'Assistant Engineer (Water Works), S Ward',
   'L.B. Shastri Marg, Bhandup (West), Mumbai 400078',
   '25947570',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'T', 'waste',
   'Solid Waste Management — T Ward',
   'Assistant Engineer (SWM), T Ward',
   'Assistant Engineer (SWM), T Ward',
   'Lala Devidayal Marg, Mulund (West), Mumbai 400080',
   '25645291',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'T', 'water_drainage',
   'Water Works (Hydraulic Engineer) — T Ward',
   'Assistant Engineer (Water Works), T Ward',
   'Assistant Engineer (Water Works), T Ward',
   'Lala Devidayal Marg, Mulund (West), Mumbai 400080',
   '25645291',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Consolidated PIO list updated to 01.01.2025. The same post the R/C '
     || 'handbook names for this work ([migration 0019]).',
   '2026-09-29')
ON CONFLICT DO NOTHING;
