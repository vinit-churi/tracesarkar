# Share kit

**Thesis:** the share is the distribution channel, and the friction that kills sharing is *writing
the post*. TraceSarkar writes it.

A share kit is generated for every verified issue, in the user's language, tailored per channel, and
pre-loaded with the facts that make the post land: the authority, the contract, the money, the
warranty status, and the timeline.

---

## 1. What gets generated

| Artefact | Channel | Notes |
|---|---|---|
| **Post text** | X / Twitter | ≤ 280 chars, correct handles, 2–3 hashtags, permalink |
| **WhatsApp card** | WhatsApp | Longer text + image; formatted for forwarding; no external-link dependence for the core facts |
| **Annotated image** | All | The photo with an overlay: location, date, authority, contract facts, TraceSarkar mark |
| **Permalink** | All | `tracesarkar.org/i/<slug>` with Open Graph and Twitter Card metadata |
| **Evidence PDF** | Email, meetings, filings | Photos + timeline + contract + jurisdiction + sources |
| **Instagram/Story card** | Instagram, Status | 9:16 variant of the annotated image |

---

## 2. The annotated image

The single most important artefact — most sharing in India is image forwarding, and the image must
carry the argument without the link.

```
┌────────────────────────────────────────────────────────┐
│                                                        │
│                  [ the photograph ]                    │
│                                                        │
│                                                        │
├────────────────────────────────────────────────────────┤
│ POTHOLE · Link Road, Kandivali West                    │
│ 11 Jun 2026, 08:14  ·  19.2094, 72.8348                │
│                                                        │
│ Authority   BMC · R/S Ward · Roads & Traffic           │
│ Contract    WS/2023/ROAD/117 · ₹4.11 cr                │
│ Warranty    valid until 28 Nov 2028   ⚠ IN WARRANTY    │
│ Reported    7 defects on this stretch since completion │
│                                                        │
│ tracesarkar.org/i/8f2a1c        Sources on the page ↗  │
└────────────────────────────────────────────────────────┘
```

Rules:
- Facts only. No adjectives, no accusation, no emoji in the fact block.
- Every fact on the card is reproducible from the permalink.
- Contractor name appears **only** when the match confidence is above the publication threshold and
  the source is cited on the linked page.
- The TraceSarkar mark and URL are always present — a forwarded image without provenance is a rumour.

---

## 3. Post text generation

Generated per channel and per language. Structure:

```
[WHAT] at [WHERE].
[AUTHORITY], this is your [WARD] ward.
[MONEY/WARRANTY FACT].
[SLA FACT].
[LINK]
[HASHTAGS]
```

Example (English, X):

> Pothole on Link Road, Kandivali West. @mybmc — R/S ward.
> This stretch is under warranty from contract WS/2023/ROAD/117 (₹4.11 cr) until Nov 2028.
> Bombay HC requires action in 48 hours.
> tracesarkar.org/i/8f2a1c
> #MumbaiRoads #RSWard

Marathi and Hindi variants are generated from the same fact set, not machine-translated from the
English string — translating a hashtag-laden English tweet produces garbage. See
[accessibility and vernacular](10-accessibility-and-vernacular.md).

### Tone rules

| Do | Don't |
|---|---|
| State the fact and the obligation | Allege corruption |
| Tag the institutional handle | Tag individual officials by personal account |
| Cite the warranty and the SLA | Use insults, sarcasm, or party references |
| Use the ward code — it makes the post routable internally | Use vague "Mumbai roads are terrible" framing |

The tone rules are enforced in the generation prompt **and** validated post-generation against a
prohibited-pattern list before the text is shown.

---

## 4. Handle registry

A curated registry maps `(authority, ward, category)` → official social handles.

```yaml
- authority: BMC
  handles:
    x: ["@mybmc"]
    ward_x:
      R/S: ["@mybmcRSward"]      # verify existence before use
  escalation_handles:
    x: ["@CMOMaharashtra"]        # used only at escalation stages, never on first report
- authority: WesternRailway
  handles:
    x: ["@WesternRly"]
    rail_madad: ["@RailMinIndia"]
```

**Rules:**
- Handles must be verified as official before entering the registry, with a source and a check date.
- Never auto-tag a personal account.
- Escalation handles (CM, ministers) are unlocked only after an SLA breach — tagging them on every
  pothole trains everyone to ignore the platform.
- Registry entries expire and require re-verification (handles get renamed and abandoned).

---

## 5. WhatsApp — the primary channel

WhatsApp is where MMR civic complaint behaviour already lives (society groups, ward groups).
Design accordingly:

1. **The card must stand alone.** Assume it will be forwarded far from the link.
2. **Text and image in one message**, so forwarding keeps them together.
3. **Marathi/Hindi first** in the default variant for MMR.
4. **A "share to my society group" affordance**, not a generic share sheet, for Sunita-type users.
5. **Inbound intake bot (A6):** a WhatsApp number that accepts a photo + location and creates a
   report. This is likely the largest single distribution unlock available. Requires WhatsApp
   Business API onboarding, template approval, and a cost model — see
   [cost model](../06-operations/04-cost-model.md).

---

## 6. Permalink page

Public, no login, fast, low-bandwidth.

Contains:
- Photograph(s), with redaction applied
- Map with a **coarsened** location marker (never the exact reporter coordinates)
- Full timeline (see [snap-to-action pipeline](04-snap-to-action-pipeline.md#stage-7--track))
- Jurisdiction resolution with confidence and basis
- Contract attribution with every source cited and dated
- "How was this matched?" expandable provenance
- Corroboration count
- Dispute link
- Structured data (`schema.org`) for search and social previews

**Open Graph card** renders the annotated image, so a bare link in any chat becomes the full argument.

---

## 7. Evidence PDF

For Sunita's meetings and for attachment to filings. Sections:

1. Cover: issue ID, category, location, authority, date range
2. Photographs with capture metadata
3. Timeline, verbatim
4. Jurisdiction resolution and basis
5. Contract attribution with source documents referenced
6. Prior reports at the same asset
7. Correspondence and official reference numbers
8. Source register: every URL, retrieval date, and hash
9. Disclaimer: this is a compiled record, not legal advice

Deterministically generated and hash-stamped so two people generating it get an identical file.

---

## 8. Anti-abuse considerations

| Risk | Control |
|---|---|
| Share kit used to brigade an official | Rate limits; no personal handles; escalation handles gated on SLA breach |
| Contractor named on thin evidence | Publication threshold on match confidence; name omitted below it |
| Image forwarded with facts stripped or altered | Mark + URL burned into the image; permalink is authoritative; a public correction log covers genuine errors |
| Share kit generated for an unverified issue | Only verified issues get a kit (except `hazard_to_life`, which gets a hazard-styled variant that names no contractor) |

---

## 9. Measurement

| Metric | Why |
|---|---|
| Kit generation rate per verified issue | Does the fact set actually motivate sharing? |
| Channel split | Where to invest |
| Permalink referral traffic by channel | Real reach, not vanity |
| Reports originating from a shared permalink | The compounding loop |
| Authority response rate on shared vs unshared issues | **The core hypothesis test:** does public exposure change resolution? |

That last row is the experiment the whole platform exists to run. It should be instrumented from
day one, with a holdout: a random subset of issues where the share kit is generated but the user is
not prompted to share, to establish a baseline.
