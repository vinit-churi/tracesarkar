# Moderation policy

Public-facing policy plus the internal operating procedure. Written to be publishable as-is.

---

## 1. What we moderate and why

TraceSarkar publishes two kinds of content:

| Kind | Example | Standard |
|---|---|---|
| **Citizen content** | Photographs, descriptions, comments | Must be a genuine civic observation; no PII of third parties; no harassment; no political campaigning |
| **Platform statements** | Jurisdiction resolution, contract attribution, scorecards | Must be sourced, factual, and non-conclusory |

The two are visually and structurally distinct in every surface, because their liability treatment
differs: citizen content is third-party content under Section 79 of the IT Act; **platform statements
are ours**.

---

## 2. Content rules (public)

### Allowed

- Photographs of civic infrastructure and its condition, taken in a public place
- Factual descriptions of a problem, in any language
- Evidence-bearing comments (a second photograph, a reference number, a correction)
- Reports naming a **contract** or **firm** where the platform's own records support it

### Not allowed

| Prohibited | Reason |
|---|---|
| Photographs whose primary subject is an identifiable person | Privacy; DPDP |
| Photographs inside private property without a civic-infrastructure subject | Trespass and privacy |
| Allegations of corruption or criminality against a named person or firm | Defamation; the platform states facts, not conclusions |
| Political party references, symbols, or campaigning | Neutrality |
| Abuse, threats, communal or caste content, harassment | Obvious |
| Targeting a private individual (their house, their vehicle, their business) | Not a civic issue |
| Deliberately false or fabricated reports | Integrity |
| Old, reused, or manipulated photographs presented as current | Integrity |
| Personal contact details of officials or citizens | Privacy and harassment |
| Bulk-identical reports intended to distort statistics | Integrity |

### Consequences

| Severity | Action |
|---|---|
| First minor breach | Content removed; user notified with the specific rule and how to fix it |
| Repeated minor | Trust score reduced; rate limits tightened |
| Deliberate fabrication | Content removed; trust score reset; account limited |
| Coordinated manipulation | Account and associated cluster suspended; incident logged |
| Illegal content | Removed; preserved for authorities as required; reported |

Every action is appealable and appeals get a human.

---

## 3. The moderation queue

```
report/comment ──▶ automated signals ──▶ risk score ──▶ queue (ranked) ──▶ human decision
                        │
                        ├─ classification: is_civic_issue=false
                        ├─ image integrity: phash reuse, EXIF anomaly, clock skew
                        ├─ people_present with poor redaction confidence
                        ├─ account trust below threshold
                        ├─ independence clustering (device/IP/graph)
                        ├─ prohibited-pattern match in text
                        └─ user flags
```

**The model ranks; a human decides.** No content is removed by an automated decision, with two
exceptions:

1. Automated redaction failure → the *image* is withheld pending review (not a removal, a hold)
2. Confirmed exact-duplicate of previously-removed content → auto-removed

### Priority order in the queue

1. Anything flagged as `hazard_to_life` (review for accuracy, never delay routing)
2. Content naming a private individual
3. Content with high fabrication signals
4. Disputes from named parties
5. User flags
6. Routine low-trust review

**Target:** first review within 4 hours during working hours, 24 hours otherwise. Hazard items are
never gated on moderation for routing — only for public display.

---

## 4. Platform statement moderation

Different process, higher bar. A platform statement is withdrawn or corrected when:

- The source turns out to be wrong or has been corrected upstream
- The match confidence was miscalculated
- A dispute establishes a factual error
- The methodology changes in a way that invalidates prior outputs

**Every correction is published in the public correction log** with the old value, the new value, the
reason, and the date. Nothing is silently edited.

When a systematic error is found (e.g. a geocoding bug affecting a class of attributions), **all**
affected records are reviewed, not just the one that was reported.

---

## 5. Disputes from named parties

See also [trust and anti-abuse §8](../02-product/08-trust-and-antiabuse.md#8-dispute-and-right-of-reply).

| Step | Detail | Target |
|---|---|---|
| 1. Receipt | Acknowledged with a reference number | 2 working days |
| 2. Triage | Is a specific fact contested, or is this a general objection? | 5 working days |
| 3. Marking | The record shows "under review" while open | Immediate |
| 4. Investigation | Re-check the source, the retrieval, and the derivation | 15 working days |
| 5. Outcome | Upheld → correct + log. Rejected → record stands + the party's response displayed alongside. Partial → correct the specific fact | 20 working days |

**Bulk disputes** (the same party disputing many records at once) are handled per-record, not
per-batch. There is no bulk dispute API. A general assertion of defamation without a specific
contested fact is triaged and answered, not actioned.

---

## 6. Takedown requests

| Source | Handling |
|---|---|
| Court order | Complied with; actioned within the IT Rules 2021 window; logged; published in the transparency report |
| Government agency order | Same, with a legal review of scope; minimum necessary compliance |
| Private party legal notice | Reviewed on the merits under the dispute process; a notice is not an order |
| Individual privacy request (a person in a photo) | Prioritised; the image is re-redacted or removed; the issue survives without it |

Every takedown request, complied with or not, is counted in the quarterly transparency report.

---

## 7. Moderator conduct

| Rule | Detail |
|---|---|
| Least privilege | Moderators cannot see phone numbers or precise coordinates |
| Break-glass | Access to archival originals requires a logged justification |
| Audit | Every action logged with actor, target, reason, timestamp |
| Conflict of interest | Moderators recuse from content involving their own ward, employer, or any party they have a relationship with |
| No political affiliation disclosure requirement, but a declared-conflicts register | Practical neutrality |
| Two-person rule | Account suspensions and platform-statement withdrawals require a second reviewer |

---

## 8. Transparency

Published quarterly:

- Content actioned, by rule and outcome
- Accounts actioned, by reason
- Disputes received, upheld, rejected
- Takedown requests received, by source and outcome
- Government data demands received
- Corrections issued
- Median time to first review

A platform that demands transparency from municipal bodies and publishes nothing about its own
moderation is not credible.

---

## 9. Appeals

Any moderation decision can be appealed once. Appeals are reviewed by someone other than the original
decision-maker, within 10 working days, with a written outcome. Appeal outcomes are counted in the
transparency report.

---

## 10. What this policy deliberately does not do

| Not doing | Why |
|---|---|
| Remove content because it is unflattering to an authority | That is the point of the platform |
| Remove content on an unsubstantiated defamation claim | A notice is not an order; the dispute process handles the merits |
| Allow an authority to close or hide an issue | Only a citizen photograph moves an issue to confirmed |
| Rank content by engagement | The feed ranks by issue state, not by outrage |
| Host general political discussion | Comments are evidence-only |
