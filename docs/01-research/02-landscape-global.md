# Landscape: global

What to steal, and what does not transfer to MMR.

---

## 1. FixMyStreet / FixMyStreet Pro (mySociety, UK — since 2007)

**Model:** citizen reports a street problem with a photo and a pin; the platform emails or API-posts
it to the correct council; everything is public by default.

### Features worth copying

| Feature | Why it matters in MMR |
|---|---|
| **Offline reporting** — capture geotagged photo and complete the form without connectivity, sync later | Monsoon is exactly when networks fail and reports spike. This is a requirement, not a nice-to-have. |
| **Emergency diversion** — during high-demand periods, administrators set site-wide messaging and divert emergency categories (fuel spillage, active flooding) to immediate dispatch rather than the standard queue | Directly applicable to open manholes and live electrical hazards during monsoon |
| **Staff-side filtering and export** by date, category, ward, and state | The Watchdog TUI is our version of this, aimed at journalists rather than council staff |
| **Automated deduplication** | Our issue-merge logic |
| **Public by default** | A deliberate design stance we adopt, with DPDP-driven modifications |

### Where it does not transfer

UK councils are largely single-tier and publish authoritative boundaries; "whose is it?" is close to
trivial. In MMR it is the hardest problem in the product. FixMyStreet also has no procurement engine
and no legal escalation ladder — it stops at "reported to the council".

---

## 2. SeeClickFix (USA — since 2008, now part of CivicPlus)

**Model:** citizen reporting front-end plus a SaaS work-order dashboard sold to municipalities, with
direct integrations into municipal asset-management systems (e.g. Cityworks).

### Features worth copying

| Feature | Adaptation |
|---|---|
| **Direct CRM / work-order integration** | Our filing adapters; the long-term goal is an official API into corporation systems |
| **Reputation and points** | Our trust score — but weighted toward *accuracy* and *verification*, not volume |
| **Public comment threads on issues** | Our evidence-only comments (see [moderation policy](../06-operations/02-moderation-policy.md)) |
| **"Loops of involvement"** — moving users from extrinsic reward-seeking to intrinsic civic motivation | Explicit design goal of the gamification model |

### Where it does not transfer

The business model. SeeClickFix's customer is the municipality; the citizen is the supply side.
TraceSarkar's customer is the citizen, and its leverage comes precisely from *not* being dependent on
municipal goodwill. Adopting SeeClickFix's revenue model would neuter the accountability layer.

---

## 3. Open311 / GeoReport v2

**Model:** a standard HTTP API for civic service requests. Six methods. UTF-8 mandatory; ISO 8601
timestamps with timezone. Implemented by Chicago, Toronto, San Francisco, Washington DC, Boston,
Baltimore, Bloomington, New Haven, Helsinki, and Bonn among others.

**Why TraceSarkar should implement it:**

1. **It is the obvious integration surface** if any MMR corporation ever opens up. Speaking the
   standard makes "please give us an endpoint" a much smaller ask.
2. **It gives third-party clients for free.** An existing ecosystem of Open311 apps and libraries can
   point at us.
3. **It is a good schema.** Service definitions, service requests, attributes, and statuses map
   cleanly onto our model.

**Where we extend it:** Open311 has no concept of an authority hierarchy, a contract, a defect
liability period, or an escalation. Those become namespaced extensions. See
[API design](../03-architecture/03-api-design.md).

---

## 4. vTaiwan / g0v (Taiwan)

**Model:** a civic-hacker community (g0v) plus a structured deliberation process (vTaiwan) using
Pol.is for opinion clustering; famously fast pandemic-era deployments (mask-availability maps,
vaccine scheduling) built on open government data.

**What transfers:** the demonstration that an API-driven, open-data architecture lets civil society
ship faster than the state; the deliberation tooling as a model for the participatory-budgeting
module.

