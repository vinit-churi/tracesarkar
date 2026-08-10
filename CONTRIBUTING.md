# Contributing to TraceSarkar

Thank you for considering it. This project touches law, personal data, and named private parties, so
a few things work differently from a typical open-source repository.

---

## Read these first

| If you are… | Read |
|---|---|
| New here | [`README.md`](README.md) → [`docs/00-overview/01-vision.md`](docs/00-overview/01-vision.md) |
| Writing code | [`CLAUDE.md`](CLAUDE.md) — the hard rules apply to humans too |
| Touching anything that publishes a name | [`docs/06-operations/03-legal-review-checklist.md`](docs/06-operations/03-legal-review-checklist.md) |
| Touching anything that collects data | [`docs/03-architecture/11-security-and-privacy.md`](docs/03-architecture/11-security-and-privacy.md) |
| Proposing an architectural change | [`docs/04-adr/`](docs/04-adr/) |

---

## The hard rules

Non-negotiable. A PR that violates any of these will be closed, however good the code is.

1. **Never auto-file a legal or administrative instrument.** Drafts only; a human taps to submit.
   See [ADR 0007](docs/04-adr/0007-never-auto-file.md).
2. **Never publish an unsourced fact about a named party.** Source reference and retrieval timestamp,
   or the field is not rendered.
3. **Never assert wrongdoing.** Facts and joins, never conclusions — in code, comments, prompts, and
   UI copy.
4. **Never expose precise reporter coordinates or PII** on a public surface, in logs, in metrics, or
   in traces.
5. **Never hard-code a legal constant.** SLAs, fees, limitation periods, and compensation amounts live
   in the `legal_constants` table with a citation and an effective-from date.
6. **Never lose a captured report.** Raw capture persists before any enrichment is attempted.
7. **Never scrape a source whose terms forbid it.**

---

## Ways to contribute that are not code

Honestly, these are more valuable right now than code:

| Contribution | Why it matters |
|---|---|
| **Ward boundary data** for MMR corporations outside Greater Mumbai | The biggest coverage blocker. See [jurisdiction map](docs/01-research/03-mmr-jurisdiction-map.md) |
| **Department, SLA, and PIO details** for any authority | Every generated instrument's accuracy depends on this |
| **Filing an RTI** for a road-ownership inventory and sharing the reply | The single biggest correctness dependency |
| **Verifying a research claim** in `docs/01-research/` against a primary source | Everything marked ⚠️ needs this |
| **Labelling photographs** for the classification evaluation set | Directly improves routing accuracy |
| **Marathi / Hindi / Gujarati review** of user-facing copy | Machine translation is not acceptable for legal text |
| **Legal review** of any instrument template | We are engineers, not lawyers |
| **Testing on a low-end device / slow connection** | Our users are on these |

Open an issue describing what you have; we will tell you where it goes.

---

## Code contributions

### Setup

```bash
git clone git@github.com:vinit-churi/tracesarkar.git
cd tracesarkar
make dev          # docker compose up: postgres+postgis, redis, minio
make migrate
make seed         # synthetic MMR data — never real citizen data
make test
```

### Conventions

Full detail in [`CLAUDE.md`](CLAUDE.md). Summary:

- Go; standard library first; dependencies added deliberately
- `cmd/` for binaries, `internal/` for everything else
- Errors wrapped with `%w`; `ctx` first on every I/O function
- Structured logging (`log/slog`); never log PII
- Money as `int64` paise; time as UTC `timestamptz`; IDs as UUIDv7
- Hand-written SQL, no ORM; `sqlc` for generated types
- Table-driven tests; geospatial logic gets golden-file tests with real MMR coordinates
- SPDX header on every source file: `// SPDX-License-Identifier: AGPL-3.0-or-later`

### Pull requests

- One logical change per PR
- Documentation updated **in the same commit** as the code
- Tests for the failure path, not just the happy path
- An ADR for any architectural decision
- The pre-commit checklist in [`CLAUDE.md`](CLAUDE.md#before-you-commit) completed
- Commits: conventional-commit style (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`)

### AI-assisted contributions

Welcome and expected — [`CLAUDE.md`](CLAUDE.md) exists for exactly this. Two requirements:

1. **You are responsible for what you submit.** Review it. If you cannot explain it, do not submit it.
2. **Disclose it** in the PR description. Not because it is a problem, but because it tells reviewers
   where to look hardest.

---

## Reporting bugs

Include: what you did, what happened, what you expected, environment, and — for anything user-facing
— whether real citizen data was involved (if so, **do not paste it**; reference an ID).

For anything with **security or privacy** implications, do not open a public issue. See
[`SECURITY.md`](SECURITY.md).

---

## Proposing features

Check [`docs/02-product/02-feature-catalog.md`](docs/02-product/02-feature-catalog.md) first — it is
long, and the answer may already be there, including for things we have deliberately cut.

A good proposal answers the one-sentence test from
[personas](docs/02-product/01-personas-and-jtbd.md#the-one-sentence-test):

> After this ships, **[persona]** can do **[specific thing]** they demonstrably could not do before,
> in under **[time]**, without **[prior knowledge they don't have]**.

---

## Licensing and the DCO

- Code: **AGPL-3.0-or-later**. Docs and data: **CC BY-SA 4.0**. See
  [ADR 0012](docs/04-adr/0012-licensing.md).
- We use the **Developer Certificate of Origin**, not a CLA. Sign off your commits:

```bash
git commit -s -m "feat: ..."
```

By signing off you certify you have the right to submit the work under the project's licence.

---

## Code of conduct

[`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md). Short version: this project exists to help people who are
routinely ignored by institutions. Behave accordingly.

---

## A note on neutrality

TraceSarkar is not aligned with any political party, and contributions that introduce partisan
framing — in code, copy, data, or issue discussions — will be rejected. The platform's usefulness
depends entirely on being trusted by people who disagree with each other.
