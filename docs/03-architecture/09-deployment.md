# Deployment

**Target:** a small cluster of VPS instances running Docker Swarm, managed by Dokploy, that survives
a monsoon traffic spike and can be operated by one person.

---

## 1. Why this stack

| Choice | Rationale | Alternative rejected |
|---|---|---|
| **VPS + Docker Swarm** | Predictable cost, no egress surprises, full control of data residency | Managed Kubernetes — operational overhead far exceeds the benefit at this scale |
| **Dokploy** | Open-source PaaS over Swarm; handles builds, env vars, Traefik, TLS, and rollouts; multi-node cluster support | Hand-rolled Ansible — more to maintain; Vercel/Render — cost and data-residency issues |
| **Traefik** | Dokploy's built-in reverse proxy; Swarm-aware service discovery, automatic Let's Encrypt | nginx — more manual config for the same result |
| **Managed Postgres or self-hosted with replication** | Postgres is the single source of truth; the backup and restore story must be boring | — |
| **S3-compatible object storage** — Cloudflare R2 ([D046](../00-overview/05-decision-log.md)) | Media and archive; lifecycle rules; CDN in front of public derivatives. A bucket lock with indefinite retention makes `archive/` write-once. `media/` stays erasable for data-principal rights | Local disk — no |

**Data residency:** Indian region where available. Civic data about Indian citizens processed under
DPDP should not be casually offshored, and a domestic region is easier to defend in any government
conversation.

---

## 2. Topology

```
                      ┌──────────────┐
        Internet ────▶│  CDN         │  public media, share images, static
                      └──────┬───────┘
                             │
                      ┌──────▼───────┐
                      │  Traefik     │  TLS, routing, rate limit  (Swarm service)
                      └──────┬───────┘
      ┌──────────────────────┼──────────────────────┐
      │                      │                      │
┌─────▼─────┐          ┌─────▼─────┐          ┌─────▼─────┐
│  node-1   │          │  node-2   │          │  node-3   │
│  manager  │          │  worker   │          │  worker   │
│           │          │           │          │           │
│  api ×2   │          │  api ×2   │          │  worker×4 │
│  realtime │          │  worker×2 │          │  ingest   │
│  admin    │          │           │          │  scheduler│
└───────────┘          └───────────┘          └───────────┘
      │                      │                      │
      └──────────────────────┼──────────────────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
      ┌───────▼────────┐           ┌────────▼───────┐
      │ Postgres 16    │           │ Redis          │
      │ + PostGIS      │──replica──│                │
      │ (primary)      │           └────────────────┘
      └────────────────┘
              │
      ┌───────▼────────┐
      │ Object storage │  media (public + private), archive (write-once)
      └────────────────┘
```

**Baseline sizing (v0.1–v0.3, one ward → Greater Mumbai):**

| Component | Spec |
|---|---|
| 3 × app nodes | 4 vCPU / 8 GB / 80 GB NVMe |
| Postgres primary | 4 vCPU / 16 GB / 200 GB NVMe (grows with media metadata and archive index) |
| Postgres replica | 2 vCPU / 8 GB |
| Redis | 2 vCPU / 4 GB |
| Object storage | Pay-per-GB; media dominates |

---

## 3. Services and scaling

| Service | Replicas (baseline) | Scale trigger |
|---|---|---|
| `api` | 4 | p95 latency > 300 ms or CPU > 70% |
| `realtime` | 2 | Connection count > 8k per instance |
| `worker` | 6 | Queue depth > 500 or oldest-job age > 120 s |
| `ingest` | 1 | Never horizontally scaled (advisory-locked) |
| `scheduler` | 1 | Singleton |
| `admin` | 1 | — |

Swarm handles rolling updates with health checks:

```yaml
deploy:
  replicas: 4
  update_config:
    parallelism: 1
    delay: 15s
    order: start-first
    failure_action: rollback
  restart_policy:
    condition: on-failure
    max_attempts: 3
  resources:
    limits:   {cpus: '1.0', memory: 1G}
    reservations: {cpus: '0.25', memory: 256M}
healthcheck:
  test: ["CMD", "/app/healthcheck"]
  interval: 15s
  timeout: 3s
  retries: 3
  start_period: 20s
```

---

## 4. Monsoon capacity plan

The defining operational risk: **10–50× baseline traffic in an 8–10 week window**, concentrated in
hours, with degraded client connectivity.

