# Watchdog — the terminal interface

A keyboard-driven terminal UI for journalists, activists, and researchers. Deliberately modelled on
a financial data terminal: dense, fast, no chrome, no marketing, no mouse.

**Who it is for:** Anjali (see [personas](01-personas-and-jtbd.md)). She has a deadline and a
sceptical editor. She needs a number nobody else has, and she needs to be able to defend it.

---

## 1. Why a TUI at all

| Reason | Detail |
|---|---|
| **Speed** | Keyboard-driven query and filter beats any web dashboard for a power user working a story |
| **Density** | A terminal shows 60 rows without scrolling; a web dashboard shows 8 with padding |
| **Low resource** | Runs over SSH on a newsroom laptop; works on a bad connection |
| **Scriptable** | Every view has a `--json` equivalent, so it composes with `jq`, `csvkit`, and the user's own tooling |
| **Signals seriousness** | It tells a journalist this is a data tool, not a campaign site |

It is a *complement* to the public web, not a replacement. Non-technical activists get the web
dashboard; the TUI is for people who live in a terminal.

---

## 2. Distribution

```
go install github.com/vinit-churi/tracesarkar/cmd/watchdog@latest
# or
brew install tracesarkar/tap/watchdog
```

Authenticates with an API key against the public API (see
[API design](../03-architecture/03-api-design.md)). Everything the TUI can see, the API can serve —
the TUI has no privileged backdoor.

---

## 3. Layout

```
┌ TraceSarkar Watchdog ─────────────────────── ward:R/S  ⟳ live ── 2026-08-10 14:22 IST ┐
│ FILTER  ward=R/S category=road_defect state=sla_breached since=2026-06-01              │
├───────────────────────────────────────────────────────────────────────────────────────┤
│  ID       AGE   CAT          SEV  STATE          AUTH  CONTRACT        DLP   CORR      │
│  8f2a1c   62d   pothole      HIG  sla_breached   BMC   WS/2023/R/117   IN     3        │
│  7b1e09   58d   pothole      CRI  sla_breached   BMC   WS/2023/R/117   IN     7        │
│  91cc2d   55d   open_manhole CRI  routed         BMC   —               —      2        │
│  4a02f8   51d   waterlog     HIG  reopened       BMC   SW/2022/D/044   IN     11       │
│  …                                                                                     │
├───────────────────────────────────────────────────────────────────────────────────────┤
│ 214 issues · 61 breached · 38 in-warranty · median age 44d · resolution 12%            │
├───────────────────────────────────────────────────────────────────────────────────────┤
│ / filter  : command  d detail  m map  t timeline  c contract  e export  a alert  ? help│
└───────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Views

| Key | View | Contents |
|---|---|---|
| `d` | **Issue detail** | Full record, timeline, photographs (rendered as sixel/kitty graphics where supported, otherwise a URL), jurisdiction basis, contract attribution with provenance |
| `t` | **Timeline** | Append-only event log for the issue, exportable verbatim |
| `c` | **Contract** | Tender → award → work order → completion → payment, with linked source documents and every issue attributed to it |
| `k` | **Contractor** | Scorecard, all contracts across MMR, all attributed defects, blacklisting history, corporate lineage |
| `w` | **Ward** | Report volume, category mix, resolution rate, claimed-vs-confirmed gap, top contractors by value, SLA breach rate |
| `m` | **Map** | Terminal map (braille/unicode density plot) — orientation, not cartography; `M` opens the web map |
| `q` | **Query** | Structured query builder over issues, contracts, contractors, authorities |
| `a` | **Alerts** | Saved queries that notify on new matches |
| `s` | **Sources** | Ingestion health: last successful run, records added, parse failures, staleness per source |

---

## 5. Query language

A small, learnable DSL — not SQL, not a form.

```
ward:R/S category:road_defect state:sla_breached since:2026-06-01
contract.dlp:active contractor:"* Infrastructure*" defects:>5
authority:MMRDA state:claimed_resolved confirmed:false
value:>10cr awarded:2024 completed:true attributed_defects:0
```

| Operator | Meaning |
|---|---|
| `field:value` | Exact match |
| `field:>N`, `field:<N` | Numeric comparison |
| `field:"glob*"` | Glob match on text |
| `since:`, `until:` | Date range |
| `field:true/false` | Boolean |
| `-field:value` | Negation |
| `a OR b` | Disjunction within a field |

Every query is shareable as a URL and as a saved alert. `:json` on any view emits the underlying API
response so a journalist can verify exactly what produced a number.

### Queries worth having as built-in presets

| Preset | Query |
|---|---|
| **In-warranty failures** | `contract.dlp:active defects:>0` |
| **Ghost completions** | `contract.completed:true attributed_defects:0 value:>1cr` — high-value contracts marked complete with zero citizen observations. Not proof of anything; a list of places to go look. |
| **Repeat offenders** | `contractor.scorecard:<D` |
| **Resolution theatre** | `state:claimed_resolved confirmed:false age:>30d` |
| **Monsoon surge** | `since:<monsoon_start> volume:>3sigma group:ward` |
| **Blacklist re-entry** | `contractor.directors_overlap_blacklisted:true` (v1+) |

---

## 6. Alerts

```
watchdog alert create \
  --name "R/S in-warranty failures" \
  --query 'ward:R/S contract.dlp:active defects:>0' \
  --notify email,webhook \
  --cadence realtime
