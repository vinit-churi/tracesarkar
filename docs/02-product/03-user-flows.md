# User flows

Screen-by-screen for the flows that matter. Anything not described here is out of MVP scope.

---

## F1 — First report (the critical path)

**Constraint:** from app open to "submitted" in **under 30 seconds**, with **zero** required text
input.

```
┌─ 1. OPEN ────────────┐   ┌─ 2. CAPTURE ─────────┐   ┌─ 3. SECOND SHOT ─────┐
│                      │   │  [ live camera ]     │   │  [ live camera ]     │
│   [ big camera CTA ] │──▶│                      │──▶│  "Step back — one    │
│                      │   │        ( ● )         │   │   wide shot so the   │
│  Nearby: 3 open      │   │                      │   │   place is clear"    │
│  issues              │   │  GPS ✓ ±6m           │   │        ( ● )  [skip] │
└──────────────────────┘   └──────────────────────┘   └──────────────────────┘
                                                                  │
┌─ 6. DONE ────────────┐   ┌─ 5. CONFIRM ─────────┐   ┌─ 4. ENRICHING ───────┐
│  ✓ Reported          │   │  Pothole · High      │   │  ⟳ Looking at photo  │
│                      │   │  BMC · R/S ward      │   │  ⟳ Finding the ward  │
│  [ SHARE THIS ]      │◀──│  Roads & Traffic     │◀──│  ⟳ Checking contracts│
│  [ view issue ]      │   │                      │   │                      │
│                      │   │  Wrong? [ change ]   │   │  (streams in ~4s)    │
└──────────────────────┘   │  [ SUBMIT ]          │   └──────────────────────┘
                           └──────────────────────┘
```

**Design notes**

| Screen | Note |
|---|---|
| 1 | Camera is the primary and largest action. "Nearby open issues" creates immediate social proof and enables dedup before capture |
| 2 | GPS accuracy is shown honestly. Below 50 m accuracy, proceed; above, prompt to step outside or drop a pin |
| 3 | Skippable. The wide shot roughly doubles evidentiary value, so ask — once, briefly — but never block |
| 4 | Progressive disclosure. Each line resolves as its stage completes. This is where the platform demonstrates it is doing something no other app does |
| 5 | Everything is pre-filled. The only affordance is to correct. `[ change ]` opens the disambiguation question, not a form |
| 6 | **Share is the primary action, not a secondary one.** The share is the distribution engine |

**Anti-patterns explicitly avoided:** a category dropdown before the photo; a required description
field; a ward selector; an account wall before the first report (auth is requested at submit, once).

---

## F2 — Disambiguation (when confidence is low)

Triggered when `requires_user_disambiguation = true`. Exactly one question, always answerable in a
second.

```
┌────────────────────────────────────────────┐
│  Quick check — where exactly is this?      │
│                                            │
│  ┌──────────────┐    ┌──────────────┐      │
│  │ [thumbnail]  │    │ [thumbnail]  │      │
│  │ ON the       │    │ UNDERNEATH   │      │
│  │ flyover      │    │ the flyover  │      │
│  └──────────────┘    └──────────────┘      │
│                                            │
│  Different authorities look after each.    │
└────────────────────────────────────────────┘
```

Never a dropdown of eleven authorities. Never more than one question per report.

---

## F3 — The attribution reveal

The moment the product justifies itself.

```
┌─────────────────────────────────────────────────────────┐
│  Pothole · Link Road, Kandivali West                     │
│  ─────────────────────────────────────────────────────   │
│  BMC · R/S ward · Roads & Traffic                        │
│  Must be attended within 48 hours (Bombay HC, Oct 2025)  │
│                                                          │
│  ┌───────────────────────────────────────────────────┐   │
│  │  This stretch is under warranty                   │   │
│  │                                                   │   │
│  │  Contract   WS/2023/ROAD/117                      │   │
│  │  Value      ₹4.11 crore                           │   │
│  │  Completed  28 Nov 2023                           │   │
│  │  Warranty   until 28 Nov 2028                     │   │
│  │                                                   │   │
│  │  7 defects reported here since completion          │   │
│  │                                                   │   │
│  │  Source: Mahatenders · retrieved 14 Jul 2026      │   │
│  │  [ how was this matched? ]     [ dispute ]        │   │
│  └───────────────────────────────────────────────────┘   │
│                                                          │
│  [ ■ SHARE ]   [ file with BMC ]   [ ask a question ]    │
└─────────────────────────────────────────────────────────┘
```

The contractor name appears only above the publication threshold. Below it, the card shows the
contract without the name and says so.

---

## F4 — Share

