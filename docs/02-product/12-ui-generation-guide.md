# UI generation guide

Instructions and ready-to-paste prompts for producing TraceSarkar screen imagery with an
image-generation model.

The design authority is [`11-screen-spec.md`](11-screen-spec.md). This file only translates that
spec into prompts. If a generated image looks good but contradicts the spec, the image is wrong.

---

## 1. How to drive the generator

| Rule | Why |
|---|---|
| **One screen per render.** Never ask for a 10-screen grid | Text degrades badly at small scale; a grid renders every string as noise. Assemble the grid afterwards in Figma/Canva |
| **Canvas 1080 × 2340 px** (19.5:9), or 1242 × 2688 for retina crops | Matches the real device target |
| Use a model with strong text rendering (GPT Image, Gemini image, Ideogram, Flux-with-text). Diffusion models with weak typography will garble every label | Screens are 70% typography |
| **Paste the style capsule (§2) verbatim into every prompt.** Do not paraphrase it between screens | Consistency across a set comes from a byte-identical style block |
| Fix the seed after the first acceptable render, and reuse it plus a reference image for every subsequent screen | Keeps palette, corner radii and type scale stable |
| Keep visible strings ≤ 12 words per element and ≤ 60 words per screen. Give the model the exact strings from the copy deck (§4) | Every extra word is another chance to hallucinate a glyph |
| Expect to rebuild text in a design tool for anything client-facing | No generator is reliable at ₹ figures, contract IDs or Devanagari |
| **Never** generate Devanagari with an image model unless it is explicitly Devanagari-capable — render Marathi screens with real text in Figma | Garbled Devanagari in a civic product aimed at Marathi speakers is worse than English-only |

**Iteration protocol:** render → check against the QA rubric (§6) → fix by *editing the prompt's
element list*, not by adding adjectives. "More civic-tech" produces nothing; "remove the drop
shadow, use a 1 px #E3E1DC border" produces a fix.

---

## 2. Style capsule — paste verbatim into every prompt

```
STYLE: A high-fidelity mobile UI screenshot for an Indian civic-accountability app. Flat,
document-like, editorial. Not glossy, not neumorphic, no glassmorphism, no gradient meshes, no
3D icons, no drop shadows except on the camera shutter and modal sheets.

PALETTE (use exactly): background #FBFAF8 warm off-white; cards #FFFFFF; all separators are 1px
#E3E1DC hairlines; primary text #141518; secondary text #4A4D55; metadata text #7C8089; single
accent plum #6B2C4F used sparingly for the active nav item and focus states; alert #A33A1F;
caution #8A6410; confirm-green #1F6A4D used ONLY on a small confirmed-status chip; evidence blocks
have a #F2F0EB fill with a 3px #6B2C4F left rule. No saffron, no orange, no blue, no green as a
dominant or brand colour.

TYPE: clean neutral grotesque sans (Inter-like). Screen titles 28px semibold, body 16px, labels
14px medium, metadata 13px, contract IDs and amounts in a tabular monospace. Left-aligned. Generous
line height. Indian number formatting.

GEOMETRY: 4pt grid, 16px screen gutters, 12px card radius, 8px chip radius, 48px minimum tap
targets, single-column layout, plenty of whitespace, data-dense but calm.

DEVICE: straight-on flat screenshot, no device bezel, no hand, no perspective, no reflections. iOS
status bar at top showing 9:41 with signal, wifi, battery.

MOOD: sober public-record tool, closer to a court filing or a transit app than to a consumer social
app. Trustworthy, plain, unexcited.
```

**Negative prompt (append to every render):**

```
NEGATIVE: drop shadows, glassmorphism, gradients, neon, purple-blue tech gradient, 3D rendered
icons, glossy buttons, stock photography of smiling people, saffron/orange/green/blue brand colour,
Indian flag colours, party symbols, political logos, watermarks, lorem ipsum, garbled text,
duplicated status bars, device bezels, hands holding phone, perspective tilt, dark mode (unless
asked), fake percentage rings, badge that says "Verified", trophy or leaderboard iconography,
confetti.
```

---

## 3. Recurring elements (define once, reference by name)

Give the model these definitions in the prompt whenever the screen uses them.

