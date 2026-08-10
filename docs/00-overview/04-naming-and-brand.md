# Naming, voice, and brand

---

## The name

**TraceSarkar** — *trace* (to follow, to track, to find the origin of) + *sarkar* (सरकार, government).

| Property | Assessment |
|---|---|
| Says what it does | ✔ "Trace the government" |
| Bilingual, works in Marathi/Hindi/English mouths | ✔ |
| Not party-affiliated | ✔ — *sarkar* is institutional, not partisan |
| Not region-locked | ✔ — starts in MMR, does not preclude expansion |
| Pronounceable and typeable | ✔ |
| Domain/handle availability | ⚠️ To be checked before public launch |
| Trademark conflict | ⚠️ To be checked before public launch |

The working title in early notes was **CivicPulse MMR**. That name is retired: "pulse" implies
monitoring and dashboards, which is exactly the category we are trying not to be in, and the "MMR"
suffix bakes in a geographic ceiling.

**Blocking task before public launch:** trademark and domain search. The name is cheap to change now
and expensive to change later.

---

## What the name commits us to

"Trace" sets an expectation the product must meet:

- Every fact is traceable to a source
- Every routing decision shows its basis
- Every attribution shows how it was matched
- Every correction is logged publicly

If the platform ever displays something a user cannot trace back, the name becomes a lie. That is a
useful constraint.

---

## Voice

| Principle | In practice |
|---|---|
| **Factual, never inflammatory** | "48 hours have passed" not "BMC has failed you again" |
| **Specific, never vague** | "Contract WS/2023/ROAD/117, ₹4.11 crore" not "crores of taxpayer money" |
| **Calm** | The situation is often outrageous. The interface is not |
| **Second person, active** | "You can ask for the records" not "An RTI may be generated" |
| **Plain register** | "The 48-hour deadline has passed" not "SLA breach recorded" |
| **Honest about gaps** | "We don't have contract data for this ward yet" not silence |
| **Never triumphant about failure** | A confirmed fix is celebrated. A breach is reported, not gloated over |

### Words we use

report · issue · authority · ward · contract · warranty · deadline · confirmed · reopened · source ·
draft · evidence

### Words we do not use

corrupt · scam · loot · negligent (as an accusation) · shameful · exposed · caught · guilty ·
crackdown · war on \_\_\_

The second list is not squeamishness. Every one of those words converts a factual record into an
allegation, and an allegation is what a defamation claim needs. See
[legal framework §8](../01-research/04-legal-framework.md#8-publication-risk-defamation-and-intermediary-liability).

---

## Tone by context

| Context | Tone |
|---|---|
| Reporting flow | Brisk, encouraging, out of the way |
| Attribution reveal | Matter-of-fact. The facts are dramatic enough |
| SLA breach | Direct, actionable, unsurprised |
| Confirmed fix | **Genuinely celebratory** — this is the one place to be warm |
| Compensation claim (death/injury) | Quiet, plain, helpful. No gamification, no share prompts, no cheerfulness |
| Rejection ("not a civic issue") | Specific and helpful, never a generic error |
| Error states | Honest about what broke and what happens to their report |

---

## Visual principles

Not a full design system — the constraints that matter.

| Principle | Reason |
|---|---|
| **The photograph is the hero** | It is the evidence and the emotional content |
| **Platform statements look different from citizen content** | A legal requirement, not a style choice |
| **Data density over decoration** for power surfaces | Watchdog TUI, ward pages |
| **Colour never carries meaning alone** | Accessibility; also, colour-coded severity is culturally variable |
| **No dark patterns** | Unsubscribe is one tap. Consent is real. Nothing is pre-ticked |
| **Works at 200% text zoom** | WCAG 2.2 AA |
| **Loads on 2G** | Hard constraint on the public pages |

### Neutral palette discipline

Avoid saffron, green, and blue as *dominant* brand colours — each maps onto Indian party symbolism.
Party-neutral perception is a functional requirement, not an aesthetic preference.

---

## What the platform never says

| Never | Instead |
|---|---|
| "This contractor is corrupt" | "This contract's defect liability period is active" |
| "BMC doesn't care" | "BMC has not responded within 48 hours" |
| "Exposed: the truth about…" | "Contract WS/2023/ROAD/117: awarded, completed, in warranty" |
| "Join the fight against…" | "Report what you see" |
| "X% of politicians are…" | Nothing. We do not model politicians |
| "Guaranteed to get results" | "Here is what you can do next" |
| "Verified by AI" | "Classified automatically — confirm if this is wrong" |

---

## Attribution requirements

Downstream users of our data must carry attribution (CC BY-SA 4.0). Our own outputs carry it too:

- Share images: the TraceSarkar mark and URL burned in
- Exports: a provenance manifest with sources, versions, and caveats
- Embeds: a visible attribution link
- API responses: `sources[]` on every record

---

## Open questions

- Domain and handle availability, and trademark search — **blocking before public launch**
- Does the name read as adversarial to a municipal officer we want as an ally? Test it with one.
  *Current view: "trace" is neutral and the framing conversation carries more weight than the name.*
- Marathi-first naming of the product surfaces (`ट्रेससरकार`?) — ask native speakers, do not guess
