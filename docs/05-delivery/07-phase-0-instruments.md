# Phase 0 — Instruments

**Tier: personal.** Nothing in this phase is visible to anyone but the maintainers.
**Target: November 2026.** See the [roadmap](01-roadmap.md).

**Status, 23 September 2026:** the foundations, P1 (works snapshotter) and the GR half of P2 are
built and running. Stored so far: 2,237 dashboard works, 2,405 road geometries, 30 wards, the
storm-water progress card, and the first Government Resolutions. Remaining: the court watcher,
resolution classification, P3 capture extension, P4 field kit, P5 RTI tracker. Progress is tracked
in the [backlog](03-backlog.md).

---

## 1. The one-sentence scope

> Build the ingestion and archive foundations the public product needs, and use them now — to keep a
> daily record of BMC's own works data, watch for the legal documents the roadmap is blocked on,
> capture contract documents by hand, and collect the evaluation sets v0.1 requires.

---

## 2. Why this comes first

| Reason | Detail |
|---|---|
| **Some evidence cannot be backfilled** | BMC's roads API re-statused two works between 24 August and 18 September 2026 while its total stayed the same. Nobody can now say which two. Every day without a snapshot is a day of history lost |
| **v0.1 needs artefacts, not features** | 500 labelled photos, 200 golden points and a hand-built contract dataset are all prerequisites of v0.1's exit criteria. They are collection work, and collection needs tools |
| **The blockers are documents** | The RTI fee (Q2), orders in PIL 71/2013 (Q28) and every new SLA arrive as GRs or court orders. Watching for them is cheaper than searching for them later |
| **The foundations are the same** | The ingester framework, the archive and `legal_constants` are exactly what the public product runs on. Nothing here is throwaway |

---

## 3. Boundaries

| In | Out |
|---|---|
| Ingester framework, archive, register enforcement | Any public surface |
| P1 works snapshotter (roads, storm-water drains) | Redaction, deduplication, issues |
| P2 blocker watchers (GRs, Bombay HC judgments) | The jurisdiction resolver (its golden set is collected here; the resolver is v0.1) |
| P3 manual-capture extension | Classification of citizen photos (labels here are human) |
| P4 field kit | The research terminal (P6), document search (P7) |
| P5 RTI and complaint tracker | Screen-spec gaps |
| P8 alerts | Any source whose register entry is `blocked` |

Terms follow the [glossary](../00-overview/03-glossary.md): P4 creates **reports**. Nothing in this
phase creates an **issue**.

---

## 4. Hosting

Revised 24 September 2026 by [ADR 0015](../04-adr/0015-indian-egress-for-collection.md): BMC's
roads API and portal refuse foreign cloud runners, so collection is split by reachability.

| Component | Choice | Note |
|---|---|---|
| Compute, all sources | A Cloud Run job in `asia-south1`, triggered by Cloud Scheduler at 02:30 and 14:30 IST ([ADR 0016](../04-adr/0016-collection-as-a-cloud-run-job.md)) | Free. No VM, no disk. [Deployment](../../deploy/cloudrun/README.md) |
| Compute, sources reachable anywhere | GitHub Actions, daily at 08:30 IST | The drain API and the GR watcher. Free, and redundant with the job |
| Database | PostgreSQL 16 + PostGIS in a container on the VPS | Nightly `pg_dump` to R2 under `backups/`, 30-day lifecycle |
| Archive | Cloudflare R2, S3-compatible — **the production archive bucket from the first run**, because snapshot history cannot be re-collected | A bucket lock with indefinite retention makes `archive/` write-once. `media/` is **not** locked: report photographs must stay erasable under the data-principal rights in [security and privacy](../03-architecture/11-security-and-privacy.md). Credentials are scoped per prefix. `backups/` is not locked |
| Secrets | Dokploy secret store | Never in the repository or an image layer |
| Alerts | One push channel, email as fallback | See P8 |

