# Roadmap

**Governing principle:** every phase must produce something a real person uses, and every phase is
gated on the previous one producing evidence — not on the previous one merely shipping.

The failure mode this roadmap is designed to avoid: building thirty features, shipping none of them
well, and discovering after a year that the contract-matching engine never worked.

**Restructured 18 September 2026** ([D047](../00-overview/05-decision-log.md)). Two things changed:
the [vertical exploration](../01-research/09-vertical-exploration.md) found BMC works data that
turns much of attribution into a spatial join, and personal instruments now come before anything
public. Phases keep their old version labels so the `M` column in the
[feature catalog](../02-product/02-feature-catalog.md) still reads correctly. Where this page and a
catalog milestone disagree, this page wins.

Every feature carries an **exposure tier** — `personal`, `flagged` or `public`
([D044](../00-overview/05-decision-log.md)). A feature moves up a tier deliberately, and that move
is where the hard rules in [`CLAUDE.md`](../../CLAUDE.md) are checked.

---

## How the phases actually run

**Re-baselined 25 September 2026 ([D055](../00-overview/05-decision-log.md)).** The original plan
was serial — research, then build, then launch — written before the collector existed. It does not
need to be. Three tracks run at the same time:

| Track | Owner | Status |
|---|---|---|
| **Collection** | Nobody. It runs twice a day by itself | Running since 25 Sep 2026 |
| **Backend** | Engineering. Nothing about it waits on collection or fieldwork | Started 25 Sep 2026 |
| **Fieldwork** | The maintainer: photographs, ward points, contract documents, the RTI | Whenever available |

A phase is finished when all three tracks have delivered what it needs, not when the calendar says
so. The dates below are therefore estimates of **when the slowest track lands**, and the slowest
track is usually fieldwork — not code.

One item has a clock somebody else controls: an RTI's reply window is 30 days from filing. Filing
early costs nothing and takes it off the critical path.

## At a glance

| Phase | Label | Tier | Target | Proves |
|---|---|---|---|---|
| 0 | Instruments | personal | Nov 2026 | The data can be collected and kept |
| 1 | v0.1 | public, R/S ward | Feb 2027, and movable | People report |
| 2 | v0.2 | public | **Apr 2027**, before the monsoon | The money can be joined, and drains follow roads |
| 3 | v0.3 | public + flagged | Jun 2027 | Consequence is generated |
| 4 | v0.4 | public | late 2027 | Escalation compounds |
| 5 | v0.5–v0.6 | public + flagged | 2028 | It generalises to a second geography and survives scrutiny |
| 6 | v0.7+ | — | — | Depth |

The calendar is anchored to the **2027 monsoon**. BMC's desilting deadline is 31 May and its C1
dangerous-buildings list appears around 30 April, so drains and building safety must be live before
then or they wait a year.

---

## Phase 0 — Instruments · personal · Nov 2026

Full specification: [Phase 0](07-phase-0-instruments.md).

| Deliverable | Detail |
|---|---|
| Foundations | Go scaffold, ingester framework, write-once archive on R2, Postgres + PostGIS on a single-node staging VPS |
| P1 works snapshotter | Daily hash-checked snapshots of BMC's roads and storm-water drain APIs, with field-level change logs |
| P2 blocker watchers | New Maharashtra GRs and Bombay HC judgments, classified and alerted |
| P3 manual-capture extension | Hand retrieval into the archive, with hashes, for sources that forbid automation |
| P4 field kit | Personal build of the v0.1 capture flow, used to collect the evaluation sets |
| P5 RTI tracker | Your own filings, with clocks from `legal_constants` |
| P8 alerts | One channel for all of the above |
| Documentation drift | The gaps the September review found, as backlog tasks |

