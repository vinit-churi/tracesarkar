# TraceSarkar

**Snap a photo of a civic problem. Get back who is responsible, what they were paid to do, and what you can do about it — in one tap.**

TraceSarkar is a civic accountability platform for the Mumbai Metropolitan Region (MMR). It is not
another grievance inbox. It is a machine that converts a single geotagged photograph into a complete,
citable, escalatable accountability record:

```
   [photo + GPS]
        │
        ▼
   ① CLASSIFY      what is this? (pothole / garbage / open manhole / illegal dumping / …)
   ② LOCATE        which ward, which road, which authority owns this asset?
   ③ ATTRIBUTE     which contract covers it, who won it, for how much, is it in warranty?
   ④ CONTEXTUALISE what else has been reported here? what did the news say? is it raining?
   ⑤ ACT           file it, tweet it, WhatsApp it, RTI it, escalate it, track it
```

Every step is a first-class, independently useful capability. Steps ③ and ⑤ are the ones nobody
else in India has built, and they are the reason this project exists.

---

## Why this, why now

Three facts make TraceSarkar buildable today in a way it was not five years ago:

1. **The Bombay High Court has made civic negligence legally expensive.** In
   *High Court on its own motion v. State of Maharashtra* (13 October 2025), a Division Bench held
   that the right to safe roads is a fundamental right under Article 21, fixed ₹6,00,000 compensation
   for pothole deaths and ₹50,000–₹2,50,000 for injuries, mandated that reported potholes be attended
   to **within 48 hours**, and ordered blacklisting and personal liability for erring officials and
   contractors. Structured, timestamped, geotagged citizen evidence is now directly convertible into
   legal consequence. See [`docs/01-research/06-key-judgments.md`](docs/01-research/06-key-judgments.md).

2. **Procurement data is public but unusable.** Mahatenders, the BMC portal, and the CPP Portal all
   publish tender and award data. None of it is joined to geography, to a warranty clock, or to the
   citizen standing on the broken road. That join is a database problem, not a policy problem.

3. **Vision-language models made classification and evidence extraction cheap.** What required a
   bespoke CV pipeline and a labelled dataset in 2020 is now a single API call with a schema.

The gap in MMR governance is not data. It is **synthesis and consequence**.

---

## What makes it different

| Everyone else | TraceSarkar |
|---|---|
| "Your complaint has been registered" | "Your complaint has been registered **against Contract WS/2023/ROAD/117, awarded to X for ₹4.1 crore, whose defect liability period runs to 2028**" |
| Routes to one municipal body | Resolves overlapping BMC / MMRDA / PWD / MSRDC / Railway / MIDC / CIDCO jurisdiction and routes correctly |
| Complaint dies at SLA breach | SLA breach auto-generates an RTI; RTI silence auto-generates a First Appeal; the paper trail auto-assembles into an NGT / Lokayukta / Consumer Commission brief |
| Complaint is private | Complaint produces a share kit: tagged tweet, WhatsApp card, ward heatmap embed |
| Dashboard for officials | Terminal UI for journalists and activists, plus a public read-only API |

---

## Repository map

This repository is currently **documentation and design**. Code lands per the roadmap.

| Path | What's in it |
|---|---|
| [`docs/00-overview/`](docs/00-overview/) | Vision, problem statement, glossary, decision log |
| [`docs/01-research/`](docs/01-research/) | MMR jurisdiction map, legal framework, data sources, key judgments, global + Indian landscape |
| [`docs/02-product/`](docs/02-product/) | Personas, full feature catalog, user flows, the snap-to-action pipeline, trust & anti-abuse, share kit, chatbot, Watchdog TUI |
| [`docs/03-architecture/`](docs/03-architecture/) | System overview, data model, API design, AI pipeline, jurisdiction engine, tender engine, ingestion, deployment, security |
| [`docs/04-adr/`](docs/04-adr/) | Architecture Decision Records — why Go, why Postgres+PostGIS, why H3, why this AI stack |
| [`docs/05-delivery/`](docs/05-delivery/) | Roadmap, MVP definition, backlog, metrics, risk register, go-to-market |
| [`docs/06-operations/`](docs/06-operations/) | Runbook, moderation policy, legal review checklist, cost model |
| [`docs/templates/`](docs/templates/) | RTI application, NGT original application, Lokayukta complaint, consumer complaint templates |

**Start here:** [`docs/00-overview/01-vision.md`](docs/00-overview/01-vision.md) →
[`docs/02-product/04-snap-to-action-pipeline.md`](docs/02-product/04-snap-to-action-pipeline.md) →
[`docs/03-architecture/01-system-overview.md`](docs/03-architecture/01-system-overview.md).

**If you only read one page:** [`docs/05-delivery/02-milestone-v0-mvp.md`](docs/05-delivery/02-milestone-v0-mvp.md).

---

## Architecture in one paragraph

A Go backend (goroutines for concurrent ingestion and WebSocket fan-out) fronts PostgreSQL 16 with
PostGIS for authoritative geometry and H3 for fast bucketing. Media goes to S3-compatible object
storage. A worker pool handles the AI pipeline (Claude vision for classification + evidence
extraction, structured outputs for schema guarantees, Batch API for backfills). Scrapers pull
tenders, news, weather, and sensor feeds on schedules into a normalised `contracts` /
`news_items` / `observations` model. Everything is containerised and orchestrated on a Docker
Swarm cluster of VPS nodes managed by Dokploy. See
[`docs/03-architecture/01-system-overview.md`](docs/03-architecture/01-system-overview.md).

---

## Scope discipline

The temptation with this idea is to build all thirty features at once and ship none of them. The
MVP is deliberately narrow:

> **v0.1 — one ward, one issue type, one authority.** Photo → classify → resolve ward → match to
> road contract → produce a filable complaint + a shareable card. Nothing else.

Everything in the feature catalog is a *later* milestone, and each is gated on the previous one
producing real usage. See [`docs/05-delivery/01-roadmap.md`](docs/05-delivery/01-roadmap.md).

---

## Status

| | |
|---|---|
| **Stage** | Design / pre-alpha |
| **Region** | Mumbai Metropolitan Region (Greater Mumbai first) |
| **Backend** | Go (planned) |
| **Docs last reviewed** | 2026-08-10 |

---

## Contributing

Read [`CONTRIBUTING.md`](CONTRIBUTING.md). This project touches law, personal data, and named private
parties — read [`docs/06-operations/03-legal-review-checklist.md`](docs/06-operations/03-legal-review-checklist.md)
before contributing anything that publishes a name.

## Licence

Code: **AGPL-3.0-or-later**. Documentation and datasets: **CC BY-SA 4.0**.
Rationale in [`docs/04-adr/0012-licensing.md`](docs/04-adr/0012-licensing.md).
