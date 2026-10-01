# Project state

**Last updated: 29 September 2026.** Read this first when you pick the project up, on any device.
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

## 2. What we have, as of 1 October 2026

Three things are live and answering. Nothing below is aspirational.

| | Where | State |
|---|---|---|
| **API** | `https://tracesarkar-api-mlkwom573a-el.a.run.app` | Cloud Run, `asia-south1`, healthy |
| **App** | `https://tracesarkar-app.infoyantra.workers.dev` | Cloudflare Workers — sign in, capture, send |
| **Android** | `…workers.dev/download/` | Signed APK, published from R2 on a version tag |
| **Review** | `…run.app/review/` | Attribution verdicts, signed-in only |
| **Collector** | Cloud Run job, `asia-south1` | Twice daily, alerting on failure |

### In the database

| | |
|---|---|
| Road works archived | 4,683 |
| Observed changes | 4,762 |
| Raw documents archived | 274 |
| Road segments in PostGIS | 2,405 · **779 km** of carriageway |
| — of those, Borivali | 198 |
| Ward boundaries | 24 · all of Greater Mumbai |
| Department mappings | 2 · both sourced |
| Government resolutions watched | 248 |
| Captures stored | 6 |
| Attribution verdicts by a human | 50 |

### The numbers that were measured, not assumed

| Measurement | Result |
|---|---|
| **Attribution precision** | 50 of 50 judged right by a human. Honest floor **≥ 92.9%** (95% Wilson). **Zero** wrong-contractor matches, ever |
| **Jurisdiction accuracy** | **98.17%** over 2,405 road works. Bar was ≥ 95% |
| **Contract coverage, Borivali** | **~43–57%** of carriageway. Best where it matters: secondary 77%, tertiary 69% |
| **Questions the citizen would face** | 0.7% of points |

### Tests

14 Go packages passing, 22 Flutter tests. Live tests gated behind `TRACESARKAR_LIVE=1`.

---

## 3. The eight systems

v1 is eight systems, built strictly one at a time ([D063](05-decision-log.md)). Scope authority is
**[v0.1 MVP §5](../05-delivery/02-milestone-v0-mvp.md)**.

| # | System | State |
|---|---|---|
| **1** | **Attribution** — point → road → contract → contractor | **Bar met on probes.** Open until real captures confirm it |
| **2** | **Jurisdiction** — ward, department, confidence gate | **Bar met.** 98.17% |
| 3 | **Classification** — what is in the photograph | **Next** |
| 4 | Issues — dedup, corroboration, SLA clock, timeline | Waiting |
| 5 | Accounts — OTP, rate limits, reset, revocation | Partly built |
| 6 | Public surface — redaction, coarsening, permalink, ward page | Waiting |
| 7 | Share kit — annotated image, per-language text | Waiting |
| 8 | Operator console — moderation, collection health, metrics | Waiting |

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

## 8. Next actions

### Yours, next week

- [ ] **Buy the guidelines-and-circulars booklet** at Municipal Head Office. BMC's own tenders say
      it is held at the Dy. Ch. Eng. (Roads)(Planning) offices and may be purchased. No 30-day
      clock, and it may contain the three pothole circulars outright
- [ ] **File [RTI 1](../06-operations/rti-01-pothole-sla-circulars.md)** — on paper, ₹10 court-fee
      stamp, registered post AD or over a counter. BMC is not on the state portal, and its own
      online form serves 2 of 24 wards ([D075](05-decision-log.md), [D077](05-decision-log.md)).
      **Photograph the stamped application and the fee** — that settles Q2
- [ ] **Record the registration number** in the RTI file when you get it. The 30-day clock starts
      on receipt

### Yours, whenever

- [ ] **Narrow what CI holds:** an R2 token scoped to the `tracesarkar` bucket, and a database user
      limited to our tables rather than `avnadmin`. The repository is public. Then rotate both
- [ ] **Revoke the temporary Dokploy API key** — that host is dead and the key is in a transcript
- [ ] **Register an Android OAuth client** in Google Cloud so the Google button works on the phone
      as well as the web: package `org.tracesarkar.app`, SHA-1
      `30:A3:36:B2:D5:03:17:66:4A:90:9C:80:22:36:47:4E:7D:CD:BB:D2`
      ([how and why](../06-operations/02-releasing-the-app.md))
- [ ] **Add the GitHub secrets and variables** the Android workflow needs, listed in the same file.
      Until they exist, a version tag fails the build rather than publishing an unsigned APK
- [ ] **Back up `app/android/tracesarkar-release.jks`** somewhere off this machine. It is not in
      the repository and cannot be regenerated; losing it strands every install
- [ ] Read §6 and confirm or reverse each decision
- [ ] Decide whether to tell BMC that its health-department map layer exposes patient records
- [ ] Send written terms requests to MCGM, CPCB, MahaRERA and IITM. Nothing collected reaches a
      public surface until those land
- [ ] Optional: set `NOTIFY_WEBHOOK_URL` so change alerts reach you rather than the logs

