-- 0024: narrow the append-only rule from "never changes" to "never changes
-- invisibly".
--
-- 0023 blocked UPDATE and DELETE on issue_events. Blocking DELETE turned out
-- to be wrong in two ways.
--
-- It made an issue undeletable, because the cascade from `issues` is itself a
-- delete on `issue_events`. The DPDP Act gives a person the right to have
-- their data erased, and a table nothing can ever be removed from cannot
-- honour that — the runbook has a data-subject-request procedure that this
-- would have made impossible to carry out.
--
-- And it protected the wrong thing. The attack the hash chain exists to defeat
-- is the *invisible* one: changing what an authority is recorded as having
-- said, or when, while the history still reads as continuous. That is what
-- UPDATE does and what verification catches. Removing an entire issue is a
-- different act — visible, because the issue is gone, and sometimes required.
--
-- So UPDATE stays forbidden and DELETE is allowed. Editing history remains
-- impossible; erasing an issue outright remains possible and obvious.
DROP TRIGGER IF EXISTS issue_events_no_update ON issue_events;

CREATE OR REPLACE FUNCTION issue_events_are_append_only() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'issue_events is append-only: an event is never edited once written. To remove a history, delete the issue.';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER issue_events_no_update
  BEFORE UPDATE ON issue_events
  FOR EACH ROW EXECUTE FUNCTION issue_events_are_append_only();
