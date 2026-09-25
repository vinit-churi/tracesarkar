-- 0007: one photograph, one media row.
--
-- A capture that is retried after a dropped connection re-uploads the same
-- bytes. SaveReport is already idempotent on the client's Idempotency-Key, but
-- the media insert was not, so each retry added another row pointing at the
-- same object. That inflates media counts and would double-count a photograph
-- in any evaluation export.
--
-- The digest is the identity: the same bytes under the same report are the same
-- photograph, whatever the role or filename says.

DELETE FROM report_media a
 USING report_media b
 WHERE a.report_id = b.report_id
   AND a.sha256 = b.sha256
   AND a.ctid > b.ctid;

CREATE UNIQUE INDEX IF NOT EXISTS report_media_report_sha_idx
    ON report_media (report_id, sha256);
