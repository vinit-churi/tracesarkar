# Template — NGT Original Application (Western Zone Bench, Pune)

**Status:** draft skeleton for engineering. **Must be reviewed by counsel before first use.** An NGT
application is a court filing; a defective one can be dismissed and can prejudice the cause of
action.

The platform's job is **assembly, not advocacy**: it compiles a chronology, an evidence index, and a
statement of facts from records it already holds. The legal framing is a human's work.

---

## Why this instrument matters

Limitation is the commonest way meritorious environmental petitions die:

| Provision | Limitation |
|---|---|
| **Section 14** (original application) | **6 months** from when the cause of action first arose, extendable by up to **60 days** for sufficient cause |
| **Section 15** (compensation / restitution) | **5 years**, extendable by up to 60 days |

TraceSarkar holds an immutable, timestamped, geotagged observation trail. It can therefore establish
the cause-of-action date precisely — which is exactly the fact a limitation objection turns on.

**Product consequence:** the moment an issue is classified as environmental, a 6-month countdown
starts and is surfaced to the user.

---

## Structure

```
BEFORE THE HON'BLE NATIONAL GREEN TRIBUNAL
WESTERN ZONE BENCH, PUNE

ORIGINAL APPLICATION NO. ______ OF {{ year }}

IN THE MATTER OF:
  Sections 14, 15, 18 and 20 of the National Green Tribunal Act, 2010
  {{ additional_statutes }}          e.g. Environment (Protection) Act 1986;
                                     Water (Prevention and Control of Pollution) Act 1974;
                                     Air (Prevention and Control of Pollution) Act 1981;
                                     Maharashtra Regional and Town Planning Act 1966

AND IN THE MATTER OF:
  {{ applicant.name }}, {{ applicant.address }}                        …APPLICANT

                              VERSUS

  1. {{ authority.name }}, through its Commissioner / Chief Executive   …RESPONDENT No. 1
  2. Maharashtra Pollution Control Board                                …RESPONDENT No. 2
  3. {{ contractor.name }} (if applicable)                              …RESPONDENT No. 3
  4. {{ other_respondents }}

────────────────────────────────────────────────────────────────────────
SYNOPSIS
{{ one_paragraph_summary }}

────────────────────────────────────────────────────────────────────────
LIST OF DATES AND EVENTS

{{#each timeline_entries}}
  {{ date }}   {{ description }}                          [Annexure {{ annexure }}]
{{/each}}

────────────────────────────────────────────────────────────────────────
1. PARTICULARS OF THE APPLICANT
2. PARTICULARS OF THE RESPONDENTS
3. JURISDICTION OF THE TRIBUNAL
     3.1  The environmental breach complained of occurred at {{ location }},
          within the territorial jurisdiction of this Hon'ble Tribunal under Rule 11.
     3.2  The application is filed within the limitation prescribed under Section 14(3).
          The cause of action first arose on {{ cause_of_action_date }}, being the date on
          which the applicant first observed and recorded the said activity. The present
          application is filed {{ days_elapsed }} days thereafter.

4. FACTS OF THE CASE
     {{ statement_of_facts }}

5. GROUNDS
     {{ grounds }}

6. RELIEFS SOUGHT
     {{#each reliefs}}
       {{ letter }}) {{ text }}
     {{/each}}

7. INTERIM RELIEF (if any)
8. CAVEAT / DECLARATION
9. VERIFICATION

────────────────────────────────────────────────────────────────────────
INDEX OF ANNEXURES
{{#each annexures}}
  Annexure {{ letter }}  —  {{ description }}                       (pages {{ from }}–{{ to }})
{{/each}}
```

---

## What the platform generates automatically

| Section | Source |
|---|---|
| **List of Dates and Events** | `issue_events` — the append-only, hash-chained timeline. This is the single most valuable auto-generated section |
| **Particulars of respondents** | `authorities` + `contractors`, with the resolved jurisdiction |
| **Jurisdiction and limitation paragraph** | Computed from the first observation timestamp |
| **Statement of facts (draft)** | Assembled from the issue record, observations, and news context — flagged for human editing |
| **Annexure index** | Photographs with capture metadata, RTI applications and replies, correspondence, official reference numbers, news items, source documents with hashes |
| **Photographic evidence pack** | Each photograph with coordinates, timestamp, direction, and the archival hash |

## What a human must supply

| Section | Why |
|---|---|
| Grounds | Legal argument, not fact assembly |
| Reliefs sought | Requires judgment about what to ask for |
| Additional statutes pleaded | Depends on the specific violation |
| Interim relief | Tactical |
| Verification and signature | Statutorily personal |

---

## Standard relief templates (starting points, to be edited)

| Relief | Text |
|---|---|
| Restoration | "Direct Respondent No. 1 to restore the {{ feature }} at {{ location }} to its condition prior to the said activity, at the cost of the person responsible" |
| Cessation | "Direct the immediate cessation of {{ activity }} at {{ location }}" |
| Environmental compensation | "Direct the levy of environmental compensation on the polluter under Section 15 read with Section 20, on the polluter-pays principle" |
| Inspection and report | "Direct Respondent No. 2 to inspect {{ location }} and file a report before this Hon'ble Tribunal within {{ n }} weeks" |
| Committee | "Constitute a joint committee comprising {{ bodies }} to inspect and report" |
| Action against officials | "Direct Respondent No. 1 to fix responsibility on the officials whose inaction permitted the said activity" |

---

## Category → statute mapping (indicative; verify per case)

| Issue | Likely statutes to plead |
|---|---|
| Debris dumping in mangroves | EPA 1986; CRZ Notification; Forest Conservation Act; MRTP Act |
| Effluent into a creek | Water Act 1974; EPA 1986 |
| Illegal tree felling | Maharashtra (Urban Areas) Protection and Preservation of Trees Act 1975 |
| Construction dust / air pollution | Air Act 1981; EPA 1986; relevant C&D waste rules |
| Nallah encroachment | MRTP Act; municipal corporation act provisions |
| Solid waste burning | Solid Waste Management Rules 2016; Air Act 1981 |

⚠️ This mapping is indicative only and must be confirmed per case by counsel.

---

## Generation rules

1. **Never file automatically.** The NGT application is generated as a draft, downloaded, and filed by
   the applicant or their counsel. See [ADR 0007](../04-adr/0007-never-auto-file.md).
2. **Limitation is computed and displayed prominently**, with the cause-of-action date and the days
   remaining.
3. **Every fact traces to a record.** No fact appears in the statement of facts that is not in the
   issue timeline or an annexure.
4. **Every photograph carries its metadata** — coordinates, timestamp, and archival hash — because
   authenticity will be challenged.
5. **The draft is marked as a draft** on every page.
6. **Counsel review is recommended in the UI**, with a link to legal-aid partners where available.
7. **No legal conclusions are asserted by the platform.** Grounds are a human's work.

---

## Review checklist shown to the citizen

- [ ] The cause-of-action date is correct — is this genuinely when you first observed it?
- [ ] You are within 6 months of that date (currently: {{ days_remaining }} days remaining)
- [ ] The respondents are the right bodies
- [ ] Every annexure referred to is actually attached
- [ ] The statement of facts is accurate and contains nothing you cannot support
- [ ] You have read the grounds and reliefs, and they say what you want to say
- [ ] **You have considered obtaining legal advice.** This is a filing before a tribunal
