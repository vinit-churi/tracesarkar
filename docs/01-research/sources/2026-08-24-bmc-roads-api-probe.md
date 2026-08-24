# Source archive — BMC roads dashboard API probe

**Endpoint base:** `https://roads.mcgm.gov.in:3000/api/`
**Discovered from:** the Angular bundle at `roads.mcgm.gov.in/publicdashboard/main.*.js`, which
carries `production:!0, baseUrl:"https://roads.mcgm.gov.in:3000/api/"`.
**Probed:** 24 August 2026. No authentication, no key, no captcha. `robots.txt` returns 404.
**Terms of use:** not located. Recorded as an open item before any production ingestion.

## Reproduce

```
curl "https://roads.mcgm.gov.in:3000/api/publicdashboard/"
curl "https://roads.mcgm.gov.in:3000/api/publicdashboard/getpublicdashboardwithstartedafter01oct2025roadlayer"
curl "https://roads.mcgm.gov.in:3000/api/geo/getwardlayer"
curl "https://roads.mcgm.gov.in:3000/api/ward/"
```

## Observed on 24 Aug 2026

| Endpoint | Shape | Size |
|---|---|---|
| `publicdashboard/` | `{status, message, count, data[]}` | count = **2237** |
| `…roadlayer` | GeoJSON `FeatureCollection`, `MultiLineString` | **2405** features |
| `geo/getwardlayer` | GeoJSON, `wardname` | 24 ward polygons |
| `ward/` | ward master: `wardID`, `wardName`, `zoneName`, `Zone` | 24 rows |

Other routes present in the bundle, not exercised: `geo/getroadlayer`, `geo/getwardlayernew`,
`geo/getelectrolwardlayer`, `geo/getphase2locationroadlayer/{id}`, `geo/getthaneboundarylayer`,
`geo/getmasticplantlayer`, `location/potholework`, `location/masticworklist`, `report/cumulative/`,
`potholework/create`, `login/`, `user/`, `syncfleet/*`, `vts/*`, `scada/*`.

## `publicdashboard/` record shape

```json
{
  "contractorName": "M/s NCC Ltd. W-421",
  "roadName": "Road from Nanasaheb Dharmadhikari Marg to Sant Dnyaneshwar Mandir Marg",
  "length": 400,
  "width": 18.3,
  "status": "Completed",
  "ward": "H/E",
  "source": "Phase 1",
  "completedRaodImagePath": "LocationCompletedRoads/ND to SD_completed_1752831187649.jpg",
  "startDate": "2023-10-01T00:00:00.000Z",
  "endDate": "2026-01-20T00:00:00.000Z"
}
```

## Geometry-layer `properties.location` — the useful object

Full field list (61 fields):

```
__v, _id, billPeriodFromDate, billPeriodToDate, billpaidAmountAfterContractorsRebate,
completionDateOfPQC, contractorName, contractorRepMobile, contractorRepName, coordinates,
createdBy, createdOn, description, dlpPeriod, endDate, excavationPlannedEndDate,
excavationPlannedStartDate, hasOctoberEntry, layingOfDuctPlannedEndDate,
layingOfDuctPlannedStartDate, length, lengthUnit, locationID, locationName, microPlanEndDate,
microPlanStartDate, modifiedBy, modifiedOn, pdfFileName, pdfFilePath, pqcPercentProgress,
pqcPercentStatus, pqcPlannedEndDate, pqcPlannedStartDate, pqcStatusmodifiedBy,
pqcStatusmodifiedOn, pqcstatus, qmaRepMobile, qmaRepName, quater1EndDate, quater2EndDate,
quater3EndDate, quater4EndDate, quater5EndDate, remarks, roadDiagrams, roadType, rwmsLength,
startDate, status, subRoadType, swdPlannedEndDate, swdPlannedStartDate,
trafficNOCApplicationDate, trafficNOCReceivedDate, wardName, width, widthUnit, workCode,
workOrderDate, zoneName
```

Sample (trimmed to the fields that matter):

```json
{
  "workCode": "W-414",
  "locationName": " Hotel Orritel West to Maurya House (C.T.S no 650/3 )",
  "contractorName": "M/s Megha Engineering and Infrastructures Ltd. W-414",
  "length": 120,
  "lengthUnit": "643ce04d81fa1b2af060df9c",
  "width": 9.15,
  "widthUnit": "643ce04d81fa1b2af060df9c",
  "dlpPeriod": null,
  "startDate": "2023-04-01T00:00:00.000Z",
  "endDate": "2026-04-10T00:00:00.000Z",
  "status": "Completed",
  "roadType": "Mega CC Road",
  "pqcstatus": "Yes",
  "pqcPercentProgress": 100,
  "completionDateOfPQC": null,
  "quater1EndDate": "2024-03-01T18:30:00.000Z",
  "excavationPlannedStartDate": "2023-04-01T00:00:00.000Z",
  "trafficNOCReceivedDate": null,
  "microPlanStartDate": "2025-10-05T00:00:00.000Z"
}
```

## Distributions across the 2405 geometry features

**Status:** Completed 1709, In Progress 392, Not Started 271, On Hold 31, None 2

**Work codes (10 distinct):** C-320, C-322, E-289, E-322, W-414, W-415, W-421, W-445, W-446, W-447

**Contractors (10 distinct):**

| Contractor | Segments |
|---|---|
| M/s. GHV (India) Pvt Ltd. | 359 |
| M/s. Gawar Construction Ltd. | 317 |
| M/s. BSCPL Infrastructure Ltd. | 286 |
| M/s N.C.C Ltd | 275 |
| M/s. RPS Infraprojects Pvt. Ltd. | 261 |
| M/s. Eagle Infra India Pvt. Ltd. E-289. | 209 |
| M/s NCC Ltd. W-421 | 209 |
| M/s. AIC Infrastructures Pvt. Ltd. | 178 |
| M/s Megha Engineering and Infrastructures Ltd. W-414 | 177 |
| M/s Dineshchandra Ramchandra Agrawal Infracon Pvt. Ltd. W-415. | 132 |

## Two findings that constrain use

1. **`dlpPeriod` is null in all 2405 records.** The field exists in the schema and carries no
   value. Defect liability period is still unavailable from this source.
2. **`contractorRepName` and `contractorRepMobile` are populated with the personal names and mobile
   numbers of named individuals.** They must be dropped at the parser and must never reach a
   derivative, a log, or an export.
