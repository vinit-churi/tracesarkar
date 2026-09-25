# Project state

**Last updated: 25 September 2026.** Read this first when you pick the project up, on any device.
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
| Code | **Collectors running; backend and client both work.** Go module, ten packages, ~130 tests, plus a Flutter client with 17 |
| Plan | Phases 0–6 — [roadmap](../05-delivery/01-roadmap.md) |
| Current phase | **Phase 0 collecting; v0.1 has a signed-in client that captures** — three parallel tracks ([D055](05-decision-log.md)) |
| Data held | 4,673 work records, 4,673 change rows, 9 archived documents, 5 Government Resolutions |
| Infrastructure | Cloudflare R2, an Aiven PostgreSQL database, a Cloud Run job in `asia-south1` collecting twice daily, and **the API live on Dokploy over HTTPS** |
| Branch | `main` |

**The 30-day snapshot clock is running.** Collection happens twice a day at 02:30 and 14:30 IST as
a Cloud Run job in Mumbai, with GitHub Actions covering the globally-reachable sources as
redundancy. Two email alerts watch it: one for a failed run, one for no successful run in 23h30m.
Collection now costs nothing; what remains is the database and the bucket.

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

## 4B. What happened on 25 September 2026

The VM worked, and then made itself redundant. Because it proved BMC accepts Indian *datacenter*
addresses, a Cloud Run job became worth testing — and a job in `asia-south1` reaches every BMC
endpoint too. That is cheaper (no disk, so nothing standing) and simpler (no machine, no startup
script, no serial console) at the same time, which is rare.

So collection moved: a Cloud Run job running `ingest all`, triggered twice a day by Cloud Scheduler.
The VM, its disk and its schedule are deleted. [ADR 0016](../04-adr/0016-collection-as-a-cloud-run-job.md)
records it and supersedes ADR 0015's mechanism, while keeping its finding.

**Alerting now exists**, which it did not before: an email when an execution fails, and an email
when nothing has succeeded for 23h30m. The second is why collection runs twice a day — Cloud
Monitoring cannot watch for a missed run on a daily schedule without crying wolf every day.

**Two drift risks found while reviewing BMC's stability.** The storm-water URL carries the desilting
season, so on 1 January it would have pointed at a path that does not exist yet; the collector now
falls back to the season still being published. The roads geometry URL contains a date BMC chose
(`...startedafter01oct2025roadlayer`), so when they cut a new phase we may keep fetching a stale
layer while everything *looks* healthy. That one has no automatic guard yet — it is in the backlog.

---

## 4C. What happened on 25 September 2026, part two: the backend started

Collection needs nobody now, so the backend began in parallel ([D055](05-decision-log.md)). The
first slice is the one the rest depends on: **a capture is stored before anything is done to it.**

**Built, test-first:**

- `internal/store` — reports, media and labels, on PostGIS. `SaveReport` is idempotent, so a phone
  retrying an upload gets the original report rather than a duplicate
- `internal/api` — `POST /v1/reports` exactly as the API design specifies: multipart, a JSON `meta`
  part, `202 Accepted`, and the photograph in object storage *before* the reply
- `cmd/api` — the binary, with a health check and bearer-token auth
- **The field kit** — a page at `/` that photographs, reads live GPS with its accuracy, takes a
  label and conditions, queues captures on the device, and uploads when there is signal

**Verified end to end** against the real database and bucket: a capture with 6.2 m accuracy stored
at 6.2 m, its image readable back from R2, its label attached. A live test caught genuine precision
loss on the way — `NULLIF($n, 0)` made Postgres read the accuracy as an integer, turning 6.2 into 6,
and accuracy is exactly what decides whether we route confidently or ask a question.

**This unblocks your fieldwork.** Run `make api`, open the page on your phone, and captures start
counting toward the evaluation set.

---

## 4D. What happened overnight, 25 September 2026: something to show

A working client, because the thing that was missing was not another document — it was a screen a
person can sign in to and send a photograph from.

**Auth, test-first, all of it verified against the live database:**

- Email and password (bcrypt, minimum 10 characters, no composition rules) and Google sign-in
  (ID token verified against Google's JWKS — a token whose `alg` is not `RS256`, whose `aud` is not
  our client, or whose `iss` is not Google is refused)
- Sessions as signed tokens; the same endpoint accepts either a session or the field kit's static
  token, so the field kit kept working unchanged
- Sign-in answers **identically** for an unknown address and a wrong password. Telling them apart
  hands an attacker the list of who is registered here

This contradicted [ADR 0011](../04-adr/0011-phone-only-identity.md), which rejects email outright.
Rather than let it drift, [ADR 0017](../04-adr/0017-email-and-google-identity-before-public-tier.md)
records the amendment and its limit: **email and Google are fine while nothing is published; phone
verification still gates the public tier.** Phase 1 does not exit without that check in the
publication path.

