# API design

Three surfaces, one implementation:

| Surface | Base | Audience |
|---|---|---|
| **Core API** | `/v1` | First-party clients (mobile, web, WhatsApp bot, Watchdog TUI) |
| **Open311** | `/open311/v2` | Standards-compatible third-party clients and any future municipal integration |
| **Public data API** | `/v1/public` | Researchers, newsrooms, embeds — read-only, coarsened, cached |

**Rule:** first-party clients use the same endpoints as everyone else. There is no privileged
backdoor. This keeps the API honest and makes the public surface a first-class product.

---

## 1. Conventions

| | |
|---|---|
| Transport | HTTPS only; HSTS; TLS 1.2+ |
| Format | JSON, UTF-8; `application/json` |
| Time | RFC 3339 with offset, e.g. `2026-08-10T14:22:01+05:30` |
| Coordinates | `{"lat": 19.2094, "lon": 72.8348}`; GeoJSON where a geometry is returned |
| Money | `{"amount_paise": 41100000, "display": "₹4.11 crore"}` |
| IDs | UUIDv7 strings; public issues also carry a short `slug` |
| Pagination | Keyset: `?after=<cursor>&limit=50`; never OFFSET |
| Errors | RFC 9457 problem details |
| Versioning | Path-versioned (`/v1`); additive changes only within a version |
| Idempotency | `Idempotency-Key` header on all POSTs |
| Rate limits | `RateLimit-*` response headers |

### Error shape

```json
{
  "type": "https://tracesarkar.org/errors/validation",
  "title": "Invalid report",
  "status": 422,
  "detail": "location.accuracy_m exceeds the maximum of 500",
  "instance": "/v1/reports",
  "errors": [{"field": "location.accuracy_m", "code": "out_of_range"}],
  "request_id": "01J9…"
}
```

---

## 2. Authentication

| Client | Method |
|---|---|
| Mobile / web | OAuth-style bearer token from phone-OTP login; short-lived access + refresh |
| WhatsApp bot | Server-to-server key; acts on behalf of a phone-verified account |
| Watchdog TUI / public API | API key (`Authorization: Bearer tsk_…`), scoped and rate-limited |
| Anonymous | Public read endpoints only, heavily rate-limited |

Scopes: `reports:write`, `issues:read`, `escalations:write`, `public:read`, `admin:*`.

---

## 3. Core endpoints

### Reports

```http
POST /v1/reports
Content-Type: multipart/form-data
Idempotency-Key: 01J9…

  meta = {
    "location": {"lat": 19.2094, "lon": 72.8348, "accuracy_m": 6.2, "heading_deg": 118},
    "captured_at": "2026-08-10T08:14:22+05:30",
    "description": "गेल्या तीन आठवड्यांपासून हा खड्डा आहे",
    "description_lang": "mr",
    "category_hint": null,
    "device_id": "…"
  }
  media[0] = <image/jpeg, role=wide>
  media[1] = <image/jpeg, role=close>

202 Accepted
{
  "report_id": "01J9…",
  "status": "pending",
  "estimated_ready_ms": 4000,
  "poll": "/v1/reports/01J9…",
  "stream": "wss://api.tracesarkar.org/v1/stream?report=01J9…"
}
```

`202`, not `201`: the report is durably stored, enrichment is asynchronous. The client renders an
optimistic card and fills it in as events arrive.

```http
GET    /v1/reports/{id}          # own report, full detail
GET    /v1/reports?mine=true     # own reports, keyset paginated
DELETE /v1/reports/{id}          # withdraw own report (soft; audited)
```

### Issues

```http
GET /v1/issues/{id_or_slug}
```

