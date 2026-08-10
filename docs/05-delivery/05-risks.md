# Risk register

Scored **L** (likelihood) × **I** (impact), 1–5 each. Ordered by score. Every risk has an owner, a
mitigation, and — where it exists — a trigger that says "stop and re-plan".

---

## Critical (score ≥ 16)

### R1 — Contract-to-geometry matching does not work at usable accuracy · L4 I5 = 20

Tender descriptions are written for humans who know the city. Geocoding "from Kandivali Station Road
junction to Charkop naka" reliably is genuinely hard, and it is the platform's differentiator.

**Mitigation:** v0.1 uses a hand-built contract dataset, so the *hypothesis* is tested before the
*engineering* is attempted. v0.2 has an explicit gate: ≥50% recall at ≥95% precision on one ward.
Precision is prioritised absolutely over recall.

**Trigger:** if the v0.2 gate fails, re-plan the roadmap around escalation (D-series) rather than
attribution, and write it down as an ADR. The product is still valuable; it is a different product.

---

### R2 — Defamation action over a published contractor name · L3 I5 = 15→ treat as critical

Naming firms is central to the value and is the largest legal exposure.

**Mitigation:** publication threshold gated on match confidence *and* geocode derivation; every
published fact carries a source and retrieval date; platform statements are visually distinct from
citizen content; documented right-of-reply, dispute, and takedown workflows; public correction log;
scorecard methodology published and versioned; IT Rules 2021 grievance officer appointed before
launch. Legal review of the scorecard methodology is a **blocking gate** for v0.5.

**Trigger:** a first notice triggers immediate legal review of the methodology, not a defensive
reaction to the specific record.

---

## High (score 10–15)

### R3 — Coordinated political fake reporting · L4 I3 = 12

**Mitigation:** trust scoring, device/IP/graph independence checks, corroboration requirements,
category-tiered verification, burst-mode handling for genuine mass events, moderation queue. See
[trust and anti-abuse](../02-product/08-trust-and-antiabuse.md).

**Trigger:** if astroturf detection cannot separate genuine corridor-wide events from coordinated
campaigns, raise verification thresholds and accept lower throughput rather than publish noise.

---

### R4 — Scrapers break and nobody notices · L5 I3 = 15

Government sites change layout without notice. A silently-failing scraper serves stale procurement
data as current — the most credibility-damaging failure available to us.

**Mitigation:** parse from the archive, not the network; row-count and field-fill-rate monitoring
with alerts; golden fixtures in CI; `reparse` capability; **freshness displayed publicly** on every
surface that shows contract data.

---

### R5 — Wrong jurisdiction routing destroys trust early · L3 I4 = 12

A citizen who sees a wrong ward once assumes everything else is wrong.

**Mitigation:** confidence gating with a single disambiguating question; golden evaluation set of 500
points including every hard case; wrong-at-high-confidence tracked as a near-zero target; road
ownership layer prioritised as a data-acquisition task.

---

### R6 — AI inference cost exceeds any sustainable budget · L3 I4 = 12

Classification is the dominant per-report cost.

**Mitigation:** Batch API for backfill (50%); prompt caching; **dedup before classify** (a viral
location produces ~1 classification, not 200); deterministic short-circuits for common chatbot
intents; model tiering; hard daily cap with graceful degradation to queue-and-classify-later.
Tracked in [cost model](../06-operations/04-cost-model.md).

---

### R7 — No adoption: citizens report once and never return · L3 I5 = 15

Every predecessor platform has a large install base and a small active base.

**Mitigation:** the share loop (distribution is the retention mechanism); the re-check notification
(gives the user a reason to return with a purpose); campaigns via RWA organisers; WhatsApp intake.
Exit criterion 5 for v0.1 exists precisely to detect this early.

**Trigger:** if repeat-report rate within 30 days is below 15% after v0.3, the problem is the product
loop, not the feature set. Stop adding features.

---

### R8 — Terms-of-use conflict with a critical data source · L3 I4 = 12

Mahatenders, MESONET, or a filing platform may prohibit automated access.

**Mitigation:** terms reviewed and recorded **before** any ingester is enabled; RTI as the fallback
acquisition channel; filing adapters degrade to "generate draft + deep link" where automation is not
permissible.

---

## Medium (score 6–9)

### R9 — Municipal hostility · L3 I3 = 9

**Mitigation:** the platform does what the Bombay High Court directed the state to do — frame
partnership conversations accordingly. Never shame individual officers for matters outside their
control. Correct routing and deduplication are genuinely useful to ward engineers. Publish a
transparency report so our own accountability claims are not hypocritical.

### R10 — DPDP non-compliance · L2 I5 = 10

**Mitigation:** the compliance checklist is a blocking pre-launch gate; consent scopes separable;
retention jobs with legal hold; redaction before public derivative; minimisation as the primary
control (we hold no name, no email, no government ID).

### R11 — Monsoon overload · L4 I2 = 8

**Mitigation:** durable-then-async writes; backpressure not errors; pre-scale in April; annual load
test at 20×; emergency mode; shed enrichment before intake.

### R12 — Single-maintainer bus factor · L4 I2 = 8

**Mitigation:** ADRs capture reasoning; `CLAUDE.md` makes the codebase legible to AI agents; runbooks
are executable commands not prose; everything is open source so the project survives the maintainer.

### R13 — Boundary data unavailable outside Greater Mumbai · L4 I2 = 8

DataMeet's MMR coverage is incomplete; several corporations publish only PDF maps.

**Mitigation:** digitisation is a scheduled task per corporation, sequenced with onboarding; a
corporation is not launched until its boundaries are verified.

### R14 — Contractor scorecard is statistically unfair · L3 I3 = 9

A firm with 200 contracts accumulates more raw defects than one with 5.

**Mitigation:** absolute counts displayed alongside the grade so the denominator is visible;
minimum-data thresholds before any grade is shown; the v1 formula is deliberately simple and marked
as such; normalisation by contract-kilometres is a research task, not a guess.

### R15 — Filing adapters break silently · L4 I2 = 8

**Mitigation:** capture and verify the official reference number as the success signal; alert on
adapter failure rate; degrade to draft + deep link; never mark an issue as filed without a reference.

---

## Low (score ≤ 5)

| # | Risk | Mitigation |
|---|---|---|
| R16 | Object-storage cost from media growth · L3 I1 | Lifecycle rules, aggressive derivative compression, retention schedule |
| R17 | Bhashini API instability · L3 I1 | Voice is an enhancement; text always works |
| R18 | Model provider changes behaviour · L2 I2 | Pinned model IDs, versioned prompts, eval sets in CI |
| R19 | Trademark or name conflict · L2 I2 | Search before public launch; the name is cheap to change now |

---

## Risks we are accepting without mitigation

| Risk | Why accepted |
|---|---|
| **The state builds its own version and it is better** | That would be the best possible outcome for citizens. The project's purpose is the outcome, not the platform. |
| **A commercial fork** | AGPL makes a closed hosted fork non-compliant; an open fork is fine. |
| **Reduced accuracy from phone-only identity** | Deliberate trade for reporter safety and inclusion. See [ADR 0011](../04-adr/0011-phone-only-identity.md). |
| **Lower escalation throughput from never auto-filing** | Deliberate. See [ADR 0007](../04-adr/0007-never-auto-file.md). |

---

## Review cadence

Reviewed monthly; scores updated; new risks added with an owner. A risk whose trigger fires goes to
the top of the next planning session regardless of its score.
