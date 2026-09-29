# v0.1 — the MVP, in full detail

**If you read one page in this repository, read this one.**

> **Amended 18 September 2026.** v0.1 is now Phase 1 of the [roadmap](01-roadmap.md) and follows
> [Phase 0](07-phase-0-instruments.md), which delivers the scaffold, the archive, the evaluation
> sets and the hand-built contract documents. Attribution for concretised roads now uses BMC's own
> roads API (K1). **Amended 29 September 2026:** the ward is **R/C, Borivali**, not R/S. Two
> reasons, and the second only became checkable once collection was running. The maintainer lives
> in Borivali, and Phase 0's photographs and ground-truth points have to be gathered on foot — a
> ward nobody walks through produces no data. And measured against the 4,673 works actually
> collected, **R/C carries 199 road works, 198 of them with geometry — the most of any ward in
> Mumbai**, ahead of H/W (161) and well ahead of R/S (118). Borivali is both the reachable ward and
> the densest one. The hand-built dataset covers roads outside the CC programme.

---

## 1. The one-sentence scope

> A resident of Borivali photographs a pothole; within six seconds they see which BMC ward and
> department owns it, which contract covers that stretch and whether it is still under warranty, and
> a ready-to-send share card — and the platform tracks the 48-hour clock the Bombay High Court
> imposed.

Nothing else.

---

## 2. Boundaries

| In | Out |
|---|---|
| BMC R/C ward (Borivali) | Every other ward, every other corporation |
| `road_defect` category | Waste, drainage, environment, transit, structural |
| BMC as the only authority | MMRDA, PWD, MSRDC, Railways |
| Web PWA | Native apps |
| English + Marathi | Hindi, Gujarati |
| Roads-API attribution (K1) + hand-built dataset for other roads | Automated tender ingestion |
| Share kit + permalink | Filing adapters, RTI, escalation |
| Email/password + Google, **and** phone OTP before anything publishes ([ADR 0017](../04-adr/0017-email-and-google-identity-before-public-tier.md)) | Full trust score |
| An operator console: moderation queue, collection health, metrics | Multi-operator roles and permissions |

Every "out" item is v0.2 or later and is listed in the [roadmap](01-roadmap.md).

---

## 3. Why this scope

- **One ward** makes the boundary data problem tractable: one polygon to verify, one road inventory
  to obtain, one gazetteer to build.
- **One category** makes the classification eval set achievable: 500 labelled pothole/road-defect
  photos, not 5,000 across nine categories.
- **One authority** makes the department mapping, SLA, and PIO research a day's work rather than a
  quarter's.
- **Manual contracts** removes the hardest engineering problem (work-site geocoding) from the
  critical path, so the *hypothesis* can be tested before the *engineering* is attempted. If citizens
  do not care about contract attribution, there is no point automating it.

That last point is the most important design decision in the whole plan.

---

## 4. The manual contract dataset

**Scope reduced 18 September 2026:** roads inside BMC's CC programme are attributed from the roads
API, so this dataset covers the rest. It is assembled in Phase 0 through the manual-capture
extension (P3), because Mahatenders may not be crawled
([D027](../00-overview/05-decision-log.md)).

For R/C ward, one financial year, road works outside the CC programme:

1. Pull every BMC road contract for R/C from Mahatenders and the BMC portal **by hand**
2. Archive each source page and PDF with a hash
3. Extract by hand: contract ID, contractor, value, award date, completion date, DLP terms
   (with verbatim quote and page number)
4. Geocode work sites by hand against the road network
5. Load into the same schema the automated pipeline will use

Expected size: on the order of 20–60 contracts. A few days of careful work.

This dataset doubles as the **evaluation set** for the automated pipeline in v0.2: the hand-built
answers are the ground truth against which geocoding recall and precision are measured.

---

## 5. Build order — eight systems, one at a time

**Amended 28 September 2026.** The previous twenty-item list was accurate and unusable: it read as
a queue rather than a sequence of things that are each finished. v1 is now eight systems. **Only one
is open at a time.** A system is not started until the one before it meets its done bar, and the
done bar is the thing that can be checked, not a feeling that the code is written.

Two of these are already partly built and are listed at their real state, not at zero.

| # | System | What it is | Done when |
|---|---|---|---|
| **1** | **Attribution** | The point-to-contract join: a reported coordinate against the road geometry in BMC's own works data, with a buffer, a confidence band, and the `MatchBasis` record that explains the join in plain language. The coverage-honesty line (K7) where nothing matches | It runs over the golden points and reports precision and recall against a hand-checked answer for each. A match that is wrong is visible in that number, not hidden behind an average |
| **2** | **Jurisdiction** | R/C ward boundary loaded and verified, BMC department mapping, and the confidence gate that decides when to ask the single disambiguating question | ≥ 95% on the golden set of known-ward points, and every miss inspected |
| **3** | **Classification** | Claude vision with a structured output over the road-defect taxonomy, prompt in a versioned file, plus the eval harness | ≥ 90% on the labelled eval set, with the failure cases written down |
| **4** | **Issues** | Dedup and issue creation, corroboration, the SLA clock from `legal_constants`, and the append-only hash-chained timeline | A second report of the same defect joins the first rather than creating a duplicate, and the 48-hour clock is rendered from config with its citation |
| **5** | **Accounts** | Finish what exists: phone OTP as the publication gate, rate limiting, password reset, email verification, token revocation | Sign-in cannot be brute-forced, a lost password is recoverable, a stolen session can be cancelled, and nothing publishes from an unverified phone |
| **6** | **Public surface** | Redaction of faces and plates before any public derivative exists, coordinate coarsening, the issue permalink, the ward page, OG cards | A public page carries no precise coordinate and no unredacted face, on a 150 KB budget at 2G |
| **7** | **Share kit** | The annotated image and the text, generated per language from the fact set | A share card carries a coarsened location, the contract ID where published, and no accusation |
| **8** | **Operator console** | Moderation queue, collection and source health, the metrics in [observability](../03-architecture/10-observability.md), and the public status page | A failed collection run, a stale source and a report needing moderation are all visible in one place without reading a log |

