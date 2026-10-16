-- 0039: the rows seeded this week now cite the archive, not just a URL.
--
-- Migrations 0033, 0036 and 0037 deliberately left `archive_key` null and said
-- why: the handbooks behind them were read from the live portal and were not
-- in R2, and pointing at another document's key would have cited a document
-- that says nothing about the subject. That was honest and temporary. The
-- documents are in the bucket now — `ingest archive-docs` fetched and stored
-- thirty-six of them on 16 October 2026 — so the rows can point at the thing
-- that will still exist when the portal has rewritten its paths again.
--
-- Keys are content hashes, so they are written here as literals rather than
-- joined on the URL. The URL is not a reliable join key: the same document is
-- spelled `Circular (Revised) Maint of Bridge.pdf` in one row and
-- `Circular%20%28Revised%29...` in another, and a join that silently matches
-- nothing would leave the row looking archived when it is not. A live test
-- asserts that every routing row has a key, and another asserts that every
-- cited key is actually in the bucket.

-- Street furniture (0033) and the ward Building & Factory desk (0036) both
-- rest on an A-ward Section 4 handbook.
UPDATE authority_departments
   SET archive_key = 'archive/bmc_portal_rti_manuals/2026/10/16/'
     || '66c55b7bb057ec3723e80ea83bd06679c8f3773a1102f6f8476d8ee2e78006e7.pdf'
 WHERE authority = 'BMC' AND category = 'street_furniture' AND archive_key IS NULL;

UPDATE authority_departments
   SET archive_key = 'archive/bmc_portal_rti_manuals/2026/10/16/'
     || '9ea624be2a10c043bee60e76a40b56ee5fbfc76348141dd2bdd312e6149d2805.pdf'
 WHERE authority = 'BMC' AND category = 'structural' AND subcategory IS NULL
   AND archive_key IS NULL;

-- The bridge rows rest on the revised maintenance circular of 28.06.2019.
UPDATE authority_departments
   SET archive_key = 'archive/bmc_portal_rti_manuals/2026/10/16/'
     || 'aa0db59ce1d11bdde51db6da013c6c0d6f50927d4b92a7a3d82158701fb790b0.pdf'
 WHERE authority = 'BMC' AND category = 'structural'
   AND subcategory = 'cracked bridge/FOB' AND archive_key IS NULL;

-- The Tree Authority rows rest on Manual 16 of the Garden handbook.
UPDATE authority_departments
   SET archive_key = 'archive/bmc_portal_rti_manuals/2026/10/16/'
     || '08268687c737678ba1c9598e93c50785c3b8c1e3a324ce8021ef5d8761fe82bf.pdf'
 WHERE authority = 'BMC' AND category = 'environment'
   AND subcategory = 'tree felling' AND archive_key IS NULL;

-- A legal constant is a published figure the platform repeats on a public
-- surface, and until now the only thing standing behind one was a URL. Hard
-- rule 6 puts the constant in a table with a citation; the citation is only
-- worth as much as the document, and the document has to be somewhere that
-- does not change under it.
ALTER TABLE legal_constants ADD COLUMN IF NOT EXISTS archive_key text;

UPDATE legal_constants
   SET archive_key = 'archive/bmc_portal_rti_manuals/2026/10/16/'
     || '6d27555248a76a96c3c176b3d13c3b93b19be2667b7ffab488bc91adf48b4f51.pdf'
 WHERE key IN ('waste.refuse_removal_hours', 'waste.dead_animal_removal_hours')
   AND effective_from = DATE '2020-12-07';

UPDATE legal_constants
   SET archive_key = 'archive/bmc_portal_rti_manuals/2026/10/16/'
     || '649b2f8f44e7ad396b0decf25a8db75a8909d9b3b0d7ae408e9b6d512306e5cb.pdf'
 WHERE key LIKE 'tree.%';

DO $$
DECLARE n integer;
BEGIN
  SELECT count(*) INTO n FROM authority_departments
   WHERE archive_key IS NULL OR archive_key = '';
  IF n <> 0 THEN
    RAISE EXCEPTION '% routing row(s) still cite no archived document', n;
  END IF;
  SELECT count(*) INTO n FROM legal_constants
   WHERE effective_from >= DATE '2016-09-29' AND archive_key IS NULL
     AND key NOT LIKE 'pothole.%';
  IF n <> 0 THEN
    RAISE EXCEPTION '% constant(s) from a portal document cite no archive key', n;
  END IF;
END $$;
