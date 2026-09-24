# Source archive — BMC hosts are geo-restricted

Probed 24 September 2026, from two vantage points: a residential connection in Mumbai, and GitHub
Actions runners (Microsoft Azure, `canadaeast` and a US runner reporting Virginia).

This was found by running the collector in CI, not by reading documentation.

---

## What was observed

| Host | Port | From India | From a US/Canada cloud runner |
|---|---|---|---|
| `roads.mcgm.gov.in` | 443 | ✅ 200 | ⛔ no connection |
| `roads.mcgm.gov.in` | 3000 (the works API) | ✅ 200 | ⛔ `dial tcp 161.248.105.55:3000: i/o timeout` |
| `portal.mcgm.gov.in` | 443 | ✅ 302 | ⛔ no connection |
| `swd.mcgm.gov.in` | 443 (drain desilting API) | ✅ 200 | ✅ 200 |
| `archive.org` (GR mirror) | 443 | ✅ | ✅ |

The failure is a connection timeout, not an HTTP error: the refusal happens before TLS, so it is
network-level filtering rather than an application response. Three retries with backoff made no
difference, and the same binary succeeded from India minutes earlier and later.

**Evidence:** GitHub Actions runs `35963304983` (the collector) and `35963831983` (a throwaway
`curl` probe, since removed) in `vinit-churi/tracesarkar`.

## What it means

Most of MCGM's estate answers only Indian traffic; `swd.mcgm.gov.in` does not share that
restriction. Any collector for the roads API or the BMC portal must run from an Indian IP address.

**Not yet distinguished:** whether the filter is by geography (non-Indian IPs refused) or by network
type (datacenter and cloud ranges refused, residential allowed). Only two vantage points were
tested, and both were at the extremes: an Indian home connection, and foreign cloud. The first boot
of a VM inside an Indian datacenter answers it, and that test is scheduled as part of the work in
[ADR 0015](../../04-adr/0015-indian-egress-for-collection.md).

## Consequences recorded elsewhere

- Collection split by reachability: [ADR 0015](../../04-adr/0015-indian-egress-for-collection.md).
- `bmc_roads_api` and `bmc_tenders` carry a `network` note in
  [`data/sources.yaml`](../../../data/sources.yaml).
- A source that is fetchable from your desk may still be unfetchable from where the code runs. Worth
  checking before any future source is scheduled.
