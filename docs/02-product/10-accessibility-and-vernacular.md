# Accessibility and vernacular access

**Premise:** a platform that works only in English, only on a good phone, and only with a good
connection is a platform for people who already have access to officials. Those are not the people
worst served by civic failure.

---

## 1. Languages

| Language | Milestone | Rationale |
|---|---|---|
| **Marathi (मराठी)** | v0.1 | The state language. Non-negotiable in MMR |
| **English** | v0.1 | Working language of much of the city and all official documents |
| **Hindi (हिन्दी)** | v0.3 | Very large MMR population |
| **Gujarati (ગુજરાતી)** | v0.5 | Significant MMR population |

### Translation rules

1. **Generate per language from the fact set, never translate a generated string.** A machine
   translation of an English tweet with hashtags and handles produces garbage. Share text, chatbot
   answers, and notification copy are all generated in the target language from the same structured
   facts.
2. **Legal and procedural terms keep their canonical form with a gloss:** `प्रथम अपील (First
   Appeal)`. The citizen will need the exact term on a government form.
3. **Official names are not translated.** "Brihanmumbai Municipal Corporation" stays, with a
   Marathi name field alongside where the body publishes one.
4. **Numbers, dates, and currency** follow Indian conventions: lakh/crore, DD/MM/YYYY, ₹.
5. **Code-mixed input is expected and accepted.** "pothole kadhi fix honar?" must work.

### Translation workflow

UI strings live in `i18n/{lang}.json`, reviewed by a native speaker before release. Machine
translation is acceptable as a first draft for UI chrome; it is **not** acceptable for legal text,
consent notices, or anything a citizen will act on.

**Built, 14 October 2026.** The first copy to exist in all three languages is the share kit's, and
it lives in `internal/share/i18n/share.{lang}.json` rather than at the repository root — beside the
package that embeds it, because `go:embed` cannot reach above its own directory. Each file carries
its own `review` state, and that file is the artefact a reviewer is handed; nothing in Go has to be
read to check Marathi ([D137](../00-overview/05-decision-log.md),
[ADR 0020](../04-adr/0020-share-text-is-generated-not-written-by-a-model.md)).

Three rules the share kit's copy follows, which any later vernacular surface should follow too:

- **Legal citations are not translated.** They pass through every language unchanged, because a
  citation is checked against the document that produced it.
- **Digits stay Latin in all three languages**, for the same reason.
- **A missing term falls back to the category's word and says so**, rather than being guessed at or
  left as English in the middle of a Marathi sentence.

**Reviewed, 16 October 2026.** `share.mr.json` and `share.hi.json` were served as labelled drafts
for two days and have now been read by native speakers; both say `"review": "reviewed"` and no post
carries a draft notice. The review changed **19 of 43** Marathi strings and **20 of 43** Hindi ones,
and the record of every change with its reason is
[13-vernacular-review-sheet.md](13-vernacular-review-sheet.md)
([D175](../00-overview/05-decision-log.md)).

Four findings generalise beyond the share kit and should shape any vernacular surface built next.
**`प्राधिकरण` is not a word people use** — the natural form is `संबंधित विभाग`, and this is the
single most common defect in the draft. **A category word needs `समस्या`**, because an adjective or
an abstract noun does not read as a problem in a sentence frame. **"Repaired" does not fit most of
what this platform reports**, so the status copy says "the work being completed" instead and
sidesteps gender agreement at the same time. And **English relative clauses must not be translated
literally** — `तो ज्या कागदपत्रातून आला आहे त्यासह` became `मूळ कागदपत्रांसह`, a third of the
length. Punctuation is the **full stop in both languages**: modern Marathi uses it exclusively, and
for Hindi either is correct but `.` is what urban Mumbai writes beside Latin ward codes.

**The sheet to hand them is [13-vernacular-review-sheet.md](13-vernacular-review-sheet.md)**, added
16 October 2026. It is written for somebody who knows nothing about the project or about code: what
the message is for, who reads it, the seven rules the copy has to follow, then every sentence and
every word with an ID to answer against. It also carries the four questions per language that a
non-speaker cannot settle — whether the danda belongs, whether `आढळले` agrees with `{n}`, whether a
bare ward letter joins correctly to `विभागात`, and whether the register is right for a ward office
*and* a WhatsApp forward.

**A model may now reword that copy at request time, 16 October 2026.** Rule 1 above is unchanged and
is the reason the layer is shaped the way it is: nothing is translated from a generated English
string. `internal/polish` takes the *Marathi* lines a Marathi template composed and asks for better
*Marathi*, line for line, and publishes the answer only if every citation, number, ward code and
account survives it
([ADR 0021](../04-adr/0021-a-model-may-reword-a-generated-post-never-author-one.md)). Rules 2 and 3
are enforced rather than asked for: a legal term and an official name are protected spans, and a
line that alters one is discarded.

Two things this does not do, both worth stating plainly.

- **It does not make a language reviewed.** A fluent sentence nobody who speaks Marathi has read is
  still an unreviewed sentence — which is why rewording may not touch the field. A test compares the
  review state before and after rather than asserting a value, so signing a language off is a change
  to one JSON file and not to any test.
- **It adds a second artefact for the same reviewer.**
  `internal/polish/prompts/share_polish_v1.md` carries the per-language register rules — that BMC's
  word for an administrative ward is विभाग and never प्रभाग, that Mumbai Hindi is neither
  Sanskritised nor Urdu-leaning, where a danda belongs and where it does not. Those are claims about
  Marathi and Hindi written by someone who speaks neither, which is exactly the condition that made
  the vocabulary files a review task.

