# Screen specification

The authoritative, screen-by-screen definition of the TraceSarkar client surfaces. Where this file
and a mockup disagree, this file wins.

This document exists because a visually competent mockup can still violate the platform's hard
rules. It defines, per screen: purpose, entry condition, what is displayed, what is *never*
displayed, the single primary action, every state, and the exact copy patterns.

Read with [`03-user-flows.md`](03-user-flows.md) (the flows), [`02-feature-catalog.md`](02-feature-catalog.md)
(what exists at which milestone), and [`../00-overview/04-naming-and-brand.md`](../00-overview/04-naming-and-brand.md)
(voice). For generating imagery from this spec, use
[`12-ui-generation-guide.md`](12-ui-generation-guide.md).

---

## 1. Global rules

These apply to every screen. A screen that breaks one is wrong, however good it looks.

| # | Rule | Consequence in UI |
|---|---|---|
| G1 | **No instrument is ever filed by the platform.** | No button anywhere says "File this RTI", "Submit complaint for me", or "We will auto-file". Buttons say `Download draft`, `Open the RTI portal`, `Copy text`. The escalation ladder says "we will draft it". See [ADR 0007](../04-adr/0007-never-auto-file.md). |
| G2 | **No fact about a named party without a source and a retrieval time.** | Every contractor name, award value, DLP date, blacklisting entry renders with a `SourceLine`: `Source · <source name> · retrieved <date>`. If the source is absent, the field is not rendered — and the absence is stated: "Contract data not available for this ward yet." |
| G3 | **No assertion of wrongdoing.** | Status labels are procedural (`Deadline passed`), never evaluative (`BMC failed`). Forbidden words list in the brand doc is enforced in i18n review. |
| G4 | **No precise reporter location on a public surface.** | Public map pins and permalinks show a coarsened point (≥100 m jitter/snap-to-segment). Exact coordinates appear only to the reporter, on their own report, and internally. |
| G5 | **No raw faces or plates in any public derivative.** | Redaction runs before the public image exists. The UI shows a `Redacted` marker and a "why?" link. The unredacted original is never rendered on a public surface. |
| G6 | **No legal constant hard-coded in a screen.** | "48 hours", "₹30", "30 days", "6 months" are all rendered from `legal_constants` and always carry their citation inline or behind an info affordance. |
| G7 | **Nothing captured is ever lost.** | Every capture surface shows queue state honestly. There is no state in which a photo silently disappears. |
| G8 | **Model output is labelled and correctable.** | Anything classified automatically shows "Classified automatically" + a `Wrong? Change` affordance. Never "Verified by AI". |
| G9 | **One primary action per screen**, visually dominant. Everything else is secondary or tertiary. |
| G10 | **The platform fills in; the citizen corrects.** No screen asks the citizen for something the platform can derive. |
| G11 | **Colour never carries meaning alone.** Every status = icon + text label + colour. |
| G12 | **Honest state.** Low confidence, stale data, missing coverage, offline queue and partial failure are shown, never hidden. |
| G13 | **Language is switchable from every screen** (chrome-level control), and generated text is generated *per language*, never translated after the fact. |
| G14 | **`claimed_resolved` and `citizen_confirmed` never collapse into one label.** Anywhere they are counted, they are counted separately. |

---

## 2. Design tokens

Not a full design system. The decisions that carry meaning.

### 2.1 Palette

Party-neutral by requirement, not taste. Saffron, green and blue are all party-coded in Maharashtra
and none may be the dominant brand colour. (The August 2026 mockup used a dominant deep green — that
is a brand-neutrality defect, not a style preference.)

| Token | Value | Use |
|---|---|---|
| `ink` | `#141518` | Primary text, primary button fill |
| `ink-70` | `#4A4D55` | Secondary text |
| `ink-45` | `#7C8089` | Tertiary text, metadata, source lines |
| `paper` | `#FBFAF8` | App background (warm off-white, not pure white) |
| `surface` | `#FFFFFF` | Cards |
| `hairline` | `#E3E1DC` | 1 px borders — the primary separation device |
| `accent` | `#6B2C4F` | Plum. The single brand accent: primary CTA on dark surfaces, active nav, focus ring |
| `accent-wash` | `#F5EAF0` | Accent background tint |
| `alert` | `#A33A1F` | Deadline passed, reopened, error. Always with an icon + label |
| `alert-wash` | `#FBEDE8` | — |
| `caution` | `#8A6410` | Low confidence, stale data, disputed |
| `confirm` | `#1F6A4D` | **Only** for `citizen_confirmed` and successful capture. Never a surface fill |
| `evidence` | `#F2F0EB` | Fill for platform-statement blocks (see §3.1) |

