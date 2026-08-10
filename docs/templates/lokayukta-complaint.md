# Template — Maharashtra Lokayukta complaint

**Status:** draft skeleton for engineering. **Must be reviewed by someone competent in Lokayukta
practice before first use.** Limitation periods and the current form set must be verified against the
official source — see [legal framework](../01-research/04-legal-framework.md).

---

## When this instrument applies

Under the Maharashtra Lokayukta and Upa-Lokayuktas Act, 1971, against maladministration or corruption
by a public servant. The typical TraceSarkar fact pattern:

> An RTI reply shows that funds were fully released for work that a documented, timestamped,
> geotagged citizen record shows was not delivered, or failed within its guarantee period, and no
> action was taken against the contractor.

**Precondition:** an RTI reply in hand. Without documents, this is an allegation without a basis and
should not be generated.

| Category | Limitation (⚠️ verify) |
|---|---|
| Grievance | ~12 months from when the action became known |
| Allegation | ~3 years |

---

## The automation limit

**This instrument cannot be fully automated.** It requires:

- A complaint **in duplicate**, signed
- An **original sworn affidavit** in support of allegations against a public servant, sworn before a
  competent authority
- Copies of prior correspondence with the concerned authorities, enclosed

The platform therefore produces a **print-ready packet plus a checklist**, not a submission.

---

## Structure

```
BEFORE THE HON'BLE LOKAYUKTA / UPA-LOKAYUKTA, MAHARASHTRA

COMPLAINT UNDER SECTION {{ section }} OF THE MAHARASHTRA LOKAYUKTA AND
UPA-LOKAYUKTAS ACT, 1971

Complainant:
  {{ applicant.name }}
  {{ applicant.address }}
  {{ applicant.phone }}   {{ applicant.email }}

Public servant complained against:
  {{ official.designation }}
  {{ official.department }}
  {{ official.authority }}
  (Name, if disclosed in the official record: {{ official.name_if_public_record }})

Subject: {{ subject_line }}

────────────────────────────────────────────────────────────────
1.  BRIEF FACTS
      {{#each numbered_facts}}
      1.{{ n }}  {{ text }}                                    [Annexure {{ annexure }}]
      {{/each}}

2.  CHRONOLOGY
      {{#each timeline_entries}}
      {{ date }}   {{ description }}                           [Annexure {{ annexure }}]
      {{/each}}

3.  THE GRIEVANCE / ALLEGATION
      {{ grievance_statement }}

4.  ACTION TAKEN BY THE COMPLAINANT PRIOR TO THIS COMPLAINT
      {{#each prior_actions}}
      4.{{ n }}  {{ description }} on {{ date }}               [Annexure {{ annexure }}]
      {{/each}}

5.  RELIEF SOUGHT
      {{#each reliefs}}
      ({{ letter }})  {{ text }}
      {{/each}}

6.  DECLARATION
      I declare that the facts stated above are true to my knowledge, that the
      matter is not pending before any court or tribunal, and that I have not
      previously filed a complaint on the same subject before this office.

                                                    {{ applicant.name }}
Place: {{ place }}                                  (Signature)
Date:  {{ today }}

────────────────────────────────────────────────────────────────
AFFIDAVIT                                     [to be sworn separately]

I, {{ applicant.name }}, aged {{ age }}, residing at {{ address }}, do
hereby solemnly affirm and state as follows:

1. I am the complainant in the accompanying complaint and am competent to
   swear this affidavit.
2. The contents of paragraphs {{ para_range }} of the complaint are true to
   my personal knowledge, and those of paragraphs {{ para_range_2 }} are based
   on records obtained under the Right to Information Act, 2005, which I
   believe to be true.
3. The annexures filed with the complaint are true copies of their originals.

                                                    DEPONENT

VERIFICATION
Verified at {{ place }} on this {{ day }} day of {{ month }}, {{ year }} that
the contents of the above affidavit are true and correct to my knowledge, and
nothing material has been concealed therefrom.

                                                    DEPONENT
```

---

## Annexure set (auto-assembled)

| Annexure | Content |
|---|---|
| A | The RTI application as filed, with its registration number |
| B | The RTI reply received, in full |
| C | Photographic record with coordinates, timestamps, and archival hashes |
| D | The issue timeline, exported verbatim from the platform |
| E | The original complaint to the authority and its reference number |
| F | Correspondence with the authority, if any |
| G | Contract documents obtained (work order, completion certificate, payment records) |
| H | News reports, where relevant |

---

## Filing checklist (shown to the citizen)

- [ ] Print the complaint **in duplicate**
- [ ] Sign both copies
- [ ] Print the affidavit
- [ ] **Swear the affidavit before a competent authority** (notary / oath commissioner)
- [ ] Attach all annexures to both copies, indexed and paginated
- [ ] Attach copies of prior correspondence with the authority
- [ ] Verify you are within the limitation period (currently: {{ days_remaining }} days remaining)
- [ ] Submit by post, in person, or by email to the office of the Lokayukta
- [ ] Record the acknowledgement reference in TraceSarkar so the matter can be tracked

---

## Generation rules

1. **Precondition enforced in code:** an RTI reply must be recorded on the issue before this
   instrument can be generated. No documents, no complaint.
2. **Designations, not names**, unless the official record itself names the officer in that capacity
   — and then only with the source cited. See
   [legal review checklist §A](../06-operations/03-legal-review-checklist.md#a-publishing-anything-about-a-named-party).
3. **The complaint states facts and asks the Lokayukta to investigate.** It does not assert that
   corruption occurred. The platform's role is to present the documented discrepancy.
4. **The discrepancy is stated arithmetically** where possible: amount sanctioned, amount paid, work
   claimed complete on date X, defect observed on date Y, guarantee period expiring on date Z.
5. **Never auto-filed.** Notarisation is a physical act; so is this filing.
6. **Limitation computed and displayed.**
7. **Disclaimer on every page:** generated draft, not legal advice.

---

## Relief templates (starting points)

| Relief | Text |
|---|---|
| Investigation | "Investigate the circumstances in which payment was released for the work described, and whether the work was in fact executed as certified" |
| Recovery | "Direct recovery of the amount paid in respect of work not executed or executed defectively within the guarantee period" |
| Action against the contractor | "Direct the concerned authority to take action against the contractor in accordance with the terms of the contract and applicable blacklisting norms" |
| Accountability | "Fix responsibility on the officials who certified completion and released payment" |
| Systemic direction | "Direct the authority to put in place a verification mechanism before release of final payment for road works" |
