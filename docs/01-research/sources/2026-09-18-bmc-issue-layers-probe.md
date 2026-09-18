# Source archive — BMC layers and lists for new issue domains

Probed 18 September 2026. Counts marked ✅ were re-fetched first-hand; ◐ rows come from the research
pass. **No personal data was retrieved**: layers that carry it were queried for counts and schema
only, and their contents are not recorded here.

`AGOL` below means `https://services8.arcgis.com/r6MmJtuWAzMawmJ8/ArcGIS/rest/services`, the
endpoint recorded in the [August probe](2026-08-24-mcgm-arcgis-and-cpcb-probe.md).

---

## 1. Incident layer — `AGOL/Disaster_Management_Resources/FeatureServer/1` ✅

Layer name `Incidents`. **70,096 rows**, `INCIDENT_DATE` from 24 Jul 2001 to 17 May 2023.

Fields: `OBJECTID, INUID, INCIDENT_ID, INTEGRATION_ID, INCIDENT_DATE, INCIDENT_CLOSE_DATE, STATUS,
LOCATION, REMARKS, TYPE, SUBTYPE, CREATED_USER, CREATED_DATE, LAST_EDITED_USER, LAST_EDITED_DATE`.

Count by `TYPE` (server-side statistics query, no records pulled):

| Type | Count |
|---|---|
| Tree Fall | 15,368 |
| Water logging | 14,220 |
| Fire | 12,975 |
| Electricity Related | 11,463 |
| Building/ Wall Collapse | 8,534 |
| Oil Spill | 2,317 |
| Chemical Hazard | 1,949 |
| Animal/ Bird Related | 936 |
| Drowning | 878 |
| Road Related | 586 |
| None | 188 |
| Landslide | 183 |
| Falling in Open Gutter/ Drain/Pit | 139 |
| F.O.B. or Bridge | 97 |

◐ Coverage is dense for 2018–2020 and thin after 2021. `STATUS` is "Open" on about 95% of rows, so
this is an incident history, not a record of closure. `REMARKS` and `LOCATION` are free text and
must be screened before display.

## 2. Hoarding survey — `AGOL/Survey_for_Capturing_Hoardings_public/FeatureServer/0` ✅

**1,063 rows.** Fields include `ward, sac_number, address, advt_hoarding_permit_number,
_1_name_of_the_advertiser, _8_advertisement_fee, _7_fee_paid_upto, _11_permit_renewed_upto,
_12_current_renewal_status, _4b_is_it_still_active, _6b_date_time_of_visit`, and
`_3a_name_of_police_team_member_`, **which is personal data and is never ingested**.

◐ Surveyed July 2022 to February 2023. Renewal status: 446 renewed, 78 in court, 77 under appeal.

## 3. Public toilets

| Layer | Count | Mark |
|---|---|---|
| `AGOL/Toilets/FeatureServer/0` | 6,694 points, seat counts by gender | ✅ |
| `AGOL/Toilet1` | 8,412 rows with ward, address, department | ◐ |

◐ Both date from August 2021.

## 4. C1 dilapidated-buildings list, 30 April 2026 ✅

| | |
|---|---|
| URL | `portal.mcgm.gov.in/irj/go/km/docs/documents/HomePage Data/Related Links/C1 list for portal 30.04.2026  final.pdf` |
| SHA-256 | `a9728501ba206e0fc2d46b739147ab6f28a16cf09f401dd518b4c33768c0aca0` |
| Rows | **178** (last serial number), 5 pages |
| Columns | Sr. No., Ward, Beat No., Name of Building, Address |
| Coordinates | None |

The building names and addresses are not reproduced here. C2A, C2B and C3 lists are not published ⛔.

## 5. Related artefacts ◐

| Artefact | SHA-256 | Note |
|---|---|---|
| MHADA pre-monsoon survey press release, 82 most-dangerous cessed buildings (Marathi) | `e6064a2553d545843f293d05f8d56fad65e3b8c1c41e87c01b334d87e38c6496` | Addresses only |
| BMC draft hoarding policy | `a73fdd66ad341598b22c223ff3a1f266d941036a4557f0feda40577ffabf3185` | Structural stability report and insurance per hoarding |
| BMC Environment Status Report 2024-25 (English) | `b66b78afcd1fef3d62f3f253884dab96bad8d009d554208eedcf86e78d367085` | Ward tree counts (Table 8.1), water-sample results (§9), 27-point dust guidelines |
| Praja Foundation, *Status of Civic Issues in Mumbai 2024* | `05b7ba18fb84a9cbe4c72a3a247a0f66054bf58f92f9c105b969cc650e5d200a` | Table 22: BMC complaint volumes for 2023 by RTI. Annex 1: Citizen's Charter escalation days |
| `mybmcid.mcgm.gov.in/server/rest/services/MCGM_UID/Dilapidated` | — | Five-row pilot layer; schema only |

## 6. Layers that expose personal data ◐

Recorded so they are blocklisted, not so they are used. Only counts or schema were requested.

| Layer | Exposure |
|---|---|
| `AGOL/Health_Department_Data` | Patient-level disease line list, Jan–Mar 2023: 1,181 rows with names, phone numbers, diagnoses, coordinates |
| `AGOL/Licensed_And_Unlicensed_Stalls_Details/1` | 1,578 D-ward stalls with vendor names and phone numbers |
| `AGOL/BMC_SWM_Complaints_Data` | Complainant `Mobile_No` |
| `AGOL/Health_Post` | Medical officers' mobile numbers |
| `AGOL/Vendor_Details_Mapping` | 49,733 supplier records with PAN |