**The client — [`app/`](../../app/README.md), one Flutter codebase for Android and the web
([D059](05-decision-log.md)):**

- Sign in or create an account; the session survives a restart, and signing out actually removes the
  token rather than only navigating away
- A capture screen that takes the position first — with the same accuracy bands the field kit uses,
  because a capture from either surface must mean the same thing — then the photograph, then sends
- Colours, type and spacing come from [the screen spec](../02-product/11-screen-spec.md) §2, so it
  is party-neutral by construction rather than by later correction

**Verified end to end against the real Aiven database and the real R2 bucket**, not against mocks:
register → 201 · login → 200 · wrong password → 401 with an identical message · `/v1/auth/me` →
the account · a 631-byte JPEG posted with 6.2 m accuracy → stored at 6.2 m, at
(19.2094, 72.8348), content-addressed in R2 · the same capture retried → the *same* report id and
`created: false`.

**That retry is what found the night's real bug.** The report was idempotent; the media row was
not, so a retried upload left two rows pointing at the same object — enough to inflate media counts
and double-count a photograph in an evaluation export. A failing test first, then migration
`0007_media_idempotency.sql` makes the digest the identity ([D060](05-decision-log.md)).

**Two tests encode hard rules rather than mechanics:** one asserts the capture screen never offers
to file anything, and one asserts signing out removes the token. Those are the assertions that fail
loudly when someone later "improves" the UI.

### To demonstrate it

```sh
set -a && . ./.env && set +a
go run ./cmd/api serve --addr :8080          # terminal one
cd app && flutter run -d chrome \
  --dart-define=API_BASE=http://localhost:8080 \
  --dart-define=DEMO_POSITION=19.2094,72.8348
```

Create an account, take a photograph, send. The report id that comes back is a row in PostGIS and an
object in R2.

`DEMO_POSITION` exists because a laptop often cannot get a GPS fix, and with no position there is
nothing to send. The screen labels it "Fixture position" and says it is worthless as evidence; drop
the flag and the app uses the device. Full walkthrough, failure modes and the questions to expect:
**[demo script](../05-delivery/08-demo-script.md)**.

### What is deliberately not there

Google's *button* is not wired, though the endpoint is built and tested — the web flow needs
`google_sign_in_web`'s rendered button, and that was not worth risking on the night before a demo.
There is no history beyond the current session (`GET /v1/reports/{id}` does not exist), no offline
queue in Flutter (the field kit has one), and no redaction, so no public derivative is written at
all.

---

## 4E. The API is deployed

It is no longer only on a laptop. The HTTP surface runs as a Docker Compose service on the existing
Dokploy instance, built from this repository, behind Traefik with a Let's Encrypt certificate
([D061](05-decision-log.md), [ADR 0018](../04-adr/0018-api-on-dokploy.md)).

```
https://tracesarkar-primarybackend-ls228s-313702-35-188-103-96.sslip.io
```

Collection stays on Cloud Run, and that separation is deliberate: the collector runs twice a day and
exits from Mumbai because BMC geo-restricts by geography; the API has to stay up and does not care
where it runs. Two lifecycles, two deploys.

**Verified against the deployed service, not the laptop:** register → 201 · wrong password → 401
with the same message an unknown address gets · `/v1/auth/me` → the account · a capture posted with
6.2 m accuracy → stored at 6.2 m in PostGIS over `verify-full` TLS, image content-addressed in R2 ·
the same capture retried → same report id, `created: false`, and **one** media row. The container
reports `(healthy)`, which is `api health` answering Docker — the runtime image is `distroless`, so
there is no curl and the binary has to check itself.

**Two things the platform forced, both improvements.** The database CA now travels as
`POSTGRES_CA_PEM` (base64), because Dokploy's only secret channel is environment variables and a
multi-line PEM crosses two parsers on the way in ([D062](05-decision-log.md)). And the compose file
is committed with **no secret in it** — every value is a `${...}` reference — because this
repository is public.

**Production has its own `AUTH_SECRET` and `API_TOKEN`**, generated at deploy rather than copied
from the laptop, so a session minted locally is not valid in production. Both are in Dokploy under
the service's Environment tab.

### What this does not yet mean

The hostname is a generated `sslip.io` name, which encodes the server's IP address — fine for a
private surface, wrong for a civic platform people are asked to trust. It must become a real domain
before the public tier. Pushing to `main` deploys, which is convenient now and needs a gate once
anyone other than you depends on the service being up. And the server is shared with unrelated
projects, so a noisy neighbour is a new way this can degrade.

---

## 4F. The deployed backend, audited

The question was whether the deployment and the auth actually hold up, so they were tested rather
than assumed — against the deployed service, not a laptop.

