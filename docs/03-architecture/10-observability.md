# Observability

**Rule:** alert on symptoms a citizen would notice, not on causes an engineer finds interesting.

---

## 1. The three signals

| Signal | Stack | Retention |
|---|---|---|
| **Metrics** | Prometheus + Grafana | 90 days (15 s), 1 year (5 min rollup) |
| **Logs** | Structured JSON via `log/slog` → Loki | 30 days |
| **Traces** | OpenTelemetry → Tempo | 7 days, sampled |

Self-hosted alongside the application on the Swarm cluster. A hosted alternative is acceptable if it
does not receive PII — which, given the redaction rules below, it would not.

---

## 2. Logging rules

```go
slog.Info("issue routed",
    "issue_id", issue.ID,
    "authority", res.Authority.Code,
    "ward", res.Ward.Code,
    "confidence", res.Confidence,
    "resolver_version", res.Version,
    "request_id", middleware.RequestID(ctx),
)
```

**Never logged, ever:**

| Forbidden | Why |
|---|---|
| Phone numbers (even hashed, at info level) | PII |
| Precise coordinates | Location PII |
| Media URLs with signed tokens | Grants access to private originals |
| Full request bodies | May contain images and descriptions |
| Model prompts containing user content | Same |
| API keys, tokens, session identifiers | Obvious |

A CI check greps for forbidden field names in log call sites. It is crude and it works.

**Always included:** `request_id`, and `issue_id` / `report_id` where applicable, so a user complaint
("my report vanished") is traceable end to end from one identifier.

---

## 3. Golden signals per service

| Service | Latency | Traffic | Errors | Saturation |
|---|---|---|---|---|
| `api` | p50/p95/p99 per route | req/s | 5xx rate, 4xx by class | CPU, goroutines, DB pool |
| `realtime` | Event delivery lag | Connections, events/s | Drops, auth failures | Connections per instance, buffer occupancy |
| `worker` | Job duration by stage | Jobs/s | Failure rate by stage | Queue depth, oldest job age |
| `ingest` | Run duration | Records/run | Parse failures, HTTP errors | Runs behind schedule |

---

## 4. Domain metrics

The ones that actually indicate whether the product works:

| Metric | Meaning |
|---|---|
| `reports_submitted_total{ward,category}` | Intake |
| `report_enrichment_duration_seconds{stage}` | The 6-second promise |
| `classification_confidence` (histogram) | Model health in production |
| `classification_user_override_rate` | The model was wrong and the user corrected it — the truest quality signal |
| `jurisdiction_confidence` (histogram) | Routing health |
| `jurisdiction_disambiguation_rate{ward}` | Where boundary data is weak |
| `jurisdiction_unresolved_total{ward}` | Coverage gaps |
| `attribution_published_rate` | Contract-matching yield |
| `attribution_confidence` (histogram) | Matching health |
| `issues_by_status{status,authority}` | The funnel |
| `sla_breach_rate{authority,ward}` | **The headline public number** |
| `claimed_vs_confirmed_ratio{authority}` | **The accountability number** |
| `recheck_response_rate` | Whether the confirmation loop works |
| `escalations_generated_total{kind}` | Consequence engine usage |
| `share_kit_generated_total{channel}` | Distribution |
| `ai_tokens_total{call,model}` / `ai_cost_paise_total` | Cost |
| `ai_cache_read_ratio` | Prompt caching effectiveness |

The last two exist because AI inference is the dominant variable cost and the easiest thing to
silently overspend on.

---

## 5. Alerting

### Page (wake someone up)

| Alert | Condition |
|---|---|
| Report submission failing | 5xx on `POST /v1/reports` > 2% for 5 min |
| Database primary down | Health check failing 1 min |
| Object storage unavailable | Upload failures > 10% for 5 min |
| Deadline scheduler stalled | `deadline_scheduler_last_run` > 30 min |
| Outbox lag | > 60 s for 5 min |
| Data breach indicator | Anomalous bulk read on sensitive tables |

### Ticket (fix within a day)

| Alert | Condition |
|---|---|
| Worker queue backing up | Oldest job > 15 min |
| Ingester stale | No success in 3× cadence |
| Parse failure rate | > 5% for a source |
| AI budget | > 80% of daily cap |
| Classification override rate | > 15% over 24 h (model regression) |
| Jurisdiction unresolved | > 10% in any covered ward |
| Certificate expiry | < 14 days |

### Inform (dashboard only)

Everything else. **Resist the urge to alert on it.** A noisy pager is an ignored pager.

---

## 6. Dashboards

| Dashboard | Audience | Contents |
|---|---|---|
| **Service health** | Operator | Golden signals, error budget, deploys |
| **Pipeline** | Operator | Queue depth, stage durations, failure rates by stage |
| **Data freshness** | Operator + editorial | Per-source last success, record counts, parse rates |
| **AI cost** | Operator | Tokens and cost by call type and model, cache hit ratio, budget burn-down |
| **Civic outcomes** | Everyone, **public** | SLA breach rates, claimed-vs-confirmed, resolution times by ward and authority |

That last dashboard is a product feature, not internal tooling. A platform that demands transparency
from municipal bodies and publishes nothing about its own operation is not credible.

---

## 7. Tracing

Traced end to end:

```
POST /v1/reports
 ├─ auth
 ├─ media upload (S3)
 ├─ db insert
 └─ enqueue
      └─ worker: enrich_report
           ├─ redact
           ├─ classify (Claude)         ← external span, tokens + cost as attributes
           ├─ locate (PostGIS)
           ├─ dedupe
           ├─ attribute
           └─ publish events
```

Sampling: 100% of errors, 100% of requests slower than 2 s, 1% baseline. Every external API call is
a span with duration, status, retry count, and — for model calls — token counts and cost.

**No PII in span attributes.** Coordinates become an H3 cell at low resolution; media becomes a hash.

---

## 8. SLOs

| SLO | Target | Window |
|---|---|---|
| Report submission availability | 99.5% | 30 days |
| Report submission p95 latency | < 3 s (upload excluded) | 30 days |
| Enrichment completion | 95% within 30 s | 30 days |
| Public page availability | 99.9% | 30 days |
| Public API availability | 99.5% | 30 days |
| Notification delivery (SLA breach) | 99% within 5 min | 30 days |

Error budgets are tracked. Burning the submission budget freezes feature deploys until it recovers.

Deliberately **not** an SLO: enrichment completeness during a monsoon spike. The design choice is to
accept a longer queue rather than reject submissions, and the SLO must not create pressure to
reverse that.

---

## 9. Health endpoints

| Endpoint | Checks |
|---|---|
| `GET /healthz` | Process alive |
| `GET /readyz` | DB reachable, migrations current, object storage reachable, Redis reachable |
| `GET /v1/status` | Public: per-subsystem status + per-source data freshness |

`/v1/status` powers the public status page. Data freshness is exposed publicly because a journalist
needs to know whether a number is current before publishing it.

---

## 10. Operational review

| Cadence | Activity |
|---|---|
| Daily | Alert triage; error-budget check |
| Weekly | Domain metrics review; classification-override sampling; ingestion health |
| Monthly | Cost review; SLO review; capacity forecast |
| Quarterly | Restore test; access review; dependency and vulnerability audit; transparency report |
| Pre-monsoon (April) | Load test at 20× baseline; pre-scale plan; emergency-mode drill |
