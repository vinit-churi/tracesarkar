# Vertical exploration — September 2026

Which domains, datasets and places the platform could grow into beyond road defects in BMC, what
public data each one has, and what each would join to. This is a survey, not a commitment: nothing
here enters the [roadmap](../05-delivery/01-roadmap.md) until it is phased.

**Survey date: 18 September 2026.** Provenance marks follow the
[August audit](08-data-availability-audit.md) exactly:

| Mark | Meaning |
|---|---|
| ✅ | Fetched directly during this survey. Endpoint, fields and counts observed first-hand |
| ◐ | Fetched by a research agent from the primary source. Credible, re-verify before building on it |
| ○ | Secondary source only, or inferred. **Do not build on this without verifying** |
| ⛔ | Checked and found absent, closed, or forbidden |

Evidence lives in [`sources/`](sources/README.md), in the four files dated 2026-09-18.

---

## 0. The findings that change the plan

1. **The defect liability period has a primary source, and it is not the one the audit named.**
   PWD GR संकीर्ण-2018/प्र.क्र.151/इमारती-2 of **14 January 2019** sets DLPs of 10 years for rigid
   pavement ≥30 cm, 5 for designed flexible pavement and 20 for bridges ✅. No GR dated 27 April 2017
   was found. The GR binds PWD works, not BMC contracts. See §7 and
   [the transcription](sources/2026-09-18-pwd-dlp-gr-2019.md).
2. **BMC publishes a second works API, for storm-water drains.** `swd.mcgm.gov.in/swdwebapi2026/`
   returns live desilting progress ✅, and its nallah-level records carry agency, work code, ward and
   coordinates ◐. Drains can follow roads as the next attributed vertical.
3. **MahaRERA is a point-to-permit join for construction sites.** One GET returns 52,529 registered
   projects with coordinates and registration numbers, 7,971 of them in Mumbai ✅.
4. **Pune publishes a richer works-to-location join than BMC does.** PMC's works GIS holds 29,070
   work lines, 22,490 polygons and 3,164 points across every department, with tender amount and
   agency ✅.
5. **Several public endpoints expose personal data.** One is a patient-level disease line list.
   They need an ingestion blocklist before any of these sources is touched by code. See §8.
6. **A direct competitor now exists.** Pothole Reporter, an MIT-licensed Android app created on
   17 August 2026, already hands off reports across Maharashtra ✅. See §6.

---

## 1. Method

Three research passes ran in parallel (issue domains, data verticals, geographies). The top claims
were then re-fetched first-hand. Personal tools were designed from the platform's own architecture.

### Scoring

| Axis | 5 means | 1 means |
|---|---|---|
| **Freq** — how often a resident runs into it | Weekly, or every monsoon for everyone | Rare |
| **Data** — best verified state of the key dataset | Geocoded and machine-readable, ✅ | Nothing public |
| **Join** — what a report attaches to | A named permit, contract or party at that point | Nothing below ward level |
| **Cost** — build effort | A week | A quarter |

Rank = Freq × Data × Join, with Cost as the tiebreak.

### Exposure tiers

Legal risk is no longer a reason to leave a feature out of scope
([D044](../00-overview/05-decision-log.md)). Each feature starts in a tier instead:

| Tier | Who sees it | What applies |
|---|---|---|
| **personal** | The maintainers only | Personal-data blocklist (§8) |
| **flagged** | Behind a feature flag, invite-only | Blocklist, plus source and date on every fact |
| **public** | Everyone | Every hard rule in [`CLAUDE.md`](../../CLAUDE.md) |

A feature moves up a tier deliberately, and that move is where the hard rules are checked.

---

## 2. Issue-domain verticals

