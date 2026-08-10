# Ingestion and scrapers

Every external feed enters through one framework with one set of guarantees. Ad-hoc scripts are not
permitted — a scraper that nobody can audit produces data nobody can cite.

---

## 1. Guarantees

Every ingester must provide all six:

| # | Guarantee | Enforcement |
|---|---|---|
| 1 | **Archive the raw artefact** before parsing | The fetch stage writes to S3 + `raw_documents` with a SHA-256; parsing reads from the archive, never from the network |
| 2 | **Idempotent** | Re-running over the same period produces no duplicates; upserts keyed on `(source_id, external_id)` |
| 3 | **Resumable** | Checkpointed; a crash resumes from the last committed cursor |
| 4 | **Polite** | Rate limits, backoff, conditional requests, honest User-Agent |
| 5 | **Loud on failure** | Zero-row runs and parse-rate drops alert; they never pass silently |
| 6 | **Provenance-preserving** | Every derived row carries `source_id`, `raw_document_id`, and `retrieved_at` |

---

## 2. Framework

```go
type Ingester interface {
    ID() string                                    // matches sources.id
    Schedule() string                              // cron expression
    Fetch(ctx context.Context, cur Cursor) (Batch, Cursor, error)
    Parse(ctx context.Context, doc RawDocument) ([]Record, error)
    Load(ctx context.Context, recs []Record) (LoadStats, error)
}
```

The runner handles: scheduling, distributed locking (one instance per ingester), archival, retry
with jitter, metrics, alerting, and cursor persistence. An ingester implementation contains only
source-specific logic.

```
   scheduler ──▶ lock ──▶ Fetch ──▶ archive(S3 + sha256 + raw_documents)
                                          │
                                          ▼
                                       Parse ──▶ validate ──▶ Load(upsert)
                                          │                      │
                                          └── parse_failures ────┴──▶ metrics + alerts
```

---

## 3. Politeness policy

Non-negotiable defaults, overridable only downward:

| Setting | Default |
|---|---|
| Requests per host | 1 per 3 seconds |
| Concurrency per host | 1 |
| User-Agent | `TraceSarkar/1.0 (+https://tracesarkar.org/about/crawler; contact@tracesarkar.org)` |
| `robots.txt` | Fetched, cached 24 h, obeyed |
| Conditional requests | `If-None-Match` / `If-Modified-Since` always sent when known |
| Backoff | Exponential with jitter on 429/5xx; circuit-break after 5 consecutive failures |
| Crawl window | Prefer 00:00–06:00 IST for bulk runs |
| Terms of use | Reviewed and recorded in the source register **before** the ingester is enabled |

**If a site's terms forbid automated access, we do not scrape it.** The decision is recorded in the
source register and RTI becomes the acquisition channel. This is a hard rule, not a risk calculation.

---

## 4. The archive

The archive — not the parsed row — is the evidentiary record.

```
s3://tracesarkar-archive/
  {source_id}/
    {yyyy}/{mm}/{dd}/
      {sha256}.{ext}          # write-once, versioning on, object lock where supported
      {sha256}.meta.json      # url, headers, retrieved_at, fetcher version
```

| Property | Value |
|---|---|
| Immutability | Write-once; object lock in compliance mode where the provider supports it |
| Deduplication | `raw_documents.sha256` is unique; identical re-fetches are recorded but not re-stored |
| Retention | Indefinite for procurement and legal documents; 24 months for news pages |
| Access | Private; archive references in public APIs are hashes, not URLs |

A citation in a published attribution is `(source_id, external_id, retrieved_at, sha256)`. Anyone
disputing a fact can be given the exact artefact it came from.

---

## 5. Ingesters

### 5.1 Procurement

| Ingester | Source | Cadence | Notes |
|---|---|---|---|
| `mahatenders` | `mahatenders.gov.in` | Daily 02:00 IST | GePNIC; session and captcha handling; paginate by closing date and organisation |
| `bmc_tenders` | `portal.mcgm.gov.in` | Daily 02:30 | SAP portal; brittle URLs; expect frequent parser maintenance |
| `cppp` | `eprocure.gov.in` | Daily 03:00 | Central bodies |
| `contract_pdfs` | Follow-on | Continuous | Downloads linked PDFs for contracts lacking extraction |

Output: `contracts`, `contract_documents`, and extraction jobs.

### 5.2 Boundaries and geography

