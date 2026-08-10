# Legal framework

The escalation engine is the product's differentiator, and it is entirely a legal-process problem.
This page specifies each instrument, its preconditions, its clock, and what the platform must
generate.

> ⚠️ **This is engineering research, not legal advice.** Fees, formats, and limitation periods
> change. Every figure below carries a "verify" flag. The platform must fetch or re-confirm fee and
> format values before generating any instrument — see
> [legal review checklist](../06-operations/03-legal-review-checklist.md).

---

## 1. Right to Information (RTI Act, 2005)

**Use when:** an authority has breached its SLA, closed a complaint without evidence, or the citizen
needs the underlying documents (work order, completion certificate, measurement book, payment
record, inspection reports).

| Attribute | Value | Verify |
|---|---|---|
| Filing route (Maharashtra state bodies) | `rtionline.maharashtra.gov.in` | ✔ live |
| Application fee | ₹30 under the Maharashtra RTI Rules notified in 2026 (previously ₹10) | ⚠️ **confirm before generating** |
| First appeal fee | ₹50 | ⚠️ confirm |
| PIO response window | 30 days | ✔ statutory |
| Life-and-liberty cases | 48 hours | ✔ statutory |
| First appeal window | 30 days from PIO reply / expiry of the 30-day period | ✔ statutory |
| Second appeal | State Information Commission, 90 days from first-appeal decision | ✔ statutory |
| Support | Citizen contact centre 1800 120 8040; `rti.support@maharashtra.gov.in` | ⚠️ confirm |
| Payment | Net banking / debit / credit card on the portal | ✔ |

**Central bodies** (Railways, etc.) file through `rtionline.gov.in`, not the state portal. The
generator must branch on authority type.

### What the platform generates

A pre-filled RTI body with:
- Correct PIO designation and address for the resolved authority
- A numbered, specific question list (vague RTIs get rejected — specificity is the whole game)
- The issue's geo-coordinates, timestamps, photograph references, and any official ticket number
- The applicable SLA and the date it was breached
- A "documents sought" list keyed to the issue category

### Question templates by category

| Category | Questions to ask |
|---|---|
| Road defect | Contract number covering this stretch; contractor name; award value; work-order and completion dates; DLP end date; last inspection report; penalty imposed for defects, if any |
| Solid waste | Collection contract for this ward; frequency stipulated; weighment records for the last 30 days; penalty clauses |
| Drain / flooding | De-silting contract and quantum for this nallah for the current year; pre-monsoon inspection report; silt disposal records |
| Streetlight | Maintenance contract; SLA for restoration; complaint register extract |
| Illegal construction / dumping | Notices issued under the applicable MMC Act sections; action-taken report |

---

## 2. Maharashtra Right to Public Services Act, 2015

**Use when:** the service in question is a *notified* service with a statutory timeline.

| Attribute | Value |
|---|---|
| In force since | 28 April 2015 |
| Scope | Government departments, local authorities (municipal corporations included) |
| Appeals | First and second appeal to higher authorities; third appeal to the Maharashtra State Commission for Right to Service |
| Penalty | Up to ₹5,000 per case on the designated officer |
| Portal | `aaplesarkar.mahaonline.gov.in` |

**Research task:** obtain the notified-service list for each MMR corporation, with timelines. Where
a civic complaint maps to a notified service, RTS is a *faster and cheaper* first escalation than
RTI, because it carries a personal penalty.

---

## 3. National Green Tribunal (NGT Act, 2010)

**Bench for MMR:** Western Zone Bench, Pune. Rule 11 requires filing in the zone where the
environmental breach occurred.

| Attribute | Value | Verify |
|---|---|---|
| §14 (original application) limitation | **6 months** from when the cause of action first arose, extendable by up to 60 days for sufficient cause | ✔ statutory |
| §15 (compensation / restitution) limitation | **5 years**, extendable by up to 60 days | ✔ statutory |
| Filing | NGT e-filing portal; most benches now mandate electronic filing | ⚠️ confirm current mandate |

**Categories that qualify:** debris dumping in mangroves or creeks, untreated effluent, illegal tree
felling, construction in CRZ, air-quality violations, nallah encroachment, biomedical/hazardous waste.

### Why the 6-month clock is the product's most important timer

