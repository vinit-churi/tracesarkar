# Backlog

Concrete tasks, ordered. Everything here is scoped to be startable without further design work.

**Legend:** `[R]` research · `[D]` data acquisition · `[E]` engineering · `[L]` legal/compliance ·
`[P]` product/design.

---

## Now — v0.1 critical path

### Research and data (blocking; start immediately)

- [ ] `[R]` Enumerate every layer on the two BMC ArcGIS endpoints; record schema, vintage, licensing
- [ ] `[D]` Load and verify the R/S ward boundary polygon against BMC's own source
- [ ] `[R]` Compile the BMC `(category → department, SLA, escalation ladder)` table for road defects
- [ ] `[R]` Obtain BMC PIO and First Appellate Authority details for the Roads department
- [ ] `[D]` **Build the manual R/S road contract dataset** — one financial year, 20–60 contracts,
      archived with hashes, DLP extracted verbatim with page numbers, work sites geocoded by hand
- [ ] `[D]` Build the R/S gazetteer: road names + aliases, landmarks, nakas, junctions
- [ ] `[R]` **Obtain and archive the current Maharashtra RTI Rules** (fee schedule) — blocking for v0.3
- [ ] `[R]` Obtain the full text of the Bombay HC order dated 13 Oct 2025; extract operative paragraph
      numbers for citation
- [ ] `[R]` Identify the compensation committees constituted under that order (composition, addresses)
- [ ] `[R]` Terms-of-use review: Mahatenders, BMC portal, BMC ArcGIS, IITM MESONET, mumbaiflood.in,
      OSM, DataMeet — record every answer in the source register
- [ ] `[D]` File the first RTI: BMC road-ownership inventory for R/S ward

### Evaluation sets (blocking for the features they gate)

- [ ] `[D]` 500 labelled road-defect photographs (day, night, rain, motion blur, varying severity)
- [ ] `[D]` 100 hazard photographs for the recall test
- [ ] `[D]` 100 non-civic photographs for the rejection test
- [ ] `[D]` 200-point jurisdiction golden set for R/S, including boundary and adversarial cases

### Engineering

- [ ] `[E]` Repo scaffold: Go module, `cmd/`/`internal/` layout, Makefile, Docker Compose, CI
- [ ] `[E]` Migration framework; schema for accounts, reports, issues, media, events, authorities, wards
- [ ] `[E]` Phone OTP auth; hashed storage with pepper; rate limits
- [ ] `[E]` `POST /v1/reports` with multipart upload, idempotency, durable-then-async
- [ ] `[E]` Object storage integration: archival originals + public derivatives
- [ ] `[E]` Worker framework: job queue, retries, idempotent stages, stage versioning
- [ ] `[E]` Redaction pipeline (server-side first) + adversarial test set
- [ ] `[E]` Classification call with structured output + prompt file + eval harness
- [ ] `[E]` Jurisdiction resolver + confidence model + golden-file tests
- [ ] `[E]` Disambiguation question flow (photo options, not dropdowns)
- [ ] `[E]` Dedup (H3 candidate generation + PostGIS distance + phash) and issue creation
- [ ] `[E]` Attribution join against the manual contract dataset
- [ ] `[E]` `legal_constants` table + the 48-hour SLA with its citation
- [ ] `[E]` SLA clock and breach detection
- [ ] `[E]` Append-only `issue_events` with hash chaining
- [ ] `[E]` Issue permalink page with OG cards, coarsened location, source citations
- [ ] `[E]` Share kit: annotated image renderer + text generation (en/mr) + prohibited-pattern validator
- [ ] `[E]` PWA capture flow (F1) with progressive enrichment display
- [ ] `[E]` Public status page with per-source data freshness
- [ ] `[E]` Metrics, structured logging with the PII lint, health endpoints
- [ ] `[E]` `tsctl` skeleton with the alert-response commands used in the runbook

### Legal and compliance (blocking for launch)

- [ ] `[L]` DPDP consent notice: draft, review, translate (en/mr)
- [ ] `[L]` Implement separable consent scopes
- [ ] `[L]` Data-principal rights endpoints (access, correct, erase) + `tsctl dsr` commands
- [ ] `[L]` Retention jobs with legal-hold support
- [ ] `[L]` Coarsening enforced at the serialisation layer, with a test that fails if bypassed
- [ ] `[L]` Appoint and publish a grievance officer
- [ ] `[L]` Takedown, right-of-reply, and correction-log workflows
- [ ] `[L]` Terms of use and privacy policy
- [ ] `[L]` Provision `security@`, `conduct@`, and the grievance officer contact
- [ ] `[L]` Trademark and domain search for "TraceSarkar"

---

## Next — v0.2

- [ ] `[E]` Ingester framework: archive, advisory locking, retry, metrics, `reparse`, `backfill`
- [ ] `[E]` `mahatenders` ingester (R/S, roads, one FY) with golden fixtures
- [ ] `[E]` `bmc_tenders` ingester
- [ ] `[E]` Contract PDF extraction (DLP, work site, penalties) with verbatim + page provenance
- [ ] `[E]` Work-site geocoding pipeline (explicit coords → road-name match → from/to → ward)
- [ ] `[E]` Automated attribution with publication thresholds
- [ ] `[E]` **Measure recall/precision against the manual dataset — the v0.2 gate**
- [ ] `[E]` Filing adapter: MyBMC, with reference-number capture
- [ ] `[E]` Offline capture and deferred sync
- [ ] `[E]` On-device redaction
- [ ] `[E]` Road-ownership layer, loaded from RTI replies
- [ ] `[P]` Marathi and Hindi UI complete, native-speaker reviewed

---

## Later — v0.3+

Condensed; see the [roadmap](01-roadmap.md) for milestone framing.

| Milestone | Headline items |
|---|---|
| v0.3 | RTI generation · SLA-breach action prompt · repeat-failure detection · trust score · WhatsApp intake · moderation queue · categories expanded |
| v0.4 | First-appeal automation · **HC compensation claim assistant** · deadline scheduler · weather/news ingestion · voice reporting · Transit Mode · photo-integrity checks |
| v0.5 | **Contractor scorecard** (with legal review gate) · blacklist register · right of reply · correction log · public API · bulk export · campaigns · Greater Mumbai coverage |
| v0.6 | **Watchdog TUI** · saved alerts · embeddable heatmap · verification missions · NGT compiler · RTS appeals · second corporation |
| v0.7 | Document search with OCR · Lokayukta packet · e-Jagriti · OCDS publication · ward sentiment · Gujarati |

---

## Standing tasks

- [ ] Monthly: risk register review; cost review; SLO review
- [ ] Quarterly: restore drill; access review; secret rotation; transparency report; dependency audit
- [ ] April: monsoon preparation (load test, pre-scale, emergency-mode drill, data refresh)
- [ ] Per release: screen-reader pass; low-end device test; vernacular copy review
- [ ] Continuous: verify one ⚠️-marked claim in `docs/01-research/` per week against a primary source

---

## Task hygiene

- A task is startable only if someone could begin it today without another design decision.
- A task that has been "in progress" for more than two weeks is either too big or blocked — split it
  or record the blocker.
- Research tasks produce an artefact (a document, a dataset, an archived source), not a conversation.
- Every `[D]` task's output goes in the source register with a licence and a review date.
