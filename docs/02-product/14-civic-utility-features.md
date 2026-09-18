# Civic utility features — beyond the report loop

The snap-to-action loop is the product's spine. It is also **episodic**: a resident photographs a
pothole perhaps twice a year. A platform that only does reporting is opened twice a year, and an
app opened twice a year is deleted.

This document specifies the surfaces that give the platform a reason to be opened weekly, using data
the [August 2026 availability audit](../01-research/08-data-availability-audit.md) established we
actually have. Feature IDs continue the [feature catalog](02-feature-catalog.md) (series A–J) and
the [data-unlocked addendum](13-data-unlocked-features.md) (series K) as **series U**.

---

## 1. The rule that keeps this from becoming a weather app

Every utility surface must **route back to an accountability action**. Air quality that is only a
number is a worse version of an app that already exists. Air quality that ends in
*"report the dust source"* or *"this construction site's consent-to-operate is on record"* is this
platform.

Concretely, three constraints on everything below:

| # | Constraint |
|---|---|
| R1 | Every utility screen carries at least one action that advances an issue, a claim, or a record — never a dead-end readout |
| R2 | Every displayed fact about a named party or a public body carries its source and retrieval date, exactly as elsewhere. Utility surfaces get no exemption from [hard rule 2](../../CLAUDE.md) |
| R3 | No feature here may be measured by engagement. The metrics stay confirmed fixes, corroborations, and instruments filed. A utility surface that raises time-in-app and nothing else is a failure, not a win |

And one anti-pattern, stated plainly because it is the obvious next idea: **no streaks, no daily
check-in rewards, no notification whose purpose is re-engagement.** See
[gamification](09-gamification.md).

---

## 2. Index

| # | Feature | Data source | Source status | Milestone |
|---|---|---|---|---|
| U1 | Deadline wallet | Own records + `legal_constants` | ✅ constants verified | v0.3 |
| U2 | Warranty expiry alert | Contract dataset (DLP) | ⛔ blocked on DLP source | v0.4 |
| U3 | Pre-monsoon watch | BMC announcements + roads API | ◐ | v0.5 |
| U4 | Civic calendar | Published institutional dates | ◐ | v0.6 |
| U5 | Works near me | BMC roads API | ✅ verified open | v0.3 |
| U6 | Air on your street | CPCB CAAQMS feed | ✅ verified open, keyless | v0.3 |
| U7 | Planned interruptions | Adani outage PDF, megablock notices | ◐ | v0.6 |
| U8 | Who do I call | BMC PIO list, ward directory, prabhag layer | ◐ | v0.2 |
| U9 | Service deadline lookup | Maharashtra RTS Act notified services | ⛔ list not retrieved | v0.6 |
| U10 | Ward works and spend | Roads API + BMC budget PDFs | ◐ | v0.5 |
| U11 | Cost sanity | BMC USOR / PWD SSR | ◐ edition unconfirmed | v0.7 |
| U12 | Evidence vault | Own records | ✅ | v0.4 |
| U13 | Injury claim helper | Bombay HC order, para 70 | ✅ verified verbatim | v0.4 |
| U14 | Contract watchlist | Roads API daily snapshot | ✅ | v0.3 |
| U15 | Weekly ward digest | Own records | ✅ | v0.6 |

---

## 3. Deadline and date features

### U1 — Deadline wallet · v0.3

**What it is.** One screen holding every clock the citizen is currently inside, sorted by what runs
out first. Not a notification stream — a standing ledger of obligations owed *to* them.

**Who it is for.** Anyone who has filed anything. The escalation ladder is otherwise homework: the
citizen has to remember that an RTI reply was due, and they will not.

**Clocks it holds.**

| Clock | Value | Source status |
|---|---|---|
| Pothole attendance | 48 h from notice to the authority | ✅ Bombay HC para 70(ix) |
| Compensation disbursal | 6–8 weeks from claim; 9% p.a. interest after | ✅ para 70(x) |
| Committee first meeting | 7 days from receipt of information | ✅ para 70(iii) |
| RTI reply | 30 days; 48 h where life or liberty is involved | ◐ verify against the bare Act |
| RTI first appeal window | 30 days from the decision or the expiry of the reply period | ◐ |
| RTI second appeal window | 90 days | ◐ |
| Municipal grievance SLA | Per category, per body | ◐ |
| NGT limitation | 6 months | ⛔ unverified — do not display until confirmed |