Contrast: every text/background pair ≥ 4.5:1; large text ≥ 3:1. Enforced by a CI check on the token
file.

### 2.2 Type

| Token | Spec |
|---|---|
| Family | System UI stack first (`-apple-system`, `Roboto`), with `Noto Sans Devanagari` for Marathi/Hindi. No webfont on the capture path — it must render on 2G |
| `display` | 28/34, -0.02em — screen titles only |
| `title` | 20/26 semibold |
| `body` | 16/24 — the floor for anything a citizen must read |
| `label` | 14/20 medium |
| `meta` | 13/18 — source lines, timestamps |
| `mono` | 14/20 tabular — contract IDs, reference numbers, amounts |
| Numerals | Tabular lining in all data displays. Indian grouping (`₹4,11,00,000`) and lakh/crore in prose |

All layouts must survive 200% text zoom and Devanagari's taller line box (Devanagari strings run
~15–30% longer than English — never design a chip or button to fit English exactly).

### 2.3 Space, shape, target

4 pt base grid; spacing scale 4/8/12/16/24/32/48. Card radius 12; chip radius 8; button radius 10.
Touch target ≥ 48×48 dp with ≥ 8 dp between targets. Screen gutter 16.

Elevation: **borders, not shadows.** One hairline. Shadow is reserved for the camera shutter control
and modal sheets. This keeps the public pages cheap to render on low-end Android.

### 2.4 Photography

The photograph is the hero. Full-bleed where it is the subject; 4:3 or 16:9 thumbnails elsewhere;
never a decorative stock image anywhere in the product.

---

## 3. Component vocabulary

Components that carry legal or epistemic meaning. Reuse these; do not invent variants.

### 3.1 `PlatformStatement`

A block that is visually distinct from citizen content, because "who is speaking" is a legal
distinction. Fill `evidence`, 3 px `accent` left rule, small caps label `TRACESARKAR RECORD`.
Contains only facts the platform can source.

### 3.2 `SourceLine`

`meta` type, `ink-45`. Pattern:

```
Source · Mahatenders · retrieved 14 Jul 2026 · [view original]
```

Mandatory under every sourced fact group. `[view original]` opens the archived artefact (our copy),
not only the live URL — the live page may change or vanish.

### 3.3 `EvidenceCard`

The contract / warranty block. Fields render only when sourced (G2). Always carries `SourceLine`,
`[how was this matched?]` and `[dispute]`.

### 3.4 `MatchBasis`

Behind `[how was this matched?]`: the join that produced the attribution, in plain language —
matched geometry, buffer distance, contract ID, confidence band, and what would change the answer.
Never a raw model score alone.

### 3.5 `StatusChip`

Icon + label + colour. One chip per issue state; the label is the glossary term in plain language:

| State | Label (en) | Icon | Colour |
|---|---|---|---|
| `unverified` | Reported | dot-outline | `ink-45` |
| `verified` | Corroborated | check-double | `ink` |
| `routed` | Filed · ref captured | paper-plane | `ink` |
| `acknowledged` | Acknowledged | inbox | `ink` |
| `sla_breached` | Deadline passed | clock-alert | `alert` |
| `claimed_resolved` | **Authority says fixed** | flag | `caution` |
| `citizen_confirmed` | **Confirmed fixed** | camera-check | `confirm` |
| `reopened` | Reopened | rotate | `alert` |
| `escalated` | Escalated | stairs | `ink` |
| `closed` | Closed | archive | `ink-45` |

`claimed_resolved` and `citizen_confirmed` must never share a chip, a filter, or a count (G14).

### 3.6 `DeadlineBanner`

Statutory clock. Shows the constant, the remaining/elapsed time, and the citation:

```
⏱ 48 hours to attend  ·  31 h left
Bombay High Court, Oct 2025 · [read the order]
```

### 3.7 `ConfidenceRow`

`Classified automatically · pothole · [Wrong? Change]`. Confidence bands (`high` / `check this`) not
raw percentages on citizen surfaces; the numeric score is available in the technical view.

