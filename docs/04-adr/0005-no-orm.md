# ADR 0005 — Hand-written SQL, no ORM

**Status:** Accepted · **Date:** 2026-08-10

## Context

The data access patterns here are unusual: spatial predicates with mixed geometry types, recursive
authority hierarchies, window functions over append-only event logs, and materialised views for
scorecards. The schema is stable and well-understood (see [data model](../03-architecture/02-data-model.md)).

## Decision

Hand-written SQL in a repository layer, with `pgx` for the driver and `sqlc` for generating typed Go
structs and method signatures from the queries.

```
internal/store/
  queries/
    issues.sql          -- annotated SQL, the source of truth
    contracts.sql
    jurisdiction.sql
  generated/            -- sqlc output, committed
  store.go              -- interfaces the domain layer depends on
```

## Alternatives

| Option | Why not |
|---|---|
| **GORM / ent** | Spatial types are second-class or unsupported; the generated SQL for the queries we care about is unpredictable; the abstraction fights PostGIS at exactly the moments it matters. |
| **Raw `database/sql` with manual scanning** | Correct but tedious and error-prone at scale; `sqlc` gives the same SQL with generated scanning. |
| **A query builder (squirrel, goqu)** | Adds dynamic SQL construction — which is precisely the thing we must not do for the Watchdog query DSL (injection surface). |

## Consequences

- Query changes require regenerating and committing `sqlc` output; CI verifies it is current.
- The Watchdog query DSL compiles to **parameterised** SQL against a whitelist of fields, not to
  dynamically concatenated SQL. This is a security requirement, and having no query builder in the
  codebase removes the temptation.
- Migrations are hand-written, forward-only, and follow expand/contract so a rolled-back application
  version still works against the new schema.
- `EXPLAIN` output for the hot spatial queries is asserted in tests, so a plan regression fails CI.
