# Template — Compensation claim under the Bombay HC pothole directions

**Status:** draft skeleton for engineering. **Must be reviewed by counsel before first use.**

This is potentially the highest-value instrument the platform produces, because the forum, the
quantum, and the timeline are **already fixed by the Court**. The citizen does not have to argue for
compensation — only to establish the facts.

---

## Basis

*High Court on its own motion v. State of Maharashtra*, Bombay High Court, Division Bench of Justice
Revati Mohite Dere and Justice Sandesh D. Patil, **13 October 2025**:

| Element | Value |
|---|---|
| Right to safe roads | Fundamental right under Article 21 |
| Compensation — death | **₹6,00,000** |
| Compensation — injury | **₹50,000 – ₹2,50,000**, graded by severity |
| Disbursal | **6–8 weeks** from receipt of the claim |
| Interest on delay | **9% per annum** |
| Forum | Committees constituted to determine and disburse compensation; may act **suo motu** or on application by victims / legal heirs |
| Funding | Contractor fines; authorities bear the liability if insufficient |
| Accountability | Blacklisting of contractors; departmental and criminal proceedings; personal responsibility of officers for payment delays |

⚠️ The constituted committees, their composition, and their addresses per district must be identified
before this instrument can be generated. This is a **blocking research task** — see
[key judgments](../01-research/06-key-judgments.md).

---

## Structure

```
TO,
The {{ committee_name }}
(constituted pursuant to the order dated 13.10.2025 of the Hon'ble Bombay High
 Court in {{ case_number }})
{{ committee_address }}

APPLICATION FOR COMPENSATION IN RESPECT OF {{ death_or_injury }} CAUSED BY A
ROAD DEFECT

Applicant:
  {{ applicant.name }}
  {{ applicant.relationship_to_victim }}          (self / legal heir)
  {{ applicant.address }}
  {{ applicant.phone }}

Victim (if different):
  {{ victim.name }}, aged {{ victim.age }}

────────────────────────────────────────────────────────────────
1.  PARTICULARS OF THE INCIDENT

    Date and time      : {{ incident.datetime }}
    Location           : {{ incident.location_description }}
    Coordinates        : {{ incident.lat }}, {{ incident.lon }}
    Ward / authority   : {{ jurisdiction.ward }} — {{ jurisdiction.authority }}
    Nature of defect   : {{ defect.description }}
    Nature of injury   : {{ injury.description }}
    Treatment at       : {{ hospital }}

2.  THE DEFECT WAS KNOWN AND UNATTENDED

    2.1  The defect at the said location was reported to {{ authority }} on
         {{ report_date }} vide reference {{ external_ref }}.       [Annexure B]
    2.2  A photographic record of the defect, geotagged and timestamped,
         was created on {{ first_observation_date }}, being {{ n }} days
         before the incident.                                       [Annexure C]
    2.3  {{ corroboration_count }} independent reports of the same defect
         were recorded between {{ from }} and {{ to }}.             [Annexure D]
    2.4  The defect was not attended to within 48 hours as directed by
         this Hon'ble Court.

3.  CONTRACTUAL RESPONSIBILITY (where established)

    3.1  The said stretch of road was constructed / resurfaced under contract
         {{ contract.external_id }} awarded to {{ contract.contractor }} for
         ₹{{ contract.value }}, completed on {{ contract.completed_on }}.
                                                                    [Annexure E]
    3.2  The defect liability period under the said contract runs until
         {{ contract.dlp_ends_on }}. The incident occurred within that period.
                                                                    [Annexure F]

4.  LOSS SUFFERED
    {{#each losses}}
    4.{{ n }}  {{ description }} — ₹{{ amount }}                    [Annexure {{ a }}]
    {{/each}}

5.  COMPENSATION CLAIMED

    Pursuant to the order dated 13.10.2025, the applicant claims compensation
    of ₹{{ claimed_amount }} on account of {{ death_or_injury_category }}.

6.  DECLARATION AND VERIFICATION

                                                {{ applicant.name }}
Place: {{ place }}                              (Signature)
Date:  {{ today }}
```

---

## Annexure set

| Annexure | Content | Source |
|---|---|---|
| A | Identity and, for legal heirs, proof of relationship | Citizen |
| B | Prior complaint(s) to the authority with reference numbers | Platform |
| C | Geotagged, timestamped photographic record of the defect **before** the incident | Platform |
| D | Independent corroborating reports | Platform |
| E | Contract documents and completion certificate | Platform (archive) |
| F | Defect liability period clause, verbatim with page number | Platform (extraction) |
| G | Medical records, discharge summary, bills | Citizen |
| H | Police record / FIR / accident report, if any | Citizen |
| I | Photographs of the vehicle or injury | Citizen |
| J | Proof of income and loss of earnings, if claimed | Citizen |
| K | The issue timeline, exported verbatim | Platform |

**Annexures B, C, D, E, F, and K are exactly what TraceSarkar exists to produce.** A citizen without
the platform would have to assemble each of them by hand, and most could not assemble C, D, E, or F
at all.

---

## Generation rules

1. **This instrument is generated only on explicit user initiation**, never suggested automatically
   by a classifier detecting an injury. That would be grotesque.
2. **The prior-observation record is the crux.** The strongest version of this claim is one where the
   platform's own timestamped record shows the defect existed and was reported *before* the incident.
   The generator surfaces this prominently and computes the interval.
3. **Contract attribution is included only above the publication threshold**, with sources cited.
4. **The claimed amount is computed from `legal_constants`**, keyed to the injury category, with the
   judgment citation.
5. **Interest tracking:** once filed, the 6–8 week disbursal window is tracked, and accruing 9%
   interest is displayed after it elapses.
6. **Counsel is recommended in the UI**, with legal-aid partner links where available.
7. **Never auto-filed.**

---

## Sensitivity

This instrument is generated in the aftermath of a death or a serious injury. Every interaction must
reflect that:

- No gamification, no points, no share prompts anywhere in this flow
- No "great job reporting!" copy
- Plain, quiet language throughout
- An offer of help — a real contact, not a chatbot — at every step
- The option to have someone else (an organiser, a family member) complete the process on the
  claimant's behalf

---

## Review checklist shown to the applicant

- [ ] The incident details are accurate
- [ ] You have the medical records and bills
- [ ] You have the police record, if one was made
- [ ] For a legal heir: you have proof of relationship
- [ ] The prior reports of this defect shown above are the correct location
- [ ] You have read the claimed amount and understand its basis
- [ ] **You have considered speaking to a lawyer.** We can put you in touch with legal aid
