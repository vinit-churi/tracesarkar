# Data availability audit — August 2026

A field survey of what public data actually exists, for whom, in what shape, and under what terms.
This complements [`05-data-sources.md`](05-data-sources.md), which is the *register* of sources we
intend to use. This document is the *evidence* behind it.

**Audit date: 24 August 2026.** Every row carries a provenance mark:

| Mark | Meaning |
|---|---|
| ✅ | Fetched directly during this audit. Endpoint, fields and counts observed first-hand |
| ◐ | Fetched by a research agent from the primary source. Credible, re-verify before building on it |
| ○ | Secondary source only, or inferred. **Do not build on this without verifying** |
| ⛔ | Checked and found absent, closed, or forbidden |

---

## 0. The three findings that change the plan

1. **BMC publishes an unauthenticated JSON API containing 2,237 road works with contractor names,
   ward, dates, status and completion photographs — plus 2,405 road geometries keyed by work code.**
   This is a large part of the contract-to-geometry problem, already solved and published by the
   authority itself, for the Mumbai CC-road programme. See §1.1. It materially de-risks the
   [v0.2 thesis gate](../05-delivery/01-roadmap.md).
2. **Mumbai has had an elected municipal council again since January 2026.** The repository's
   assumption that BMC is administrator-run is stale. Every routing, escalation and
   "who represents this ward" surface has a corporator layer again. See §5.1.
3. **The Bombay High Court order was read in full, and it gives the product a claim forum it did
   not know it had.** The 48-hour rule, the ₹6,00,000 death and ₹50,000–₹2,50,000 injury figures are
   all verified verbatim — and the order routes compensation claims to a **Municipal Commissioner +
   DLSA Committee** that may act on "information from any source". See §6.0.

---

## 1. Procurement and works data

### 1.1 BMC roads dashboard API ✅ — the single most valuable source found

BMC's public road-works dashboard at `roads.mcgm.gov.in/publicdashboard/` is an Angular SPA backed
by an **open, unauthenticated JSON API**:

```
Base: https://roads.mcgm.gov.in:3000/api/
```

Verified endpoints (fetched 24 Aug 2026):

| Endpoint | Response | Observed |
|---|---|---|
| `publicdashboard/` | JSON | `count: 2237` road works. Fields: `contractorName`, `roadName`, `length`, `width`, `status`, `ward`, `source` (Phase 1/2), `completedRaodImagePath`, `startDate`, `endDate` |
| `publicdashboard/getpublicdashboardwithstartedafter01oct2025roadlayer` | GeoJSON | **2,405 features**, `MultiLineString` geometry, each with a nested `location` object |
| `geo/getwardlayer` | GeoJSON | 24 ward boundary polygons (`wardname`), CRS84 |
| `ward/` | JSON | Ward master: `wardID`, `wardName`, `zoneName`, `Zone` |

Other routes present in the bundle (not exercised): `geo/getroadlayer`, `geo/getwardlayernew`,
`geo/getelectrolwardlayer`, `geo/getphase2locationroadlayer/{id}`, `geo/getthaneboundarylayer`,
`geo/getmasticplantlayer`, `location/potholework`, `location/masticworklist`, `report/cumulative/`,
`potholework/create`, `login/`, `user/`. The `location/*` write routes and `login/` indicate this is
the public face of BMC's internal road works management system.

**The geometry layer's `location` object is the important part.** Observed fields:

```
workCode          e.g. "W-414"           ← joins geometry to contract package
locationName      "Hotel Orritel West to Maurya House (C.T.S no 650/3)"
contractorName    "M/s Megha Engineering and Infrastructures Ltd. W-414"
length, width     120, 9.15  (metres)
roadType          "Mega CC Road" | "Mega CC Phase 2"
status            Completed | In Progress | Not Started | On Hold
startDate, endDate
pqcstatus, pqcPercentProgress, completionDateOfPQC
quater1EndDate … quater5EndDate
excavationPlanned{Start,End}Date, layingOfDuctPlanned*, swdPlanned*, pqcPlanned*
trafficNOCApplicationDate, trafficNOCReceivedDate
microPlan{Start,End}Date
dlpPeriod         ← present in the schema but NULL in all 2,405 records
contractorRepName, contractorRepMobile   ← personal data, see the warning below
roadDiagrams, remarks
```

Observed distributions across the 2,405 geometry features:

- `roadType`: Mega CC Phase 2 = 1,676; Mega CC Road = 727
- `status`: Completed 1,709 · In Progress 392 · Not Started 271 · On Hold 31
- **10 distinct contractors** across **11 work codes** (`W-414`, `W-415`, `W-421`, `W-446`, `W-447`,
  `E-289`, `E-322`, `C-320`, `C-322`, …)
- On the `publicdashboard/` list: 24 wards, statuses including 58 `Deleted`

**What this means for the product**

- The join TraceSarkar was built to make — location → contract package → contractor → dates — is
  *already published for BMC's CC-road programme*. For a pothole on a concretised road, attribution
  is a spatial query against a public GeoJSON layer, not a research project.
- `workCode` is the key to Mahatenders' award records: the contract package identifier appears in
  both the dashboard's `contractorName` string and the `workCode` field.