Rationale: [D046](../00-overview/05-decision-log.md).

---

## 5. The register at runtime

`data/sources.yaml` is the source of truth. The `ingest` binary loads it at start and upserts the
`sources` table. The runner refuses to schedule an ingester unless the policy below allows it.

| Tier | May run when |
|---|---|
| **personal** | `status` is `candidate` or later · `robots_ok` is not `false` · terms are not known to forbid automated access · `acquisition` is not `manual` |
| **flagged / public** | All of the above, **and** `terms_reviewed_on` is filled |

`status: blocked` never runs at any tier. Mahatenders is blocked. Output from a source whose terms
are unreviewed may not reach a public surface. Rationale: [D045](../00-overview/05-decision-log.md).

The **personal-data blocklist** in each register entry (`blocklist:`) applies at every tier. The
parser drops those fields before a record exists; they are never written to the database.

---

## 6. Components

### 6.1 Foundations

| Unit | Responsibility |
|---|---|
| `cmd/ingest` | Scheduled ingesters: advisory lock per ingester, retry with jitter, circuit breaker, metrics |
| `cmd/api` | Two personal-only endpoints in this phase: `POST /internal/capture` (P3) and `POST /v1/reports` (P4). Bearer token; no public routes |
| `cmd/tsctl` | Operator CLI. In this phase: `sources`, `snapshots`, `changes`, `rti` |
| `internal/sources` | Register loader and the policy in §5 |
| `internal/ingest` | The `Ingester` interface from [ingestion](../03-architecture/07-ingestion-and-scrapers.md) and its runner |
| `internal/archive` | Write-once put to R2 with SHA-256; `raw_documents` rows |
| `internal/store` | Hand-written SQL; no ORM |
| `internal/notify` | P8 |

### 6.2 P1 — Works snapshotter

| Source | Endpoints |
|---|---|
| `bmc_roads_api` | `publicdashboard/`, the road-geometry layer, `ward/`, `geo/getwardlayer` |
| `bmc_swd_api` | `getprogresscard`, `getgraphdata`, `gettargetquantity`, and the nallah-level read the public dashboard itself makes |

For each endpoint, once a day:

1. Fetch. Record the attempt in `fetch_log` whatever the outcome.
2. Hash the body. If the hash matches the last stored document, stop.
3. Otherwise archive the body, write `raw_documents`, and parse.
4. The parser drops blocklisted fields, then keys each record by a **natural key fixed per
   endpoint** (the roads list has no id, so it uses road name + ward + contractor; the geometry
   layer uses work code + location name). A key collision is an error, not a silent merge.
5. Diff each record against `works.current`. Every changed field becomes a `work_changes` row with
   both values and both document ids. New and vanished records are changes too.
6. Alert on any change, on a zero-row response, and on a schema change.

The storm-water API's path carries the season (`swdwebapi2026`). From January 2027 the ingester also
probes the next season's path monthly and alerts when it answers, so the 2027 desilting season is
captured from its first day.

### 6.3 P2 — Blocker watchers

| Watcher | Source | Cadence | Output |
|---|---|---|---|
| `maha_gr` | Internet Archive mirror of `gr.maharashtra.gov.in` | Daily | New GRs archived. Each is classified from its first pages: department, subject, and any time limit, fee or period it sets |
| `hc_judgments` | Open Bombay HC judgments parquet | Weekly | Rows matching PIL 71/2013, IA No. 29119/2025, or a watched party name |

GR classification uses Claude with structured output, because the Marathi text layer is
glyph-corrupted. The prompt lives in a versioned file; the model is the documented default. A
classification is a lead for a human to read, never a `legal_constants` value.

### 6.4 P3 — Manual-capture extension

A Chrome extension. On your click, it sends the current page, or the PDF it is showing, to
`POST /internal/capture` with the URL and a source id. The server archives it with
`captured_by: manual:<account>`. It never follows links, never runs unattended, and never submits a
form. You browse, and you solve any captcha yourself. This is how the hand-built R/S contract
dataset is assembled from Mahatenders without crawling it
([D027](../00-overview/05-decision-log.md)).