### 3.8 `ActionRow`

Icon + title + one-line consequence + availability state. Unavailable actions stay visible with the
reason ("Available after the 48-hour deadline passes"), never hidden.

### 3.9 `ProvenanceFooter`

On every public page: boundary vintage, contract retrieval date, licence, correction-log link.

### 3.10 `QueueBar`

Sticky, appears only when the outbox is non-empty: `3 reports waiting to send · [details]`.

### 3.11 `LanguageControl`

`EN · मराठी · हिंदी` in the app chrome and on every generated-text surface.

---

## 4. Screen inventory

| ID | Screen | Milestone | Notes |
|---|---|---|---|
| S01 | Home / launcher | v0.1 | Camera is the screen |
| S02 | Capture — frame 1 | v0.1 | GPS honesty |
| S03 | Capture — wide frame | v0.1 | Skippable |
| S04 | Enriching | v0.1 | Streamed stages |
| S05 | Disambiguation | v0.1 | Exactly one question |
| S06 | Confirm and submit | v0.1 | Correct-only, no form |
| S07 | Submitted | v0.1 | Share is primary |
| S08 | Issue detail (own) | v0.1 | The attribution reveal |
| S09 | Contract evidence | v0.1 | Sourced or absent |
| S10 | Share kit | v0.1 | Per-language generation |
| S11 | My reports | v0.1 | Two "fixed" states |
| S12 | Issue timeline | v0.1 | Append-only, hash-chained |
| S13 | Phone OTP | v0.1 | At submit, once |
| S14 | Public issue permalink | v0.1 | Web, no login, 2G |
| S15 | Ward page | v0.1 | Web, the claimed-vs-confirmed contrast |
| S16 | Rejection ("not a civic issue") | v0.1 | Specific, never generic |
| S17 | Offline outbox | v0.2 | Honest queue |
| S18 | Re-check prompt + capture | v0.2 | The highest-value interaction |
| S19 | Deadline passed → next actions | v0.2 | Highest leverage first |
| S20 | File with the authority | v0.2 | Adapter or draft + deep link |
| S21 | Reference number capture | v0.2 | The evidentiary spine |
| S22 | Nearby / map | v0.2 | Coarsened for others' reports |
| S23 | RTI draft builder | v0.3 | Review → edit → checklist → download |
| S24 | RTI filing record | v0.3 | Starts the 30-day clock |
| S25 | Escalation ladder | v0.4 | Drafts only |
| S26 | Ask a question (chatbot) | v0.3 | Validated answers only |
| S27 | Context (news / rain / repeats) | v0.4 | Sourced, linked out |
| S28 | Compensation claim assistant | v0.4 | Quiet tone. No share prompt |
| S29 | Contractor record | v0.5 | Threshold-gated; right of reply |
| S30 | Verification mission | v0.6 | Trust-gated |
| S31 | Campaign / ward digest | v0.5 | Collective weight |
| S32 | Settings — language, privacy, data | v0.1 | Language from day one |
| S33 | Correction log + takedown | v0.5 | Public, dated |
| S34 | Watchdog TUI | v0.6 | Terminal, separate spec |

---

## 5. Screens

### S01 — Home / launcher · v0.1

**Purpose:** get to the camera in one tap; show that the neighbourhood is already active.

**Layout (top → bottom)**

1. Chrome: wordmark, `LanguageControl`, notifications.
2. **Primary: full-width capture button**, ≥ 96 dp tall, `ink` fill, camera icon + `Report a problem`.
   It is the largest object on the screen.
3. **Nearby strip** — `3 open issues within 500 m` with three thumbnails. Tapping one opens it; this
   is the dedup and social-proof surface. If none: `No open issues nearby. That may mean nothing has
   been reported here yet.`
4. **Your contribution** (not "impact"): three counters — `Confirmed fixes you helped close`,
   `Re-checks done`, `Reports corroborated by others`. **Never** raw report count, views, or a rupee
   figure. Volume metrics reproduce the failure mode of every predecessor platform
   ([gamification](09-gamification.md) §1).
5. Recent reports list — thumbnail, category, locality, `StatusChip`, relative time.
6. Bottom nav: Home · Map · **Capture** · Reports · You.

**States:** first run (no reports → an explainer card, one screen, dismissible); offline (`QueueBar`
pinned under chrome); location denied (nearby strip replaced by `Turn on location to see issues
around you`, capture still works).