| Name | Prompt fragment |
|---|---|
| `STATUS CHIP` | small pill, 8px radius, 1px border in its status colour, tinted background, a small line icon on the left and a text label — never colour alone |
| `SOURCE LINE` | a 13px metadata line in #7C8089 reading `Source · Mahatenders · retrieved 14 Jul 2026`, with an underlined `view original` link at the end |
| `EVIDENCE BLOCK` | a #F2F0EB block with a 3px #6B2C4F left rule, a 12px uppercase letterspaced label `TRACESARKAR RECORD`, a label/value field list in two columns, and a SOURCE LINE at the bottom |
| `DEADLINE BANNER` | a full-width strip with a clock icon, bold statement of the deadline, remaining time, and a smaller citation line `Bombay High Court, Oct 2025 · read the order` |
| `BOTTOM NAV` | 5 items — Home, Map, a circular capture button raised in the centre in #141518 with a white camera glyph, Reports, You. Active item in plum #6B2C4F, line icons, text labels under every icon |
| `PRIMARY BUTTON` | full-width, 56px tall, 10px radius, #141518 fill, white 16px semibold label, optional left line icon |
| `SECONDARY BUTTON` | full-width, 56px tall, 10px radius, white fill, 1px #E3E1DC border, #141518 label |
| `PHOTO` | a real-looking Indian street photograph of the stated subject, slightly overcast monsoon light, no people's faces visible |

---

## 4. Copy deck

Use these strings verbatim. They are spec-compliant; improvised copy usually is not.

| Key | English string |
|---|---|
| `app.name` | TraceSarkar |
| `app.tagline` | Snap it. Trace it. Act on it. |
| `home.cta` | Report a problem |
| `home.nearby` | 3 open issues within 500 m |
| `home.contrib` | Confirmed fixes · 4 — Re-checks done · 11 — Corroborated · 7 |
| `capture.gps` | GPS ±6 m |
| `capture.coach` | Capture the defect and its surroundings |
| `capture.wide` | One step back — a wide shot so the place is identifiable |
| `capture.skip` | Skip |
| `enrich.1` | Looked at the photo — Pothole · road defect |
| `enrich.2` | Found the ward — BMC · R/S · Kandivali West |
| `enrich.3` | Found the department — Roads & Traffic |
| `enrich.4` | Checking contracts… |
| `enrich.5` | Checking recent reports |
| `disambig.q` | Where exactly is this? |
| `disambig.a` | ON the flyover |
| `disambig.b` | UNDERNEATH the flyover |
| `disambig.note` | Different authorities look after each |
| `confirm.class` | Classified automatically · Pothole · Wrong? Change |
| `confirm.cta` | Submit report |
| `submitted.title` | Reported |
| `submitted.cta` | Share this |
| `deadline.title` | 48 hours to attend · 31 h left |
| `deadline.cite` | Bombay High Court, Oct 2025 · read the order |
| `deadline.passed` | 48 hours have passed. No action recorded. |
| `evidence.title` | This stretch is under warranty |
| `evidence.id` | WS/2023/ROAD/117 |
| `evidence.value` | ₹4.11 crore |
| `evidence.completed` | Completed 28 Nov 2023 |
| `evidence.warranty` | Warranty until 28 Nov 2028 |
| `evidence.repeat` | 7 defects reported on this stretch since completion |
| `evidence.source` | Source · Mahatenders · retrieved 14 Jul 2026 · view original |
| `evidence.match` | How was this matched? |
| `evidence.dispute` | Dispute this |
| `evidence.withheld` | Contractor name withheld — awaiting a second independent source |
| `nodata` | We don't have contract data for this ward yet |
| `action.share` | Share |
| `action.file` | File with BMC |
| `action.ask` | Ask a question |
| `action.rti` | Ask for the records (RTI) |
| `action.rti.sub` | We'll draft it. You review and file it. |
| `rti.cta` | Download the draft |
| `rti.cta2` | Open the RTI portal |
| `rti.never` | Nothing is filed automatically |
| `ref.prompt` | Filed it? Paste the reference number |
| `ref.why` | Without it we can't track the deadline or build the appeal |
| `recheck.title` | BMC says this pothole is fixed |
| `recheck.body` | You're 300 m away. Two minutes? |
| `recheck.yes` | It's fixed |
| `recheck.no` | Still broken |
| `recheck.note` | Either way, one photograph please |
| `status.claimed` | Authority says fixed |
| `status.confirmed` | Confirmed fixed |
| `status.deadline` | Deadline passed |
| `ward.stats` | Open 214 · Past deadline 61 · Authority says fixed 88 · Confirmed by citizens 12 |
| `provenance` | Boundaries v2023 · contracts retrieved 9 Aug 2026 · CC BY-SA 4.0 |
| `queue` | 3 reports waiting to send |

