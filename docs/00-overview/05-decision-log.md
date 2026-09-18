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
