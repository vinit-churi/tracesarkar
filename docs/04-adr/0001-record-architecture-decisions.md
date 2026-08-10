# ADR 0001 — Record architecture decisions

**Status:** Accepted · **Date:** 2026-08-10

## Context

This project spans law, geospatial data, procurement, AI, and infrastructure. Decisions in each area
constrain the others, and the reasoning behind them will be forgotten within weeks. Contributors —
human and AI — need to know not just what was decided but what was rejected and why, or they will
relitigate settled questions.

## Decision

Every significant architectural decision is recorded as a numbered Markdown file in `docs/04-adr/`,
following a light Nygard-style format:

```
# ADR NNNN — Title
**Status:** Proposed | Accepted | Superseded by ADR-XXXX · **Date:** YYYY-MM-DD
## Context      what forces are at play
## Decision     what we are doing
## Alternatives what we rejected and why
## Consequences what this costs us
```

ADRs are immutable once accepted. A changed decision gets a new ADR that supersedes the old one; the
old file stays, with its status updated.

## Alternatives

- **A wiki** — drifts from the code, and has no review process.
- **Comments in code** — invisible to anyone reading the docs, and lost on refactor.
- **Nothing** — the default; produces repeated argument about settled questions.

## Consequences

- Small ongoing discipline cost.
- Contributors can answer "why is it like this?" without asking.
- `CLAUDE.md` points AI agents here, so they inherit the reasoning rather than re-deriving it.
