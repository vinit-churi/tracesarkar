# How much of Borivali has contract data

**Measured 29 September 2026.** First-hand, against the collected archive and OpenStreetMap.

The attribution join answers *which contract covers this stretch*. This measures how often it can
answer at all — which decides whether "this stretch is under warranty" is the product's headline or
its exception.

---

## The number

**Roughly half of Borivali's public road network lies along a stretch BMC has published contract
geometry for.** The honest range is 43–57%, and the reason it is a range rather than a figure is
below.

| Measure | Result |
|---|---|
| OSM public carriageway inside R/C | 279.2 km |
| BMC contract geometry we hold | 54.2 km |
| Network within 5 m of BMC geometry | 121.0 km · **43%** |
| Network within 10 m | 138.6 km · **50%** |
| Network within 15 m | 148.6 km · **53%** |
| Network within 25 m | 159.8 km · **57%** |
| Raw length ratio (54.2 / 279.2) | 19.4% |

**The raw ratio understates badly and should not be quoted.** OSM splits one street into many ways
and maps dual carriageways as two lines; BMC publishes a single centreline per work. Comparing
total lengths compares different things.

**The 25 m figure overstates**, because 25 m can reach a parallel service lane. The curve flattens
between 10 m and 25 m, which is what puts the answer near half.

---

## Which roads

| Class | Total | Covered (25 m) | Share |
|---|---|---|---|
| secondary | 28.6 km | 21.9 km | **77%** |
| tertiary | 61.0 km | 41.8 km | **69%** |
| primary | 14.8 km | 10.0 km | **68%** |
| residential | 151.9 km | 80.1 km | 53% |
| living_street | 10.2 km | 5.2 km | 51% |
| trunk | 5.2 km | 0.7 km | 14% |
| unclassified | 7.6 km | 0.0 km | 0% |

Coverage is **best on the busier roads** — secondary, tertiary and primary all sit near 70% or
above. That matters more than the headline: a pothole on a road people actually drive on is more
likely to be reported, and more likely to be attributable.

**The 14% on trunk roads is correct, not a gap.** Trunk in Borivali is the Western Express Highway,
which is not BMC's road. It belongs to the state, and the absence of BMC contract data for it is
the jurisdiction engine's problem to state, not the attribution engine's to fill.

---

## A cross-check worth recording

**100% of BMC's contract geometry lies within 20 m of an OSM road.** The two datasets describe the
same streets; BMC's is a subset, not a different network. That rules out the failure where the
geometry we attribute against turns out to be drawn somewhere else entirely.

---

## What this means for the product

1. **The coverage-honesty line (K7) is the common case, not an edge case.** Something close to half
   of reports will have no contract to name. "We don't have contract data for this stretch yet" has
   to read as a normal, informative answer — not as an apology or an error.
2. **Jurisdiction is the floor; attribution is the differentiator.** Routing works for every report.
   Contract attribution works for about half. The product has to be worth using at the floor.
3. **The thesis is testable at this coverage.** At half, there will be enough attributed reports to
   see whether attribution changes what a citizen does — which is v0.1's real exit criterion.
4. **Hand-built contract data has a defined target.** The gap is ~120 km of mostly residential
   street. That is what the manual dataset in the MVP plan is for, and it now has a size.

---

## Method, so it can be repeated or disputed

1. R/C boundary from BMC's own `geo/getwardlayer`, loaded into PostGIS as geography.
2. Road centrelines from OpenStreetMap via Overpass, bounding box of R/C, classes `motorway`,
   `trunk`, `primary`, `secondary`, `tertiary`, `residential`, `unclassified`, `living_street`.
   `service` was **excluded**: BMC's concretisation programme does not cover private lanes and
   driveways, and including them would understate coverage unfairly.
3. Clipped to the ward polygon with `ST_Intersection`, measured with `ST_Length` on geography.
4. Covered length is OSM carriageway within *n* metres of any loaded `road_segments` geometry.

**Caveats.** OSM completeness in Mumbai is good but not uniform, and an unmapped street counts as
neither covered nor uncovered. BMC's published geometry is the concretisation programme only;
asphalt works under other programmes are not in this API at all, so the true share of roads with
*some* contract somewhere is higher than the share we can currently attribute.

OpenStreetMap data © OpenStreetMap contributors, ODbL. Used here for measurement; no derived
geometry is published.
