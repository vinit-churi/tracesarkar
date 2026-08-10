# ADR 0002 — Go for the backend

**Status:** Accepted · **Date:** 2026-08-10

## Context

The backend must handle three distinct loads: bursty HTTP writes during monsoon spikes, many
concurrent WebSocket connections, and a scheduled ingestion fleet doing IO-bound scraping. It will be
operated by a very small team — possibly one person — on modest hardware.

## Decision

**Go**, for all server-side components: `api`, `realtime`, `worker`, `ingest`, `admin`, and the
`watchdog` TUI. One module, one language, one deployment story.

Rationale:

1. **Concurrency is the workload.** Goroutines and channels map directly onto per-connection
   WebSocket handling, bounded worker pools, and parallel enrichment stages.
2. **Single static binaries** make container images small and deployment trivial.
3. **Predictable resource usage** on small VPS instances; no runtime tuning rabbit holes.
4. **Strong stdlib** for HTTP, TLS, JSON, and templating — few dependencies to audit.
5. **The TUI ships in the same language** (Bubble Tea), so domain types are shared.
6. **The AI work is API calls, not model training.** There is no Python-only dependency in the
   critical path, and the official Anthropic Go SDK covers what we need.

## Alternatives

| Option | Why not |
|---|---|
| **Python (FastAPI)** | Best ML ecosystem, but we do not train models. Async ergonomics and deployment footprint are worse for the WebSocket and ingestion workloads. |
| **Node/TypeScript** | Shares language with the web client, but single-threaded CPU work (image handling, hashing) is awkward and memory behaviour under burst is less predictable. |
| **Rust** | Excellent fit technically; the contribution barrier and development speed are wrong for a project whose hard problems are domain problems, not performance problems. |
| **Elixir** | Genuinely great for the realtime piece; too small a hiring and contributor pool, and a second ecosystem for everything else. |

## Consequences

- **Cost:** no first-class local ML libraries. Redaction (face/plate blurring) needs either a Go
  binding to an ONNX runtime or a small sidecar. This is the one accepted seam. Recorded as a known
  gap in [ADR 0008](0008-redaction-approach.md).
- **Cost:** more boilerplate than Python for data wrangling in ad-hoc analysis. Analysis notebooks
  may use Python against the read replica; that is out-of-band and acceptable.
- **Benefit:** one toolchain, one test harness, one deployment pattern, one on-call skill set.
