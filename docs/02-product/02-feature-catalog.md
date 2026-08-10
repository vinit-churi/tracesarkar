# Feature catalog

Every feature considered for TraceSarkar, with an honest assessment of value, cost, dependencies,
and the milestone it belongs to. **Being in this catalog is not a commitment to build.**

Legend — **V** value (1–5), **C** build cost (1–5), **R** risk (1–5), **M** milestone.

---

## A. Capture and reporting

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| A1 | One-tap photo report with auto-geotag | 5 | 2 | 1 | v0.1 | The product's front door |
| A2 | AI classification with structured output | 5 | 2 | 2 | v0.1 | See [AI pipeline](../03-architecture/04-ai-pipeline.md) |
| A3 | Offline capture + deferred sync | 5 | 3 | 1 | v0.2 | Monsoon-critical |
| A4 | Multi-frame capture (wide + close) | 3 | 1 | 1 | v0.1 | Materially improves both classification and evidentiary value |
| A5 | Voice-note reporting (vernacular) | 4 | 3 | 2 | v0.4 | Bhashini ASR; removes the typing barrier |
| A6 | WhatsApp intake bot | 5 | 3 | 3 | v0.3 | Likely the largest single distribution unlock in MMR |
| A7 | SMS / missed-call fallback | 2 | 3 | 2 | v1+ | Low value once WhatsApp exists |
| A8 | On-device face and number-plate redaction | 4 | 3 | 2 | v0.2 | DPDP requirement for public display |
| A9 | Dashcam / bulk upload for volunteers | 3 | 4 | 2 | v1+ | A fleet-style survey path |
| A10 | Report-by-forwarding (share a photo from gallery into the app) | 3 | 1 | 2 | v0.3 | Loses live geotag; must ask for location |

## B. Routing and jurisdiction

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| B1 | Ward/authority resolution via PostGIS | 5 | 3 | 3 | v0.1 | Core |
| B2 | Road-ownership override layer | 5 | 4 | 4 | v0.2 | Blocked on RTI data acquisition |
| B3 | Elevated-structure disambiguation | 4 | 3 | 2 | v0.3 | Flyover vs road below |
| B4 | Railway geofence → Transit Mode | 4 | 3 | 2 | v0.4 | RailMadad-compatible fields |
| B5 | Category-conditional department mapping | 5 | 2 | 2 | v0.1 | Cheap, high value |
| B6 | Time-versioned boundaries (CIDCO→NMMC handovers) | 3 | 3 | 2 | v0.5 | Correctness for historical issues |
| B7 | `disputed` jurisdiction state (MHADA layouts) | 4 | 2 | 1 | v0.3 | Naming the dispute is itself useful |
| B8 | Confidence-gated single disambiguating question | 5 | 2 | 1 | v0.1 | Prevents the wrong-routing reputation |

## C. Follow-the-money (procurement)

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| C1 | Tender/award ingestion (Mahatenders, BMC, CPPP) | 5 | 4 | 3 | v0.2 | See [ingestion](../03-architecture/07-ingestion-and-scrapers.md) |
| C2 | OCDS normalisation | 4 | 3 | 2 | v0.3 | Precedent exists; enables publication |
| C3 | Contract → geometry matching | 5 | 5 | 4 | v0.2 | **The hardest and most valuable component** |
| C4 | DLP extraction from contract PDFs | 5 | 4 | 3 | v0.3 | VLM/OCR; the legally decisive field |
| C5 | In-warranty defect flagging | 5 | 1 | 2 | v0.3 | Trivial once C3+C4 exist |
| C6 | Contractor scorecard (MMR-wide) | 5 | 3 | 5 | v0.5 | Highest legal risk; needs methodology review |
| C7 | Corporate lineage mapping (directors, aliases) | 4 | 4 | 5 | v1+ | MCA data; blacklist-evasion detection |
| C8 | Blacklist register aggregation | 4 | 3 | 4 | v0.5 | Per-corporation notices, often PDF |
| C9 | Tender text analytics (restrictive specs, collusion signals) | 3 | 3 | 4 | v1+ | Well-defined NLP; assertions must stay descriptive |
| C10 | Satellite "ghost project" check | 2 | 5 | 4 | v2+ | Sentinel-2 at 10 m cannot resolve road works. Scope to large sites only or drop — see [open questions](../01-research/07-open-questions.md) |
| C11 | Payment vs milestone reconciliation | 4 | 4 | 3 | v1+ | Requires RTI-obtained payment records |

