# Project state

**Last updated: 14 October 2026.** Read this first when you pick the project up, on any device.
It is a snapshot, not a spec: every line links to the document that holds the detail.

---

## 1. The goal

A resident photographs a civic problem. TraceSarkar tells them which authority owns that spot,
which public contract covers it, and whether it is still under warranty. It then drafts the next
step: a complaint, an RTI, an appeal, a compensation claim. The resident files it themselves;
the platform never does.

It starts in one ward (BMC R/C, Borivali) with one kind of problem (road defects). The bet
behind it is that **showing the contract changes what people do**. Every phase is gated on evidence
that the bet is holding. If it isn't, the plan says to stop and re-plan around escalation instead.

More: [vision](01-vision.md) · [v0.1 MVP](../05-delivery/02-milestone-v0-mvp.md).

---

## 2. What we have, as of 16 October 2026

Thirteen things are live and answering. Nothing below is aspirational.

| | Where | State |
|---|---|---|
| **API** | `https://tracesarkar-api-mlkwom573a-el.a.run.app` | Cloud Run, `asia-south1`, healthy · revision `00035-66c`, deployed 16 Oct |
| **App** | `https://tracesarkar-app.infoyantra.workers.dev` | Cloudflare Workers — sign in, capture, send |
| **Android** | `…workers.dev/download/` | Signed APK, published from R2 on a version tag. v0.1.1 queues captures offline |
| **Review** | `…run.app/review/` | Attribution verdicts, signed-in only |
| **Label** | `…run.app/label/` | Ground truth for the eval set, two keystrokes a capture |
| **Methodology** | `…run.app/methodology/` | Public, no sign-in — sources, measurements, limits |
| **Ward** | `…run.app/ward/` | Public, no sign-in — who owns what in R/C, how to reach them, what binds them |
| **Issue** | `…run.app/issue/{id}` | Public, no sign-in — one problem: what, roughly where, how many sightings, whether it was lodged. A bare link unfurls into the facts and the card |
| **Share kit** | on `…run.app/issue/{id}` | Public — the post, written for you in English, Marathi or Hindi, editable, WhatsApp first. The card comes in feed and 9:16, and a signed-in reporter can compile the record as a PDF |
| **Mark up** | `…run.app/truth/` | Drag a box round every face and plate, so the detector can be graded. Signed-in only |
| **Field kit** | `…run.app/` | Capture with ground truth recorded on the spot; signs in |
| **Collector** | Cloud Run job, `asia-south1` | Twice daily, alerting on failure |
| **Sweep** | Cloud Run job, `asia-south1` | Every 10 min — enrich, redact, classify, issues. A capture resolves itself |

### In the database

| | |
|---|---|
| Road works archived | 4,878 |
| Observed changes | 12,062 |
| Raw documents archived | 596 · including the 36 handbooks, charters and circulars the routing tables rest on ([D162](05-decision-log.md)) |
| Road segments in PostGIS | 2,404 · **778 km** of carriageway |
| — of those, Borivali | 198 |
| Ward boundaries | 24 · all of Greater Mumbai |
| Department mappings | 4 · road defects, waste, water supply — all sourced |
| BMC departments registered | 86 · every department BMC lists, with its handbook ([D150](05-decision-log.md)) |
| Wards with an escalation contact | 24 of 24 · Assistant Commissioner and zone DMC, role mailboxes ([D149](05-decision-log.md)) |
| Routable categories | **6 of 6** · `environment` only for tree felling, which is the only part of it BMC owns ([D158](05-decision-log.md)) |
| Legal constants | 12 · two pothole clocks, a penalty, four waste clocks, and five from the Tree Authority — the first that come from a statute ([D154](05-decision-log.md), [D160](05-decision-log.md)) |
| Government resolutions watched | 461 |
| Captures stored | 5 · all real. Six synthetic rows were deleted on 13 Oct ([D129](05-decision-log.md)) |
| — redacted, with a verdict recorded | 5 of 5 |
| Issues formed | 2 |
| Captures labelled for the eval set | 5 |
| Wards with a desk to complain to | 24 of 24, six categories · 337 rows, from 4 ([D145](05-decision-log.md), [D147](05-decision-log.md), [D152](05-decision-log.md), [D156](05-decision-log.md), [D158](05-decision-log.md), [D163](05-decision-log.md), [D165](05-decision-log.md), [D167](05-decision-log.md)) |
| — and an office above it to escalate to | 24 of 24 · 31 rows across 7 zones, every one with both a mailbox and an address ([D149](05-decision-log.md), [D159](05-decision-log.md)) |
| Captures marked up for faces and plates | 0 · the set exists, the photographs do not |
| Attribution verdicts by a human | 50 |

### The numbers that were measured, not assumed

