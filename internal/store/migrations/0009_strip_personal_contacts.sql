-- 0009: remove personal contact details from the parsed works rows.
--
-- BMC's roads API returns the quality-monitoring agency representative's name
-- and personal mobile number on every record. The source register's blocklist
-- named contractorRepName and contractorRepMobile — fields the API does not
-- return — so it stripped nothing, and 2,387 mobile numbers reached the
-- database before anyone looked. The register is corrected; this removes what
-- was already stored.
--
-- Scope, deliberately: the parsed `works` rows and nothing else. The archived
-- raw documents are the evidentiary record and are not rewritten — that is the
-- whole point of keeping them. They are access-controlled and never rendered
-- on any surface, while `works` is the queryable projection everything
-- downstream reads from and is where a leak would actually happen.

UPDATE works
   SET current = current - 'qmaRepName' - 'qmaRepMobile'
 WHERE current ? 'qmaRepName' OR current ? 'qmaRepMobile';

-- The change log stores both values of every field it saw change. If a mobile
-- number ever changed, the old and new numbers are both in there.
DELETE FROM work_changes
 WHERE field IN ('qmaRepName', 'qmaRepMobile');
