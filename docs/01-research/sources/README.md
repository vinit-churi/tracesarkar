# Source archive

Retrieved artefacts and probe records, one file per source, named `YYYY-MM-DD-<source>`.

The archive — not a parsed row and not a summary — is the evidentiary record. Anything in
[`../08-data-availability-audit.md`](../08-data-availability-audit.md) or
[`../09-vertical-exploration.md`](../09-vertical-exploration.md) marked ✅ has its evidence here.

| File | What it records |
|---|---|
| [Bombay HC PIL 71/2013, paras 69–70](2026-08-24-bombay-hc-pil-71-2013-paras-69-70.md) | The operative pothole directions, verbatim. Working transcript; the certified copy is still to be obtained |
| [BMC roads API probe](2026-08-24-bmc-roads-api-probe.md) | Endpoints, field schema, counts, distributions, and the two constraints on use |
| [BMC roads sample](2026-08-24-bmc-roads-sample-25.json) | First 25 records of the public dashboard response |
| [MCGM ArcGIS and CPCB probe](2026-08-24-mcgm-arcgis-and-cpcb-probe.md) | 106 ArcGIS services, two layer schemas, the live air-quality feed, and what was found closed |
| [PWD DLP GR, 14 Jan 2019](2026-09-18-pwd-dlp-gr-2019.md) | The defect liability table and inspection duties, translated, with the PDF hash |
| [BMC issue-domain layers](2026-09-18-bmc-issue-layers-probe.md) | Incident history, hoardings, toilets, the C1 list, and the layers that expose personal data |
| [Works, permits and legal sources](2026-09-18-works-permits-and-law-probe.md) | Roads API re-check, SWD desilting API, MahaRERA, the GR archive, BMC portal sources, court data |
| [Geographies beyond BMC](2026-09-18-geography-probe.md) | Pune's works GIS, the MMR corporations, other cities, and the Pothole Reporter repository |

**When production ingestion starts, this hand-kept directory is replaced by the object-storage
archive described in [ingestion](../../03-architecture/07-ingestion-and-scrapers.md)** — hashed,
timestamped, one object per retrieval. These files are the pre-code stand-in, not the design.
