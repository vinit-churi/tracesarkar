# ADR 0012 — AGPL-3.0 for code, CC BY-SA 4.0 for data and docs

**Status:** Accepted · **Date:** 2026-08-10

## Context

TraceSarkar is civic infrastructure built on public data. Two failure modes are worth guarding
against:

1. A vendor takes the codebase, closes it, and sells it back to the same municipal corporations as a
   proprietary dashboard — which is precisely the market position we decline to occupy (see
   [personas §P4](../02-product/01-personas-and-jtbd.md)).
2. The normalised procurement corpus — arguably the most valuable artefact the project produces —
   gets enclosed by someone downstream.

We also want researchers, journalists, and other civic-tech projects to use everything freely.

## Decision

| Asset | Licence |
|---|---|
| Source code | **AGPL-3.0-or-later** |
| Documentation | **CC BY-SA 4.0** |
| Normalised datasets (contracts, boundaries, aggregates) | **CC BY-SA 4.0** |
| Contributed content (citizen photographs) | Licensed to the platform for platform purposes under the consent notice; **not** relicensed for bulk redistribution |

## Rationale

**AGPL over MIT/Apache:** the network-use clause is the point. A hosted proprietary fork is the
realistic enclosure risk here, and only the AGPL addresses it. The cost — some organisations will not
touch AGPL — is acceptable, because the target adopters are civic-tech projects and public bodies,
not proprietary SaaS vendors.

**CC BY-SA over CC0 for data:** share-alike keeps derived datasets open. Attribution also matters for
a different reason: when a number from this project appears in a story or a filing, the attribution
chain is part of its credibility.

**Citizen photographs are treated separately and deliberately.** They are personal data contributed
for a specific purpose under a DPDP consent notice. Relicensing them for bulk redistribution would
be inconsistent with that consent, regardless of what a code licence says. Public derivatives are
displayed on the platform and shareable through the share kit; they are not offered as a bulk
open-data dump.

## Alternatives

| Option | Why not |
|---|---|
| **MIT / Apache-2.0** | Maximises adoption; permits exactly the proprietary hosted fork we are guarding against. |
| **GPL-3.0** | Does not cover network use, which is how this software will actually be deployed. |
| **Source-available / BSL** | Not open source; would undermine the project's own transparency argument. |
| **CC0 for data** | Simpler for reusers; permits enclosure of derived datasets and loses attribution. |
| **Dual licence with a commercial exception** | Requires a CLA, which raises the contribution barrier for a project that needs volunteer contributors. Reconsider only if a genuine funding path depends on it. |

## Consequences

- Every source file carries an SPDX header: `// SPDX-License-Identifier: AGPL-3.0-or-later`.
- Dependencies must be licence-compatible; a CI check enforces this and blocks incompatible additions
  (notably, the redaction model in [ADR 0008](0008-redaction-approach.md) must be permissively
  licensed).
- Contributors are asked to affirm the DCO rather than sign a CLA — lower friction, and sufficient
  given we are not dual-licensing.
- Organisations with AGPL policies may decline to contribute. Accepted.
- If a municipal corporation wants to self-host, they can — and must publish their modifications.
  That is a feature.