## D. Action and escalation

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| D1 | Filing adapters (MyBMC, Aaple Sarkar, Swachhata, RailMadad) | 5 | 4 | 4 | v0.2 | Terms-of-use review required per platform |
| D2 | Official reference-number capture and tracking | 5 | 3 | 2 | v0.2 | Evidentiary spine |
| D3 | SLA clock (48 h for road defects per Bombay HC) | 5 | 1 | 1 | v0.1 | Config-driven, judgment-cited |
| D4 | One-click RTI generation | 5 | 3 | 3 | v0.3 | Fee/format must be verified live |
| D5 | RTI first-appeal automation | 4 | 2 | 2 | v0.4 | Pure clock + template |
| D6 | NGT Original Application compiler | 4 | 4 | 4 | v0.6 | 6-month limitation countdown is the killer feature |
| D7 | Lokayukta complaint packet | 3 | 3 | 4 | v0.7 | Cannot be fully automated (notarisation) |
| D8 | e-Jagriti consumer complaint | 3 | 3 | 3 | v0.7 | Good for quantifiable-loss cases |
| D9 | **HC pothole compensation claim assistant** | 5 | 3 | 3 | v0.4 | Quantum and forum already fixed by the Court; Maharashtra-only moat |
| D10 | RTS Act appeal | 3 | 2 | 2 | v0.6 | Carries a personal penalty; needs the notified-service list |
| D11 | Escalation ladder auto-advance (draft only, never auto-file) | 5 | 3 | 3 | v0.4 | The compounding mechanism |
| D12 | Legal-aid / pro-bono handoff | 3 | 2 | 2 | v1+ | Partnership-dependent |

## E. Distribution and social

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| E1 | Share kit: tweet / WhatsApp card / image | 5 | 2 | 2 | v0.1 | See [share kit](05-share-kit.md) |
| E2 | Auto-tagging of correct official handles | 4 | 2 | 3 | v0.2 | Handle registry must be curated and rate-limited |
| E3 | Public issue permalinks with OG cards | 4 | 1 | 1 | v0.1 | Free reach |
| E4 | Ward-level campaigns | 4 | 3 | 2 | v0.5 | Collective weight |
| E5 | Embeddable ward heatmap widget | 4 | 3 | 2 | v0.6 | For newsrooms |
| E6 | Public read-only API | 4 | 2 | 2 | v0.5 | Researchers, newsrooms; Open311-compatible |
| E7 | Weekly ward digest (email/WhatsApp) | 3 | 2 | 1 | v0.6 | Retention |
| E8 | Councillor / MLA accountability page | 4 | 3 | 5 | v1+ | High political risk; strict sourcing rules |

## F. Trust, safety, and quality

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| F1 | Phone-number verification | 4 | 2 | 2 | v0.1 | Baseline sybil resistance |
| F2 | Trust score | 5 | 3 | 3 | v0.3 | Drives verification thresholds and rate limits |
| F3 | Corroboration-based verification | 5 | 3 | 2 | v0.2 | With a life-safety fast path |
| F4 | Device/IP/graph clustering for astroturf detection | 4 | 4 | 3 | v0.5 | Adversaries are certain |
| F5 | AI-assisted moderation queue | 4 | 3 | 3 | v0.3 | Human decides; AI ranks |
| F6 | Photo manipulation / reuse detection | 4 | 3 | 3 | v0.4 | Perceptual hash + EXIF + reverse-match |
| F7 | Right-of-reply workflow | 5 | 2 | 2 | v0.5 | Legal necessity once names are published |
| F8 | Public correction log | 5 | 1 | 1 | v0.5 | Cheapest, strongest good-faith defence |
| F9 | Takedown workflow + grievance officer | 5 | 2 | 1 | v0.5 | IT Rules 2021 compliance (36-hour window) |
| F10 | Rate limiting and abuse throttles | 4 | 2 | 1 | v0.1 | — |

## G. Power users

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| G1 | Watchdog TUI | 4 | 3 | 2 | v0.6 | See [Watchdog TUI](07-watchdog-tui.md) |
| G2 | Saved queries and alerts | 4 | 2 | 1 | v0.6 | The "set a trap" feature journalists want |
| G3 | Bulk export (CSV/GeoJSON/Parquet) | 4 | 2 | 2 | v0.5 | — |
| G4 | Document search over the scraped corpus (with OCR) | 5 | 4 | 3 | v0.7 | Turns the archive into a research tool |
| G5 | Anomaly alerts (unusual award patterns, ward-level spikes) | 3 | 4 | 3 | v1+ | — |
| G6 | Story-inquiry templates | 2 | 1 | 1 | v1+ | Nice, not load-bearing |

