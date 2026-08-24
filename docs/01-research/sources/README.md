# Source archive

Retrieved artefacts and probe records, one file per source, named `YYYY-MM-DD-<source>`.

The archive — not a parsed row and not a summary — is the evidentiary record. Anything in
[`../08-data-availability-audit.md`](../08-data-availability-audit.md) marked ✅ has its evidence
here.

| File | What it records |
|---|---|
| [Bombay HC PIL 71/2013, paras 69–70](2026-08-24-bombay-hc-pil-71-2013-paras-69-70.md) | The operative pothole directions, verbatim. Working transcript; the certified copy is still to be obtained |
| [BMC roads API probe](2026-08-24-bmc-roads-api-probe.md) | Endpoints, field schema, counts, distributions, and the two constraints on use |
| [BMC roads sample](2026-08-24-bmc-roads-sample-25.json) | First 25 records of the public dashboard response |
| [MCGM ArcGIS and CPCB probe](2026-08-24-mcgm-arcgis-and-cpcb-probe.md) | 106 ArcGIS services, two layer schemas, the live air-quality feed, and what was found closed |

**When production ingestion starts, this hand-kept directory is replaced by the object-storage
archive described in [ingestion](../../03-architecture/07-ingestion-and-scrapers.md)** — hashed,
timestamped, one object per retrieval. These files are the pre-code stand-in, not the design.