```jsonc
{
  "id": "01J9…",
  "slug": "8f2a1c",
  "status": "sla_breached",
  "category": "road_defect",
  "subcategory": "pothole",
  "severity": "high",
  "hazard_to_life": false,
  "location": {"lat": 19.2094, "lon": 72.8348},   // coarsened on public surfaces
  "jurisdiction": {
    "authority": {"code": "BMC", "name": "Brihanmumbai Municipal Corporation"},
    "ward": {"code": "R/S", "name": "Kandivali West"},
    "department": "Roads & Traffic",
    "confidence": 0.93,
    "basis": [ /* … */ ],
    "resolver_version": "2026.08.1"
  },
  "sla": {
    "hours": 48,
    "due_at": "2026-08-12T08:19:00+05:30",
    "breached": true,
    "source": {
      "citation": "High Court on its own motion v. State of Maharashtra, 13 Oct 2025",
      "url": "https://…"
    }
  },
  "attribution": {
    "published": true,
    "contract": {
      "id": "01J9…", "external_id": "WS/2023/ROAD/117",
      "title": "Improvement of Link Road…",
      "value": {"amount_paise": 41100000, "display": "₹4.11 crore"},
      "awarded_on": "2023-04-12", "completed_on": "2023-11-28",
      "dlp_ends_on": "2028-11-28", "in_dlp": true
    },
    "contractor": {"id": "01J9…", "name": "…", "scorecard_grade": "D"},
    "confidence": 0.81,
    "basis": [ /* every signal and weight */ ],
    "sources": [
      {"source_id": "mahatenders", "external_id": "…",
       "retrieved_at": "2026-07-14T…", "sha256": "…"}
    ],
    "dispute_url": "https://tracesarkar.org/dispute/01J9…"
  },
  "counts": {"reports": 3, "corroborations": 2},
  "timeline": [ /* append-only events */ ],
  "context": {"weather": {}, "news": [], "history": {}},
  "actions": [
    {"kind": "generate_rti", "available": true},
    {"kind": "share", "available": true},
    {"kind": "escalate_ngt", "available": false, "reason": "category not environmental"}
  ]
}
```

```http
GET  /v1/issues?ward=R/S&status=sla_breached&since=2026-06-01&after=<cursor>
GET  /v1/issues/near?lat=…&lon=…&radius_m=500
POST /v1/issues/{id}/corroborate          # "I see this too" — requires a fresh photo
POST /v1/issues/{id}/confirm_resolved     # requires a fresh photo; only citizens
POST /v1/issues/{id}/reopen               # requires a fresh photo
```

The `?` on corroborate/confirm is deliberate: **all three require a photograph**. A tap alone is not
evidence.

### Filing and escalation

```http
POST /v1/issues/{id}/file
{ "channel": "mybmc" }

202  { "filing_id": "…", "status": "submitting" }
```

```http
POST /v1/issues/{id}/escalations
{ "kind": "rti" }

201
{
  "escalation_id": "…",
  "kind": "rti",
  "status": "draft",
  "addressee": {"pio": "…", "designation": "…", "address": "…"},
  "content": { "questions": [ … ], "facts": "…", "annexures": [ … ] },
  "rendered_url": "https://…/escalations/…/download.pdf",
  "fee": {"amount_paise": 3000, "display": "₹30",
          "source": "Maharashtra RTI Rules (verify before filing)"},
  "limitation_ends_on": null,
  "disclaimer": "This is a generated draft, not legal advice. Review before filing."
}

PATCH /v1/escalations/{id}          # user edits the draft
POST  /v1/escalations/{id}/mark_filed
{ "external_ref": "MAHRT/A/2026/00931", "filed_at": "2026-08-14T11:40:00+05:30" }
```

**There is no `POST /escalations/{id}/submit`.** Filing a legal instrument is always a human act
outside the platform, or an explicit adapter run the user triggers. The platform records that it
happened.

### Share kit

```http
GET /v1/issues/{id}/share_kit?lang=mr&channel=whatsapp

200
{
  "text": "…",
  "image_url": "https://cdn…/share/8f2a1c-mr.png",
  "permalink": "https://tracesarkar.org/i/8f2a1c",
  "handles": ["@mybmc"],
  "hashtags": ["#MumbaiRoads", "#RSWard"]
}
```

### Chatbot

```http
POST /v1/chat
{ "message": "खड्डा कधी दुरुस्त होणार?", "scope": {"issue_id": "01J9…"}, "lang": "mr" }

200
{
  "answer": "…",
  "basis": [{"type": "issue", "id": "01J9…"}, {"type": "legal_constant", "key": "sla.road_defect.MH"}],
  "actions": [{"kind": "generate_rti", "label": "…"}]
}
```

---

## 4. Open311 compatibility

```http
GET  /open311/v2/services.json
GET  /open311/v2/services/{service_code}.json
POST /open311/v2/requests.json
GET  /open311/v2/requests.json
GET  /open311/v2/requests/{service_request_id}.json
GET  /open311/v2/tokens/{token}.json
```

