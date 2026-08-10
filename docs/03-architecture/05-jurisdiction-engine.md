# Jurisdiction engine

**Answers:** *given a point, a category, and a time — which authority, which ward, which department,
and how confident are we?*

This is the component that decides whether TraceSarkar is trusted or dismissed. A citizen who sees a
wrong ward once will assume everything else is wrong too.

---

## 1. The core insight

Jurisdiction is **not** a function of geometry alone:

```
jurisdiction = f(point, category, timestamp)
```

- A pothole and a mangrove complaint at the same coordinate go to different bodies.
- A point on a flyover and the same point on the road below go to different bodies.
- A point in a Navi Mumbai node resolves differently before and after the CIDCO→NMMC handover date.

Any engine that treats it as `f(point)` will be wrong often enough to lose credibility.

---

## 2. Resolution algorithm

```
resolve(point, accuracy_m, category, at_time) →

  1. CANDIDATE GATHERING  (all layers queried in parallel)
     a. ward containment           PostGIS ST_Contains on wards valid at `at_time`
     b. authority containment      same, on authority_boundaries
     c. road ownership             ST_DWithin(point, road.geom, road.buffer_m + accuracy_m)
     d. elevated structures        ST_DWithin(point, structure.geom, 30 m)
     e. railway premises           ST_Contains
     f. active contract sites      ST_DWithin(point, site.geom, 25 m) on live contracts
     g. asset proximity            nearest known asset within 20 m

  2. SCORING
     each candidate authority accumulates weighted evidence (see §3)

  3. OVERRIDES
     road_ownership beats ward containment when confidence ≥ 0.85
     railway_premises containment beats everything for transit categories
     category → department mapping applied per authority

  4. AMBIGUITY DETECTION
     if (top_score - second_score) < 0.15  → ambiguous
     if elevated structure within 30 m     → ambiguous (deck vs ground)
     if top_score < 0.75                   → low confidence

  5. RESULT
     { authority, ward, department, confidence, basis[], alternatives[],
       requires_user_disambiguation }
```

Steps 1a–1g run concurrently (`errgroup`) against the read replica. Target p95 latency: **< 120 ms**.

---

## 3. Scoring weights

Initial weights. These are configuration, tuned against a labelled evaluation set — not constants
in code.

| Signal | Weight | Notes |
|---|---|---|
| Ward polygon containment | 0.50 | The baseline |
| Road-ownership match | 0.85 | **Overrides** ward when present and confident |
| Railway premises containment | 0.90 | For `transit` category; 0.40 otherwise (a pothole outside a station is usually municipal) |
| Elevated structure proximity | −0.30 to top / +0.30 to structure owner | Triggers disambiguation rather than deciding |
| Active contract site containment | 0.20 | Corroborating, not deciding |
| Asset ownership (known asset within 20 m) | 0.70 | Strong when the asset inventory exists |
| Category-department mapping exists for this authority | 0.10 | A body with no department for this category is a weak candidate |
| GPS accuracy penalty | ×(1 − min(accuracy_m/100, 0.5)) | A 90 m fix cannot decide a 20 m boundary question |

---

## 4. Handling the hard cases

### Flyover vs road below

```
if elevated_structure within 30 m of point:
    ask user: [ photo of deck ] "On the flyover"  |  [ photo of ground ] "Underneath"
    → deck    ⇒ structure.authority_id (MMRDA / MSRDC)
    → ground  ⇒ structure.ground_authority_id (corporation)
```

One binary question with two thumbnails. Never a dropdown.

### State highway inside city limits

The `road_ownership` layer is the only correct solution, and it must come from an authoritative
source — realistically an RTI-obtained road inventory per ward. Until it exists for a ward, the
engine caps confidence at 0.70 for `road_defect` on roads it cannot classify, which routes those
reports through the disambiguation question.

**This is the single biggest data-acquisition dependency in the project.** See
[roadmap](../05-delivery/01-roadmap.md).

### Railway adjacency

Station premises are geofenced. Inside the fence:
- The UI switches to Transit Mode (RailMadad-compatible fields: station code, platform, UTS/PNR
  optional).
- Category options change (FOB, escalator, coach cleanliness, platform surface).
- Outside the fence but within 100 m, the engine flags `near_railway` and offers both channels.

### MHADA layouts and unadopted roads

Resolution returns `disputed` with **both** authorities named:

```json
{
  "status": "disputed",
  "candidates": [
    {"authority": "BMC",   "confidence": 0.44, "reason": "ward containment"},
    {"authority": "MHADA", "confidence": 0.41, "reason": "layout polygon containment"}
  ],
  "message": "Responsibility for this road is contested between BMC and MHADA. We will file with both."
}
```

Naming the dispute is itself a product feature — it converts the citizen's confusion into a
documented institutional failure, and both bodies receive the complaint.

### Time versioning

Every boundary table carries `valid_from` / `valid_to`. Resolution always passes `at_time`. A 2021
issue resolves against 2021 boundaries. This matters for historical analysis and for any filing that
concerns a past period.

