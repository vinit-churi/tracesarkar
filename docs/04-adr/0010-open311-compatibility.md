# ADR 0010 — Expose an Open311 GeoReport v2 surface

**Status:** Accepted · **Date:** 2026-08-10

## Context

No MMR municipal corporation currently exposes a public API for civic complaints. Our filing
adapters will therefore be form automation or "generate a draft and hand it to the citizen" for the
foreseeable future.

Meanwhile, Open311 GeoReport v2 is the established international standard for exactly this domain,
implemented by Chicago, Toronto, San Francisco, Boston, Washington DC, Helsinki, and Bonn among
others. It defines six methods, mandates UTF-8, and requires ISO 8601 timestamps with timezone.

## Decision

Expose an Open311 GeoReport v2-compatible surface at `/open311/v2`, covering:

```
GET  /services.json
GET  /services/{service_code}.json
POST /requests.json
GET  /requests.json
GET  /requests/{service_request_id}.json
GET  /tokens/{token}.json
```

TraceSarkar's richer model is exposed through namespaced extension attributes
(`tracesarkar_status`, `authority`, `ward`, `sla_due_at`, `contract_attribution`), leaving the core
fields spec-compliant.

## Rationale

1. **It makes the ask small.** If any corporation ever considers opening an endpoint, "implement the
   same standard Chicago and Helsinki use, and we'll consume it" is a far easier conversation than
   "implement our proprietary API".
2. **It gives us third-party clients for free.** An existing ecosystem of Open311 apps and libraries
   can point at TraceSarkar with no work from us.
3. **The schema is good.** Service definitions, service requests, attributes, and statuses map
   cleanly onto our domain.
4. **It signals interoperability rather than lock-in**, which matters for a project asking public
   bodies to trust it.

## Alternatives

| Option | Why not |
|---|---|
| **Our own API only** | Loses the standards argument and the client ecosystem for a marginal saving. |
| **Open311 as the *primary* internal model** | Too thin — it has no concept of an authority hierarchy, a contract, a defect liability period, or an escalation. It is a compatibility surface, not our domain model. |
| **Wait until a corporation asks** | The point of implementing it early is to make asking easy. |

## Consequences

- A second API surface to maintain and test. Mitigated: it is a thin translation layer over the same
  read models, with contract tests against the spec.
- Some Open311 semantics are a poor fit (its binary `open`/`closed` status vs our ten-state
  lifecycle). Handled by extensions, and documented.
- We should also be able to *consume* Open311 if a corporation ever publishes one — the client side
  is a small addition to the filing-adapter framework, and worth building speculatively only once a
  real endpoint exists.