### Where each one stands today

| System | State |
|---|---|
| 1 Attribution | **Not started — this is the current system.** The works data it joins against is already collected twice a day |
| 2 Jurisdiction | Not started |
| 3 Classification | Not started |
| 4 Issues | Not started. Capture, storage and retry-safety are built and are its foundation |
| 5 Accounts | **Partly built.** Email/password, Google, sessions and forgery resistance work. Missing: rate limiting, reset, email verification, revocation, phone OTP |
| 6 Public surface | Not started |
| 7 Share kit | Not started |
| 8 Operator console | Not started |

Systems 1, 2 and 3 each need their evaluation artefact built **before** the system is called done.
A classifier without an eval set is a demo, and so is a spatial join without a precision number.

### Why attribution goes first

It is the hypothesis. If a citizen does not behave differently when told a contract covers the
stretch and is still under warranty, the roadmap needs replanning — and that is cheaper to discover
now than after three more systems are built on the assumption. It is also the one system that can
be tested **today**: it needs a coordinate and the works data, both of which exist, and not the
classifier or the resolver.

---

## 6. Definition of done

A v0.1 feature is done when:

- [ ] It works end to end in production
- [ ] It has tests, including the failure path
- [ ] It degrades correctly when its dependency is unavailable
- [ ] It emits the metrics listed in [observability](../03-architecture/10-observability.md)
- [ ] It logs no PII, no precise coordinates, no signed media URLs
- [ ] Any user-visible string has been reviewed for allegation-free phrasing
- [ ] Any displayed fact about a named party carries a source and a retrieval date
- [ ] Its documentation is updated in the same commit

---

## 7. Pre-launch blocking checklist

Legal and compliance, from
[security and privacy §11](../03-architecture/11-security-and-privacy.md#11-compliance-checklist-pre-launch-blocking):

- [ ] DPDP consent notice drafted, reviewed, translated (en/mr)
- [ ] Consent scopes separable (`report_processing`, `public_display`, `notifications`)
- [ ] Data-principal rights endpoints live
- [ ] Redaction verified against an adversarial set
- [ ] Public coarsening enforced at the serialisation layer, with a test
- [ ] Grievance officer named and published
- [ ] Takedown and right-of-reply workflows documented and reachable
- [ ] Correction log public
- [ ] `SECURITY.md` published with a working contact
- [ ] Backup restore tested end to end
- [ ] Terms of use and privacy policy published

Plus:

- [ ] Contractor names are **not displayed** in v0.1 unless the manual dataset gives ≥ 0.95 confidence
      and the source is cited — the safest possible starting posture
- [ ] Every generated share text passes the prohibited-pattern check

---

## 8. Exit criteria

All five must hold before v0.2 begins:

| # | Criterion | Why it matters |
|---|---|---|
| 1 | **200 reports** from people outside the team | Proves the capture flow works for real users |
| 2 | **Jurisdiction accuracy ≥ 95%** on the golden set, with wrong-at-high-confidence ≈ 0 | Proves routing is trustworthy |
| 3 | **Classification accuracy ≥ 90%**, hazard recall ≥ 99% | Proves the model tier is right |
| 4 | **Share rate ≥ 30%** of verified issues | Proves the distribution loop |
| 5 | **≥ 1 documented case** where contract attribution changed citizen behaviour or produced an outcome | **Proves the thesis** |

If criterion 5 fails, do not proceed to v0.2 as planned. Re-plan around escalation instead. Write
that decision down as an ADR.

---

## 9. Explicit non-goals for v0.1

- Do not build filing adapters. Citizens can file with BMC themselves; we capture the reference.
- Do not build the escalation engine.
- Do not build the scorecard.
- Do not build the TUI.
- Do not build native apps.
- Do not onboard a second ward "since it's easy". It is not easy — every ward needs boundary
  verification, department mapping, and a gazetteer.

---

## 10. Team and time

Realistic for one to two people working seriously:

| Phase | Duration |
|---|---|
| Research and data acquisition (ward boundary, departments, PIO, manual contracts) | 2–3 weeks |
| Build items 1–13 | 5–7 weeks |
| Build items 14–19 | 2–3 weeks |
| Eval sets (labelling 500 photos, 200 golden points) | 2 weeks, overlapping |
| Legal and compliance work | 2 weeks, overlapping |
| **Total** | **~3 months** |

Since the September 2026 restructure, the research and data-acquisition row and the eval-set row
are done in [Phase 0](07-phase-0-instruments.md), so v0.1 itself is closer to two months.

The research phase is not a warm-up. Ward boundaries, department mappings, and the manual contract
dataset are the project's actual moat, and rushing them produces a product that routes wrongly — the
one failure it cannot recover from.