**What does not:** vTaiwan depends on an unusually cooperative state that publishes data and treats
civic-tech output as legitimate input. MMR's baseline is closer to indifference. TraceSarkar must be
useful with zero cooperation, and better with some.

---

## 5. VoiceCast (Uganda) and SMS-first civic tools

**Model:** toll-free SMS and voice channels to capture opinion from populations without smartphones
or data.

**What transfers:** the reminder that a smartphone-only platform is a platform for people who already
have some access. An SMS/IVR fallback (report by missed call + SMS, or by WhatsApp) is a
distribution question, not a charity feature — significant parts of MMR's most-affected population
are exactly the population a data-hungry app excludes.

**Realistic MMR adaptation:** WhatsApp-first intake is a better fit than SMS. It supports photographs
and location natively, has near-universal penetration, and is where civic complaint behaviour
already happens (ward WhatsApp groups). See [share kit](../02-product/05-share-kit.md).

---

## 6. Participatory budgeting (Porto Alegre 1989 → Pune 2005 → present)

**Model:** citizens directly allocate a portion of the municipal capital budget.

**The Indian precedent that matters:** Pune has run participatory budgeting since 2005 — citizens
submit proposals for civic works (historically capped around ₹5 lakh per proposal), which are
screened for feasibility and routed to the ward committee (*prabhag samiti*) for inclusion in the
city budget.

**What to copy:**
- **Hard budget constraint in the UI.** If allocating to road paving depletes the drainage line, the
  interface shows it. This is what converts civic anger into financial literacy.
- **Deliberation before voting.** Reactionary voting produces reactionary budgets.
- **Automated compilation into the official format** for submission to the prabhag samiti.

**Reality check:** this is a v3+ feature. It requires a functioning ward-committee relationship, a
digitised ward budget, and a user base that already trusts the platform. Listing it in the roadmap
is honest; building it early would be a mistake. See [roadmap](../05-delivery/01-roadmap.md).

---

## 7. Procurement-integrity precedents

| Effort | Technique | Transferability to MMR |
|---|---|---|
| **CPARS (US federal)** | Contractor Performance Assessment Reporting System under FAR Subpart 42.15 — structured past-performance evaluations that feed source selection | High conceptually, zero directly: there is no MMR equivalent, which is exactly the gap the scorecard fills |
| **LADWP CPEP** | Municipal contractor performance evaluation covering safety and responsiveness | Model for the scorecard's dimension design |
| **Satellite-based "ghost project" detection** | Optical change detection against claimed completion coordinates | Sentinel-2's 10 m resolution is too coarse for most road works; usable for large sites, afforestation, land filling, and building footprints. Treat as an experiment, not a core feature. |
| **NLP on tender documents** | Detecting restrictive specifications, anomalous phrasing, and collusion signals | Genuinely applicable — restrictive-specification detection is a well-defined text task and needs no imagery |

**Honest assessment of "ghost infrastructure detection":** it is the most exciting idea in the
original concept note and the weakest one technically. Free optical satellite imagery cannot resolve
whether a 6 m road was resurfaced. What it *can* do: detect whether a claimed new road, building,
plantation, or landfill exists at all. Scope it to that, or drop it. See
[open questions](07-open-questions.md).

---

## 8. Summary — what we are actually borrowing

| From | What |
|---|---|
| FixMyStreet | Offline capture, emergency diversion, public-by-default, dedup |
| SeeClickFix | Work-order integration ambition, reputation loops |
| Open311 | The API contract |
| g0v | Open-data architecture and shipping speed |
| VoiceCast | Non-smartphone intake (adapted to WhatsApp) |
| Pune PB | Hard-constraint budget UI, prabhag samiti submission format |
| CPARS / CPEP | Contractor scorecard dimension design |
| CivicDataLab | OCDS as the normalisation target |

What nobody has combined: **cross-jurisdiction routing + procurement attribution + automated legal
escalation, on the citizen's side.**
