# Data sources

Every external feed the platform depends on, with acquisition strategy, expected format, refresh
cadence, and risk. Sources marked ⚠️ need verification before they are relied on.

---

## 1. Procurement and contracts

| Source | URL | What it gives | Acquisition | Cadence | Risk |
|---|---|---|---|---|---|
| **Mahatenders** (GePNIC) | `mahatenders.gov.in` | State-wide tender notices, corrigenda, award-of-contract for Maharashtra departments and many ULBs | HTML scrape; captcha and session handling likely; check for any published API | Daily | Layout changes; rate limiting; ⚠️ terms of use |
| **BMC tender portal** | `portal.mcgm.gov.in` | BMC-specific tenders and awards | Scrape | Daily | SAP-portal URLs are brittle |
| **CPP Portal** | `eprocure.gov.in` | Central-government tenders (Railways, port) | Scrape | Daily | — |
| **GeM** | `gem.gov.in` | Goods/services procurement, sometimes services relevant to civic works | Scrape | Weekly | — |
| **RTI** | — | Work orders, completion certificates, measurement books, DLP terms, payment records | Manual + generated RTI | Per issue | Slow; the only route to DLP data at present |
| **OCDS reference implementations** | `open-contracting.in`, CivicDataLab's Assam and Himachal publications, OCP Data Registry | Normalisation target and precedent for Indian OCDS mapping | Reference | Once | — |

**Critical gap:** the **defect liability period** is almost never in the machine-readable tender
fields. It lives in the contract conditions, usually a PDF annexure. Extracting it is a VLM/OCR task
and the single highest-value extraction the platform performs. See
[tender engine](../03-architecture/06-tender-engine.md).

**Known baseline (verify per contract):** BMC concrete roads carry a guarantee of roughly 5–10 years;
asphalt roads roughly 3 years; ~20% of contract value is typically withheld until the guarantee
period ends; defects appearing inside the DLP must be repaired free of cost.

---

## 2. Geospatial

