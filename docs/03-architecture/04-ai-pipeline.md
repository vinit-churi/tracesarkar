# AI pipeline

Every model call in the system, with its purpose, schema, model tier, cost control, and failure
behaviour.

**Stance:** the model is a structured-extraction engine and a language surface. It is never an
authority on civic facts, never a decision-maker on publication, and never in the path of a legal
filing without human review.

---

## 1. Model selection

Defaults, from the current Claude model catalog:

| Call | Model | Rationale |
|---|---|---|
| Image classification (interactive) | `claude-opus-5` | Vision quality directly determines routing correctness; 1M context, $5/$25 per MTok |
| Image classification (backfill/bulk) | `claude-opus-5` via the **Batch API** | 50% cost reduction; latency irrelevant for backfill |
| Contract PDF extraction | `claude-opus-5` | Legal-document extraction; accuracy dominates cost |
| Share-kit text generation | `claude-haiku-4-5` | Short, templated, high volume; $1/$5 per MTok |
| Chatbot | `claude-opus-5` with tools | Grounded retrieval + reasoning |
| Escalation drafting | `claude-opus-5` | Legal instruments; highest accuracy tier |
| News entity extraction | `claude-haiku-4-5` | High volume, simple schema |

**Do not silently downgrade a model to save cost.** A tier change is a documented decision with an
eval-set comparison, recorded as an ADR.

### Shared configuration

- **Structured outputs** (`output_config.format` with a JSON schema) on every extraction and
  classification call. Never parse free text into a domain object.
- **Adaptive thinking** with `output_config.effort` tuned per call: `low` for classification and
  share text, `high` for contract extraction and escalation drafting.
- **Prompt caching:** system prompt, taxonomy, and tool definitions are stable and placed before any
  volatile content, so the prefix caches. Verify with `usage.cache_read_input_tokens`.
- **Streaming** for anything with a large `max_tokens`.

---

## 2. Call: image classification

