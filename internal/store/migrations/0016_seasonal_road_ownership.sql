-- 0016: which department owns a road defect depends on the season.
--
-- 0015 recorded two owners for a road defect in R/C and left it at that. The
-- Chief Engineer (Roads & Traffic) manual, chapter 2 §2.5(iv), is more
-- specific than that: outside the monsoon the ward maintains both major and
-- minor roads, and during the monsoon major roads and bus routes pass to
-- contractors under central Roads & Traffic while minor roads stay with the
-- ward.
--
-- So the answer changes with the date of the report, not only with the road.
-- Recorded in the notes rather than modelled as a rule, because the monsoon's
-- start is declared, not fixed, and inventing a date range would be a legal
-- constant with no citation.
--
-- Also corrects an assumption worth naming: there is NO published width
-- threshold separating major from minor. Classification is per-road enumerated
-- lists in each ward manual, and width does not separate them reliably —
-- H/West lists a 19.6 m road as minor and 10 m roads as major. Jurisdiction
-- must never be derived from road width.

UPDATE authority_departments
   SET note = note || ' Seasonal: during the monsoon this department holds major '
                   || 'roads and bus routes through contractors; outside the monsoon '
                   || 'the ward maintains both classes. Ch.E (Roads & Traffic) manual '
                   || 'ch.2 s.2.5(iv). No published width threshold exists — never '
                   || 'derive the class from road width.'
 WHERE ward = 'R/C' AND category = 'road_defect'
   AND department = 'Chief Engineer (Roads & Traffic) — Western Suburbs';

UPDATE authority_departments
   SET note = note || ' Seasonal: this department holds minor roads year round, and '
                   || 'both classes outside the monsoon. Ch.E (Roads & Traffic) manual '
                   || 'ch.2 s.2.5(iv).'
 WHERE ward = 'R/C' AND category = 'road_defect'
   AND department = 'R/Central Ward — Maintenance & Repair';