**Mechanics.** Every clock is a row in `legal_constants` with a value, an effective-from date and a
citation URL. Nothing is computed from a literal. A clock instance is attached to an issue or an
instrument, and the wallet is a query over open instances. When a constant changes, running clocks
keep the version they started under, and the row records which version applied.

**Surface.** A dedicated screen (S35), plus a badge on the profile tab when something is due within
48 hours. Each row: what is owed, by whom, remaining time, the citation, and the single action
available now (`Draft the appeal`, `Record the reply`, `Nothing to do yet — we will tell you`).

**Notifications.** One at T−24 h, one at expiry, one when a new action unlocks. Never more.

**Guardrail.** A clock with an unverified constant is **not shown at all** — an authoritative-looking
countdown built on a blog post is worse than no countdown. See open question Q2.

**Failure mode.** If the reference number was never captured, the row shows *"We can't track this
without the registration number"* and links to S21 rather than guessing a start date.

---

### U2 — Warranty expiry alert · v0.4

**What it is.** A standing alert on a street the citizen has reported on or followed: *"The contract
covering this stretch leaves its defect liability period on 28 Nov 2028. Defects reported before
then must be repaired at the contractor's cost."*

**Why it matters.** This is the single most useful fact the platform holds and the one no citizen
has ever been told. It converts a vague grievance into a timed right. It is also the feature nobody
can copy without doing the contract join.

**Data.** Contract identity and dates come from the contract dataset. **The DLP duration itself is
currently unavailable for BMC contracts** — BMC's roads API carries a `dlpPeriod` field that is null
in all 2,405 records. The state reference point is now sourced: PWD GR of 14 January 2019 sets 10
years for rigid pavement ≥30 cm and 5 years for designed flexible pavement, **for PWD works only**
([transcription](../01-research/sources/2026-09-18-pwd-dlp-gr-2019.md)). It does not bind BMC, so it
may be shown as a labelled reference, never as this contract's period. See audit §1.5 and open
question Q8.

**Therefore, two rendering modes:**

| State | What is shown |
|---|---|
| DLP known and sourced | The date, the countdown, and *"defects in this period are repaired free of cost"*, with the citation |
| DLP unknown | *"BMC recorded this road complete on 28 Nov 2023. The Bombay High Court records an expectation that roads should not require repair for five to ten years (para 69)."* Two sourced facts, no inference |

The second mode ships first, is honest, and is still more than any other product says.

**Notifications.** At 6 months, 3 months and 30 days before expiry, to anyone following that stretch.
The 30-day one is the valuable one: it is the last moment to get a defect on the record inside the
window.

**Guardrail.** Never state or imply that a specific defect *is* a warranty breach. The platform
states the dates and the overlap; the conclusion belongs to the authority.

---

### U3 — Pre-monsoon watch · v0.5

**What it is.** Every year BMC announces deadlines for road works and nallah desilting to be
complete before the monsoon. This tracks announced against actual, per ward, using BMC's own works
data.

**Mechanics.** Capture the announcement as a dated artefact (press note, circular). Snapshot the
roads API on the announced deadline. Publish the two side by side: *N works were to be complete by
31 May; on 1 June, BMC's own dashboard recorded M as complete.*

**Surface.** A seasonal ward page section, live from April to July, and a digest entry.

**Guardrail.** Both numbers come from the authority. The platform performs subtraction and nothing
else — no adjectives, no "failed", no percentage framed as a grade.

**Failure mode.** If the announcement cannot be sourced as a document, the feature does not run that
year. A remembered news headline is not a baseline.

---

### U4 — Civic calendar · v0.6

**What it is.** The dates a resident cannot find and would use if they could: BMC budget tabling
(February), ward committee meeting dates, CAG report tabling, public consultation windows on
development plans, RTS Act service reviews, and — carefully — election dates.

**Surface.** A month view on the public web, plus opt-in reminders. Each entry carries the body, the
source of the date, and whether the meeting is open to the public.