| Source | URL | What it gives | Licence | Risk |
|---|---|---|---|---|
| **BMC ArcGIS REST** | `services8.arcgis.com/r6MmJtuWAzMawmJ8/ArcGIS/rest/services`, `mybmcid.mcgm.gov.in/server/rest/services` | BMC's published map layers — enumerate for wards, roads, drains, assets | ⚠️ unstated | Endpoint may change or close |
| **DataMeet Municipal_Spatial_Data** | `github.com/datameet/Municipal_Spatial_Data` | Municipal ward polygons, GeoJSON, WGS84/EPSG:4326 | CC BY-SA 2.5 IN | Vintage varies; MMR coverage incomplete (see issue #53) |
| **sanjanakrishnan/mumbai_spatial_data** | GitHub | MCGM census wards, 2017 and 2022 prabhags | ⚠️ check | — |
| **DataMeet PincodeBoundary** | GitHub | Mumbai pincode polygons | CC BY-SA | Useful as a coarse fallback |
| **OpenStreetMap / Overpass** | `overpass-api.de` | Road centrelines, names, `admin_level` relations, POIs | ODbL | Not authoritative for ownership; ODbL share-alike obligations |
| **MMRDA planning pages** | `mmrda.maharashtra.gov.in` | Regional plan, Extended Notified Area | ⚠️ | PDF-only; needs digitisation |
| **Bhuvan (ISRO)** | `bhuvan.nrsc.gov.in` | Base imagery, some admin layers | ⚠️ | Access terms vary |
| **Sentinel-2 (Copernicus)** | `dataspace.copernicus.eu` | 10 m optical, ~5-day revisit — change detection for "ghost project" checks | Open | 10 m is too coarse for most road work; useful for large sites, afforestation, land filling |

---

## 3. Weather, flooding, and environment

| Source | URL | What it gives | Notes |
|---|---|---|---|
| **IITM Mumbai MESONET** | `mumbairain.tropmet.res.in` | Live rainfall at ~139 sites across Mumbai; data stated to be freely available for research/academic use | ⚠️ confirm terms for a public platform; ~165 automatic rain gauges installed jointly by IITM, BMC and IMD |
| **Mumbai Flood (IIT-B / BMC / HDFC ERGO)** | `mumbaiflood.in` | Near-real-time waterlogging and flood forecasting; Android app | Participating agencies include IITM, IMD, MCGM, CR, WR, NMMC |
| **BMC AWS network** | via BMC disaster management | ~60 BMC automatic weather stations | ⚠️ access route unclear |
| **IMD** | `mausam.imd.gov.in` | Colaba, Santacruz, Marine Lines observatories; warnings | Public |
| **iFLOWS-Mumbai** | NCCR/MoES | Integrated flood warning system for Mumbai | ⚠️ citizen-facing feed availability unclear |
| **SAFAR** | `safar.tropmet.res.in` | Air quality and weather across Mumbai | Public |
| **CPCB / MPCB** | `cpcb.nic.in`, `mpcb.gov.in` | AQI stations, consent-to-operate records, action taken | Useful for environmental escalations |

**Product use:** correlate rainfall intensity with report volume; pre-warn wards with known
waterlogging hotspots; contextualise an issue ("this location has flooded on 6 of the last 9
>50 mm/day events").

---

## 4. News and press

| Source | Acquisition | Use |
|---|---|---|
| Local English dailies (Mumbai Mirror, Free Press Journal, Mid-Day, Hindustan Times Mumbai, Indian Express Mumbai, Times of India Mumbai) | RSS where available; otherwise polite scraping | News contextualisation on issues |
| Marathi dailies (Loksatta, Maharashtra Times, Lokmat, Sakal) | RSS / scrape | Vernacular coverage, often earlier and more local |
| **PIB** | `pib.gov.in` | Official union announcements |
| **Maharashtra DGIPR** | `dgipr.maharashtra.gov.in` | State press releases, Government Resolutions |
| **Corporation press notes** | Per-corporation sites | Mega-block notices, works announcements, tender notices |
| **Railway mega-block notices** | WR/CR sites, `RailMadad` | Transit-mode context |

**Legal note:** store URL, headline, publication, date, and a short extract. Do **not** republish full
article text. Link out.

---

## 5. Grievance and government platforms

| Platform | Direction | Notes |
|---|---|---|
| **MyBMC MARG / 24x7** | Write (file) + read (status) | 114 grievance types; no public API known — likely form automation |
| **Aaple Sarkar grievance** | Write + read | Token number; ~21 working days stated |
| **CPGRAMS** | Write + read | For union-government subjects |
| **Swachhata (MoHUA)** | Write | Nationwide sanitation; ward sanitary-inspector routing |
| **RailMadad** | Write | Web/app/SMS/139; PNR or UTS linkage; 1,000-char limit; no post-submission edits |
| **e-Jagriti** | Write | Consumer commission filing |
| **rtionline.maharashtra.gov.in** | Write | RTI + first appeal |
| **NGT e-filing** | Write | Original applications |

⚠️ **None of these is confirmed to expose an official API.** Assume form automation with a
human-in-the-loop, and design filing adapters accordingly. Where automation is not permissible under
a platform's terms, the adapter degrades to "generate a filled draft + deep link + instructions".

---

## 6. Corporate and blacklisting data

| Source | What it gives | Notes |
|---|---|---|
| **MCA21 / MCA master data** | Company CIN, directors, registered address, status | The join key for corporate-lineage mapping |
| **MCA DIN records** | Director identification across companies | Enables "same directors, new company" detection |
| **GSTIN lookup** | Registration status, legal name | Secondary identity check |
| **Corporation blacklist notices** | Debarred contractor lists | Published irregularly; often PDF; per-corporation |
| **CVC / departmental debarment lists** | Central debarments | Occasionally relevant |

**Precedent worth citing in the docs:** after the 2016 BMC road-repair scandal several contractors
were blacklisted, and allegations persist about re-entry through proxy firms. More recently a
contractor de-silting minor nallahs in M-West ward was blacklisted for adulterating silt with
construction debris to inflate weighment — a fraud detected using AI video analysis. ⚠️ Both need
primary-source verification before publication.

---

## 7. Reference / standards

| Source | Use |
|---|---|
| **Open311 GeoReport v2** (`wiki.open311.org/GeoReport_v2/`) | API compatibility target. 6 methods; ISO 8601 with timezone; UTF-8 mandatory. Adopted by Chicago, Toronto, SF, Boston, Helsinki, Bonn and others. |
| **OCDS** (`open-contracting.org/data-standard/`) | Contract lifecycle normalisation |
| **Bhashini / ULCA** (`bhashini.gitbook.io/bhashini-apis`) | ASR, NMT, TTS across 22 Indian languages. Pipeline config call → compute call; `userID` + `ulcaApiKey` + inference key. |
| **DigiLocker / Aadhaar-based verification** | ⚠️ Deliberately deferred — see [trust and anti-abuse](../02-product/08-trust-and-antiabuse.md) |
| **data.gov.in** | Miscellaneous open datasets; quality varies |

---

## 7A. Verified additions — August 2026 audit

Full evidence in [`08-data-availability-audit.md`](08-data-availability-audit.md). These rows were
confirmed by direct fetch on 24 August 2026 and supersede the assumptions above where they conflict.

| Source | Endpoint | What it gives | Status |
|---|---|---|---|
| **BMC roads dashboard API** | `https://roads.mcgm.gov.in:3000/api/` | `publicdashboard/` → 2,237 CC-road works with contractor, ward, status, dates, completion photo. `…roadlayer` → 2,405 road geometries with `workCode`, contractor, dates, PQC progress. `geo/getwardlayer` → 24 ward polygons. `ward/` → ward master | **Confirmed open, unauthenticated.** Terms not located; `robots.txt` 404. Highest-value source found |
| **BMC ArcGIS Online** | `services8.arcgis.com/r6MmJtuWAzMawmJ8/…` | 106 services. `Prabhaag_Boundary` (227 polygons, `COUNCILLOR`, `JR_ENGG`), `Streets` (`CARRIAGE_W`, `NO_OF_LANE`, `DRV_STATUS`), `BMC_Ward`, `DP_KYW`, `DP_MIS`, building footprints | Confirmed open, no token. **`copyrightText` empty — licence unstated.** Get a written position before republishing |
| **MIDC GIS** | `gis.midcindia.org/server/rest/services/CitizenPortal/…` | 306 industrial-area polygons, plot boundaries, ROW | Confirmed open; licence unstated |
| **CPPP debarment lists** | `eprocure.gov.in/cppp/` → Debarred Bidders | Central debarment register with archive and revocations | Crawlable per robots.txt |

### Acquisition decisions recorded

| Source | Decision | Basis |
|---|---|---|
| **Mahatenders** | **No automated collection.** Link, cite, and retrieve manually; use RTI or a written permission request for anything larger | `robots.txt` is `Disallow: /`; AOC search is captcha-gated; copyright policy requires permission to reproduce. [Hard rule 8](../../CLAUDE.md) |
| GeM | No automated collection | Captcha-gated |
| Commercial tender aggregators | Not used without a licence agreement | Their terms forbid automated access |
| BMC roads API | Ingest, with a conservative rate limit, full artefact archival, and a terms review recorded before the first production run | Public endpoint serving a public dashboard |
| Survey of India Nakshe | Not used | Terms forbid export and commercialisation |
| Google / Esri imagery | Not used as data; Esri tiles display-only if ever needed | Licence |

### Mandatory ingestion filter

The BMC roads API returns `contractorRepName` and `contractorRepMobile` — **personal contact details
of named individuals**. The parser drops both fields at ingestion. They are never stored, never
logged, and never reach a derivative. The same rule applies to any personal contact field
encountered in any source.

## 8. Acquisition ethics and hygiene

Non-negotiable rules for every scraper:

1. **Identify honestly.** A descriptive User-Agent with a contact URL.
2. **Respect `robots.txt`** and any published terms. Where terms forbid automated access, do not
   scrape — use RTI instead and record that decision in the source register.
3. **Rate limit conservatively.** Government infrastructure is fragile and often shared. Default:
   one request per 3 seconds per host, with exponential backoff on any 4xx/5xx.
4. **Cache aggressively.** Never re-fetch an unchanged document. Store ETag/Last-Modified.
5. **Archive the raw artefact.** Every scraped page and PDF is stored verbatim with a SHA-256 and a
   retrieval timestamp. The archive — not the parsed row — is the evidentiary record.
6. **Fail loudly.** A parser that silently produces zero rows is worse than one that errors.
7. **Never scrape personal data** that is not already published for the purpose we are using it for.

## 9. Source register

Every source must be registered in `data/sources.yaml` (to be created) with:

```yaml
- id: mahatenders
  name: Maharashtra eProcurement (Mahatenders)
  url: https://mahatenders.gov.in
  category: procurement
  acquisition: scrape
  cadence: daily
  licence: unverified
  terms_reviewed: null          # date + reviewer
  robots_txt_checked: null
  contact_attempted: null
  archive_prefix: s3://tracesarkar-archive/mahatenders/
  parser: internal/ingest/mahatenders
  owner: null
  status: planned
```

## 10. Research tasks

- [ ] Enumerate and document every BMC ArcGIS layer
- [ ] Confirm terms of use for IITM MESONET and Mumbai Flood data
- [ ] Determine whether Mahatenders exposes any documented API or bulk export
- [ ] Confirm whether any MMR corporation exposes an Open311 or equivalent endpoint
- [ ] Build the per-corporation grievance-channel survey (all 9 corporations + 9 councils)
- [ ] Review Bhashini's terms for non-government platform usage
- [ ] Establish the archive bucket, retention policy, and hash-chain for scraped artefacts
