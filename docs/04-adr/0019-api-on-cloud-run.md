# ADR 0019 — The API moves to Cloud Run

**Status:** Accepted · **Date:** 2026-09-29 · **Supersedes:**
[ADR 0018](0018-api-on-dokploy.md) · **Related:** [ADR 0016](0016-collection-as-a-cloud-run-job.md)

## Context

[ADR 0018](0018-api-on-dokploy.md) put the API on a shared Dokploy instance because that instance
already existed, with Traefik and Let's Encrypt in front of it. It recorded the risk plainly: *"the
server is shared with unrelated projects. A noisy neighbour is now a way this can degrade, and the
blast radius of that host is wider than this project."*

On 29 September 2026 that host became unreachable — 100% packet loss, port 443 filtered, and the
Dokploy control panel down with it. Not our container: the whole machine. The API was down and
there was no way to restart it from inside this project, because the machine belongs to a wider
estate.

The risk the ADR named is the one that happened, four days later.

That ADR also rejected Cloud Run for the API, reasoning that *"a request-scheduled container
cold-starts, and a capture upload is exactly the request that must not wait."* Measured rather than
assumed, that concern was overstated: a warm request answers in ~35 ms, and a cold start costs a
couple of seconds on the first request after idle — once, for a capture that is then stored
durably. Against a service that is simply *down*, it is not a close call.

## Decision

**The API runs as a Cloud Run service in `asia-south1`, in the same project as the collector.**

- Built from `Dockerfile.api` via `deploy/cloudrun/api.cloudbuild.yaml`. A plain
  `gcloud builds submit --tag` would use the repository's default `Dockerfile`, which builds the
  *collector*; the API's build has to name its own file.
- Secrets come from Secret Manager, reusing the collector's service account: `tracesarkar-env` and
  `tracesarkar-ca` mount as files, `tracesarkar-auth-secret` and `tracesarkar-api-token` map to
  environment variables.
- `--min-instances 0`, so an idle service costs nothing. Raise it to 1 if cold starts ever become
  visible to a person rather than to a test.
- TLS, certificate renewal and the public hostname are Google's problem, not ours.

**Health is served on `/v1/health` as well as `/healthz`.** Google's frontend answers `/healthz`
itself: on Cloud Run the request never reaches the container, returns Google's own 404 page, and
produces no entry in the request log, while `/` and `/v1/auth/me` on the same service logged and
responded normally. `/healthz` is kept because it works over loopback, where no proxy is involved,
which is what the image's `HEALTHCHECK` uses.

## Alternatives

| Option | Why not |
|---|---|
| **Wait for the shared host to come back** | It is not ours to restart, and the next outage has the same shape. The dependency is the problem, not this instance of it |
| **A dedicated GCE VM** | A machine to patch, a TLS certificate to renew, a deploy mechanism to build — all of which Cloud Run provides. An `e2-micro` in `asia-south1` also costs real money monthly, where this is likely to stay inside the free tier |
| **Our own Dokploy on our own VM** | Same maintenance burden as a VM, plus Dokploy itself to keep upgraded, to regain a control panel we used four times |
| **Cloud Run in a US region** | The database and bucket are reachable from anywhere, but the users are in Mumbai. `asia-south1` is where the collector already runs |

## Consequences

- **No machine to maintain.** No patching, no certificate renewal, no disk filling up.
- **Cold starts exist.** At `--min-instances 0` the first request after an idle period pays a
  couple of seconds. Acceptable at personal tier; revisit before the public tier, where it becomes
  one flag and a small monthly cost.
- **The hostname changed**, and it is a `run.app` address. Like the `sslip.io` name before it, this
  must become a real domain before the public tier — a civic platform asking people to trust it
  cannot live on an opaque generated hostname.
- **Production secrets rotated again**, because they are new secrets in Secret Manager rather than
  values copied from the dead host. The Dokploy environment's copies are stale and should be
  treated as revoked.
- **Deploys are now explicit**, not push-to-main. `gcloud builds submit` then `gcloud run deploy`.
  That is a loss of convenience and a gain in control; a GitHub Actions trigger can restore the
  convenience once anyone other than the operator depends on the service.
- **Collection and the API now sit in one project**, sharing a service account and two secrets.
  They remain separate workloads with separate lifecycles — the collector runs twice a day and
  exits, the API serves — and neither deploy touches the other.
