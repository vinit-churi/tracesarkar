-- 0014: who owns a road defect in R/C (Borivali).
--
-- Verified first-hand on 29 September 2026 from BMC's own Section 4(1)(b) RTI
-- handbooks, which are archived in R2 because BMC drops KM paths without
-- notice — its Citizen's Charter URL already 404s.
--
-- A pothole in a BMC ward has TWO owners, not one, and getting this wrong
-- sends the complaint to a desk that cannot act on it:
--
--   * major roads and the concretisation programme -> Chief Engineer
--     (Roads & Traffic), through zonal staff and contractors;
--   * minor roads -> the ward's own Maintenance & Repair staff.
--
-- Both rows are recorded. Which applies depends on the road class, and the
-- platform states both rather than guessing between them.
--
-- Designations only. The source PDFs name the individuals currently holding
-- these posts; a name is personal data with a short shelf life, and the post
-- is what a complaint should address.

INSERT INTO authority_departments
  (id, authority, ward, category, department, officer_designation,
   pio_designation, pio_address, source_ref, retrieved_at, archive_key, note,
   effective_from)
VALUES
  (gen_random_uuid(), 'BMC', 'R/C', 'road_defect',
   'Chief Engineer (Roads & Traffic) — Western Suburbs',
   'Assistant Engineer (Roads) R/Central',
   'Assistant Engineer (Roads) R/Central',
   'Office of Dy. Ch. Eng. Roads (W.S.), 5th Floor, P/S Ward Office Building, '
     || 'S. V. Road, Goregaon (West), Mumbai 400062',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/MCGM%20Department%20List/'
     || 'Roads%20and%20Traffic/RTI%20Manuals/Westernsuburbs/Westernsuburbs_RTI_E01.pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/roads-western-suburbs-rti-manual-f2e1119fe64b.pdf',
   'Major roads and the concretisation programme. First Appellate Authority: '
     || 'Executive Engineer (Roads) K/E, R/C & R/N — one EE covers three wards.',
   '2026-09-29'),

  (gen_random_uuid(), 'BMC', 'R/C', 'road_defect',
   'R/Central Ward — Maintenance & Repair',
   'Assistant Engineer (Maintenance & Repair) R/Central',
   'Assistant Engineer (Maintenance) R/Central',
   'Mahapalika Mandai Building, Swami Vivekanand Marg, Borivali (West), Mumbai 400092',
   'https://portal.mcgm.gov.in/irj/go/km/docs/documents/RTI/List%20of%20RTI%20PIOs%20'
     || 'and%20FAOs/LIST%20OF%20P.I.O''s%20AND%20F.A.A''s(8800662725).pdf',
   '2026-09-29T00:00:00Z',
   'archive/bmc_portal/2026-09-29/pio-and-faa-list-2025-01-01-0080cc66b40f.pdf',
   'Minor roads, maintained by ward staff. Consolidated PIO list updated to '
     || '01.01.2025; jurisdiction covers ward roads, storm-water drains, sewerage '
     || 'and minor-works contractors. First Appellate Authority: Executive '
     || 'Engineer, R/Central.',
   '2026-09-29');