**Input:** 1–3 images + optional user text.
**Output:** the schema in
[snap-to-action pipeline §Stage 1](../02-product/04-snap-to-action-pipeline.md#stage-1--classify).

### Vision constraints that matter here

| Constraint | Value | Consequence |
|---|---|---|
| Max resolution (Opus 5 tier) | 2576 px on the long edge | Send at native resolution up to this; coordinates map 1:1 to pixels |
| Token cost per image | up to ~4,784 tokens at full resolution | ~3× a pre-4.7-generation image. Downsample only if the eval set shows no accuracy loss |
| Practical setting | 1080p long edge | Good accuracy/cost balance for civic photos; validate against the eval set before locking it in |

Cost sanity check at the interactive tier: a 3-image classification at ~1,500 tokens per image plus
a ~600-token prompt is roughly 5,100 input tokens ≈ **₹2.2 per report** at $5/MTok and ₹88/USD.
That is the dominant per-report cost and the primary reason for the batch path and the caching
strategy. See [cost model](../06-operations/04-cost-model.md).

### Prompt structure (cache-friendly ordering)

```
[system]  ← stable, cached
  role, taxonomy, severity definitions, hazard criteria,
  output schema description, refusal rules, MMR-specific guidance
  (monsoon conditions, common Indian road materials, typical signage)

[user]    ← volatile
  images
  optional citizen text
  optional coarse locality hint ("Kandivali West") — never precise coordinates
```

The locality hint is a **coarse** string, not coordinates. It helps with landmark reading without
sending precise personal location data to a third-party API.

### Guardrails

- `is_civic_issue: false` → polite, specific rejection; never a generic error
- `people_present: true` → redaction pipeline is mandatory before any public derivative
- `image_quality: "unusable"` → ask for a retake rather than guessing
- Low confidence → present the top 3 categories for the user to pick from
- **No `hazard_to_life` decision is model-final:** any `true` immediately routes to the emergency
  channel; any `false` on a subcategory in the known-hazard list (open manhole, live wire, collapse)
  is overridden to `true` by rule

---

## 3. Call: contract PDF extraction

**Input:** contract PDF (via the Files API or base64 document blocks).
**Output:** the DLP / work-site / penalty schema in
[tender engine §3](06-tender-engine.md#3-stage-3-extraction-the-dlp-problem).

| Setting | Value |
|---|---|
| Model | `claude-opus-5` |
| Effort | `high` |
| Structured output | Required |
| Citations | Enabled where the document format supports it — page provenance is mandatory |
| Batch | Yes for backfill |

**Non-negotiable:** every extracted legal term carries a `verbatim` quote and a page number. An
extraction without them is discarded, not published.

Documents over the per-request page limit are chunked by page range, extracted per chunk, and merged
with conflict detection. Two chunks reporting different DLP values is a review-queue event, not a
silent pick.

---

## 4. Call: chatbot

Tool-based retrieval. See [chatbot](../02-product/06-chatbot.md) for the full specification.

| Setting | Value |
|---|---|
| Model | `claude-opus-5` |
| Effort | `low` for navigational queries, `high` for multi-record reasoning |
| Tools | Typed tools only — **never text-to-SQL** |
| Caching | System prompt + tool definitions cached |
| Validation | Post-generation entity check against retrieved context |

**Deterministic short-circuit:** template-answer the top intents ("what's the status of my report?",
"what happens next?") from the state machine with **no model call**. A large fraction of real
questions are three or four intents; routing them deterministically is the single biggest cost lever
in the whole pipeline.

---

## 5. Call: share-kit text generation

**Input:** the issue fact set (structured), the target channel, the target language.
**Output:** post text within channel constraints.

| Setting | Value |
|---|---|
| Model | `claude-haiku-4-5` |
| Effort | `low` |
| Structured output | Yes — `{text, hashtags[], handles[], char_count}` |
| Post-validation | Prohibited-pattern check (allegations, insults, party references, personal handles) before display |

Generated **per language from the fact set**, never translated from a generated English string —
translating a hashtag-laden tweet produces garbage. See
[accessibility and vernacular](10-accessibility-and-vernacular.md).

---

## 6. Call: escalation drafting

**Input:** issue record, timeline, jurisdiction, contract attribution, instrument type, applicable
legal constants.
**Output:** structured instrument content, rendered to PDF/DOCX by a deterministic template engine.

| Setting | Value |
|---|---|
| Model | `claude-opus-5` |
| Effort | `high` |
| Structured output | Required — the model fills a schema; it does not write a free-form document |

**Architecture matters here:** the model produces *structured content* (question list, statement of
facts, list of dates and events, reliefs sought). A deterministic renderer produces the document.
This means:

- Formatting, margins, and annexure numbering are code, not model output
- Statutory citations come from the `legal_constants` table, not from the model
- The output is diffable and reviewable
- A prompt regression cannot silently corrupt a legal document's structure

**Every generated instrument is a draft.** It is labelled as such, carries a disclaimer, and requires
an explicit human action to file.

---

## 7. Call: news entity extraction

**Input:** headline + extract.
**Output:** `{places[], authorities[], contractors[], categories[], location_hint}`.

`claude-haiku-4-5`, `low` effort, batched. Used to link news items to issues and wards.

---

## 8. Redaction (not a Claude call)

Face and number-plate detection runs **on-device where possible** and server-side as a fallback,
using a local model — not a hosted API. Rationale: it runs on every image, latency matters, and
sending un-redacted images containing identifiable people to a third party before redaction is
exactly what we are trying to avoid.

Server-side fallback runs inside our own infrastructure. Redaction results are recorded in
`report_media.redaction` with the model version so a redaction-model improvement can trigger a
re-run over the corpus.

---

## 9. Cost controls

| Control | Detail |
|---|---|
| **Prompt caching** | Stable prefixes on every repeated call type. Monitor `cache_read_input_tokens`; a zero read rate means something volatile leaked into the prefix |
| **Batch API** | 50% discount for all backfill and re-processing |
| **Model tiering** | Haiku for high-volume simple schemas; Opus where accuracy is legally load-bearing |
| **Deterministic short-circuits** | Template answers for common chatbot intents; rule-based classification for unambiguous cases (a report with subcategory pre-selected by the user skips classification) |
| **Image sizing** | Validate the accuracy cost of downsampling to 1080p; adopt if the eval set is flat |
| **Daily budget cap** | A hard global cap with graceful degradation to queue-and-classify-later, never a 5xx |
| **Per-account rate limits** | Trust-scaled |
| **Dedup before classify** | If a report merges into an existing issue on spatial+phash grounds alone, skip the classification call entirely |

That last one is significant during a viral event: 200 reports of one pothole should produce
approximately one classification call.

---

## 10. Evaluation

Every call type has a versioned eval set, run in CI on any prompt or model change.

| Call | Set | Metric | Target |
|---|---|---|---|
| Classification | 500 labelled MMR photos across all categories, incl. night, rain, motion blur | Category accuracy | ≥ 92% |
| Classification | 100 hazard photos | Hazard recall | ≥ 99% (false negatives are dangerous) |
| Classification | 100 non-civic photos | Rejection precision | ≥ 95% |
| DLP extraction | 200 labelled contracts | Field accuracy | ≥ 90% |
| Work-site extraction | 200 labelled contracts | Usable-geometry rate | ≥ 50% |
| Chatbot | 200 questions with gold answers | Groundedness | 100% |
| Chatbot | 50 unanswerable questions | Correct refusal | ≥ 98% |
| Share text | 100 fact sets | Prohibited-pattern violations | 0 |
| Escalation drafting | 50 issue records | Schema validity + citation correctness | 100% |

Eval sets live in the repository (photos anonymised and consented) and are versioned. **A prompt
change without an eval run does not merge.**

---

## 11. Prompt management

- Prompts live in `internal/ai/prompts/*.md`, versioned, never inline string literals
- Each prompt file has a header: version, last-evaluated date, eval-set reference, owner
- Changes are reviewed like code
- The prompt version is recorded on every artefact the call produces, so any output can be traced to
  the exact prompt that produced it

---

## 12. Failure behaviour

| Failure | Response |
|---|---|
| API unavailable | Queue the job; the report is stored and acknowledged regardless; UI shows "classifying" |
| Rate limited (429) | SDK retry with backoff; then queue for batch |
| Structured output validation fails | Retry once with the validation error appended; then human triage queue |
| Confidence below threshold | Ask the user; never guess silently |
| Model returns a refusal | Log with the category, route to human triage, never surface a raw refusal to the citizen |
| Daily budget exhausted | Degrade to queue-and-classify-later with an honest UI message |

**Invariant:** no AI failure loses a report, and no AI failure publishes an unverified fact.

---

## 13. Things deliberately not done with models

| Not doing | Why |
|---|---|
| Text-to-SQL | Injection surface into a database containing personal data |
| Model decides publication | Publication thresholds are rules, auditable and versioned |
| Model asserts wrongdoing | Legal conclusion; out of scope by policy |
| Model files anything | Filing is always an explicit human action |
| Fine-tuning on citizen photos | Consent, cost, and DPDP complexity vastly exceed the marginal gain over a strong general VLM |
| Model-generated moderation decisions | The model ranks the queue; a human decides |
