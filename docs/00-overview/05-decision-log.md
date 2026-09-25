# Decision log

Chronological record of decisions taken. Architectural ones link to their ADR; product, scope, and
policy decisions live here.

**Format:** date · decision · rationale · status.

---

## 2026-08-10 — Repository established

| # | Decision | Rationale | Status |
|---|---|---|---|
| D001 | **Name: TraceSarkar.** Retire "CivicPulse MMR" | "Pulse" implies monitoring/dashboards — the category we are avoiding. "MMR" bakes in a ceiling. "Trace" commits us to provenance, which is the product's discipline | Accepted; trademark/domain search pending |
| D002 | **Model the asset and the obligation as primary; the grievance is an observation about them** | This is the architectural difference that makes accountability possible. Grievance-centric systems have no memory that three tickets are the same 40 m of road | Accepted |
| D003 | **Jurisdiction is `f(point, category, timestamp)`**, not `f(point)` | A pothole and a mangrove complaint at the same coordinate go to different bodies; boundaries change over time | Accepted → [jurisdiction engine](../03-architecture/05-jurisdiction-engine.md) |
| D004 | **Precision over recall for contract attribution** | A wrong published attribution is a defamation fact pattern. A missing one costs a feature | Accepted → [tender engine](../03-architecture/06-tender-engine.md) |
| D005 | **The v0.1 contract dataset is built by hand** | Tests the hypothesis before attempting the hardest engineering. Doubles as the evaluation set for the automated pipeline | Accepted → [v0.1 MVP](../05-delivery/02-milestone-v0-mvp.md) |
| D006 | **Only a citizen photograph moves an issue to `citizen_confirmed`** | Self-certified resolution is the documented failure mode of every predecessor | Accepted |
| D007 | **The claimed-vs-confirmed ratio is the headline public statistic** | It measures the gap between what a body says and what a citizen sees — the thing nobody currently measures | Accepted → [metrics](../05-delivery/04-metrics.md) |
| D008 | **North star is confirmed fixes, not complaints filed** | Volume without resolution is the failure we are reacting against | Accepted |
| D009 | **Never auto-file a legal instrument** | Legal responsibility follows the signature; a wrong auto-filed instrument is unrecoverable; bulk automated filing invites dismissal as vexatious | Accepted → [ADR 0007](../04-adr/0007-never-auto-file.md) |
| D010 | **Phone-only identity; Aadhaar rejected** | Excludes the most-affected populations, creates a high-value breach target, chills reporting, and adds little sybil resistance over phone | Accepted → [ADR 0011](../04-adr/0011-phone-only-identity.md) |
| D011 | **Publication threshold for contractor names (confidence ≥ 0.75 + acceptable geocode derivation)** | This is a legal control, not a UX preference | Accepted |
| D012 | **Life-safety alerts are never gated on corroboration** | A delayed open-manhole alert is worse than a false one. The "three independent users" rule applies to reputational claims, not hazards | Accepted → [trust and anti-abuse](../02-product/08-trust-and-antiabuse.md) |
| D013 | **Gamification rewards accuracy and closure, never volume** | Rewarding volume produces the exact noise the platform exists to cut through | Accepted → [gamification](../02-product/09-gamification.md) |
| D014 | **No municipal SaaS business model** | Would make the customer the entity being held accountable | Accepted → [GTM](../05-delivery/06-gtm-and-partnerships.md) |
| D015 | **Go backend, Postgres+PostGIS, Docker Swarm + Dokploy** | Boring infrastructure; the novelty budget is spent on the domain | Accepted → [ADR 0002](../04-adr/0002-go-backend.md), [0003](../04-adr/0003-postgres-postgis.md) |
| D016 | **H3 for bucketing, PostGIS for containment** | H3 cells do not respect administrative boundaries; using them for jurisdiction would be wrong at every ward line | Accepted → [ADR 0004](../04-adr/0004-h3-indexing.md) |
| D017 | **AGPL-3.0 for code, CC BY-SA 4.0 for docs and data** | The realistic enclosure risk is a proprietary hosted fork; only the AGPL addresses it | Accepted → [ADR 0012](../04-adr/0012-licensing.md) |
| D018 | **Legal constants live in a database table with citations, never as literals** | Judgments and fee schedules change; the 48-hour SLA is a citation, not a number | Accepted |
| D019 | **The archive, not the parsed row, is the evidentiary record** | Every published fact must be traceable to the exact artefact it came from | Accepted → [ingestion](../03-architecture/07-ingestion-and-scrapers.md) |
| D020 | **Cut: IoT structural health monitoring, bridge deflection prediction, blockchain, anonymous drop** | No sensor access; different threat model; no trust problem solved by a ledger | Accepted → [feature catalog](../02-product/02-feature-catalog.md#explicitly-cut) |
| D021 | **De-scope satellite "ghost project" detection to large-site existence checks, or cut** | Sentinel-2 at 10 m cannot resolve whether a 6 m road was resurfaced. Promising otherwise would be misleading | Provisional; pending the Q9 experiment |
| D022 | **Expose an Open311 GeoReport v2 surface** | Makes a future municipal integration a small ask; gives third-party clients for free | Accepted → [ADR 0010](../04-adr/0010-open311-compatibility.md) |
| D023 | **v0.1 scope: one ward, one category, one authority** | The failure mode to avoid is thirty half-built features and an unvalidated thesis | Accepted → [v0.1 MVP](../05-delivery/02-milestone-v0-mvp.md) |
| D024 | **WhatsApp, not SMS, as the non-app channel** | Supports photos and location natively, near-universal penetration, and it is where MMR civic complaint behaviour already happens | Accepted; revisit if evidence shows a served population WhatsApp misses |
| D025 | **Marathi from v0.1, not "later"** | An English-only platform is a platform for people who already have access to officials | Accepted |

---

## Pending decisions

Tracked in [open questions](../01-research/07-open-questions.md). The ones that will most change the
plan:

| # | Question | Gates |
|---|---|---|
| Q1 | Can tender work-sites be geocoded at usable accuracy? | v0.2 — and possibly the whole thesis |
| Q2 | What is the current Maharashtra RTI fee and format? | v0.3 |
| Q3 | Terms of use for Mahatenders, MESONET, Bhashini | v0.2, v0.4 |
| Q4 | Does attribution change citizen behaviour? | Everything |
| Q6 | Can we obtain per-ward road-ownership inventories? | Routing correctness |

---

## How to add an entry

1. Architectural decisions → write an **ADR** in `docs/04-adr/`, then add a one-line entry here
   pointing to it.
2. Product, scope, or policy decisions → add a row here directly.
3. Reversing a decision → add a **new row** referencing the old one. Never edit history.
4. A decision that turned out wrong is worth recording *as* wrong. The reasoning is the value, not
   the correctness.

## 2026-08-24 — Data availability audit and UI specification

| # | Decision | Rationale | Status |
|---|---|---|---|
| D026 | **Ingest BMC's roads dashboard API as a first-class source** | It publishes 2,237 works with contractor, dates and completion photos plus 2,405 road geometries keyed by work code — much of the contract-to-geometry join, already done by the authority | Accepted → [audit §1.1](../01-research/08-data-availability-audit.md) |
| D027 | **Mahatenders is link-and-cite only; no automated collection** | `robots.txt` is `Disallow: /`, the award search is captcha-gated, and the copyright policy requires permission to reproduce. Hard rule 8 applies | Accepted |
| D028 | **Strip personal contact fields at ingestion, always** | BMC's roads API returns contractor representatives' names and mobile numbers. They must never enter a derivative | Accepted |
| D029 | **Open311 read side only; no `POST /requests`** | Open311's create semantics let a third party file with only an API key, which contradicts ADR 0007 | Accepted → amends [ADR 0010](../04-adr/0010-open311-compatibility.md) |
| D030 | **Rebuild the compensation assistant around the DLSA Committee route** | The Bombay HC order routes claims to a Municipal Commissioner + DLSA Secretary committee that may act on information from any source — a far lower barrier than a High Court petition | Accepted → [key judgments](../01-research/06-key-judgments.md) |
| D031 | **Model a representative layer; BMC has an elected council again since January 2026** | Escalation routing and "who is responsible" surfaces need corporators. The `COUNCILLOR` field in BMC's GIS is stale and must not be displayed | Accepted |
| D032 | **Publish records about the office, never judgments about the person.** No credibility scores, no "promise broken", no party affiliation as a filter | The verbatim layer is well precedented and DPDP-exempt; the derived layer is the platform's own speech, unprotected by safe harbour, in a jurisdiction with no anti-SLAPP statute | Accepted → [ADR 0013](../04-adr/0013-political-accountability-scope.md) |
| D033 | **Election-period freeze as a product control** | RPA s.126's 48-hour silence window is treated by the ECI as covering websites; the political-advertisement question is unresolved. A freeze makes it survivable | Accepted → [ADR 0013](../04-adr/0013-political-accountability-scope.md) |
| D034 | **Drop the green-dominant palette; brand accent is plum on an ink/paper neutral base** | Saffron, green and blue are all party-coded in Maharashtra. Party-neutral perception is a functional requirement | Accepted → [screen spec §2.1](../02-product/11-screen-spec.md) |
| D035 | **`claimed_resolved` and `citizen_confirmed` never share a label, filter or count in any UI** | The gap between them is the platform's headline statistic. A single "Resolved" chip destroys it | Accepted → [screen spec](../02-product/11-screen-spec.md) |
| D036 | **The capture path has no wizard, no category picker and no description field** | The F1 constraint is under 30 seconds with zero required text input. A "details" step is a form | Accepted → [screen spec S02–S06](../02-product/11-screen-spec.md) |

---

## 2026-08-25 — Filing mechanics and the utility layer

| # | Decision | Rationale | Status |
|---|---|---|---|
| D037 | **No server-side submission to any government portal, by any mechanism** — no headless drivers, no request replay | Responsibility follows the signature; a wrongly filed instrument is unrecoverable; bulk filing invites dismissal of the platform as vexatious | Accepted → [ADR 0014](../04-adr/0014-assisted-filing-not-automated-submission.md) |
| D038 | **Never hold a citizen's government-portal credentials or intercept their OTP** | Every filing route requires an account, several require payment. Server-side filing means a credential vault for government identities — a breach consequence out of all proportion to the convenience | Accepted → [ADR 0014](../04-adr/0014-assisted-filing-not-automated-submission.md) |
| D039 | **No anti-detection or fingerprint-spoofing tooling against any government system** | A bot defence is the operator's answer. Circumvention is a different thing to defend, and it would taint every record the platform publishes | Accepted → [ADR 0014](../04-adr/0014-assisted-filing-not-automated-submission.md) |
| D040 | **Build a utility layer, on the condition that every surface routes back to an accountability action** | Reporting is episodic; an app opened twice a year is deleted. But a utility surface with no route back is a worse version of an app that already exists | Accepted → [series U](../02-product/14-civic-utility-features.md) |
| D041 | **Utility features are never measured by engagement** | Time-in-app is the metric that turns a civic record into a feed. The measures stay confirmed fixes, corroborations and instruments filed | Accepted |
| D042 | **Warranty alerts ship in an honest mode before the DLP source is obtained** — completion date plus the Court's five-to-ten-year expectation, both sourced | Two sourced facts beat one inferred one, and it is still more than any other product tells a citizen | Accepted → U2 |
| D043 | **Notification ceiling of five per user per week**, quiet hours 21:00–08:00, no re-engagement messages ever | The utility layer can generate unlimited messages. The budget is the control | Accepted → U9 §9 |

---

## 2026-09-18 — Vertical exploration

| # | Decision | Rationale | Status |
|---|---|---|---|
| D044 | **Exposure tiers replace legal risk as a scoping filter.** Every feature starts as `personal` (maintainers only), `flagged` (feature flag, invite-only) or `public`. The hard rules in `CLAUDE.md` are checked when a feature moves to `public`; the personal-data blocklist applies at every tier | Legal exposure is controlled by who can see a feature, not by whether it is built. Scoping on legal risk was stopping exploration of features that are safe to run privately | Accepted → [vertical exploration §1](../01-research/09-vertical-exploration.md#exposure-tiers) |
| D045 | **At the `personal` tier, an ingester may run on a `candidate` source whose terms are unreviewed**, provided `robots_ok` is not `false`, the terms are not known to forbid automation, and acquisition is not manual. Output may not reach a `flagged` or `public` surface until `terms_reviewed_on` is filled. `blocked` sources never run | Waiting for written terms from MCGM would lose snapshot history that cannot be re-collected. Hard rule 8 — never scrape a source whose terms forbid it — is unchanged; this relaxes only the stricter "review before any run" wording in the register | Accepted → [Phase 0 §5](../05-delivery/07-phase-0-instruments.md#5-the-register-at-runtime) |
| D046 | **Phase 0 runs on one VPS (Dokploy single-node staging) with Cloudflare R2 as the archive.** A bucket lock with indefinite retention makes `archive/` write-once; `media/` is not locked so report photographs stay erasable. The Phase 0 ingesters write to the production archive bucket from their first run | R2 is S3-compatible with no egress fees and supports bucket locks. Public-source documents must be immutable; citizen media must be deletable under data-principal rights. Snapshot history cannot be re-collected, so it goes to the production bucket from day one | Accepted → [deployment §9](../03-architecture/09-deployment.md#9-environments) |
| D047 | **Personal instruments before public v0.1; the roadmap is restructured into Phases 0–6 anchored to the 2027 monsoon.** v0.1 attribution uses BMC's roads API for CC roads | Phase 0 starts archives that cannot be backfilled and builds v0.1's evaluation sets. R/S ward has 118 CC-road works in BMC's API, so the thesis can be tested without waiting for geocoding. Drains and building safety must be live before May 2027 or wait a year | Accepted → [roadmap](../05-delivery/01-roadmap.md) |

---

## 2026-09-24 — Where collection runs

| # | Decision | Rationale | Status |
|---|---|---|---|
| D048 | **Collection that touches an India-restricted source runs from a VM in an Indian region**, started and stopped on a schedule (02:30 IST, on-demand `e2-micro` in `asia-south1`, ~₹45/month). GitHub Actions keeps the sources reachable from anywhere | `roads.mcgm.gov.in` and `portal.mcgm.gov.in` refuse foreign cloud runners; `swd.mcgm.gov.in` and the GR mirror do not. Found by running the collector in CI, not by reading documentation | Accepted → [ADR 0015](../04-adr/0015-indian-egress-for-collection.md) |
| D049 | **On-demand, not spot, for the collector VM** | At two minutes a day the disk dominates the bill, so spot saves about ₹2 a month while adding the one failure this phase exists to prevent: a night that cannot start for want of capacity. A missed night cannot be recovered, because the API publishes current state, not history | Accepted → [ADR 0015](../04-adr/0015-indian-egress-for-collection.md) |
| D050 | **Check reachability from where the code will run, before scheduling any new source** | A source that answers a laptop in Mumbai may refuse a datacenter anywhere. This cost us a failed scheduled run to learn | Accepted → source register |

---

## 2026-09-25 — Collection moves to Cloud Run

| # | Decision | Rationale | Status |
|---|---|---|---|
| D051 | **Collection runs as a Cloud Run job in `asia-south1`, triggered by Cloud Scheduler.** The VM, its disk and its instance schedule are deleted | A job in that region reaches BMC, verified by running one. No disk, no machine to patch, no serial console to lose, and it costs nothing: about 1.5% of the free vCPU allowance and 2 of 3 free scheduler jobs. Cheaper and simpler at once | Accepted → [ADR 0016](../04-adr/0016-collection-as-a-cloud-run-job.md) |
| D052 | **Collect twice a day, 02:30 and 14:30 IST, not once** | Cloud Monitoring's absence conditions cap at 23h30m, so a daily schedule cannot be watched for a missed run without a false alarm every day. Two runs make the dead-man alert meaningful, and catch BMC's daytime edits | Accepted → ADR 0016 |
| D053 | **One command (`ingest all`) rather than per-schedule argument overrides** | Overrides need `run.jobs.runWithOverrides`, which `roles/run.invoker` does not grant. Widening the trigger's permissions in order to collect resolutions is the wrong trade | Accepted → ADR 0016 |
| D054 | **Alert by email on a failed execution, and on no success for 23h30m** | Until now a failure reached a log nobody reads, and a night that never ran was invisible. A missed night cannot be recovered | Accepted → ADR 0016 |

---

## 2026-09-25 — Parallel tracks, and the first backend slice

| # | Decision | Rationale | Status |
|---|---|---|---|
| D055 | **The roadmap runs as three parallel tracks — collection, backend, fieldwork — not as a serial sequence** | Collection now runs itself, and nothing about the backend waits on it or on fieldwork. The original ordering was written before the collector existed. A phase ends when the slowest track lands, and that is usually fieldwork, not code | Accepted → [roadmap](../05-delivery/01-roadmap.md) |
| D056 | **The capture endpoint stores the photograph before it replies** | A 202 promises the capture is safe. Returning before the image is in object storage would make that a lie, and hard rule 7 forbids losing a capture | Accepted → `internal/api` |
| D057 | **The field kit queues captures on the device and uploads when it can** | A walk through a ward passes through dead spots. Losing a photograph because of signal is the same failure as losing a citizen's report | Accepted → `internal/api/fieldkit` |
| D058 | **Email/password and Google sign-in are accepted before the public tier; phone verification still gates publication** | Phone OTP needs a DLT-registered sender and a paid SMS route in India. Gating the first working client on a vendor would stall the field kit and the evaluation set, and the thing ADR 0011 protects — published claims about named parties — does not exist yet | Accepted → [ADR 0017](../04-adr/0017-email-and-google-identity-before-public-tier.md) |
| D059 | **One Flutter codebase serves Android and the web** | The same capture surface is needed on a phone in a ward and in a browser for a demo. Two codebases for one screen is the expensive way to get there. The web build deploys to Cloudflare Pages | Accepted → `app/` |
| D060 | **The media insert is idempotent on the content digest, not only the report** | A retried upload re-sends the same bytes. `SaveReport` was already idempotent but `AddReportMedia` was not, so each retry added a row pointing at the same object — inflating media counts and double-counting a photograph in any evaluation export | Accepted → migration `0007_media_idempotency.sql` |

