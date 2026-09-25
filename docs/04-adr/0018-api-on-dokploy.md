# ADR 0018 — The API runs on Dokploy; collection stays on Cloud Run

**Status:** Accepted · **Date:** 2026-09-25 · **Related:**
[ADR 0015](0015-indian-egress-for-collection.md), [ADR 0016](0016-collection-as-a-cloud-run-job.md)

## Context

The v0.1 client needs an HTTPS endpoint it can reach. Until now the API ran only on a laptop, which
is enough to develop against and useless for anything else — a phone in a ward cannot reach it, and
a browser will not send a capture to a plain-HTTP address because geolocation requires a secure
context.

Collection already has a home and a reason for it: BMC's portal geo-restricts by geography, so
[ADR 0015](0015-indian-egress-for-collection.md) put the collector in Mumbai and
[ADR 0016](0016-collection-as-a-cloud-run-job.md) made it a Cloud Run job that runs twice a day and
exits. **The API has no such constraint.** It talks to Aiven and Cloudflare R2, neither of which
cares where the request comes from, and it must stay up rather than run and exit.

A Dokploy instance already exists and already hosts other projects, with Traefik and Let's Encrypt
in front of it.

## Decision

**The API is a Docker Compose service on the existing Dokploy instance, built from this repository.
Collection stays on Cloud Run.** Two workloads, two schedules, two deploys.

- `Dockerfile.api` builds `cmd/api` only, onto `distroless/static`. It is separate from the
  collector's `Dockerfile`: one image with two entrypoints would couple deploys that have no reason
  to move together.
- `docker-compose.yml` is committed and contains **no secret**. Every value is an `${...}`
  reference; Dokploy writes the service environment to a `.env` beside the compose file before
  building.
- The database CA travels as `POSTGRES_CA_PEM`, base64-encoded. Dokploy's only secret channel is
  environment variables, and a multi-line PEM crosses two parsers on the way in. The process
  materialises it into a `0600` temp file at startup and connects with `sslmode=verify-full`.
- The container's `HEALTHCHECK` runs `api health`, because `distroless` has no shell and no curl.
- The hostname is a generated `sslip.io` name resolving to the server's address, with a Let's
  Encrypt certificate. It works without touching DNS, which is the right trade while the surface is
  personal-tier.

## Alternatives

| Option | Why not |
|---|---|
| **Cloud Run for the API too** | A request-scheduled container cold-starts, and a capture upload is exactly the request that must not wait. It also bills per request, where this is a process that should simply stay up |
| **The free US VM** | No orchestration, no TLS termination, no deploy story — every one of which would have to be built and then maintained. The Dokploy instance already has all three |
| **A managed platform (Fly, Render, Railway)** | Another account, another bill, another place secrets live. Nothing here needs what they add |
| **Run the API next to the collector, in Mumbai** | Would couple two workloads with different lifecycles for a geography constraint only one of them has |

## Consequences

- The server is **shared with unrelated projects**. A noisy neighbour is now a way this can degrade,
  and the blast radius of that host is wider than this project.
- No autoscaling. One container, restarted on failure. Adequate at personal tier and a known limit
  to revisit before anything public.
- Nothing stateful lives in the stack: the database is Aiven and the bucket is Cloudflare R2, so the
  whole thing can be destroyed and recreated from the repository.
- **The hostname must change before the public tier.** An `sslip.io` name encodes the server's IP
  address, which is fine for a private surface and wrong for a civic platform people are asked to
  trust.
- Pushing to `main` deploys. That is convenient now and will need a gate — a tag or a manual
  promotion — once anyone other than the operator depends on the service being up.
- Production holds its **own** `AUTH_SECRET` and `API_TOKEN`, generated at deploy. A session minted
  on a laptop is not valid in production, which is the point.
