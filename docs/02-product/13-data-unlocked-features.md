# Features unlocked by verified public data

Derived from the [August 2026 data availability audit](../01-research/08-data-availability-audit.md).
Each feature here is grounded in a source that was actually checked. Features whose data turned out
not to exist are in §7 so nobody proposes them again.

Legend — **V** value, **C** cost, **R** risk (1–5), **M** milestone. Feature IDs continue the
[feature catalog](02-feature-catalog.md).

---

## 1. From BMC's roads API — the big one

`roads.mcgm.gov.in:3000/api/` publishes 2,237 CC-road works and 2,405 road geometries with
contractor, work code, dates, status and completion photographs. Verified open.

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| K1 | **Instant contract attribution on concretised roads** — spatial join of the report point to the published road geometry | 5 | 2 | 2 | v0.2 | Turns the hardest planned component into a PostGIS query for this road class |
| K2 | **"BMC says this road is complete" vs a dated citizen photograph** | 5 | 2 | 3 | v0.2 | The dashboard carries BMC's own completion photo and date. A defect photographed after it is the platform's sharpest single artefact |
| K3 | **Work-code → tender award join** | 5 | 3 | 3 | v0.2 | `workCode` (W-414, E-289, …) is the key into Mahatenders award records, retrieved manually |
| K4 | **Programme progress tracker** — status, PQC progress, quarter milestones, traffic-NOC dates per road | 4 | 2 | 2 | v0.3 | Ward-level "what was promised, what is built" from the authority's own numbers |
| K5 | **Schedule-slip detection** — planned vs actual dates already in the payload | 4 | 2 | 3 | v0.3 | Report the delta as a fact. Never characterise it |
| K6 | **Change-log watch** — snapshot the API daily; surface silent edits to status, dates or contractor | 5 | 2 | 2 | v0.3 | Because the archive is the evidentiary record, and government dashboards get edited |
| K7 | **Coverage honesty layer** — "this road is not in BMC's CC programme, so no contract is matched" | 5 | 1 | 1 | v0.2 | Prevents the false impression that unmatched means unaccountable |

**Ingestion rules:** drop `contractorRepName` and `contractorRepMobile` at the parser (personal
data). `dlpPeriod` is null throughout — never render a warranty claim from this source alone.

## 2. From BMC ArcGIS and MIDC GIS

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| K8 | Ward + prabhag resolution with the junior engineer's designation | 5 | 2 | 2 | v0.1 | `Prabhaag_Boundary` carries `JR_ENGG`; 227 polygons |
| K9 | Street context — carriageway width, lane count, road status | 3 | 1 | 1 | v0.2 | From `Streets`. Adds factual texture; **not** an ownership claim |
| K10 | **DP 2034 reservation lookup** — "this land is reserved as a 9.15 m DP road; acquisition status: pending" | 4 | 3 | 2 | v0.4 | `DP_KYW` + `DP_MIS`. Explains why a road does not exist, which is often the real answer |
| K11 | MIDC estate containment → route to MIDC, not the municipal ward | 4 | 2 | 2 | v0.3 | 306 industrial-area polygons, statewide |
| K12 | Rail-proximity flag → RTI path instead of a municipal complaint | 3 | 1 | 1 | v0.3 | No public railway cadastre exists; naming that is the feature |
| K13 | **Site-noticeboard capture** — prompt for a photo of the project board when jurisdiction is unresolved | 5 | 2 | 2 | v0.2 | Often the single most reliable artefact on the ground. Feeds OCR extraction of contractor and contract number |

## 3. From CPCB, satellite and weather

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| K14 | **Air-quality context on dust, burning and construction reports** | 4 | 1 | 1 | v0.3 | CPCB feed is keyless and hourly; nearest station, value, timestamp, station name |
| K15 | **Landfill fire watch** via NASA FIRMS around Deonar, Kanjurmarg, Mulund | 3 | 2 | 2 | v0.5 | Detect, then draft — never auto-file |
| K16 | Sentinel-1 SAR monsoon flood extent as a neutral evidentiary layer | 3 | 4 | 2 | v0.6 | The only satellite input that works through Mumbai cloud cover |
| K17 | Rainfall correlation on an issue | 4 | 3 | 3 | v0.4 | **Blocked on licensing**: MESONET is research/academic-scoped, IMD's API is whitelist-gated. Partnership, not scraping |
| K18 | Planned-outage suppression for electricity complaints | 2 | 2 | 1 | v0.6 | Adani publishes a 7-day PDF schedule |

## 4. From representation, finance and audit data

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| K19 | **"Who represents this location"** — corporator, MLA, MP from prabhag and constituency geometry | 4 | 3 | 3 | v0.5 | Live again since the January 2026 BMC election. Source from the State Election Commission record, never from the stale `COUNCILLOR` field in BMC's GIS |
| K20 | **Cost framing from the authority's own rate book** — "this repair is rated at ₹X/m² under BMC USOR <edition>" | 4 | 2 | 2 | v0.5 | Versioned config with an effective-from date and citation, per hard rule 6 |
| K21 | Ward budget vs ward complaint volume, both sourced | 3 | 3 | 3 | v0.6 | BMC budget PDFs plus Praja Foundation's ward analysis, labelled as NGO-derived |
| K22 | **Assembly-question mining** — surface the on-record Vidhan Sabha questions and answers about a ward's civic conditions | 4 | 4 | 2 | v0.7 | `mls.org.in` session PDFs. Underused, and it is the government answering itself on the record |
| K23 | CAG and Local Fund Audit findings attached to a ward or body | 3 | 3 | 2 | v0.7 | Systemic findings, which suits a platform that does not accuse |
| K24 | Peer municipal finance comparison | 2 | 2 | 1 | v1+ | `cityfinance.in` standardised ULB accounts |