---

## 5. Per-screen prompts

Each block is: `[style capsule] + [screen block] + [negative prompt]`.

### S01 — Home

```
SCREEN: Home screen of TraceSarkar. Top to bottom:
1. Slim app bar: small plum outline logo mark and wordmark "TraceSarkar" at left; "EN | मराठी" text
   toggle and a bell icon at right.
2. A very large primary button, 96px tall, #141518 fill, white camera icon and white label
   "Report a problem". It is the visually dominant element on the screen.
3. A one-line row: "3 open issues within 500 m" with three small 44px rounded photo thumbnails of
   street defects at the right.
4. A card titled "Your contribution" with three stacked rows, each a label and a tabular number:
   "Confirmed fixes  4", "Re-checks done  11", "Corroborated reports  7". No rupee amounts, no view
   counts, no trophies.
5. Section header "Recent" with two list rows: each has a 48px square photo thumbnail, a title
   ("Pothole on SK Bole Marg", "Garbage overflow, Andheri East"), a grey locality line, and a
   STATUS CHIP at the right reading "Deadline passed" in alert red and "Acknowledged" in neutral.
6. BOTTOM NAV.
```

### S02 — Capture

```
SCREEN: Camera capture screen. Full-bleed live PHOTO of a rain-slicked Mumbai street with a large
water-filled pothole in the foreground; no people's faces. Overlays:
- Top row: white "×" at left, flash icon at right, and a centred dark translucent pill reading
  "GPS ±6 m" with a small filled location dot.
- A single dark translucent coaching bar near the lower third: "Capture the defect and its
  surroundings" with a small "?" at the right.
- Bottom bar over the photo: a 72px white shutter ring in the centre, a small gallery thumbnail at
  the left, a flip-camera icon at the right.
NO step indicator, NO "1 2 3 4" wizard dots, NO form fields, NO category dropdown.
```

### S03 — Wide frame

```
SCREEN: Same camera viewfinder, stepped back so the street and a building line are visible.
Overlay text at the top centre: "One step back — a wide shot so the place is identifiable".
Bottom bar: white shutter ring in the centre and a text button "Skip" at the right, same size and
weight as the other controls. GPS pill still visible.
```

### S04 — Enriching

```
SCREEN: Processing screen. Top: the captured pothole PHOTO as a 16:9 rounded thumbnail. Title
"Working on your report". Below it, a vertical list of five rows separated by 1px hairlines; each
row has a status glyph at the left, a bold label, and a grey result line under it:
✓ "Looked at the photo" / "Pothole · road defect"
✓ "Found the ward" / "BMC · R/S · Kandivali West"
✓ "Found the department" / "Roads & Traffic"
◐ "Checking contracts" / "…" (this row shows a small spinner)
○ "Checking recent reports" (dimmed, not started)
At the bottom, a quiet grey note: "Your report is already saved."
NO circular percentage ring, NO "85%", NO progress bar.
```

### S05 — Disambiguation

```
SCREEN: A single-question sheet. Title "Where exactly is this?". Two large side-by-side
photographic option cards, each 1:1, with a 1px hairline border and a caption bar: left card is a
PHOTO looking along the top deck of a flyover captioned "ON the flyover"; right card is a PHOTO of
the road underneath a flyover captioned "UNDERNEATH the flyover". Below them a grey line:
"Different authorities look after each." A tertiary text button "Not sure" beneath.
NO dropdown, NO list of authority names, only two photo choices.
```

### S06 — Confirm