| # | Vertical | Freq | Data | Join | Cost | Starts | Key evidence |
|---|---|---|---|---|---|---|---|
| V1 | **Drains and flooding** | 5 | 4 | 5 | 2 | public | SWD desilting API ✅; 14,220 geocoded waterlogging incidents ✅ |
| V2 | **Construction sites** | 4 | 5 | 5 | 2 | flagged | MahaRERA, 7,971 Mumbai projects ✅; AutoDCR permission stage ◐ |
| V3 | **Building safety** | 3 | 4 | 4 | 3 | flagged | C1 list of 30 Apr 2026, 178 buildings ✅; 8,534 geocoded collapses ✅ |
| V4 | **Hoardings** | 2 | 4 | 5 | 2 | flagged | 1,063 surveyed hoardings with permit and renewal status ✅ |
| V5 | **Trees** | 3 | 3 | 3 | 3 | public | 15,368 geocoded tree falls ✅; statutory objection window ◐ |
| V6 | **Water supply** | 5 | 2 | 2 | 2 | public | Complaint volumes and closure times by RTI ◐; ward-level only |
| V7 | **Public toilets** | 3 | 4 | 3 | 1 | public | 6,694 mapped toilet blocks ✅, 2021 vintage ◐ |

**Parked:** noise, streetlights and dark spots, hawkers. **Accessibility** stays a subcategory of
`street_furniture` rather than a vertical. **Mosquito breeding** folds into V2, because construction
sites are the joinable source.

Complaint volumes below marked [CC23] are BMC's 2023 figures obtained by RTI and published in Praja
Foundation's *Status of Civic Issues in Mumbai 2024*, Table 22 ◐. Charter days marked [Charter] are
BMC Citizen's Charter first-level escalation days, reproduced in the same report's Annex 1 ◐. BMC's
own charter document was not located, so neither may go into `legal_constants` until it is.

### V1 — Drains and flooding

- **What residents hit:** waterlogging and blocked drains every monsoon. BMC's incident layer
  records 14,220 geocoded waterlogging events from 2001 to 2023 ✅, which answers audit §4.3's
  "no public flooding-spot history".
- **Owner:** BMC Storm Water Drains department. [Charter] "Flooding during monsoon": 1 day.
- **Data:** the SWD desilting API. The progress card is GET and live ✅. Daily weighbridge totals
  and ward targets are GET ◐. Nallah-level records (`AgencyName`, `Workcode`, `Ward`, `NallahNo`,
  `Lat`, `Long`) and BMC's own anomaly flags are POST ◐. Earlier seasons sit at `WMS2025` and
  `wms_2022` ◐.
- **Join:** nallah → work code → agency, which is the same shape as the roads API.
- **Route back:** *"BMC recorded this nallah desilted on <date> by <agency>"*, beside a dated flood
  photograph. That is the K2 pattern applied to drains.
- **Watch for:** `VehicleNo` and `SlipNo` never enter a derivative.

### V2 — Construction sites

- **What residents hit:** dust, debris on the road, night work, blocked footpaths, standing water
  that breeds mosquitoes, trees felled for a project. BMC issued 27-point dust guidelines on
  25 Oct 2023, cited in the Environment Status Report 2024-25 ◐.
- **Data:** MahaRERA's project map embeds the whole statewide registry: 52,529 projects, of which
  Mumbai City 1,407, Mumbai Suburban 6,564, Thane 7,192, Palghar 3,078 ✅. Fields are
  `CertificateNo`, `Latitude`, `Longitude`, `plotBearing` (CTS/CS numbers), locality, district and
  pincode. `projectName` actually holds the promoter's name ◐. A complaint-count report covers
  7,443 projects ◐. AutoDCR's citizen search gives BMC permission stage (IOD, CC, OC) by CTS number,
  but needs a form postback ◐.
- **Join:** report point → nearest registered project → registration number and promoter →
  permission stage by CTS.
- **Route back:** a dust or debris complaint pre-attributed to a registered project. This is the
  action [U6](../02-product/14-civic-utility-features.md) says air quality must lead to.
- **Watch for:** coordinates are declared by the promoter. Absence from AutoDCR is phrased *"no
  matching proposal found as of <date>"*, never as unauthorised.

### V3 — Building safety

