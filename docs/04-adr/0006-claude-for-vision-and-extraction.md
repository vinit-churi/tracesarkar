# ADR 0006 — Claude models for classification and extraction, with structured outputs

**Status:** Accepted · **Date:** 2026-08-10

## Context

Two AI-shaped problems dominate:

1. **Classify a civic photograph** into a category, subcategory, severity, and hazard flag, with
   supporting attributes (dimensions, surface, water present, readable signage).
2. **Extract legal terms from a scanned contract PDF** — principally the defect liability period,
   with a verbatim quote and page number.

Neither has a usable labelled dataset available to us. Both need to work on day one.

## Decision

Use the Claude API with **structured outputs** (`output_config.format` with a JSON schema) for every
classification and extraction call.

| Call | Model | Effort |
|---|---|---|
| Interactive classification | `claude-opus-5` | `low` |
| Bulk/backfill classification | `claude-opus-5` via Batch API (50% cost) | `low` |
| Contract PDF extraction | `claude-opus-5` | `high` |
| Share text, news entities | `claude-haiku-4-5` | `low` |
| Chatbot | `claude-opus-5` with typed tools | `low`–`high` by intent |
| Escalation drafting | `claude-opus-5` | `high` |

Supporting decisions:

- **Structured outputs everywhere.** Never parse free text into a domain object.
- **Prompt caching** on the stable prefix (system prompt, taxonomy, tool definitions).
- **Prompts in versioned files**, never inline literals; the prompt version is recorded on every
  artefact produced.
- **Evaluation sets in CI.** A prompt or model change without an eval run does not merge.
- **Vision at native resolution up to 2576 px long edge**, with 1080p as the cost/accuracy default
  pending eval-set validation.

## Alternatives

| Option | Why not |
|---|---|
| **Train a YOLO-family pothole detector** | Needs a labelled MMR dataset we do not have. BBMP's AI road survey shows this works at fleet scale with a procurement budget; it is not the right first move for a citizen-sourced platform. Revisit once we have a large corpus of our own labelled photos. |
| **A smaller hosted VLM** | Cheaper, but classification accuracy directly determines routing correctness, and routing errors are the fastest way to lose citizen trust. |
| **Self-hosted open-weight VLM** | Attractive for cost and data residency at scale; the GPU operational burden is wrong for a one-person team at v0.1. Explicitly revisit at v1 — recorded as a future ADR. |
| **Free-text prompting with regex parsing** | Silent, undetectable failure mode. Rejected. |
| **A commercial OCR service for contracts** | Handles the OCR but not the semantic extraction (which clause is the DLP, what does "from the date of the completion certificate" imply). A VLM does both in one pass with page provenance. |

## Consequences

- **Inference cost is the dominant variable cost.** Mitigated by the Batch API, prompt caching,
  dedup-before-classify, deterministic short-circuits, and a hard daily budget cap with graceful
  degradation.
- **A third-party dependency in a critical path.** Mitigated: reports are stored durably before any
  model call, and every enrichment stage is a retryable job. An API outage delays enrichment; it
  never loses a report.
- **Data sent to a third party.** Mitigated: images are redacted before any external call where
  people are detected, precise coordinates are never sent (a coarse locality string is), and the
  DPDP consent notice discloses the processing.
- **Model behaviour changes over time.** Mitigated by pinned model IDs, versioned prompts, and eval
  sets that run in CI.
