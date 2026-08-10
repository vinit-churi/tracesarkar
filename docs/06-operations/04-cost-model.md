# Cost model

**Purpose:** know the unit economics before building, so the architecture is shaped by them.

All figures are estimates for planning. Currency conversions at ₹88/USD. Verify current pricing
before committing to any number.

---

## 1. Fixed infrastructure (monthly)

Baseline sizing from [deployment](../03-architecture/09-deployment.md).

| Item | Spec | Est. USD/mo | Est. ₹/mo |
|---|---|---|---|
| 3 × app nodes | 4 vCPU / 8 GB | $60 | ₹5,300 |
| Postgres primary | 4 vCPU / 16 GB / 200 GB NVMe | $48 | ₹4,200 |
| Postgres replica | 2 vCPU / 8 GB | $24 | ₹2,100 |
| Redis | 2 vCPU / 4 GB | $12 | ₹1,050 |
| Object storage | 500 GB + egress | $15 | ₹1,300 |
| CDN | Public media | $10 | ₹900 |
| Backups (off-site) | 200 GB | $6 | ₹530 |
| Monitoring | Self-hosted on existing nodes | $0 | ₹0 |
| Domain + TLS | Let's Encrypt | $2 | ₹180 |
| **Total fixed** | | **~$177** | **~₹15,500** |

Comfortably within a modest grant or personal budget. **Fixed cost is not the problem.**

---

## 2. Variable cost — AI inference

This dominates, and it is the reason for every cost-control mechanism in the architecture.

### Per-report classification

| Component | Tokens | Notes |
|---|---|---|
| System prompt + taxonomy | ~900 | **Cached** — reads at ~0.1× after the first call |
| Images (2 @ 1080p) | ~2,400 | ~1,200 each; up to ~4,784 each at full 2576 px resolution |
| User text | ~100 | |
| Output | ~350 | Structured classification |

At `claude-opus-5` rates ($5/MTok input, $25/MTok output), with the system prompt cached:

```
uncached input  : 2,500 tok × $5/1M   = $0.0125
cached read     :   900 tok × $0.5/1M = $0.00045
output          :   350 tok × $25/1M  = $0.00875
                                        ─────────
per report                            ≈ $0.0217  ≈ ₹1.91
```

### Effect of the cost controls

| Control | Effect |
|---|---|
| **Dedup before classify** | During a burst at one location, 200 reports → ~1 classification. In steady state, assume ~25% of reports merge without classification |
| **Batch API for backfill** | 50% discount |
| **Image sizing** | Dropping to 1080p from full resolution roughly thirds the image-token cost; validate against the eval set first |
| **Prompt caching** | Already reflected above; verify `cache_read_input_tokens` is non-zero |

**Effective classification cost ≈ ₹1.45 per submitted report** after dedup savings.

### Monthly AI cost by scale

| Reports/month | Classification | Chatbot* | Extraction† | Share text | Total ₹/mo |
|---|---|---|---|---|---|
| 1,000 | ₹1,450 | ₹600 | ₹800 | ₹50 | **~₹2,900** |
| 10,000 | ₹14,500 | ₹4,000 | ₹1,500 | ₹500 | **~₹20,500** |
| 50,000 | ₹72,500 | ₹15,000 | ₹2,500 | ₹2,500 | **~₹92,500** |
| 200,000 (monsoon peak month) | ₹290,000 | ₹40,000 | ₹3,000 | ₹10,000 | **~₹343,000** |

\* Assumes deterministic short-circuits handle ~60% of chatbot intents with no model call.
† Contract extraction is a one-time-per-contract cost, not per-report; it plateaus.

**The monsoon peak row is the planning constraint.** At 200k reports in a month, inference alone is
~₹3.4 lakh. Mitigations, in order of preference:

1. Aggressive dedup (the biggest lever — a viral location should cost one classification)
2. Batch classification for non-urgent categories, with a longer displayed ETA
3. Hard daily budget cap with graceful degradation to queue-and-classify-later
4. Haiku tier for routine categories where the eval set shows no accuracy loss
5. Self-hosted VLM (revisit at this scale — the GPU cost becomes competitive)

---

## 3. Variable cost — storage

| Item | Per report | Notes |
|---|---|---|
| Archival originals (2 @ ~3 MB) | ~6 MB | Private, retained 3–7 years |
| Public derivatives (2 @ ~250 KB) | ~0.5 MB | CDN-served |
| Share image | ~150 KB | |
| **Total** | **~6.7 MB** | |