- **What residents hit:** partial collapses every monsoon. The incident layer holds 8,534 geocoded
  building or wall collapses ✅. [CC23] 14,572 building complaints.
- **Owner:** ward Building and Factory department; MHADA's repair board for cessed buildings.
- **Standard:** MMC Act s.353B, structural stability certificate once a building turns 30 and every
  10 years after ◐ (bare text via Indian Kanoon, not the gazette).
- **Data:** BMC's C1 dilapidated-buildings list of 30 April 2026, 178 buildings with ward, beat
  number, name and address, no coordinates ✅. MHADA's pre-monsoon list of 82 most-dangerous cessed
  buildings ◐. A five-row pilot layer, `MCGM_UID/Dilapidated`, whose schema (s.354 notice date,
  disconnection, vacated, demolished, court case number) is a ready-made RTI field list ◐. The C2A,
  C2B and C3 lists are not published ⛔.
- **Join:** address → building footprint (SAC number); beat number → prabhag.
- **Route back:** "Is this building listed?" → an RTI for its s.353B audit and status.

### V4 — Hoardings

- **Why it matters:** the Ghatkopar collapse of 13 May 2024 killed 17 people.
- **Data:** `Survey_for_Capturing_Hoardings_public`: 1,063 hoardings with permit number, fee paid up
  to, permit renewed up to and current renewal status ✅, surveyed July 2022 to February 2023 ◐. A
  draft BMC hoarding policy requires a structural stability report and insurance per hoarding ◐.
- **Join:** report point → hoarding → permit → renewal status.
- **Route back:** a report on a hoarding whose recorded permit has lapsed, stated as the two dates.
- **Watch for:** the layer names police team members, which must be stripped. Hoardings on railway
  land are outside BMC.

### V5 — Trees

- **What residents hit:** tree falls every monsoon (15,368 geocoded, 2001–2023 ✅) and felling for
  projects year-round.
- **Standard:** Maharashtra (Urban Areas) Protection and Preservation of Trees Act 1975, s.8: notice
  of felling with at least 7 days for public objections ◐. [Charter] fallen tree: 5 days.
- **Data:** the incident layer ✅; ward tree counts in the Environment Status Report ◐. Tree
  Authority felling notices sit on a portal page that returns an empty shell without a browser ⛔.
- **Route back:** an objection drafted inside the 7-day window, which is a natural
  [deadline wallet](../02-product/14-civic-utility-features.md) clock once the section is read
  from a primary source.

### V6 — Water supply

- **What residents hit:** supply cuts, low pressure and contamination. [CC23] 14,752 complaints.
  Contamination is chartered at 2 days [Charter]; closure took 36 days on average in 2023 [CC23].
- **Data:** ward-level only. Unfit-sample percentages per ward are in the Environment Status Report
  ◐. No pipe network is public ⛔.
- **Route back:** a complaint with the charter clock, and an RTI for the zone's sample results.

### V7 — Public toilets

- **Data:** two BMC layers: `Toilets` (6,694 points with seat counts ✅) and `Toilet1` (8,412 with
  ward, address and department ◐). Both are from August 2021 ◐.
- **Route back:** "nearest toilet" is a utility surface that passes R1 only because a dirty or
  locked block turns into a complaint with the 2-day [Charter] clock.

---

## 3. Datasets that feed several verticals

