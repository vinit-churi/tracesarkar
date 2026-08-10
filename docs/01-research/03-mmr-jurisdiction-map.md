# MMR jurisdiction map

The single most valuable thing TraceSarkar can do on day one is answer, correctly and instantly:
**"whose is this?"** This document specifies what that requires.

> ⚠️ Every boundary, ward code, and responsibility split below must be re-verified against the
> primary source before it is encoded. Treat this page as a research index, not as ground truth.

## 1. The bodies

### Municipal corporations (9)

| Corporation | Area covered | Ward scheme |
|---|---|---|
| Brihanmumbai (BMC/MCGM) | Greater Mumbai (island city + suburbs) | 24 lettered administrative wards; separate *prabhag* electoral divisions |
| Thane (TMC) | Thane city | Numbered ward committees |
| Kalyan-Dombivli (KDMC) | Kalyan, Dombivli | Ward committees A–I (verify) |
| Navi Mumbai (NMMC) | Navi Mumbai handed-over nodes | 8 ward offices |
| Ulhasnagar (UMC) | Ulhasnagar | 4 zones |
| Bhiwandi-Nizampur (BNCMC) | Bhiwandi | Prabhag samitis |
| Vasai-Virar (VVCMC) | Vasai, Virar, Nalasopara, Navghar-Manikpur | 9 prabhag samitis (verify) |
| Mira-Bhayandar (MBMC) | Mira Road, Bhayandar | 6 prabhag samitis (verify) |
| Panvel (PMC) | Panvel + former CIDCO colonies (corporation since Oct 2016) | Verify |

### Municipal councils (9)

Ambernath, Kulgaon-Badlapur, Matheran, Karjat, Khopoli, Pen, Uran, Alibaug, Palghar — plus 1,000+
villages across Thane, Raigad, and Palghar districts under Zilla Parishads and Gram Panchayats.

### Parastatals and state bodies

| Body | Owns |
|---|---|
| **MMRDA** | Regional plan, Metro corridors, major flyovers, MTHL, regional roads under its own works wing |
| **MSRDC** | Expressways (Mumbai–Pune), Bandra–Worli Sea Link, coastal/major corridors |
| **PWD (Maharashtra)** | State highways and many arterial roads outside corporation limits |
| **MIDC** | Roads, water, and drainage inside notified industrial estates |
| **CIDCO** | Navi Mumbai nodes not yet handed to NMMC; Panvel-area development |
| **MHADA** | Internal roads and services in MHADA layouts (a notorious grey zone) |
| **Western Railway / Central Railway** | Station premises, FOBs, railway land, level crossings |
| **Mumbai Port Authority** | Port estate |
| **MPCB** | Pollution enforcement (not asset ownership, but the correct escalation target) |
| **Traffic Police** | Signals, signage, enforcement (distinct from road surface ownership) |
| **UMMTA** | Cross-agency transport coordination — the correct mirror target for systemic transport issues |

## 2. The hard cases

These are where naive "point-in-ward" routing produces the wrong answer. The jurisdiction engine
must handle each explicitly.

| Case | Why it's hard | Resolution strategy |
|---|---|---|
| **Road inside city limits but state-owned** | An arterial road inside BMC limits may be PWD or MSRDC. Ward polygon says BMC; asset ownership says otherwise. | Maintain a road-ownership layer that overrides ward containment. Requires an authoritative road inventory (RTI target). |
| **Under a flyover** | Deck = MMRDA/MSRDC. Ground beneath = corporation. Reports must distinguish. | Ask one disambiguating question in the UI when the point is within N metres of a known elevated structure. |
| **Railway land adjacency** | A pothole on the approach road to a station may be corporation, railway, or MRVC. | Geofence station premises; switch the app into "Transit Mode" with RailMadad-compatible fields. |
| **Metro construction corridor** | Active MMRDA/MMRC works often *cause* the defect but the corporation formally owns the surface. | Contract layer: if an active works contract's site polygon contains the point, name both. |
| **CIDCO → NMMC handover** | Handover dates vary by node; the answer changes over time. | Time-versioned jurisdiction: every boundary record carries `valid_from` / `valid_to`. |
| **MHADA layout internal roads** | Often unadopted by the corporation; residents get bounced. | Explicit `disputed` jurisdiction state, with both bodies named and the citizen told this is contested. |
| **Ward vs prabhag** | Electoral prabhags ≠ administrative wards. Datasets conflate them. | Store both; use administrative wards for routing, prabhags for the deliberative/PB features only. |
| **Nallah and creek edges** | Storm-water drains (SWD dept), mangroves (Forest/MCZMA), and creek dumping (MPCB) overlap. | Category-conditional routing: the *issue category* co-determines the authority, not just geometry. |

**Design consequence:** jurisdiction is a function of `(geometry, category, timestamp)` — not of
geometry alone. See [jurisdiction engine](../03-architecture/05-jurisdiction-engine.md).

## 3. Boundary data sources

