# Source archive — MCGM ArcGIS and CPCB probes

Both probed 24 August 2026 by direct fetch.

---

## 1. MCGM / BMC ArcGIS Online

**Endpoint:** `https://services8.arcgis.com/r6MmJtuWAzMawmJ8/ArcGIS/rest/services?f=json`
**Auth:** none required.
**Licence:** `copyrightText` is **empty on every layer inspected**. Fetchability is not a licence —
a written position from MCGM's IT department is an open item before republishing raw layer content.

**106 services returned.** Full list:

```
AQIRealTime, Acquisition20221004, Assessment_Buildings_With_Sac_view, BED_VACANCY,
BED_Vacancy_View, BMC_Dispensary_WFL1, BMC_Estate, BMC_SWM_Complaints_Data, BMC_Ward,
BMConMaps_Nov26gdb, BaseLayersgdb, Bed_Tracker, Bed_Tracker_API, Bed_Tracker_View,
BuidlingSACs_view, Building_SAC, Building_SAC_view, CPWM_Address, DP_KYW, DP_MIS, DRC2,
Disaster_Management_Resources, Discover_near_me, Discover_near_me_featureLayer,
ElectionGDB_170522, EstateAmenities, Estate_Editor_Info, Estate_PAP, Estate_Property_join,
Estate_SQ, Existing_Suburban_Line, Existing_Suburban_Stations, Facebook, FacilitiesHealth,
Fire_Station, GANESHOSTAV_Mandap_Locations_Updated, Ganesh_Immersion_Locations,
Ganeshotsav_Visarjan_Details_2024_gdb, HBT_Davakhana, Health_Department_Data, Health_Facilities,
Health_Post, IPVS, IPVS_Building_SAC, IPVS_SWM, Insta, Instagram,
Join_Features_to_MCGM_Covid19_Dummy_WFL1___MCGM_Wards, Join_Features_to_Ward_Boundary,
LOI_Category_Table, Licensed_And_Unlicensed_Stalls_Details, MCGM_COVID_BED_TRACKER,
MCGM_COVID_DASHBOARD, MCGM_COVID_DASHBOARD_VIEW, MCGM_Covid19_Dummy_WFL1, MEMSLatest,
MEMSLatestView, MYBMCONMAP1, Main_Mandaps, Mandap_Locations, Maternity_Homes_New_WFL1,
Metro_Lines, Metro_Stations, Mumbai_COVID, Mumbai_COVID_View, MyBMCOnMap_web,
On_Off_street_parking, PAP_MIS, Parking_lots, Police_Stations, Prabhaag_Boundary,
RESPONCESURVEY, RajData_BMC_WFL1, SAC_Buildings, SQ_Master_data, Staff_Quaters_Info_view,
Streets, Survey_Form_Response, Survey_for_Capturing_Hoardings_public,
Sweeping_Beats_And_House_Gullies_Data, Toilet1, Toilets, Twitter, VIPL_view, VLT_Plot_Layer,
VLT_Plots, Vendor_Details_Mapping, Ward_Boundary12, Whatsapp, Yoga_Centers, Youtube,
buid_testttt, parking_Charges, pklt, service_0d99a84be6ca48389fa9ef15eefdb712,
service_4ac33b2f83be40e79ae0aa3d38a9657e, service_8c7e7462ac4f43708d3c8567fd99e2ac_form,
service_b2ebb0fe1e7d4294a9095b344f8a6c3f, survey123_71be1d42068b4d4fac81ed2173bfa8a7,
survey123_71be1d42068b4d4fac81ed2173bfa8a7_form,
survey123_71be1d42068b4d4fac81ed2173bfa8a7_results, survey123_82d8df41025144a7ba07743ef006d66f,
survey123_82d8df41025144a7ba07743ef006d66f_form,
survey123_82d8df41025144a7ba07743ef006d66f_results,
survey123_87714386e417454ca85e8b49c9783147_form, workforce_73ff1765b6674f0598a47228920944a1
```

### Layers inspected in detail

`Prabhaag_Boundary/FeatureServer/0` — polygon, **227 features** (`returnCountOnly=true` confirmed).

```
OBJECTID, POPULATION, WARD, PRABHAG_NO, COUNCILLOR, JR_ENGG, Beat_No, Shape__Area, Shape__Length
```

`Streets/FeatureServer/0` — polyline.

```
OBJECTID, ID, WARD_ID, CARRIAGE_W, NO_OF_LANE, NAME, DRV_STATUS, LKM, Shape__Length
```

**Caution on `COUNCILLOR`:** the field is personal data, and its vintage is unstated. Mumbai had no
elected corporators between March 2022 and January 2026, so any value in it predates the current
council. It must never be rendered without re-verification against the State Election Commission
record.

**Caution on `Streets`:** there is no owner or maintaining-agency attribute. Membership of this layer
is a hint that a road is BMC's, not evidence of it.

---

## 2. CPCB CAAQMS live feed

**Endpoint:** `https://airquality.cpcb.gov.in/caaqms/rss_feed`
**Auth:** none. **Format:** XML. **Size on 24 Aug 2026:** 378 KB.
**Freshness observed:** `lastupdate="24-08-2026 16:00:00"` — same-day, hourly.
**Coverage observed:** `<State id="Maharashtra">` contains **88 stations**, of which **29 carry
Mumbai or Navi Mumbai in the station name** (BMC, MPCB and IITM operated). Each station carries
lat/long and per-pollutant sub-indices (PM2.5, PM10, NO2, NH3, SO2, CO, O3) plus an AQI value.

**Terms:** no terms-of-use statement located for this endpoint. Open item before it becomes a
production dependency.

---

## 3. What was checked and found closed

| Target | Result |
|---|---|
| `roads.mcgm.gov.in/robots.txt` | 404 — no restriction stated |
| `mumbaidp24seven.in` | DNS does not resolve |
| `pmis.mahapwd.gov.in` | Redirects to a login |
| `gtfs.chalobest.in` | DNS does not resolve |
| `city.imd.gov.in/api/v1/cityforecast` | `{"status":false,"message":"Unauthorised Access"}` |
| `api.openaq.org/v3/locations` | HTTP 401 without an API key |
