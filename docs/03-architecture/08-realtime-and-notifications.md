# Realtime and notifications

Two related systems: **live updates** (WebSocket fan-out to open clients) and **notifications**
(push, email, WhatsApp to closed clients). They share an event source and nothing else.

---

## 1. Event source

Every state change writes an `issue_events` row inside the same transaction as the change, then
publishes to Redis pub/sub via a transactional outbox. This guarantees that a published event
corresponds to committed state — a delivered "issue resolved" for a rollback that never committed
would be worse than a missed one.

```
tx: UPDATE issues … ; INSERT issue_events … ; INSERT outbox …   COMMIT
                                                    │
                              outbox relay ─────────┘
                                    │
                      ┌─────────────┴─────────────┐
                      ▼                           ▼
               Redis pub/sub              notification queue
               (realtime fan-out)         (push/email/WhatsApp)
```

---

## 2. Realtime — WebSocket

```
wss://api.tracesarkar.org/v1/stream?topics=ward:R/S,issue:01J9…&since=<cursor>
```

### Topics

| Topic | Content | Access |
|---|---|---|
| `issue:{id}` | All events for one issue | Public issues: anyone; own reports: owner |
| `ward:{code}` | New and changed issues in a ward | Public, coarsened |
| `report:{id}` | Enrichment progress for one's own report | Owner only |
| `feed:mmr` | Region-wide firehose, coarsened | API key required |
| `ingest` | Ingestion health | Admin |

### Connection model

| Property | Value |
|---|---|
| Concurrency | One goroutine per connection, owned by a hub per topic |
| Buffering | Bounded per-connection channel (256 events) |
| Slow consumers | Dropped with a close frame carrying a resume cursor; the client reconnects with `?since=` |
| Heartbeat | Server ping every 30 s; client must pong within 10 s |
| Auth | Bearer token at connect; re-validated on token expiry, connection closed if invalid |
| Limits | 5 concurrent connections per account; 50 topics per connection |
| Fallback | `GET /v1/events?since=<cursor>` long-poll when WebSocket is unavailable |

**Why one goroutine per connection is fine:** Go handles tens of thousands of them comfortably, and
`realtime` is deployed as its own service so a connection surge cannot starve the API. Connection
count is the scaling trigger.

### Event shape

```json
{
  "cursor": "01J9…",
  "topic": "issue:01J9…",
  "type": "issue.status_changed",
  "occurred_at": "2026-08-12T08:19:00+05:30",
  "data": {"from": "routed", "to": "sla_breached", "reason": "sla_elapsed"}
}
```

Cursors are monotonic and resumable. A client that reconnects with `since` receives everything it
missed, up to a retention window of 24 hours.

---

## 3. Notifications

### Channels

| Channel | Use | Notes |
|---|---|---|
| **Push** (FCM / APNs / Web Push) | Primary for app users | Cheap, immediate |
| **WhatsApp** | High-value events for users who opted in | Template-approved messages only; per-message cost — reserve for events that need action |
| **Email** | Digests, escalation deadlines, data-rights responses | Optional; many users will not provide one |
| **In-app inbox** | Everything, always | The durable record; other channels are best-effort |

### Event → channel matrix

| Event | Push | WhatsApp | Email | Inbox |
|---|---|---|---|---|
| Report classified / issue created | ✓ | — | — | ✓ |
| Issue verified | ✓ | — | — | ✓ |
| Filed, official reference captured | ✓ | ✓ | — | ✓ |
| **SLA breached** | ✓ | ✓ | — | ✓ |
| Authority claims resolved | ✓ | ✓ | — | ✓ |
| **Re-check request** (go confirm the fix) | ✓ | ✓ | — | ✓ |
| Escalation deadline approaching (RTI appeal 30 d, NGT 6 mo) | ✓ | ✓ | ✓ | ✓ |
| Corroboration received | ✓ (batched) | — | — | ✓ |
| Campaign milestone | ✓ (digest) | — | ✓ | ✓ |
| Dispute filed against your report | ✓ | — | ✓ | ✓ |
| Hazard alert near you | ✓ | ✓ | — | ✓ |