**Guardrail.** Election entries are dates only. The moment the calendar carries anything about a
candidate it falls under [ADR 0013](../04-adr/0013-political-accountability-scope.md), including the
election-period freeze.

---

## 4. Daily utility

### U5 — Works near me · v0.3

**What it is.** A map and list of road works in progress near the citizen: what is being built, by
which package, planned start and end, current status, and whether a traffic NOC has been received.

**Data.** ✅ Verified open. BMC's roads API returns 2,405 road geometries whose `location` object
carries `workCode`, `status`, `startDate`, `endDate`, `roadType`, `pqcstatus`,
`pqcPercentProgress`, `excavationPlannedStartDate`, `excavationPlannedEndDate`,
`layingOfDuctPlanned*`, `swdPlanned*`, `trafficNOCApplicationDate`, `trafficNOCReceivedDate`, and
`microPlan*`. Ingestion strips `contractorRepName` and `contractorRepMobile` — see audit §1.1.

**Why people will open it.** It answers "why is my road dug up and when will it stop", which is a
weekly question in Mumbai and currently unanswerable. It is the retention surface.

**Surface (S36).** Map with work segments coloured by status (icon and label too — [G11](11-screen-spec.md)),
a bottom sheet listing each with dates, and a per-work detail carrying the source line.

**Route back (R1).** Every work detail offers *"Is this site unsafe or abandoned? Report it"*, which
enters the normal capture flow pre-attributed to that work code.

**Guardrail.** Planned dates are BMC's, presented as BMC's. A date that has passed is shown as
*"planned end 12 Jun 2026 · still recorded in progress"* — the delta, not a verdict.

---

### U6 — Air on your street · v0.3

**What it is.** The nearest continuous monitoring station's current reading, attached to the
citizen's location and to any dust, burning or construction report.

**Data.** ✅ Verified open, keyless, hourly: `airquality.cpcb.gov.in/caaqms/rss_feed`, 88
Maharashtra stations of which 29 are named for Mumbai or Navi Mumbai, each with coordinates,
per-pollutant sub-indices and an AQI. Terms of use for the endpoint were not located — a terms
review goes in the source register before this ships.

**Surface.** A compact card on the home screen and inside relevant issue types: value, station name,
distance, timestamp.

**Route back.** *"Dust from a construction site? Report it"* — and where the site can be matched to a
work code, the report is pre-attributed.

**Guardrail.** Show the station, the distance and the time, always. An AQI number without its
station is a number pretending to be about your street. Never interpolate between stations and
present the result as a measurement.

---

### U7 — Planned interruptions · v0.6

**What it is.** Scheduled power cuts, water cuts, and railway megablocks affecting the citizen's
area, in one place.

**Data.** ◐ Adani Electricity publishes a rolling 7-day outage schedule as a dated PDF. Megablock
notices are press and social only. MSEDCL, Tata Power and MGL had no structured feed at survey time.

**Second use, and the better one:** suppress false escalation. A "no power" report that coincides
with a published outage window is contextualised rather than escalated — *"a planned outage is on
record for this area, 10:00–16:00 today"* — which protects the platform's signal-to-noise and the
authority's goodwill.

**Guardrail.** A published schedule is a claim by the utility, labelled as such. If the power is
still out at 17:00, that is a fact worth capturing, not an exception to swallow.

---

## 5. Civic navigation

### U8 — Who do I call · v0.2

**What it is.** For this exact location: the ward, the prabhag, the responsible department, the
junior engineer's designation, the ward office address and phone, the Public Information Officer for
RTI purposes, and the working helpline numbers.

**Data.** ◐ The prabhag layer carries `JR_ENGG` and `Beat_No`; BMC's public disclosure page publishes
a downloadable list of all PIOs; helpline and WhatsApp channels are published. The corporator layer
is now live again — see [ADR 0013](../04-adr/0013-political-accountability-scope.md) for what may be
shown about a person as opposed to an office.

**Why it is early.** It is cheap, it is the most-asked question in every civic forum in the city,
and it makes the platform useful before a single contract is matched.

**Surface (S37).** A card on every issue, and a standalone lookup by pin drop.