Why each of these exists, and what breaks without it, is in
[Phase 0 §10A](07-phase-0-instruments.md#10a-why-these-criteria-and-what-each-one-buys).

**Exit criteria — all must hold:**

- 30 consecutive days of P1 snapshots with no missed run
- 500 labelled road-defect photographs and 200 golden jurisdiction points for R/S ward
- At least 20 R/S contract documents captured through P3
- One RTI filed and tracked through P5, which settles the fee question (Q2)

---

## Phase 1 · v0.1 — One ward, one category, one authority · public · Feb 2027

**Scope:** Kandivali West (BMC R/S ward) · `road_defect` only · BMC only. Full detail in the
[v0.1 MVP](02-milestone-v0-mvp.md).

| Deliverable | Detail |
|---|---|
| Photo capture with geotag | The PWA from P4, opened to the public |
| Classification | Claude vision, structured output, road-defect taxonomy |
| Jurisdiction | R/S ward boundary + BMC department mapping + confidence gate |
| Contract attribution | **Spatial join to BMC's roads API** for the 118 CC-road works in R/S (K1). Other roads get the coverage-honesty line (K7), backed by the hand-built dataset from P3 where it exists |
| Who do I call (U8) | Ward, department, engineer designation, helplines |
| SLA clock | 48 h, from the Bombay HC direction, in `legal_constants` |
| Share kit | Annotated image + text (en/mr) + permalink |
| Trust | Phone OTP, rate limits, basic corroboration |
| Public issue pages | Permalink with OG cards |

**Exit criteria — all must hold:**

- 200 real reports from people who are not the team
- Jurisdiction accuracy ≥ 95% on the golden set
- Classification accuracy ≥ 90% on the eval set
- ≥ 30% of verified issues produce a share
- **≥ 1 documented case where the contract attribution changed what the citizen did**

That last criterion is the whole hypothesis. If reporting and share rates are unchanged with
attribution vs without, the thesis is wrong and the roadmap needs re-planning.

**Closes by the end of this phase:** Q35 — partner with Pothole Reporter, or compete.

---

## Phase 2 · v0.2 — Money, and drains before the monsoon · public · Apr 2027

| Deliverable | Detail |
|---|---|
| **BMC says complete vs a dated citizen photo (K2)** | BMC's completion photograph and date beside the citizen's later one |
| **Public change log (U14, K6)** | Published from P1's archive, which by then holds six months of history |
| **Drains and flooding (V1)** | SWD desilting API, 14,220 historical waterlogging incidents, the nallah-to-agency join. Live before the 31 May desilting deadline |
| Works near me (U5) | The retention surface, from the roads API |
| Non-CC attribution | Work-site geocoding against the hand-built dataset, with publication thresholds |
| Filing drafts (K30) | Prefilled drafts and deep links for MyBMC; reference-number capture |
| Offline capture | Deferred sync |
| Redaction | On-device + server fallback |
| Marathi + Hindi UI | Full |

**Gate (explicit, scheduled):** for roads **outside** the CC programme, if work-site geocoding
cannot reach **≥ 50% recall at ≥ 95% precision** on R/S ward, stop and re-plan around escalation
(D-series) rather than attribution. For CC roads the question is already answered by the roads
API. This decision point has a date, an owner, and a written outcome.

---

## Phase 3 · v0.3 — Consequence, and permits behind a flag · public + flagged · Jun 2027

| Deliverable | Tier | Detail |
|---|---|---|
| **RTI generation** | public | Correct PIO, category-specific questions, fee settled by P5 |
| Deadline wallet (U1) | public | Every clock the citizen is inside |
| SLA breach detection and notification | public | The action prompt |
| Repeat-failure detection | public | Same asset, N times |
| WhatsApp intake | public | The distribution unlock |
| Trust score, moderation queue | public | AI-ranked, human-decided |
| Categories expanded | public | `waste`, `water_drainage`, `street_furniture` |
| Wards expanded | public | All of BMC R-zone |
| **Construction sites (V2)** | flagged | MahaRERA point-to-project join; dust, debris and breeding reports pre-attributed |
| **Building safety (V3)** | flagged | The C1 list (published around 30 April), collapse history, s.353B RTI drafts |
| **Hoardings (V4)** | flagged | Permit and renewal status per hoarding |
| Research terminal (P6) | flagged | For invited journalists, on the local database |

**Closes by the end of this phase:** Q34 — whether the second geography is Pune or an MMR
corporation, decided on its works-to-location data, its open GIS, and need.

---

## Phase 4 · v0.4 — Escalate · public · late 2027

| Deliverable | Detail |
|---|---|
| **Compensation claim via the DLSA committee (D9, U13)** | The administrative route the Bombay HC order created, not a High Court filing |
| Escalation ladder auto-advance (drafts) | With the deadline scheduler |
| RTI first appeal | Clock + template |
| Trees (V5) | Fall history, and the s.8 objection window as a deadline-wallet clock once read from the primary text |
| Evidence vault (U12) | Citable export with a provenance manifest |
| Warranty watch (U2), honest mode | Completion date, the Court's expectation, and the PWD GR as a labelled reference |
| News and context | Duplicate-panic suppression |
| Transit Mode | Railway geofence + RailMadad-compatible fields |
| Photo-integrity checks | phash, EXIF, reuse detection |
| Voice reporting | **Blocked** until Bhashini production terms are settled (Q31) |

---

## Phase 5 · v0.5–v0.6 — A second geography, and aggregation · 2028

| Deliverable | Tier | Detail |
|---|---|---|
| **Second geography** | public | Pune or KDMC, per Q34 |
| Public read-only API | public | Open311 read side + data API |
| Bulk export | public | CSV / GeoJSON / Parquet with provenance manifests |
| Wards expanded | public | All of Greater Mumbai |
| Right of reply, correction log, takedown workflow | public | Required before any named-party surface goes public |
| Contractor record | flagged → public | Methodology, minimum-data thresholds, dispute channel, external review before public |
| Corporator lookup (K19a) | flagged | From the 2026 election gazette; the election-period freeze ships with it |
| Watchdog terminal, saved alerts | public | The research terminal, opened up |
| Verification missions, weekly ward digests | public | — |
| NGT application compiler, RTS appeals | public | RTS depends on Q33 |

**Legal review gate:** no named-party aggregate moves from `flagged` to `public` without a
documented external review of its methodology and publication thresholds.

---

## Phase 6 · v0.7+ — Depth

Document search over the archive · OCDS publication of the contracts corpus · Lokayukta packet ·
e-Jagriti · water supply (V6) · public toilets (V7) · Gujarati UI · ward intensity from report text.

**v1.0 is defined by coverage and reliability, not by new features:** all nine MMR corporations,
filing drafts for the major channels, the escalation ladder complete, the public API stable, a
quarterly transparency report.

---

## Beyond v1

Ordered by expected value, not by excitement:

| Item | Notes |
|---|---|
| Corporate lineage mapping (DIN-based) | High value, high legal care required |
| Payment vs milestone reconciliation | Needs RTI-obtained payment records |
| Tender text analytics (restrictive specifications) | Descriptive observations only |
| Representative record pages | Scope fixed by [ADR 0013](../04-adr/0013-political-accountability-scope.md) |
| Participatory budgeting engine | Needs ward-committee relationships; Pune precedent |
| Anomaly detection on award patterns | — |
| Satellite change detection | **Only** for large sites. Sentinel-2 at 10 m cannot resolve road works — see [open questions](../01-research/07-open-questions.md) |
| Self-hosted VLM | Revisit when inference cost justifies GPU operations |

**Explicitly not planned:** IoT structural-health monitoring, bridge-deflection prediction,
blockchain anything, anonymous whistleblower drop. See
[feature catalog §Explicitly cut](../02-product/02-feature-catalog.md#explicitly-cut).

**Parked verticals:** noise, streetlights and dark spots, hawkers — see
[vertical exploration §2](../01-research/09-vertical-exploration.md#2-issue-domain-verticals).

---

## Sequencing logic

```
Phase 0  prove the data can be collected and kept   ─┐
Phase 1  prove people report                          ├─ if any fails, the product is different
Phase 2  prove the money can be joined               ─┤
Phase 3  prove consequence is generated              ─┘
Phase 4  prove escalation compounds
Phase 5  prove it generalises, and that aggregation survives scrutiny
Phase 6  prove depth
```

Each row is falsifiable. Each has an exit criterion that can fail.

---

## Standing pre-monsoon task (every April)

Independent of phase:

- [ ] Load test at 20× baseline
- [ ] Pre-scale the cluster
- [ ] Verify offline capture and sync end to end
- [ ] Emergency-mode drill
- [ ] Refresh ward boundary and contract data
- [ ] Capture the C1 list and the desilting-season API path the day they appear
- [ ] Verify the deadline scheduler and its alert
- [ ] Restore test

Monsoon is when the platform matters most and when everything is most likely to break.