### Mine, in order

- [ ] **Schedule the enrichment pass** so a verdict appears without anyone running
      `ingest classify` by hand. The loop closes on the phone today only because that is run
      manually
- [ ] **System 4 — issues and the SLA clock.** Then: finishing accounts, the public surface, the
      share kit, the operator console

Carried, not forgotten:

- [ ] **Rate-limit sign-in** — currently the cheapest attack on the platform
- [ ] **Phone verification before anything publishes** — the gate
      [ADR 0017](../04-adr/0017-email-and-google-identity-before-public-tier.md) promises
- [ ] **A real domain** before the public tier. `run.app` and `workers.dev` are fine for now and
      wrong for a civic platform asking to be trusted
- [ ] Runner hardening: advisory lock per ingester, circuit breaker, metrics

Full task list: [backlog](../05-delivery/03-backlog.md).

---

## 9. Open questions that matter now

| # | Question | Why now |
|---|---|---|
| Q2 | What is the Maharashtra RTI fee? | Blocks RTI generation; your first RTI settles it |
| Q27, Q32 | Terms for BMC's APIs, ArcGIS layers and MahaRERA | Blocks anything from them going public |
| Q28 | Orders in the pothole PIL after Nov 2025 | Every pothole deadline and compensation figure rests on that order |
| Q34 | Second geography: Pune or an MMR corporation? | Decided at the end of Phase 3 |
| Q35 | Partner with Pothole Reporter, or compete? | Decided by the end of Phase 1 |

All of them: [open questions](../01-research/07-open-questions.md).

---

## 10. Where to look

| For | Read |
|---|---|
| The rules that are never broken | [`CLAUDE.md`](../../CLAUDE.md) |
| Why the project exists | [Vision](01-vision.md) |
| The whole plan | [Roadmap](../05-delivery/01-roadmap.md) |
| What to build next | [Phase 0](../05-delivery/07-phase-0-instruments.md), then the [backlog](../05-delivery/03-backlog.md) |
| What public data exists | [August audit](../01-research/08-data-availability-audit.md), [September exploration](../01-research/09-vertical-exploration.md) |
| Every feature ever considered | [Feature catalog](../02-product/02-feature-catalog.md) and series [K](../02-product/13-data-unlocked-features.md), [U](../02-product/14-civic-utility-features.md) |
| What was decided and why | [Decision log](05-decision-log.md) |
| Every screen | [Screen spec](../02-product/11-screen-spec.md) |

---

## 11. Resuming on another device

```sh
git clone git@github.com:vinit-churi/tracesarkar.git   # or: git pull
cd tracesarkar
cp .env.example .env        # fill in from your password manager
# copy ca.pem across too; both files are gitignored and never leave your machine
make test                   # offline tests
make status                 # what has been collected so far
```

**The two files you must carry across yourself:** `.env` and `ca.pem`. They hold the R2 keys and
the database password, so they are not in the repository and never will be.

### Everyday commands

```sh
make status     # what has been collected, per endpoint, and recent runs
make changes    # what changed in BMC's published data
make snapshot   # collect now, rather than waiting for the schedule
make test       # offline tests
make test-live  # tests that touch the real bucket and database

go run ./cmd/ingest wards                      # reload ward boundaries
go run ./cmd/ingest roads --ward R/C           # project road geometry into PostGIS
go run ./cmd/ingest review --count 50          # queue attribution answers to judge
```

Most of these need the environment exported first, because the API and the tools read secrets from
the environment rather than the file:

```sh
set -a && . ./.env && set +a
```

### Deploying

```sh
# API -> Cloud Run
IMAGE="asia-south1-docker.pkg.dev/cloud-mcp-501616/cloud-run-source-deploy/tracesarkar-api:$(git rev-parse --short HEAD)"
gcloud builds submit --config deploy/cloudrun/api.cloudbuild.yaml --substitutions "_IMAGE=$IMAGE" --region asia-south1
gcloud run deploy tracesarkar-api --image "$IMAGE" --region asia-south1

# App -> Cloudflare Workers. API_BASE and GOOGLE_CLIENT_ID are read from .env,
# so neither is retyped from memory — a wrong API_BASE compiles a client that
# talks to nothing and says nothing.
make app-deploy

# Android -> a signed APK on the download page
git tag v0.1.1 && git push origin v0.1.1
```

Deploys are explicit, not push-to-main ([ADR 0019](../04-adr/0019-api-on-cloud-run.md)). The
Android build is the exception: it fires on a version tag, because the thing that makes it
reproducible — the signing key — lives in CI, not here. See
[releasing the app](../06-operations/02-releasing-the-app.md).

### Capturing on your phone

Open **`https://tracesarkar-app.infoyantra.workers.dev`**, sign in, allow location, photograph. No
install, no token. The field kit at the API's `/` still works and queues offline, but the app is
the easier path now.

Then start a session with: *"Read `docs/00-overview/06-project-state.md` and continue from §8."*
