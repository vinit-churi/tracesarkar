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
| Marathi and Hindi native-speaker review of all user-facing copy | Per release |
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
