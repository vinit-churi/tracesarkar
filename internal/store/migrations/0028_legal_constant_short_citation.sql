-- 0028: a citation short enough to fit in a post.
--
-- `legal_constants.citation` is the full reference and often the quoted text,
-- which is right for a page and impossible in 280 characters. The temptation
-- is to truncate it at the edge, and a truncated legal citation is a wrong
-- citation: "High Court of Bombay, PIL 71/2013, order of 13 Octo…" cites
-- nothing a person can look up.
--
-- So the short form is recorded, not derived. A constant with no short form
-- simply does not appear in a post — the deadline sentence is dropped whole
-- rather than stated without the document that creates it.

ALTER TABLE legal_constants ADD COLUMN IF NOT EXISTS short_citation text;

UPDATE legal_constants SET short_citation = 'Bombay HC, PIL 71/2013, para 70(ix)'
 WHERE key = 'pothole.attend_hours.court' AND short_citation IS NULL;

UPDATE legal_constants SET short_citation = 'BMC tender ETH_8000040832, cl. 10.11'
 WHERE key = 'pothole.attend_hours.contract' AND short_citation IS NULL;
