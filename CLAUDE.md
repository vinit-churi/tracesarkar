# Repository context for AI coding agents

This file is read by Claude Code and compatible agents. `AGENTS.md` is a copy for other tools.

---

## What this project is

TraceSarkar is a civic accountability platform for the Mumbai Metropolitan Region. A citizen
photographs a civic problem; the platform resolves jurisdiction, matches the location to a public
procurement contract, and generates the correct administrative or legal escalation.

Read [`README.md`](README.md) first, then
[`docs/00-overview/01-vision.md`](docs/00-overview/01-vision.md) and
[`docs/03-architecture/01-system-overview.md`](docs/03-architecture/01-system-overview.md).

**Current state: documentation and design. There is no application code yet.** Do not assume code
exists; check before referencing it.

**Start every session with
[`docs/00-overview/06-project-state.md`](docs/00-overview/06-project-state.md)** — the current
phase, decisions awaiting confirmation, and next actions.

---

## Hard rules

These are not style preferences. Violating any of them creates legal or safety exposure.

1. **Never auto-file anything.** RTIs, complaints, tribunal applications, and social posts are
   *generated as drafts*. A human taps to submit. There is no code path that submits a legal
   instrument without an explicit, logged user action.
2. **Never publish an unsourced fact about a named party.** Every contractor name, award value,
   blacklisting entry, or official designation displayed must carry a source reference and a
   retrieval timestamp. If the source is missing, the field is not rendered.