The two rows in bold are the ones that make the platform work. Everything else is noise management.

### The re-check request

The single highest-value notification in the system. When an authority marks an issue
`claimed_resolved`:

```
┌──────────────────────────────────────────────┐
│ BMC says the pothole on Link Road is fixed.  │
│                                              │
│ You're 300 m away. Two minutes to check?     │
│                                              │
│  [ It's fixed ✓ ]   [ Still broken ✗ ]       │
│                                              │
│ Either answer needs one photo.               │
└──────────────────────────────────────────────┘
```

Sent to the original reporter and, if they do not respond within 48 hours, to nearby high-trust
accounts as a verification mission. The claimed-vs-confirmed ratio this produces is the platform's
core public statistic.

---

## 4. Notification hygiene

A civic app that over-notifies gets muted, and a muted app cannot deliver an SLA breach alert.

| Control | Detail |
|---|---|
| **Frequency cap** | Max 3 push per user per day, except hazard alerts and deadlines |
| **Batching** | Corroborations, campaign updates, and digests are batched into one daily message |
| **Quiet hours** | 22:00–07:00 IST, except hazard alerts |
| **Per-category preferences** | Granular opt-outs, honoured immediately |
| **Deduplication** | One notification per (user, issue, event class) per window |
| **Actionability test** | If a notification does not enable an action, it goes to the inbox only |
| **Unsubscribe** | One tap, in every message; no dark patterns |

WhatsApp specifically: messages cost money and templates require approval. Use it only for events in
the bold rows above, and only for users who opted in explicitly. Budget it in the
[cost model](../06-operations/04-cost-model.md).

---

## 5. Delivery guarantees

| Property | Value |
|---|---|
| Ordering | Per-topic ordering guaranteed by cursor; cross-topic ordering not guaranteed |
| Delivery | At-least-once; clients deduplicate on `cursor` |
| Durability | The inbox is the durable record; push/WhatsApp/email are best-effort |
| Retry | Exponential backoff, 3 attempts, then dead-letter with alerting |
| Failure isolation | A failing channel (e.g. WhatsApp API down) does not block others |

---

## 6. Deadline scheduler

Escalation deadlines are the notifications with the highest consequence for being missed.

```
scheduler (every 15 min):
  SELECT * FROM escalations
   WHERE limitation_ends_on IS NOT NULL
     AND status IN ('draft','ready')
     AND limitation_ends_on <= now()::date + INTERVAL '30 days'

  → schedule reminders at T-30d, T-14d, T-7d, T-2d, T-1d
```

| Instrument | Clock | Reminders |
|---|---|---|
| RTI first appeal | 30 days from reply/expiry | T-14, T-7, T-2 |
| RTI second appeal | 90 days from first-appeal decision | T-30, T-14, T-7 |
| NGT §14 | 6 months from cause of action | T-60, T-30, T-14, T-7, T-2 |
| Lokayukta grievance | ~12 months from knowledge | T-60, T-30, T-7 |
| HC compensation claim | 6–8 week disbursal window | T-14 after filing, then weekly |

A missed limitation period kills a meritorious case. This scheduler is the most legally valuable
cron job in the system, and it needs a test that fails loudly if it stops running.

---

## 7. Failure behaviour

| Failure | Response |
|---|---|
| Redis down | Realtime degrades; clients fall back to polling. The outbox retains events; nothing is lost |
| WebSocket service down | Clients poll `GET /v1/events` |
| Push provider down | Retries, then inbox-only; a banner explains the delay |
| WhatsApp API down | Falls back to push; retries the WhatsApp send for 6 hours |
| Outbox relay stalled | Alert on outbox lag > 60 s — this is a paging alert, because it means state changes are invisible |

---

## 8. Metrics

| Metric | Alert |
|---|---|
| `outbox_lag_seconds` | Page above 60 s |
| `ws_connections_active` | Scale trigger |
| `ws_slow_consumer_drops` | Alert on a spike |
| `notification_send_failures{channel}` | Alert above 5% |
| `deadline_scheduler_last_run` | **Page** if older than 30 min |
| `recheck_response_rate` | Product metric, not an alert — but tracked prominently |