```

Delivered by email, webhook, or a desktop notification while the TUI is running. This is the
"set a trap and go work on something else" feature, and it is what makes the tool sticky for a
newsroom.

---

## 7. Export and provenance

Every export carries provenance. Non-negotiable — an export without provenance is unusable to a
journalist.

```
watchdog export --query '...' --format csv     > issues.csv
watchdog export --query '...' --format geojson > issues.geojson
watchdog export --query '...' --format parquet > issues.parquet
watchdog export --issue 8f2a1c --format pdf    > evidence.pdf
```

Every export includes a sidecar manifest:

```json
{
  "generated_at": "2026-08-10T14:22:01Z",
  "query": "ward:R/S contract.dlp:active defects:>0",
  "api_version": "v1",
  "dataset_version": "2026-08-10T02:00:00Z",
  "record_count": 38,
  "sources": [
    {"id": "mahatenders", "retrieved": "2026-08-09T22:14:00Z", "sha256": "…"},
    {"id": "bmc_ward_boundaries", "version": "2023", "sha256": "…"}
  ],
  "caveats": [
    "Contract attribution confidence ≥ 0.75; 4 issues below threshold excluded",
    "DLP extracted from contract PDF by OCR; 2 records flagged for manual review"
  ]
}
```

The `caveats` array is generated, not boilerplate. It is what lets a reporter answer "how confident
are you in this?" precisely.

---

## 8. The "show, don't tell" rule

The TUI shows raw values, never editorialised summaries.

| Show | Don't show |
|---|---|
| `defects_in_dlp: 41` | "Poor performer" |
| `claimed_resolved: 88, confirmed: 12` | "Resolution theatre" |
| `match_confidence: 0.81` | (hiding the confidence) |
| `source: mahatenders, retrieved 2026-08-09` | (an unsourced fact) |
| `2 records flagged for manual review` | (silently dropping them) |

Where the platform computes a derived score (a contractor grade), the TUI shows the inputs and the
formula alongside it, always.

---

## 9. Document search (v0.7)

The archive of scraped tender PDFs, RTI replies, and official notices, made searchable with OCR.

```
watchdog docs search "defect liability" --ward R/S --year 2023
watchdog docs search "desilting" --contractor "*" --type work_order
```

Results show the matching page as an image with the hit highlighted, plus the archival hash so the
document can be independently verified. This turns a pile of scraped PDFs into the thing journalists
actually want: a searchable civic archive.

---

## 10. Implementation notes

| Choice | Rationale |
|---|---|
| Go + Bubble Tea | Same language as the backend; single static binary; good TUI ecosystem |
| API-only access | No direct database connection; the TUI is a client like any other, which keeps the API honest |
| Local cache | Recently viewed issues and contracts cached on disk for offline review of a story in progress |
| Config | `~/.config/tracesarkar/config.toml` — API key, default ward, colour scheme |
| Graphics | Sixel/kitty protocol where the terminal supports it; graceful degradation to URLs |
| Accessibility | Full functionality without colour; screen-reader-friendly plain output mode (`--plain`) |

---

## 11. Non-goals

- Not a moderation console. Moderators use a separate, access-controlled tool.
- Not a write interface for issues. Journalists read; they don't file on citizens' behalf.
- Not a replacement for the web dashboard for non-technical activists.
- Not a real-time trading-style tick feed. Civic data changes on the order of minutes, not
  milliseconds; the "live" mode polls, and says so.
