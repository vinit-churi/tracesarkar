# Source archive — works, permits and legal sources

Probed 18 September 2026. ✅ rows were fetched first-hand; ◐ rows come from the research pass.

---

## 1. BMC roads API, re-checked ✅

`https://roads.mcgm.gov.in:3000/api/publicdashboard/` returned HTTP 200 and `count: 2237`, the same
as on 24 August. All 25 records in [the August sample](2026-08-24-bmc-roads-sample-25.json) were
found unchanged, matched on road name, ward and contractor.

| `status` | Count |
|---|---|
| Completed | 1,589 |
| In Progress | 333 |
| Not Started | 233 |
| Deleted | 60 |
| On Hold | 22 |

The [August audit](../08-data-availability-audit.md) §1.1 recorded **58** `Deleted` on the same
list. Two works were re-statused between 24 August and 18 September while the total stayed at
2,237. Which two cannot be recovered, because only a 25-record sample was archived in August.

`ward/` also returned 200.

## 2. BMC storm-water drain desilting API

Base `https://swd.mcgm.gov.in/swdwebapi2026/`. `robots.txt` returns 404 ◐.

`report.svc/report/getprogresscard` (GET) ✅, response verbatim:

```json
{"Data":{"Counter":"4733","Machinary":"--","Manpower":"--","NallahCount":"317544",
"NallahLength":"321335","TargetQuantity":"833298.17","TodayCumulativeTrips":"51941",
"TodayTrips":"7","TotalQuantity":"924335.65"},"Msg":"Successful","ServiceResponse":1}
```

Units are not stated. Other routes ◐:

| Route | Method | Content |
|---|---|---|
| `dashboard/getgraphdata` | GET | Daily weighbridge weights, 76 rows |
| `report/gettargetquantity` | GET | Targets, 107 rows |
| `publicreport/getnallacleaningdataPublic` | POST | Nallah-level cleaning records |
| `report/knowyourwarddetails` | POST | Ward detail |
| `PublicDashboard/getAnomalies` | POST | BMC's own anomaly flags |

Schema seen in the app bundle: `AgencyName, Workcode, Ward, NallahNo, Lat, Long, SlipNo,
VehicleNo`. `VehicleNo` and `SlipNo` are never ingested. Earlier seasons are served under `WMS2025`
and `wms_2022`.

## 3. MahaRERA project map ✅

| | |
|---|---|
| URL | `https://maharera.maharashtra.gov.in/map-projects-search-result` |
| Response | HTML page with the statewide registry embedded as JSON, 21,689,493 bytes |
| SHA-256 | `1f9f7ab128754866b9f83d463b8b7e56703261ae89547e9bfec105a29bb5e371` |
| Rows | 52,529 |
| `robots.txt` | Returns 403 ◐ |

| District | Projects |
|---|---|
| Mumbai City | 1,407 |
| Mumbai Suburban | 6,564 |
| Thane | 7,192 |
| Palghar | 3,078 |
| Pune | 13,692 |

Fields: `CertificateNo, Latitude, Longitude, plotBearing, street, locality, project_Village,
project_Taluka, project_District, project_Division, project_State, pincode`, plus a name field
that holds the promoter's name ◐. The page itself is not committed; only its hash is.

◐ A complaint report gives counts for 7,443 projects. A project-detail API
(`maharerait.maharashtra.gov.in/api/maha-rera-public-view-project-registration-service/…`) exposes
litigation, professionals, documents and estimated vs actual cost, but only by POST, and was not
called.

## 4. Maharashtra GR archive

| | |
|---|---|
| Primary | `gr.maharashtra.gov.in`. Listing is GET; search is captcha-gated; PDFs are directly linkable ◐ |
| Mirror | Internet Archive items `in.gov.maharashtra.gr.<sanketank>` ✅ |
| Count | **172,402** items ✅. ◐ 12,073 from 2026; newest dated 17 Sep 2026 |
| API | `https://archive.org/advancedsearch.php?q=identifier:in.gov.maharashtra.gr.*&output=json` — GET, no key ✅ |

