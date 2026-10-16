-- 0045: one spelling per ward, because two spellings read as two wards.
--
-- `escalates_to` is printed on the public ward page under a heading that
-- carries the ward's own code. Every ward's rows name that ward the way the
-- code does — "F/N Ward", "K/E Ward", "M/E Ward" — except R/Central, which
-- was seeded from documents that spell it out, and which ended up carrying
-- both: 2 rows saying "R/C Ward" and 6 saying "R/Central Ward".
--
-- That is not a typo, it is an ambiguity. A resident reading "R/C" in the
-- heading and "Assistant Commissioner, R/Central Ward" in the appeal line has
-- no way to know those are the same office, and the one thing an escalation
-- line has to be is unambiguous about which office to write to.
--
-- The code form wins rather than BMC's expanded form, for one reason: the
-- heading above it is the code, and the two have to agree. Expanding every
-- ward instead would mean rewriting twenty-four wards' rows and the heading
-- on the page, to gain nothing a reader needs.
--
-- The invariant is asserted by a live test — every `escalates_to` that names
-- a ward names *this* row's ward — so the next document read from a BMC PDF
-- cannot quietly reintroduce a second spelling.

UPDATE authority_departments
   SET escalates_to = replace(escalates_to, 'R/Central Ward', 'R/C Ward')
 WHERE escalates_to LIKE '%R/Central Ward%';

-- The address inside one of those strings names a different ward's office on
-- purpose — the Dy. Hydraulic Engineer for the Western Suburbs sits in the
-- K/West ward office — so only the ward's own name is rewritten, not every
-- ward name in the string.

DO $$
DECLARE n integer;
BEGIN
  SELECT count(*) INTO n FROM authority_departments
   WHERE escalates_to LIKE '%R/Central Ward%';
  IF n <> 0 THEN
    RAISE EXCEPTION '% row(s) still spell R/C as R/Central', n;
  END IF;

  -- Every row whose escalation names a ward names its own ward. The
  -- exception is a row that names another ward's *office address*, which is
  -- why the check is "contains this ward" rather than "contains no other".
  SELECT count(*) INTO n FROM authority_departments
   WHERE escalates_to ILIKE '%ward%'
     AND escalates_to NOT LIKE '%' || ward || ' Ward%'
     AND escalates_to NOT LIKE '%' || ward || ' ward%';
  IF n <> 0 THEN
    RAISE EXCEPTION '% escalation(s) name a ward other than their own', n;
  END IF;
END $$;
