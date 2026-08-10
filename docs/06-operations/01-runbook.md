# Operations runbook

Every procedure is a command, not a paragraph. If a step cannot be written as a command, it is not
ready to be an operational procedure.

**Assumption:** one person on call. Everything here must be executable by that person at 3 a.m.

---

## 1. Alert response

### `report_submission_failing` — PAGE

```bash
# 1. Confirm scope
curl -sf https://api.tracesarkar.org/readyz | jq
tsctl status --service api

# 2. Common causes, in order of likelihood
tsctl logs api --since 15m --level error | head -50   # app errors
tsctl db ping                                          # database
tsctl storage ping                                     # object storage
tsctl swarm ps --filter desired-state=running          # container health

# 3. If object storage is down: submissions must fail, not silently drop.
#    Confirm the client shows the offline-retry message.
tsctl feature set submissions.offline_message true

# 4. If a bad deploy: roll back
tsctl deploy rollback api
```

**Never** "fix" this by accepting reports without storing media. A report we cannot store is a report
we cannot use.

---

### `deadline_scheduler_stalled` — PAGE

The most legally consequential alert in the system: a missed limitation period kills a case.

```bash
tsctl scheduler status
tsctl logs scheduler --since 1h
tsctl scheduler run --job deadline_reminders --dry-run   # inspect what it would send
tsctl scheduler run --job deadline_reminders             # run manually
```

Then: identify any escalation whose reminder window was missed and notify those users individually.

---

### `outbox_lag_high` — PAGE

State changes are not reaching clients or notifications.

```bash
tsctl outbox stats                       # depth, oldest entry
tsctl logs outbox-relay --since 15m
tsctl redis ping
tsctl outbox drain --limit 1000          # manual drain if the relay is wedged
```

---

### `worker_queue_backing_up` — TICKET

```bash
tsctl queue stats                        # depth by job type, oldest age
tsctl workers scale --replicas 12        # scale up
tsctl queue inspect --job enrich_report --failed --limit 20
```

If the cause is AI API rate limiting or an outage:

```bash
tsctl feature set ai.mode batch          # switch classification to the batch path
tsctl feature set ui.enrichment_delay_notice true
```

---

### `ingester_stale` — TICKET

```bash
tsctl ingest status                      # per-source last success, record counts
tsctl ingest run --source mahatenders --dry-run
tsctl ingest logs --source mahatenders --since 24h
```

If the parser is broken by a layout change:

```bash
# Fix the parser, then replay from the archive — no re-scraping needed
tsctl ingest reparse --source mahatenders --since 2026-07-01
```

**Always** surface staleness publicly while it persists:

```bash
tsctl status set --source mahatenders --note "Contract data updates paused; investigating"
```

---

### `ai_budget_80pct` — TICKET

```bash
tsctl ai usage --today --by call_type,model
tsctl ai cache-stats                     # is prompt caching actually working?
```

If the spike is a burst at one location, dedup should have absorbed it — investigate why it did not:

```bash
tsctl issues burst --since 6h --min-reports 20
```

Levers, in order:

```bash
tsctl feature set ai.mode batch
tsctl feature set ai.image_max_edge 1080
tsctl ai budget set --daily <lower>       # last resort; degrades to queue-and-classify-later
```

---

## 2. Deployment

```bash
# Standard
git tag v0.2.3 && git push origin v0.2.3      # CI builds, tests, deploys to staging
tsctl deploy promote --version v0.2.3         # staging → production, after the manual gate

# Migration (runs before the rolling deploy)
tsctl migrate status
tsctl migrate up --dry-run
tsctl migrate up

# Rollback (app only — migrations are forward-only and expand/contract safe)
tsctl deploy rollback api
```

**Migration rule:** every migration must be safe against the previous application version. If it is
not, it is split into two deploys (expand, then contract).

---

## 3. Backup and restore

```bash
# Verify backups are current
tsctl backup status

# Restore drill (quarterly, into a scratch environment — NEVER production)
tsctl backup restore --target scratch --point-in-time '2026-08-01T00:00:00Z'
tsctl verify --env scratch                    # schema, row counts, spatial sanity checks
```

**A backup that has never been restored is a hypothesis.** The quarterly drill is a calendar item.

---

## 4. Monsoon preparation (every April)