3. **Never assert wrongdoing.** The platform states facts and joins ("this contract covers this
   location; its DLP is active"). It never concludes ("this contractor is corrupt"). This applies to
   code comments, generated text, prompts, and UI copy.
4. **Never store or display precise reporter location publicly.** Public surfaces get coarsened
   coordinates. Exact coordinates stay internal.
5. **Never store raw faces or number plates in a public derivative.** Redaction happens before the
   public derivative is written.
6. **Never hard-code a legal constant.** SLAs, fees, limitation periods, and compensation amounts
   live in a versioned config table with an effective-from date and a citation. The 48-hour pothole
   SLA is not a literal in Go code.
7. **Never lose a captured report.** Raw capture is persisted before any enrichment is attempted.
   Every downstream stage must be retryable.
8. **Never scrape a source whose terms forbid it.** Record the decision in the source register and
   use RTI instead.

---

## Conventions

### Language and stack

- **Backend:** Go. Standard library first; add dependencies deliberately.
- **Database:** PostgreSQL 16 + PostGIS. PostGIS is authoritative for containment; H3 is an index,
  never an authority.
- **Migrations:** forward-only, numbered, reversible where practical.
- **Object storage:** S3-compatible.
- **Orchestration:** Docker Swarm via Dokploy.

### Code

- Package layout: `cmd/` for binaries, `internal/` for everything not intended for external import.
- Errors: wrap with `fmt.Errorf("...: %w", err)`. No bare `err` returns from a layer boundary.
- Context: every function that does I/O takes `ctx context.Context` first.
- Logging: structured (`log/slog`), with `issue_id`, `report_id`, and `request_id` where applicable.
  Never log PII, exact coordinates, or media URLs with tokens.
- Time: `time.Time` in UTC internally; format for display at the edge only.
- Money: `int64` paise. Never `float64`.
- IDs: UUIDv7.
- SQL: written by hand in the repository layer. No ORM.
- Tests: table-driven. Geospatial logic gets golden-file tests with real MMR coordinates.

### Naming

Use the canonical identifiers from [`docs/00-overview/03-glossary.md`](docs/00-overview/03-glossary.md).
`report` and `issue` are different things and must not be used interchangeably.

### Documentation

- Every architectural decision gets an ADR in `docs/04-adr/`.
- Every external data source gets an entry in the source register.
- Docs are Markdown, wrapped at 100 columns, with relative links between them.

---

## Working on the AI pipeline

See [`docs/03-architecture/04-ai-pipeline.md`](docs/03-architecture/04-ai-pipeline.md) for the
authoritative specification. Key points for agents:

- **Use structured outputs** (`output_config.format`) for every classification and extraction call.
  Never parse free text into a domain object.
- **Default model is `claude-opus-5`**; the batch/backfill path may use a cheaper tier where the
  eval set shows no quality loss. Do not silently downgrade a model to save cost — that is a
  documented decision, not an implementation detail.
- **Prompt caching:** system prompts, tool definitions, and taxonomy blocks are stable and must be
  placed before any volatile content so the prefix caches.
- **Every model output that reaches a user must be validated** against the platform's own records
  before display. See [`docs/02-product/06-chatbot.md`](docs/02-product/06-chatbot.md) §2.
- **Prompts live in versioned files**, not inline string literals, so they can be diffed and evaluated.

---

## Working on the UI

[`docs/02-product/11-screen-spec.md`](docs/02-product/11-screen-spec.md) is the authority for every
client surface — global rules, design tokens, component vocabulary, and 34 screens. Where a mockup
and that file disagree, the file wins. Its §1 restates the hard rules above as UI rules; a screen
that breaks one is wrong however good it looks. Generating screen imagery goes through
[`12-ui-generation-guide.md`](docs/02-product/12-ui-generation-guide.md).

Three that are violated most often: share is the primary action on an issue, not a header icon;
`claimed_resolved` and `citizen_confirmed` never share a label or a count; and no button anywhere
implies the platform files something.

---

## Before you rely on a data source

[`docs/01-research/08-data-availability-audit.md`](docs/01-research/08-data-availability-audit.md)
is a field survey from 24 August 2026 with a provenance mark on every row — ✅ means it was fetched
first-hand, ○ means it was not. Retrieved artefacts live in
[`docs/01-research/sources/`](docs/01-research/sources/).

Two things it establishes that change how you should reason:

- **BMC publishes an open JSON API** at `roads.mcgm.gov.in:3000/api/` carrying contractor, work
  code, dates and geometry for 2,237 road works. Much of the contract-to-geometry join is already
  done for the CC-road programme. Its `dlpPeriod` field is null throughout, and it returns personal
  mobile numbers that must be stripped at ingestion.
- **Mumbai has elected corporators again** (BMC polled 15 January 2026). Any doc that assumes an
  administrator is stale, and the `COUNCILLOR` field in BMC's GIS predates the election.

---

## Things that look like good ideas and are not

| Idea | Why not |
|---|---|
| Text-to-SQL for the chatbot | Injection surface into a database containing personal data. Use typed tools. |
| An ORM | The queries here are spatial and hand-tuned. An ORM will fight you. |
| Storing money as float | Rounding errors in contract values that get published. |
| Using H3 for jurisdiction containment | H3 cells do not respect administrative boundaries. Use PostGIS. |
| Auto-translating generated English text to Marathi | Produces garbage for hashtags and legal terms. Generate from the fact set per language. |
| Skipping the raw-artefact archive to save storage | The archive, not the parsed row, is the evidentiary record. |
| Adding a "resolved" button for authorities | Only a citizen photograph moves an issue to `citizen_confirmed`. |
| A points leaderboard by report count | Rewards volume, which is the failure mode of every predecessor platform. |

---

## Before you commit

- [ ] Does any new user-visible string assert wrongdoing? Remove it.
- [ ] Does any new displayed fact about a named party lack a source reference? Fix it.
- [ ] Does any new legal constant appear as a literal? Move it to config.
- [ ] Does any new public endpoint expose precise coordinates or PII? Coarsen it.
- [ ] Does any new external call have a timeout, a retry policy, and a circuit breaker?
- [ ] Is there a test for the failure path, not just the happy path?
- [ ] If you added a data source, is it in the source register with a terms review?

---

## Repository map

See the table in [`README.md`](README.md#repository-map).