---

## 5. Category → department mapping

```sql
SELECT department, sla_hours, sla_source, filing_channel, pio
FROM authority_departments
WHERE authority_id = $1 AND category = $2;
```

Examples (illustrative — must be verified per authority):

| Authority | Category | Department |
|---|---|---|
| BMC | `road_defect` | Roads & Traffic |
| BMC | `waste` | Solid Waste Management |
| BMC | `water_drainage` (drain) | Storm Water Drains |
| BMC | `water_drainage` (supply) | Hydraulic Engineer |
| BMC | `structural` | Building & Factory / Bridges |
| MMRDA | `road_defect` | Transport & Communication |
| WR/CR | `transit` | Divisional Railway Manager |
| — | `environment` | MPCB (regulator) **plus** the land-owning authority |

`environment` is special: the correct escalation target is often a regulator, not the land owner. The
engine returns both, and the escalation engine files with both.

---

## 6. Output contract

```jsonc
{
  "authority": {"code": "BMC", "name": "Brihanmumbai Municipal Corporation"},
  "ward": {"code": "R/S", "name": "Kandivali West"},
  "department": "Roads & Traffic",
  "sla_hours": 48,
  "sla_source": "bombay_hc_2025-10-13",
  "confidence": 0.62,
  "status": "ambiguous",             // resolved | ambiguous | disputed | unresolved
  "basis": [
    {"layer": "wards", "version": "bmc_2023", "match": "contains", "weight": 0.50},
    {"layer": "osm_highway", "match": "nearest=Link Road (12 m)", "weight": 0.10},
    {"layer": "road_ownership", "match": "none", "weight": 0.00},
    {"layer": "elevated_structures", "match": "Kandivali flyover, 18 m", "weight": -0.30}
  ],
  "alternatives": [
    {"authority": "MSRDC", "confidence": 0.31, "reason": "within 30 m of flyover deck"}
  ],
  "requires_user_disambiguation": true,
  "disambiguation": {
    "question_key": "flyover_or_ground",
    "options": [
      {"id": "deck",   "label": "On the flyover",   "authority": "MSRDC"},
      {"id": "ground", "label": "Underneath it",     "authority": "BMC"}
    ]
  },
  "resolver_version": "2026.08.1"
}
```

`basis` is stored on the issue and rendered in the "how was this routed?" panel. Users must be able
to audit the routing, and moderators must be able to debug it.

---

## 7. Evaluation

The engine needs a labelled test set, built once and maintained:

| Set | Size (target) | Content |
|---|---|---|
| **Golden** | 500 points | Hand-verified `(point, category, expected authority/ward)` across all 9 corporations, including every hard case |
| **Boundary** | 100 points | Points within 20 m of a ward or authority boundary |
| **Adversarial** | 50 points | Flyovers, station approaches, MHADA layouts, MIDC edges, creek margins |

Metrics: exact-match accuracy, "correct authority, wrong ward" rate, ambiguity rate, and
**wrong-with-high-confidence rate** — the last is the one that must be near zero, because a
confidently wrong route is worse than an honest question.

Golden-file tests run in CI. Any change to weights or layers that regresses the golden set fails the
build.

---

## 8. Caching

| Cacheable | Key | TTL |
|---|---|---|
| Resolution for an H3 r11 cell + category + boundary version | `res:{h3_11}:{category}:{ver}` | 24 h |
| Ward polygon lookups | in-process, refreshed on boundary version bump | — |
| Department/SLA config | in-process, invalidated on config change | — |

H3 r11 cells are roughly 25 m across — small enough that a cached resolution is safe for most points,
with the caveat that cells straddling a boundary must not be cached. The cache writer checks
`ST_Intersects(cell_boundary, ward_boundary)` and skips caching straddlers.

---

## 9. Failure behaviour

| Failure | Response |
|---|---|
| No boundary contains the point (outside MMR) | `unresolved` + "we don't cover this area yet" + a notify-me option |
| Boundary data missing for the corporation | `unresolved` + manual triage queue + the corporation's public channel list |
| Confidence below threshold | Ask the disambiguating question; if unanswered, route to the highest-scoring candidate **and label the issue as low-confidence** |
| PostGIS unavailable | Report is still accepted and queued; jurisdiction resolved on retry |

**Never** silently route on low confidence without labelling it.

---

## 10. Implementation checklist

- [ ] Load and version BMC ward polygons; verify against the BMC ArcGIS layers
- [ ] Digitise or source ward polygons for the other 8 corporations
- [ ] Build the elevated-structures layer (flyovers, metro viaducts, sea link)
- [ ] Geofence all suburban railway station premises
- [ ] Populate `authority_departments` for BMC (all categories) as the v0.1 scope
- [ ] Build the golden evaluation set (500 points) before shipping v0.1
- [ ] File the first road-ownership RTIs
- [ ] Implement the disambiguation question UI with photo options
