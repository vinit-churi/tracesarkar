# System overview

**Design bias: boring infrastructure, one language, one database.** The novelty budget is spent
entirely on the domain (jurisdiction resolution, contract matching, escalation), not on the stack.

---

## 1. Context diagram

```
   ┌──────────┐   ┌──────────┐   ┌───────────┐   ┌──────────┐
   │  Mobile  │   │   Web    │   │ WhatsApp  │   │ Watchdog │
   │  (PWA)   │   │  public  │   │   bot     │   │   TUI    │
   └────┬─────┘   └────┬─────┘   └─────┬─────┘   └────┬─────┘
        │              │               │              │
        └──────────────┴───────┬───────┴──────────────┘
                               │  HTTPS / WSS
                        ┌──────▼───────┐
                        │   API Gateway │  Traefik (TLS, routing, rate limit)
                        └──────┬───────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
 ┌──────▼──────┐        ┌──────▼──────┐        ┌──────▼──────┐
 │  api        │        │  realtime   │        │  admin      │
 │  (Go, REST) │        │  (Go, WS)   │        │  (Go)       │
 └──────┬──────┘        └──────┬──────┘        └──────┬──────┘
        │                      │                      │
        └──────────┬───────────┴──────────────────────┘
                   │
        ┌──────────▼───────────────────────────────────────────┐
        │  PostgreSQL 16 + PostGIS       │  Redis (queue,cache) │
        │  (single source of truth)      │  S3 (media, archive) │
        └──────────▲───────────────────────────────▲───────────┘
                   │                               │
     ┌─────────────┴──────────┐         ┌──────────┴──────────┐
     │  worker pool (Go)      │         │  ingest (Go)        │
     │  · classify            │         │  · tenders          │
     │  · locate              │         │  · news             │
     │  · dedupe              │         │  · weather/sensors  │
     │  · attribute           │         │  · boundaries       │
     │  · generate share kit  │         │  · blacklists       │
     │  · generate instrument │         └──────────┬──────────┘
     │  · file (adapters)     │                    │
     └─────────┬──────────────┘                    │
               │                                   │
     ┌─────────▼─────────┐              ┌──────────▼──────────┐
     │  Claude API       │              │  External sources   │
     │  (vision, text,   │              │  Mahatenders, BMC,  │
     │   batch)          │              │  IITM, IMD, news…   │
     └───────────────────┘              └─────────────────────┘
```

---

## 2. Services

| Service | Responsibility | Scaling characteristic |
|---|---|---|
| **api** | REST + Open311 endpoints, auth, read models | CPU-light, IO-bound; scale horizontally |
| **realtime** | WebSocket fan-out for live feeds and TUI `--live` | Connection-bound; goroutine per connection |
| **worker** | The enrichment pipeline; consumes the job queue | Bursty; scale on queue depth |
| **ingest** | Scheduled scrapers and feed pullers | Time-boxed; runs on a cron-like scheduler |
| **admin** | Moderation console, source register, config editing | Low volume, high privilege; separate deployment and network policy |

All five are Go binaries from one module. Shared domain code lives in `internal/`.

**Why one language:** the team is small, and the cost of context-switching between a Python ML
service and a Go API exceeds any benefit. The AI work is API calls, not model training — there is no
Python-only dependency in the critical path.

---

## 3. The write path (report submission)

```
POST /v1/reports  (multipart: images + metadata)
   │
   ├─▶ 1. validate, authenticate, rate-limit
   ├─▶ 2. stream media to S3 (archival original)
   ├─▶ 3. INSERT report (status=pending)         ◀── durable here; nothing after this may lose data
   ├─▶ 4. enqueue enrich_report job
   └─▶ 5. return 202 with report_id + optimistic estimate

worker: enrich_report
   ├─▶ redact (faces, plates) → public derivative → S3
   ├─▶ classify (VLM, structured output)          [retryable, idempotent]
   ├─▶ locate (PostGIS + overrides)               [retryable, idempotent]
   ├─▶ dedupe → attach to existing issue OR create issue
   ├─▶ attribute (contract match)                 [retryable, may defer]
   ├─▶ contextualise (weather, news, history)     [best-effort]
   ├─▶ evaluate verification threshold
   ├─▶ generate share kit (if verified)
   └─▶ publish realtime events
```

**Invariant:** step 3 is the durability boundary. Everything after it is a retryable job. A VLM
outage delays enrichment; it never loses a report.

Each worker stage is idempotent and keyed by `(report_id, stage, stage_version)` so re-runs after a
model or rule change are safe and auditable.

---

## 4. The read path

| Surface | Pattern |
|---|---|
| Issue detail | Direct read with joins; cached in Redis with an issue-version key |
| Ward/map tiles | Pre-aggregated materialised views refreshed on a schedule; H3-bucketed |
| Feed | Keyset pagination, never OFFSET |
| Watchdog queries | Read replica; query DSL compiled to parameterised SQL with a whitelist of fields |
| Public API | Same read models, coarsened geometry, aggressive caching |

