# Metrics

**The metric we refuse to optimise is total complaints filed.** Volume without resolution is the
documented failure mode of every predecessor platform — a 93% self-certified resolution rate that
citizens do not recognise on the ground is worse than no metric at all.

---

## 1. The North Star

> **Confirmed fixes attributable to the platform, per month.**

A confirmed fix requires a citizen photograph taken after the authority's resolution claim. It cannot
be self-certified by the body responsible for the failure.

It is a deliberately hard metric. It will be small for a long time. That is the point.

---

## 2. The accountability metrics (public)

These are published on the public dashboard. They are the platform's product, not its internal
telemetry.

| Metric | Definition | Why |
|---|---|---|
| **Claimed-vs-confirmed ratio** | `citizen_confirmed / claimed_resolved`, by authority and ward | The single most important number the platform produces. It measures the gap between what a body says and what a citizen sees. |
| **SLA breach rate** | Issues past their SLA / total routed, by authority and ward | Directly actionable; grounded in the 48-hour Bombay HC direction for road defects |
| **Median time to resolution** | `citizen_confirmed_at − first_reported_at` | Honest resolution time, not ticket-closure time |
| **Reopen rate** | Issues reopened after `citizen_confirmed` within 90 days | Measures fix quality, not fix speed |
| **In-warranty defect rate** | Defects on assets inside a contract DLP / total defects | The procurement accountability number |
| **Coverage** | Wards with verified boundaries, department mappings, and contract data | Honesty about what we actually know |
| **Data freshness** | Per-source last successful ingestion | Lets a journalist know whether a number is current |

---

## 3. Funnel metrics (internal)

```
capture started
   └─▶ report submitted            [ capture completion rate ]
        └─▶ classified              [ classification success rate ]
             └─▶ located            [ jurisdiction resolution rate ]
                  └─▶ verified      [ verification rate ]
                       └─▶ attributed  [ attribution publication rate ]
                            └─▶ shared    [ share rate ]
                                 └─▶ filed     [ filing rate ]
                                      └─▶ acknowledged
                                           └─▶ confirmed  ★
```

| Stage metric | v0.1 target |
|---|---|
| Capture completion (started → submitted) | ≥ 80% |
| Classification success (no human override) | ≥ 90% |
| Jurisdiction resolved without disambiguation | ≥ 75% |
| Jurisdiction accuracy on the golden set | ≥ 95% |
| Verification rate | ≥ 60% |
| Attribution published | ≥ 40% (with the manual dataset) |
| **Share rate** | **≥ 30%** |
| Filing rate | ≥ 50% |

---

## 4. Quality metrics

| Metric | Target | Meaning |
|---|---|---|
| Classification user-override rate | < 10% | The truest model-quality signal in production |
| Hazard recall | ≥ 99% | False negatives on open manholes are dangerous |
| Wrong-at-high-confidence jurisdiction rate | ≈ 0 | A confidently wrong route is worse than an honest question |
| Attribution disputes upheld | < 2% of published attributions | Legal exposure indicator |
| DLP extraction accuracy | ≥ 90% | Against the hand-labelled set |
| Duplicate merge precision | ≥ 95% | Wrongly merging two distinct issues hides one |
| Chatbot groundedness | 100% | Any failure is a bug, not a metric |

---

## 5. Engagement metrics — used carefully

| Metric | Use | Do not use it to |
|---|---|---|
| Repeat-report rate within 30 days | Is the loop working? | Optimise for report volume |
| Re-check response rate | Is the confirmation loop working? | — |
| Verification missions completed | Community health | Rank users publicly |
| Campaign participation | Organiser leverage | — |
| Permalink referral traffic by channel | Where distribution actually works | Chase vanity reach |

**Deliberately not tracked as a goal:** DAU, session length, time in app. A civic tool that
maximises time-in-app is misdesigned. The ideal interaction is sixty seconds long.

---

## 6. The core experiment

The whole platform is a bet:

> **Does public exposure of procurement attribution change authority behaviour?**

Instrumented from day one:

| Arm | Treatment |
|---|---|
| **A** | Issue with attribution displayed, share prompted |
| **B** | Issue with attribution displayed, share **not** prompted |
| **C** | Issue with attribution withheld (below threshold — a natural control) |

Measured: time to acknowledgement, time to claimed resolution, time to confirmed resolution, and
reopen rate.

This is an observational study, not a clean RCT — assignment to arm C is not random, and confounders
abound (severity, ward, season). It should be treated and reported as such. The honest version of
this analysis, published, is worth more than an overclaimed one.

---

## 7. Operational metrics

Defined in [observability](../03-architecture/10-observability.md). Summary of what gets alerted:
report submission availability, enrichment queue depth, deadline-scheduler liveness, ingestion
freshness, AI budget burn.

---

## 8. Cost metrics

| Metric | Watch |
|---|---|
| AI cost per report | The dominant variable cost |
| AI cache read ratio | Prompt caching effectiveness; a low ratio means something volatile leaked into the prefix |
| Storage cost per 1,000 reports | Media dominates |
| WhatsApp message cost per notified user | Per-message pricing; reserve for high-value events |
| Total cost per confirmed fix | **The efficiency number that matters** |

---

## 9. Reporting cadence

| Cadence | Report | Audience |
|---|---|---|
| Real-time | Public dashboard | Everyone |
| Weekly | Funnel + quality review | Team |
| Monthly | Cost, SLOs, capacity, risk register | Team |
| Quarterly | Transparency report (takedowns, disputes, corrections, incidents) | Public |
| Annually | Impact report with the core experiment's findings, caveats included | Public |

---

## 10. Anti-metrics — signals that we are drifting

If any of these appear, stop and reassess:

| Anti-signal | What it means |
|---|---|
| Reports rising while confirmed fixes stay flat | We have built a complaint app, not an accountability platform |
| Share rate falling as features are added | The product is getting complicated |
| Dispute rate rising | Attribution precision is slipping |
| Time-in-app rising | We have built engagement, not utility |
| Coverage claims outpacing verified boundary data | We are overstating what we know |
| A published number nobody can reproduce from the API | We have broken the provenance guarantee |
