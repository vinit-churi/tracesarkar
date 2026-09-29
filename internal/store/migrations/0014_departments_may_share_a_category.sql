-- 0014: a category may have more than one owning department.
--
-- 0013 keyed on (authority, ward, category, effective_from), which quietly
-- assumed one department owns one kind of complaint in one ward. BMC's own
-- RTI handbooks disprove that for the first category we looked at: a pothole
-- in a ward is split between the Chief Engineer (Roads & Traffic), who holds
-- major roads and the concretisation programme, and the ward's own
-- Maintenance & Repair staff, who hold minor roads.
--
-- Forcing one row would have meant choosing between them, and the choice is
-- not ours to make — it depends on the road. The platform should state both.

ALTER TABLE authority_departments
  DROP CONSTRAINT IF EXISTS authority_departments_unique;

ALTER TABLE authority_departments
  ADD CONSTRAINT authority_departments_unique
  UNIQUE (authority, ward, category, department, effective_from);