- The completion photographs (`completedRaodImagePath`, served from `roads.mcgm.gov.in:3000/`) are
  BMC's own claim of completion. A citizen photograph of a defect at the same coordinates, dated
  after that claim, is an extremely strong evidentiary pairing — and is exactly the
  claimed-vs-confirmed contrast the platform exists to publish.

**Warnings**

1. `dlpPeriod` is **null in every record**. The defect liability period remains unavailable from
   this source. It still has to come from contract documents or an RTI. Do not report otherwise.
2. `contractorRepName` and `contractorRepMobile` are **personal contact details of named
   individuals**. They must never be ingested into a public derivative. Drop them at the parser.
3. Coverage is the Mega CC-road programme only (~2.2k roads), not Mumbai's ~2,000 km of roads in
   general, and not asphalt works, footpaths, drains or any other category.
4. Terms of use for `roads.mcgm.gov.in` were not located. `robots.txt` returns 404 (no restriction
   stated). The API is served by BMC to its own public dashboard, so consumption is technically
   equivalent to using the dashboard — but the terms question is open and belongs in the source
   register before ingestion starts.
5. Port 3000 on a `.gov.in` host with `login/` and write routes is a fragile arrangement. Archive
   every response; assume it can disappear without notice.

### 1.2 Mahatenders (GePNIC, Maharashtra) ◐

- Award data is reachable: `ResultOfTenders` gives AOC date, e-published date, tender title and
  reference, organisation chain, and a link to the AOC document. A **Debarment List** is published.
- **`robots.txt` is `Disallow: /` for all agents**, and the AOC search is captcha-gated.
- The Disclaimer/copyright policy permits linking freely but requires permission to reproduce
  content.
- **Conclusion: no automated collection.** Under [hard rule 8](../../CLAUDE.md), this source is
  link-and-cite only, with manual retrieval for the v0.1 hand-built dataset, and an RTI or a written
  permission request as the route to anything larger. Record the decision in the source register.

### 1.3 CPP Portal `eprocure.gov.in` ◐

Central-government tenders only — Railways, port, CPWD. Not the state or municipal works that make
up most MMR civic infrastructure. Its `robots.txt` blocks only admin paths, and it publishes a
**Debarred Bidders search plus archive and revocation lists**, which is a genuinely useful
blacklisting source with a citable URL.

### 1.4 Everything else ◐/⛔

| Source | Finding |
|---|---|
| GeM `view_contracts` | Captcha-gated; no bulk export ⛔ |
| `portal.mcgm.gov.in` | Legacy SAP portal. Individual tender PDFs are directly linkable, no index, no API ◐ |
| MahaPWD PMIS | `pmis.mahapwd.gov.in` redirects to a login. Not public ⛔ |
| MMRDA / MSRDC / MHADA / CIDCO | Tender notices as PDFs; bidding routed through Mahatenders ◐ |
| data.gov.in | No Maharashtra procurement dataset. GODL-India licence would be workable if one existed ◐ |
| OCDS in India | CivicDataLab publishes Assam and Himachal Pradesh under CC BY 4.0. **No Maharashtra footprint** — but a precedent to cite when asking for one ◐ |
| Commercial aggregators (BidAssist, TenderTiger…) | Terms prohibit automated access. Licensed purchase is the only legitimate route ⛔ |

### 1.5 Defect liability periods — still the critical gap