## H. Intelligence and prediction

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| H1 | Weather + sensor ingestion (MESONET, IMD, mumbaiflood.in) | 4 | 3 | 2 | v0.4 | Context and pre-warning |
| H2 | News ingestion and issue linking | 4 | 3 | 2 | v0.4 | Prevents duplicate panic-reporting |
| H3 | Ward-level sentiment/intensity from report text | 3 | 3 | 3 | v0.7 | Interesting; not decision-grade |
| H4 | Waterlogging hotspot prediction | 3 | 4 | 3 | v1+ | Needs several monsoons of data |
| H5 | Pothole degradation forecasting | 2 | 5 | 4 | v2+ | Requires longitudinal repeat imagery we won't have early |
| H6 | Bridge/structural health prediction from IoT | 1 | 5 | 4 | ✗ | We have no sensor access. Cut. |
| H7 | Repeat-failure detection (same 40 m, N times) | 5 | 1 | 1 | v0.3 | Trivial with the asset model, and devastating in a filing |

## I. Participation

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| I1 | Issue upvote / me-too | 3 | 1 | 2 | v0.3 | Weight by trust, not raw count |
| I2 | Evidence-only comments | 3 | 2 | 3 | v0.4 | Not a discussion forum |
| I3 | Gamification (points, badges, ward leaderboards) | 3 | 2 | 3 | v0.5 | Reward accuracy and confirmation, never volume |
| I4 | Participatory budgeting engine | 4 | 5 | 4 | v2+ | Pune precedent; needs ward-committee relationships |
| I5 | Deliberation with hard budget constraints | 3 | 5 | 3 | v2+ | With I4 |
| I6 | Volunteer verification missions ("re-check these 5 nearby") | 4 | 3 | 2 | v0.6 | Converts `claimed_resolved` → confirmed |

## J. Access and inclusion

| # | Feature | V | C | R | M | Notes |
|---|---|---|---|---|---|---|
| J1 | Marathi + Hindi UI | 5 | 2 | 1 | v0.2 | Non-negotiable in MMR |
| J2 | Gujarati UI | 3 | 1 | 1 | v0.5 | Significant MMR population |
| J3 | Bhashini ASR for voice reports | 4 | 3 | 3 | v0.4 | 22 languages; Indian-accent optimised |
| J4 | TTS read-back of official replies and legal text | 3 | 2 | 2 | v0.6 | Low-literacy access |
| J5 | WCAG 2.2 AA compliance | 4 | 2 | 1 | v0.3 | Also improves everything for everyone |
| J6 | Low-bandwidth / low-end device mode | 4 | 3 | 1 | v0.3 | Aggressive image compression, no heavy JS |

---

## Explicitly cut

| Feature | Why |
|---|---|
| IoT structural-health monitoring (accelerometers, strain gauges, LSTM/Autoencoder analysis) | We have no access to instrumented bridges. Building this is a hardware and government-partnership programme, not an app feature. |
| Bridge vertical-deflection prediction (B-IBk / hybrid ML models) | Same. Interesting research; not a product. |
| Blockchain anything | No trust problem here is solved by a distributed ledger that isn't better solved by a hash-chained append-only log and published source citations. |
| Anonymous whistleblower drop | Different threat model, different security posture. Would compromise the platform's public-evidence stance. |
| Real-time video streaming of issues | Bandwidth cost, moderation cost, marginal evidentiary gain over stills. |
| Political-party affiliation tagging | Guarantees capture and destroys neutrality. |

---

## Dependency graph (critical path to v1)

```
A1 ──▶ A2 ──▶ B1 ──▶ B5 ──▶ B8 ──▶ E1  ......................  v0.1 shippable
                │
                ├──▶ F1, F3, F10
                │
                └──▶ C1 ──▶ C3 ──▶ C4 ──▶ C5 ──▶ C6 ............  the moat
                              │
                              └──▶ D1 ──▶ D2 ──▶ D3 ──▶ D4 ──▶ D11 ..  consequence
                                                        │
                                                        └──▶ D9, D6
```

**Everything else is optional until C3 works.** Contract-to-geometry matching is the load-bearing
wall. If it cannot be made to work at usable accuracy, the product degrades to "a better complaint
app", and the roadmap should be re-planned around D-series escalation instead.
