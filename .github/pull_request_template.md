## What this changes

<!-- One paragraph. What and why. -->

## Related

<!-- Issue, ADR, or doc section this implements. -->

---

## Hard-rule checklist

These are non-negotiable — see [CLAUDE.md](../CLAUDE.md) and
[CONTRIBUTING.md](../CONTRIBUTING.md).

- [ ] No path auto-files a legal or administrative instrument
- [ ] No displayed fact about a named party lacks a source reference and retrieval timestamp
- [ ] No user-visible string asserts wrongdoing (facts and joins only)
- [ ] No public surface, log, metric, or trace exposes PII or precise coordinates
- [ ] No legal constant (SLA, fee, limitation period, compensation amount) appears as a literal
- [ ] A captured report cannot be lost by this change
- [ ] No source is scraped whose terms forbid it

## Engineering checklist

- [ ] Tests included, covering the failure path
- [ ] Degrades correctly when its dependency is unavailable
- [ ] Every external call has a timeout, retry policy, and circuit breaker
- [ ] Metrics emitted per [observability](../docs/03-architecture/10-observability.md)
- [ ] Documentation updated **in this commit**
- [ ] ADR written if this is an architectural decision
- [ ] Migration is expand/contract safe against the previous app version
- [ ] SPDX header on new source files
- [ ] Commits signed off (DCO): `git commit -s`

## If this touches AI

- [ ] Structured outputs used (no free-text parsing into domain objects)
- [ ] Prompt lives in a versioned file, not an inline literal
- [ ] Eval set run; results in the PR description
- [ ] Prompt/model version recorded on produced artefacts
- [ ] Cost impact considered

## If this touches published facts about named parties

- [ ] Publication threshold enforced in code, not by convention
- [ ] Provenance chain visible to users ("how was this matched?")
- [ ] Dispute path reachable from the record
- [ ] Reviewed against
      [legal review checklist §A](../docs/06-operations/03-legal-review-checklist.md#a-publishing-anything-about-a-named-party)

## AI-assisted?

- [ ] Yes — reviewed and understood before submitting
- [ ] No