## 4A. From candidate and representative disclosure

Scope, permitted forms and the election freeze are set by
[ADR 0013](../04-adr/0013-political-accountability-scope.md). Nothing here is a scorecard.

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| K19a | **Representative resolution** — ward/prabhag/constituency → office holder, sourced from the SEC or ECI record | 4 | 3 | 3 | v0.5 | Prerequisite for escalation routing now that corporators exist |
| K19b | **Verbatim Form 26 record page** — declared cases, assets, education, each next to its source affidavit and retrieval date | 3 | 3 | 4 | v1+ | Pending cases labelled *pending, charges framed on <date>*, never "criminal" |
| K19c | **Commitment register** — verbatim, specific, checkable promises with a closed status vocabulary | 3 | 4 | 4 | v1+ | `Not yet due` / `Evidence found` / `No verified evidence as of <date>`. Never "broken" |
| K19d | **Assembly question mining for a ward** | 4 | 4 | 2 | v0.7 | The government's own on-record answers about local conditions |
| K19e | **Election-period freeze mechanism** | 5 | 2 | 1 | Ships with K19a | A hard product control, not a policy note. Per-constituency, per-phase |
| ⛔ | Credibility or performance score | — | — | 5 | never | Outside the DPDP exemption, unprotected by safe harbour, and the conclusion the platform exists not to draw |
| ⛔ | Party affiliation as a filter, colour or grouping | — | — | 5 | never | Destroys neutrality; already cut in the feature catalog |
| ⛔ | Attributing a civic failure to a named politician | — | — | 5 | never | The platform links contracts to locations and duties to offices, not potholes to people |

## 5. From debarment and procurement registers

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| K25 | Central debarment lookup (CPPP debarred bidders, archive, revocations) | 4 | 2 | 3 | v0.5 | Crawlable; citable URL per entry |
| K26 | Mahatenders debarment list | 4 | 2 | 4 | v0.5 | **Manual retrieval only** — that site's robots.txt forbids automated access |
| K27 | World Bank / ADB debarment cross-check | 2 | 1 | 2 | v1+ | Open registers; rarely relevant to municipal works, occasionally decisive |

## 6. Interoperability

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| K28 | **Open311 GeoReport v2 read-only endpoint** | 4 | 2 | 1 | v0.5 | Emit, never consume. No `POST /requests` — it conflicts with ADR 0007 |
| K29 | WhatsApp intake on the Cloud API | 5 | 3 | 3 | v0.3 | Citizen messages first → opt-in implicit and the session is free. Templates only for breach and status nudges |
| K30 | Prefilled-draft adapters for MyBMC, Swachhata, CPGRAMS, Aaple Sarkar, RailMadad | 5 | 3 | 3 | v0.2 | None of them has a citizen API. Draft + deep link + reference capture is the ceiling, and it is also what ADR 0007 requires |

---

## 7. Proposed and struck off — the data does not exist

| Idea | Why it dies |
|---|---|
| Consume a municipal complaint API | No MMR body publishes one. Verified, not assumed |
| Consume Open311 from an Indian city | No Indian Open311 deployment found |
| Integrate NHAI's Rajmargyatra | Reporting is geo-fenced to a phone physically on the highway — incompatible with deferred reporting |
| Live BEST bus positions | `gtfs.chalobest.in` is dead; no open replacement |
| Mumbai transit GTFS generally | No official feed for suburban rail or metro |
| Pull DLP from BMC's roads API | The field exists and is empty in all 2,405 records |
| Bulk tender data from Mahatenders | `Disallow: /` plus captcha. Manual or RTI only |
| Smart Cities Mission data for Mumbai | Greater Mumbai was never a Smart Cities Mission city |
| IUDX data for Mumbai | Mumbai, Navi Mumbai and Thane are not on the exchange |
| PRS-style MLA report cards for Maharashtra | PRS does not track Maharashtra MLA attendance or questions |
| A live dataset of dug-up roads | BMC's trenching permit system is an internal workflow, not published |
| Free sub-metre satellite imagery | Does not exist for India on open terms |
| Ward-level SECC data | Urban SECC was never fully released |

---

## 8. Preconditions before any of this ships

1. A written terms position for `roads.mcgm.gov.in` and for the MCGM ArcGIS layers.
2. A terms review recorded in the source register for the CPCB feed.
3. IITM's position on MESONET use by a non-academic public platform.
4. Primary documents for every legal constant: the DLP GR, the HC order, the BMC SLA circular.
5. The PII strip on ingestion, tested, before the first roads-API production run.