| Dataset | Access | Mark | Unlocks | Tier |
|---|---|---|---|---|
| **Maharashtra GR archive** and its Internet Archive mirror (172,402 GRs) | `archive.org/advancedsearch.php`, JSON, no key | ✅ | A citation for every `legal_constants` row; a watcher for new PWD, UDD and GAD resolutions | public |
| **BMC incident layer** (70,096 incidents, Jul 2001 – May 2023) | ArcGIS, no auth | ✅ | "N recorded incidents at this spot" for V1, V3, V5 | public |
| **BMC e-tender listing** (352 notices, 199 Mahatenders IDs) | SAP portal page; needs a browser session | ◐ | "Out to tender in this ward"; the tender ID for a manual award lookup. Old notices return 404, so archive on sight | public |
| **2026 BMC election gazette** (227 prabhags) | PDF, glyph-corrupted text | ◐ | Corporator by prabhag from a primary source, replacing the stale GIS field (K19a) | flagged |
| **Roads USOR 2023** | PDF, 71 pages | ◐ | Unblocks K20 and U11 at current rates | public |
| **Bombay HC judgments** (Bombay benches, 2025–26) | Open parquet on S3 | ◐ | "This contractor is a party to case X", as a bare fact | personal |
| **ecMPCB consents** (~247,000 records) | Paged HTML; terms bar commercial reuse | ◐ | Consent status for an industrial source near a report | personal |
| **BMC supplier register** (49,733 records with PAN) | ArcGIS | ◐ | A contractor identity key | personal |

**Dead ends, so nobody proposes them again:** Standing and Improvements Committee agendas and
minutes are not published ⛔. No corporator or ward works lists exist after 2010-11 ⛔. Bhulekh 7/12
is captcha-gated ⛔. MHADA's GIS needs a login ⛔. `data.gov.in` disallows all crawling ⛔.
`tendering.mcgm.gov.in` no longer resolves ⛔. No streetlight asset register or dark-spot list was
found ⛔.

---

## 4. Geographies

The Bombay High Court's pothole directions address municipal corporations and councils generally,
not BMC alone (paras 70(ii) and 70(ix) in
[the transcript](sources/2026-08-24-bombay-hc-pil-71-2013-paras-69-70.md)). The 48-hour clock
already reaches every candidate in Maharashtra below.

### Inside the MMR

| Body | Open GIS | Works joined to location | Maturity |
|---|---|---|---|
| **KDMC** | ArcGIS server, no login: 12 services ✅; admin wards, election wards, 4,331 road segments ◐ | Project layers exist but hold 0 records ◐ | 4 |
| **NMMC** | Ward polygons only in a vendor copy ◐ | ERP dashboard of ward-level works totals, refreshed daily; no rows ◐ | 3 |
| **Ulhasnagar** | ArcGIS Enterprise, 41 services, road centrelines 2005–2024 ◐ | Yearly totals only ◐ | 3 |
| **Thane** | None found ⛔ | A captcha-free awarded-tender list on its own procurement site ◐ | 2 |
| Mira-Bhayandar, Vasai-Virar, Panvel | PDF or image maps ◐ | Tender PDFs; Panvel has 6 projects on a dashboard ◐ | 2 |
| Bhiwandi-Nizampur | None ⛔ | None ⛔ | 1 |

**Order:** KDMC (best maps, and its empty project layers are exactly what to ask for by RTI), then
NMMC (the ERP clearly holds the rows), then Thane (largest population, adjoins BMC, publishes
awards).

### Outside the MMR

| City | Works joined to location | Mark | Notes |
|---|---|---|---|
| **Pune (PMC)** | 29,070 work lines, 22,490 polygons, 3,164 points, all departments | ✅ | `Tender_Amount`, `Agency`, `Budget_Code`, `Work_Type`, `Department`, `ward`. `Status` and completion date sparse ◐; no DLP field |
| Bengaluru | Work orders by ward and text, not geometry; 8,644 rows for FY22-23 | ◐ | No court-set SLA; the 2025 split into five corporations breaks ward joins; Pothole Reporter is already there |
| Kerala PWD | Contractor and DLP start and end per road section | ◐ | The only DLP-by-segment feed found in India, but sparse and state roads only. It was reached with a key embedded in the portal's own script; do not build on it without asking Kerala PWD |
| Ahmedabad | Road centrelines with surface type, no works | ◐ | — |
| Delhi | PWD road ownership by division, no contractors | ◐ | — |
| Hyderabad, Chennai, Surat | Thin or login-gated | ◐/⛔ | — |

**Pune ranks first overall.** Its join is broader than BMC's, it is under the same High Court
order, and [D001](../00-overview/05-decision-log.md) chose the name partly to avoid an MMR ceiling.
Whether a second geography should be Pune or an MMR corporation is a phasing decision.