Dismissal for delay is one of the commonest ways environmental petitions die. TraceSarkar's
immutable, timestamped observation trail establishes the cause-of-action date precisely — and the
platform should surface a countdown the moment an issue is classified as environmental.

### What the platform generates

- Cause title and party array (applicant, respondent authority, respondent contractor/polluter, MPCB)
- **List of Dates and Events** — auto-assembled from the issue timeline
- Statement of facts with embedded geo-referenced photographs
- The specific statutory provisions alleged to be violated
- Reliefs sought (restoration, environmental compensation, directions to the authority)
- Annexure index: photographs, RTI replies, correspondence, news items
- A limitation-period statement with the computed cause-of-action date

---

## 4. Maharashtra Lokayukta and Upa-Lokayuktas Act, 1971

**Use when:** documents (usually obtained via RTI) show funds released for work not delivered, or
maladministration by a named public servant.

| Attribute | Value | Verify |
|---|---|---|
| Complaint (grievance) limitation | 12 months from when the action became known | ⚠️ confirm |
| Allegation limitation | 3 years | ⚠️ confirm |
| Format | Signed complaint in **duplicate**, with enclosures; prior correspondence attached | ✔ per official guidance |
| Affidavit | **Original sworn affidavit required** in support of allegations against a public servant | ✔ |
| Submission | Post, in person, or email `soadm.lokayukta@maharashtra.gov.in` | ⚠️ confirm |
| Prescribed format | Available but **not obligatory** | ✔ |

