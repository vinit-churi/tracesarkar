# Open questions

Things we do not know, ordered by how much the answer would change the plan. Each has a proposed way
to find out.

---

## Blocking — must be answered before the milestone they gate

### Q1 — Can tender work-sites be geocoded at usable accuracy? · gates v0.2

The entire "follow the money" thesis rests on turning
*"Improvement of Link Road from Kandivali Station Road junction to Charkop naka"* into geometry.

**How to find out:** hand-build the R/S ward contract dataset for v0.1 (20–60 contracts). Use it as
ground truth. Build the gazetteer and matcher for that ward. Measure recall and precision.

**Decision rule:** ≥50% recall at ≥95% precision → automate and expand. Below that → re-plan the
roadmap around escalation, and write the ADR.

---

### Q2 — Is the current Maharashtra RTI fee ₹30, and what is the exact current format? · gates v0.3

Reporting indicates the fee rose from ₹10 to ₹30 (first appeal ₹50) under RTI Rules notified in 2026.
Generating an instrument with a wrong fee wastes the citizen's application.

**How to find out:** obtain and archive the current notified rules from the state portal; verify by
filing one RTI ourselves end to end.

---

### Q3 — What are the terms of use for our critical data sources? · gates v0.2 and v0.4

Specifically: Mahatenders (scraping), BMC ArcGIS (redistribution), IITM MESONET and mumbaiflood.in
("free for research and academic use" — does a public platform qualify?), Bhashini (non-government
platform usage and cost).

**How to find out:** read the published terms; where ambiguous, write and ask. Record every answer in
the source register. Where the answer is no, use RTI.

---

### Q4 — Does contract attribution actually change citizen behaviour? · gates everything

The core hypothesis. If showing the contract does not increase reporting, sharing, or persistence,
the product is a better complaint app and the roadmap is wrong.

**How to find out:** v0.1 exit criterion 5. Instrument the arms described in
[metrics §6](../05-delivery/04-metrics.md#6-the-core-experiment) from day one.

---

## High-value — would materially change design

### Q5 — Does any MMR corporation expose a usable API, or would one?

We assume no, and design filing adapters as form automation or generate-and-hand-off. If BMC has
even an undocumented endpoint, the filing story improves dramatically.

**How to find out:** inspect the MyBMC MARG app's network traffic (for our own understanding only —
not to build against an undocumented endpoint without permission); RTI asking for the grievance
system's technical documentation; ask directly.

---

### Q6 — Can we obtain per-ward road-ownership inventories?

The single biggest correctness dependency for jurisdiction resolution. Without it, arterial roads
inside city limits route wrongly.

**How to find out:** RTI to BMC's Roads department for R/S ward as a pilot; if refused or fobbed off,
first appeal. Repeat per corporation. This is a first-quarter task.

---

### Q7 — How much re-photograph compliance can we actually get?

The claimed-vs-confirmed ratio is our headline statistic, and it depends entirely on citizens going
back to check.

**How to find out:** measure the re-check response rate in v0.1 with a small sample. If it is under
10%, the verification-mission mechanic (routing to nearby high-trust users) becomes essential rather
than supplementary.

---

### Q8 — What is the real defect liability period distribution across BMC road contracts?

We have secondary reporting (concrete ~5–10 years, asphalt ~3 years, 20% retention until the
guarantee period ends) but no primary confirmation.

**How to find out:** the hand-built v0.1 contract dataset answers this directly, with verbatim
quotes and page numbers.

---

### Q9 — Is the "ghost project" idea viable at all with free imagery?

Sentinel-2 at 10 m resolution cannot determine whether a 6 m road was resurfaced. It *can* detect
whether a claimed new road, building, plantation, or landfill exists at all.

**How to find out:** a bounded experiment on 20 large completed contracts with known outcomes. If
detection accuracy on that narrow class is not high, drop the feature rather than shipping a
misleading one.

**Current position:** de-scope to large-site existence checks, or cut. Do not promise pothole-level
satellite verification.

---

### Q10 — Where does the SMS/WhatsApp line actually fall?

We assume WhatsApp reaches almost everyone SMS would. If a significant MMR population is
WhatsApp-less, that assumption excludes exactly the people the platform is for.

**How to find out:** ask the RWA organisers in the pilot ward; look for published MMR connectivity
data.

---

## Design questions with a working position

| # | Question | Current position |
|---|---|---|
| Q11 | Do we ever publish an individual official's name? | Designations by default; a name only where the official record names them in that capacity |
| Q12 | Do we publish reporter handles on public issues? | No by default; opt-in per user |
| Q13 | Cool-off before `citizen_confirmed` → `closed`? | 90 days for road defects; category-tuned thereafter |
| Q14 | Do we let authorities upload proof-of-work photographs? | Yes, via a scoped API key — but it remains a *claim* that still requires citizen confirmation to close |
| Q15 | Assets created eagerly or lazily? | Lazily, on the second nearby report, to avoid an asset table full of one-off noise |
| Q16 | TimescaleDB for observations? | Not at v0.1. Revisit above 50M rows/year — [ADR 0009](../04-adr/0009-timeseries.md) |
| Q17 | Should the chatbot draft an RTI conversationally? | No — trigger the structured generator; a conversationally-produced legal document is harder to validate |
| Q18 | Separate progression for organisers? | Probably; their contribution is distribution, not reporting |
| Q19 | Is level-5 moderation-queue access safe? | Only with training, a two-person rule, and recusal policy. Consider deferring past v1 |
| Q20 | Publish generated alt text as canonical? | Yes, labelled as auto-generated, with an edit affordance |

---

## Research questions (lower urgency, high long-term value)

| # | Question |
|---|---|
| Q21 | What is the correct normalisation for the contractor scorecard — per contract, per contract-kilometre, or per rupee? Needs real data before choosing |
| Q22 | Is there a detectable relationship between award value per kilometre and subsequent defect rate? |
| Q23 | Does rainfall intensity predict pothole formation at a usable lead time, given several monsoons of data? |
| Q24 | Do wards with active TraceSarkar organisers show different resolution rates than comparable wards without? (Needs careful confounder handling) |
| Q25 | What proportion of MMR road contracts have overlapping or ambiguous work-site descriptions with another contract? |
| Q26 | How often do blacklisted firms' directors appear in newly-awarded contracts? (DIN-based; publishable only as a factual overlap) |

---

## How to use this page

- A question moves to a decision (with an ADR if it is architectural) once answered.
- Answered questions are struck through with a link to where the answer lives, not deleted — the
  history is useful.
- A blocking question that cannot be answered before its milestone **blocks the milestone**. It does
  not get assumed away.