---

## 5. Personal tools

The personal tier is the cheapest place to start, and most of it is also groundwork for the public
product. Each tool either starts an archive that cannot be backfilled later, builds an evaluation
set the roadmap already requires, or answers an open question.

| # | Tool | What it does | Builds or unblocks | Cost |
|---|---|---|---|---|
| P1 | **Works-API snapshotter** | Daily fetch of the roads and SWD APIs; hash-compare; archive only on change; field-level diff | U14 and K6 change logs; a longitudinal record nobody else is keeping | 1 |
| P2 | **Blocker watchers** | Query the GR mirror and the HC judgments parquet on a schedule; alert on matches | Q2 (RTI fee), Q28 (orders in PIL 71/2013), new SLA resolutions | 1 |
| P3 | **Manual-capture helper** | A bookmarklet that saves the page or PDF you are viewing into the archive with its hash, URL and time. You browse and solve any captcha yourself | The hand-built v0.1 contract dataset, without crawling Mahatenders (D027) | 2 |
| P4 | **Field kit** | Your own capture app for walking R/S ward: photo bursts, GPS, noticeboard frames, a quick label | The 500-photo evaluation set, the 200-point jurisdiction golden set, a noticeboard corpus (K13) | 2 |
| P5 | **RTI and complaint tracker** | Log what you file; run the clocks from `legal_constants`; draft from the templates | An end-to-end check of Q2; dogfoods U1 | 2 |
| P6 | **Research terminal v0** | The Watchdog TUI pointed at a local database, with presets over the works APIs and contracts | Tests the journalist workflow on real data before a public API exists | 3 |
| P7 | **Document search** | OCR and full-text search over everything P1–P3 archive | The DLP extraction corpus (C4) | 3 |
| P8 | **Private alerts** | Air quality near home, fires near the three landfills, a status change on a watched road | Notification copy for U6, K15 and U14 | 1 |

P1 matters most, and the reason showed up during this survey. BMC's roads API was still live on
18 September 2026 with the same total of 2,237 works, and all 25 records sampled on 24 August were
unchanged ✅. But the list now carries **60 works with status `Deleted`, up from 58** on 24 August
✅. Two works changed status in 25 days, and nobody archiving only the current state would know
which. The dataset moves slowly, so a daily hash check is cheap and every change is worth keeping.

---

## 6. Competitor: Pothole Reporter

`github.com/coding-parrot/pothole-reporter` ✅: an Android app for potholes, garbage and open
manholes with a background drive mode and "Maharashtra-wide handoffs". MIT licence, 144 stars,
created 17 August 2026, last pushed 17 September 2026. In Karnataka it drafts an email with a
probable tender number, which a person sends ◐. Its own documentation says Maharashtra has no
authoritative road-linked award and defect-liability feed ◐, so it does not use the BMC roads API.

What it lacks is what this platform is for: contract attribution from the authority's own works
data, the escalation ladder, and a public record of claimed versus confirmed fixes. The MIT licence
allows reuse of its routing work in an AGPL project. Partner or compete is an open question (Q35).

---

## 7. Corrections to earlier documents

Applied in the same change as this survey:

| # | What was wrong | Correct position | Where fixed |
|---|---|---|---|
| 1 | A PWD GR of 27 Apr 2017 sets 15/30/100-year DLPs | No such GR was found. The operative GR is 14 Jan 2019: 10 / 5 / 20 years, PWD works only ✅ | Audit §1.5, §8; Q8; U2 |
| 2 | `bmc.gov.in/public-disclosure/485` lists BMC's PIOs | That domain is Bhubaneswar Municipal Corporation ◐ | Audit §6.3 |
| 3 | No public history of flooding spots | 14,220 geocoded waterlogging incidents in BMC's incident layer ✅ | Audit §4.3 |
| 4 | No BMC USOR newer than 2018 confirmed | Roads USOR 2023 exists ◐ | Doc 14 §11 |
| 5 | Orders after 21 Nov 2025 in PIL 71/2013 unknown | The 13 Oct 2025 order was passed in IA No. 29119/2025. State GRs of 19 Nov 2025, 11 Dec 2025 and 20 May 2026 implement it ◐. The orders themselves are still behind the captcha | Q28 |
| 6 | RTS notified-services list not retrieved | A UDD GR of 27 Jun 2025 shortens limits for 25 municipal services, e.g. water connection 15 → 7 days ◐. Whether it binds BMC is open | Q33 |