| Measurement | Result |
|---|---|
| **Attribution precision** | 50 of 50 judged right by a human. Honest floor **≥ 92.9%** (95% Wilson). **Zero** wrong-contractor matches, ever |
| **Jurisdiction accuracy** | **98.17%** over 2,404 road works. Bar was ≥ 95% |
| **Contract coverage, Borivali** | **~43–57%** of carriageway. Best where it matters: secondary 77%, tertiary 69% |
| **Questions the citizen would face** | 0.7% of points |
| **Classification** | 5 labelled cases on 12 Oct — category 5/5, subcategory 4/5, no hazard missed |
| **Redaction** | Verdict on 5 of 5, 0.5–0.9 s each on the re-scan, 4 false positives. **Recall still unmeasured**: none of the five contains a face or a plate ([D127](05-decision-log.md)). The harness that will measure it exists — `/truth/` and `make redact-eval` — and found a real defect on its first run ([D134](05-decision-log.md)) |
| **Public image budget** | 2.5 MB → 135 KB and 1.7 MB → 127 KB on the two live issues, against a 150 KB budget |

### Tests

22 Go packages passing, 22 Flutter tests. Live tests gated behind `TRACESARKAR_LIVE=1`, and the
whole live store suite is green as of 16 October. Three of them had rotted and were repaired that
day: two asserted that most wards have nothing recorded, which stopped being true at migration
0029, and one could never have passed — it expected a capture to be classifiable before anything
had redacted it, and `UnclassifiedReports` joins on the redacted derivative. The CA-path gotcha
in [§8](#8-picking-this-up-on-another-machine) applies when running them.

---

## 3. The eight systems

v1 is eight systems, built strictly one at a time ([D063](05-decision-log.md)). Scope authority is
**[v0.1 MVP §5](../05-delivery/02-milestone-v0-mvp.md)**.

| # | System | State |
|---|---|---|
| **1** | **Attribution** — point → road → contract → contractor | **Bar met on probes.** Open until real captures confirm it |
| **2** | **Jurisdiction** — ward, department, confidence gate | **Bar met.** 98.17%. Road defects, waste and water now route to a named desk in **all 24 wards** — 73 rows, up from 4 — read out of BMC's own PIO list ([D145](05-decision-log.md), [D147](05-decision-log.md)) |
| 3 | **Classification** — what is in the photograph | **Built, measurement running.** Six categories, versioned prompt, guardrails. `ingest eval` scores whatever is labelled and re-runs as the set grows; 5 cases on 12 Oct — category 5/5, subcategory 4/5, no hazards missed. The ≥90% bar accrues from use ([D114](05-decision-log.md)) |
| **4** | **Issues** — dedup, corroboration, SLA clock, hash-chained timeline | **Built, surfaced, and now public.** A second sighting joins the first and is shown as corroboration; the clock starts only when the reporter records the complaint reference; history is hash-chained and verifies, and is readable at `/issue/{id}` without the actor |
| 5 | Accounts — OTP, rate limits, reset, revocation | **Built but for phone OTP.** Email/password, Google, sessions, forgery resistance and throttling ([D113](05-decision-log.md)), and now password reset, email verification and token revocation ([D171](05-decision-log.md), migration 0044). Revocation is a `token_version` on the account rather than a list of dead tokens: bumping it invalidates every token at once, which is what a reset and "sign out everywhere" both need. Mail is plain text through any SMTP server, or written to the log at the personal tier; with `MAIL_MODE` unset the two mail endpoints answer 501 rather than minting a credential nobody will receive. **Missing: phone OTP**, which [ADR 0017](../04-adr/0017-email-and-google-identity-before-public-tier.md) keeps as the gate on the public tier and which needs a DLT-registered sender ID and a paid SMS route |
| **6** | **Public surface** — redaction, coarsening, permalink, ward page | **Mostly built.** Coarsening, the ward page, and both detectors — faces on an embedded Pico cascade, plates on an edge-density localiser, neither leaving the process ([D124](05-decision-log.md)). Scan, verdict and output object are recorded on `report_media`, and the database refuses a published verdict with no object. The permalink is live at `/issue/{id}`: coarsened position, hash-chained history with no actor, and a photograph only behind a completed redaction ([D130](05-decision-log.md)). Images are fitted to the 150 KB budget on the way out — 2.5 MB became 135 KB on a live issue ([D131](05-decision-log.md)). The detector is now gradable: `/truth/` records what a person can see in a photograph and `make redact-eval` scores both detectors against it ([D132](05-decision-log.md), [D133](05-decision-log.md)). Remaining, and still the gate: photographs with faces and plates in them. Five captures measured no recall because none contains either ([D127](05-decision-log.md)) |
| **7** | **Share kit** — annotated image, per-language text | **Mostly built.** The post is assembled from the fact set with no model in the path, in three languages and two channels, and the annotated image puts the facts under the photograph rather than across it ([ADR 0020](../04-adr/0020-share-text-is-generated-not-written-by-a-model.md), [D135](05-decision-log.md)). A handle is printed only with its source and a re-verification date ([D136](05-decision-log.md)). **All three languages are native-speaker reviewed** as of 16 October 2026 ([D137](05-decision-log.md), [D175](05-decision-log.md)); the review changed 19 of 43 Marathi strings and 20 of 43 Hindi ones, and all 36 subcategories now have a word in each language — before it, 27 fell back to the category word and a felled tree composed as "Environment in H/W ward". A link pasted in any chat unfurls into the facts and the annotated card, which meant moving the redaction gate ahead of the first byte rather than relaxing it ([D139](05-decision-log.md)), and the card is drawn in 9:16 for Status and Stories ([D140](05-decision-log.md)). The evidence pack compiles the record as a deterministic, stamped PDF — the one artefact here that is signed in, because it carries the exact position and the reference a complaint needs ([D141](05-decision-log.md), [D142](05-decision-log.md)). Not built: the WhatsApp intake bot |
| 8 | Operator console — moderation, collection health, metrics | **Started.** `/ops/` answers the three questions that were being read out of logs ([D178](05-decision-log.md)): the claimed-versus-confirmed counts, what is waiting with *what clears it* on every row, and sources and runs with a never-returned run distinguished from a failed one. Deliberately **no moderation** — `/review/`, `/label/` and `/truth/` are the surfaces that take a verdict, and a test asserts the console issues no write. Not built: metrics over time, and anything multi-operator |

**The first real captures landed on 5 October 2026, and earned their keep.** Two photographs of a
waste problem — one in R/C, one in R/N — went through the whole chain, and reading what came back
found three defects no probe could have shown: a road contractor named against a garbage report, a
subcategory vocabulary nothing was checking, and coverage that ignored the ward
([D101](05-decision-log.md)–[D104](05-decision-log.md)). All three are fixed. The lesson is the one
the limit below already stated, now with evidence: probes test the join, photographs test the
product.

**Photograph quality is now a measured constraint.** Both evening captures were rated `poor` by the
classifier and scored 0.50 and 0.62 against a confidence floor of 0.65, so both correctly asked
rather than guessed. Daylight, close, filling the frame.

**System 1 was investigated on 1 October 2026 and holds.** A live test reading "matched the wrong
work" turned out to be asserting on a tie-break, not a mis-attribution — see
[D090](05-decision-log.md). The investigation did measure something worth knowing: 269 pairs of
segments from *different* contract packages lie within a metre of each other, so a defect there
could belong to either contract. The gate answers those honestly rather than guessing.

**Systems 1 and 2 share a limit worth stating.** Both were validated against BMC's own data and a
human reading the same map the code read. Neither has been tested against a photograph taken at a
place a person can identify. That is what real captures close, and it is the one thing fieldwork is
genuinely needed for.

---

## 4. What we know about the pothole deadline

Three numbers exist and the platform must say which is which
([D079](05-decision-log.md), [detail](../01-research/11-pothole-sla-sources.md)):

| Deadline | Binds | Source |
|---|---|---|
| **48 hours** | the corporation | Bombay HC, `2025:BHC-OS:18736-DB` para 70(ix) |
| **24 hours** | the contractor, in DLP | BMC tender ETH_8000040832 §10.11 |
| ₹5,000/day/pothole | the contractor | same tender §7.4 |

The June 2026 "24 hours from the Commissioner" is **not a document** — a verbal instruction at a
review meeting. Not a constant.

Three traps, each recorded because each would otherwise be walked into: the court order spells it
**"forty-eight hours"** in words; road ownership is **seasonal, not width-based**; and BMC's own
contract accepts **WhatsApp as valid intimation**, which is the same channel the share kit targets.

---

## 5. History, compressed

Older detail lives in git. The turns that changed the plan:

| When | What changed |
|---|---|
| 18 Sep | Verticals explored; legal risk became exposure tiers, not a gate ([D044](05-decision-log.md)) |
| 23–24 Sep | Collection moved to a Cloud Run job in Mumbai — BMC geo-restricts by geography, verified from three vantage points ([ADR 0015](../04-adr/0015-indian-egress-for-collection.md), [0016](../04-adr/0016-collection-as-a-cloud-run-job.md)) |
| 25 Sep | Capture endpoint, field kit, auth, and the Flutter client. Deployed to a shared Dokploy host |
| 28 Sep | v1 restated as eight systems, one at a time ([D063](05-decision-log.md)) |
| 29 Sep | Dokploy host died; API moved to Cloud Run ([ADR 0019](../04-adr/0019-api-on-cloud-run.md)). Ward switched to Borivali on measured data ([D068](05-decision-log.md)). Systems 1 and 2 hit their bars |

Two defects found by checking rather than trusting, both worth remembering:

- **2,387 personal mobile numbers** were in the database. The blocklist named fields BMC's API does
  not return, so it stripped nothing. Fixed, scrubbed, and a test now asserts the register names
  fields the source actually returns.
- **Every capture was filed under one shared account**, even when a signed-in person sent it. The
  middleware verified the session; the handler ignored it.

---

## 6. Decisions to confirm

These were approved quickly during the session. They are recorded in the [decision log](05-decision-log.md) and the docs now depend on them. **Read the right-hand column; if
any is wrong, say so and it gets reversed with a new log row.**

| # | Decision | What it means in practice |
|---|---|---|
| D044 | Exposure tiers instead of legal-risk filtering | Any feature can be built. It starts as `personal` (only you), `flagged` (invite-only) or `public`. The hard rules in `CLAUDE.md` bite when something goes public |
| D045 | Personal-tier ingesters may run before terms are reviewed | The snapshotter can start on BMC's APIs now, without waiting for MCGM to answer a terms request. Nothing it collects goes public until the terms are reviewed. Mahatenders stays off entirely |
| D046 | One VPS + Cloudflare R2 | You need a small VPS running Dokploy, and an R2 bucket. Archived government documents are locked forever; report photos are not, so they can be deleted on request. The Phase 0 archive is the real, permanent one from day one |
| D047 | Personal tools before the public app | Nothing public ships until Phase 1 (target Feb 2027). Drains must be live by April 2027, before the monsoon |

Also chosen, not logged as decisions: the verticals order (drains first, then construction sites,
building safety and hoardings behind a flag, then trees), and starting with personal tools.

### Two model choices that are not decisions yet, and should be

`CLAUDE.md` says the default model is `claude-opus-5`, and that a cheaper tier is permitted on the
batch path *where the eval set shows no quality loss* — explicitly "a documented decision, not an
implementation detail". Neither of the two models in the codebase has that document.

- **The classifier runs `deepseek-flash`** (`CLASSIFY_BASE_URL=https://api.deepseek.com/v1`), chosen
  because it was the vision model available without a card on file and because DeepSeek's own
  listing does not say which of its two models can see images — that much is recorded in
  `.env.example`. What is **not** recorded anywhere is a decision naming it, or a comparison against
  the default on the eval set. [D114](05-decision-log.md) has the measurement (5 cases, category
  5/5, subcategory 4/5) and does not name a model as the choice.
- **The rewording model is unchosen**, deliberately: `SHARE_POLISH` without `POLISH_MODEL` is a
  startup error precisely so this does not happen twice ([D170](05-decision-log.md)).

What this needs is one decision-log row per model, naming it, saying what it was compared against,
and saying what would change the answer. Until then the classifier's model is a configuration value
nobody decided on, which is the state `CLAUDE.md` says it must not be in.

---

## 7. The plan

| Phase | What | Target |
|---|---|---|
| **0 · Instruments** | Daily archive of BMC's works data, watchers for new government resolutions and court judgments, a browser extension for saving documents by hand, a field-capture app, an RTI tracker | Nov 2026 |
| 1 · v0.1 | The public report loop in one ward, with contract attribution from BMC's own roads data | Feb 2027 |
| 2 · v0.2 | Drains, "BMC says complete vs your photo", public change log, works near me | Apr 2027 |
| 3 · v0.3 | RTIs, deadline wallet, WhatsApp; construction sites, building safety and hoardings behind a flag | Jun 2027 |
| 4 · v0.4 | Compensation claims, escalation ladder, trees, evidence vault | late 2027 |
| 5 · v0.5–0.6 | A second city or corporation, public API, contractor records | 2028 |
| 6 · v0.7+ | Depth | — |

Phase 0 exits when: 30 days of unbroken snapshots · 500 labelled photos and 200 known-ward points ·
20 contract documents saved · one RTI filed by you.

---

## 8. Picking this up on another machine

Everything the project needs is in the repository except the secrets, which are deliberately not.

### What you have to bring

Neither of these is in git, and nothing runs without them:

| File | What it is | Where to get it |
|---|---|---|
| `.env` | Database URL, R2 keys, `API_TOKEN`, `AUTH_SECRET`, `CLASSIFY_*`, `GOOGLE_CLIENT_ID` | Copy it across from the other machine, or rebuild it from `.env.example` and the Cloud Run `tracesarkar-env` secret |
| `ca.pem` | Aiven's certificate for the Postgres connection | Download again from the Aiven console; `POSTGRES_CA_PATH` points at it |

Also not in git and not reproducible: `app/android/tracesarkar-release.jks`. **Losing it means no
further update can be installed over an existing APK** — back it up somewhere that is not a laptop.

### Then

```bash
make test          # 24 Go packages, offline, no credentials needed
make migrate       # applies anything newer than the deployed schema (0046 is current)
make sweep         # enrich, redact, classify, issues — what the Cloud Run job runs
make api           # serves the API, the field kit, and every public page on :8080
make test-live     # the tests that touch the real bucket and database
```

`make test` passing with no `.env` present is the quickest check that a clone is sound.

**One gotcha with the live tests.** `POSTGRES_CA_PATH` in `.env.example` is `./ca.pem`, a *relative*
path, and `go test ./internal/store/` runs with that package as its working directory — so every
live test fails with `postgres CA file: stat ./ca.pem: no such file or directory` even though the
certificate is sitting in the repository root. Point it at an absolute path when running them:

```bash
POSTGRES_CA_PATH=$PWD/ca.pem TRACESARKAR_LIVE=1 go test ./internal/store/
```

### Where the code actually is

| Want to change | Look in |
|---|---|
| Point → road → contract | `internal/attribute/` |
| Ward, department, the confidence gate | `internal/jurisdiction/` |
| What is in a photograph | `internal/classify/` — prompts are versioned files, not literals |
| Dedup, the clock, the hash chain | `internal/issue/` |
| Faces, plates, the mosaic, the hold | `internal/redact/` |
| Grading the detector against marked-up photographs | `internal/redact/eval.go`, `cmd/ingest/redact_eval.go` |
| The one place a model may touch a generated post | `internal/polish/` — rewords composed lines, never composes one; `Verify` is the gate, `Corpus`/`Measure` the eval set |
| A reset or verification link, and how it is delivered | `internal/mail/` — plain text, no tracking; `internal/auth/onetime.go` mints the credential, `internal/store/credentials.go` holds the hash |
| What a public surface may show | `internal/publish/` |
| The next step a report earns | `internal/action/` |
| HTTP, and every page | `internal/api/` — static pages are embedded under `internal/api/{ward,issue,label,review,methodology,fieldkit,truth}/ and the share kit's\nvocabulary in `internal/share/i18n/`` |
| SQL | `internal/store/` — hand-written, no ORM. Migrations are forward-only and numbered |
| Who a complaint escalates to | `internal/store/escalation.go` — `EscalationFor` (ward + zone), `Departments` (the BMC register) |

### What is deployed, and how

Three Cloud Run resources in `asia-south1`, all from this repository:

- `tracesarkar-api` — the API and every page, `maxScale=3`. Revision **`00042-cnq`** (`ab8a86f`,
  16 October) lets a reporter answer the question the classifier could not
  ([D179](05-decision-log.md), migration 0046) — the field kit carries the block, the endpoint is
  authenticated and scoped to the caller's own captures, and three captures stuck since 5 October
  are answerable. `00041-725` added `/ops/`, system 8's first surface ([D178](05-decision-log.md)), and takes the
  permalink from `PUBLIC_BASE_URL` rather than from the request's `Host` header
  ([D177](05-decision-log.md)) — now set to the service's own URL, which needs no DNS; change it to
  the custom domain when one exists. Verified after deploy: `/ops/` serves, `/v1/ops/health`
  refuses an unauthenticated request, and a spoofed `Host` no longer reaches a generated post. The
  authenticated `/v1/ops/health` response was **not** checked through the deployed service, because
  `API_TOKEN` is a Secret Manager reference rather than a literal and a smoke test is not a reason
  to read a secret; its data is covered by the live store tests, which do run against this
  database. `00039-vbl` served the native-speaker-reviewed Marathi and Hindi with no draft notice on any post
  and a word for all 36 subcategories ([D175](05-decision-log.md)), and no longer drops the status
  line to fit X's window — which had been making *the authority says attended*, *a resident
  photographed it fixed* and *seen again afterwards* the same post
  ([D176](05-decision-log.md)). Verified on a live issue: `draft_languages` absent, all three
  languages `reviewed`, no vocabulary fallbacks. `00036-x6t` added password reset, email
  verification and session revocation
  ([D171](05-decision-log.md), migration 0044 — the two mail endpoints answer 501 until `MAIL_MODE`
  is set), the share kit's `wording` field and the rewording layer it belongs to
  ([D170](05-decision-log.md), off until `SHARE_POLISH` is set), and one spelling per ward in every
  escalation line ([D172](05-decision-log.md), migration 0045). It carries `/truth/`, the share kit, the unfurling permalink, the evidence pack,
  obligations shown with the scope they bind, desks shown with the scope *they* bind, and a card
  that says **"Not BMC"** when the office belongs to another body, and an issue block stating
  what the law requires for the thing photographed
  ([D156](05-decision-log.md), [D166](05-decision-log.md), [D168](05-decision-log.md))
- `tracesarkar-collector` — twice daily, alerts on failure. Both jobs now run `3c0d386`
- `tracesarkar-sweep` — every ten minutes, now with a fifth stage that removes spent
  password-reset and verification links a week after they stop working ([D171](05-decision-log.md));
  it belongs on a job that already runs rather than a timer of its own, and matches nothing most
  passes. Both jobs share the collector image, rebuilt at
  `b1aa75e`: a newly captured photograph is scanned by `edge-density-2` rather than the localiser
  that was blind to large plates ([D134](05-decision-log.md)), and the next step it drafts quotes
  the court's forty-eight hours only where that order actually binds ([D138](05-decision-log.md))

The Flutter client is a Cloudflare Worker; the APK is published to R2 on a version tag. Deployment
detail is in [the runbook §2](../06-operations/01-runbook.md).

**The archive bucket is private and must stay that way.** It holds photographs taken by people who
did not agree to publish them. It was briefly public and was closed; the public page serves the
*redacted copy* through the API, never a bucket URL, so that the gate sits in front of every read.

---

## 9. Next actions

### Worth doing, and no longer a blocker

- [ ] **Fill the adversarial redaction set.** The machinery is built: `/truth/` to drag a box round
      every face and plate in a photograph, and `make redact-eval` to score the detectors against
      what you recorded ([D132](05-decision-log.md), [D133](05-decision-log.md)). What it has no
      supply of is photographs. Nothing in the five stored captures contains a face or a plate, so
      recall is still unmeasured. **This no longer blocks anything**: the owner has accepted the
      residual risk and asked that it not be treated as a defect ([D144](05-decision-log.md)). The
      detectors still run on every capture and still hold an image they cannot account for; what is
      gone is the rule that said no public image may exist until recall is measured.
      **What the set needs:** faces at distance, in profile, partly occluded, at night, backlit, in
      a crowd; plates at an angle, on two-wheelers, the yellow commercial ones, hand-painted, and
      half-hidden behind another vehicle. Mark them up the same evening — the conditions are what
      explain a miss, and they are hard to remember a week later.
      One annotated photograph was already enough to find the localiser blind to any plate large in
      the frame ([D134](05-decision-log.md)), which is what the set is for.

### Next in the code — the BMC portal thread, picked up where it was left

The portal is **no longer blocked**; the obstacle was our own User-Agent, not BMC
([D148](05-decision-log.md)). Everything below is now reachable without asking anybody for a link.

- [x] **Settled: a street light is BMC's, and the ward's.** Done 16 October
      ([D152](05-decision-log.md), migration 0033). Duty 38 of the ward Assistant Engineer
      (Maintenance & Repair) is to provide street lighting to the licensee's LUX value — Adani and
      Tata Power publish the standard, the ward does the work — and the same item appears word for
      word in A Ward and in P/North inside Adani's licence area. The Mechanical & Electrical
      handbook, read as images, claims high-mast lamps and garden illumination and **not** street
      lighting, so the citywide desk that looked obvious would have been wrong. `street_furniture`
      now routes in all 24 wards, to the same office that holds minor roads and footpaths
- [x] **Rendered the Citizen Charter — by not rendering it.** Done 16 October
      ([D153](05-decision-log.md), [D154](05-decision-log.md), [D155](05-decision-log.md)). The
      page never loads; its inner iView times out on every attempt. The document store behind it
      is SAP Knowledge Management, which answers **WebDAV `PROPFIND`** — so the folder lists
      itself even though its HTML view says it has no displayable content. 50 folders, 141 files,
      one charter per department and ward. BMC's waste clocks are now its own: 12 hours for the
      category, 48 for a dead animal (migration 0034). Roads & Traffic publishes no pothole clock
      and the Hydraulic Engineer none for a leak, both recorded so nobody looks again
- [x] **Seeded the footbridge desk — and taught routing about subcategories.** Done 16 October
      ([D156](05-decision-log.md), [D157](05-decision-log.md), migrations 0035 and 0036).
      `structural` is one category and two authorities: the Chief Engineer (Bridges) holds the
      structural audit and repair of every MCGM bridge above 6 m of span, the ward's Assistant
      Engineer (Building & Factory) is the Designated Officer under MMC Act s.354 for a ruinous or
      dangerous structure. So `authority_departments` gained a nullable `subcategory`, and a
      cracked bridge now routes differently from a leaning building in all 24 wards. Manual 16 of
      the Bridges handbook gave a ward-by-ward Assistant Engineer with a role mailbox for every
      ward. The Storm Water Drains desk for spans of 6 m and under followed on the same day
      ([D164](05-decision-log.md)); a bridge in a private layout belongs to the society that owns
      it and is still not seeded, because nothing in it is a public desk
- [x] **Seeded tree felling, and `environment` with it.** Done 16 October
      ([D158](05-decision-log.md), migration 0037). Not from the Garden charter, which covers
      saplings and plant hire — from the Garden Department's own Section 4 handbook, whose stated
      mission is "To Control illegal tree cutting" and whose Manual 16 names an Assistant
      Superintendent of Gardens for every ward, answering to that ward's Assistant Commissioner.
      Scoped to the subcategory: a mangrove is the Forest Department's and an effluent discharge
      the pollution board's, so the category stays empty and those still answer "not known".
      **All six categories now reach a desk.** The same manual closed migration 0031's two open
      addresses ([D159](05-decision-log.md))
- [x] **Split `water_drainage`, which had been routing to the water supply desk.** Done 16
      October ([D163](05-decision-log.md), migration 0040). Since migration 0019 a photograph of a
      flooded lane in any of the twenty-four wards was addressed to Water Works, the supply
      department. The ward maintenance handbooks give duties 32–34 — minor nallas, road-side SWD,
      flooding spots, missing manholes — and the Storm Water Drains handbook claims the network
      and the major nallas. Waterlogging and blocked drains now go to the ward, a silted nallah to
      the SWD department by division, and burst pipelines and leaks stay where they belonged
- [x] **Seeded the sewage overflow desk.** Done 16 October ([D165](05-decision-log.md),
      migration 0041). `water_drainage` is now fully split: supply to Water Works, the street to
      the ward, the network and major nallas to Storm Water Drains, and the sewers to the Chief
      Engineer (Sewerage Operation). One honest gap recorded with it — the department's English
      directory has no Eastern Suburbs divisional office, so every ward points at the head of
      department rather than at two divisions out of three
- [x] **Taught routing that a desk need not be the corporation's — and seeded the pollution
      control board.** Done 16 October ([D166](05-decision-log.md), [D167](05-decision-log.md),
      migrations 0042 and 0043). `covers_authority` separates who a desk *serves* from who it
      *belongs to*, so an effluent discharge or an air pollution source inside a BMC ward now
      names the Maharashtra Pollution Control Board, and the ward page says **"Not BMC — this is
      MPCB"** on that card. Two wards the Board's own sub-regional lists omit, E and R/Central,
      are given its Regional Office rather than an inferred one
- [ ] **The last three subcategories of `environment`.** Mangrove destruction and encroachment on
      a water body belong to the Forest Department's Mangrove Cell and the coastal zone authority
      (MCZMA); creek dumping is shared between the board and the corporation on a basis no
      document in hand settles. The machinery to record them exists now; the sources do not. All
      three still answer "not known"

- [x] **Read what BMC has notified under the Right to Public Services Act 2015 — and closed the
      thread.** Done 16 October ([D161](05-decision-log.md)). BMC publishes it properly, in
      `MCGM Department List/Public Service Charter` and in per-department notice boards. Four were
      read, and every notified service is a **permission, a certificate or a connection** — a zone
      certificate in 7 days, an IOD in 60, a sewerage connection in 15, tree-felling permission in
      45. Not one is a complaint, so the Act cannot supply the deadlines this platform's categories
      need. The Garden figures are seeded as unbound constants ([D160](05-decision-log.md))
- [x] **Rendered the Tree Authority facts.** Done 16 October ([D168](05-decision-log.md)). An
      issue of a felled tree now carries a block headed "What the law requires for this" — the
      permission, its published fee per tree, and the 45 statutory days to decide an application —
      each with its citation, apart from the clock and saying in its own words that it is not one
      and that it says nothing about whether anybody followed it. The share kit is deliberately
      not given these: a fee beside a photograph of a stump, in a post that travels beyond any
      page this platform can correct, is an accusation however it is worded

- [x] **Marathi and Hindi are reviewed.** Done 16 October ([D175](05-decision-log.md)). A native
      speaker read both files: 19 of 43 Marathi strings changed and 20 of 43 Hindi ones, and the
      verdict was that the Marathi was close to the mark while the Hindi was "Google Translate
      syndrome" — over-relying on `प्राधिकरण`, which nobody in Mumbai says, and on literal
      renderings of English grammar. Four findings generalise to any later vernacular surface and
      are recorded in [the review sheet](../02-product/13-vernacular-review-sheet.md). Punctuation
      is the full stop in both languages, which also means the rewording prompt's instruction to use
      the danda in Hindi was wrong and is corrected. **All 36 subcategories now have a word in all
      three languages** — including English, which was missing 27 of them, so a felled tree
      composed as "Environment in H/W ward" in every language. The three tests that hard-coded
      "Marathi is a draft" now compare the review state before and after instead, so signing off a
      language is a change to one JSON file and to no test

- [x] **Finished system 5 but for phone OTP.** Done 16 October ([D171](05-decision-log.md),
      migration 0044). Password reset, email verification and token revocation, with the design
      decision in the middle: revocation is a `token_version` on the account rather than a list of
      dead tokens, so bumping it kills every session at once and a reset and "sign out everywhere"
      are the same operation. A one-time credential is 256 random bits stored as a SHA-256 hash and
      nothing else; the purpose is part of the lookup, so a verification link cannot reset a
      password. The forgotten-password endpoint answers 202 with an identical body whether or not
      the address is registered, and a successful reset returns no session

- [x] **Put a model in the share path, for rewording only.** Done 16 October
      ([D170](05-decision-log.md),
      [ADR 0021](../04-adr/0021-a-model-may-reword-a-generated-post-never-author-one.md)).
      `internal/polish` hands a model the numbered prose lines of an already-composed post — not
      the fact set, not the permalink, not the issue id — and publishes what comes back only if
      every protected span, every number and every account survives the line it replaced. One
      failure discards the whole candidate and the composed post is served, so a provider that is
      down, slow or wrong costs nothing. The eval set
      [ADR 0020](../04-adr/0020-share-text-is-generated-not-written-by-a-model.md) demanded before
      any of this exists: 16 fact sets, 96 posts, all six categories, three languages, including
      the states that remove a sentence. **Off unless `SHARE_POLISH` is set, with no default
      model** — that choice has to be a recorded decision with the measurement behind it, and is
      not yet made

- [x] **Archived what was read.** Done 16 October ([D162](05-decision-log.md), migration 0039).
      Thirty-six documents — every ward and departmental handbook, charter, circular and Right to
      Public Services board this week's work rests on — fetched, hashed and stored by a new
      command, `ingest archive-docs`, with the manifest kept in
      [`docs/01-research/sources/bmc-portal-manifest.txt`](../01-research/sources/bmc-portal-manifest.txt).
      **169 of 169 routing rows now cite an archive key**, and a live test fails if one does not.
      Another asserts every cited key is really in the bucket — which also confirmed the four keys
      cited since September were there

### Next in the code

- [ ] **Decide whether to turn the rewording on.** It is measured and off
      ([D173](05-decision-log.md)). `deepseek-flash` published **92 of 96** corpus posts with
      **0 invented facts** — no lost span, no changed number, no invented account or link, no tone
      violation — and changed the vernacular in 49 of 64 posts against English in 9 of 32, which is
      the asymmetry [ADR 0021](../04-adr/0021-a-model-may-reword-a-generated-post-never-author-one.md)
      was written about. `deepseek-v4-pro` is ruled out on latency, not fidelity: two minutes a call
      against an 8-second budget. **`SHARE_POLISH` stays unset until you say otherwise** — turning
      it on changes what a public surface publishes, which is not a consequence of a good test
      result. Re-run any candidate with
      `POLISH_BASE_URL=… POLISH_API_KEY=… POLISH_MODEL=… POLISH_SCHEMA_MODE=json_object \
      TRACESARKAR_LIVE=1 go test ./internal/polish/ -run TestLive -v` (about 5 minutes, 96 calls)
- [ ] **Phone OTP, the last of system 5.** Reset, email verification and revocation are done
      ([D171](05-decision-log.md)); phone verification is not, and
      [ADR 0017](../04-adr/0017-email-and-google-identity-before-public-tier.md) keeps it as the
      gate on the public tier. It is a procurement task before it is a coding one: a
      DLT-registered sender ID and a paid SMS route in India
- [ ] **Choose a mail route.** The code is done and off: `MAIL_MODE=smtp` with `SMTP_HOST`,
      `MAIL_FROM` and `PUBLIC_BASE_URL`, or `MAIL_MODE=log` to read your own links out of the log
      at the personal tier. A domain sending password resets also needs an SPF record, or the mail
      is spam ([D171](05-decision-log.md))
- [x] **Answered the question the platform forgot it asked.** Done 16 October
      ([D179](05-decision-log.md), migration 0046). The console's "waiting for the citizen" count
      turned out to be a capture the classifier had asked about with a list of options that
      `SaveClassification` never persisted — computed, discarded, never shown, and skipped by the
      sweep for ever. The options are stored now, the field kit asks, and a reporter's answer writes
      a new classification row marked `source = 'reporter'` rather than editing the model's. Three
      real captures from 5 and 6 October are answerable

- [x] **Activity over time in the console.** Done 16 October ([D180](05-decision-log.md)).
      `/ops/` draws fetches, collector runs, captures and issues per day over 7, 30 or 90 days,
      from timestamps the rows already carried — no new storage, so the history was already there.
      A quiet day is drawn as a hairline rather than omitted, and *never returned* stays separate
      from *failed*. It found two failed collector runs on 24 and 25 September that nothing had
      surfaced. **Queue depth has no history and cannot be given one retrospectively**, because a
      queue is counted by asking what is unfinished now and that answer was never recorded; the page
      says so rather than drawing arrivals as if they were depth
- [ ] **Finish system 8.** What is left is moderation — which is deliberately *not* in the console
      ([D178](05-decision-log.md)), so the question is whether the existing narrow surfaces
      (`/review/`, `/label/`, `/truth/`) are enough or whether v1 wants something more. Roles and
      permissions are explicitly out of v1 scope. If queue depth over time is ever wanted, it needs
      the sweep to start writing a snapshot row; that is a decision, not a backlog

### Yours, in the field

- [ ] **Photograph more than potholes.** Garbage, overflowing bins, illegal dumping, blocked
      drains, waterlogging and burst pipes all route now, and the eval set is far more useful
      across three categories than one. Anywhere, not only Borivali — a label is ground truth
      wherever the photograph was taken ([D115](05-decision-log.md))
- [ ] **Photograph people and vehicles too, deliberately**, for the adversarial set above — then
      open `…run.app/truth/` and drag a box round every face and plate in them. That is the one
      kind of capture the walk is not producing and the one the detector needs
- [ ] **Label as you go.** Open `…run.app/label/` on a laptop and work through the walk's captures:
      a number for the category, a number for what it is, `s` to skip. Skip anything you cannot
      read from the photograph alone — a guess becomes a wrong answer the classifier is marked
      against forever
- [ ] **Lodge one real pothole complaint yourself** — MyBMC MARG or the Pothole QuickFix app — and
      record what reference number comes back and in what form. The whole escalation ladder hangs
      off that reference, and nothing in the product can be designed properly until we have seen
      one ([D085](05-decision-log.md))

### Yours, on paper

- [ ] **Buy the guidelines-and-circulars booklet** at Municipal Head Office. BMC's own tenders say
      it is held at the Dy. Ch. Eng. (Roads)(Planning) offices and may be purchased. No 30-day
      clock, and it may contain the three pothole circulars outright
- [ ] **File [RTI 1](../06-operations/rti-01-pothole-sla-circulars.md)** — on paper, ₹10 court-fee
      stamp, registered post AD or over a counter. BMC is not on the state portal, and its own
      online form serves 2 of 24 wards ([D075](05-decision-log.md), [D077](05-decision-log.md)).
      **Photograph the stamped application and the fee** — that settles Q2
- [ ] **Record the registration number** in the RTI file when you get it. The 30-day clock starts
      on receipt

### Housekeeping that is overdue

- [ ] **Register the Android OAuth client** — package `org.tracesarkar.app`, SHA-1
      `30:A3:36:B2:D5:03:17:66:4A:90:9C:80:22:36:47:4E:7D:CD:BB:D2`. Google sign-in on the APK does
      not work until this exists
- [ ] **Back up `app/android/tracesarkar-release.jks`** somewhere that is not a laptop. It cannot
      be regenerated, and losing it means no update can ever be installed over an existing APK
- [ ] **Narrow what CI holds:** an R2 token scoped to the `tracesarkar` bucket, and a database user
      limited to our tables rather than `avnadmin`. The repository is public. Then rotate both
- [ ] **Revoke the temporary Dokploy API key** — that host is dead and the key is in a transcript
