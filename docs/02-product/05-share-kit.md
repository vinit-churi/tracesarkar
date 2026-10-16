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
> Bombay HC requires action within 48 hours of a pothole being reported.
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

---

## 10. What was built, 14 October 2026

The written post exists; the evidence PDF and the Instagram variant do not. What runs is
`internal/share` behind `GET /v1/public/issue/{id}/share` and `/card`, with the kit itself on the
issue permalink — language toggle, editable text, and WhatsApp first.

| Artefact | State |
|---|---|
| **Post text**, per channel and per language | Built. English, Marathi, Hindi × WhatsApp, X |
| **Annotated image** | Built. The fact block sits *below* the photograph, not across it |
| **Permalink** | Built earlier ([D130](../00-overview/05-decision-log.md)), and it now unfurls: a bare link in any chat becomes the title, the facts and the annotated card ([D139](../00-overview/05-decision-log.md)) |
| **Handle registry** | Built, with one row in it |
| **Evidence PDF** | Built. Signed in, scoped by sighting, deterministic and stamped ([D141](../00-overview/05-decision-log.md), [D142](../00-overview/05-decision-log.md)) |
| **Instagram / Story card** | Built. `?shape=story` draws the same facts into 9:16 ([D140](../00-overview/05-decision-log.md)) |
| **WhatsApp intake bot** | Not built |

### What §3's structure became in practice

The spec's example post leans on contract facts. No public surface carries a contract attribution
yet, so a real post today states the thing, the ward, the authority's own account, whether anybody
has been told, and the obligation where one binds. It degrades truthfully rather than leaving a
gap: an issue nobody has lodged says so in a sentence, which is also the sentence that most often
prompts somebody to lodge it.

Four rules are tests rather than intentions ([ADR 0020](../04-adr/0020-share-text-is-generated-not-written-by-a-model.md)):

- **No model writes a post.** Every sentence is assembled from fields the platform can prove, so
  what a post may say is bounded by the database rather than by a prompt.
- **A deadline with no citation is dropped whole**, never shown as a bare number. On X the short
  reference is a recorded column, not a truncation of the long one — a citation cut to fit a window
  cites nothing a person can look up (migration 0028).
- **An escalation handle waits for an elapsed deadline**, enforced in `share.Compose` and not left
  to the template.
- **An unlodged issue never claims anybody was told**, because photographing something here
  notifies nobody.

### The evidence pack, and why it is the one private artefact

Everything else here is public, because a post nobody can read is not distribution. §7's pack is
the opposite: it carries the **exact** position and the complaint reference, because it exists to
be attached to a complaint and a complaint needs the place and the number. Both already belong to
the person asking for it, so it is served to them and to nobody else, scoped by sighting rather
than by ownership — a stranger is told the issue does not exist rather than that they may not have
it ([D141](../00-overview/05-decision-log.md)).

The photographs in it are still the **redacted** copies. A pack is a document to hand to somebody
else, which is exactly the case redaction is for, and the archival original stays break-glass
however much more useful it would be in a filing.

§7 asked for it to be "deterministically generated and hash-stamped so two people generating it get
an identical file", and that turned out to decide the implementation: most PDF libraries stamp a
creation date and a generated id, which breaks determinism quietly, so the writer is four hundred
lines here with no dependency ([D142](../00-overview/05-decision-log.md)). The stamp is computed
over the facts rather than the rendered bytes, so a later layout change does not invalidate a stamp
already quoted in a filing.

**What the first real pack found.** Compiled from an actual issue rather than a fixture, it put
BMC's RTI manual URL past the right margin where it could not be read, and cited the contract as
`bmc_roads_api` — a key into our own register, and nothing at all to the desk the document is
handed to. A citation nobody can check satisfies hard rule 2's letter and defeats its point
([D143](../00-overview/05-decision-log.md)).

### What the unfurl cost

Making a link unfurl meant giving up the rule that the page shell was the same bytes for every id.
That rule existed so a crawler reached nothing before the gate had been consulted, so the gate moved
*earlier* rather than being relaxed: the issue is read before a byte is written, an unknown id is a
404 from the server, and `publish.Derive` decides whether an image may be named at all. A held
photograph is never offered to a preview crawler, which matters more than on the page itself —
WhatsApp and X cache what they fetch, out of reach of any later correction
([D139](../00-overview/05-decision-log.md)).

### The two things it changed elsewhere

Writing the post surfaced a defect on the page it links to. The Bombay High Court's forty-eight
hours was keyed on the *category*, so the first real kit quoted "All potholes…" beside a missing
manhole cover. It now binds the subcategory it names ([D138](../00-overview/05-decision-log.md)).

And §4's "verify existence before use" is now a column: rows carry a `verify_by` date and stop
being served when it passes. Only `@mybmc` is seeded, marked **secondary** because BMC's own portal
publishes no social links; none of the twenty-four ward accounts is seeded at all
([D136](../00-overview/05-decision-log.md)).

### The rewording layer, 16 October 2026

§3 above specified post text as a model call; [ADR 0020](../04-adr/0020-share-text-is-generated-not-written-by-a-model.md)
shipped it as none. [ADR 0021](../04-adr/0021-a-model-may-reword-a-generated-post-never-author-one.md)
splits the difference along the line that turned out to matter: a model may **reword** a composed
sentence and may not **compose** one.

What that looks like in the response. Every post now carries a `wording` field — `template` or
`model` — and the kit carries a `polish` block naming the model, the prompt version, how many posts
were reworded and how many kept the composed text. `review` is untouched: a reworded Marathi post
is still a draft, and `draft_languages` still lists `mr` and `hi`.

What the model is given: the prose lines, numbered. Not the fact set, not the permalink, not the
issue id. What comes back is checked span by span against what went out — citation, ward code,
authority, `@handle` byte for byte; numbers as a multiset; no new account, tag or link; the right
script; the tone list of §3; the channel window — and a single failure anywhere serves the composed
post instead. A half-reworded post is refused too, because two registers in one message read as two
authors.

Why it was worth doing at all: §1's thesis is that the share *is* the distribution channel, and a
stilted post in the reader's own language is a post that does not get sent. The English templates
read acceptably because a person wrote them; the Marathi and Hindi were written by the same person,
who speaks neither.

### What has to happen before this is finished

- **A native speaker reads `internal/share/i18n/share.mr.json` and `share.hi.json`.** Both ship as
  drafts, labelled as drafts in the API and on the page. This is the one outstanding item that a
  user sees. The sheet to hand them is
  [13-vernacular-review-sheet.md](13-vernacular-review-sheet.md); the rewording layer above does
  **not** close it — it makes a stilted sentence
  fluent, which is not the same as making it right. The register rules in
  `internal/polish/prompts/share_polish_v1.md` are now part of what that reader should check.
- **First-hand verification of `@mybmc`**, and a decision on ward handles.
- **A WhatsApp intake bot** (§5.5), which is a distribution unlock rather than a share artefact,
  and the only item in this document that has not been started.