**Design consequence:** Lokayukta cannot be fully automated — notarisation is a physical act. The
platform generates a print-ready packet and a checklist ("print, sign, notarise the affidavit,
enclose these 6 annexures, post to this address").

---

## 5. Consumer Protection Act, 2019 — e-Jagriti

**Use when:** the citizen is a tax-paying consumer of a municipal service (water supply, sanitation,
road maintenance) and can plead deficiency in service, ideally with quantifiable loss (vehicle
damage, medical expenses, property damage).

| Attribute | Value | Verify |
|---|---|---|
| Portal | `e-jagriti.gov.in` (launched 1 Jan 2025; replaced eDaakhil) | ✔ |
| Levels | District, State, National commissions | ✔ |
| Fee | Free for district claims up to ₹5 lakh; ₹200–₹7,500 above | ⚠️ confirm slab table |
| Auth | OTP-based registration for consumers and advocates | ✔ |

### What the platform generates

Facts-only complaint text (emotional language actively stripped), the defect chronology, the relief
sought with a computed claim amount, and the annexure index. Class-action framing where multiple
citizens report the same asset is a v2 feature.

---

## 6. Bombay High Court pothole directions

See [key judgments](06-key-judgments.md) for the full account. Operationally:

- Potholes must be attended **within 48 hours** of being brought to notice → this is the SLA the
  platform enforces for road defects in Maharashtra, regardless of what an authority's own charter says.
- Compensation: **₹6,00,000** for death; **₹50,000–₹2,50,000** for injury by severity.
- Disbursal within **6–8 weeks** of a claim, **9% interest** on delay.
- Committees may act *suo motu* or on application by victims/legal heirs.
- Funds sourced from contractor fines; authorities bear the liability if insufficient.
- Blacklisting and departmental/criminal proceedings against erring contractors and officials.
- Earlier directions in the same PIL line require single-window complaint systems, multiple reporting
  modes (WhatsApp, app, toll-free), and information boards at works sites.

**Product implication:** a *compensation claim assistant* is a distinct, high-value instrument —
arguably higher-value than RTI for injury cases, because the quantum and the forum are already fixed
by the Court.

---

## 7. Data protection: DPDP Act 2023 + DPDP Rules 2025

| Attribute | Value |
|---|---|
| Rules notified | 13 November 2025 (G.S.R. 846(E)) |
| Phase 1 (from 13 Nov 2025) | Data Protection Board established |
| Phase 2 (from 13 Nov 2026) | Consent manager registration framework |
| Phase 3 (full compliance) | Consent notices, data-principal rights, breach notification, SDF obligations; Schedule 1 penalties up to ₹250 crore effective **13 May 2027** |
| Breach penalty | Up to ₹200 crore for failure to notify |

### What this means concretely for TraceSarkar

1. **Photographs are personal data** when they contain identifiable faces or number plates.
   → Automatic face and plate blurring on ingest, before storage of the display derivative.
2. **Consent notice** must be itemised, plain-language, and specify purpose, categories, retention,
   and withdrawal mechanism. A checkbox saying "I agree to the terms" is not compliant.
3. **Retention limits** and pre-erasure notification for inactive users (Third Schedule) must be
   designed in, not bolted on.
4. **Breach notification** to the Board and affected principals within the specified timelines.
5. **Location data** of a reporting citizen is precise personal data. Public display must be
   coarsened (see [security and privacy](../03-architecture/11-security-and-privacy.md)).

---

## 8. Publication risk: defamation and intermediary liability

This is the product's largest non-technical risk. Naming contractors and officials is central to
its value.

### The safe posture

| Rule | Implementation |
|---|---|
| **Publish only what the state published** | Every contractor name, award value, and blacklisting entry displayed must carry a link/citation to the source document and its retrieval date. |
| **State facts, not conclusions** | "Contract WS/2023/ROAD/117 covers this segment; DLP ends 2028-04" ✔. "This contractor is corrupt" ✘. |
| **Separate observed from alleged** | Citizen text is user-generated content, clearly labelled as such, and moderated. Platform-generated joins are labelled as platform statements. |
| **Right of reply** | A documented process for a named firm or official to submit a response, displayed alongside. |
| **Correction log** | Public, append-only record of corrections — the single strongest defence of good faith. |
| **Takedown process** | IT Rules 2021 require intermediaries to act on court/government takedown orders within 36 hours on grounds including defamation. Build the workflow and the grievance officer role before launch, not after the first notice. |
| **Grievance officer** | Named, contactable, with published response timelines, per the IT Rules. |

### Safe harbour

Section 79 of the IT Act protects intermediaries for third-party content subject to due diligence
and content neutrality. **Platform-generated statements are not third-party content** — the
procurement joins and scorecards are ours. That distinction must be reflected in the UI (visually
distinct treatment for platform assertions vs citizen assertions) and in the moderation policy.

### Contractor scorecards specifically

A scorecard is defensible if and only if:
- Every input is a public record or a verified citizen observation, both cited
- The methodology is published and versioned
- The score is presented as a computed metric with its formula visible, not as a verdict
- Firms can dispute individual inputs through a documented channel

See [moderation policy](../06-operations/02-moderation-policy.md) and
[tender engine](../03-architecture/06-tender-engine.md).

---

## 9. Instrument decision table

The escalation engine selects instruments using this table. Multiple may apply simultaneously.

| Trigger | Instrument | Precondition |
|---|---|---|
| SLA breached (any category) | RTI | Issue routed and acknowledged, or filing evidence exists |
| RTI unanswered after 30 days | First Appeal | RTI registration number recorded |
| First appeal unanswered / rejected | SIC Second Appeal | First-appeal decision or 45 days elapsed |
| Notified service delayed | RTS appeal | Category maps to a notified service |
| Environmental harm | NGT §14 OA | Within 6 months of first observation |
| Environmental harm causing damage | NGT §15 claim | Within 5 years |
| Documents show payment without delivery | Lokayukta complaint | RTI reply in hand; affidavit required |
| Consumer suffered loss | e-Jagriti complaint | Quantifiable loss with proof |
| Death or injury from road defect | HC compensation claim | Medical/police records; Maharashtra only |
| Defect inside DLP | Contractor liability notice + scorecard entry | Contract matched with confidence ≥ threshold |

---

## 10. Research tasks

- [ ] Obtain and archive the current Maharashtra RTI Rules (fee schedule) — **blocking for v1**
- [ ] Confirm current NGT Western Zone e-filing requirements and any prescribed format
- [ ] Confirm Lokayukta limitation periods and current form set (Form I / II, Schedules A / B)
- [ ] Obtain the e-Jagriti fee slab table
- [ ] Obtain the RTS notified-service list per MMR corporation
- [ ] Retain counsel (or a pro-bono partner) for a pre-launch review of the scorecard methodology
- [ ] Draft the DPDP consent notice and the retention schedule
- [ ] Draft the takedown, right-of-reply, and correction-log procedures
