# ADR 0016 — Collection runs as a Cloud Run job, triggered by Cloud Scheduler

**Status:** Accepted, and running · **Date:** 2026-09-25
**Supersedes the mechanism of** [ADR 0015](0015-indian-egress-for-collection.md); its finding stands.

## Context

[ADR 0015](0015-indian-egress-for-collection.md) established the constraint that matters: BMC
answers Indian addresses and refuses foreign ones. It met that constraint with a scheduled
start/stop VM in `asia-south1`, because a Cloud Run job's egress was not known to present as
Indian, and the documented fix — Cloud NAT — cost more than the VM.

Building the VM answered a second question: BMC filters by **geography, not network type**. A
Google datacenter address in Mumbai is accepted. That made the Cloud Run question worth testing
rather than assuming, and a job deployed to `asia-south1` on 25 September 2026 reached every BMC
endpoint, including the roads API on port 3000.

## Decision

**Collection runs as a Cloud Run job in `asia-south1`, triggered by Cloud Scheduler.**

- One job, `tracesarkar-collector`, built from the repository `Dockerfile`.
- Its command is `ingest all`: the snapshot, then the resolution watch, in one invocation.
- Two schedules, 02:30 and 14:30 IST, both calling the job with no arguments.
- Credentials are mounted from Secret Manager; the source register is baked into the image.
- The VM, its disk and its instance schedule are deleted.

## Rationale

- **Cheaper and simpler at once**, which is unusual. No disk, so no standing charge; no machine, so
  nothing to patch; no startup script, no binary download, no serial console to lose.
- **Costs nothing.** The job uses about 1.5% of Cloud Run's monthly free vCPU allowance, and the
  triggers are 2 of 3 free Cloud Scheduler jobs.
- **Twice daily rather than nightly.** Cloud Monitoring's absence conditions cap at 23h30m, so a
  daily schedule cannot be watched for a missed run without false alarms every day. Running twice
  makes the dead-man alert meaningful, and it catches BMC's daytime edits as well.
- **One command, not argument overrides.** Passing different arguments per schedule requires
  `run.jobs.runWithOverrides`, which `roles/run.invoker` does not grant. Widening the trigger's
  permissions to collect resolutions is the wrong trade, so `ingest all` does both.

## Alternatives

| Option | Why not |
|---|---|
| **Keep the start/stop VM** (ADR 0015) | Works, and is now proven, but costs ~₹42/month for the boot disk and carries a machine, a startup script and a schedule policy. Kept in git history; `deploy/gcp/` can be restored if Cloud Run's egress ever stops presenting as Indian |
| **Cron on the free US VM triggering the job** | Free and uses a machine that exists, but collection would then depend on that machine's health, with nothing watching it. Cloud Scheduler is the managed form of the same thing |
| **GitHub Actions triggering the job** | Needs Workload Identity Federation, and GitHub delays scheduled runs under load and disables schedules in repositories inactive for 60 days — unacceptable for an archive that must not miss days |
| **Create and delete a VM nightly from a backend** | Reaches the same ₹0 with far more machinery, and needs `compute.instanceAdmin` on an internet-facing box. It would also be discarded at Phase 1 |
| **Cloud Run with Cloud NAT for a static egress IP** | The documented way to pin egress geography, at roughly $30/month. Only worth it if Google's shared pool ever routes `asia-south1` jobs through a non-Indian address |

## Consequences

- **Cost falls to zero** for collection. What remains is the Aiven database and the R2 bucket.
- **The egress IP is not ours.** It comes from a shared Google pool that presents as Indian today.
  If that changes, collection fails visibly (the run log and the failure alert), and the fallback is
  the VM kit in git history or Cloud NAT.
- **Alerting now exists**, which it did not under ADR 0015: an email on a failed execution, and a
  second on no successful execution for 23h30m. Both go to the account's email address.
- At **Phase 1** the backend becomes a long-running service in `asia-south1` and can own its own
  schedule. Cloud Scheduler and this job go away then, and collection becomes one more thing the
  backend does.
