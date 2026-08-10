# Template — RTI application

**Status:** draft template for engineering. **Must be reviewed by someone competent in RTI practice
before first use.** Fees and format must be verified against the current Maharashtra RTI Rules at
generation time — see [legal framework](../01-research/04-legal-framework.md).

Placeholders are `{{ }}`. Values in `[[ ]]` come from `legal_constants` and carry citations.

---

## Structure

```
TO,
The Public Information Officer
{{ authority.rti.pio.designation }}
{{ authority.rti.pio.department }}
{{ authority.rti.pio.address }}

SUBJECT:  Application under Section 6(1) of the Right to Information Act, 2005 —
          information regarding {{ issue.category_label }} at {{ issue.location_description }}

Sir / Madam,

1. I am a citizen of India and seek the following information under the Right to
   Information Act, 2005.

2. PARTICULARS OF THE MATTER

   Location            : {{ issue.location_description }}
   Coordinates         : {{ issue.location.lat }}, {{ issue.location.lon }}
   Ward                : {{ issue.ward.code }} — {{ issue.ward.name }}
   Nature of the issue : {{ issue.category_label }} ({{ issue.subcategory }})
   First observed on   : {{ issue.first_reported_at | date }}
   Reported to         : {{ filing.channel_name }} on {{ filing.filed_at | date }}
   Complaint reference : {{ filing.external_ref }}
   Status as on date   : {{ issue.status_label }}
   Days elapsed        : {{ issue.days_since_report }}

3. INFORMATION SOUGHT

{{#each questions}}
   ({{ index }})  {{ text }}
{{/each}}

4. I request that the information be provided in {{ preferred_format }} form.

5. The application fee of [[ rti.fee.MH ]] is paid by {{ payment_method }}.
   {{#if bpl}}
   I belong to a Below Poverty Line household; a copy of the certificate is enclosed
   and no fee is payable under Section 7(5).
   {{/if}}

6. ENCLOSURES
{{#each annexures}}
   Annexure {{ letter }} — {{ description }}
{{/each}}

Yours faithfully,

{{ applicant.name }}
{{ applicant.address }}
{{ applicant.phone }}
Date: {{ today }}
Place: {{ applicant.place }}
```

---

## Question sets by category

The single most important part of the instrument. Vague RTIs get rejected; specific ones do not.
Each question must be answerable from a document the authority holds.

### `road_defect`

1. Please provide the name and address of the contractor responsible for the construction /
   resurfacing / maintenance of the road at {{ location }}, together with the contract or work order
   number.
2. Please provide the date of the work order, the date of completion, and a copy of the completion
   certificate for the said work.
3. Please state the defect liability period / guarantee period applicable to the said work, and the
   date on which it expires, with a copy of the relevant clause of the contract.
4. Please provide the total amount sanctioned and the total amount paid to the contractor for the
   said work, with dates of each payment.
5. Please state the amount, if any, retained or withheld against the guarantee period, and its
   current status.
6. Please provide copies of all inspection reports for the said road for the period
   {{ from_date }} to {{ to_date }}.
7. Please state the number of complaints received regarding the said location during the said
   period, with their dates and current status.
8. Please state what action has been taken on complaint reference {{ filing.external_ref }} dated
   {{ filing.filed_at }}, with the date of each step taken.
9. Please state whether any penalty has been imposed on the contractor in respect of defects at this
   location, and if so, provide details; if not, please state the reasons.
10. Please provide the name and designation of the officer responsible for supervising the said work
    and for attending to complaints at this location.

### `waste`

1. Contract or work order number, contractor name, and address for solid waste collection in
   {{ ward }} ward, for the period {{ from_date }} to {{ to_date }}.
2. The collection frequency stipulated in the said contract for {{ location }}.
3. Weighment records for waste collected from {{ location }} for the last 30 days.
4. Penalty clauses in the said contract for missed collection, and details of any penalties imposed
   during the said period.
5. Action taken on complaint reference {{ filing.external_ref }}.
6. Name and designation of the supervising officer.

### `water_drainage`

1. Contract number, contractor name, and quantum of de-silting work sanctioned for
   {{ nallah_or_drain }} for the current year.
2. Copy of the pre-monsoon inspection report for the said drain.
3. Records of silt removed, with dates, quantities, and disposal location.
4. Copies of any weighment or measurement records relied on for payment.
5. Amount sanctioned and amount paid for the said work.
6. Action taken on complaint reference {{ filing.external_ref }}.

### `structural`

1. Date of the last structural audit of {{ structure_name }} and a copy of the audit report.
2. Name of the agency that conducted the audit and the contract under which it was conducted.
3. Recommendations made in the audit and the action taken on each.
4. Contract number and contractor for any repair work carried out, with dates.
5. Present classification of the structure (safe / repairable / dilapidated) and the basis for it.

### `environment`

1. Copies of all notices issued in respect of {{ activity }} at {{ location }} under the applicable
   provisions, with dates.
2. Action-taken report on each such notice.
3. Whether the said activity has environmental clearance / consent to operate; if so, a copy.
4. Complaints received regarding this location during {{ from_date }} to {{ to_date }} and their
   disposal.
5. Name and designation of the officer responsible for enforcement at this location.

---

## Generation rules

| Rule | Detail |
|---|---|
| **PIO must be resolved**, not guessed | From `authority_departments.pio`. If unknown, the instrument is not generated; the platform instead offers an RTI seeking the PIO details |
| **Central vs state portal** | Branch on authority kind. Railways and other union bodies go to `rtionline.gov.in`, not the Maharashtra portal |
| **Fee from `legal_constants`** | With its citation and effective-from date, verified at generation time |
| **Facts from the issue record only** | No model-invented facts; every particular is a field |
| **Questions capped at 10** | Long RTIs invite rejection for being omnibus. Prioritise |
| **Each question answerable from one document** | Questions that ask for opinion or inference are rejected under the Act |
| **Applicant is the citizen** | Never the platform |
| **Draft status unmissable** | On screen and on the rendered document |
| **Disclaimer** | "This is a generated draft, not legal advice. Review before filing." |

---

## Post-filing

Once the citizen records the registration number:

- The 30-day clock starts
- Reminders at T-7 and T-2 before the reply is due
- On expiry with no reply, or an inadequate reply, the **First Appeal** instrument becomes available,
  pre-filled with the RTI particulars and the specific deficiency
- On first-appeal expiry, the **Second Appeal** to the State Information Commission becomes available

---

## Review checklist shown to the citizen

- [ ] The PIO's designation and address look right for this department
- [ ] The location description matches where the problem actually is
- [ ] The questions ask for what you actually want to know
- [ ] Your name, address, and phone number are correct
- [ ] You have the fee ready (`[[ rti.fee.MH ]]`)
- [ ] You understand this is a draft you are filing in your own name
