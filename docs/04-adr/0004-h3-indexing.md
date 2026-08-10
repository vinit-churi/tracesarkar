# ADR 0004 — H3 for bucketing, never for jurisdiction

**Status:** Accepted · **Date:** 2026-08-10

## Context

Two operations happen constantly and must be fast: finding candidate duplicate reports near a point,
and aggregating issues into heatmap tiles. Both are approximate operations over a lot of points.
Meanwhile, jurisdiction resolution is an exact operation over administrative polygons.

The temptation is to use one spatial index for both.

## Decision

Use **H3** (Uber's hexagonal hierarchical index) for:

- **Dedup candidate generation** — `h3_r11` cell plus its `k_ring(1)` neighbours as the candidate
  set, then exact distance filtering in PostGIS.
- **Heatmap aggregation** — `h3_r8`/`h3_r9` cells as the tiling unit for public heatmaps.
- **Cache keys** — jurisdiction resolutions cached per `(h3_r11, category, boundary_version)`.

**Never** use H3 for:

- Jurisdiction containment
- Ward assignment
- Any published statement about which authority owns a location

## Rationale

H3 cells do not respect administrative boundaries. A single r11 hexagon (~25 m across) can straddle
a ward line. Assigning jurisdiction by cell would be wrong for every point near a boundary — which
is exactly where routing errors are most damaging and most visible.

The cache is the subtle case: caching a resolution per cell is safe only for cells wholly inside one
ward. The cache writer checks `ST_Intersects(cell_boundary, ward_boundary)` and refuses to cache
straddling cells.

## Alternatives

| Option | Why not |
|---|---|
| **PostGIS for everything** | Correct, but dedup candidate generation over millions of points is measurably slower without a cheap pre-filter. |
| **Geohash** | Rectangular cells with non-uniform neighbour distances; worse for radius queries. |
| **S2** | Comparable quality; H3's uniform neighbour distance is a better fit for radius-style dedup, and the Go and Postgres tooling is good. |
| **H3 for jurisdiction too** | Wrong at every boundary. Rejected on correctness. |

## Consequences

- Two spatial representations to keep in sync. Mitigated by computing H3 cells as generated columns
  from the geometry, so they cannot diverge.
- The `h3` Postgres extension is a dependency. If it proves troublesome to operate, H3 cells can be
  computed in Go and stored as plain text with a btree index, at a small cost in query flexibility.
- Developers must understand the distinction. It is stated in `CLAUDE.md` and enforced by the fact
  that jurisdiction code has no access to H3 columns.
