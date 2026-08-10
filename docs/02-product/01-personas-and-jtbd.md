# Personas and jobs-to-be-done

Five users. Only the first three matter for v1.

---

## P1 — Ravi, the annoyed commuter

**Profile:** 34, works in Andheri, lives in Vasai, rides a two-wheeler, Android mid-range, Marathi +
Hindi + workable English, has never filed an RTI.

**Job:** *"When I hit the same crater for the third week running, I want to do something that
actually costs someone something — in under 60 seconds, from the roadside, without learning
anything."*

| Currently | Because |
|---|---|
| Posts a photo to a ward WhatsApp group | It's the only channel he knows works socially |
| Occasionally tweets tagging @mybmc | Sometimes gets a reply, almost never a fix |
| Has downloaded a civic app and abandoned it | The form asked for a ward he didn't know |

**What wins him:** the moment he sees "this road is under warranty until 2028 and this firm has 41
open defects across MMR". That fact is *shareable*, and shareability is the only distribution channel
that compounds.

**Failure modes to avoid:** any form field he can't answer; any routing he can tell is wrong; any
outcome that reads as "registered" with no consequence.

**Success metric:** repeat reporting within 30 days.

---

## P2 — Sunita, the RWA secretary / hyperlocal organiser

**Profile:** 52, runs a housing-society federation in Kandivali, has filed RTIs before, knows the
ward officer's name, moderates a 400-member WhatsApp group.

**Job:** *"I want to accumulate evidence across my whole locality so that when I meet the AMC I'm not
one person complaining — I'm a documented pattern."*

**What she needs that Ravi doesn't:**
- Campaigns: group issues by corridor or society cluster
- Exportable evidence packs for meetings
- Someone-else's-issue tracking (she monitors, she doesn't only report)
- Volunteer verification missions she can distribute to her group
- Meeting-ready summaries: "our ward, this quarter, by category, with resolution rates"

**Why she matters disproportionately:** she is the distribution channel. Each Sunita is worth several
hundred Ravis, and she will onboard them herself if the tool makes her look effective.

**Success metric:** campaigns created; volunteer missions completed.

---

## P3 — Anjali, the local reporter

**Profile:** 29, city desk at a Mumbai daily, files 4–6 stories a week, deadline-driven, deeply
sceptical of civic-tech dashboards.

**Job:** *"Give me something I can put in a story today: a number nobody else has, sourced well
enough to survive my editor and the corporation's PR team."*

**What she needs:**
- The Watchdog TUI — fast, keyboard-driven, no marketing UI
- Saved queries and alerts ("tell me when any ward's monsoon report volume exceeds 3σ")
- Document search across the scraped corpus with OCR
- Exports with full provenance so every number is defensible
- Contractor scorecards with visible methodology
- A named human she can call to verify a figure

**What loses her instantly:** an unsourced claim, a broken link, or a number she can't reproduce.

**Why she matters:** one published story does more for the platform's legitimacy — and for the
political cost of ignoring it — than 10,000 app installs.

**Success metric:** stories published citing platform data.

---

## P4 — Prakash, the ward-level municipal officer

**Profile:** 45, Assistant Engineer, Roads, handling several hundred complaints a month across
multiple channels, personally liable under the 2025 HC directions.

**Job:** *"Show me the complaints that are real, are mine, and will get me in trouble — and let me
show that I fixed them."*

**Important framing:** TraceSarkar is adversarial to *the system's opacity*, not to Prakash. He is
often the person best positioned to fix things and worst served by the existing channels. Under the
2025 directions he faces personal liability for delays — which makes accurate, deduplicated,
correctly-routed, evidence-backed reports genuinely useful to him.

**What would make him an ally:**
- Correct routing (he is not blamed for MSRDC's road)
- Deduplication (100 reports of one pothole arrive as one issue with a count of 100)
- A proof-of-work upload path so his fix is publicly recorded
- No public shaming of individuals for things outside their control

**What we will not do:** build him a paid SaaS dashboard. That business model would compromise the
accountability layer. If a corporation wants integration, it consumes the public API or we build a
free adapter.

**Milestone:** v1+, and only via the public API.

---

## P5 — Dr. Menon, the researcher

**Profile:** urban-planning academic or think-tank analyst.

**Job:** *"Give me a clean, documented, longitudinal dataset of MMR infrastructure failure joined to
procurement, and I'll produce the analysis that changes policy."*

**Needs:** bulk export, stable schemas, versioned datasets, a data dictionary, a citation format, and
an OCDS publication of the contracts corpus.

**Why worth serving:** low marginal cost, high legitimacy return, and the OCDS publication is
independently valuable civic infrastructure.

**Milestone:** v0.5 (export) and v0.7 (OCDS publication).

---

## Jobs-to-be-done, consolidated

| JTBD | Persona | Feature response |
|---|---|---|
| "Report without learning anything" | Ravi | A1, A2, B1, B8 |
| "Make it cost someone something" | Ravi, Sunita | C3–C6, D-series |
| "Make it shareable" | Ravi | E1–E3 |
| "Accumulate a pattern" | Sunita | H7, E4, G3 |
| "Distribute work to my group" | Sunita | I6 |
| "Find a number nobody else has" | Anjali | G1–G4 |
| "Defend the number to my editor" | Anjali | provenance chain, F8 |
| "See only what's real and mine" | Prakash | B-series accuracy, dedup, public API |
| "Analyse it properly" | Dr. Menon | G3, C2 |

---

## Anti-personas

Users we deliberately do not optimise for, and the design consequences.

| Anti-persona | Behaviour | Design response |
|---|---|---|
| **The party worker** | Bulk-reports in an opponent's ward; suppresses reports in their own | Trust scoring, device/graph clustering, corroboration requirements, no party tagging |
| **The contractor's PR agency** | Files disputes en masse; astroturfs "resolved" confirmations | Right-of-reply is a documented single-instance process, not a bulk API; confirmation requires photographic proof |
| **The rage poster** | Wants an outrage feed, not a fix | Comments are evidence-only; the feed ranks by issue state, not by engagement |
| **The scraper** | Bulk-harvests citizen data | Public API returns coarsened locations and no PII; rate limits; ToS |
| **The vendor** | Wants to sell the corporation a dashboard using our data | AGPL on the code, CC BY-SA on the data, and an explicit non-endorsement stance |

---

## The one-sentence test

Before any feature ships, it must answer this for at least one persona:

> **"After this ships, [persona] can do [specific thing] that they demonstrably could not do before,
> in under [time], without [prior knowledge they don't have]."**

If it can't be phrased that way, it is a dashboard feature and it goes back in the catalog.
