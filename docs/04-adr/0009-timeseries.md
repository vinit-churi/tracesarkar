# ADR 0009 — Plain Postgres partitioning for observations, revisit TimescaleDB later

**Status:** Accepted · **Date:** 2026-08-10

## Context

The `observations` table holds sensor and weather readings: ~139 IITM MESONET rain-gauge sites plus
BMC AWS stations plus IMD and SAFAR feeds, sampled as often as every 15 minutes during monsoon. That
is on the order of a few million rows per year — meaningful, but not large.

The queries are: "rainfall near this point in the last 72 hours", "was this location flooded on the
date of this report", and ward-level aggregation for context and dashboards.

## Decision

Use **native Postgres declarative partitioning** by month on `observed_at`, with a BRIN index on
`observed_at` and a GiST index on `location`. Retain raw observations for 24 months, roll up to
hourly and daily aggregates beyond that.

Do **not** adopt TimescaleDB at v0.1.

## Rationale

- The volume does not require it. A few million rows per year with monthly partitions is comfortably
  within plain Postgres capability.
- TimescaleDB adds an extension dependency that constrains hosting options (some managed Postgres
  providers do not offer it) and complicates the backup and upgrade story.
- Continuous aggregates are genuinely nice, but materialised views refreshed on a schedule cover the
  need at this volume.
- Adding it later is straightforward; removing it later is not.

## Alternatives

| Option | Why not (yet) |
|---|---|
| **TimescaleDB** | Better ergonomics for time-series; not yet justified by volume. Revisit if observation volume exceeds ~50M rows/year or if continuous aggregates become a bottleneck. |
| **A separate time-series database (InfluxDB, VictoriaMetrics)** | A second data store to operate, back up, and join against — for data whose main value is being joined to spatial civic data. Rejected. |
| **No storage; query sources live** | Sources are rate-limited and sometimes unavailable, and historical context ("this location flooded on 6 of the last 9 heavy-rain days") requires our own history. |

## Consequences

- Partition management (creating next month's partition) is a scheduled job that must not silently
  fail — an alert covers it.
- The rollup job must be written and tested; it is not free.
- The `create_hypertable` call shown in the data-model document is illustrative and **not** part of
  the v0.1 schema. It is left in the document as a marked future option.
- Revisit condition is explicit: >50M observation rows/year, or aggregate query p95 above 500 ms.
