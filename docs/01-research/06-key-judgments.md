# Key judgments and their product consequences

These decisions are not background reading. Each one converts directly into an SLA constant, a
generated instrument, or a UI affordance.

---

## 1. *High Court on its own motion v. State of Maharashtra* — Bombay HC, 13 October 2025

**Bench:** Justice Revati Mohite Dere and Justice Sandesh D. Patil (Division Bench).
**Lineage:** continues the long-running suo motu pothole PIL line (PIL 71/2013 and related matters).

### Holdings and directions

| Direction | Value |
|---|---|
| Right to safe roads | A fundamental right under **Article 21** |
| Compensation — death | **₹6,00,000** |
| Compensation — injury | **₹50,000 to ₹2,50,000**, graded by severity |
| Pothole attention window | **48 hours** from being brought to notice |
| Compensation disbursal | **6–8 weeks** from receipt of claim |
| Delay interest | **9% per annum** |
| Committees | Constituted to determine and disburse compensation; may act **suo motu** or on application by victims/legal heirs |
| Funding | From contractor fines; authorities bear liability if the fund is insufficient |
| Accountability | Strict disciplinary action against erring contractors and officials; **blacklisting** for defective work; departmental and criminal proceedings as warranted; **personal responsibility** of officers for payment delays |

The Court observed that refusing to award compensation "would amount to rendering lip service" to
Article 21 rights.

### Earlier directions in the same line that the Court reiterated

- A **single-window** grievance system for pothole complaints (originating in a 2015 Government
  Resolution, intended to run through the "Voice of the Citizens" website), which the Court has
  repeatedly found to be functioning poorly
- **Multiple reporting modes**: toll-free numbers, a website accepting photograph uploads, SMS, and
  complaint tracking — maintained **year-round**, not only in monsoon
- **Information boards at works sites**
- A **Centralised Grievance Redressal Mechanism** for road complaints, whose non-establishment the
  Court criticised in 2023, describing road deaths as "man-made disasters"

### Product consequences

1. **The road-defect SLA is 48 hours in Maharashtra.** Hard-code it; do not defer to a corporation's
   softer internal charter. Every road issue gets a 48-hour countdown from the routed timestamp.
2. **Build the compensation claim assistant.** For injury/death cases the forum, the quantum, and the
   timeline are already fixed by the Court. This is the highest-leverage instrument in the entire
   escalation set, and it is Maharashtra-specific — a genuine moat.
3. **Blacklisting is a live remedy.** The contractor scorecard is not merely informational; it feeds
   a remedy the Court has already directed.
4. **Photographic upload with tracking is judicially mandated.** TraceSarkar is not an adversarial
   overlay here — it is doing what the Court told the state to do. Frame partnership conversations
   accordingly.
5. **The 9% interest clock is a UI element.** Once a claim is filed, the platform tracks the 6–8 week
   window and surfaces accruing interest.

### Verification tasks

- [ ] Obtain the full order text from `bombayhighcourt.nic.in` and archive it in `docs/01-research/sources/`
- [ ] Extract the exact operative paragraph numbers for citation inside generated instruments
- [ ] Identify the constituted committees and their addresses per district — required for the
      claim assistant
- [ ] Track subsequent compliance orders in the same PIL

---

## 2. The single-window / centralised mechanism line (2015 → 2023 → 2025)

**Timeline as reconstructed from the record:**

| Date | Event |
|---|---|
| 07–08 Aug 2015 | Court notes a single-window pothole grievance system incorporated via Government Resolution; directs effective implementation through the "Voice of the Citizens" website |
| Subsequent hearings | Intervenor repeatedly reports the single-window system is not functioning; Court gives MCGM time |
| 2023 | Court criticises the State for failing to establish a Centralised Grievance Redressal Mechanism; characterises road deaths as man-made disasters |
| 13 Oct 2025 | Compensation framework and 48-hour direction (above) |

**Product consequence:** the state has been under a judicial obligation to build something very
close to TraceSarkar's routing layer for a decade and has not delivered it. This is both the market
gap and the strongest argument for an official integration.

---

## 3. Statutory limitation periods that behave like judgments

| Instrument | Limitation | Effect on product |
|---|---|---|
| NGT §14 | 6 months from cause of action (+60 days for sufficient cause) | Countdown timer on every environmental issue from the first observation timestamp |
| NGT §15 | 5 years (+60 days) | Long-tail compensation claims remain viable |
| RTI first appeal | 30 days from reply or expiry | Auto-scheduled reminder and pre-drafted appeal |
| RTI second appeal | 90 days from first-appeal decision | Same |
| Lokayukta grievance | ~12 months from knowledge (verify) | Countdown from the RTI reply date |
| Lokayukta allegation | ~3 years (verify) | Same |

Limitation is the most common way meritorious civic litigation dies. A platform that owns the
timestamp owns the limitation defence.

---

## 4. Doctrines worth encoding in generated instruments

| Doctrine | Use |
|---|---|
| Article 21 — right to life, extended to safe roads (Bombay HC, 2025) | Opening paragraph of road-defect escalations in Maharashtra |
| Public trust doctrine | Environmental filings concerning creeks, mangroves, and open land |
| Polluter pays / precautionary principle (§20, NGT Act) | Environmental compensation claims |
| Deficiency in service (CPA 2019) | Consumer commission complaints against municipal bodies |
| Legitimate expectation | SLA-breach arguments where a published charter exists |

---

## 5. Standing caveats

- None of the above is legal advice, and the platform must say so, prominently, on every generated
  instrument.
- Generated instruments are **drafts for human review**. The citizen (or their counsel) signs and
  files. See [legal review checklist](../06-operations/03-legal-review-checklist.md).
- Judgments get modified, stayed, and clarified. Every hard-coded constant derived from a judgment
  (the 48-hour SLA, the ₹6,00,000 figure) must live in a versioned configuration table with an
  effective-from date and a citation — never as a literal in code.
