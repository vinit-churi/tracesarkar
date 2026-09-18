# Documentation index

Everything in this repository, in reading order.

---

## Start here

| # | Document | What it answers |
|---|---|---|
| 0 | [Project state](00-overview/06-project-state.md) | **Where things stand right now, and what to do next** |
| 1 | [Vision](00-overview/01-vision.md) | What are we building and why |
| 2 | [Problem statement](00-overview/02-problem-statement.md) | What is actually broken in MMR |
| 3 | [Snap-to-action pipeline](02-product/04-snap-to-action-pipeline.md) | How the product works, end to end |
| 4 | [v0.1 MVP](05-delivery/02-milestone-v0-mvp.md) | What we build first, and what "done" means |
| 5 | [System overview](03-architecture/01-system-overview.md) | How it is built |

---

## 00 — Overview

| Document | Contents |
|---|---|
| [Vision](00-overview/01-vision.md) | The three layers (synthesis, consequence, leverage), design principles, non-goals, success conditions |
| [Problem statement](00-overview/02-problem-statement.md) | Jurisdictional fragmentation, grievance-vs-asset modelling, procurement opacity, why predecessors under-delivered |
| [Glossary](00-overview/03-glossary.md) | Canonical terms, statuses, MMR authorities, legal instruments, conventions |
| [Naming and brand](00-overview/04-naming-and-brand.md) | The name, tone, and what the platform never says |
| [Decision log](00-overview/05-decision-log.md) | Chronological record of decisions, with pointers to ADRs |
| [Project state](00-overview/06-project-state.md) | Snapshot: goal, current phase, decisions to confirm, next actions, how to resume |

## 01 — Research

| Document | Contents |
|---|---|
| [Landscape: India](01-research/01-landscape-india.md) | Swachhata, IChangeMyCity, MyBMC, Aaple Sarkar, RailMadad, BBMP, open contracting — and where each stops |
| [Landscape: global](01-research/02-landscape-global.md) | FixMyStreet, SeeClickFix, Open311, g0v, VoiceCast, participatory budgeting, CPARS — what transfers and what doesn't |
| [MMR jurisdiction map](01-research/03-mmr-jurisdiction-map.md) | The 9 corporations, 9 councils, parastatals; the hard cases; boundary data sources; the confidence model |
| [Legal framework](01-research/04-legal-framework.md) | RTI, RTS Act, NGT, Lokayukta, e-Jagriti, DPDP, defamation and intermediary liability; the instrument decision table |
| [Data sources](01-research/05-data-sources.md) | Every external feed, with acquisition strategy, format, cadence, and risk |
| [Key judgments](01-research/06-key-judgments.md) | The Bombay HC 2025 pothole directions and their direct product consequences |
| [Open questions](01-research/07-open-questions.md) | What we don't know, ordered by how much the answer changes the plan |
| [Data availability audit](01-research/08-data-availability-audit.md) | **Aug 2026 field survey** — what public data actually exists, verified by direct fetch, with licence and terms verdicts |
| [Vertical exploration](01-research/09-vertical-exploration.md) | **Sep 2026 survey** — new issue domains, datasets, geographies and personal tools, scored and provenance-marked; corrections to the audit |

## 02 — Product

| Document | Contents |
|---|---|
| [Personas and JTBD](02-product/01-personas-and-jtbd.md) | Five users, their jobs, and the anti-personas we design against |
| [Feature catalog](02-product/02-feature-catalog.md) | Every feature considered, scored, milestoned — including what we cut |
| [User flows](02-product/03-user-flows.md) | Screen-by-screen for the flows that matter |
| [Snap-to-action pipeline](02-product/04-snap-to-action-pipeline.md) | The seven-stage core loop |
| [Share kit](02-product/05-share-kit.md) | The distribution engine |
| [Chatbot](02-product/06-chatbot.md) | Grounded retrieval, not an oracle |
| [Watchdog TUI](02-product/07-watchdog-tui.md) | The journalist's terminal |
| [Trust and anti-abuse](02-product/08-trust-and-antiabuse.md) | Threat model, trust score, verification policy, image integrity, disputes |
| [Gamification](02-product/09-gamification.md) | Rewarding accuracy and closure, never volume |
| [Accessibility and vernacular](02-product/10-accessibility-and-vernacular.md) | Languages, voice, low bandwidth, WCAG, literacy |
| [Screen specification](02-product/11-screen-spec.md) | The authoritative screen-by-screen definition: global rules, design tokens, components, 34 screens |
| [UI generation guide](02-product/12-ui-generation-guide.md) | Style capsule, copy deck and per-screen prompts for producing screen imagery |
| [Data-unlocked features](02-product/13-data-unlocked-features.md) | Series K — features grounded in sources verified by the August 2026 audit, and the ones the audit killed |
| [Civic utility features](02-product/14-civic-utility-features.md) | Series U — the surfaces that earn a weekly open: deadline wallet, warranty alerts, works near me, claim helper, evidence vault |

## 03 — Architecture

