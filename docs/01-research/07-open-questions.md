# Open questions

Things we do not know, ordered by how much the answer would change the plan. Each has a proposed way
to find out.

---

## Status after the August 2026 audit

The [data availability audit](08-data-availability-audit.md) answered several of these and changed
the shape of others. Read that page before treating anything below as current.

| Question | What changed on 2026-08-24 |
|---|---|
| Q1 — work-site geocoding | **Narrowed, not answered.** For Mumbai's Mega CC-road programme, BMC publishes 2,405 road geometries with contractor and work code, so no geocoding is needed there. Q1 now applies to asphalt roads, other categories, and every other MMR body |
| Q2 — Maharashtra RTI fee | **Still open, and now known to be contested.** ₹10 is long-standing; a ₹30 figure circulates with no gazette notification found. Needs a GAD notification, not a blog |
| Q3 — terms of use | **Partly answered.** Mahatenders: `Disallow: /` plus captcha → no automated collection. GeM: captcha. CPPP: crawlable. **Still unknown:** `roads.mcgm.gov.in`, `portal.mcgm.gov.in`, MCGM's ArcGIS layers, the CPCB feed, and IITM MESONET for non-academic use |
| Q5 — does any MMR corporation expose an API | **Answered: no.** Verified across MyBMC, Swachhata, CPGRAMS, Aaple Sarkar, RailMadad, MSRDC, MMRDA and the utilities. CPGRAMS has an integration path, but it is government-to-government only. Filing stays draft-plus-deep-link |
| Q6 — road-ownership inventories | **Answered: none exists.** No dataset, state or central, carries a maintaining-agency attribute on a road geometry. The resolution ladder and the site-noticeboard capture idea are in the audit §2.2 |
| Q8 — DLP distribution | **Still open for BMC; the state reference point is now sourced.** BMC's roads API has a `dlpPeriod` field that is null in all 2,405 records. ~~The 2017 PWD GR that reportedly sets 15/30-year DLPs is press-sourced only and must be retrieved by GR number~~ — **18 Sep 2026:** the PWD GR is dated 14 Jan 2019 and sets 10 years for rigid pavement ≥30 cm, 5 for designed flexible pavement, 20 for bridges, for PWD works only. BMC contracts still need their own documents. See [the transcription](sources/2026-09-18-pwd-dlp-gr-2019.md) |

### New blocking questions

| # | Question | Gates |
|---|---|---|
| Q27 | What are the terms of use for `roads.mcgm.gov.in`, and what licence covers MCGM's ArcGIS layers? Fetchability is not permission | v0.2 ingestion |
| Q28 | Are there orders after 21 November 2025 in PIL 71/2013? Compliance reports were called for on that date, so the directions may have moved. **Partly answered 18 Sep 2026:** the 13 Oct 2025 order was passed in IA No. 29119/2025, and State GRs of 19 Nov 2025, 11 Dec 2025 and 20 May 2026 implement it ◐. The orders themselves remain behind the captcha | Every legal constant derived from that order |
| Q29 | Does the DPDP Act's publicly-available-data exemption cover republishing director and representative names, read properly by a lawyer rather than inferred from the text? | v0.5 contractor records, v1+ representative records |
| Q30 | Is TraceSarkar a "publisher of news and current affairs content" under the IT Rules 2021, and can its content be characterised as a political advertisement requiring MCMC pre-certification? | Any representative-linked surface, and the first live election cycle |
| Q31 | Can Bhashini be used in production by a non-government platform, and at what cost? The published terms scope the free tier to proof-of-concept | v0.4 voice |

### Added by the September 2026 vertical exploration

See [vertical exploration §10](09-vertical-exploration.md#10-still-open).

| # | Question | Gates |
|---|---|---|
| Q32 | What are the terms for MahaRERA's project map and BMC's storm-water drain desilting API? | Construction-site and drainage verticals at the public tier |
| Q33 | Does the UDD GR of 27 June 2025, which shortens limits for 25 municipal services, bind BMC? | U9 service deadline lookup |
| Q34 | Should the second geography be Pune or an MMR corporation? | Phasing |
| Q35 | Partner with Pothole Reporter, or compete with it? | Phasing |

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
