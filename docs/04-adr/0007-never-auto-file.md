# ADR 0007 — Never auto-file a legal or administrative instrument

**Status:** Accepted · **Date:** 2026-08-10

## Context

The escalation engine can generate an RTI, a first appeal, an NGT original application, a Lokayukta
complaint, or a consumer complaint automatically once its preconditions are met. Automating the
submission as well would be trivially easy and would obviously increase throughput.

It would also be wrong.

## Decision

**The platform never submits a legal or administrative instrument without an explicit, per-instance
human action.**

Concretely:

- There is **no** `POST /escalations/{id}/submit` endpoint.
- Generated instruments are `draft` → the citizen reviews and edits → the citizen files (either
  outside the platform, or by triggering a filing adapter with an explicit tap) → the citizen or the
  adapter records the reference number.
- The same applies to social posts: a share kit is generated; the user posts it.
- The same applies to routine grievance filing with an authority, though the friction there is a
  single tap rather than a review flow, because the consequences of a wrongly-filed municipal
  complaint are small.

Every generated instrument carries: a "this is a draft, not legal advice" disclaimer, the citizen's
name as the applicant (never the platform's), and a review checklist.

## Rationale

1. **Legal responsibility follows the signature.** An RTI or a tribunal application is filed by a
   named person who is accountable for its contents. The platform cannot assume that.
2. **A wrong auto-filed instrument is unrecoverable.** An RTI to the wrong PIO wastes the fee and the
   30-day clock. A defective NGT application can be dismissed and prejudice the cause of action.
3. **Bulk automated filing would be abused** — by us, by a griefer with a script, or by a political
   operative — and would give authorities a legitimate reason to dismiss the entire platform as
   vexatious.
4. **It preserves the citizen's agency**, which is the point of the product.

## Alternatives

| Option | Why not |
|---|---|
| **Auto-file with an opt-out** | Same risks; the opt-out will not be read. |
| **Auto-file after a delay with a cancel window** | Better, but still puts a legal document out under someone's name without them reading it. |
| **Auto-file only for low-stakes instruments (municipal complaints)** | This is effectively what we do — a single tap, no review flow, for grievance filing. The line is drawn at instruments with a statutory clock or a legal forum. |

## Consequences

- Lower escalation throughput than a fully automated system. Accepted deliberately.
- The product must make review *fast* — a good draft, clearly presented, with the specific edits a
  citizen might want highlighted. This is a UX investment, not a compromise.
- The deadline scheduler becomes more important, because the citizen must be prompted in time to act.
- Recorded in `CLAUDE.md` as a hard rule so no contributor, human or AI, adds the endpoint.