**Guardrail.** **Designations by default, names only where an official record names that person in
that capacity**, with source and date — the standing position in
[open questions](../01-research/07-open-questions.md) Q11. Never publish a personal mobile number,
even one that appears in a public dataset.

---

### U9 — Service deadline lookup · v0.6

**What it is.** "How long is the government allowed to take?" — for birth and death certificates,
water connections, building permissions, and the rest of the notified services, with the appeal path
and the penalty when they miss it.

**Data.** ⛔ Blocked. The Maharashtra Right to Public Services Act 2015 notified-services list is
rendered dynamically on the state portal and was not retrievable in the audit. It needs a
browser-based pass or an RTI. Until then this feature does not exist, because a wrong SLA published
here would be actively harmful.

**When unblocked.** Each service becomes a `legal_constants` row: service, timeline, first appeal,
second appeal, penalty, citation. The lookup is then trivial, and the RTS appeal joins the
escalation ladder as a first-class instrument.

---

## 6. Money

### U10 — Ward works and spend · v0.5

**What it is.** What is being built in this ward, at what stage, and — where the budget documents
support it — what was allocated against what is visible on the ground.

**Data.** ◐ Roads API for works; BMC budget documents (PDF, per year) for allocation; Praja
Foundation's ward analysis as a labelled third-party comparison.

**Guardrail.** Allocation and expenditure are different things, and a PDF budget line does not map
cleanly to a work code. Show them as two separately sourced facts on the same page. **Do not
compute a "value for money" figure** — that is a conclusion, and the data does not support it.

---

### U11 — Cost sanity · v0.7

**What it is.** *"Work of this type is rated at ₹X per square metre under BMC's Unified Schedule of
Rates, edition <year>."* Context for a citizen looking at a contract value.

**Data.** ◐ BMC USOR and the PWD State Schedule of Rates are published as PDFs. The most recent BMC
edition confirmed in the audit is 2018 — verify for a newer one before shipping.

**Guardrail.** The rate is a published norm, not an audit finding. Never juxtapose a rate and a
contract value in a way that implies overpricing; contract scope, site conditions and escalation
clauses all legitimately move the number. If this feature cannot be built without implying a
verdict, it is not built.

---

## 7. Evidence as a product

### U12 — Evidence vault · v0.4

**What it is.** Everything the citizen has captured — photographs with capture-time GPS and
timestamps, the enrichment record, the filing references, the timeline — exportable as a single
citable PDF with a provenance manifest and a verification link.

**Who it is for.** People will use this for reasons the product did not design: an insurance claim,
a society dispute, a lawyer's first meeting, a journalist's sourcing. That is fine, and it is a
strong reason to keep the archive rigorous.

**Mechanics.** The export carries: each artefact with its SHA-256, the hash-chained event timeline,
every source and retrieval date used in the enrichment, the platform version, and a statement of
what the platform does and does not assert. Redaction is applied to the public derivative; the
export to the citizen's own report may include the unredacted original **only to that citizen**.

**Guardrail.** The export must be honest about limits: it is not certified, it is not a legal
opinion, and the platform's own enrichment is labelled as such. A document that overstates its
standing damages the person relying on it.

---

### U13 — Injury and compensation claim helper · v0.4

Already catalogued as D9. Restated here because the audit changed what it is.

**What it is.** For a pothole or open-manhole death or injury: the quantum, the forum, the evidence
checklist, the drafted application, and the clocks.

**Data.** ✅ Verified verbatim from the order (PIL 71/2013, 13 Oct 2025, paragraph 70):

- Death ₹6,00,000; injury ₹50,000–₹2,50,000 by gravity — para 70(i)
- **Forum: a committee of the Municipal Commissioner and the District Legal Services Authority
  secretary** (Chief Officer or District Collector outside a corporation; the CEO or Principal
  Secretary for MMRDA, MSRDC, PWD, BPT, NHAI) — para 70(ii)
- The committee may act on an application, suo motu, **or "on receipt of information from any
  source"** — para 70(iv)
- First meeting within 7 days; thereafter at least every 15 days — para 70(iii)
- Disbursal within 6–8 weeks, then 9% per annum and personal responsibility of the named officer —
  para 70(x)

