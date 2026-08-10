# Glossary

Terms used consistently across this repository. When a term appears in code, the identifier in the
right column is the canonical spelling.

## Domain objects

| Term | Meaning | Identifier |
|---|---|---|
| **Report** | A single citizen submission: photo(s) + location + time + optional text/audio. The atomic unit of input. | `report` |
| **Issue** | A deduplicated, verified civic problem. One issue may aggregate many reports. What gets routed, tracked, and escalated. | `issue` |
| **Asset** | A physical thing with an owner: road segment, footpath, manhole, streetlight, drain, bridge, FOB. | `asset` |
| **Authority** | A body with jurisdiction: BMC, MMRDA, PWD, Western Railway, etc. | `authority` |
| **Jurisdiction** | A resolved (authority, ward, department) tuple with a confidence score and a provenance chain. | `jurisdiction_resolution` |
| **Contract** | A procurement record: tender → award → work order → completion → payment, joined into one lifecycle object. | `contract` |
| **Contractor** | A legal entity that holds contracts. Tracks aliases, directors, and blacklisting history. | `contractor` |
| **DLP** | Defect Liability Period. The warranty window during which a contractor must repair defects free of cost. | `defect_liability_period` |
| **Observation** | A non-citizen data point about a place and time: rainfall, water level, AQI, news mention. | `observation` |
| **Escalation** | A generated legal or administrative instrument (RTI, First Appeal, NGT OA, Lokayukta complaint, consumer complaint). | `escalation` |
| **Filing** | An escalation that a human has actually submitted, with a reference number. | `filing` |
| **Share kit** | The generated set of shareable artefacts for an issue (tweet, WhatsApp card, image, permalink). | `share_kit` |
| **Scorecard** | A contractor's aggregated performance record across all MMR authorities. | `contractor_scorecard` |

## Statuses

### Report

`pending` → `classified` → `merged` (into an issue) | `rejected` (spam/duplicate/unclear)

### Issue

| Status | Meaning |
|---|---|
| `unverified` | Reported once, not yet corroborated |
| `verified` | Passed the trust threshold (see [trust model](../02-product/08-trust-and-antiabuse.md)) |
| `routed` | Filed with the responsible authority; official reference captured |
| `acknowledged` | Authority responded with a ticket ID or acceptance |
| `sla_breached` | Authority SLA elapsed with no resolution |
| `claimed_resolved` | Authority marked it resolved |
| `citizen_confirmed` | A citizen re-photographed and confirmed the fix |
| `reopened` | Re-reported at the same location after `claimed_resolved` |
| `escalated` | At least one escalation instrument generated and filed |
| `closed` | Confirmed resolved and not re-reported for the cool-off window |

`claimed_resolved` and `citizen_confirmed` are deliberately distinct. The gap between them is the
single most important statistic the platform produces.

## Authorities in MMR

| Abbrev | Full name | Typical responsibility |
|---|---|---|
| **BMC / MCGM** | Brihanmumbai Municipal Corporation (Municipal Corporation of Greater Mumbai) | Greater Mumbai roads, drains, SWM, water |
| **MMRDA** | Mumbai Metropolitan Region Development Authority | Regional infrastructure, metro, flyovers, planning |
| **TMC** | Thane Municipal Corporation | Thane city |
| **KDMC** | Kalyan-Dombivli Municipal Corporation | Kalyan, Dombivli |
| **NMMC** | Navi Mumbai Municipal Corporation | Navi Mumbai (post-CIDCO handover areas) |
| **MBMC** | Mira-Bhayandar Municipal Corporation | Mira Road, Bhayandar |
| **VVCMC** | Vasai-Virar City Municipal Corporation | Vasai, Virar, Nalasopara |
| **PMC (Panvel)** | Panvel Municipal Corporation | Panvel (distinct from Pune MC) |
| **UMC** | Ulhasnagar Municipal Corporation | Ulhasnagar |
| **BNCMC** | Bhiwandi-Nizampur City Municipal Corporation | Bhiwandi |
| **PWD** | Public Works Department, Government of Maharashtra | State highways, many arterial roads |
| **MSRDC** | Maharashtra State Road Development Corporation | Expressways, sea link, major corridors |
| **MIDC** | Maharashtra Industrial Development Corporation | Industrial estate roads and services |
| **CIDCO** | City and Industrial Development Corporation | Navi Mumbai areas not yet handed over |
| **MHADA** | Maharashtra Housing and Area Development Authority | MHADA colony internal infrastructure |
| **WR / CR** | Western Railway / Central Railway | Station premises, FOBs, tracks, railway land |
| **UMMTA** | Unified Mumbai Metropolitan Transport Authority | Regional transport coordination |
| **MPCB** | Maharashtra Pollution Control Board | Effluent, air, and noise enforcement |

