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

## 2. What we have, as of 15 October 2026

Thirteen things are live and answering. Nothing below is aspirational.

| | Where | State |
|---|---|---|
| **API** | `https://tracesarkar-api-mlkwom573a-el.a.run.app` | Cloud Run, `asia-south1`, healthy |
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
| Raw documents archived | 540 |
| Road segments in PostGIS | 2,404 · **778 km** of carriageway |
| — of those, Borivali | 198 |
| Ward boundaries | 24 · all of Greater Mumbai |
| Department mappings | 4 · road defects, waste, water supply — all sourced |
| BMC departments registered | 86 · every department BMC lists, with its handbook ([D150](05-decision-log.md)) |
| Wards with an escalation contact | 24 of 24 · Assistant Commissioner and zone DMC, role mailboxes ([D149](05-decision-log.md)) |
| Routable categories | 3 of 6 · `road_defect`, `waste`, `water_drainage` |
| Legal constants | 5 · two pothole clocks, a penalty, two waste clocks |
| Government resolutions watched | 461 |
| Captures stored | 5 · all real. Six synthetic rows were deleted on 13 Oct ([D129](05-decision-log.md)) |
| — redacted, with a verdict recorded | 5 of 5 |
| Issues formed | 2 |
| Captures labelled for the eval set | 5 |
| Wards with a desk to complain to | 24 of 24, for road defects, waste and water · 73 rows, from 4 ([D145](05-decision-log.md), [D147](05-decision-log.md)) |
| — and an office above it to escalate to | 24 of 24 · 31 rows across 7 zones ([D149](05-decision-log.md)) |
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

22 Go packages passing, 22 Flutter tests. Live tests gated behind `TRACESARKAR_LIVE=1`.

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
| 5 | Accounts — OTP, rate limits, reset, revocation | **Partly built.** Email/password, Google, sessions, forgery resistance and throttling ([D113](05-decision-log.md)) work. Missing: reset, email verification, revocation, phone OTP |
| **6** | **Public surface** — redaction, coarsening, permalink, ward page | **Mostly built.** Coarsening, the ward page, and both detectors — faces on an embedded Pico cascade, plates on an edge-density localiser, neither leaving the process ([D124](05-decision-log.md)). Scan, verdict and output object are recorded on `report_media`, and the database refuses a published verdict with no object. The permalink is live at `/issue/{id}`: coarsened position, hash-chained history with no actor, and a photograph only behind a completed redaction ([D130](05-decision-log.md)). Images are fitted to the 150 KB budget on the way out — 2.5 MB became 135 KB on a live issue ([D131](05-decision-log.md)). The detector is now gradable: `/truth/` records what a person can see in a photograph and `make redact-eval` scores both detectors against it ([D132](05-decision-log.md), [D133](05-decision-log.md)). Remaining, and still the gate: photographs with faces and plates in them. Five captures measured no recall because none contains either ([D127](05-decision-log.md)) |
| **7** | **Share kit** — annotated image, per-language text | **Mostly built.** The post is assembled from the fact set with no model in the path, in three languages and two channels, and the annotated image puts the facts under the photograph rather than across it ([ADR 0020](../04-adr/0020-share-text-is-generated-not-written-by-a-model.md), [D135](05-decision-log.md)). A handle is printed only with its source and a re-verification date ([D136](05-decision-log.md)). Marathi and Hindi ship as **drafts** until a native speaker reads them, labelled as such in the API and on the page ([D137](05-decision-log.md)). A link pasted in any chat unfurls into the facts and the annotated card, which meant moving the redaction gate ahead of the first byte rather than relaxing it ([D139](05-decision-log.md)), and the card is drawn in 9:16 for Status and Stories ([D140](05-decision-log.md)). The evidence pack compiles the record as a deterministic, stamped PDF — the one artefact here that is signed in, because it carries the exact position and the reference a complaint needs ([D141](05-decision-log.md), [D142](05-decision-log.md)). Not built: the WhatsApp intake bot |
| 8 | Operator console — moderation, collection health, metrics | Waiting |

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
make migrate       # applies anything newer than the deployed schema (0032 is current)
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
| What a public surface may show | `internal/publish/` |
| The next step a report earns | `internal/action/` |
| HTTP, and every page | `internal/api/` — static pages are embedded under `internal/api/{ward,issue,label,review,methodology,fieldkit,truth}/ and the share kit's\nvocabulary in `internal/share/i18n/`` |
| SQL | `internal/store/` — hand-written, no ORM. Migrations are forward-only and numbered |
| Who a complaint escalates to | `internal/store/escalation.go` — `EscalationFor` (ward + zone), `Departments` (the BMC register) |