| Ingester | Source | Cadence |
|---|---|---|
| `bmc_arcgis` | BMC ArcGIS REST services | Weekly |
| `datameet_boundaries` | DataMeet GitHub | Monthly (or on release) |
| `osm_roads` | Overpass, MMR bbox | Monthly |

Boundary changes create a **new version row**, never an update. A boundary version bump invalidates
the jurisdiction cache and triggers a re-resolution job for open issues.

### 5.3 Weather and sensors

| Ingester | Source | Cadence |
|---|---|---|
| `iitm_mesonet` | `mumbairain.tropmet.res.in` | 15 min during monsoon, hourly otherwise |
| `mumbai_flood` | `mumbaiflood.in` | 15 min during monsoon |
| `imd` | IMD observatories and warnings | Hourly |
| `safar_aqi` | SAFAR | Hourly |

⚠️ Terms of use for MESONET and Mumbai Flood must be confirmed before enabling — the data is stated
to be freely available for research and academic use, which may not cover a public platform.

### 5.4 News

| Ingester | Source | Cadence |
|---|---|---|
| `news_rss` | English + Marathi Mumbai dailies | 30 min |
| `pib` | `pib.gov.in` | Hourly |
| `dgipr` | Maharashtra DGIPR | Hourly |
| `corp_notices` | Per-corporation press/notice pages | Daily |
| `railway_notices` | WR/CR mega-block notices | Daily |

**Store headline, URL, publication, date, and a short extract only. Never republish full text.**
Entity extraction runs downstream (see [AI pipeline](04-ai-pipeline.md)).

### 5.5 Corporate and blacklist

| Ingester | Source | Cadence |
|---|---|---|
| `mca_lookup` | MCA master data | On demand, per contractor |
| `blacklist_notices` | Per-corporation debarment notices | Weekly |

---

## 6. Parser resilience

Government sites change layout without notice. Design for it:

| Practice | Detail |
|---|---|
| **Parse from the archive** | A layout change breaks the parser, not the fetch — historical documents remain reparseable |
| **Schema assertions** | Every parser asserts expected field presence and types; violations increment `parse_failures` |
| **Row-count sanity** | A run producing < 20% of the trailing-7-day median alerts |
| **Field-fill monitoring** | Per-field fill rates tracked over time; a field silently going empty is the classic scraper failure |
| **Golden fixtures** | Archived sample pages in the test suite; parser changes run against them in CI |
| **Reparse capability** | `reparse --source=mahatenders --since=2026-01-01` re-runs the current parser over archived documents |

That last capability is why the archive exists. When a parser bug is found, the fix is retroactive.

---

## 7. Scheduling and locking

| Concern | Approach |
|---|---|
| Scheduling | Cron expressions in the source register; a single scheduler service |
| Locking | Postgres advisory lock per ingester; only one instance runs at a time |
| Overlap | A run that exceeds its interval skips the next tick and alerts |
| Backfill | `ingest backfill --source=X --from=… --to=…` with a lower rate limit |
| Manual run | `ingest run --source=X --dry-run` prints what it would load |

---

## 8. Monitoring

Per-ingester metrics:

| Metric | Alert |
|---|---|
| `ingest_last_success_timestamp` | Alert if older than 3× the cadence |
| `ingest_records_loaded` | Alert on a zero run or < 20% of the 7-day median |
| `ingest_parse_failures` | Alert above 5% of documents in a run |
| `ingest_http_errors` | Alert on circuit-break |
| `ingest_duration_seconds` | Alert above 3× the median |
| `field_fill_rate{source,field}` | Alert on a drop above 20 points |

Ingestion health is exposed in the Watchdog TUI `s` view — a journalist should be able to see that a
number is stale before they publish it.

---

## 9. Handling failure honestly

If a source has been failing for N days, the platform says so:

- Ward pages show "contract data last updated 14 days ago" when the tender ingester is stale
- Published attributions carry `retrieved_at` so staleness is visible in every citation
- The public status page lists per-source freshness

Silently serving stale procurement data as current is the failure mode that would most damage
credibility.

---

## 10. Implementation checklist

- [ ] Build the ingester framework with archive, lock, retry, and metrics
- [ ] Populate the source register with terms reviews for every planned source
- [ ] Implement `mahatenders` for one ward, one category, one financial year
- [ ] Implement `bmc_arcgis` boundary loader with versioning
- [ ] Implement `iitm_mesonet` (after terms confirmation)
- [ ] Implement `news_rss` for 6 publications
- [ ] Golden fixtures and parser tests for each
- [ ] `reparse` and `backfill` commands
- [ ] Freshness display on public surfaces