| Source | What it gives | Notes |
|---|---|---|
| **BMC ArcGIS REST services** — `https://services8.arcgis.com/r6MmJtuWAzMawmJ8/ArcGIS/rest/services` and `https://mybmcid.mcgm.gov.in/server/rest/services` | BMC's own published map layers | Highest-authority BMC source found. Enumerate layers; check for ward, road, drain, and asset layers. Licensing must be confirmed. |
| **DataMeet `Municipal_Spatial_Data`** — `github.com/datameet/Municipal_Spatial_Data/tree/master/Mumbai` | Ward boundaries in GeoJSON (WGS84 / EPSG:4326), CC BY-SA 2.5 IN | Community-maintained; attribution required; vintage must be checked. |
| **`sanjanakrishnan/mumbai_spatial_data`** | MCGM-region GeoJSON incl. census wards, 2017 and 2022 prabhags | Useful for the ward-vs-prabhag distinction. |
| **OpenStreetMap / Overpass** | Road centrelines, `admin_level` relations, named features | Good for road naming and snap-to-road; unreliable for legal ownership. |
| **MMRDA** planning pages | Regional plan and Extended Notified Area extents | Mostly PDF; digitisation needed. |
| **Bhuvan (ISRO) / Survey of India** | Base layers, some administrative boundaries | Licensing varies. |
| **RTI** | Authoritative road-ownership inventory per ward | The only reliable path to the road-ownership layer. Treat as a first-quarter task. |

> **Open gap:** DataMeet issue #53 explicitly requests MMR-wide municipal ward polygons and notes
> only PDF maps are available for several corporations. Outside Greater Mumbai, expect to digitise.

## 4. Ward-level operational data we need per authority

For each authority the platform must store:

```yaml
authority:
  code: BMC
  name: Brihanmumbai Municipal Corporation
  parent: null
  boundary: <geometry>            # time-versioned
  wards: [...]                    # each with boundary, code, name
  departments:                    # category -> department mapping
    pothole: Roads & Traffic
    garbage: Solid Waste Management
    drain: Storm Water Drains
    water: Hydraulic Engineer
  channels:                       # how to actually file
    - {type: app,     name: MyBMC MARG}
    - {type: phone,   value: "1916", hours: 24x7}
    - {type: web,     url: "portal.mcgm.gov.in/..."}
    - {type: open311, url: null}   # if/when available
  sla:                            # per category, in hours
    pothole: 48                   # per Bombay HC directions (Oct 2025)
    garbage: 24
  escalation_ladder:
    - {level: 1, role: Ward Officer}
    - {level: 2, role: Assistant Municipal Commissioner}
    - {level: 3, role: Deputy Municipal Commissioner}
  rti:
    pio: {name: ..., designation: ..., address: ...}
    first_appellate_authority: {...}
```

The `channels`, `sla`, `escalation_ladder`, and `rti` blocks are per-authority research tasks. They
are the least glamorous and most valuable data in the entire project — the accuracy of every
generated instrument depends on them.

## 5. Grievance channels currently in the field

| Channel | Scope | Notes |
|---|---|---|
| **MyBMC MARG** | BMC, 114 grievance types | "Management and Redressal of Grievances"; status tracking |
| **MyBMC 24x7** | BMC, English + Marathi | Information and service delivery |
| **BMC helpline 1916** | BMC | 24×7 |
| **Aaple Sarkar grievance portal** (`grievances.maharashtra.gov.in`) | State departments | Token number; stated 21 working days |
| **CPGRAMS** | Union government departments | Applies to Railways-adjacent complaints |
| **Swachhata (MoHUA)** | All ULBs, sanitation | Huge base; ward-level sanitary inspector routing |
| **RailMadad** (web/app/SMS/139) | Indian Railways | PNR/UTS-linked; 1,000-char limit; no post-submit edits |
| **Per-corporation apps** | TMC, KDMC, NMMC, etc. | Coverage and quality vary widely; needs a per-body survey |

**Implication for TraceSarkar:** for each `(authority, category)` we need a *filing adapter*. Some
will be API-backed, most will be form automation or a "generate + hand off to the citizen" flow. See
[API design](../03-architecture/03-api-design.md) and
[snap-to-action pipeline](../02-product/04-snap-to-action-pipeline.md).

## 6. Confidence model

The engine must never silently guess. Every resolution returns:

```json
{
  "authority": "BMC",
  "ward": "R/S",
  "department": "Roads & Traffic",
  "confidence": 0.62,
  "basis": [
    {"layer": "bmc_ward_boundaries_2023", "match": "contains", "weight": 0.5},
    {"layer": "osm_highway", "match": "nearest_road=Link Road", "weight": 0.2},
    {"layer": "road_ownership_rti_2026Q1", "match": "none", "weight": 0.0}
  ],
  "alternatives": [
    {"authority": "MSRDC", "confidence": 0.31, "reason": "within 30m of flyover deck polygon"}
  ],
  "requires_user_disambiguation": true
}
```

Below a threshold (default 0.75), the app asks the citizen one question rather than routing blindly.
This is the difference between a tool that survives contact with MMR and one that gets a reputation
for wrong-routing in its first week.

## 7. Research tasks

- [ ] Enumerate every layer exposed by the two BMC ArcGIS endpoints; record schema and licensing
- [ ] Verify ward counts and codes for all 9 corporations against official sources
- [ ] Digitise ward polygons for the corporations DataMeet does not cover
- [ ] File RTIs for per-ward road-ownership inventories (BMC first, then TMC/KDMC/VVCMC)
- [ ] Build the elevated-structure polygon layer (flyovers, metro viaducts, sea link)
- [ ] Geofence all suburban railway station premises (WR + CR + Harbour + Trans-Harbour)
- [ ] Compile the `(authority, category) → department + SLA + escalation ladder` table
- [ ] Compile the PIO / First Appellate Authority directory per authority
- [ ] Confirm CIDCO→NMMC handover dates per node