---

## 2. Voice

Typing a civic complaint on a phone is a real barrier for elderly users, low-literacy users, and
anyone standing in the rain.

| Capability | Implementation | Milestone |
|---|---|---|
| **Speech-to-text (report descriptions)** | Bhashini ASR — 22 Indian languages, optimised for Indian accents | v0.4 |
| **Speech-to-text (chatbot)** | Same | v0.4 |
| **Text-to-speech (official replies, legal text, instrument summaries)** | Bhashini TTS | v0.6 |

Bhashini's ULCA pipeline model: a pipeline-config call returns service IDs and a compute
authorisation value; the compute call then performs ASR / NMT / TTS, individually or chained.
Authentication uses a user ID, a ULCA API key, and an inference API key.

**Design rules for voice:**
- Voice is always an *alternative*, never the only path. Text always works.
- The transcript is shown for correction before submission.
- The original audio is retained with the report (it is often better evidence than the transcript).
- A Bhashini outage degrades to text with an honest message, never to a broken flow.

---

## 3. Low-bandwidth and low-end devices

Monsoon is exactly when networks degrade and reports spike.

| Constraint | Response |
|---|---|
| **Slow / intermittent connectivity** | Offline capture with a durable queue; opportunistic sync; visible queue state |
| **Expensive data** | Client-side image compression before upload (target ≤ 400 KB per upload frame while preserving defect legibility); no video; no autoplay |
| **Low-end Android** | PWA first; small JS bundle; no heavy client-side framework on the capture path; server-rendered public pages |
| **Small screens** | Single-column layouts; large touch targets (≥ 48 dp) |
| **Older Android WebView** | Progressive enhancement; the capture flow must work without modern APIs |

**Budget:** the public issue page must be usable on a 2G connection. That is a hard constraint that
shapes the whole front-end approach, not an aspiration.

---

## 4. WCAG 2.2 AA

Target from v0.3, with the highest-value items applied from v0.1.

| Requirement | Implementation |
|---|---|
| Colour contrast ≥ 4.5:1 | Design tokens enforce it; CI check on the palette |
| Never colour alone | Severity uses icon + text + colour; status uses a label |
| Keyboard operable | Full app; visible focus indicators |
| Screen reader | Semantic HTML; ARIA only where semantics are insufficient; every image has meaningful alt text |
| Alt text for citizen photographs | **Generated from the classification** ("Pothole with standing water on an asphalt road, Link Road, Kandivali West") — a genuine accessibility win from the AI pipeline |
| Touch targets ≥ 24×24 CSS px | Design system |
| Text resize to 200% | Fluid layouts, relative units |
| Motion | `prefers-reduced-motion` respected; no essential information conveyed by motion |
| Forms | Labels, error identification, error suggestion |
| Timeouts | No session timeouts on the capture flow |

Audited with automated tooling in CI **and** a manual screen-reader pass per release. Automated
tooling catches perhaps a third of real issues.

---

## 5. Literacy

Beyond language: many users can speak a language they cannot comfortably read, especially in a
bureaucratic register.

| Technique | Application |
|---|---|
| **Icons + text**, never icons alone | Category selection, status |
| **Photographs as options** | Disambiguation questions use thumbnails, not descriptions |
| **Plain-language default** | The UI never uses bureaucratic register. "The 48-hour deadline has passed" not "SLA breach recorded" |
| **Legal text summarised** | Every generated instrument has a plain-language summary above the formal text |
| **Read-aloud** | TTS on official replies and instrument summaries |
| **Numbers spoken plainly** | "₹4.11 crore" not "41100000" |

---

## 6. Non-smartphone access

| Channel | Status |
|---|---|
| **WhatsApp intake** | v0.3 — the realistic answer for MMR. Supports photos and location natively; near-universal penetration; where civic complaint behaviour already happens |
| **SMS / missed call** | Deferred. Low value once WhatsApp exists; revisit only if evidence shows a served population that WhatsApp misses |
| **Assisted reporting** | An RWA organiser or volunteer reports on someone's behalf, with attribution ("reported by X on behalf of a resident"). A v0.5 feature and a genuinely important one |

---

## 7. What we will not do

| Not doing | Why |
|---|---|
| Require an email address | Excludes a large population for no benefit |
| Require Aadhaar | See [ADR 0011](../04-adr/0011-phone-only-identity.md) |
| Ship English-only "for now" | "For now" becomes forever, and it defines who the platform is for |
| Machine-translate legal text without review | A mistranslated statutory term in a filing is a real harm |
| Gate any core function behind an app install | The web PWA is the primary client |

---

## 8. Testing

| Test | Cadence |
|---|---|
| Screen reader pass (TalkBack, NVDA) | Per release |
| Low-end device test (₹8,000-class Android, throttled network) | Per release |
| Marathi and Hindi native-speaker review of all user-facing copy | Per release · **done for the share kit** 16 October 2026 ([D175](../00-overview/05-decision-log.md)); every later vernacular surface needs its own pass |
| Voice input in each supported language | Per release |
| 2G throughput test on the public page | Per release |
| Accessibility audit (automated + manual) | Quarterly |

---

## 9. Open questions

- Does Bhashini's licensing permit use by a non-government public platform at our scale, and at what
  cost? **Blocking for v0.4.**
- Is there a served population that WhatsApp genuinely misses in MMR, and would SMS reach them?
- For assisted reporting, how is the on-behalf-of relationship represented without creating an
  impersonation vector?
- Should generated alt text be published as the canonical description of a photograph, given it is
  model output? *Leaning: yes, labelled as auto-generated, with an edit affordance.*