```
SCREEN: Review-before-submit. Top: horizontal strip of two captured PHOTO thumbnails, the first
with a small "Redacted" tag. Then four hairline-separated rows, each a label with a small trailing
text link:
- "Classified automatically · Pothole" with a link "Wrong? Change"
- "SK Bole Marg, Dadar West · ±6 m" with a link "Move pin"
- "BMC · R/S ward · Roads & Traffic" with a link "Not right?"
- A collapsed row "Add a note (optional)" with a "+" at the right.
Then a DEADLINE BANNER reading "If BMC does not attend within 48 hours, you'll be able to escalate"
with the citation line under it. At the bottom a PRIMARY BUTTON "Submit report".
```

### S07 — Submitted

```
SCREEN: Success state. A single small confirm-green check glyph, then the title "Reported" in 28px.
A grey line: "BMC · R/S ward · Roads & Traffic — 48-hour clock started". Then a compact EVIDENCE
BLOCK with one line "This stretch is under warranty until 28 Nov 2028" and a SOURCE LINE. Then a
PRIMARY BUTTON with a share glyph reading "Share this", and beneath it two SECONDARY BUTTONs side by
side: "View issue" and "Report another". No confetti, no illustration, no celebration graphics.
```

### S08 — Issue detail (the hero screen)

```
SCREEN: Issue detail, scrollable. Top to bottom:
1. Back arrow, title "Issue", share glyph, bookmark glyph.
2. Full-width 4:3 PHOTO of a pothole with a small "Redacted" tag bottom-left.
3. Title "Pothole · SK Bole Marg, Dadar West" and a STATUS CHIP "Deadline passed" in alert red.
4. Grey line "Classified automatically · Pothole — Wrong? Change".
5. An EVIDENCE BLOCK labelled "TRACESARKAR RECORD" with rows: "Authority  BMC · R/S ward",
   "Department  Roads & Traffic", and a SOURCE LINE reading "Ward boundary v2023 · department
   mapping retrieved 2 Aug 2026".
6. A DEADLINE BANNER: clock icon, "48 hours have passed. No action recorded.", citation line
   "Bombay High Court, Oct 2025 · read the order".
7. A second EVIDENCE BLOCK titled "This stretch is under warranty" with a two-column field list:
   "Contract  WS/2023/ROAD/117" (monospace), "Value  ₹4.11 crore", "Completed  28 Nov 2023",
   "Warranty  until 28 Nov 2028", "Contractor  name withheld — awaiting a second source",
   "Repeat defects  7 since completion", then a SOURCE LINE, then two small text buttons side by
   side: "How was this matched?" and "Dispute this".
8. A small rounded map thumbnail with a single plum pin and a grey caption "Approximate location".
9. Action list: a PRIMARY BUTTON "Share", then two SECONDARY BUTTONs "File with BMC" and "Ask a
   question", then a dimmed row "First appeal — available 30 days after an RTI is filed".
NO badge reading "Verified". NO percentage confidence. NO rupee figure without the source line
directly beneath its block.
```

### S09 — Contract record

```
SCREEN: Contract detail. Header "Contract record" with a back arrow. A dark #141518 card at the top
containing: monospace "WS/2023/ROAD/117", a small light label "Roads & footpaths work", and a small
outline tag reading "Cross-checked · 2 sources" (NOT the word "Verified" alone).
Below, a white card of label/value rows separated by hairlines: Work type, Work order date, Award
amount ₹4,11,00,000, Contractor, Defect liability period 12 months, Start date, End date, and
"Warranty status — In defect liability period, ends in 6 months" with a small caution-coloured chip.
Then a SOURCE LINE.
Then a "Documents" section: three rows with a PDF glyph, filename ("Work order", "Agreement",
"Bill of quantities"), a size and retrieval date in metadata grey, and a download glyph.
Then two full-width SECONDARY BUTTONs: "How was this matched?" and "View the archived copy", and a
tertiary text row "Dispute this record · Right of reply".
```

### S10 — Share kit

```
SCREEN: Share sheet. Title "Share this". A preview card showing the annotated share image: the
pothole PHOTO with a burned-in caption strip at the bottom containing a small logo, "Dadar West ·
24 Aug 2026" and a contract ID. Under it, a language segmented control "EN | मराठी | हिंदी" with EN
active. Then an editable text block with a 1px border containing four short lines of the share text
and a small "Edit" link. Then a row of four square channel buttons with labels beneath: WhatsApp
first, then X, then Copy, then More.
```

