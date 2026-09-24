# The Indian collector VM

BMC's roads API and portal refuse connections from foreign cloud runners
([probe](../../docs/01-research/sources/2026-09-24-bmc-geo-restriction-probe.md)), so the part of
collection that touches them runs from a VM in `asia-south1`. The reasoning, and what was rejected,
is in [ADR 0015](../../docs/04-adr/0015-indian-egress-for-collection.md).

## What this is

| | |
|---|---|
| Machine | `e2-micro`, `asia-south1-c`, on-demand, 10 GB standard disk |
| Lifetime | About three minutes a night |
| Schedule | Instance schedule starts it at 02:30 IST; stops it at 03:00 IST if it has not already powered off |
| Credentials | Secret Manager, read at boot into tmpfs, gone at shutdown |
| Binary | Downloaded at boot from the `collector-latest` release, checksum verified |
| Cost | Roughly ₹45/month, mostly the disk |

Nothing is installed on the disk beyond Debian. The VM is disposable: delete it, re-run `setup.sh`,
and it comes back identical.

## Setting it up

You need a GCP project with billing enabled, and `gcloud` installed and logged in.

```sh
gcloud config set project <your-project-id>
./deploy/gcp/setup.sh
```

It creates a service account, stores `.env` and `ca.pem` as secrets, creates the VM, and attaches
the schedule. Re-running it is safe: existing pieces are left alone, and the secrets and startup
script are refreshed.

## Testing it without waiting for 02:30

```sh
gcloud compute instances start tracesarkar-collector --zone asia-south1-c
gcloud compute instances tail-serial-port-output tracesarkar-collector --zone asia-south1-c
```

The VM shuts itself down when it finishes, so the serial output ends on its own.

**What to look for on the first boot.** It answers a question no documentation can:

- `"msg":"snapshot" ... "endpoint":"publicdashboard"` — BMC accepts Indian datacenter IPs. This
  design stands.
- `dial tcp ...:3000: i/o timeout` — the filter is by network type, not geography. Then no cloud
  works, collection moves to a machine on a residential Indian connection, and ADR 0015 gets a
  successor.

## Changing the schedule

```sh
gcloud compute resource-policies describe tracesarkar-nightly --region asia-south1
```

To change the times, delete the policy and re-run `setup.sh` with `START_CRON` and `STOP_CRON` set.

## Turning it off

```sh
gcloud compute instances delete tracesarkar-collector --zone asia-south1-c
```

Collection of the drain API and Government Resolutions continues on GitHub Actions, because those
sources answer from anywhere. Only the roads API and the BMC portal need this machine.

## What runs where

| Source | Where it is collected | Why |
|---|---|---|
| BMC roads API (works, geometries, wards) | This VM | India-only |
| BMC portal (tenders, C1 list, rate schedules) | This VM, when those collectors exist | India-only |
| BMC storm-water drain API | GitHub Actions, and here as well | Answers from anywhere |
| Government Resolutions (Internet Archive) | GitHub Actions, and here as well | Answers from anywhere |

The overlap is deliberate. Collecting the same unchanged bytes twice costs one row in `fetch_log`
and nothing else, and it means neither runner is a single point of failure.