Read replicas absorb analytical load so a journalist's expensive query cannot degrade report
submission.

---

## 5. Data flow for the contract join

The most valuable and most complex path. See [tender engine](06-tender-engine.md).

```
ingest/tenders ──▶ raw_documents (S3 + sha256)
                        │
                        ├──▶ parse structured fields ──▶ contracts (OCDS-shaped)
                        │
                        └──▶ VLM/OCR extraction ──────▶ contract_terms
                                                          · defect_liability_period
                                                          · work-site description
                                                          · milestones
                                                          │
                     geocode / gazetteer / road matching ─┘
                                                          │
                                                          ▼
                                                   contract_sites (geometry)
                                                          │
                       issue.location ──── spatial join ──┘
                                                          │
                                                          ▼
                                              issue_contract_match
                                              (confidence, basis, version)
```

`issue_contract_match` is versioned. When the matcher improves, old matches are re-run and the change
is recorded — a published attribution that later turns out to be wrong must leave a visible trail.

---

## 6. Storage layout

| Store | Contents | Notes |
|---|---|---|
| **PostgreSQL + PostGIS** | All domain data, geometry, timelines, config | Single source of truth |
| **Read replica** | Analytical and Watchdog queries | Streaming replication |
| **Redis** | Job queue, rate limits, ephemeral caches, WebSocket pub/sub | Nothing durable |
| **S3 (media)** | Archival originals (private), public derivatives (CDN) | Lifecycle rules; server-side encryption |
| **S3 (archive)** | Every scraped page and PDF, verbatim, with SHA-256 | The evidentiary record; write-once |

**Why not a separate search engine?** Postgres full-text plus `pg_trgm` covers the initial document
search need. Add OpenSearch only when the corpus and query complexity demonstrably outgrow it — and
record that as an ADR.

---

## 7. Concurrency model

Go's concurrency is the reason for the language choice, and it is used in three specific places:

| Use | Pattern |
|---|---|
| **Monsoon submission spikes** | Bounded worker pools with a queue; backpressure to the API returns 202 with a longer estimate, never a 5xx |
| **WebSocket fan-out** | One goroutine per connection, a hub per ward topic, buffered channels with slow-consumer eviction |
| **Parallel enrichment** | Independent enrichment stages (contextualise sub-tasks) run concurrently with `errgroup`, bounded by a semaphore against external rate limits |

**Rules:** every goroutine has an owner and a lifetime tied to a `context.Context`. No unbounded
`go func()`. Every external call has a timeout, a retry budget, and a circuit breaker.

---

## 8. Failure and degradation

| Component down | Behaviour |
|---|---|
| VLM API | Reports accepted, queued unclassified; UI says "classifying"; batch catch-up when restored |
| PostGIS boundaries stale | Jurisdiction returns lower confidence, asks the user |
| Contract matcher | Issue proceeds without attribution; attribution backfilled later |
| Filing adapter | Falls back to draft + deep link; user can file manually and paste the reference |
| Redis | Queue drains from a Postgres-backed outbox; degraded throughput, no data loss |
| Object storage | Submission blocked (we will not accept a report we cannot store); clear error and offline retry |
| Realtime | Clients fall back to polling |

Degradation is designed and tested, not discovered. Each row above has a failure-injection test.

---

## 9. Environments

| Environment | Purpose | Data |
|---|---|---|
| `local` | Development | Docker Compose; seeded synthetic MMR data |
| `staging` | Pre-production | Anonymised subset; real boundary and contract data |
| `production` | Live | Real data |

Production media and archive buckets are never readable from non-production environments.

---

## 10. Key architectural constraints

1. **Postgres is the single source of truth.** Redis, S3, and caches are derived and reconstructible.
2. **Every enrichment is versioned and re-runnable.** Model and rule changes must be replayable.
3. **The archive is immutable.** Scraped artefacts are write-once with a hash.
4. **Timelines are append-only.** Corrections are new entries, never edits.
5. **The public API and the TUI use the same endpoints as the first-party clients.** No privileged
   backdoor keeps the API honest.
6. **No PII in logs, metrics, or traces.**
7. **All legal constants are configuration**, versioned, with citations.

---

## 11. Related documents

- [Data model](02-data-model.md)
- [API design](03-api-design.md)
- [AI pipeline](04-ai-pipeline.md)
- [Jurisdiction engine](05-jurisdiction-engine.md)
- [Tender engine](06-tender-engine.md)
- [Ingestion and scrapers](07-ingestion-and-scrapers.md)
- [Deployment](09-deployment.md)
- [Security and privacy](11-security-and-privacy.md)
- ADRs: [`docs/04-adr/`](../04-adr/)
