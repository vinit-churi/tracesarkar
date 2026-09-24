# ADR 0015 — Collection runs from an Indian IP, on a scheduled start/stop VM

**Status:** Accepted · **Date:** 2026-09-24

## Context

Phase 0 collects BMC's published works data daily. The first scheduled run from GitHub Actions
failed in a way that documentation would never have revealed: `roads.mcgm.gov.in` and
`portal.mcgm.gov.in` refuse connections from foreign cloud runners, while `swd.mcgm.gov.in` and the
Internet Archive answer normally. Evidence in
[the probe record](../01-research/sources/2026-09-24-bmc-geo-restriction-probe.md).

The roads API is the single most valuable source the platform has: 2,237 works and 2,405 geometries
carrying contractor, work code, dates and status. It cannot be collected from abroad.

A further constraint shapes the choice: the job is **short, daily, and unrepeatable**. It runs for
two to three minutes. A night that does not run is a day of BMC's own record that nobody kept, and
it cannot be recovered afterwards — the API publishes current state, not history.

## Decision

**Collection that touches an India-restricted source runs from a VM inside an Indian region**, on
an on-demand `e2-micro` in `asia-south1`, started and stopped on a schedule:

- A Compute Engine **instance schedule** starts the VM at 02:30 IST and stops it at 03:00 IST.
- A **startup script** runs the collector and shuts the machine down when it finishes; the scheduled
  stop is the safety net if it hangs.
- Credentials come from **Secret Manager** at boot, never from a disk image.
- **GitHub Actions keeps** the sources reachable from anywhere: the storm-water drain API and the
  Government Resolution watcher.

Cost is roughly ₹45 per month, dominated by the boot disk rather than by compute.

## Rationale

- **An Indian IP is not optional.** It is the one requirement that rules out every free option we
  would otherwise have taken: GitHub Actions, and Google's always-free `e2-micro`, which exists only
  in US regions.
- **Start/stop rather than always-on**, because the duty cycle is two minutes a day. Always-on costs
  about ₹600 a month for 99.8% idle.
- **On-demand rather than spot.** At this duty cycle spot saves about ₹2 a month, because the disk
  dominates the bill, and it introduces the one failure this phase exists to prevent: a night that
  cannot start because the region is short of capacity. Spot suits long, restartable work. This work
  is short, daily and unrepeatable.
- **Instance schedules rather than Cloud Scheduler**, for the VM design. Scheduler plus Pub/Sub plus
  a function is four moving parts to start a machine that Compute Engine can start by itself.
- **Splitting collection by reachability** keeps half the record flowing free of charge, and means a
  broken VM does not stop the drains or the resolutions.

## Alternatives

| Option | Why not |
|---|---|
| **Cloud Run job + Cloud Scheduler** | Genuinely free and has no server to maintain, but a job's default egress is not guaranteed to geolocate to its region, and the documented fix — routing egress through Cloud NAT — costs more per month than the VM it was meant to replace. Worth revisiting if a job's egress is ever confirmed to present as Indian. |
| **Always-on VM in Mumbai** | ~₹600/month for a machine idle 99.8% of the time. Revisit at Phase 1, when a server is needed continuously for the public API; the collector then moves onto it and this VM goes away. |
| **Spot VM** | Saves ~₹2/month and risks missed nights. See rationale. |
| **Oracle Cloud Always Free (Mumbai)** | Free permanently and a real Indian IP. Rejected for now because free-tier instances can be reclaimed when idle, which is precisely our failure mode, and because it adds a second cloud account to operate. A reasonable fallback if cost ever matters more than certainty. |
| **A machine at home (Raspberry Pi, old laptop)** | Free, and a residential Indian IP, which is the only option that certainly works today. Rejected as the default because it depends on domestic power and broadband. **It becomes the answer if the VM's first boot shows that Indian datacenter ranges are blocked too.** |
| **Keep everything on GitHub Actions** | Loses the roads API entirely, which is most of the platform's value. |

## Consequences

- A GCP project, billing account and `gcloud` are now operational dependencies. The deployment kit
  lives in [`deploy/gcp/`](../../deploy/gcp/README.md).
- The collector binary is built by CI and published as a release asset, so the VM downloads a
  known artefact at boot instead of compiling on a micro instance.
- **The first boot is also an experiment.** If BMC answers an Indian datacenter IP, this design
  stands. If it refuses, the filter is by network type rather than geography, the VM is discarded,
  and collection moves to a machine on a residential Indian connection. That outcome gets its own
  ADR.
- Secrets now exist in three places — a local `.env`, GitHub Actions secrets, and GCP Secret
  Manager. Each should be scoped down and rotated; that is a backlog item, not a blocker.
- Every future source needs a reachability check from where the code will run, not from a laptop.
  The source register records this per source.
