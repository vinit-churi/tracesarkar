-- 0019: who owns garbage and a water leak in R/C (Borivali).
--
-- Until now the platform could route exactly one category. The classifier has
-- recognised six since it was built, so a photograph of a garbage pile was
-- classified correctly and then returned not_yet_covered — honest, and useless.
-- Coverage is read from this table at runtime, so these rows make waste and
-- water_drainage routable without a line of code changing.
--
-- Verified first-hand on 3 October 2026 from BMC's own Section 4(1)(b) RTI
-- handbooks for R/Central, archived in R2 because BMC drops KM paths without
-- notice.
--
-- Designations only, as in 0015. The source PDFs name the individuals holding
-- these posts; a name is personal data with a short shelf life, and the post is
-- what a complaint should address.

INSERT INTO authority_departments
  (id, authority, ward, category, department, officer_designation,
   pio_designation, pio_address, source_ref, retrieved_at, archive_key, note,
   effective_from)
VALUES
  (gen_random_uuid(), 'BMC', 'R/C', 'waste',
   'Solid Waste Management — R/Central Ward',
   'Assistant Engineer (SWM) R/Central',
   'Assistant Engineer (SWM) R/Central',
   'Office of Assistant Engineer (SWM) R/Central, Chandavarkar Road, Borivali (West), Mumbai 400092',
   'https://www.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/Wards/Assistant%20Commissioner%20(R-Central)/RTI%20Manuals/AESWMRCentralWardManuals.pdf',
   '2026-10-03', 'archive/bmc_portal/2026-10-03/solid-waste-rcentral-rti-manual-edb814131993.pdf',
   'Sweeping, refuse removal, silt and debris removal, dead animals. Parent department Chief Engineer (SWM); reports to Assistant Commissioner R/Central. The manual appoints this officer PIO for the department in its own introduction. Section 4(1)(b)(iii) publishes a 24-hour time limit for both refuse removal and silt/debris removal — see legal_constants.',
   '2026-10-03'),

  (gen_random_uuid(), 'BMC', 'R/C', 'water_drainage',
   'Water Works (Hydraulic Engineer) — R/Central Ward',
   'Assistant Engineer (Water Works) R/Central',
   'Assistant Engineer (Water Works) R/Central',
   'Office of Assistant Engineer (W.W.) R/Central, Room no. 02, Ground Floor, R/Central Ward Building, Chandavarkar Road, Borivali (West), Mumbai 400092',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/Wards/Assistant%20Commissioner%20(R-Central)/RTI%20Manuals/RTI%2017%20manual%20for%20AEWW%20RC%20ward.pdf',
   '2026-10-03', 'archive/bmc_portal/2026-10-03/water-works-rcentral-rti-manual-348da19b03ec.pdf',
   'Water supply: burst pipeline, leakage, contaminated supply. Part of the Hydraulic Engineer''s department. IMPORTANT: this manual publishes NO repair deadline. Every time limit in its Section 4(1)(b)(iii) concerns granting a water connection, meter reading or disconnection; public complaints are recorded as registrable with the ward Complaint Officer and the city Water Control Office, with no timeline attached. The platform must say it does not have a deadline for this category rather than borrow the pothole one. An RTI for the repair timeline is the obvious follow-up.',
   '2026-10-03');

-- The deadlines the platform shows. The table was empty until now, which meant
-- every clock in the product was either absent or a literal somewhere — the
-- thing hard rule 6 exists to prevent.
INSERT INTO legal_constants
  (id, key, value, citation, citation_url, verification, effective_from)
VALUES
  (gen_random_uuid(), 'pothole.attend_hours.court', '48'::jsonb,
   'High Court of Bombay, PIL 71/2013, order of 13 October 2025, paragraph 70(ix), neutral citation 2025:BHC-OS:18736-DB. "All potholes, once brought to the notice of the concerned Corporation or Authority, shall be attended to forthwith and, in any event, within forty-eight hours."',
   NULL, 'verified', '2025-10-13'),

  (gen_random_uuid(), 'pothole.attend_hours.contract', '24'::jsonb,
   'BMC road tender ETH_8000040832, clause 10.11: "All Defective work must be rectified within 24 hours of intimation". Binds the contractor during the defect liability period.',
   'https://www.mcgm.gov.in/irj/go/km/docs/documents/Tenders/ETH/ETH_8000040832_020523.pdf',
   'verified', '2023-05-02'),

  (gen_random_uuid(), 'pothole.penalty_per_day_paise', '500000'::jsonb,
   'BMC road tender ETH_8000040832, clause 7.4: "A penalty of Rs. 5000/- per day per Defect/Pothole/Trench if not attended in stipulated time."',
   'https://www.mcgm.gov.in/irj/go/km/docs/documents/Tenders/ETH/ETH_8000040832_020523.pdf',
   'verified', '2023-05-02'),

  (gen_random_uuid(), 'waste.refuse_removal_hours', '24'::jsonb,
   'BMC R/Central Solid Waste Management RTI manual, Section 4(1)(b)(iii): activity "Sweeping of roads & Removal of refuse", time limit "Within 24 Hours". MMC Act 1888 s.365(a); circular DMC/ENV SWM/4345 dated 16.03.2006.',
   'https://www.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/Wards/Assistant%20Commissioner%20(R-Central)/RTI%20Manuals/AESWMRCentralWardManuals.pdf',
   'verified', '2006-03-16'),

  (gen_random_uuid(), 'waste.silt_debris_removal_hours', '24'::jsonb,
   'Same manual, Section 4(1)(b)(iii): activity "Removal of Silt & Debris", time limit "Within 24 Hours". MMC Act 1888 s.375(A); same circular.',
   'https://www.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/Wards/Assistant%20Commissioner%20(R-Central)/RTI%20Manuals/AESWMRCentralWardManuals.pdf',
   'verified', '2006-03-16');
