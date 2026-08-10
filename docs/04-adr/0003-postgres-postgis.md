# ADR 0003 — PostgreSQL + PostGIS as the single source of truth

**Status:** Accepted · **Date:** 2026-08-10

## Context

Almost every question the platform answers is spatial: which ward contains this point, which road
segment is nearest, which contract work-site intersects this location, which issues fall inside this
bounding box. Alongside that we need transactional integrity for the report → issue → escalation
lifecycle and an append-only event log with tamper-evidence.

## Decision

**PostgreSQL 16 with PostGIS** is the single source of truth. Everything else — Redis, object
storage caches, materialised views, search indexes — is derived and reconstructible from it.

Specifically:

- `geography(Point,4326)` and `geography(MultiPolygon,4326)` with GiST indexes for all geometry.
- PostGIS is **authoritative** for containment and distance. H3 is an index, never an authority
  (see [ADR 0004](0004-h3-indexing.md)).
- Native enums for status columns, so invalid states fail at write time.
- Hand-written SQL in a repository layer. **No ORM** (see [ADR 0005](0005-no-orm.md)).
- Read replica for analytical and Watchdog queries, so a journalist's expensive scan cannot degrade
  report submission.

## Alternatives

| Option | Why not |
|---|---|
| **Postgres + a separate spatial service** | Two sources of truth for the same geometry; join complexity for no benefit at this scale. |
| **MongoDB with geospatial indexes** | Weaker spatial predicates, no polygon topology operations, and we need transactions across five tables per report. |
| **Elasticsearch as primary** | Excellent for search, wrong for a transactional lifecycle with foreign keys. |
| **SQLite + Litestream** | Genuinely tempting for a single-node MVP, but PostGIS has no adequate SQLite equivalent for the polygon work, and we need a read replica. |
| **A managed geospatial platform** | Cost, lock-in, and data-residency friction. |

## Consequences

- Operating Postgres well (vacuum, connection pooling, replication, backups) becomes a core
  operational skill. Accepted; it is boring, documented technology.
- Spatial query performance must be watched: GiST index bloat and poor plan choices on mixed
  geometry types are the known failure modes. Golden-file query tests with `EXPLAIN` assertions
  guard the hot paths.
- One database means one thing to back up, one thing to restore, one thing to secure.
