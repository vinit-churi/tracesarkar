# The snap-to-action pipeline

This is the core loop. Everything else in the product is an elaboration of one of these seven stages.

**Contract with the user:** they point a camera at a problem and press one button. Every subsequent
decision is either made by the platform or asked as a single, concrete question.

```
 ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
 │ 0 CAPTURE│──▶│1 CLASSIFY│──▶│ 2 LOCATE │──▶│3 DEDUP + │
 │          │   │          │   │          │   │  VERIFY  │
 └──────────┘   └──────────┘   └──────────┘   └────┬─────┘
                                                    │
      ┌─────────────────────────────────────────────┘
      ▼
 ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
 │4 ATTRIBUTE│─▶│5 CONTEXT │──▶│  6 ACT   │──▶│ 7 TRACK  │
 │ (contract)│  │          │   │          │   │          │
 └──────────┘   └──────────┘   └──────────┘   └──────────┘
```

Target end-to-end latency for stages 0–5: **under 6 seconds on a mid-range Android on 4G**, with
stage 1 results shown optimistically at ~1.5 s and later stages streaming in.

---

## Stage 0 — Capture

**Input:** camera, GPS, optional voice note, optional text.

| Requirement | Detail |
|---|---|
| **Geotag at capture** | Read GPS at shutter time, not at submit time. Store accuracy radius, heading, and altitude. |
| **Timestamp integrity** | Record device time *and* server-received time. Divergence is a fraud signal. |
| **EXIF preservation** | Retain original EXIF in the archival copy; strip it from public derivatives. |
| **Offline queue** | Full capture works with no network. Queue persists across app restarts and syncs opportunistically. Show queue state honestly. |
| **Multi-shot** | Encourage 2–3 frames: a wide contextual shot (so the location is recognisable) and a close shot (so the defect is measurable). Prompt for the wide shot if only a close-up is given. |
| **Scale reference** | Optional prompt to place a common object (a shoe, a water bottle) for size estimation. |
| **Consent notice** | DPDP-compliant, itemised, shown at first capture — not buried in a ToS. |
| **On-device redaction** | Face and number-plate blurring applied before the public derivative is produced. |

**Anti-pattern to avoid:** a long form. Every field beyond the photo costs a measurable fraction of
submissions. Category, description, and authority are all *inferred*, then confirmed.

---

## Stage 1 — Classify

**Input:** photo(s) + optional text/voice.
**Output:** a structured, schema-guaranteed classification.

Implemented as a vision-language model call with **structured outputs** so the response is
guaranteed to parse. Model selection, cost envelope, and fallback behaviour are specified in
[AI pipeline](../03-architecture/04-ai-pipeline.md).

```jsonc
{
  "category": "road_defect",
  "subcategory": "pothole",
  "severity": "high",                    // low | medium | high | critical
  "hazard_to_life": true,                // triggers emergency diversion
  "estimated_dimensions": {              // null when not confidently estimable
    "length_cm": 120, "width_cm": 80, "depth_cm": 25, "confidence": 0.55
  },
  "surface_type": "asphalt",
  "water_present": true,
  "obstructs_traffic": true,
  "visible_landmarks": ["Reliance Fresh signboard", "bus stop"],
  "readable_text": ["Link Road", "MH 02 ..."],   // OCR'd; plates redacted before storage
  "people_present": false,
  "image_quality": "good",               // good | poor | unusable
  "is_civic_issue": true,                // false => reject politely
  "confidence": 0.91,
  "rationale": "Large depression in asphalt with standing water, exposed aggregate at edges"
}
```

### Category taxonomy (v1)

| Category | Subcategories |
|---|---|
| `road_defect` | pothole, crack, subsidence, missing manhole cover, open manhole, damaged speed breaker, utility-dig damage, missing/faded markings |
| `waste` | uncollected garbage, illegal dumping, debris (C&D waste), overflowing bin, dead animal, burning waste |
| `water_drainage` | waterlogging, burst pipeline, leakage, blocked drain, silted nallah, sewage overflow |
| `street_furniture` | streetlight out, damaged railing, broken footpath, missing bollard, damaged signage |
| `structural` | cracked bridge/FOB, distressed building, unsafe scaffolding, collapsed wall |
| `environment` | mangrove destruction, creek dumping, effluent discharge, tree felling, air pollution source, encroachment on water body |
| `transit` | station infrastructure, FOB, escalator/lift out, unsanitary premises |
| `illegal_construction` | unauthorised structure, encroachment, hawking obstruction |
| `other` | escalate to human triage |

`hazard_to_life = true` (open manhole, live wire, imminent collapse) bypasses normal queuing and
routes to the emergency channel — a direct adaptation of FixMyStreet Pro's emergency diversion.