### 6.5 P4 — Field kit

A personal build of the v0.1 capture PWA:

- Photograph bursts with GPS, reported accuracy and time
- Frame type: close, wide, or site noticeboard (K13)
- A label from the road-defect taxonomy, plus `not_civic`
- Conditions: day, night, rain, motion blur
- For golden points: the ward, as known on the ground
- Offline queue with deferred upload, which is v0.2's offline capture, early

Uploads go to `POST /v1/reports` with an idempotency key. Originals land under the private `media/`
prefix, which is erasable, unlike `archive/`. Labels export as a JSONL manifest that refers to
archive keys. **Images are never committed to the repository.**

### 6.6 P5 — RTI and complaint tracker

`tsctl rti add | list | due | draft`. Each filing records authority, mode, date and reference
number, and opens its clocks from `legal_constants`. Drafts render from
[`docs/templates/`](../templates/rti-application.md).

`legal_constants` gains a `verification` column (`verified`, `secondary`, `unverified`). At the
personal tier an unverified clock is shown with that label. On any public surface it is not shown at
all, as [U1](../02-product/14-civic-utility-features.md) already requires.

The first RTI you file through P5 is also the experiment that settles the fee (Q2).

### 6.7 P8 — Alerts

One channel for P1 changes and failures, P2 matches and P5 deadlines. Every alert names its source
and links to the archived document. There is no volume budget at the personal tier, but duplicates
are bundled per source per day.

---

## 7. Data flow

```
cron ─▶ ingest ─▶ fetch ─▶ fetch_log
                     │
                     └─▶ sha256 ─┬─ unchanged ─▶ stop
                                 └─ new ─▶ R2 archive/ (locked) ─▶ raw_documents
                                             │
                                             └─▶ parse (blocklist strip)
                                                   ├─▶ works ─▶ diff ─▶ work_changes ─▶ notify
                                                   └─▶ gr_items / court_items ─▶ notify

extension (P3) ─▶ api /internal/capture ─▶ R2 archive/ ─▶ raw_documents
field kit (P4) ─▶ api /v1/reports ─▶ R2 media/ (private) ─▶ reports ─▶ label manifest
tsctl rti (P5) ─▶ filings + clock_instances ◀─ legal_constants ─▶ notify
```