`service_code` maps to `(category, subcategory)`. Statuses map to the Open311 `open`/`closed` pair,
with the richer TraceSarkar status exposed as an extension attribute:

```json
{
  "service_request_id": "8f2a1c",
  "status": "open",
  "extensions": {
    "tracesarkar_status": "sla_breached",
    "authority": "BMC",
    "ward": "R/S",
    "sla_due_at": "2026-08-12T08:19:00+05:30",
    "contract_attribution": { "external_id": "WS/2023/ROAD/117", "in_dlp": true }
  }
}
```

Mandatory per the spec: UTF-8 everywhere, ISO 8601 with timezone on every datetime.

**Why bother:** if any MMR corporation ever opens an endpoint, speaking the standard makes the ask
small. And it gives third-party clients a path in for free.

---

## 5. Public data API

Read-only, coarsened, aggressively cached.

```http
GET /v1/public/issues?ward=R/S&category=road_defect&since=2026-01-01&format=geojson
GET /v1/public/wards/{code}/stats?range=90d
GET /v1/public/contractors/{id}/scorecard
GET /v1/public/contracts?authority=BMC&year=2024&dlp_active=true
GET /v1/public/heatmap?bbox=…&resolution=8&format=geojson       # H3 aggregation
GET /v1/public/exports/issues-2026-08.parquet                    # monthly bulk dumps
```

Guarantees:

| Guarantee | Detail |
|---|---|
| **No PII** | No reporter identity, no exact coordinates, no un-redacted media |
| **Coarsened geometry** | Points snapped to a grid; heatmaps aggregated to H3 cells |
| **Provenance on every record** | `sources[]` with retrieval timestamps and hashes |
| **Stable schema** | Documented; breaking changes only at a version bump |
| **Cached** | CDN-fronted with explicit `Cache-Control`; ETag support |
| **Licensed** | CC BY-SA 4.0, attribution required |

Every response carries a `dataset_version` so an analysis is reproducible.

---

## 6. Realtime

```
wss://api.tracesarkar.org/v1/stream?topics=ward:R/S,issue:01J9…
```

Server → client events: `report.enriched`, `issue.created`, `issue.status_changed`,
`issue.corroborated`, `issue.attributed`, `filing.updated`.

| Property | Value |
|---|---|
| Auth | Bearer token in the connection request |
| Backpressure | Bounded per-connection buffer; slow consumers are dropped with a resume cursor |
| Heartbeat | 30 s ping; client reconnects with `?since=<cursor>` |
| Fallback | Long-poll `GET /v1/events?since=<cursor>` when WebSocket is unavailable |
| Scope | Public topics carry coarsened data; private topics require ownership |

---

## 7. Rate limits

| Scope | Limit |
|---|---|
| `POST /v1/reports` | 20/hour, 100/day per account (trust-scaled) |
| `POST /v1/chat` | 60/hour per account |
| `GET /v1/public/*` (keyed) | 1,000/hour |
| `GET /v1/public/*` (anonymous) | 60/hour per IP |
| Bulk export | 10/day per key |

`429` responses carry `Retry-After` and `RateLimit-Reset`.

---

## 8. Idempotency and consistency

- Every `POST` accepts `Idempotency-Key`; keys are retained 24 h with their response.
- Report submission is idempotent on `(account, device, captured_at, media sha256)` as a secondary
  guard against duplicate submissions from a retrying offline queue.
- Reads are eventually consistent against a replica; a read immediately after a write is routed to
  the primary using a session token.

---

## 9. Documentation and contract testing

- OpenAPI 3.1 spec generated from code annotations, published at `/v1/openapi.json`
- Contract tests run in CI against the spec; a breaking change fails the build
- A public changelog for the API, with deprecation windows measured in months, not weeks
- Example requests for every endpoint in the docs, runnable with `curl`

---

## 10. Open questions

- Do we expose a write endpoint for authorities to post proof-of-work photographs? *Leaning yes, via
  a scoped API key, because it improves the confirmation loop — but only as a claim that still
  requires citizen confirmation to close the issue.*
- Do we publish reporter pseudonymous handles on public issues at all? *Leaning no by default,
  opt-in per user.*
- GraphQL for the Watchdog TUI? *No. The query DSL compiles to parameterised SQL server-side; GraphQL
  would add a second query surface to secure for no gain.*