### What is deployed, and how

Three Cloud Run resources in `asia-south1`, all from this repository:

- `tracesarkar-api` — the API and every page, `maxScale=3`. Revision `00032-84w` (`b5c9551`)
  carries `/truth/`, the share kit, the unfurling permalink, the evidence pack, and obligations
  shown with the scope they bind
- `tracesarkar-collector` — twice daily, alerts on failure
- `tracesarkar-sweep` — every ten minutes. Both jobs share the collector image, rebuilt at
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

- [ ] **Settle whether a street light is BMC's or Adani's.** This is the one thing standing between
      the registry and a street-light desk. Adani Electricity lists a dedicated Street Light officer
      in CONNECT 2026, so a citywide Mechanical & Electrical desk would misroute much of the
      suburbs ([D150](05-decision-log.md)). The answer is inside the M&E RTI manuals,
      `qlrticemne` → `Chief Eng M&E_RTI_M_01…18.pdf`, which are **legacy-font Marathi and must be
      read as images** — `pdftotext` returns mojibake, not merely dropped matras. Manual 1 is
      organisation and functions; manual 3 is powers and duties. Until it is settled, **do not seed
      a street-light desk**
- [ ] **Render the Citizen Charter** (`qlcitizenchar`). Its inner iView timed out on every attempt;
      retrying usually works, and a browser screenshot is a fallback. This is BMC's own published
      service-level table and is the proper source for `legal_constants` — the platform currently
      has five constants, four of them from a court rather than from BMC
- [ ] **Seed the footbridge desk from the Bridges handbooks.** `qlCeb` →
      `Chief Engineer_Bridges_E01…E18.pdf`, clean English, already downloadable. Note the file
      numbering is offset: `E(n)` is Manual `(n−1)`, so decision process and time limits are
      **E04**, norms **E05**, officer directory **E10**. Of the three blocked categories this is
      the one with no jurisdictional ambiguity
- [ ] **Seed tree felling.** The Assistant Superintendent of Gardens sits in every ward's CONNECT
      entry, and Gardens & Tree Authority is in the register (`qlgarden`). About an hour's work and
      nothing is blocking it
- [ ] **Read the remaining 23 wards' RTI handbooks** for office hours, visiting hours and the
      per-activity time limits. The path generalises:
      `.../Wards/Assistant Commissioner (<X>-Ward)/RTI Manuals/`. The A-ward handbook gave the real
      escalation chain and a circular register ([D151](05-decision-log.md)); the other 23 will give
      the same, and `authority_departments.office_hours` and `visiting_hours` are null for all but
      two rows
- [ ] **Consider seeding the 227 councillors** — `councillor<N>@mcgm.gov.in` by electoral ward,
      from the English Diary 2026. This supersedes BMC's own GIS `COUNCILLOR` field, stale since
      the January 2026 poll. Weigh it first: a councillor is an elected politician rather than an
      office, so it is closer to a named party under hard rule 2 than the role mailboxes in
      [D149](05-decision-log.md) are, and the share kit's rule against political reference
      ([D136](05-decision-log.md)) bears on how it could ever be surfaced
- [ ] **Archive what was read into R2.** CONNECT 2026, the English Diary 2026 and the handbooks
      downloaded this session are on a laptop, not in the archive, and the archive — not the parsed
      row — is the evidentiary record. `authority_departments` rows cite an `archive_key`; the new
      `ward_escalation_contacts` and `bmc_departments` rows cite a live URL only, which is weaker

### Next in the code

- [ ] **Have a native speaker read the share kit's Marathi and Hindi** —
      `internal/share/i18n/share.mr.json` and `share.hi.json`, one file each, and change
      `"review": "pending"` to `"reviewed"` when they have. Both are live and labelled as drafts
      on every issue page, so this is the only outstanding item a user can see
      ([D137](05-decision-log.md))
- [ ] **Finish system 5** — password reset, email verification, token revocation. Throttling is
      done ([D113](05-decision-log.md)); the rest is not, and the public tier needs it
      ([ADR 0017](../04-adr/0017-email-and-google-identity-before-public-tier.md))
- [ ] **System 8, the operator console** — collection health is currently read from logs

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