**Still open after this pass:** the Maharashtra RTI fee (Q2). A GAD circular of 29 Dec 2025 changes
how the fee is communicated but sets no amount ◐.

---

## 8. Personal-data blocklist

These fields and layers are public today. None may be ingested into a derivative at any tier, and
none of their content is kept in this repository.

| Source | Layer or field | What it exposes |
|---|---|---|
| BMC ArcGIS | `Health_Department_Data` | **Patient-level disease records**: names, phone numbers, diagnoses, coordinates ◐. Only aggregate counts were queried |
| BMC ArcGIS | `Licensed_And_Unlicensed_Stalls_Details` | Vendor names and phone numbers ◐ |
| BMC ArcGIS | `BMC_SWM_Complaints_Data` | Complainant `Mobile_No` ◐ |
| BMC ArcGIS | `Health_Post` | Medical officers' mobile numbers ◐ |
| BMC ArcGIS | `Survey_for_Capturing_Hoardings_public` | `_3a_name_of_police_team_member_` ✅ |
| BMC ArcGIS | `Vendor_Details_Mapping` | PAN, which is personal data for individual suppliers ◐ |
| BMC ArcGIS | Incident layer `REMARKS`, `LOCATION` | Free text; screen before any display |
| BMC SWD API | `VehicleNo`, `SlipNo` | Vehicle registration numbers ◐ |
| PMC works GIS | `Name_of_JE`, `Contact_Number`; `Agency` is often a person | Officer names and phone numbers ✅ |
| Kerala iROADS | Section detail | Contractor and officer phone numbers ◐ |
| Bengaluru work orders | Contractor field | Phone numbers joined to names ◐ |
| AutoDCR | Search results | Applicant names ◐ |
| BMC RTI s.4 manuals | Staff lists | Named staff with gross salaries ◐ |

Whether to tell BMC about the health layer is an open decision (§10).

---

## 9. Terms, as far as they were checked

| Source | robots.txt | Terms |
|---|---|---|
| `gr.maharashtra.gov.in` | Allows `/` ◐ | Search is captcha-gated; direct PDF links are not |
| MahaRERA | Returns 403, unread ◐ | Unknown |
| `swd.mcgm.gov.in` | 404 ◐ | None located |
| `portal.mcgm.gov.in` | Present ◐ | None located, as for the roads API |
| ecMPCB | Disallows `/public/pdf` ◐ | Bars reproducing or storing content for commercial purposes ◐ |
| PMC works GIS | Allows all ◐ | None located |
| `data.gov.in` | `Disallow: /` ⛔ | — |

Every candidate is registered in [`data/sources.yaml`](../../data/sources.yaml) with
`status: candidate` and its terms fields left null until reviewed.

---

## 10. Still open

| # | Question | Gates |
|---|---|---|
| Q32 | Terms for MahaRERA's map page and BMC's SWD API | V1, V2 at the public tier |
| Q33 | Does the UDD GR of 27 Jun 2025 bind BMC's service limits? | U9 |
| Q34 | Should the second geography be Pune or an MMR corporation? | Phasing |
| Q35 | Partner with Pothole Reporter, or compete? | Phasing |
| — | Should BMC be told its health layer is public? | Nothing; a maintainer decision |

---

## 11. Next step

Phase these into the roadmap: pick which verticals, which geography and which personal tools enter
which milestone, and fold in the drift the September review found between the roadmap, backlog and
source register and the August audit.