### S11 — My reports

```
SCREEN: List screen titled "My reports". A horizontally scrolling row of six filter chips:
"All" (active, plum), "Open", "Deadline passed", "Authority says fixed", "Confirmed fixed",
"Escalated". Then four list rows, each with a 56px photo thumbnail, a title, a grey locality and
date line, and a STATUS CHIP: row 1 "Deadline passed" alert red; row 2 "Authority says fixed"
caution amber; row 3 "Confirmed fixed" confirm green; row 4 "Escalated" neutral. Row 2 has a
right-aligned small button "Re-check". BOTTOM NAV at the bottom with "Reports" active.
```

### S15 — Ward page (web, desktop)

```
SCREEN: A desktop web page, 1440px wide, same palette, editorial and data-dense. Header
"R/S — Kandivali West · BMC". A row of five statistic tiles with big tabular numbers and small
labels: "Open 214", "Past deadline 61", "Authority says fixed 88", "Confirmed by citizens 12",
"Median age 44 days". The fourth tile is visually emphasised with a thin plum underline. Below, a
tab row "Map | List | By category | Contracts", with a muted map of a Mumbai ward showing clustered
dots. A footer strip in metadata grey: "Boundaries v2023 · contracts retrieved 9 Aug 2026 ·
CC BY-SA 4.0 · Corrections".
```

### S17 — Offline outbox

```
SCREEN: Titled "Waiting to send". A sticky bar at the top in a light caution tint reading
"3 reports waiting to send". Three rows, each with a photo thumbnail, a capture time, a small
"captured at 18:42 · GPS ±8 m" metadata line, and a right-aligned state: "Waiting for network",
"Uploading 40%" with a thin progress line, "Failed — will retry" in alert red with a small "Retry"
text button. A grey footer note: "Nothing is deleted. Reports send when you're back online."
```

### S18 — Re-check

```
SCREEN: A focused prompt card on a plain background. Small photo thumbnail of the original pothole
at the top. Title "BMC says this pothole is fixed". Body "You're 300 m away. Two minutes?". Two
equally weighted SECONDARY BUTTONs stacked: "It's fixed" with a camera glyph, "Still broken" with a
camera glyph. A grey line beneath: "Either way, one photograph please."
```

### S19 — Deadline passed

```
SCREEN: Titled "What you can do now". At the top a DEADLINE BANNER in an alert tint: "48 hours have
passed. No action recorded." with the citation line. Then a stacked action list:
- A dominant card: bold "Ask for the records (RTI)", grey sub-line "We'll draft it. You review and
  file it.", chevron.
- A normal row: "Share the missed deadline".
- A normal row: "Add to your ward campaign".
- A dimmed row with a lock glyph: "First appeal — available 30 days after an RTI is filed".
```

### S23 — RTI draft

```
SCREEN: Titled "RTI draft". A stepper of four small text labels at the top: "Review · Edit ·
Checklist · Download", with "Review" active. A document preview card with a 1px border showing a
small Ashoka-emblem-free plain heading "Right to Information Act, 2005 — Application", an addressee
block, and three numbered request lines in 14px. Above the document, a light #F2F0EB summary strip:
"In plain language: you are asking BMC for the work order, inspection reports and complaint history
for this stretch."
At the bottom: a checklist card with three ticked items "₹10 application fee", "An account on the
Maharashtra RTI portal", "The PIO address (filled in)", each with a small metadata source note.
Then a PRIMARY BUTTON "Download the draft" and a SECONDARY BUTTON "Open the RTI portal", and under
them a small centred grey line: "Nothing is filed automatically."
NO button reading "File this RTI".
```

### S24 — Reference capture

```
SCREEN: Titled "Record the filing". A single large input field with the placeholder "Registration
number", a small "or photograph the receipt" text button with a camera glyph, and a grey
explanation: "Without it we can't track the deadline or build the appeal." A PRIMARY BUTTON "Save".
A quiet secondary text button "I haven't filed it yet".
```

### S25 — Escalation ladder

