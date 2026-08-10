# Problem statement

## The user's problem

A resident of Vasai, Kandivali, or Dombivli walks past an open manhole, a collapsed footpath, a
burst water line, or a mound of construction debris dumped into a nallah. They want it fixed. What
stands between intent and outcome:

| Friction | Consequence |
|---|---|
| **"Whose is it?"** A single stretch of road may be BMC, MMRDA, PWD, MSRDC, MIDC, CIDCO, or Railway property. Boundaries are not visible on the ground. | Complaint filed with the wrong body → "not our jurisdiction" → dead. |
| **"Which app?"** MyBMC MARG, Aaple Sarkar, CPGRAMS, Swachhata, RailMadad, plus per-corporation apps and 1916 / 139 helplines. | Citizen picks wrong channel or gives up. |
| **"Then what?"** Complaint is "closed" with a photo of unrelated work, or the ticket vanishes. | Learned helplessness; the citizen never reports again. |
| **"What do I even say?"** Effective escalation requires citing civic codes, SLA timelines, and the correct authority. | Only lawyers and activists escalate successfully. |
| **"It's my word against theirs."** No durable, timestamped, geotagged record survives the ticket's closure. | Nothing accumulates. Every monsoon starts from zero. |

## The systemic problem

### 1. Jurisdictional fragmentation is a feature of the excuse, not a bug of the map

MMR spans ~6,328 km² and contains **nine municipal corporations** (Greater Mumbai, Thane,
Kalyan-Dombivli, Navi Mumbai, Ulhasnagar, Bhiwandi-Nizampur, Vasai-Virar, Mira-Bhayandar, Panvel),
**nine municipal councils** (Ambernath, Kulgaon-Badlapur, Matheran, Karjat, Khopoli, Pen, Uran,
Alibaug, Palghar) and 1,000+ villages, overlaid with MMRDA, MSRDC, PWD, MIDC, CIDCO, MHADA,
the Central and Western Railways, and the port trust. See
[MMR jurisdiction map](../01-research/03-mmr-jurisdiction-map.md).

Each has its own complaint channel, its own SLA, and its own definition of "resolved". There is no
citizen-facing router.

### 2. Complaint systems are designed to close tickets, not to fix assets

Existing platforms model the **grievance** as the primary object. TraceSarkar models the **asset**
and the **obligation** as primary, and treats the grievance as an observation about them. This is
the architectural difference that makes accountability possible:

```
Grievance-centric (status quo)          Asset-centric (TraceSarkar)
────────────────────────────           ─────────────────────────────
ticket #48213  →  closed               road_segment#8811
ticket #51902  →  closed                 ├── contract: WS/2023/ROAD/117 (DLP → 2028-04)
ticket #55110  →  open                   ├── observation 2025-07-02  pothole  (photo, GPS)
                                         ├── observation 2025-08-19  pothole  (photo, GPS)  ← same 40 m
(no memory that these are the             ├── observation 2026-06-11  pothole  (photo, GPS)
 same 40 m of road, three times)          └── 3 defects inside warranty → contractor scorecard −3
```

Under the asset-centric model, a "closed" ticket that produces a re-report 40 days later is not a
success — it is evidence.

### 3. Procurement transparency exists but is not actionable

Data is published (Mahatenders, BMC portal, CPP Portal) in forms that are:

- **Keyed by tender ID**, not by geography — you cannot ask "what contract covers this coordinate?"
- **Split across award, work order, completion certificate, and payment**, with no stable join key
- **Frequently scanned PDFs**, requiring OCR
- **Silent on defect liability period** in the machine-readable fields, even though the DLP is the
  single most legally useful attribute (BMC asphalt roads carry ~3 years; concrete roads 5–10 years;
  20% of the contract value is typically withheld until the guarantee period expires)

The result: an enormous transparency apparatus that no citizen can use at the moment of the pothole.

### 4. Legal remedies exist but are gated on procedural literacy

| Remedy | Barrier |
|---|---|
| RTI (Maharashtra) | Correct PIO, correct department, correct fee, correct format; 30-day clock nobody tracks |
| First Appeal | Must be filed within 30 days of the PIO's reply or the expiry of the response period |
| NGT (Western Zone, Pune) | **Section 14 applications must be filed within 6 months** of the cause of action (extendable by up to 60 days for sufficient cause); Section 15 compensation within 5 years |
| Maharashtra Lokayukta | Notarised affidavit, complaint in duplicate, prior correspondence enclosed |
| Consumer Commission (e-Jagriti) | Framing municipal failure as "deficiency in service"; free below ₹5 lakh claim value |
| Bombay HC 2025 pothole directions | Knowing the directions exist at all |

Each is a form-filling problem with a deadline. Form-filling with a deadline is exactly what
software is for.

### 5. Nothing accumulates across corporations

A contractor blacklisted by one corporation can bid in the next one over. Directors of a blacklisted
firm can incorporate a new entity. Because no MMR-wide ledger exists, past performance is
structurally invisible at the moment of award. See
[tender engine](../03-architecture/06-tender-engine.md).

## Why previous attempts under-delivered

| Platform | What it got right | Where the model runs out |
|---|---|---|
| **Swachhata (MoHUA)** | Enormous scale (1.8 crore users, 93%+ reported resolution) | Sanitation-only; "resolved" is self-certified by the same ULB; no procurement link |
| **IChangeMyCity (Janaagraha)** | Genuine community layer, open complaint dataset | Grievance-centric; no jurisdiction resolution across parastatals; no legal escalation |
| **MyBMC MARG / 24x7** | Official, 114 grievance categories, real-time status | Single corporation; closure is unilateral; no contract attribution |
| **FixMyStreet (UK)** | Offline capture, emergency diversion, council data export | No procurement engine; UK's single-tier councils make jurisdiction trivial |
| **SeeClickFix (US)** | Work-order CRM integration, reputation/gamification | SaaS-to-government business model; citizen is not the customer |
| **vTaiwan / g0v** | Fast, open, API-driven deliberation | Requires an unusually cooperative state |

The unclaimed position is: **citizen-side, cross-jurisdiction, procurement-linked, legally
escalatory.**

## Constraints we must design around

1. **Monsoon spike.** Reports concentrate in 8–10 weeks with 10–50× baseline traffic, exactly when
   connectivity is worst.
2. **Scanned PDFs.** A large share of the most valuable procurement documents are images.
3. **No official API.** We must assume scraping and RTI are the acquisition channels; any official
   integration is upside, not a dependency.
4. **Defamation exposure.** Naming firms and officials is central to the product's value and is its
   largest legal risk. Everything published must be sourced from a public record and displayed as
   such. See [legal review checklist](../06-operations/03-legal-review-checklist.md).
5. **DPDP Act 2023 + DPDP Rules 2025.** Photographs are personal data when they contain identifiable
   people or number plates. Consent notices, retention limits, and breach notification are
   compliance obligations, not future work. See
   [security and privacy](../03-architecture/11-security-and-privacy.md).
6. **Zero budget assumption.** Design for a single-operator cost envelope until there is proof of
   usage. See [cost model](../06-operations/04-cost-model.md).

## The bet

> If a citizen can see, at the moment of reporting, **who was paid how much to prevent exactly this**,
> the reporting rate goes up, the share rate goes up, and the political cost of ignoring it goes up.

Everything in this repository is downstream of that bet. The MVP is designed to falsify it cheaply.