| Document | Contents |
|---|---|
| [System overview](03-architecture/01-system-overview.md) | Services, write path, read path, storage, concurrency, degradation |
| [Data model](03-architecture/02-data-model.md) | The full schema, in SQL |
| [API design](03-architecture/03-api-design.md) | Core API, Open311 surface, public data API, realtime |
| [AI pipeline](03-architecture/04-ai-pipeline.md) | Every model call, its schema, cost control, and evaluation |
| [Jurisdiction engine](03-architecture/05-jurisdiction-engine.md) | Resolving "whose is this?" with confidence and provenance |
| [Tender engine](03-architecture/06-tender-engine.md) | Follow the money: ingest → extract → geocode → match → score |
| [Ingestion and scrapers](03-architecture/07-ingestion-and-scrapers.md) | The framework, politeness policy, and the archive |
| [Realtime and notifications](03-architecture/08-realtime-and-notifications.md) | WebSocket fan-out, notification hygiene, the deadline scheduler |
| [Deployment](03-architecture/09-deployment.md) | Docker Swarm, Dokploy, monsoon capacity, backup and recovery |
| [Observability](03-architecture/10-observability.md) | Metrics, logging rules, alerting, SLOs, the public dashboard |
| [Security and privacy](03-architecture/11-security-and-privacy.md) | Data inventory, DPDP compliance, location privacy, incident response |

## 04 — Architecture Decision Records

| ADR | Decision |
|---|---|
| [0001](04-adr/0001-record-architecture-decisions.md) | Record architecture decisions |
| [0002](04-adr/0002-go-backend.md) | Go for the backend |
| [0003](04-adr/0003-postgres-postgis.md) | PostgreSQL + PostGIS as the single source of truth |
| [0004](04-adr/0004-h3-indexing.md) | H3 for bucketing, never for jurisdiction |
| [0005](04-adr/0005-no-orm.md) | Hand-written SQL, no ORM |
| [0006](04-adr/0006-claude-for-vision-and-extraction.md) | Claude models with structured outputs |
| [0007](04-adr/0007-never-auto-file.md) | Never auto-file a legal instrument |
| [0008](04-adr/0008-redaction-approach.md) | On-device-first redaction |
| [0009](04-adr/0009-timeseries.md) | Plain Postgres partitioning for observations |
| [0010](04-adr/0010-open311-compatibility.md) | Expose an Open311 surface |
| [0011](04-adr/0011-phone-only-identity.md) | Phone-only identity; no Aadhaar |
| [0012](04-adr/0012-licensing.md) | AGPL-3.0 code, CC BY-SA 4.0 data |
| [0013](04-adr/0013-political-accountability-scope.md) | What may be published about elected representatives and candidates |
| [0014](04-adr/0014-assisted-filing-not-automated-submission.md) | Assisted filing on the citizen's device; never server-side submission or credential custody |

## 05 — Delivery

| Document | Contents |
|---|---|
| [Roadmap](05-delivery/01-roadmap.md) | Phases 0–6 with exposure tiers, falsifiable exit criteria and stop-and-replan gates |
| [v0.1 MVP](05-delivery/02-milestone-v0-mvp.md) | Scope, build order, blocking checklist, exit criteria |
| [Backlog](05-delivery/03-backlog.md) | Concrete tasks, ordered |
| [Metrics](05-delivery/04-metrics.md) | North star, accountability metrics, funnel, anti-metrics |
| [Risks](05-delivery/05-risks.md) | Scored register with triggers |
| [GTM and partnerships](05-delivery/06-gtm-and-partnerships.md) | Distribution thesis, launch sequence, sustainability |
| [Phase 0 — instruments](05-delivery/07-phase-0-instruments.md) | The first build: archive, snapshotter, watchers, capture tools, exit criteria |

## 06 — Operations

| Document | Contents |
|---|---|
| [Runbook](06-operations/01-runbook.md) | Alert response, deployment, backup, monsoon prep, incident response |
| [Moderation policy](06-operations/02-moderation-policy.md) | Content rules, the queue, disputes, takedowns, transparency |
| [Legal review checklist](06-operations/03-legal-review-checklist.md) | Blocking gates by feature type and milestone |
| [Cost model](06-operations/04-cost-model.md) | Unit economics, AI cost dominance, budget scenarios |

## Templates

| Template | Instrument |
|---|---|
| [RTI application](templates/rti-application.md) | Right to Information, with question sets by category |
| [NGT Original Application](templates/ngt-original-application.md) | National Green Tribunal, Western Zone |
| [Lokayukta complaint](templates/lokayukta-complaint.md) | Maharashtra Lokayukta, with the affidavit |
| [Consumer complaint](templates/consumer-complaint-ejagriti.md) | e-Jagriti, deficiency in service |
| [HC compensation claim](templates/hc-compensation-claim.md) | Bombay HC pothole compensation framework |

---

## Conventions in these documents

| Marker | Meaning |
|---|---|
| ⚠️ | Unverified — must be checked against a primary source before it is relied on |
| **Gate** | A decision point that can stop or re-plan the roadmap |
| "Open question" | Genuinely undecided; see [open questions](01-research/07-open-questions.md) |

Documents are Markdown, wrapped at 100 columns, cross-linked with relative paths. Every architectural
decision has an ADR. Every external data source has an entry in the source register.