**Why this changes the feature.** It was scoped as a High Court filing. It is an administrative
application to a named committee — a far lower barrier, and a route a citizen can actually walk.

**Tone.** Quiet, plain, no accent colour, no share prompt, no celebratory language, no gamification.
See [brand](../00-overview/04-naming-and-brand.md) tone table and screen S28.

**Guardrail.** Not legal advice, stated once and plainly. Legal-aid handoff offered. The draft is the
citizen's application, in their name, filed by them.

---

## 8. Watching and digesting

### U14 — Contract watchlist · v0.3

**What it is.** A daily snapshot of the works dataset, diffed. When a status, a date or a contractor
changes, the change is recorded and anyone following that stretch is told.

**Why.** Government dashboards are edited. A road that moves from `In Progress` to `Completed`
overnight, or an `endDate` that shifts by a year, is a fact about the record — and the archive, not
the current value, is the evidentiary object ([hard rule](../../CLAUDE.md), audit §1.1).

**Surface.** A change log on the contract record; a notification to followers; a feed for the
Watchdog TUI.

**Guardrail.** Report the change and both values with timestamps. Never characterise an edit as
concealment.

---

### U15 — Weekly ward digest · v0.6

One message a week, opt-in, per ward: new issues, deadlines passed, confirmed fixes, works started
and finished, and one thing the reader can do. Available by email and WhatsApp.

**Guardrail.** Opt-in, one tap to leave, no re-engagement copy. The digest reports the ward, not the
reader's activity.

---

## 9. Notification budget

These features can generate a lot of messages. The budget is fixed, and it is small.

| Rule | Value |
|---|---|
| Hard ceiling per user | 5 notifications per week, excluding ones they explicitly asked for (a watched stretch, a filed instrument) |
| Deadline notifications | T−24 h and expiry only |
| Quiet hours | 21:00–08:00 IST, except a life-safety hazard within 500 m |
| Bundling | Multiple events for the same ward in a day become one message |
| Never | Re-engagement, streaks, "you haven't reported in a while", digests the user did not opt into |

Full policy in [realtime and notifications](../03-architecture/08-realtime-and-notifications.md).

---

## 10. Build order

The core loop first. Then, in this order, and the reasoning matters more than the sequence:

1. **U8 Who do I call** (v0.2) — cheapest, answers the most-asked question, useful before any
   contract is matched.
2. **U5 Works near me** (v0.3) — the retention surface. Data is verified and open today.
3. **U1 Deadline wallet** (v0.3) — turns the escalation ladder from homework into a service.
4. **U6 Air on your street** and **U14 Contract watchlist** (v0.3) — both nearly free once ingestion
   exists.
5. **U13 Claim helper** (v0.4) — highest value per user reached, and fully unblocked.
6. **U12 Evidence vault** (v0.4) — makes everything already captured more useful.
7. **U2 Warranty alerts** (v0.4) — ships in its honest mode now, its full mode when the DLP source
   is obtained.

Everything after that is v0.5+ and should be re-argued against evidence from the first five.

**If only three are built:** U2, U5, U1. The first is unique and uncopyable, the second is the daily
habit, the third is what makes the accountability machinery feel like a service.

---

## 11. Blocked on someone else

| Feature | Blocker | Who resolves it |
|---|---|---|
| U2 full mode | ~~The DLP government resolution, by GR number~~ Found 18 Sep 2026 (PWD works only). Now: BMC's own contract DLP clauses | Contract documents, or an RTI |
| U9 | The RTS notified-services list. A UDD GR of 27 Jun 2025 revises 25 municipal service limits; whether it binds BMC is Q33 | A browser pass on the state portal, or an RTI |
| U11 | ~~Whether a BMC USOR edition supersedes 2018~~ Roads USOR 2023 found 18 Sep 2026 ◐ | — |
| U5, U14 | Terms of use for `roads.mcgm.gov.in` | A written request to MCGM |
| U6 | Terms for the CPCB feed | A written request to CPCB |
| Rainfall context | IITM's position on MESONET use by a non-academic platform | A written request to IITM |
