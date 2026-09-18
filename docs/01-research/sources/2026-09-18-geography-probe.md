# Source archive — geographies beyond BMC

Probed 18 September 2026. ✅ rows were fetched first-hand; ◐ rows come from the research pass.

---

## 1. Pune Municipal Corporation works GIS

Base `https://iwmsgis.pmc.gov.in/geoserver/ows` (GeoServer WFS). `robots.txt` allows all ◐.

| Layer | Features | Mark |
|---|---|---|
| `pmc:IWMS_line` | 29,070 | ✅ |
| `pmc:IWMS_polygon` | 22,490 | ✅ |
| `pmc:IWMS_point` | 3,164 | ✅ |

Counted with `request=GetFeature&resultType=hits`. Fields on `IWMS_line` ✅:

```
Agency, Budget_Code, Budget_Year, Contact_Number, Created_At, Department, GIS_Created_At, GIS_ID,
Label, LenghtM, Name_of_JE, Name_of_Work, PID, Project_Office, Project_Office_Id, Project_Time,
Remark, Scope_of_Work, Status, Tender_Amount, Update_Date, Used_Amount, Work_Completion_Date,
Work_ID, Work_Type, area, conc_appr_, conceptual, date, department_Id, id, length, length_1,
measure_in, no_of_road, project_fi, stage, village, ward, ward_id, width, zone, zone_id
```

`Name_of_JE` and `Contact_Number` are personal data and are never ingested. ◐ `Agency` is often an
individual's name. In a 200-record sample, `Status` and `Work_Completion_Date` were empty. 9,724
lines belong to the Road department and 12,538 carry an `Agency`. The same server has prabhag,
ward and road-centreline layers (25,485 centrelines). There is no DLP field.

## 2. MMR corporations

| Body | Finding | Mark |
|---|---|---|
| KDMC | `https://maps.kdmc.gov.in/agserver/rest/services`: folders `Hosted, icc, image, Utilities`, 12 top-level services, no login | ✅ |
| KDMC | 10 admin wards, 122 election wards, 4,331 road segments; `KDMC_Projects_Point/Line/Poly` exist with `project_name` and `project_id` but hold 0 records; `copyrightText` empty | ◐ |
| NMMC | Public Power BI works dashboard with ward, budget-head and department totals and work orders vs bills, refreshed 18 Sep 2026; no contractor or geometry. Ward polygons only in a vendor's 2022 ArcGIS Online copy | ◐ |
| Ulhasnagar | ArcGIS Enterprise at `oneulhas.umc.gov.in`, 41 services, no login; city, ward and panel boundaries; road centrelines 2005–2024; `PWD_1..4` tables are yearly totals. Frequent timeouts | ◐ |
| Thane | No open GIS found. Own procurement site `tmc.abcprocure.com` has a captcha-free awarded-tender view with tender number, awardee, amount and date | ◐ |
| Mira-Bhayandar | Map PDFs; about 545 public-works PDFs; a notice on the HC pothole order | ◐ |
| Vasai-Virar | Map as an image; tender and budget PDFs 2013-14 to 2026-27 | ◐ |
| Panvel | PDF maps; a citizen dashboard with 6 AMRUT 2.0 projects, cost and progress | ◐ |
| Bhiwandi-Nizampur | Dashboard 404; works page is prose | ⛔ |

## 3. Other cities ◐

| City | Finding |
|---|---|
| Bengaluru | BBMP work orders on OpenCity: FY22-23 has 8,644 rows with ward, description, contractor and amount; earlier years by ward. Contractor fields carry phone numbers. Bill portals timed out. No court-set pothole SLA found; the pothole PIL is monitored through affidavits |
| Kerala PWD | iROADS returns contractor and DLP start and end per road section, e.g. a section with DLP from 27 Dec 2022 to 27 Dec 2027. Sparse: 1 of 181 and 1 of 633 sections in two sampled tiles. Reached with a key embedded in the portal's own script. Responses include phone numbers |
| Delhi | PWD road ownership as a Google My Maps KML, 867 roads by division; no contractor data |
| Ahmedabad | Open ArcGIS: 48 wards and 38,508 road centrelines with surface type; hosted folder needs a token |
| Hyderabad | Works system login-only; site lists from 2015-16 |
| Chennai | GIS server returned HTTP 500; a road portal lists contractor and work order by date |
| Surat | GIS needs registration |

## 4. Competitor

`https://api.github.com/repos/coding-parrot/pothole-reporter` ✅:

| | |
|---|---|
| Description | "Pothole Reporter is an independent Android app for documenting potholes, garbage and open manholes, with background Drive Mode, Maharashtra-wide handoffs and reviewed routes across supported Indian areas." |
| Licence | MIT |
| Stars | 144 |
| Created | 2026-08-17 |
| Last push | 2026-09-17 |

◐ In Karnataka it drafts an email carrying a probable tender number, which the user sends. Its
documentation states that Maharashtra has no authoritative statewide road-linked award and
defect-liability feed.