New tables are specified in
[data model §9A](../03-architecture/02-data-model.md#9a-phase-0-instruments).

---

## 8. Build order

| # | Work | Depends on |
|---|---|---|
| 1 | Go module, `cmd/` and `internal/` layout, Makefile targets, Compose for local, CI | — |
| 2 | VPS with Dokploy; Postgres + PostGIS; R2 bucket with locks; nightly backup | — |
| 3 | Migrations: `sources`, `raw_documents`, `fetch_log`, `works`, `work_changes`, `legal_constants` | 1 |
| 4 | `internal/sources` with the §5 policy | 3 |
| 5 | `internal/archive` and the `internal/ingest` runner | 2, 3 |
| 6 | **P1 roads ingester**, deployed. **The 30-day clock starts here** | 4, 5 |
| 7 | P8 alerts | 6 |
| 8 | P1 storm-water ingester and season probe | 6 |
| 9 | P2 GR watcher with the classification prompt; P2 HC watcher | 5, 7 |
| 10 | `cmd/api` with personal auth; P3 endpoint and extension | 5 |
| 11 | P4 field kit and `POST /v1/reports`; label manifest export | 10 |
| 12 | P5 and the `legal_constants` seed | 3 |
| 13 | Documentation drift tasks from the [backlog](03-backlog.md) | parallel |

Item 6 is on the critical path for the exit criteria and should be live in October 2026.

---

## 9. Tests

| Kind | What |
|---|---|
| Golden fixtures | The archived August sample and the September responses; parsers must reproduce known records exactly |
| Table-driven | Natural-key derivation, diffing (changed, added, vanished), register policy decisions |
| **Blocklist** | A test fails if any blocklisted field from any register entry reaches a parsed record |
| Failure paths | 5xx, timeout, zero rows, schema change, unchanged hash, a key collision, R2 unavailable. Each must be loud and retryable, and none may lose a fetch record |
| Idempotency | Re-running a day produces no new rows |

---

## 10. Definition of done

A Phase 0 component is done when:

- [ ] It runs unattended on the staging VPS
- [ ] It has tests, including the failure path
- [ ] It fails loudly through P8
- [ ] It logs no personal data and no signed URLs
- [ ] Its register entry is current
- [ ] Its documentation is updated in the same commit

---

## 11. Exit criteria

All four must hold before Phase 1 begins:

| # | Criterion | Why |
|---|---|---|
| 1 | **30 consecutive days of P1 snapshots with no missed run** | Proves the record can be kept, which Phase 2's public change log depends on |
| 2 | **500 labelled road-defect photos and 200 golden jurisdiction points** for R/S | v0.1's classification and jurisdiction exit criteria cannot be measured without them |
| 3 | **≥ 20 R/S contract documents** captured through P3 | The non-CC attribution set, and the evaluation set for the Phase 2 gate |
| 4 | **One RTI filed and tracked through P5** | Settles Q2 by doing it |

---

## 12. Running it

```sh
cp .env.example .env          # then fill in R2 and Postgres credentials
make migrate                  # apply the schema
make snapshot                 # snapshot every schedulable source
make watch                    # archive newly published Government Resolutions
make all-collect              # both, as the scheduled job runs them
make status                   # per-endpoint collection health
make changes                  # what changed in the published data
make test                     # offline tests
make test-live                # tests that touch the real bucket and database
```

Daily collection runs in two places, because BMC only answers some of it from abroad:

- **A Cloud Run job in `asia-south1`** (`deploy/cloudrun/`) collects everything, including the roads
  API, twice a day. Credentials come from Secret Manager.
- **GitHub Actions** (`.github/workflows/snapshot.yml`) collects the drain API and watches for new
  Government Resolutions. It needs these repository secrets: `R2_BUCKET_URL`, `R2_BUCKET_NAME`,
  `R2_ACCESS_KEY`, `R2_SECRET_ACCESS_KEY`, `POSTGRESQL_CONNECTION`, `POSTGRES_CA` (the certificate
  itself) and, optionally, `NOTIFY_WEBHOOK_URL`.

**What the runner guarantees, and where it is enforced:**

| Guarantee | Enforced by |
|---|---|
| Every attempt is recorded, including failures | `fetch_log`, written before any parse |
| Unchanged bytes are not re-archived | SHA-256 compared against the last **parsed** document |
| An archived body that fails to parse is retried, never skipped | `raw_documents.parsed_at`, set only after a snapshot is applied |
| Blocklisted fields never reach the database | the parsers, with a test per source |
| A natural-key collision stops the run | `buildRecords`, which refuses to merge two records silently |
| Sources the register forbids are never scheduled | `sources.MayRun`, checked by the snapshotter and the watcher alike |
| A failed night is diagnosable after the machine is gone | `collector_runs`, written before the work starts and updated when it ends |
| A run that fails, or never happens at all, reaches a human | Two Cloud Monitoring alerts by email: a failed execution, and no success for 23h30m |
| A source whose URL changes with the season keeps collecting | `Endpoint.Alternate`, tried when the primary is absent |
| A source unreachable from a runner is collected elsewhere, not dropped | The `network` field in the register, and the split above |

---

## 13. Non-goals

- No public endpoint, page, share card or API.
- No classification calls on citizen photographs; the eval harness is v0.1.
- No issue creation, deduplication or jurisdiction resolution.
- No ingester for a `blocked` source, at any tier.
- No native app.