> **Correction, 18 September 2026.** No GR dated 27 April 2017 was found. The operative resolution
> is PWD GR संकीर्ण-2018/प्र.क्र.151/इमारती-2 of **14 January 2019** ✅: 10 years for rigid pavement
> ≥30 cm, 5 years for designed flexible pavement, 20 years for bridges. It binds PWD works, not BMC
> contracts. See
> [vertical exploration §7](09-vertical-exploration.md#7-corrections-to-earlier-documents) and the
> [transcription](sources/2026-09-18-pwd-dlp-gr-2019.md).

- ~~○ A Maharashtra PWD GR dated 27 April 2017 is *reported* to set 15-year DLP for bituminous
  roads, 30 years for cement concrete, and up to 100 years for bridges. **Sourced only from press
  coverage.** The GR itself must be retrieved from `gr.maharashtra.gov.in` before any of these
  numbers goes into `legal_constants`.~~ The press figures describe expected life, not DLP.
- ◐ MoRTH's national EPC standard is 5 years flexible / 10 years rigid pavement, reportedly doubled
  for EPC contracts in 2024.
- ✅ BMC's own works API carries a `dlpPeriod` field that is empty.
- Retention money percentages for Maharashtra PWD contracts are **not verified**; the commonly cited
  5% is generic industry practice, not a sourced Maharashtra figure.

**This remains the highest-value extraction the platform performs, and it still requires contract
PDFs.**

---

## 2. Geospatial and jurisdiction

### 2.1 BMC ArcGIS — open, rich, unlicensed ✅

`https://services8.arcgis.com/r6MmJtuWAzMawmJ8/ArcGIS/rest/services?f=json` returns **106 hosted
feature services**, no token required. Verified directly:

| Layer | Verified detail |
|---|---|
| `BMC_Ward` | 24 admin ward polygons ◐ |
| `Prabhaag_Boundary` | ✅ **227 polygons**; fields `WARD`, `PRABHAG_NO`, `COUNCILLOR`, `JR_ENGG`, `Beat_No`, `POPULATION` |
| `Streets` | ✅ polyline; fields `WARD_ID`, `CARRIAGE_W`, `NO_OF_LANE`, `NAME`, `DRV_STATUS`, `LKM` (~56.8k segments ◐) |
| `DP_KYW` | DP-2034 reservations, ~5,902 records: `RES_CODE`, `RES_MAINTYPE`, `DISCRIPTION` ◐ |
| `DP_MIS` | Land acquisition case file: `NAME_OF_OWNER`, `ACQUISITION_STATUS_OF_LAND`, `COMPENSATION_AMOUNT` ◐ |
| `Assessment_Buildings_With_Sac_view` | ~305k building footprints with address, floors, usage ◐ |
| `BaseLayersgdb` | MCGM_Street, Railway_Lines, Waterbody, Prabhag_Bnd, Ward, Zone ◐ |
| Others | `BMC_Estate`, `Metro_Lines`, `Metro_Stations`, `Existing_Suburban_Line`, `Fire_Station`, `Police_Stations`, `Toilets`, `On_Off_street_parking` ◐ |

Two cautions:

- **`copyrightText` is empty on every layer.** Fetchability is not a licence. Republishing raw MCGM
  layer content needs a written clarification from MCGM's IT department. Using it internally for
  containment is a different and much lower risk than redistributing it.
- **`Prabhaag_Boundary.COUNCILLOR` is a personal-data field of unknown vintage**, almost certainly
  predating the January 2026 election. Never display a councillor name from this layer without
  re-verification against the State Election Commission's own record.

Also confirmed: `mybmcid.mcgm.gov.in/server/rest/services` (property/UID services) ◐, and MIDC's
statewide `gis.midcindia.org` service with 306 industrial-area polygons, edited as recently as
February 2026 ◐.

### 2.2 The unsolved problem: road ownership

**No dataset — state or central — carries a "maintaining agency" attribute on a road geometry.** ◐
BMC's `Streets` layer has no owner field; membership is a hint, not proof.

The practical resolution ladder is: ward/prabhag containment (PostGIS against BMC layers) → DP 2034
reservation and acquisition status → MIDC estate containment → OSM `ref=NH*/MH SH*` and `operator=*`
as a secondary signal → railway-line proximity as an RTI trigger → **site noticeboard photographed
by the citizen**, which is often the most reliable single artefact → RTI to the ward office.

That last-but-one step deserves emphasis: NHAI, MSRDC and MMRDA erect project boards with contract
and contractor details at work sites. A citizen photograph of that board is a first-class evidentiary
artefact and a *feature*, not a fallback.

### 2.3 Boundaries, geocoding and imagery ◐

| Source | Licence | Verdict |
|---|---|---|
| DataMeet `Municipal_Spatial_Data`, `PincodeBoundary` | CC BY-SA 2.5 IN | Usable; vintage varies |
| data.gov.in all-India pincode boundary GeoJSON (2026) | GODL-India | Usable, commercial use permitted |
| OSM / Overpass / Nominatim | ODbL | Usable with share-alike care. Self-host for production |
| Photon | ODbL | Lighter self-hosted geocoder |
| Mappls (MapMyIndia) | Commercial, free tier | Best Indian address parsing; quota unverified |
| Copernicus Sentinel-2 | Free and open | Usable; 10 m is too coarse for road defects |
| Bhuvan | Bhuvan ToU, login for bulk | Not open by default |
| Survey of India Nakshe | **Restricted** — no export, no commercialisation | ⛔ unusable as a base layer |
| Esri World Imagery | Display-only under Esri MLA | Basemap tile only, never redistributed |
| Google Maps/Earth imagery | Personal/non-commercial only | ⛔ needs a paid Maps Platform licence |
| MRSAC / MahaGIS | Licence unstated | Verify before use |
| CIDCO, SRA, MMRDA, MSRDC, PWD GIS | Internal, PDF or DWG | ⛔ RTI or manual digitisation |
| CRZ / CZMP (MCZMA) | PDF map sheets only | Shapefiles exist but are not published |

---

## 3. Grievance platforms and interoperability ◐

**No MMR civic body exposes a citizen-facing complaint API.** That is now a verified finding, not an
assumption.

| Platform | API | Verdict |
|---|---|---|
| MyBMC / MARG / 24×7 / Pothole FixIt | None found | Deep link + prefilled draft only |
| BMC WhatsApp chatbots (`MyBMC Assist` 8999228999; garbage 8169681697) | n/a | Citizen-operated channel to surface, not automate |
| Swachhata (MoHUA) | None found | Deep link |
| CPGRAMS | **Partner-only** — integration exists for 17 states/UTs and 4 ministries, not private platforms | Prefilled draft, citizen submits |
| Aaple Sarkar | None citizen-facing | Draft + deep link. 21 working days stated |
| RailMadad | None found | Draft + deep link |
| NHAI Rajmargyatra | **Geo-fenced to a phone physically on the highway** — structurally incompatible with deferred reporting | Cite, do not integrate |
| MSRDC / MMRDA / MSEDCL / Adani / Tata / BEST / traffic police | None | Surface the channel to the citizen |
| DIGIT / eGov PGR | Open-source, self-hostable, real API | No Maharashtra deployment found. Good schema reference |
| IUDX | Live for Pune, Pimpri-Chinchwad, Nagpur, Kalyan-Dombivli. **Not Mumbai, Navi Mumbai or Thane** | Not usable in MMR core today |
| Open311 GeoReport v2 | Spec, not a service. **No Indian deployment found** | Emit it, do not consume it |

**Open311 decision:** implement the **read** side only. Open311's `POST /requests` lets a third
party create a request with only an API key, which is incompatible with
[ADR 0007](../04-adr/0007-never-auto-file.md) and with the platform's human-taps-to-submit rule. A
read-only GeoReport v2 surface costs little and buys compatibility with existing civic tooling. Map
`status` to open/closed only, and never expose anything but coarsened coordinates and redacted media
through it.

**Bhashini:** a free ULCA key is obtainable, but the published terms state the APIs are **for
proof-of-concept use only**; production use requires arrangement with the Bhashini team. This is a
blocking dependency for v0.4 voice, exactly as
[the accessibility doc](../02-product/10-accessibility-and-vernacular.md) §9 anticipated. Treat
`bhashini.ai` (a commercial-looking key-sale site) as unverified and unaffiliated until proven
otherwise.

**WhatsApp Cloud API:** location and image messages arrive natively over webhooks. Billing moved to
**per-message** for template messages on 1 July 2025; replies inside the 24-hour session window are
free. Design intake so the citizen messages first — that makes opt-in implicit and the exchange free
— and reserve templates for SLA-breach and status nudges. ○ Per-message India rates come from
secondary sources; price from Meta's own rate card before budgeting.

---

## 4. Environment, weather and context

### 4.1 Air quality — open and keyless ✅

`https://airquality.cpcb.gov.in/caaqms/rss_feed` returns the **national CAAQMS live feed as XML with
no key and no authentication**. Verified 24 Aug 2026: 378 KB, `lastupdate="24-08-2026 16:00:00"`,
**88 Maharashtra stations of which 29 are named for Mumbai or Navi Mumbai** (BMC, MPCB and IITM
operated), each with lat/long and per-pollutant sub-indices (PM2.5, PM10, NO2, NH3, SO2, CO, O3)
plus AQI.

This is the cheapest high-quality context feed available to the platform. Hourly cadence. No terms
of use statement was located for the endpoint itself — record that gap in the source register before
it becomes a production dependency.

SAFAR (10 Mumbai stations) has no API ◐. MPCB's real-time page largely duplicates stations already
in the CPCB feed ◐. OpenAQ v3 requires an API key ◐ and adds nothing over CPCB for MMR.

### 4.2 Rainfall — available, but the licence is wrong ◐

| Source | Finding |
|---|---|
| IITM Mumbai MESONET | 131 gauges, 15-minute cadence. Undocumented PHP endpoints (`get_data.php`) return `##`-delimited text, not JSON. **The site scopes access to "research and academic purpose"** — a public civic platform is neither. Needs a written terms determination from IITM before use |
| BMC AWS (120 stations) | `dm.mcgm.gov.in/auto-weather-station` is a client-rendered SPA; no static endpoint found by curl. Worth a browser network-trace pass |
| IMD | `api.imd.gov.in` documents ~28 endpoints, but a live call returned `Unauthorised Access`. Access is by IP whitelist / institutional arrangement, **not public self-serve** |
| IMD gridded rainfall | 0.25° daily NetCDF, 1901–2024. Free, and far too coarse for intra-city work |
| data.gov.in rainfall datasets | CSV/JSON behind a free `api.data.gov.in` key |

**Consequence:** high-resolution live rainfall for MMR is *technically* reachable and *legally*
unsettled. Treat rainfall context as a partnership dependency (IITM or BMC DM), not a scraping
target.

### 4.3 Flooding ◐

`mumbaiflood.in` (IIT-B) and iFLOWS-Mumbai both exist and both are dashboards. **Neither publishes a
citizen-accessible feed.** iFLOWS output is described as going to IMD-Mumbai and BMC's disaster
management department. BRIMSTOWAD documents and BMC's annual pre-monsoon chronic-flooding spot lists
have no stable public URL — RTI or a direct request to BMC's Storm Water Drains department.

> **Update, 18 September 2026.** A spot *history* is public: BMC's ArcGIS incident layer holds
> 14,220 geocoded waterlogging incidents from 2001 to 2023 ✅, and BMC's storm-water drain desilting
> API is open ✅. See [vertical exploration](09-vertical-exploration.md#v1--drains-and-flooding).

### 4.4 Satellite ◐

| Source | Use |
|---|---|
| Sentinel-2 (10 m, free, open) | Optical change detection on large sites |
| **Sentinel-1 SAR (down to 5 m, free)** | All-weather, day-night — the only viable satellite flood-mapping input during a Mumbai monsoon, when optical is cloud-blocked |
| NASA FIRMS | Free key, <3 h latency active-fire hotspots. Directly useful for dumping-ground fires (Deonar, Kanjurmarg, Mulund) |
| Bhuvan/NRSC | AWiFS 56 m, LISS-III 23.5 m, flood hazard zonation; free after registration |
| Sub-metre imagery | ⛔ Not free for India. Google and Esri basemap terms forbid derivative redistribution |

### 4.5 Utilities, mobility and context ◐

| Source | Finding |
|---|---|
| Adani Electricity | Publishes a **7-day planned-outage schedule as a dated PDF**. Useful for suppressing "no power" reports that coincide with a disclosed outage |
| MSEDCL, Tata Power, MGL | No structured feed found |
| BMC trenching permits | A G2B permit *workflow* exists (~6,000 applications/year). ⛔ No public dataset of currently dug-up roads |
| BEST GTFS (`gtfs.chalobest.in`) | ⛔ **Dead** — DNS no longer resolves |
| Mumbai suburban rail, Metro | ⛔ No official GTFS. Community conversions only |
| Open Transit Data | ⛔ Delhi only. No Mumbai equivalent |
| Waze Connected Citizens | Free but restricted to bodies that manage transport infrastructure; a private platform likely needs a government sponsor |
| Railway megablock notices | Press and social channels only |
| Census 2011 ward data + ward boundary KML (OpenCity) | **Public domain**, Open Definition conformant |
| BMC Environment Status Report | Annual PDF on `mcgm.gov.in`, 2020-21 through 2024-25 confirmed |
| Praja Foundation ward budget and civic-issue reports | Annual PDF. Cite as NGO-derived analysis, distinct from official BMC data |


---

## 5. Representation, finance and audit ◐

### 5.1 Mumbai has an elected council again — the repo's assumption is stale

| Event | Date | Note |
|---|---|---|
| Corporators' term ended; administrator appointed | 7 Mar 2022 | The state the repo currently assumes |
| Supreme Court ordered local body elections by 31 Jan 2026 | 16 Sep 2025 | Delimitation by 31 Oct 2025 |
| Municipal councils / nagar panchayats polled | 2 Dec 2025 | Results 21 Dec 2025 |
| **29 municipal corporations including BMC polled** | **15 Jan 2026** | Results 16 Jan 2026; 227 seats; turnout ~52.9% |
| Mayor of Mumbai took office | 7 Feb 2026 | ○ Reported as Ritu Tawde (BJP) |

○ Party-wise seat totals come from secondary sources and **must** be replaced with the Maharashtra
State Election Commission's own record (`mahasec.maharashtra.gov.in`, results via `mahasecelec.in`)
before anything is displayed. A cleaned CSV exists at `data.opencity.in/dataset/bmc-election-results-2026`
◐ — useful, but the SEC gazette is the citable source.

**Consequences for TraceSarkar**

- Ward pages, escalation ladders and "who is responsible" surfaces need a **corporator layer** that
  the docs currently do not model.
- `Prabhaag_Boundary.COUNCILLOR` in BMC's ArcGIS is almost certainly pre-2022 and must not be
  trusted.
- Delimitation moved roughly a quarter of ward boundaries; **boundary vintage now matters for
  historical issues**, which is feature B6 (time-versioned boundaries) arriving earlier than planned.

### 5.2 Money and cost norms ◐

| Source | Use |
|---|---|
| BMC Unified Schedule of Rates (`portal.mcgm.gov.in`, USOR PDFs) | Per-item civic work rates. Lets the platform state "this repair is rated at ₹X/m² under USOR <edition>" with a citation. ~~○ Latest confirmed edition is 2018 — verify for a newer one~~ ◐ Roads USOR 2023 found 18 Sep 2026 — see [vertical exploration §7](09-vertical-exploration.md#7-corrections-to-earlier-documents) |
| Maharashtra PWD State Schedule of Rates | ◐ Latest confirmed 2022-23 |
| BMC budget documents (Budget A/B/G) | PDF only, per year; historical ward-level breakdowns exist |
| `cityfinance.in` | Standardised audited ULB accounts for 4,000+ bodies — peer comparison |
| MPLADS (`mplads.mospi.gov.in`, data.gov.in) | MP fund allocation and utilisation, open extracts |
| MLA-LAD Maharashtra | ⛔ No public dataset found |
| CAG local-body reports; DLFAA Maharashtra | PDF performance and compliance audits. Systemic findings, rarely naming contractors — which suits the platform's non-accusatory posture exactly |
| Praja Foundation annual reports | Ward and category complaint volumes derived from BMC's own CCRS. Good benchmark, PDF only |
| Swachh Survekshan, MoHUA EoLI/MPI | City scores. ○ Latest confirmed EoLI/MPI round is 2020 |
| Smart Cities Mission portal | ⛔ **Greater Mumbai was never a Smart Cities Mission city.** Thane, Kalyan-Dombivli, Pune, PCMC, Nagpur, Nashik, Solapur, Aurangabad only |
| Maharashtra Vidhan Sabha starred/unstarred questions (`mls.org.in`) | Rich, underused: MLA questions frequently extract ward-level civic statistics from the government on the record. PDF per session |
| PRS Legislative Research | ⛔ Does **not** track individual Maharashtra MLA attendance or questions — only post-election profiles |

### 5.3 Candidate and representative disclosure ◐

| Source | Finding |
|---|---|
| **ECI affidavit portal** `affidavit.eci.gov.in` | Form 26 affidavits searchable by election, state, constituency and candidate. **Scanned images, no API, no bulk export.** Footer disclaims the listing as tentative and points to the Returning Officer |
| Form 26 contents | Pending criminal cases and convictions; assets and liabilities of candidate, spouse and dependants; education; PAN and income-tax particulars for five years; government dues; sources of income. Statutory basis: RPA 1951 s.33A + Conduct of Election Rules 1961 r.4A |
| ADR / MyNeta | Parsed affidavit data, Lok Sabha 2004–2024, state assemblies, some local bodies — reportedly including BMC 2026 candidates. No public API or bulk export found. Reuse terms not established |
| PRS Legislative Research | **CC BY 4.0.** Strong for Parliament; ⛔ no per-MLA attendance or question tracking for Maharashtra |
| Maharashtra Vidhan Sabha questions | Session-wise starred and unstarred question lists as PDFs. Underused and genuinely valuable |
| MPLADS | Fund allocation and utilisation; open extracts via data.gov.in |
| MLA-LAD Maharashtra | ⛔ No public dataset found |
| Party manifestos | Self-published by parties. ⛔ No ECI archive found. Filing a manifesto with the ECI is not a verification route |

**The legal position on publishing any of it is settled in one direction and unsettled in another,
and the split matters enough that it is written up as
[ADR 0013](../04-adr/0013-political-accountability-scope.md).** In brief: verbatim republication of
Form 26 fields with a source and a date is well precedented and squarely within the DPDP Act's
publicly-available-data exemption (s.3(c)(ii), confirmed against the Act's text). Anything the
platform *derives* — a score, a "promise broken" label, a link between a named person and a specific
civic failure — is the platform's own speech, outside that exemption, unprotected by intermediary
safe harbour, and exposed under a defamation provision where truth alone is not enough without
public good. India has neither an actual-malice standard for public figures nor an anti-SLAPP
statute, so the practical risk is a civil suit that runs for years regardless of merit.

Election periods add a separate regime: RPA s.126's 48-hour silence window, which the Election
Commission treats as covering websites and social media, plus an unresolved question about whether
civic content counts as a political advertisement requiring MCMC pre-certification.

---

## 6. Legal and corporate

### 6.0 The Bombay High Court pothole order — verified against the primary text ✅

**The repository's premise holds.** The order in `PIL 71/2013`, *High Court on its own motion v.
State of Maharashtra*, uploaded 13 October 2025, was retrieved and read directly during this audit.
Every figure the repository relies on is in the operative directions at paragraph 70. Verbatim:

| Direction | Text (condensed, quoting the order) |
|---|---|
| (i) Compensation | "In cases of death caused by potholes or open manholes, a sum of **Rs.6,00,000/-** shall be paid to the legal heirs… In cases of injury, compensation ranging from **Rs.50,000/- to Rs.2,50,000/-**, depending upon the nature and gravity of the injury." Payable by "the Municipal Corporations, MMRDA, MSRDC, MHADA, BPT, NHAI, and the PWD, as the case may be", and "independent of, and in addition to" other remedies |
| (ii) Forum | A **Committee** per jurisdiction: Municipal Commissioner + Secretary, District Legal Services Authority (DLSA) inside a corporation; Chief Officer + DLSA in a council; District Collector + DLSA outside municipal limits; Principal Secretary / Chairperson / CEO + DLSA for MMRDA, MSRDC, PWD, BPT, NHAI |
| (iii) Cadence | First meeting **within 7 days** of receiving information; thereafter **at least once every 15 days**, "more particularly, during the monsoon period" |
| (iv) Trigger | The Committee "may act **suo motu** or on an application… It may also take cognizance **on receipt of information from any source, including newspaper reports**." The officer in charge of the police station must communicate any such incident to the Committee **within forty-eight hours** |
| (v)–(vii) Recovery | The authority pays first; the amount is then recovered, after inquiry, "from the officers, engineers, or contractors found responsible" |
| (viii) Sanction | "Strict disciplinary and penal action… Such action shall include **blacklisting**, imposition of penalties, and initiation of appropriate departmental or criminal proceedings" |
| (ix) **The 48-hour rule** | "**All potholes, once brought to the notice of the concerned Corporation or Authority, shall be attended to forthwith and, in any event, within forty-eight hours.** Failure to do so shall constitute **gross negligence** and shall warrant departmental action against the responsible officers and contractors" |
| (x) Payment clock | Compensation disbursed **within six to eight weeks** of the claim; delay makes the Municipal Commissioner / Chief Officer / Collector / CEO / Chairperson / Principal Secretary **personally responsible**, and the amount then carries **interest at 9% per annum** from the date of claim |
| (xi) Publicity | The State must give "wide publicity" to these directions |

Paragraph 69 also records the Court's expectation that "roads are constructed and maintained in
such a manner that they do not require repairs for a **minimum period of five to ten years**", and
directs authorities to bear that in mind when awarding contracts. Compliance reports were called for
on 21 November 2025 — **later orders in this PIL very likely exist and must be tracked.**

A note on method: an earlier automated pass reported that these figures were *absent* from the
order. They are not. The order spells the number as "forty-eight hours", and a literal search for
"48 hours" misses it. Anything derived from an automated text search of a judgment gets a
human re-read before it changes a legal constant.

**Four product consequences, all new:**

1. **The claim forum is a named Committee, not a writ petition.** The compensation assistant (D9)
   should generate a claim to the *Municipal Commissioner + DLSA Secretary* Committee, not a High
   Court filing. That is a much lower barrier for a citizen and a materially better product.
2. **The Committee may act on "information from any source, including newspaper reports."** A
   sourced, timestamped, geotagged TraceSarkar record is precisely such information. This is a
   legitimate, court-sanctioned route for platform-assembled evidence — with a human filing it.
3. **There is a second clock.** Beyond the 48-hour attendance rule, a filed claim has a 6–8 week
   disbursal deadline after which 9% interest accrues and named officials become personally
   responsible. Both belong in `legal_constants` and both belong on the issue timeline.
4. **"Gross negligence" is the Court's characterisation of a missed 48-hour deadline, not ours.**
   The platform may quote it with the citation. It still never says it in its own voice.

### 6.1 The 24-hour BMC directive alongside the 48-hour order ◐

Two different numbers are circulating for pothole attendance:

- The Bombay High Court's 48-hour requirement (October 2025), which the repository currently treats
  as authoritative.
- A **24-hour** BMC monsoon directive reported in 2025.

The 48-hour figure is now verified against the order itself (§6.0). The 24-hour figure is still only
press-sourced; get the BMC circular by number before using it. If both hold, the platform shows the
**shorter applicable** clock and names the body that imposed each, with both citations.

### 6.2 Corporate identity and debarment ◐

| Source | Finding |
|---|---|
| MCA V3 master data | CIN, status, registered office, directors and DINs, capital — free web lookup. Filed documents cost ₹100 per company-year |
| data.gov.in "Company Master Data" | Bulk CSV of ~3.6M companies under GODL-India. The practical route to lineage work |
| API Setu | A directory of government APIs; per-API access is negotiated, not open |
| GST public search | Free manual lookup of GSTIN → legal name, status, address. The developer portal appears GSP-gated |
| Udyam | Free verification by URN / PAN / Aadhaar |
| **CPPP debarment list** (GFR 2017 Rule 151) | Live, public, searchable: bidder, category, debarment start and end. Central procuring entities only |
| **GeM suspended / debarred sellers** | Live page plus a PDF archive; three categories — overall suspension, temporary moratorium, buyer-specific debarment |
| World Bank / ADB debarment | Open lists, cross-debarment noted. ADB publishes only a public subset |
| BMC / PWD blacklisting | ⛔ **No standing public register.** Announced ad hoc by circular and press note. Each instance must be tracked individually with its own citation |

**There is no single Maharashtra-wide blacklist.** Any "blacklisted" badge in the product is a claim
about a *specific register on a specific date*, and must render that way or not at all.

### 6.3 Courts, tribunals and RTI machinery ◐

| Source | Finding |
|---|---|
| eCourts / Bombay HC case status | Free, manual, captcha-gated. ⛔ **No official public API** — every "eCourts API" found is a commercial scraper |
| NJDG | Public dashboard; its Open API is scoped to government departments |
| Supreme Court eSCR (`digiscr.sci.gov.in`) | Free full-text search of ~34,000 judgments |
| Indian Kanoon | Free to browse. Paid API (~₹5 per 100-result search page); free credits for verified non-commercial use. **Mandatory "Powered by IKanoon" attribution** on rendered results |
| CivicDataLab Justice Hub / Open Justice | Open datasets: ~81M eCourts case records, High Court judgments 1950–2025 in JSON/Parquet. The best bulk legal corpus available |
| NGT | E-filing and judgment search. Note a TLS certificate mismatch on the bare domain |
| Maharashtra SIC | Bench-wise RTI decision listings |
| e-Jagriti | Consumer commission filing and case status |
| RTI Online Maharashtra | State RTI filing and first appeal |
| ~~**BMC public disclosure (RTI s.4(1)(b))**~~ | ~~`bmc.gov.in/public-disclosure/485` — includes a downloadable list of all BMC PIOs.~~ **Withdrawn 18 Sep 2026:** `bmc.gov.in` is Bhubaneswar Municipal Corporation. Mumbai's PIO list must come from `portal.mcgm.gov.in` |
| MMRDA RTI page | SPIO contact only; no consolidated PIO/FAA directory found |

### 6.4 Legal constants — verified, disputed, and unverified

Only the first block may be encoded as-is. Everything else needs a primary document first.

**Verified against the primary text ✅** (all: Bombay HC, PIL 71/2013, order of 13 Oct 2025)

| Constant | Value |
|---|---|
| `pothole_attendance_deadline` | 48 hours from notice to the Corporation or Authority |
| `pothole_death_compensation` | ₹6,00,000 |
| `pothole_injury_compensation_range` | ₹50,000 – ₹2,50,000 by gravity |
| `compensation_disbursal_window` | 6–8 weeks from receipt of claim |
| `compensation_delay_interest` | 9% per annum from date of claim |
| `committee_first_meeting` | 7 days from receipt of information |
| `committee_meeting_interval` | 15 days |
| `police_report_to_committee` | 48 hours |
| `expected_road_life_no_repair` | 5–10 years (paragraph 69, expectation not mandate) |

**Well-established, secondary-sourced ◐ — verify against the bare Act before encoding**

RTI response 30 days; life-and-liberty 48 hours; via APIO 35 days; third-party 40 days; first appeal
window 30 days; second appeal window 90 days; PIO penalty ₹250/day capped at ₹25,000; IT Rules
takedown on a court or government order 36 hours; grievance acknowledgement 24 hours and resolution
15 days.

**Disputed or unverified ⛔ — do not encode**

| Constant | Problem |
|---|---|
| Maharashtra RTI fee | ₹10 is long-standing; a commercial site claims ₹30 under "2026 rules". **No gazette notification found.** The first-appeal fee is equally unclear |
| NGT limitation (6 months, 30-day appeal) | Not confirmed against the NGT Act this pass |
| Consumer commission pecuniary limits and 2-year limitation | 2019 Act values not confirmed as current |
| DPDP s.3(c)(ii) publicly-available-data exemption | The clause could not be retrieved from a primary source. **This is load-bearing for publishing director names** — it needs a proper legal read, not a web fetch |
| BNS s.356 defamation | Section number and exceptions from a secondary reference site, not the gazette |
| Maharashtra RTS Act notified-services list and SLAs | Portal renders dynamically; not retrievable this pass |

### 6.5 What may lawfully be published about a named contractor ◐

Publishable, each with source URL and retrieval timestamp: corporate register facts (CIN, status,
registered office, directors, capital); GSTIN status and legal name from the official search;
debarment status **as recorded in a named register on a named date**; contract award facts; and case
status as a bare fact ("X Pvt Ltd is respondent in NGT Appeal No. Y").

The defence that matters is BNS s.356 Exception 1 — **truth published for the public good** — which
works only if every fact is sourced and true. Exceptions 2 and 3 give more latitude for comment on a
*public servant's* conduct in office than for a *private contractor's* reputation. The platform's
rule against ever concluding wrongdoing is therefore doing real legal work, not signalling caution.

Two operational consequences: a designated **Grievance Officer** and a documented intake process
must exist **before** launch, not after the first notice; and because most debarment sources are
point-in-time PDFs rather than live registers, "last verified on <date>" is load-bearing text, not
decoration.

---

## 7. What this audit changes

| # | Change | Where |
|---|---|---|
| 1 | Add BMC roads API as a first-class source; build the ingester early | Source register, [tender engine](../03-architecture/06-tender-engine.md) |
| 2 | Re-scope the v0.2 contract-to-geometry gate: for CC roads it is a spatial join against published data, not a geocoding research problem | [Roadmap](../05-delivery/01-roadmap.md) |
| 3 | Add a corporator/representative layer to the data model | [Data model](../03-architecture/02-data-model.md) |
| 4 | Mark Mahatenders "link and cite only, no automated collection" | Source register |
| 5 | Add `contractorRepMobile`-style PII stripping as an ingestion rule | [Ingestion](../03-architecture/07-ingestion-and-scrapers.md) |
| 6 | Encode the nine verified constants from the HC order; rebuild D9 around the DLSA Committee route, not a High Court filing | `legal_constants`, [feature catalog](../02-product/02-feature-catalog.md) D9 |
| 7 | Open311: read-side only, explicitly | [ADR 0010](../04-adr/0010-open311-compatibility.md) |
| 8 | Bhashini's PoC-only terms become a v0.4 blocking item | Accessibility doc §9 |
| 9 | Site-noticeboard photography becomes a designed capture affordance | [Screen spec](../02-product/11-screen-spec.md) |
| 10 | Boundary vintage matters now: delimitation moved ~25% of ward boundaries in 2025 | Jurisdiction engine |

---

## 8. Still open

- Terms of use for `roads.mcgm.gov.in` and `portal.mcgm.gov.in` — neither located.
- A written licence position from MCGM for the ArcGIS layers.
- ~~The 27 April 2017 PWD GR on defect liability periods, by GR number.~~ Answered 18 Sep 2026:
  the GR is dated 14 January 2019 — see §1.5.
- The BMC 24-hour pothole circular, by number and date.
- **Orders after 21 November 2025 in PIL 71/2013** — compliance reports were called for on that
  date, so the directions may have been extended or modified.
- The DPDP Act's publicly-available-data exemption, read properly by a lawyer.
- Maharashtra RTI fee, by gazette notification.
- SEC's structured ward-to-corporator mapping for the 2026 result.
- ~~Whether a newer BMC USOR edition supersedes 2018.~~ Answered 18 Sep 2026: Roads USOR 2023 ◐.
- Whether Mahatenders' non-AOC searches are also captcha-gated.
- Bhashini production licensing terms and cost.
