# Project state

**Last updated: 24 September 2026.** Read this first when you pick the project up, on any device.
It is a snapshot, not a spec: every line links to the document that holds the detail.

---

## 1. The goal

A resident photographs a civic problem. TraceSarkar tells them which authority owns that spot,
which public contract covers it, and whether it is still under warranty. It then drafts the next
step: a complaint, an RTI, an appeal, a compensation claim. The resident files it themselves;
the platform never does.

It starts in one ward (BMC R/S, Kandivali West) with one kind of problem (road defects). The bet
behind it is that **showing the contract changes what people do**. Every phase is gated on evidence
that the bet is holding. If it isn't, the plan says to stop and re-plan around escalation instead.

More: [vision](01-vision.md) · [v0.1 MVP](../05-delivery/02-milestone-v0-mvp.md).

---

## 2. Where things stand

| | |
|---|---|
| Code | **Phase 0 collectors are built and running.** Go module, eight packages, ~100 tests |
| Plan | Phases 0–6 — [roadmap](../05-delivery/01-roadmap.md) |
| Current phase | **Phase 0, in progress** — [Phase 0 spec](../05-delivery/07-phase-0-instruments.md) |
| Data held | 4,673 work records, 4,673 change rows, 9 archived documents, 5 Government Resolutions |
| Infrastructure | Cloudflare R2, an Aiven PostgreSQL database, and a nightly collector VM in `asia-south1` — all live |
| Branch | `main` |

**The 30-day snapshot clock starts with the first scheduled run, 02:30 IST on 25 September 2026.**
Collection is verified end to end from both runners: the Mumbai VM covers everything including the
roads API, and GitHub Actions covers the drain API and Government Resolutions from anywhere.

---

## 3. What happened on 23 September 2026

Phase 0's foundations and its first instrument were built, test-first, and run against the live
BMC APIs.

**Built:** configuration, the R2 archive client (SigV4 checked against the AWS test vector), the
migration runner and Phase 0 schema, the source register with its tier policy, the works parsers
and differ, the polite fetcher, the snapshot runner, the alert channel, and the `ingest` command.

**Running:** `ingest run` collects BMC's roads dashboard, road geometry, ward master and the
storm-water progress card. `ingest watch` archives newly published Government Resolutions from the
Internet Archive mirror and flags the ones whose OCR text mentions a watched term — the RTI fee,
defect liability, potholes, Right to Public Services. Bytes are archived to R2 only when they
change; every attempt is logged; every change is stored with both values.

Verified end to end: archived objects read back from R2 byte-for-byte, and the Marathi keyword
match confirmed against the 2019 defect-liability resolution.

**Two bugs the live run found**, both now fixed and covered by tests:

1. A document that was archived but failed to parse counted as "already seen", so the next run
   would have skipped it and the data would have been lost quietly. Bytes now count as seen only
   after a snapshot is applied.
2. The road layer's key was wrong. `(workCode, locationName)` is not unique — 2,405 features share
   1,919 pairs — `locationID` is null in 1,678 of them, and the nested `location._id` is shared
   between features and was overwriting each feature's own id. The key-collision check caught it;
   without that check 486 records would have merged silently.

**Deviation worth knowing:** the database adapter was written before its tests, against the
project's test-first rule. Its tests were added immediately after and found two real defects (a
null blocklist column and a missing `endpoint` column). Everything else was written test-first.

---

## 4. What happened on 18 September 2026

1. **Checked the Screen Book artifact.** It draws 56 screens covering 113 of 129 catalogued
   features. Still pending: screens S13 and S24 are not drawn; 21 drawn screens have no entry in the
   [screen spec](../02-product/11-screen-spec.md). Both are backlog tasks.
2. **Found the docs out of step with the August data audit.** Only 1 of its 10 follow-ups had
   landed. The roadmap still planned to crawl Mahatenders, which the project had already ruled out.
   The ones that matter now are fixed; the rest are backlog tasks.
3. **Explored new verticals**, with three research passes and first-hand checks. The result is
   [vertical exploration](../01-research/09-vertical-exploration.md). Headlines:
   - BMC also publishes an open API for storm-water drain desilting — drains can follow roads.
   - MahaRERA publishes 52,529 construction projects with coordinates in one download.
   - Pune publishes more contract-to-location data than BMC does.
   - A competitor exists: Pothole Reporter, an Android app launched in August 2026.
   - Several public BMC map layers expose personal data, including patient-level health records.
4. **Corrected the August audit.** The warranty-period (DLP) rule is a PWD resolution of
   14 Jan 2019, not 27 Apr 2017. A cited "BMC" page was Bhubaneswar's. Flooding history is public.
   A 2023 rate schedule exists.
5. **Caught a live change.** BMC's roads data listed 58 deleted works on 24 August and 60 on
   18 September. Which two changed is unknowable, because nothing was archiving. That is why Phase 0
   exists.
6. **Re-planned** the roadmap into phases and wrote the Phase 0 spec.

---

## 4A. What happened on 24 September 2026

The repository secrets went in and the workflow ran for real — and failed in a way worth the
failure. From a GitHub runner in the US, `roads.mcgm.gov.in` and `portal.mcgm.gov.in` refuse
connections outright, on every port, while `swd.mcgm.gov.in` and the Internet Archive answer
normally. All four respond from your home connection in Mumbai.