```bash
# 1. Load test at 20x baseline against staging
tsctl loadtest --profile monsoon-20x --duration 30m --env staging

# 2. Pre-scale production
tsctl scale api --replicas 8
tsctl scale worker --replicas 16

# 3. Refresh reference data
tsctl ingest run --source bmc_arcgis
tsctl ingest run --source mahatenders --backfill --from 2026-01-01

# 4. Verify offline capture end to end (manual, on a real device)
# 5. Emergency-mode drill
tsctl emergency enable --dry-run
tsctl emergency status

# 6. Restore drill
tsctl backup restore --target scratch --point-in-time now
```

---

## 5. Emergency mode

For a declared civic emergency (severe flooding, a structural collapse, a major incident).

```bash
tsctl emergency enable \
  --banner "Heavy rainfall alert. Report open manholes and flooding immediately." \
  --priority hazard_to_life \
  --divert-to emergency_channel
```

Effects: a site-wide banner; hazard reports jump the queue and route immediately; optional enrichment
is shed; the share kit switches to a hazard-styled variant that names no contractor.

```bash
tsctl emergency disable
```

Modelled on FixMyStreet Pro's emergency diversion. Drilled annually.

---

## 6. Secret rotation

```bash
tsctl secrets list                                  # names and last-rotated dates, never values
tsctl secrets rotate --name anthropic_api_key
tsctl secrets rotate --name db_password --coordinated   # rolls app instances after the change
tsctl secrets rotate --name phone_hash_pepper --plan    # ⚠ requires a re-hash migration; read the plan first
```

Quarterly for API keys; immediately on suspected exposure. The phone-hash pepper is special — rotating
it invalidates every stored hash and needs a planned migration.

---

## 7. Data-subject requests

```bash
tsctl dsr export --account <id> --out ./export.zip     # access request
tsctl dsr correct --account <id> --field handle --value <new>
tsctl dsr erase --account <id> --plan                  # shows what will be deleted and what is retained
tsctl dsr erase --account <id> --confirm
```

Erasure removes personal data and public linkage; anonymised issue aggregates are retained, and
records under legal hold are exempt. The plan output shows exactly which is which.

---

## 8. Incident response

```bash
tsctl incident open --severity P1 --summary "..."      # creates the record, posts to the status page
tsctl incident update <id> --note "..."
tsctl incident close <id> --postmortem ./pm.md
```

Severity: **P1** citizen-facing outage or data exposure · **P2** degraded function · **P3** internal.

For a suspected data exposure, in order:

1. `tsctl incident open --severity P1`
2. Contain: revoke credentials, disable the affected surface
3. Assess: what data, how many principals, what window
4. Notify: Data Protection Board and affected principals per DPDP timelines; status page immediately
5. Remediate and verify
6. Blameless post-mortem, published in the transparency report

---

## 9. Moderation operations

```bash
tsctl mod queue --priority high --limit 20
tsctl mod decide --item <id> --action remove --rule harassment --note "..."
tsctl mod appeal list
tsctl dispute list --status open
tsctl correction publish --target issue:<id> --field contractor_name \
  --old "..." --new "..." --reason "source corrected upstream"
```

Every action writes to the audit log. Account suspensions and platform-statement withdrawals require
a second reviewer (`--second-reviewer <id>`).

---

## 10. Routine schedule

| Cadence | Task |
|---|---|
| Daily | Alert triage; error-budget check; moderation queue |
| Weekly | Funnel and quality metrics; classification override sampling; ingestion health |
| Monthly | Cost review; SLO review; risk register; dependency updates |
| Quarterly | Restore drill; access review; secret rotation; vulnerability audit; transparency report |
| April | Monsoon preparation (§4) |
| Annually | Impact report; legal review of published methodologies |

---

## 11. Escalation

| Situation | Action |
|---|---|
| P1 beyond one person's capacity | Call the secondary contact; if none, post publicly on the status page and be honest about the ETA |
| Legal notice | Preserve everything; do not modify the disputed record; obtain counsel before responding |
| Suspected data breach | §8, immediately; do not wait for certainty |
| Physical safety concern for a reporter | Make the account's public footprint private; contact the user; review coarsening controls |

---

## 12. `tsctl` — the operator CLI

Every command above is part of one binary, deployed alongside the services. It talks to the admin
API, not directly to the database, so every operational action is authenticated and audited.

**There are no manual production database edits.** If a fix requires one, it requires a migration or
an admin-API endpoint — both of which are reviewed and logged.
