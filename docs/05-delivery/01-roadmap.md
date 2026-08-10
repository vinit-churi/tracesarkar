# Roadmap

**Governing principle:** every milestone must produce something a real person uses, and every
milestone is gated on the previous one producing evidence — not on the previous one merely shipping.

The failure mode this roadmap is designed to avoid: building thirty features, shipping none of them
well, and discovering after a year that the contract-matching engine never worked.

---

## v0.1 — One ward, one category, one authority

**Scope:** Kandivali West (BMC R/S ward) · `road_defect` only · BMC only.

| Deliverable | Detail |
|---|---|
| Photo capture with geotag | Web PWA first (fastest to iterate); native later |
| Classification | Claude vision, structured output, road-defect taxonomy |
| Jurisdiction | R/S ward boundary + BMC department mapping + confidence gate |
| Contract attribution | **Manual for v0.1** — one financial year of BMC road contracts for R/S, hand-geocoded, to validate the concept before automating it |
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

That last criterion is the whole hypothesis. If reporting rates and share rates are unchanged with
attribution vs without, the thesis is wrong and the roadmap needs re-planning.

---

## v0.2 — Automate the money, add the channels

| Deliverable | Detail |
|---|---|
| **Mahatenders + BMC tender ingestion** | Automated, archived, parsed |
| **DLP extraction** | VLM over contract PDFs, with verbatim + page provenance |
| **Work-site geocoding** | Gazetteer + road-name matching for R/S ward |
| **Automated attribution** | With publication thresholds |
| Filing adapter: MyBMC | Official reference capture |
| Offline capture | Deferred sync |
| Redaction | On-device + server fallback |
| Marathi + Hindi UI | Full |
| Road-ownership layer | First RTI replies loaded |

**Gate (explicit, scheduled):** if work-site geocoding cannot reach **≥ 50% recall at ≥ 95%
precision** on R/S ward, stop and re-plan the roadmap around escalation (D-series) rather than
attribution. This decision point has a date, an owner, and a written outcome.

---

## v0.3 — Consequence

| Deliverable | Detail |
|---|---|
| **One-click RTI generation** | Correct PIO, category-specific questions, fee verified live |
| SLA breach detection and notification | The action prompt |
| Repeat-failure detection | Same asset, N times — the devastating statistic |
| Trust score | Full model |
| WhatsApp intake bot | The distribution unlock |
| Moderation queue | AI-ranked, human-decided |
| Categories expanded | `waste`, `water_drainage`, `street_furniture` |
| Wards expanded | All of BMC R-zone |

---

## v0.4 — Escalate and contextualise

| Deliverable | Detail |
|---|---|
| RTI first-appeal automation | Clock + template |
| **HC compensation claim assistant** | Maharashtra-only; quantum and forum already fixed by the Court |
| Escalation ladder auto-advance (drafts) | With the deadline scheduler |
| Weather + sensor ingestion | MESONET, IMD, mumbaiflood.in |
| News ingestion and linking | Context and duplicate-panic suppression |
| Voice reporting | Bhashini ASR |
| Transit Mode | Railway geofence + RailMadad-compatible fields |
| Photo-integrity checks | phash, EXIF, reuse detection |

---

## v0.5 — Aggregate and expose

| Deliverable | Detail |
|---|---|
| **Contractor scorecard** | Published methodology, minimum-data thresholds, dispute channel |
| Blacklist register aggregation | Per-corporation notices |
| Right-of-reply workflow | Legal necessity once names are aggregated |
| Public correction log | Cheapest strong good-faith defence |
| Takedown workflow + grievance officer | IT Rules 2021 compliance |
| Public read-only API | Open311 + data API |
| Bulk export | CSV / GeoJSON / Parquet with provenance manifests |
| Campaigns | Ward and corridor level |
| Wards expanded | All of Greater Mumbai |

**Legal review gate:** the scorecard does not ship without a documented external review of its
methodology and publication thresholds.

---

## v0.6 — Power users and the second corporation

| Deliverable | Detail |
|---|---|
| **Watchdog TUI** | Query DSL, alerts, exports with provenance |
| Saved queries and alerts | The "set a trap" feature |
| Embeddable ward heatmap | For newsrooms |
| Volunteer verification missions | Converts `claimed_resolved` → confirmed |
| Weekly ward digests | Retention |
| NGT Original Application compiler | With the 6-month limitation countdown |
| RTS Act appeals | Needs the notified-service list |
| **Second corporation onboarded** | TMC or KDMC — proves the model generalises beyond BMC |

---

## v0.7 — Depth

| Deliverable | Detail |
|---|---|
| Document search over the archive | OCR'd, page-level hits, archival hashes |
| Lokayukta complaint packet | Print-and-notarise workflow |
| e-Jagriti consumer complaint | Quantifiable-loss cases |
| OCDS publication of the contracts corpus | Independently valuable civic infrastructure |
| Ward sentiment/intensity | From report text |
| Gujarati UI | — |

---

## v1.0 — MMR

All nine corporations. Filing adapters for the major channels. The escalation ladder complete.
Public API stable and documented. Transparency report published quarterly.

**v1.0 is defined by coverage and reliability, not by new features.**

---

## Beyond v1

Ordered by expected value, not by excitement:

| Item | Notes |
|---|---|
| Corporate lineage mapping (DIN-based) | High value, high legal care required |
| Payment vs milestone reconciliation | Needs RTI-obtained payment records |
| Tender text analytics (restrictive specifications) | Descriptive observations only |
| Councillor / MLA accountability pages | Highest political risk in the catalog |
| Participatory budgeting engine | Needs ward-committee relationships; Pune precedent |
| Anomaly detection on award patterns | — |
| Satellite change detection | **Only** for large sites, afforestation, buildings. Sentinel-2 at 10 m cannot resolve road works — see [open questions](../01-research/07-open-questions.md) |
| Self-hosted VLM | Revisit when inference cost justifies GPU operations |

**Explicitly not planned:** IoT structural-health monitoring, bridge-deflection prediction,
blockchain anything, anonymous whistleblower drop. See
[feature catalog §Explicitly cut](../02-product/02-feature-catalog.md#explicitly-cut).

---

## Sequencing logic

```
v0.1  prove people report              ─┐
v0.2  prove the money can be joined     ├─ if either fails, the product is different
v0.3  prove consequence is generated   ─┘
v0.4  prove escalation compounds
v0.5  prove aggregation survives scrutiny (legal + dispute)
v0.6  prove journalists use it
v0.7  prove depth
v1.0  prove it generalises across corporations
```

Each row is falsifiable. Each has an exit criterion that can fail.

---

## Standing pre-monsoon task (every April)

Independent of milestone:

- [ ] Load test at 20× baseline
- [ ] Pre-scale the cluster
- [ ] Verify offline capture and sync end to end
- [ ] Emergency-mode drill
- [ ] Refresh ward boundary and contract data
- [ ] Verify the deadline scheduler and its alert
- [ ] Restore test

Monsoon is when the platform matters most and when everything is most likely to break.