```
SCREEN: Titled "Escalation path". A vertical timeline with five rungs connected by a 1px line.
Each rung: a small circular node, a bold role ("You", "Ward Engineer", "Assistant Engineer",
"Executive Engineer", "Municipal Commissioner"), a grey sub-line with the body and the window
("F/South Ward · 48 hours"), a tiny citation link ("source"), and a right-aligned state chip:
"Done" for the first, "Pending" for the second, "Not yet due" dimmed for the rest.
At the bottom, a #F2F0EB card: "No response? We'll draft the next instrument for you to review and
file." with a smaller line "Nothing is ever filed automatically · why".
```

### S28 — Compensation assistant

```
SCREEN: Titled "Compensation claim". Deliberately plain and quiet: no accent colour anywhere, no
chips, generous whitespace, 16px body text. A short paragraph, then a plain field list: "Forum —
Bombay High Court", "Amount fixed for a death — ₹6,00,000", "Injury — ₹50,000 to ₹2,50,000", each
with a small citation line beneath. Then an evidence checklist of five items with empty checkboxes.
A SECONDARY BUTTON "Talk to a legal-aid partner" and a PRIMARY BUTTON "Start the claim file".
No share button, no celebratory colour, no illustration.
```

### S29 — Contractor record

```
SCREEN: Titled "Contractor record". A plain header with the company name and a monospace CIN line.
A field list: "Contracts on record 14", "Total awarded ₹212 crore", "Stretches in warranty 6",
"Defects reported on covered stretches 41". Each row has its own small SOURCE LINE. Then a section
"Blacklisting entries" with one row: a date, an issuing body, and a "view notice" link — or, if
none, the line "No blacklisting entries found in the sources we monitor."
At the bottom a bordered #F2F0EB card in 14px: "These are records, not allegations. TraceSarkar does
not conclude that any party has acted improperly." with two text links "Right of reply" and
"Dispute a fact".
```

---

## 6. QA rubric — reject a render that fails any of these

1. Any button that says or implies the platform files something (`File this RTI`, `Auto-file`).
2. A rupee amount, contractor name or contract ID with no `SOURCE LINE` in the same block.
3. A badge reading `Verified` with no stated object of verification.
4. A single `Resolved` status where the spec requires `Authority says fixed` and `Confirmed fixed`.
5. Saffron, orange, green or blue used as a dominant surface or brand colour; any party symbol.
6. A capture screen with a wizard stepper, a category dropdown or a description field.
7. A percentage ring on the enrichment screen.
8. Missing GPS-accuracy indicator on a capture screen.
9. Status conveyed by colour with no icon and no text label.
10. Vanity metrics: report counts, view counts, rupees "tracked by you", trophies, leaderboards.
11. Faces or number plates legible in any screenshot photograph.
12. Garbled or invented text anywhere — including plausible-looking fake contract IDs in a screen
    that will be shown publicly. Use the copy deck's fixed sample values so every asset agrees.
13. Drop shadows, gradients or glass effects.
14. Any Devanagari string that a native reader cannot read.

---

## 7. Sample-data discipline

Every generated asset uses the same fictional-but-plausible sample record, so screens agree with
each other and with the docs:

| Field | Value |
|---|---|
| Issue | Pothole · SK Bole Marg, Dadar West, Mumbai 400028 |
| Ward | BMC · F/South (for Dadar) — use R/S · Kandivali West only for the ward-page example |
| Contract | `WS/2023/ROAD/117` |
| Value | ₹4,11,00,000 (₹4.11 crore) |
| Completed | 28 Nov 2023 |
| Warranty ends | 28 Nov 2028 |
| Contractor | withheld in screenshots by default |
| Retrieval date | 14 Jul 2026 |

Label these as sample data in any public-facing use. Do **not** put a real contractor's name into a
generated mockup: an unsourced name in a marketing image is the same violation as an unsourced name
in the product.

---

## 8. Assembling a set

1. Render S01, S02, S04, S08, S23, S18 first — they carry the product story.
2. Fix the seed and re-render the rest with the same capsule and reference image.
3. Assemble in Figma at 1×, replace every text layer with live text, set Marathi variants with a
   real Devanagari font.
4. Run the QA rubric on the assembled sheet, not only on individual screens.