### Rejection path

`is_civic_issue = false` returns a polite, specific message ("this looks like a private property
dispute — here's who handles that"), never a generic error. Rejections are logged and sampled for
review; a high rejection rate in a ward is usually a UX failure, not a user failure.

---

## Stage 2 — Locate

**Input:** GPS point + accuracy radius + classification category + timestamp.
**Output:** a jurisdiction resolution with confidence and provenance.

See [jurisdiction engine](../03-architecture/05-jurisdiction-engine.md) for the algorithm. Summary:

1. PostGIS containment against time-versioned ward and authority polygons
2. Road-ownership layer override (state highway inside city limits → PWD/MSRDC, not the corporation)
3. Elevated-structure proximity check (flyover deck vs ground below)
4. Railway premises geofence (→ Transit Mode)
5. Category-conditional department mapping (nallah → SWD; mangrove → Forest/MCZMA/MPCB)
6. Active-works overlay (if a live contract's site polygon contains the point, name it)
7. Confidence scoring; below threshold → **one** disambiguating question in the UI

The one-question rule matters. "Is this on the flyover or the road underneath it?" with two photo
thumbnails is answerable in a second. A dropdown of eleven authorities is not.

---

## Stage 3 — Deduplicate and verify

### Deduplication

Candidate matching within an H3 cell neighbourhood and a category-specific radius (default 40 m for
road defects, 25 m for street furniture, 100 m for waste), plus a recency window. Candidates are
scored on:

- Spatial distance (weighted by GPS accuracy)
- Category and subcategory agreement
- Visual similarity (perceptual hash + embedding cosine similarity on the close-up frame)
- Textual similarity of the description

Above the merge threshold → the report attaches to the existing issue, incrementing its corroboration
count. Between thresholds → surfaced to the user: "Is this the same as this one?" with a thumbnail.
Below → a new issue.

**Merging is the accountability multiplier.** Three separate tickets are three ignorable
inconveniences. One issue with three independent corroborations from three accounts is evidence.

### Verification

Issue moves `unverified → verified` when the trust threshold is met. The threshold is
category-dependent:

| Category | Verification requirement |
|---|---|
| `hazard_to_life = true` | **Fast path:** publish immediately as `unverified-urgent`, route to the emergency channel, and require corroboration within a short window for public display. A delayed open-manhole alert is worse than a false one. |
| `road_defect`, `waste`, `street_furniture` | One report from a trusted account, or two independent reports from distinct accounts and devices within the dedup radius |
| `environment`, `illegal_construction`, anything naming a private party | Higher bar: independent corroboration from accounts with no shared device/IP/social graph, plus moderator review before public display |

The "three independent users within a timeframe" idea from the original concept is right in spirit
but must not gate life-safety alerts. See [trust and anti-abuse](08-trust-and-antiabuse.md).

---

## Stage 4 — Attribute (the "follow the money" join)

**Input:** resolved location + asset + category + date.
**Output:** matched contract(s) with confidence, or an honest "no matching contract found".

See [tender engine](../03-architecture/06-tender-engine.md) for the matching algorithm.

What the citizen sees when a match succeeds:

```
┌─────────────────────────────────────────────────────────────┐
│  This road was last worked on under                          │
│                                                              │
│  Contract  WS/2023/ROAD/117                                  │
│  Contractor  ███████ Infrastructure Pvt Ltd                  │
│  Awarded  ₹4.11 crore  ·  12 Apr 2023                        │
│  Completed  28 Nov 2023                                      │
│  Defect Liability Period ends  28 Nov 2028   ⟵ IN WARRANTY   │
│                                                              │
│  Defects reported on this stretch since completion:  7       │
│  This contractor's MMR scorecard:  D  (41 in-warranty defects)│
│                                                              │
│  Source: Mahatenders tender ID …  ·  retrieved 2026-07-14    │
│  [ view documents ]  [ how was this matched? ]  [ dispute ]  │
└─────────────────────────────────────────────────────────────┘
```

**Three non-negotiables in this display:**

1. **Every fact carries a source and a retrieval date.** Without it, this is defamation exposure.
2. **"How was this matched?"** opens the provenance chain. Users must be able to audit the join.
3. **"Dispute"** is a real, staffed channel. See [moderation policy](../06-operations/02-moderation-policy.md).

When no match is found, say so plainly and offer to generate an RTI that asks for the contract
details. A confident wrong attribution is far more damaging than an honest gap.

---

## Stage 5 — Contextualise

Attach, where available:

| Context | Source |
|---|---|
| Prior reports at this location and their outcomes | Internal |
| Rainfall in the last 24/72 h at the nearest gauge | IITM MESONET / IMD |
| Waterlogging status | mumbaiflood.in |
| Air quality (for pollution categories) | SAFAR / CPCB |
| Recent news mentioning this location or asset | News ingestion pipeline |
| Active works notices (mega-block, road closure, utility dig) | Corporation and railway press notes |
| Ward-level issue density and resolution rate | Internal |

Context does two jobs: it prevents duplicate panic-reporting during a known event ("this ward is
under a declared mega-block until Sunday"), and it strengthens the escalation ("this stretch has
flooded on 6 of the last 9 heavy-rain days").

---

## Stage 6 — Act

The user is presented with a ranked action set. Every action is one tap. Nothing is performed
without an explicit tap.

| Action | What happens |
|---|---|
| **File with the authority** | The filing adapter for `(authority, category)` runs — API where one exists, form automation where permitted, otherwise a pre-filled draft plus a deep link and instructions. The official reference number is captured back into the issue. |
| **Share** | Generates the share kit: a tagged tweet/X post, a WhatsApp card, an image with the annotated photo + contract facts, and a permalink. See [share kit](05-share-kit.md). |
| **Generate RTI** | Pre-filled RTI addressed to the correct PIO with category-specific questions. Draft only — the citizen reviews, edits, and files. |
| **Add to a campaign** | Attaches the issue to a ward-level or corridor-level campaign for collective weight. |
| **Escalate** | Available when preconditions are met (SLA breached, RTI unanswered, environmental harm within limitation). Produces the appropriate instrument from the [decision table](../01-research/04-legal-framework.md#9-instrument-decision-table). |
| **Claim compensation** | For injury/death from a road defect in Maharashtra: assembles a claim under the Bombay HC 2025 framework. |
| **Ask** | Opens the chatbot scoped to this issue. See [chatbot](06-chatbot.md). |

**Ordering rule:** the highest-leverage action available given the issue state is shown first and
largest. For a fresh pothole that is a share + file. For a 60-day-old unresolved issue it is
"generate RTI". For an issue where the RTI came back showing full payment for undelivered work, it
is "Lokayukta".

---

## Stage 7 — Track

Every issue carries a public timeline. The timeline is the product's memory and its evidentiary
spine.

```
2026-06-11 08:14  reported            by @user  (photo, GPS ±6 m)
2026-06-11 08:14  classified          pothole / high / hazard=false
2026-06-11 08:14  located             BMC · R/S ward · Roads & Traffic (conf 0.93)
2026-06-11 08:15  attributed          contract WS/2023/ROAD/117 · in DLP (conf 0.81)
2026-06-11 08:16  corroborated        2nd independent report, 11 m away
2026-06-11 08:16  verified
2026-06-11 08:19  filed               MyBMC MARG · ticket MARG-2026-0611-88213
2026-06-13 08:19  sla_breached        48 h elapsed (Bombay HC direction, 13 Oct 2025)
2026-06-13 09:02  rti_generated       draft ready
2026-06-14 11:40  rti_filed           MAHRT/A/2026/00931
2026-07-14 11:40  rti_overdue         30 days elapsed → first appeal available
```

Timeline entries are append-only and hash-chained. Nothing is deleted; corrections are additional
entries. This is what makes the record usable in a filing.

### Notifications

| Event | Channel |
|---|---|
| Issue verified | Push |
| Filed / reference captured | Push |
| SLA breach | Push + in-app action prompt |
| Authority claims resolved | Push + **"go back and check"** prompt with a re-photograph flow |
| Limitation period approaching (NGT 6 months, RTI appeal 30 days) | Push + email |
| A campaign the user joined hits a milestone | Push (digest) |

**The re-photograph prompt is the highest-value notification in the system.** It converts
`claimed_resolved` into `citizen_confirmed` or `reopened`, and the ratio between those two is the
single number that makes the platform matter.

---

## Failure modes and their handling

| Failure | Handling |
|---|---|
| Classification confidence < threshold | Ask the user to pick from the top 3 categories; log for model evaluation |
| GPS accuracy > 50 m | Show a map pin and ask the user to drag it; store both raw and corrected positions |
| No jurisdiction match | Route to human triage queue; tell the user honestly and give the manual channel list |
| No contract match | Say so; offer an RTI seeking the contract details |
| Filing adapter fails | Fall back to "generate draft + deep link"; never lose the report |
| VLM API unavailable | Queue for later classification; the report is stored and acknowledged regardless |
| Duplicate storm (one viral location) | Auto-merge aggressively, show the aggregate count, and stop asking for corroboration |

**Invariant:** a captured report is never lost, whatever fails downstream. Storage of the raw capture
happens before any enrichment is attempted.