```
┌─────────────────────────────────────────┐
│  Share this                             │
│                                         │
│  [ preview of the annotated image ]     │
│                                         │
│  Language:  [English] [मराठी] [हिंदी]     │
│                                         │
│  ┌─────────────────────────────────┐    │
│  │ Pothole on Link Road, Kandivali │    │
│  │ West. @mybmc — R/S ward.        │    │
│  │ This stretch is under warranty  │    │
│  │ from contract WS/2023/ROAD/117  │    │
│  │ (₹4.11 cr) until Nov 2028.      │    │
│  │ Bombay HC requires action in    │    │
│  │ 48 hours. tracesarkar.org/i/…   │    │
│  └─────────────────────────────────┘    │
│  [ edit ]                               │
│                                         │
│  [ WhatsApp ]  [ X ]  [ copy ]  [ more ]│
└─────────────────────────────────────────┘
```

WhatsApp is first because it is where the behaviour is. The text is editable — a user who rewrites
it is more likely to send it.

---

## F5 — The re-check (highest-value interaction)

Triggered when an authority marks an issue `claimed_resolved`.

```
┌────────────────────────────────────────┐
│  BMC says the pothole on Link Road     │
│  is fixed.                             │
│                                        │
│  You're 300 m away. Two minutes?       │
│                                        │
│  [ It's fixed ✓ ]   [ Still broken ✗ ] │
│                                        │
│  Either way, one photo please.         │
└────────────────────────────────────────┘
                 │
                 ▼
        camera → confirm → issue moves to
        citizen_confirmed OR reopened
```

If the original reporter does not respond within 48 hours, the same prompt goes to nearby high-trust
accounts as a **verification mission**. The claimed-vs-confirmed ratio this produces is the
platform's headline public statistic.

---

## F6 — SLA breach and escalation

```
┌────────────────────────────────────────┐
│  ⚠ 48 hours have passed                │
│                                        │
│  BMC has not attended to this. The     │
│  Bombay High Court requires action     │
│  within 48 hours of a report.          │
│                                        │
│  What you can do now:                  │
│                                        │
│  ■ Ask for the records (RTI)           │
│    We'll draft it — you review and file│
│                                        │
│  □ Share the breach                    │
│  □ Add to your ward campaign           │
└────────────────────────────────────────┘
```

The highest-leverage available action is first and largest. Actions become available as preconditions
are met; unavailable ones show why.

---

## F7 — RTI generation

```
1. Review          the draft, with the questions the platform chose
2. Edit            add or remove questions; edit the facts paragraph
3. Checklist       "you will need: a ₹30 fee, an account on the state RTI portal"
4. Download / open the filled draft, plus a deep link to the portal
5. Record          "filed? paste the registration number" → starts the 30-day clock
```

**Step 5 is critical.** Without the registration number, the first-appeal clock cannot run and the
escalation ladder stops. The UI nags gently until it is entered.

**There is no "file for me" button.** See [ADR 0007](../04-adr/0007-never-auto-file.md).

---

## F8 — Offline capture

```
capture → queued locally (visible, honest count)
        → "3 reports waiting to send"
        → auto-sync when connectivity returns
        → notification per report as it enriches
```

Rules: capture must work with zero connectivity; the queue survives app restarts; the queue state is
always visible; nothing is silently dropped; GPS and timestamp are recorded at capture, not at sync.

---

## F9 — WhatsApp intake

```
citizen                          bot
──────                           ───
[sends photo]        ─────────▶
                     ◀───────── "Got it. Send your location? 📍"
[shares location]    ─────────▶
                     ◀───────── "Pothole on Link Road, Kandivali West.
                                 BMC · R/S ward.
                                 This stretch is under warranty until Nov 2028.
                                 Reported ✓  tracesarkar.org/i/8f2a1c
                                 Reply SHARE for a forwardable card."
[SHARE]              ─────────▶
                     ◀───────── [annotated image + text]
```

Constraints: WhatsApp's own location sharing is the only reliable geotag path (photos forwarded
through WhatsApp lose EXIF); first-time users get a one-message consent notice with a link to the
full notice; template approval is required for business-initiated messages.

---

## F10 — Ward page (public, no login)

```
R/S — Kandivali West · BMC
──────────────────────────────────────────
Open issues        214
Past SLA            61      ⚠
Claimed resolved    88
Citizen confirmed   12      ← the number that matters
Median age        44 days

[ map ]  [ list ]  [ by category ]  [ contractors ]

Data: boundaries v2023 · contracts retrieved 9 Aug 2026
```

The claimed-vs-confirmed contrast is placed deliberately. Data freshness is always visible.

---

## Flow principles

1. **Every screen has one primary action**, visually dominant.
2. **The platform fills in; the citizen corrects.** Never the reverse.
3. **Progressive disclosure during enrichment** — show work happening, stream results in.
4. **Never block on optional input.** Skip is always available.
5. **Photograph or it didn't happen** — corroborate, confirm, and reopen all require a fresh photo.
6. **Honest state** — offline queues, low confidence, stale data, and missing coverage are all shown,
   never hidden.
7. **One question, never a form.**
