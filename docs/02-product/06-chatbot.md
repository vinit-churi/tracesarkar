# The chatbot ("Ask TraceSarkar")

A conversational surface over the platform's own data. **It is a retrieval interface, not an
oracle.** Every answer is grounded in a record the user can open.

---

## 1. What it is for

| Question type | Example | Answer source |
|---|---|---|
| **About my issue** | "What happens now?" | Issue state machine + escalation decision table |
| **About procedure** | "How do I file an RTI for this?" | Legal framework docs + the generated draft |
| **About my area** | "How many potholes in R/S ward are still open?" | Live query over issues |
| **About a contractor** | "What else has this firm built in MMR?" | Contracts + scorecard |
| **About a rule** | "What is the SLA for a pothole?" | Config table with judgment citation |
| **Navigational** | "Show me my reports from last monsoon" | Query + UI deep link |

---

## 2. What it must never do

| Prohibited | Why |
|---|---|
| Give legal advice | We are not lawyers; every legal output carries a disclaimer and is a draft for review |
| Assert corruption or wrongdoing | Defamation exposure; the platform states facts and joins, never conclusions |
| Answer from parametric knowledge when the platform has no record | Hallucinated civic facts are worse than "I don't know" |
| File anything | Filing is always an explicit user tap, never a conversational side effect |
| Speculate about an individual official | Designations, not persons |
| Invent a contract, a ward, or a phone number | Every entity mentioned must resolve to a record |

**Enforcement:** the system prompt states these constraints, *and* a post-generation validator checks
that every entity named in the answer (contract IDs, contractor names, ward codes, phone numbers,
statutory provisions) exists in the retrieved context. An answer that names something not in context
is discarded and regenerated once, then falls back to "I don't have a record for that."

---

## 3. Architecture

```
user message
    │
    ▼
┌─────────────────┐
│ intent + scope  │  scoped to an issue? a ward? global? a contractor?
└────────┬────────┘
         ▼
┌─────────────────┐
│ tool selection  │  the model calls typed tools; it does not free-form SQL
└────────┬────────┘
         ▼
┌───────────────────────────────────────────────────────────┐
│ tools                                                     │
│  get_issue(id)               search_issues(filters)       │
│  get_contract(id)            search_contracts(filters)    │
│  get_contractor(id)          get_scorecard(contractor_id) │
│  get_authority(code)         get_sla(authority, category) │
│  get_ward_stats(ward, range) get_escalation_options(issue)│
│  get_legal_reference(topic)  get_my_reports(user)         │
└────────┬──────────────────────────────────────────────────┘
         ▼
┌─────────────────┐
│ grounded answer │  + citations to records + suggested actions
└────────┬────────┘
         ▼
┌─────────────────┐
│ validator       │  every named entity must exist in retrieved context
└─────────────────┘
```

**Tools, not text-to-SQL.** Typed tools with narrow parameters give bounded blast radius, cacheable
results, auditable calls, and no injection surface into the database. Text-to-SQL over a civic
database with personal data is not a risk worth taking.

Implementation notes and model configuration are in
[AI pipeline](../03-architecture/04-ai-pipeline.md).

---

## 4. Answer format

Every answer has three parts:

1. **The answer**, in the user's language, short.
2. **The basis** — links to the specific records used.
3. **The next action** — a tappable action, when one exists.

Example:

> **You reported this 62 days ago and BMC has not resolved it.**
>
> The Bombay High Court requires potholes to be attended within 48 hours (order dated 13 Oct 2025).
> That deadline passed on 13 June. Your MyBMC ticket MARG-2026-0611-88213 is still open.
>
> *Basis:* [issue 8f2a1c timeline] · [SLA config: road_defect/MH] · [ticket record]
>
> **→ Generate an RTI asking for the inspection and repair records**  ·  **→ Share this**

---

## 5. Language

Full support for English, Marathi, and Hindi at launch; Gujarati next.

- Input is accepted in any supported language, including code-mixed ("pothole kadhi fix honar?").
- Answers are generated **in the language of the question**, from the same fact set — never
  translated from a generated English answer.
- Legal and procedural terms keep their canonical form with a gloss ("प्रथम अपील (First Appeal)"),
  because the citizen will need the exact term on a government form.
- Voice input via Bhashini ASR; voice output via TTS for low-literacy users.

---

## 6. Scoping and privacy

| Scope | Data visible |
|---|---|
| Anonymous | Public aggregates and public issues only |
| Signed in | The above, plus the user's own reports and filings |
| Issue-scoped (opened from an issue) | That issue's full record plus related public context |
| Never | Another user's identity, another user's report history, exact coordinates of other users' reports |

Chat transcripts are retained only as long as needed for the session plus a short window for abuse
investigation, then deleted. They are covered by the DPDP consent notice.

---

## 7. Cost control

The chatbot is the easiest place to burn budget. Controls:

| Control | Detail |
|---|---|
| Prompt caching | The system prompt, tool definitions, and legal reference blocks are stable and cached |
| Effort tuning | Low effort for navigational and factual lookups; higher only for multi-step reasoning over several records |
| Retrieval budget | Hard cap on records injected per turn; summarise beyond it |
| Rate limits | Per account per hour, trust-scaled |
| Deterministic short-circuits | "What's the status of my report?" is answered by a template from the state machine, with no model call at all |

That last one matters more than it sounds: a large fraction of real questions are three or four
templatable intents. Route them deterministically and reserve the model for genuinely open questions.

---

## 8. Evaluation

A held-out question set, refreshed monthly, scored on:

| Dimension | Target |
|---|---|
| Groundedness (every claim traceable to a retrieved record) | 100% — any failure is a bug, not a metric |
| Correct refusal (says "no record" when there is none) | ≥ 98% |
| Correct action suggestion | ≥ 90% |
| Language fidelity (answers in the asked language) | ≥ 99% |
| Prohibited-content violations (legal advice, allegations) | 0 |

Failures are logged with the full retrieval context so regressions are reproducible.

---

## 9. Open questions

- Should the chatbot be able to *draft* an RTI directly in conversation, or only trigger the
  structured generator? *Current position: trigger the generator — a conversationally-produced legal
  document is harder to validate.*
- How do we handle questions about wards or authorities not yet covered? *Answer honestly with
  coverage status and offer to notify on launch.*
- Do we expose the chatbot to anonymous users? *Yes, scoped to public data — it is a strong
  acquisition surface and public data is public.*
