# The collector on Cloud Run

Collection runs as a Cloud Run job in `asia-south1`, triggered by Cloud Scheduler
([ADR 0016](../../docs/04-adr/0016-collection-as-a-cloud-run-job.md)). BMC answers Indian addresses
only, and a job in that region satisfies it — with no VM and no disk.

The VM this replaced is in git history at `deploy/gcp/`, before commit `ADR 0016`.

## What exists in GCP

| Resource | Name | Notes |
|---|---|---|
| Cloud Run job | `tracesarkar-collector` | `asia-south1`, 1 vCPU, 512 MiB, command `ingest all` |
| Scheduler | `tracesarkar-nightly` | 02:30 IST |
| Scheduler | `tracesarkar-midday` | 14:30 IST — also what makes the missed-run alert reliable |
| Secrets | `tracesarkar-env`, `tracesarkar-ca` | Mounted at `/etc/ts/env` and `/etc/ts-ca/ca.pem` |
| Service account | `tracesarkar-collector@…` | Reads those secrets; may invoke the job |
| Alerts | Two policies | A failed execution, and no success for 23h30m |

Cost: nothing. The job uses about 1.5% of the monthly free vCPU allowance, and 2 of 3 free
scheduler jobs.

## Deploying a change

```sh
gcloud run jobs deploy tracesarkar-collector --source . --region asia-south1 --quiet \
  --service-account tracesarkar-collector@$(gcloud config get-value project).iam.gserviceaccount.com \
  --set-secrets "/etc/ts/env=tracesarkar-env:latest,/etc/ts-ca/ca.pem=tracesarkar-ca:latest" \
  --set-env-vars "TRACESARKAR_ENV_FILE=/etc/ts/env,POSTGRES_CA_PATH=/etc/ts-ca/ca.pem,TRACESARKAR_TIER=personal" \
  --max-retries 1 --task-timeout 15m --memory 512Mi --cpu 1
```

Cloud Build compiles the image from the `Dockerfile` at the repository root.

## Running it now, rather than waiting for a schedule

```sh
gcloud run jobs execute tracesarkar-collector --region asia-south1 --wait
# or exercise the whole chain, trigger included:
gcloud scheduler jobs run tracesarkar-nightly --location asia-south1
```

Then, from your laptop:

```sh
make status     # per-endpoint health, and the recent runs with their outcomes
make changes    # what changed in BMC's published data
```

## When something breaks

Two emails can arrive: one when an execution fails, one when nothing has succeeded for 23h30m
(which means at least two scheduled runs were missed). Both point here.

```sh
gcloud run jobs executions list --job tracesarkar-collector --region asia-south1
gcloud logging read 'resource.type="cloud_run_job"' --limit 50 --freshness=1d
```

The run log in the database survives the container, so `make status` is usually the faster answer.

## If Cloud Run's egress ever stops reaching BMC

The job's outbound address comes from a shared Google pool that presents as Indian today. If BMC
starts refusing it, the failure is loud. Two ways back: restore the VM kit from git history, or put
Cloud NAT in front of the job with a static Indian address (~$30/month).
