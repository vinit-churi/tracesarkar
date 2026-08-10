# Template — Consumer complaint (e-Jagriti)

**Status:** draft skeleton for engineering. **Must be reviewed by someone competent in consumer law
before first use.** Fee slabs must be verified against the current schedule — see
[legal framework](../01-research/04-legal-framework.md).

---

## When this applies

Under the Consumer Protection Act, 2019, a citizen who pays taxes and charges for municipal services
(water supply, sanitation, road maintenance) is a consumer of those services, and their failure can
be pleaded as **deficiency in service**.

Strongest where there is **quantifiable loss**:

| Loss type | Evidence needed |
|---|---|
| Vehicle damage from a road defect | Repair invoice, mechanic's report, photographs |
| Medical expenses from a fall or accident | Medical records, bills, police record if any |
| Property damage from flooding or a burst line | Photographs, repair estimates, insurance correspondence |
| Loss of earnings | Employer letter, medical certificate |

**Filing:** `e-jagriti.gov.in` (launched 1 January 2025, replacing eDaakhil). District, State, and
National commissions. OTP-based registration. Filing is free for district claims up to ₹5 lakh;
₹200–₹7,500 above (⚠️ verify the current slab table).

---

## Structure

```
BEFORE THE {{ district }} DISTRICT CONSUMER DISPUTES REDRESSAL COMMISSION

CONSUMER COMPLAINT NO. ______ OF {{ year }}

  {{ complainant.name }}
  {{ complainant.address }}                                    …COMPLAINANT

                            VERSUS

  {{ authority.name }},
  through its Commissioner / Chief Officer,
  {{ authority.address }}                                      …OPPOSITE PARTY

COMPLAINT UNDER SECTION 35 OF THE CONSUMER PROTECTION ACT, 2019

────────────────────────────────────────────────────────────────
1.  The Complainant is a resident of {{ locality }} and a consumer within the
    meaning of Section 2(7) of the Act, having paid {{ tax_or_charge_description }}
    to the Opposite Party.                                    [Annexure A]

2.  The Opposite Party is a statutory body responsible for {{ service_description }}
    within {{ jurisdiction }}.

3.  FACTS
    {{#each numbered_facts}}
    3.{{ n }}  {{ text }}                                     [Annexure {{ annexure }}]
    {{/each}}

4.  DEFICIENCY IN SERVICE
    4.1  The Opposite Party was under an obligation to {{ obligation }}.
    4.2  Despite the complaint dated {{ complaint_date }} bearing reference
         {{ external_ref }}, the Opposite Party failed to attend to the said
         defect within {{ sla_hours }} hours as required.     [Annexure {{ x }}]
    4.3  The said failure constitutes a deficiency in service within the meaning
         of Section 2(11) of the Act.

5.  LOSS AND INJURY SUFFERED
    {{#each losses}}
    5.{{ n }}  {{ description }} — ₹{{ amount }}              [Annexure {{ annexure }}]
    {{/each}}
    Total: ₹{{ total_loss }}

6.  CAUSE OF ACTION
    The cause of action arose on {{ cause_of_action_date }} and is continuing.

7.  JURISDICTION AND LIMITATION
    7.1  The Opposite Party carries on business within the territorial
         jurisdiction of this Commission.
    7.2  The value of the claim is ₹{{ claim_value }}, within the pecuniary
         jurisdiction of this Commission.
    7.3  The complaint is filed within two years of the cause of action.

8.  RELIEF SOUGHT
    {{#each reliefs}}
    ({{ letter }})  {{ text }}
    {{/each}}

9.  The requisite fee of ₹{{ fee }} has been paid.

VERIFICATION
I, {{ complainant.name }}, verify that the contents of paragraphs 1 to 9 are
true to my knowledge and belief, and that nothing material has been concealed.

Place: {{ place }}                                {{ complainant.name }}
Date:  {{ today }}                                (Complainant)
```

---

## Relief templates

| Relief | Text |
|---|---|
| Direction to repair | "Direct the Opposite Party to repair / restore {{ asset }} at {{ location }} within {{ n }} weeks" |
| Compensation for loss | "Direct the Opposite Party to pay ₹{{ amount }} towards the loss and damage suffered" |
| Compensation for mental agony | "Direct the Opposite Party to pay ₹{{ amount }} towards mental agony and harassment" |
| Litigation costs | "Direct the Opposite Party to pay ₹{{ amount }} towards the costs of this complaint" |
| Systemic direction | "Direct the Opposite Party to publish and adhere to a time-bound mechanism for attending to complaints of this nature" |

---

## Generation rules

| Rule | Detail |
|---|---|
| **Facts only** | The generator actively strips emotional language. "The road remained unrepaired for 62 days" ✔. "The BMC's shocking apathy" ✘. Consumer commissions respond to dates, documents, and amounts |
| **Quantify or omit** | A claim without a quantified loss is much weaker. If no loss is quantifiable, prompt the citizen and suggest RTI first |
| **Consumer status must be pleaded** | Paragraph 1 requires evidence of the tax or charge paid — property tax receipt, water bill. The generator prompts for it |
| **SLA breach is the core of the deficiency** | Cite the applicable SLA with its source from `legal_constants` (for road defects in Maharashtra, the Bombay HC 48-hour direction) |
| **Pecuniary jurisdiction check** | The generator selects District / State / National based on the claim value and flags it |
| **Limitation** | Two years from the cause of action; computed and displayed |
| **Never auto-filed** | The citizen registers on e-Jagriti and files. See [ADR 0007](../04-adr/0007-never-auto-file.md) |

---

## Class-action variant (v2+)

Where multiple citizens report the same asset and suffer comparable loss, Section 35(1)(c) permits a
complaint by one or more consumers on behalf of numerous consumers with the same interest.

The platform is unusually well-placed to assemble this: it already holds the corroborated reports,
the shared asset, the common authority, and the individual loss records. But it requires:

- Explicit, individual consent from each participating consumer
- Verification of each claimed loss
- A lead complainant willing to be named
- Almost certainly, counsel

Treat as a v2 feature with a legal-partnership prerequisite, not as an automation target.

---

## Review checklist shown to the citizen

- [ ] You have proof that you pay for this service (property tax receipt, water bill)
- [ ] Every loss claimed has a document supporting it
- [ ] The amounts are accurate — you may be asked to prove each one
- [ ] The complaint reference and dates are correct
- [ ] You are within two years of the cause of action
- [ ] The claim value is within this Commission's pecuniary jurisdiction
- [ ] You understand you are filing this in your own name and may need to attend hearings