**Never:** a rupee "public money tracked" counter attributed to the user; a city-wide leaderboard;
a report-count badge.

---

### S02 — Capture, frame 1 · v0.1

**Purpose:** one photograph, honestly geotagged, in under five seconds.

**Layout:** full-bleed live viewfinder. Overlaid:

- Top-left `×` (exit), top-right flash toggle.
- **GPS pill, top-centre:** `GPS ✓ ±6 m` (`ink` on translucent). Degrades honestly:
  `±40 m` in caution colour; above 50 m → `Weak GPS. Step into the open, or [drop a pin]`.
- One coaching line, dismissible, max 8 words: `Capture the defect and some surroundings.`
- Shutter (72 dp), gallery affordance left, flip camera right.

**Rules:** no category picker, no description field, no stepper. There is **no** `Photo → Location →
Details → Review` wizard: a details step is a form, and the flow constraint is zero required text
input ([F1](03-user-flows.md#f1--first-report-the-critical-path)).

**States:** permission not granted (a single explainer + `Allow camera`); offline (`Offline — this
will send when you're back` chip; capture proceeds); low light (`Use flash?`).

---

### S03 — Capture, wide frame · v0.1

Same viewfinder. Copy: `One step back — a wide shot so the place is identifiable.` Persistent
`Skip` at equal weight to the shutter. Never blocks. Shown once per report.

---

### S04 — Enriching · v0.1

**Purpose:** show the work no other app does, while it happens.

Photo thumbnail pinned top. Then a vertical stage list, each line resolving independently as its
stage completes, with the *result* substituted into the line:

```
✓ Looked at the photo        Pothole · road defect
✓ Found the ward             BMC · R/S · Kandivali West
✓ Found the department       Roads & Traffic
◐ Checking contracts         …
○ Checking recent reports
```

- Streamed, not a percentage ring. A single spinner with a fake percentage hides which stage is slow
  and is dishonest when a stage fails.
- Any stage may resolve to a *negative* and still be a success:
  `— No contract data for this ward yet` in `ink-45`, with `[why?]`.
- If a stage fails: `— Couldn't check contracts right now. Your report is saved; we'll retry.`
- Target: first stage resolves < 1.5 s, all stages < 6 s. Beyond 10 s, offer
  `Submit now — we'll finish in the background`.

**Never:** block submission on enrichment. The report is already persisted (G7).

---

### S05 — Disambiguation · v0.1

Shown **only** when `requires_user_disambiguation`. Exactly one question, two or three
**photographic** options (thumbnails, not text lists — literacy, §5 of the accessibility doc).

```
Quick check — where exactly is this?
[ photo: ON the flyover ]   [ photo: UNDER the flyover ]
Different authorities look after each.
```

Never a dropdown of authorities. Never two questions. `Not sure` is always an option and routes to
the higher-authority default with `disputed` flagged.

---

### S06 — Confirm and submit · v0.1

Everything pre-filled; the only affordances are correct and submit.

- Photo strip (1–2 frames) with a `Redacted` marker if redaction fired + `[what was redacted?]`.
- `ConfidenceRow`: `Classified automatically · Pothole · [Wrong? Change]`.
- Location line: `SK Bole Marg, Dadar West · ±6 m · [move pin]`.
- Authority line: `BMC · R/S ward · Roads & Traffic · [not right?]`.
- `DeadlineBanner` preview: `If BMC does not attend within 48 hours, you'll be able to escalate.`
- Optional: `Add a note (optional)` — collapsed, never focused by default, never required.
- Primary: `Submit report`.

`[Wrong? Change]` opens the single disambiguation question or a 6-item category list — not a form.

---

### S07 — Submitted · v0.1

Confirmation, then **share as the primary action** — share is the distribution engine, not a
courtesy.

```
✓ Reported
BMC · R/S ward · Roads & Traffic · 48-hour clock started

[ ■ Share this ]            ← primary, full width
[ View issue ]  [ Report another ]
```

If the attribution found a contract, one line of the payload previews here:
`This stretch is under warranty until Nov 2028.` — with its `SourceLine`.

---

### S08 — Issue detail · v0.1

**The screen the product exists for.** Order matters: evidence → responsibility → money → clock →
action.

1. **Photo** (full-bleed, tappable to full screen; redaction marker).
2. Title line: `Pothole · SK Bole Marg, Dadar West` + `StatusChip`.
3. `ConfidenceRow`.
4. **`PlatformStatement` — responsibility:** `BMC · R/S ward · Roads & Traffic`, PIO/officer
   designation if sourced, `SourceLine` for the boundary vintage and the department mapping.
5. **`DeadlineBanner`** with citation (G6).
6. **`EvidenceCard` — the contract**, if and only if sourced:
   ```
   This stretch is under warranty
   Contract    WS/2023/ROAD/117
   Value       ₹4.11 crore
   Completed   28 Nov 2023
   Warranty    until 28 Nov 2028   (in defect liability period)
   Contractor  <name, only above the publication threshold>
   7 defects reported on this stretch since completion
   Source · Mahatenders · retrieved 14 Jul 2026 · [view archived copy]
   [ how was this matched? ]   [ dispute this ]
   ```
   Below the publication threshold, the contractor row is replaced with:
   `Contractor name withheld — awaiting a second independent source.` Absence is stated, never
   silent.
   If no contract matched: `No contract matched to this location yet. [what does that mean?]`
7. Map (small, own report → exact pin; anyone else's → coarsened, labelled `approximate`).
8. **Actions**, in leverage order, `ActionRow` each:
   `■ Share` (primary) · `File with BMC` · `Ask a question` · `Track`.
   Escalation actions appear here only once their precondition is met, with the reason shown until
   then.
9. Timeline preview → S12.

**Never on this screen:** a "Verified" badge on a contract; an award value without a source; the word
"escalate" as the primary action before the SLA has elapsed.

---

### S09 — Contract evidence (expanded) · v0.1

Full contract record. Every row is `field · value · source`. Sections: identification, parties,
money (award value, retention/withheld %, payments if sourced), dates (award, work order, start,
completion, DLP end), documents (work order, agreement, BOQ — each with size, retrieval date and
the archived copy), and `MatchBasis`.

Header badge, if used, states **what** was verified and by what:
`Cross-checked against Mahatenders and the BMC portal · 14 Jul 2026`. A bare `Verified` badge is
prohibited — it reads as the platform vouching for the contractor.

Footer: `[dispute this record]` → S33 correction workflow, and `[right of reply]` for the named
party.

---

### S10 — Share kit · v0.1

Preview of the annotated image + the generated text + `LanguageControl` (EN · मराठी · हिंदी),
generated per language from the fact set, never translated.

Text is **editable** — a user who rewrites it is likelier to send it. Channel order: WhatsApp, X,
Copy, More. WhatsApp first because that is where the behaviour is.

Burned into the image: coarsened location, date, contract ID if published, the mark and URL.
Never: an exact coordinate, an unredacted face or plate, an accusation.

---

### S11 — My reports · v0.1

Filters: `All · Open · Deadline passed · Authority says fixed · Confirmed fixed · Escalated`.
Six chips, because collapsing the middle two into "Resolved" destroys the platform's headline
statistic (G14).

Rows: thumbnail, category, locality, `StatusChip`, date, and — where a clock runs — the remaining
time. Rows with an action waiting (`re-check`, `enter your reference number`) carry a right-aligned
call-to-action, sorted to the top.

---

### S12 — Issue timeline · v0.1

Append-only event list, newest last: reported → corroborated → filed (ref) → acknowledged →
deadline passed → RTI drafted → RTI filed (reg. no.) → claimed resolved → re-check → confirmed /
reopened. Each event: actor (citizen / authority / platform), timestamp, and for platform events a
source or artefact link. Footer: `This timeline is hash-chained. [verify]`.

---

### S13 — Phone OTP · v0.1

Requested **at submit, once** — never as a wall before the first capture. One field, autofill-aware,
resend timer, and a plain statement of what the number is used for and what it is not used for
(no email required, no Aadhaar — [ADR 0011](../04-adr/0011-phone-only-identity.md)).

---

### S14 — Public issue permalink (web) · v0.1

Server-rendered, no login, must be usable on 2G. Same content as S08 with: coarsened location only,
no owner-only actions, an OG card for social unfurls, and a `ProvenanceFooter`. Total transfer
budget: ≤ 150 KB including the hero image at low bandwidth.

---

### S15 — Ward page (web) · v0.1

```
R/S — Kandivali West · BMC
Open 214 · Past deadline 61 · Authority says fixed 88 · Confirmed by citizens 12 · Median age 44 days
[ map ] [ list ] [ by category ] [ contractors ]
Boundaries v2023 · contracts retrieved 9 Aug 2026 · CC BY-SA 4.0 · [corrections]
```

The `88 → 12` contrast is the single most important number pair on the platform and is placed
adjacently by design. The contractors tab is threshold-gated and links to S29.

---

### S16 — Rejection · v0.1

When a report is classified as out of scope. Specific, never a generic error, and never accusatory:
`This looks like a private building repair. We only cover public infrastructure. [I disagree]`
The photo is retained for the user and deletable by them.

---

### S17 — Offline outbox · v0.2

`3 reports waiting to send`, each with thumbnail, capture time, capture-time GPS, and a per-item
state (`waiting for network`, `uploading 40%`, `failed — will retry`). Manual `Retry now`. Queue
survives restarts. Nothing is auto-deleted.

---

### S18 — Re-check · v0.2

The highest-value interaction and the only path to `citizen_confirmed`.

```
BMC says the pothole on Link Road is fixed.
You're 300 m away. Two minutes?
[ It's fixed ]   [ Still broken ]
Either way, one photograph please.
```

Both buttons lead to the camera; neither state can be set without a fresh photograph. Outcome moves
the issue to `citizen_confirmed` or `reopened`, and the celebratory tone is used **only** here for a
confirmed fix (brand doc, tone table).

---

### S19 — Deadline passed → next actions · v0.2

`DeadlineBanner` in `alert`, then `ActionRow`s in leverage order, the top one visually dominant:

1. `Ask for the records (RTI)` — `We'll draft it. You review and file it.`
2. `Share the missed deadline`
3. `Add to your ward campaign`
4. Greyed with reasons: `First appeal — available 30 days after an RTI is filed`.

Tone: direct, unsurprised, never gloating.

---

### S20 — File with the authority · v0.2

Two variants, chosen per platform terms:

- **Adapter available:** a summary of exactly what will be sent, to which channel, and
  `Send to BMC MARG` — the user's explicit, logged tap.
- **Adapter not permitted:** `Copy the complaint text` + `Open the BMC portal` + a 3-step
  instruction list. Degradation is stated plainly: `BMC's portal doesn't allow automated
  submission, so this is a copy-and-paste.`

Both end at S21.

---

### S21 — Reference number capture · v0.2

`Filed it? Paste the reference number.` One field, paste-friendly, with a photograph-of-the-receipt
alternative. Explains the stake plainly: `Without it we can't track the deadline or build the
appeal.` Gentle recurring nudge; never a blocking modal.

---

### S22 — Nearby / map · v0.2

Clusters with counts; own reports at exact position, others coarsened and labelled. Filters by
category, state, and age. A visible legend (icon + label + colour, G11). Bottom sheet lists the
issues in view.

---

### S23 — RTI draft builder · v0.3

Four steps, all reviewable, none automated:

1. **Review** — the drafted application with the questions the platform chose, each with a one-line
   rationale, each removable.
2. **Edit** — add questions; edit the facts paragraph. A plain-language summary sits above the
   formal text.
3. **Checklist** — `You will need:` the current fee (from `legal_constants`, with citation), an
   account on the state RTI portal, and the correct PIO address (sourced).
4. **Download / open** — `Download the draft (PDF)` + `Open the RTI portal`.

The screen's primary button is **never** `File this RTI` (G1). Nothing on this screen submits
anything to any government system.

---

### S24 — RTI filing record · v0.3

`Filed? Paste the registration number.` Starts the 30-day clock, which then drives the first-appeal
draft. Shows the clock and the appeal window with citations. Without this number the ladder stops —
so the screen states that consequence.

---

### S25 — Escalation ladder · v0.4

A vertical ladder of rungs (officer designation, body, statutory or administrative window, source
for the window). Each rung shows: reached / pending / available-when. The bottom card reads:

```
No response? We'll draft the next instrument for you to review and file.
Nothing is ever filed automatically. [why]
```

Never `we will auto-file`. Never a countdown implying platform action.

---

### S26 — Ask a question (chatbot) · v0.3

Entry from S08. Answers only from validated platform records, each with a citation chip; anything
the platform cannot source produces `I don't have that on record` plus the RTI route. Suggested
questions are typed tools, not free-form SQL ([chatbot](06-chatbot.md) §2).

---

### S27 — Context · v0.4

Tabs: `Nearby`, `Repeat failures`, `Rain`, `News`. Each item carries a source and a date; news shows
headline + publication + date + link out, never full text. Repeat failures render as
`7 defects at this asset since 28 Nov 2023` with the list.

---

### S28 — Compensation claim assistant · v0.4

For death or injury. Quiet, plain, helpful. **No gamification, no share prompt, no celebratory
language, no colour accents.** Shows the quantum and forum fixed by the Court with citation, the
evidence checklist, and a legal-aid handoff. Every figure sourced from `legal_constants`.

---

### S29 — Contractor record · v0.5

Threshold-gated. Facts and joins only: contracts held, values, DLP status, defect counts on covered
stretches, blacklisting entries **with the issuing notice**. Every row sourced and dated.

Mandatory on the page: `[right of reply]`, `[dispute a fact]`, a link to the correction log, and a
standing statement: `These are records, not allegations. TraceSarkar does not conclude that any
party has acted improperly.`

---

### S30 — Verification mission · v0.6

`5 issues near you are waiting to be re-checked.` Trust-gated, opt-in, distance-sorted, each with
the reason it needs a re-check. Accepting one opens the camera.

---

### S31 — Campaign / ward digest · v0.5

Collective view of an issue set with a shared share kit. Counters use corroborated and confirmed
figures only.

---

### S32 — Settings · v0.1

Language (EN · मराठी · हिंदी · ગુજરાતી as they land), notification channels, `My data` (export,
delete), privacy notice, `What we publish about your reports` (plain language, with an example),
and the grievance-officer contact required by the IT Rules.

---

### S33 — Correction log + takedown · v0.5

Public, chronological, permanent: what was wrong, when it was corrected, what the source now says.
Plus the takedown/right-of-reply intake and the grievance officer's name, address and response
window. This is the cheapest and strongest good-faith defence the platform has.

---

## 6. Copy patterns

| Situation | Pattern | Never |
|---|---|---|
| Deadline elapsed | `48 hours have passed. BMC has not recorded an action.` | `BMC failed you` |
| Contract found | `This stretch is under warranty until 28 Nov 2028.` | `The contractor must be held accountable` |
| No data | `We don't have contract data for this ward yet.` | silence, or an empty field |
| Model output | `Classified automatically — change it if this is wrong.` | `Verified by AI` |
| Escalation | `We'll draft it. You review and file.` | `We'll file it for you` |
| Confirmed fix | `Fixed, and confirmed by a photograph. Thank you.` | `You beat BMC` |
| Named party | `<fact> · Source · <source> · retrieved <date>` | `<fact>` alone |

Marathi and Hindi strings are generated per language from the fact set and reviewed by a native
speaker before release — never machine-translated from the English string, and never guessed by a
designer or an agent.

---

## 7. Mockup audit — August 2026 concept

Recorded so the same defects are not reintroduced.

| Mockup element | Defect | Rule | Fix |
|---|---|---|---|
| "No response? We will auto-file an RTI" | Asserts automated filing | G1 | S25 copy |
| `File this RTI` primary button | Same | G1 | S23 step 4 |
| No registration-number step | Escalation ladder cannot run | — | S24 |
| Contractor + award value, no source line | Unsourced fact about a named party | G2 | S08, S09 |
| `Verified` badge on contract | Reads as vouching | G3 | S09 header |
| No `how was this matched?` / `dispute` | Untraceable attribution | G2 | `EvidenceCard` |
| 4-step capture wizard with a Details step | Form on the capture path | G10 | S02–S06 |
| No GPS accuracy display | Dishonest state | G12 | S02 |
| Share demoted to a header icon | Distribution engine demoted | G9 | S07, S08 |
| Take Action lacks "file with the authority" | No official reference number | — | S20, S21 |
| `Resolved` filter chip | Collapses claimed vs confirmed | G14 | S11 |
| `23 Reports · 2.1k Views · ₹8.4L tracked` | Volume/vanity metrics | — | S01 |
| No language control anywhere | Marathi is non-negotiable | G13 | chrome |
| No re-check, offline, disambiguation or ward screens | Missing core flows | — | S18, S17, S05, S15 |
| Dominant deep-green brand colour | Party-coded palette | — | §2.1 |
| Street-precision pins for others' reports | Location exposure | G4 | S22 |