Wards in Greater Mumbai are lettered (A, B, C, D, E, F/N, F/S, G/N, G/S, H/E, H/W, K/E, K/W, L,
M/E, M/W, N, P/N, P/S, R/C, R/N, R/S, S, T). *Prabhags* are electoral divisions and are **not** the
same as administrative wards — the distinction matters when joining datasets.

## Legal instruments

| Term | Meaning |
|---|---|
| **RTI** | Right to Information application under the RTI Act, 2005. Maharashtra online portal: `rtionline.maharashtra.gov.in`. Under the Maharashtra RTI Rules notified in 2026 the application fee is ₹30 and the first-appeal fee ₹50 — **verify the current fee before generating any instrument** (see [legal framework](../01-research/04-legal-framework.md)). |
| **PIO** | Public Information Officer — the addressee of an RTI. |
| **First Appeal** | Appeal to the First Appellate Authority when the PIO does not reply within 30 days or replies inadequately. |
| **SIC** | State Information Commission — the second appeal forum. |
| **NGT** | National Green Tribunal. MMR falls under the **Western Zone Bench at Pune**. |
| **OA** | Original Application — the NGT's primary filing under §14 of the NGT Act, 2010. |
| **Lokayukta** | Maharashtra Lokayukta and Upa-Lokayuktas Act, 1971 — investigates maladministration and corruption by public servants. |
| **e-Jagriti** | The national consumer-commission e-filing portal (launched Jan 2025; successor to eDaakhil). |
| **CPGRAMS** | Centralised Public Grievance Redress and Monitoring System (Union). |
| **Aaple Sarkar** | Maharashtra state grievance portal; stated redressal window 21 working days. |
| **RTS Act** | Maharashtra Right to Public Services Act, 2015 — notified services with statutory timelines, three levels of appeal, penalty up to ₹5,000 on the erring officer. |
| **DPDP** | Digital Personal Data Protection Act, 2023 + DPDP Rules, 2025 (notified 13 Nov 2025; phased compliance). |

## Technical terms

| Term | Meaning |
|---|---|
| **PostGIS** | PostgreSQL spatial extension. Authoritative for polygon containment and distance queries. |
| **H3** | Uber's hexagonal hierarchical spatial index. Used for fast bucketing, dedup candidates, and heatmap tiles — never as the authority for containment. |
| **Open311 / GeoReport v2** | The international standard for civic issue reporting APIs. TraceSarkar exposes an Open311-compatible read/write surface. |
| **OCDS** | Open Contracting Data Standard. Target normalisation format for the contracts model. |
| **VLM** | Vision-Language Model. Used for classification and evidence extraction from photographs. |
| **Structured outputs** | Schema-constrained model responses (`output_config.format`), so classification results are guaranteed-parseable. |
| **Dedup radius** | Distance within which two reports of the same category are candidates for merging into one issue. Default 40 m for road defects; tunable per category. |
| **Trust score** | Per-account reliability score driving verification thresholds and rate limits. |
| **Share kit** | See domain objects above. |
| **Watchdog TUI** | The terminal user interface for journalists and activists. |

## Conventions

- **Coordinates** are stored as `geography(Point, 4326)`; display is `lat, lon` in that order.
- **Times** are stored in UTC (`timestamptz`) and displayed in `Asia/Kolkata`.
- **Money** is stored in paise as `bigint`, never as float. Display in ₹ lakh/crore.
- **IDs** are UUIDv7 (time-ordered) for all primary keys.
- **Enums** are Postgres native enums, not string columns, for status fields.
- **All external identifiers** (tender IDs, ticket numbers, RTI registration numbers) are stored
  verbatim as text with their source system recorded — never reformatted.