GRs read or located in this pass ◐ unless marked:

| Sanketank | Date | Subject | SHA-256 |
|---|---|---|---|
| `201901141237173518` | 14 Jan 2019 | PWD defect liability periods ✅ — [transcription](2026-09-18-pwd-dlp-gr-2019.md) | `171bb59d09a268e43ed359ca8b3a47d483486c131a8cea2a27d4a50b28eeb4bd` |
| `202506271657286725` | 27 Jun 2025 | UDD: shorter time limits for 25 municipal services | `6f852ecaf1dfdbe6c7c314700a27b0a57120742a9940307dd2d71bcef2422b6b` |
| — | 20 May 2026 | Implementation of the Bombay HC pothole order; monthly reviews | `39f424577d8eedad42c5b8a0886456e185c5b2acd34a7fd12685975530b02d04` |
| `202512291634307907` | 29 Dec 2025 | GAD: how RTI fees are communicated; no amount | — |

## 5. BMC portal sources ◐

| Source | Finding | SHA-256 |
|---|---|---|
| E-tender listing, `portal.mcgm.gov.in/irj/portal/anonymous/qletenders_new` | 352 notices across about 40 departments, 199 distinct Mahatenders tender IDs of the form `2026_MCGM_<n>_<k>`, the rest GeM bids. A plain first-hand fetch returned only a 1,872-byte SAP shell ✅, so the listing needs a browser session | — |
| Government Gazette Extraordinary, Part II No. 11, 19 Jan 2026 | All 227 prabhags with winner, party and valid votes. Text layer glyph-corrupted | `df7771111f020b151605387e498dacd9365b76e083c0d2d918945f20ee405712` |
| Roads USOR 2023 | 71 pages; created 16 Feb 2023; internal title "Roads USOR 2022_16.02.2023", printed heading "Unified Schedule of Rates 2023" | `5b045587c6e71a183e7abf3987b304456d488dda739602195b4c3f10ec92c3d6` |
| AutoDCR citizen search, `autodcr.mcgm.gov.in/CitizenSearch/CitizenSearch.aspx` | No captcha; filters by ward, CTS, SAC and stage; needs an ASP.NET postback. Results carry applicant names | — |
| "Public Disclosure Law" capital-works reports | Contract cost vs budget vs spend per work; mostly Hydraulic divisions; latest Oct–Dec 2025 | — |

## 6. Legal and regulatory ◐

| Source | Finding |
|---|---|
| Bombay HC judgments, `s3://indian-high-court-judgments`, `court=27_1` | Metadata parquet updated 18 Sep 2026. 2026: 1,897 Original Side and 9,775 Appellate Side rows. **No PIL 71/2013 entries for 2025 or 2026**; the set holds judgments and final orders, not interim orders |
| PIL 71/2013 | The 13 Oct 2025 order was passed in **IA No. 29119/2025**. State GRs of 19 Nov 2025, 11 Dec 2025 and 20 May 2026 implement it. A hearing on 29 Jun 2026 is reported by LiveLaw ○ |
| ecMPCB consents, `ecmpcb.in/cms` | About 12,366 pages of 20. Search captcha-gated; `cms/legal_action` access denied; `robots.txt` disallows `/public/pdf`; terms bar reproduction or storage for commercial purposes |

## 7. Checked and closed ⛔

| Source | Finding |
|---|---|
| Standing and Improvements Committee agendas and minutes | Not on the portal; committee pages fall back to the homepage |
| Corporator or ward works lists | Only 2008-09 and 2010-11 PDFs |
| `tendering.mcgm.gov.in` | Does not resolve |
| Bhulekh 7/12 | Captcha |
| MHADA GIS | Login |
| `data.gov.in` | `robots.txt`: `Disallow: /` |
| `bmc.gov.in/public-disclosure/485` | Bhubaneswar Municipal Corporation, not Brihanmumbai |