**What holds.** Valid Let's Encrypt certificate to 24 Dec 2026, HTTP/2, and `http://` redirects to
`https://`. Registration refuses a duplicate address (409), a password under ten characters (400)
and a malformed address (400). Sign-in with a wrong password and sign-in with an address that was
never registered return **byte-identical** responses, so the service cannot be used to discover who
has an account. Every token forgery was refused: a flipped signature character, a payload rewritten
to point at another account id, and an `alg: none` header. CORS echoes only the configured origin
and returns nothing for an unknown one. Posting a capture without a token is refused.

**What did not hold — and this one mattered.** Every report was being filed under the server's own
`field-kit` account, even when a signed-in person posted it with their session token. The
middleware was already verifying the session and putting the claims in the request context; the
capture handler ignored them and used the configured account unconditionally. Four reports in the
database, all attributed to `field-kit`, including ones posted by accounts with an email on them.

It is not cosmetic. **My reports** could never have worked, no report could be traced back to the
person who took the photograph, and the per-account corroboration and trust scoring the platform
depends on would all have been computed against one shared account. Fixed test-first, deployed, and
confirmed in the database: a capture posted with a session token is now attributed to that account,
while the field kit's static token — which names nobody — still resolves to the server's account.

### What is still missing from auth, stated plainly

None of these are bugs; they are things that were never built, and each one is a reason this is not
ready for anyone but the operator:

| Gap | Why it matters |
|---|---|
| **No rate limiting** | Twelve wrong passwords in a row, all answered 401 at full speed. Online password guessing is currently free |
| **No password reset** | A forgotten password means a new account. There is no recovery path at all |
| **No email verification** | Anyone can register any address, including one they do not control |
| **No token revocation** | Signing out clears the token on the device. The token itself stays valid for its full 30 days, so a stolen one cannot be cancelled |
| **No phone verification** | The gate [ADR 0017](../04-adr/0017-email-and-google-identity-before-public-tier.md) promises before the public tier. Needs an SMS route and a DLT sender ID |
| **Google's button is unwired** | The endpoint is built and tested; the web flow needs `google_sign_in_web`'s rendered button and a client ID only you can create |

The static `API_TOKEN` is also worth understanding rather than forgetting: it is a shared secret
that grants capture access without an account, which is what lets the field kit work. Anyone holding
it can post. That is acceptable while the only holder is you, and is a thing to remove before the
surface is public.

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

**The next working session (backend track):**

- [ ] **Rate-limit sign-in** before anyone else has an account — currently the cheapest attack
      against the platform
- [ ] **Revoke the temporary Dokploy API key** used for this deploy, and mint a scoped one if you
      want automated deploys to continue
- [ ] **Point a real domain at the API** before anything public. The current `sslip.io` hostname
      encodes the server's IP address
- [ ] **Wire Google's button** — create a Web OAuth client ID in the GCP console, set
      `GOOGLE_CLIENT_ID` on the backend, and add `google_sign_in_web`'s rendered button. Only you
      can make the client ID
- [ ] **Phone verification before anything publishes** — the gate [ADR 0017](../04-adr/0017-email-and-google-identity-before-public-tier.md)
      promises. Phase 1 does not exit without it
- [ ] Classification: Claude vision with a structured schema, prompt in a versioned file, run over
      whatever the field kit has collected
- [ ] Jurisdiction: load the R/S ward boundary, resolve a point to ward and department, with the
      confidence gate that decides when to ask the one disambiguating question
- [ ] Attribution: spatially join a report to the roads-API geometry we already collect nightly
- [ ] `GET /v1/reports/{id}` so the field kit can show what happened to a capture
- [ ] P3 capture extension, P5 RTI tracker
- [ ] Runner hardening: advisory lock per ingester, circuit breaker, metrics

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
make api        # serve the API and the field kit on :8080 (needs API_TOKEN set)
make status     # what has been collected, per endpoint, and recent runs
make changes    # what changed in BMC's published data
make snapshot   # collect now, rather than waiting for the schedule
make watch      # archive new Government Resolutions now
make test       # offline tests
make test-live  # tests that touch the real bucket and database
```

**Using the field kit on your phone.** It needs HTTPS or localhost for the camera and GPS, so on a
phone use a tunnel (`cloudflared tunnel --url http://localhost:8080`, or any equivalent), open the
URL, paste the `API_TOKEN` once, and capture. Photographs queue on the device and upload when there
is signal, so a walk through a dead spot loses nothing.

Then start a session with: *"Read `docs/00-overview/06-project-state.md` and continue from §7."*

**The two files you must carry across yourself:** `.env` and `ca.pem`. They hold the R2 keys and
the database password, so they are not in the repository and never will be.
