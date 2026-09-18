# Backlog

Concrete tasks, ordered. Everything here is scoped to be startable without further design work.

**Legend:** `[R]` research · `[D]` data acquisition · `[E]` engineering · `[L]` legal/compliance ·
`[P]` product/design.

**Restructured 18 September 2026** around the phases in the [roadmap](01-roadmap.md).

---

## Now — Phase 0: instruments

Specification: [Phase 0](07-phase-0-instruments.md). Tier: personal.

### Infrastructure

- [ ] `[E]` Provision the staging VPS: Dokploy single-node Swarm, PostgreSQL 16 + PostGIS, nightly
      `pg_dump` to R2 `backups/` with a 30-day lifecycle
- [ ] `[E]` Create the R2 bucket; a bucket lock with indefinite retention on `archive/` only;
      `media/` stays erasable

### Foundations

- [ ] `[E]` Repo scaffold: Go module, `cmd/ingest`, `cmd/api`, `cmd/tsctl`, `internal/` layout,
      Makefile targets, Docker Compose for local, CI
- [ ] `[E]` Migrations for the Phase 0 tables in
      [data model §9A](../03-architecture/02-data-model.md#9a-phase-0-instruments), plus
      `sources`, `raw_documents`, `legal_constants`, a minimal `accounts`, `reports`,
      `report_media`, `filings`
- [ ] `[E]` `internal/sources`: register loader and the tier policy
      ([D045](../00-overview/05-decision-log.md)), table-driven tests
- [ ] `[E]` `internal/archive` (R2, SHA-256, write-once) and the `internal/ingest` runner (advisory
      lock, retry with jitter, circuit breaker, metrics)
- [ ] `[E]` A test that fails if any register `blocklist` field reaches a parsed record

### Instruments

- [ ] `[E]` **P1 `bmc_roads_api` ingester** with golden fixtures, deployed — starts the 30-day clock
- [ ] `[E]` P8 alerts
- [ ] `[E]` P1 `bmc_swd_api` ingester and the next-season path probe
- [ ] `[E]` P2 `maha_gr` watcher, with the classification prompt in a versioned file and structured
      output
- [ ] `[E]` P2 `hc_judgments` watcher
- [ ] `[E]` `cmd/api` with personal auth; `POST /internal/capture`; the P3 Chrome extension
- [ ] `[E]` P4 field kit PWA, `POST /v1/reports`, label manifest export
- [ ] `[E]` P5 `tsctl rti`; seed `legal_constants` with the nine verified HC constants and the RTI
      clocks marked `secondary`

### Collection — the exit criteria

- [ ] `[D]` 500 labelled road-defect photographs through P4 (day, night, rain, motion blur, varying
      severity), plus 100 hazard and 100 non-civic photographs
- [ ] `[D]` 200-point jurisdiction golden set for R/S, including boundary and adversarial cases
- [ ] `[D]` **≥ 20 R/S contract documents through P3**, DLP extracted verbatim with page numbers,
      work sites geocoded by hand — for roads outside BMC's CC programme
- [ ] `[D]` R/S gazetteer: road names and aliases, landmarks, nakas, junctions
- [ ] `[L]` File the first RTI through P5: BMC's road-ownership inventory for R/S ward. Its fee
      settles Q2

### Research still blocking v0.1

- [ ] `[D]` Load and verify the R/S ward boundary against BMC's own source, noting the 2025
      delimitation
- [ ] `[R]` Compile the BMC `(category → department, SLA, escalation ladder)` table for road defects
- [ ] `[R]` BMC PIO and First Appellate Authority for Roads — from `portal.mcgm.gov.in`, **not**
      `bmc.gov.in`, which is Bhubaneswar
- [ ] `[R]` Obtain a certified copy of the Bombay HC order of 13 Oct 2025 (IA No. 29119/2025)
- [ ] `[R]` Identify the compensation committees constituted under that order
- [ ] `[L]` Written terms requests: MCGM (roads API, SWD API, ArcGIS layers), CPCB, MahaRERA, IITM.
      Record every answer in the source register
- [ ] `[R]` Read Pothole Reporter's routing code and documentation (Q35)
- [ ] `[R]` Human re-read of the DLP GR page images before any row enters `legal_constants`

### Documentation drift found in the September review

- [ ] `[P]` Screen spec: S-numbers and §5 entries for the 21 screens drawn in the Screen Book but
      absent from [`11-screen-spec.md`](../02-product/11-screen-spec.md); §5 entries for S35–S39;
      the site-noticeboard capture affordance (K13); draw S13 and S24
- [ ] `[P]` Feature catalog: rebuild D9 around the DLSA committee route (D030)
- [ ] `[E]` Data model: the representative layer (D031)
- [ ] `[E]` Jurisdiction engine: boundary vintage and the 2025 delimitation
- [ ] `[P]` Accessibility doc: Bhashini's proof-of-concept-only terms as a v0.4 blocker (Q31)
- [ ] `[P]` Tender engine: BMC's roads API as a source; Mahatenders manual-only (D027)
- [ ] `[P]` README: refresh "Docs last reviewed"

---

## Next — Phase 1: v0.1

Specification: [v0.1 MVP](02-milestone-v0-mvp.md). Tier: public, R/S ward only.

### Engineering

- [ ] `[E]` Phone OTP auth; hashed storage with pepper; rate limits
- [ ] `[E]` Open `POST /v1/reports` to the public: idempotency, durable-then-async
- [ ] `[E]` Worker framework: job queue, retries, idempotent stages, stage versioning
- [ ] `[E]` Redaction pipeline (server-side first) + adversarial test set
- [ ] `[E]` Classification call with structured output + prompt file + eval harness over the P4 set
- [ ] `[E]` Jurisdiction resolver + confidence model + golden-file tests over the P4 golden set
- [ ] `[E]` Disambiguation question flow (photo options, not dropdowns)
- [ ] `[E]` Dedup (H3 candidate generation + PostGIS distance + phash) and issue creation
- [ ] `[E]` **Attribution: spatial join to the roads-API geometry (K1)**, plus the hand-built
      dataset for non-CC roads, plus the coverage-honesty line (K7)
- [ ] `[E]` Who do I call (U8)
- [ ] `[E]` SLA clock and breach detection over `clock_instances`
- [ ] `[E]` Append-only `issue_events` with hash chaining
- [ ] `[E]` Issue permalink page with OG cards, coarsened location, source citations
- [ ] `[E]` Share kit: annotated image renderer + text generation (en/mr) + prohibited-pattern
      validator
- [ ] `[E]` Public status page with per-source data freshness
- [ ] `[E]` Metrics, structured logging with the PII lint, health endpoints

### Legal and compliance (blocking for public launch)

- [ ] `[L]` DPDP consent notice: draft, review, translate (en/mr)
- [ ] `[L]` Separable consent scopes
- [ ] `[L]` Data-principal rights endpoints (access, correct, erase) + `tsctl dsr` commands
- [ ] `[L]` Retention jobs with legal-hold support
- [ ] `[L]` Coarsening enforced at the serialisation layer, with a test that fails if bypassed
- [ ] `[L]` Appoint and publish a grievance officer
- [ ] `[L]` Takedown, right-of-reply, and correction-log workflows
- [ ] `[L]` Terms of use and privacy policy
- [ ] `[L]` Provision `security@`, `conduct@`, and the grievance officer contact
- [ ] `[L]` Trademark and domain search for "TraceSarkar"

---

## Later — Phases 2 to 6

Condensed; see the [roadmap](01-roadmap.md) for gates and tiers.

| Phase | Headline items |
|---|---|
| 2 · v0.2 | Completion photo vs citizen photo (K2) · public change log (U14) · **drains and flooding (V1) before 31 May 2027** · works near me (U5) · non-CC attribution and the thesis gate · MyBMC drafts (K30) · offline · redaction on-device · Marathi + Hindi |
| 3 · v0.3 | RTI generation · deadline wallet (U1) · SLA-breach prompt · repeat failures · WhatsApp · trust score · moderation · categories and R-zone expanded · **flagged:** construction sites (V2), building safety (V3), hoardings (V4), research terminal (P6) · decide Q34 |
| 4 · v0.4 | DLSA compensation claim (D9/U13) · escalation ladder · first appeal · trees (V5) · evidence vault (U12) · warranty watch honest mode (U2) · news · Transit Mode · photo integrity · voice (blocked on Q31) |
| 5 · v0.5–0.6 | **Second geography** · public API and Open311 read side · bulk export · Greater Mumbai · right of reply, correction log, takedown · contractor record and corporator lookup (flagged) · Watchdog terminal · verification missions · digests · NGT compiler · RTS appeals |
| 6 · v0.7+ | Document search · OCDS publication · Lokayukta · e-Jagriti · water supply (V6) · toilets (V7) · Gujarati |

---

## Standing tasks

- [ ] Monthly: risk register review; cost review; SLO review
- [ ] Quarterly: restore drill; access review; secret rotation; transparency report; dependency audit
- [ ] April: monsoon preparation (load test, pre-scale, emergency-mode drill, data refresh, capture
      the C1 list and the new desilting-season API path)
- [ ] Per release: screen-reader pass; low-end device test; vernacular copy review
- [ ] Continuous: verify one ⚠️-marked claim in `docs/01-research/` per week against a primary source

---

## Task hygiene

- A task is startable only if someone could begin it today without another design decision.
- A task that has been "in progress" for more than two weeks is either too big or blocked — split it
  or record the blocker.
- Research tasks produce an artefact (a document, a dataset, an archived source), not a conversation.
- Every `[D]` task's output goes in the source register with a licence and a review date.