**BMC geo-restricts most of its estate.** The drain API collected fine from CI and even recorded
five changed values, so the pipeline is sound; the roads API simply cannot be reached from abroad.

We decided how to handle it ([ADR 0015](../04-adr/0015-indian-egress-for-collection.md)): a small
VM in `asia-south1` that an instance schedule starts at 02:30 IST, which collects and then powers
itself off — about ₹45 a month, alive three minutes a night. On-demand rather than spot, because at
this duty cycle spot saves about ₹2 a month and risks a night that cannot start for want of
capacity, and a missed night cannot be recovered. GitHub Actions keeps the sources that answer from
anywhere, as redundancy.

**Answered the same day.** The VM was built and run: from `34.100.176.104` in Mumbai, every BMC
host answers, `roads.mcgm.gov.in:3000` in 18 ms. BMC filters by geography, not by network type, so
a datacenter in India is enough.

The collector now runs unattended there: it boots, reads its credentials from Secret Manager,
downloads the binary and the source register from the `collector-latest` release, collects, and
powers itself off. Four bugs were found by running it rather than reading it: `/run` is mounted
`noexec` so the binary could not execute from the tmpfs holding the secrets; the VM has no
repository checkout so the source register had to ship with the binary; the GR watcher collected
without consulting the register at all; and a run that failed during setup left no trace anywhere.

Every run is now recorded in `collector_runs` — opened as soon as the database is reachable, so
setup failures are captured too — and `make status` prints the recent ones. That fix immediately
caught the fourth bug: the VM's `watch` was failing silently because it was not given the register.

**Billing note:** a stopped VM costs nothing for CPU or memory, and its ephemeral IP is released.
The only standing charge is the 10 GB boot disk, about ₹42/month. Deleting and recreating the VM
nightly would save that, but instance schedules cannot create machines, so it would mean an
instance template plus a scheduler plus a function — three parts for ₹42. Cloud Run has no disk at
all and is now more plausible than when it was rejected, since BMC accepts Google's Mumbai
addresses; the open question is only whether a job's egress presents as Mumbai.

---

## 5. Decisions to confirm

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

## 6. The plan

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

## 7. Next actions

**Only you can do these:**

- [x] ~~Add the repository secrets~~ — done 24 Sep 2026; collection runs daily at 02:30 IST
- [x] ~~Build the Indian collector VM~~ — done 24 Sep 2026 in project `cloud-mcp-501616`,
      `tracesarkar-collector` in `asia-south1-a`, starting nightly at 02:30 IST
- [ ] **Narrow what CI holds:** an R2 token scoped to the `tracesarkar` bucket, and a database user
      limited to our tables rather than `avnadmin`. The repository is public, so a leak should cost
      as little as possible. Then rotate both. The same credentials now also sit in GCP Secret
      Manager
- [ ] Optional: set `NOTIFY_WEBHOOK_URL` (any endpoint that accepts a JSON POST) so changes reach
      you instead of only the workflow logs
- [ ] Read §5 and confirm or reverse each decision
- [ ] Decide whether to tell BMC that its health-department map layer exposes patient records
- [ ] Send written terms requests to MCGM, CPCB, MahaRERA and IITM. Nothing collected can reach a
      public surface until those answers land
- [ ] Decide whether you want to talk to the Pothole Reporter maintainer (Q35)
- [ ] Optional: pick an alert channel (any endpoint that accepts a JSON POST) and set
      `NOTIFY_WEBHOOK_URL`, so changes reach you instead of only the logs

**The next working session:**

- [ ] P2 watchers: new Maharashtra GRs from the Internet Archive mirror, and the Bombay HC
      judgments parquet
- [ ] Runner hardening: an advisory lock per ingester, a circuit breaker, and a metrics endpoint
- [ ] The SWD nallah-level endpoints, which are POST, plus the next-season path probe
- [ ] Then P3 (capture extension), P4 (field kit), P5 (RTI tracker)

Full task list: [backlog](../05-delivery/03-backlog.md).

---

## 8. Open questions that matter now

| # | Question | Why now |
|---|---|---|
| Q2 | What is the Maharashtra RTI fee? | Blocks RTI generation; your first RTI settles it |
| Q27, Q32 | Terms for BMC's APIs, ArcGIS layers and MahaRERA | Blocks anything from them going public |
| Q28 | Orders in the pothole PIL after Nov 2025 | Every pothole deadline and compensation figure rests on that order |
| Q34 | Second geography: Pune or an MMR corporation? | Decided at the end of Phase 3 |
| Q35 | Partner with Pothole Reporter, or compete? | Decided by the end of Phase 1 |

All of them: [open questions](../01-research/07-open-questions.md).

---

## 9. Where to look

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

## 10. Resuming on another device

```sh
git clone git@github.com:vinit-churi/tracesarkar.git   # or: git pull
cd tracesarkar
cp .env.example .env        # fill in from your password manager
# copy ca.pem across too; both files are gitignored and never leave your machine
make test                   # offline tests
make status                 # what has been collected so far
```

Everyday commands:

```sh
make snapshot   # collect BMC's works data
make watch      # archive new Government Resolutions
make status     # what has been collected, per endpoint
make changes    # what changed in the published data
make test       # offline tests
```

Then start a session with: *"Read `docs/00-overview/06-project-state.md` and continue from §7."*

**The two files you must carry across yourself:** `.env` and `ca.pem`. They hold the R2 keys and
the database password, so they are not in the repository and never will be.