| Reports | Storage | Est. ₹/mo (at ~₹2/GB/mo) |
|---|---|---|
| 10,000 | 67 GB | ₹135 |
| 100,000 | 670 GB | ₹1,340 |
| 1,000,000 | 6.7 TB | ₹13,400 |

Cheap, and it compounds. Retention rules and derivative compression matter more than they appear.

---

## 4. Variable cost — WhatsApp

Per-message pricing with template approval requirements. Assume a per-conversation cost in the range
of ₹0.30–₹0.80 for utility templates (verify current Indian pricing).

| Notified users/month | Messages (3 high-value events each) | Est. ₹/mo |
|---|---|---|
| 1,000 | 3,000 | ₹900–₹2,400 |
| 10,000 | 30,000 | ₹9,000–₹24,000 |

**Consequence:** WhatsApp is reserved for the four events that need action (filed, SLA breach,
resolution claim, re-check request). Everything else is push and inbox. This is why the
[notification matrix](../03-architecture/08-realtime-and-notifications.md#event--channel-matrix) is
so restrictive.

---

## 5. Human cost (not in the infrastructure budget, but real)

| Role | Trigger | Effort |
|---|---|---|
| Moderation | ~2% of reports need human review | ~1 min each → 10k reports = ~3 hours/month |
| Dispute handling | Rare but time-consuming | ~2 hours per dispute |
| Ward onboarding | Per new ward | ~2–3 days (boundary verification, department mapping, PIO details, gazetteer) |
| Scraper maintenance | Per layout change | ~4 hours, roughly quarterly per source |
| Legal review | Per major feature touching names | Days |

**Ward onboarding is the real scaling constraint**, not compute. Nine corporations × N wards × 2–3
days is the actual expansion cost, and it does not parallelise easily.

---

## 6. Cost per confirmed fix

The efficiency number that matters. Illustrative at 10,000 reports/month:

```
total monthly cost          ≈ ₹15,500 (fixed) + ₹20,500 (AI) + ₹1,340 (storage) + ₹12,000 (WhatsApp)
                            ≈ ₹49,340

reports 10,000 → verified issues ~4,000 → routed ~2,000 → claimed resolved ~600
                             → citizen confirmed ~250

cost per confirmed fix      ≈ ₹197
```

For context: the Bombay High Court has fixed compensation for a pothole death at ₹6,00,000. If the
platform prevents one, the arithmetic is not close.

The funnel assumptions above are guesses and must be replaced with measured values as soon as v0.1
produces them.

---

## 7. Budget scenarios

| Scenario | Monthly | Viable on |
|---|---|---|
| **v0.1 pilot** (one ward, ~500 reports/mo) | ~₹18,000 | Personal budget |
| **v0.5** (Greater Mumbai, ~10k reports/mo) | ~₹50,000 | A modest grant |
| **v1.0** (MMR, ~50k reports/mo, monsoon peaks) | ~₹150,000 average, ~₹400,000 peak month | Grant + institutional API access |

---

## 8. Cost controls, in priority order

1. **Dedup before classify.** Biggest single lever, especially during bursts.
2. **Prompt caching.** Free; verify it is actually working via `cache_read_input_tokens`.
3. **Deterministic short-circuits** for common chatbot intents — no model call at all.
4. **Batch API** for all backfill and reprocessing.
5. **Hard daily budget cap** with graceful degradation. Never a 5xx to a citizen.
6. **Model tiering** — Haiku for high-volume simple schemas, validated against the eval set.
7. **Image sizing** — validate 1080p against full resolution before locking it in.
8. **Storage lifecycle rules** and derivative compression.
9. **WhatsApp restricted** to action-bearing events.
10. **Alert at 80% of the daily AI budget**, not at 100%.

---

## 9. What we will not do to cut cost

| Not doing | Why |
|---|---|
| Silently downgrade the model tier | A documented decision with an eval comparison, or not at all |
| Skip the archive | The archive is the evidentiary record; deleting it destroys the platform's central claim |
| Skip redaction | Compliance and ethics, not cost |
| Reject reports under load | Backpressure and queueing, never rejection |
| Sell data | See [GTM §6](../05-delivery/06-gtm-and-partnerships.md#6-sustainability) |
