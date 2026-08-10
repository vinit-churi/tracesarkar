# Vision

> Governance failure in the Mumbai Metropolitan Region is not caused by a shortage of data.
> It is caused by a shortage of **synthesis** and a shortage of **consequence**.

## The one-sentence version

TraceSarkar turns a photograph of a civic failure into an accountability record that names the
responsible authority, the responsible contract, and the responsible money — and then hands the
citizen the specific next action that makes ignoring it expensive.

## The three layers

### Layer 1 — Synthesis (make the invisible visible)

Today, five facts about a single pothole live in five disconnected places:

| Fact | Where it lives today |
|---|---|
| That the pothole exists | Nowhere, until a citizen complains |
| Which ward / authority owns that stretch of road | A PDF ward map, an ArcGIS layer, and institutional memory |
| Which contract covers its construction | Mahatenders / BMC tender portal, keyed by tender ID, not by geography |
| Whether it is inside its defect liability period | The contract document, if you can find it |
| Whether the same contractor has failed elsewhere in MMR | Nowhere — each corporation keeps its own list |

TraceSarkar's first job is the **join**. Geometry is the join key: every asset, every ward
boundary, every contract work-site, every complaint and every sensor reading is a point or polygon
in the same spatial database. Once joined, the pothole stops being an anonymous inconvenience and
becomes a named breach of a specific, priced obligation.

### Layer 2 — Consequence (make ignoring it expensive)

Synthesis without consequence is a dashboard, and Indian civic tech is littered with dashboards.
Consequence in the Indian legal system is a sequence, and the sequence is mechanical:

```
complaint  ──SLA breach──▶  RTI  ──no reply in 30d──▶  First Appeal  ──▶  State Information Commission
    │                        │
    │                        └──documents show funds released for undelivered work──▶  Lokayukta
    │
    ├──environmental harm (mangroves, creeks, debris, trees)──▶  NGT (Western Zone, Pune)
    ├──service deficiency to a tax-paying consumer──▶  District Consumer Commission (e-Jagriti)
    └──death or injury from a road defect──▶  compensation claim under the Bombay HC 2025 directions
```

Every one of those transitions is (a) deterministic, (b) driven by a timestamp, and (c) currently
performed by hand by the tiny minority of citizens who know it exists. TraceSarkar automates the
paperwork and the clock. It does not — and must not — file anything without an explicit human
instruction. See [`docs/06-operations/03-legal-review-checklist.md`](../06-operations/03-legal-review-checklist.md).

### Layer 3 — Leverage (make one citizen count like a thousand)

A single complaint is ignorable. What is not ignorable:

- A **contractor scorecard** aggregated across all nine MMR corporations, showing that the firm that
  is currently the lowest bidder in Thane has 41 in-warranty defects in Kandivali.
- A **ward heatmap** embedded live in a news broadcast during monsoon.
- A **share kit** that turns a complaint into a pre-written, correctly-tagged post that a citizen
  actually sends, because writing it themselves is the friction that kills 90% of intent.
- A **journalist terminal** where a reporter can ask "show me every completed contract in P-North
  with zero photographic evidence of work in the last 24 months" and get an answer in two seconds.

## Design principles

1. **Evidence over opinion.** Every claim the platform makes is traceable to a source: a document, a
   coordinate, a timestamp, a photograph. If we can't cite it, we don't display it as fact.
2. **The photograph is the API.** The citizen's only obligation is to point and shoot. Everything
   else — jurisdiction, contract, code section, phone number, hashtag — is our job.
3. **Names of officials and firms are published only from public records.** We publish what the
   government published. We annotate, we join, we never allege. See the
   [legal review checklist](../06-operations/03-legal-review-checklist.md).
4. **Nothing is filed on a citizen's behalf without their explicit act.** Auto-generate; never
   auto-submit.
5. **Assume adversaries.** Political fake-reporting, contractor astroturfing, and coordinated
   downvoting are certainties, not risks. Trust design is a v1 feature, not a v3 feature. See
   [`docs/02-product/08-trust-and-antiabuse.md`](../02-product/08-trust-and-antiabuse.md).
6. **Vernacular first, not vernacular later.** A platform that only works in English is a platform
   for the people who already have access to officials.
7. **Degrade gracefully.** Monsoon is exactly when the network fails and exactly when the platform
   matters most. Offline capture with deferred sync is a requirement, not a nice-to-have.
8. **Boring infrastructure.** One language, one database, one orchestrator. The novelty budget is
   spent entirely on the domain, not the stack.

## Non-goals

TraceSarkar is explicitly **not**:

- A replacement for MyBMC / Aaple Sarkar / RailMadad. It is an **overlay** that files into them and
  independently tracks the lifecycle they hide.
- A political campaigning tool. Party-affiliated framing is out of scope and is a moderation
  violation.
- A predictor of guilt. Correlating a contract with a defect is a factual join; asserting corruption
  is a legal conclusion we do not make.
- A general-purpose social network. Comments exist to add evidence, not to argue.
- An anonymous whistleblower drop. That is a different threat model and needs a different product.

## What success looks like

| Horizon | Success condition |
|---|---|
| 6 months | 1 ward, 1,000 verified reports, ≥1 documented case where the tender attribution changed the outcome |
| 12 months | Greater Mumbai coverage; a journalist publishes a story sourced from the Watchdog TUI |
| 18 months | Contractor scorecard cited in a municipal standing-committee discussion or a court filing |
| 24 months | A municipal corporation consumes the public API or requests an official integration |

The metric we do **not** optimise is total complaints filed. Volume without resolution is the exact
failure mode of every predecessor platform. See [`docs/05-delivery/04-metrics.md`](../05-delivery/04-metrics.md).

## Related reading

- [Problem statement](02-problem-statement.md)
- [Snap-to-action pipeline](../02-product/04-snap-to-action-pipeline.md)
- [Roadmap](../05-delivery/01-roadmap.md)