| Mechanism | Detail |
|---|---|
| **Pre-scale** | Manually scale `api` and `worker` at monsoon onset. Cheap insurance; do not rely on reactive autoscaling for a predictable seasonal event |
| **Queue-first writes** | Report submission is durable-then-async. A spike lengthens the enrichment queue; it does not fail submissions |
| **Backpressure, not errors** | When the queue is deep, `202` responses carry a longer `estimated_ready_ms`. Never 5xx a citizen standing in the rain |
| **Shed enrichment, never intake** | Under extreme load, disable optional enrichment (context, share-kit pregeneration) before touching classification, and classification before intake |
| **Aggressive dedup** | During a burst at one location, merge on spatial+phash alone and skip per-report classification |
| **CDN everything public** | Share images, permalinks, and heatmaps are static and CDN-cacheable |
| **Emergency mode** | An admin switch that surfaces a site-wide banner, prioritises `hazard_to_life` reports, and diverts them to the emergency channel — modelled on FixMyStreet Pro's emergency diversion |

**Load test before every monsoon.** A scripted scenario replaying a 20× spike against staging, run in
April, is a scheduled task not a good intention.

---

## 5. Configuration and secrets

| Item | Where |
|---|---|
| Non-secret config | Environment variables via Dokploy, per environment |
| Secrets (DB, S3, API keys) | Docker secrets / Dokploy secret store; never in the repo, never in image layers |
| Legal constants | The `legal_constants` **database table**, not environment variables — they need versioning and citations |
| Feature flags | Database-backed, admin-editable, with an audit trail |

Secret rotation: quarterly for API keys, immediately on any suspected exposure. The rotation runbook
is in [operations runbook](../06-operations/01-runbook.md).

---

## 6. CI/CD

```
push to main
   │
   ├─▶ lint (golangci-lint) · vet · gosec
   ├─▶ unit tests + race detector
   ├─▶ golden-file tests (jurisdiction eval set)
   ├─▶ AI eval suite (on prompt/model changes only)
   ├─▶ migration dry-run against a staging snapshot
   ├─▶ build multi-arch image · SBOM · sign
   ├─▶ deploy to staging · smoke tests
   └─▶ manual gate ──▶ deploy to production (rolling)
```

**Migrations run as a separate, explicit step before the rolling deploy**, and are forward-only.
Every migration must be safe against the previous application version (expand/contract pattern), so
a rollback of the app does not require a rollback of the schema.

---

## 7. Backup and recovery

| Asset | Strategy | RPO | RTO |
|---|---|---|---|
| Postgres | Continuous WAL archiving + nightly base backup, off-site | 5 min | 1 h |
| Object storage (media) | Cross-region replication | 15 min | 4 h |
| Object storage (archive) | Versioned, replicated, write-once | 15 min | 24 h |
| Redis | None — reconstructible | ∞ | 5 min |
| Configuration | In git; secrets in the secret store with backup | — | 30 min |

**Restore is tested quarterly**, end to end, into a scratch environment, with the runbook followed
verbatim. A backup that has never been restored is a hypothesis.

---

## 8. Domains and TLS

| Host | Serves |
|---|---|
| `tracesarkar.org` | Public web |
| `api.tracesarkar.org` | Core + Open311 + public API |
| `cdn.tracesarkar.org` | Public media and share images |
| `admin.tracesarkar.org` | Admin/moderation — IP-restricted and behind SSO |
| `status.tracesarkar.org` | Status page, hosted externally so it survives an outage |

TLS via Let's Encrypt through Traefik, auto-renewed. HSTS with preload once stable.

---

## 9. Environments

| | `local` | `staging` | `production` |
|---|---|---|---|
| Orchestration | Docker Compose | Swarm (1 node) | Swarm (3+ nodes) |
| Data | Synthetic seed | Anonymised subset + real boundary/contract data | Real |
| External APIs | Mocked | Sandbox/low-quota keys | Production keys |
| Media bucket | MinIO | Separate bucket | Production bucket |
| Access | Developer | Team | Break-glass, audited |

**Production media and archive buckets are never readable from non-production environments.** This is
enforced by IAM policy, not by convention.

**Phase 0 exception ([D046](../00-overview/05-decision-log.md)).** Snapshot history cannot be
re-collected later, so the Phase 0 ingesters write to the **production archive bucket** from their
first run, even though they run on the staging node. They hold write-only credentials scoped to
`archive/` and `media/`. When production exists, the ingesters move there and staging loses those
credentials.

---

## 10. Cost envelope

See [cost model](../06-operations/04-cost-model.md) for the full breakdown. Infrastructure at the
baseline sizing is a small monthly figure; the **variable AI inference cost dominates** and is the
reason for the batch path, prompt caching, dedup-before-classify, and the daily budget cap.

---

## 11. Operational principles

1. **One person must be able to operate it.** Every runbook step is a documented command.
2. **Every deploy is reversible** within one rolling update.
3. **No manual production database edits.** Everything through migrations or the admin service, both
   audited.
4. **Alert on symptoms, not causes.** Page on "reports failing", not on "CPU high".
5. **Degrade loudly.** Users are told what is not working; the status page is not optional.
