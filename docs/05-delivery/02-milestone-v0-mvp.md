# v0.1 — the MVP, in full detail

**If you read one page in this repository, read this one.**

---

## 1. The one-sentence scope

> A resident of Kandivali West photographs a pothole; within six seconds they see which BMC ward and
> department owns it, which contract covers that stretch and whether it is still under warranty, and
> a ready-to-send share card — and the platform tracks the 48-hour clock the Bombay High Court
> imposed.

Nothing else.

---

## 2. Boundaries

| In | Out |
|---|---|
| BMC R/S ward (Kandivali West) | Every other ward, every other corporation |
| `road_defect` category | Waste, drainage, environment, transit, structural |
| BMC as the only authority | MMRDA, PWD, MSRDC, Railways |
| Web PWA | Native apps |
| English + Marathi | Hindi, Gujarati |
| Manual contract dataset | Automated tender ingestion |
| Share kit + permalink | Filing adapters, RTI, escalation |
| Phone OTP + rate limits | Full trust score, moderation queue |

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

For R/S ward, one financial year, road works only:

1. Pull every BMC road contract for R/S from Mahatenders and the BMC portal **by hand**
2. Archive each source page and PDF with a hash
3. Extract by hand: contract ID, contractor, value, award date, completion date, DLP terms
   (with verbatim quote and page number)
4. Geocode work sites by hand against the road network
5. Load into the same schema the automated pipeline will use

Expected size: on the order of 20–60 contracts. A few days of careful work.

This dataset doubles as the **evaluation set** for the automated pipeline in v0.2: the hand-built
answers are the ground truth against which geocoding recall and precision are measured.

---

## 5. Build order

| # | Work | Depends on |
|---|---|---|
| 1 | Repo scaffold, CI, Docker Compose, migrations | — |
| 2 | Schema: accounts, reports, issues, media, events, authorities, wards | 1 |
| 3 | Phone OTP auth, rate limits | 2 |
| 4 | `POST /v1/reports` with media upload to object storage | 2, 3 |
| 5 | Worker framework + job queue | 2 |
| 6 | Redaction (server-side; on-device deferred to v0.2) | 5 |
| 7 | Classification call + eval set (500 labelled photos) | 5 |
| 8 | R/S ward boundary loaded and verified; BMC department mapping | 2 |
| 9 | Jurisdiction resolver + golden set (200 points) | 8 |
| 10 | Dedup + issue creation + corroboration | 2, 9 |
| 11 | Manual contract dataset loaded | 2 |
| 12 | Attribution join (spatial, against the manual dataset) | 10, 11 |
| 13 | SLA clock from `legal_constants` | 10 |
| 14 | Issue permalink page + OG cards | 10, 12 |
| 15 | Share kit: annotated image + text (en/mr) | 14 |
| 16 | PWA capture flow | 4, 15 |
| 17 | Timeline + append-only events with hash chain | 2 |
| 18 | Public status page + data freshness | — |
| 19 | DPDP consent notice + data-rights endpoints | 2 |
| 20 | Deploy to production | all |

Items 7, 9, and 12 each need their evaluation artefact built **before** the feature is considered
done. A classifier without an eval set is a demo.

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

The research phase is not a warm-up. Ward boundaries, department mappings, and the manual contract
dataset are the project's actual moat, and rushing them produces a product that routes wrongly — the
one failure it cannot recover from.
