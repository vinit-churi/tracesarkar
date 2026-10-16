# Review sheet — Marathi and Hindi copy for the share post

**For a native speaker of Marathi or Hindi. No knowledge of the project or of code is needed.**
Reply however is easiest: against each ID, write the correction, or `ok`. You can do one language
and ignore the other.

---

## 1. What you are reading

TraceSarkar is a tool for reporting civic problems in Mumbai — a pothole, an uncollected bin, a
flooded lane, a cracked foot-overbridge. A resident photographs the problem; the platform works out
which municipal department owns it and then **writes the message the resident sends**, so that
sending it is one tap instead of twenty minutes of typing.

That message goes to two readers at once, and it has to work for both:

- **A municipal desk** — a ward officer at the Municipal Corporation of Greater Mumbai, or an
  officer of a state board, who gets hundreds of complaints and acts on the ones that are specific,
  civil and checkable. Bureaucratic language is fine for this reader. Sloppy or angry language is
  not: it gets the message filed and forgotten.
- **The sender's neighbours**, because the message is forwarded on WhatsApp. For them it has to read
  like a neighbour wrote it, not like a form.

The message is built from a fixed set of sentences with blanks in them. The English was written by a
person. **The Marathi and Hindi below were drafted by somebody who speaks neither**, which is why
they need you. They are live today, labelled on screen as unreviewed drafts, and editable by the
sender before they send.

## 2. Rules the copy has to follow

These are not style preferences — they are what keeps the message truthful and usable.

1. **A blank in braces must survive, spelled exactly.** `{ward}`, `{thing}`, `{hours}`, `{date}`,
   `{n}`, `{authority}`, `{citation}`, `{short_ref}`, `{handles}`, `{ward_word}`. They are filled in
   at the moment the message is written. Move them if the sentence reads better, but do not rename,
   translate or drop one.
2. **Digits stay as Latin numerals** — `48`, `2026` — in every language. These are checked against
   government documents that print them that way.
3. **A legal citation is never translated.** `{citation}` and `{short_ref}` are filled with English
   text such as *Bombay High Court order in PIL 85 of 2013*. It stays English inside a Marathi or
   Hindi sentence, which is also how it is said. Please end the sentence **before** the citation
   rather than putting punctuation after it.
4. **Nobody is accused of anything.** The message states what happened and what the record says. It
   never says or implies that anyone is corrupt, negligent, lazy or at fault — not even when it says
   a deadline has passed. *That a deadline has passed* is a fact; *that somebody ignored it* is an
   accusation, and this platform does not make those.
5. **Nothing is softened either.** No pleading, no apology, no "we humbly request". Flat and
   factual.
6. **Shorter is better.** One of the two channels has a hard 280-character window, and everything is
   read on a phone.
7. **The corporation's own word for an administrative ward is विभाग.** **प्रभाग** is an *electoral*
   division and is a different thing — please do not use it, however natural it sounds.

## 3. Marathi

### What it produces today

A pothole in K/E ward, reported to the corporation on 10 September, with the 48-hour deadline
already passed:

```
K/E विभागात खड्डा. प्राधिकरण: BMC.
@mybmc — हा तुमचा विभाग आहे.
अद्याप प्रलंबित.
2 सप्टेंबर 2026 पासून 4 वेळा आढळले.
प्राधिकरणाला 10 सप्टेंबर 2026 रोजी कळवले गेले.
कार्यवाहीसाठी 48 तास. ही मुदत संपली आहे — Bombay High Court order in PIL 85 of 2013
वरील प्रत्येक तपशील, तो ज्या कागदपत्रातून आला आहे त्यासह, या पानावर आहे:
https://tracesarkar.org/i/abc
#KEWard #MumbaiRoads
```

Uncollected garbage in R/C ward, seen once, and **nobody has reported it to the corporation yet**:

```
R/C विभागात न उचललेला कचरा. प्राधिकरण: BMC.
@mybmc — हा तुमचा विभाग आहे.
अद्याप प्रलंबित.
याबद्दल प्राधिकरणाला कळवल्याची नोंद अजून नाही, त्यामुळे कोणतीही मुदत सुरू झालेली नाही.
प्राधिकरणाला कळवल्यापासून कार्यवाहीसाठी 48 तास — Bombay High Court order in PIL 85 of 2013
वरील प्रत्येक तपशील, तो ज्या कागदपत्रातून आला आहे त्यासह, या पानावर आहे:
https://tracesarkar.org/i/abc
#MumbaiWaste #RCWard
```

### The sentences

| ID | What it has to say in English | Marathi draft |
|---|---|---|
| M1 | {thing} in {ward} ward. | `{ward} विभागात {thing}.` |
| M2 | {thing} in {ward} ward, {authority}. | `{ward} विभागात {thing}. प्राधिकरण: {authority}.` |
| M3 | {handles} — this is your ward. | `{handles} — हा तुमचा विभाग आहे.` |
| M4 | Still open. | `अद्याप प्रलंबित.` |
| M5 | The authority has recorded this as attended. No resident has photographed it fixed. | `प्राधिकरणाने याची दखल घेतल्याची नोंद केली आहे. कोणत्याही रहिवाशाने ते दुरुस्त झाल्याचा फोटो दिलेला नाही.` |
| M6 | A resident photographed it fixed. | `एका रहिवाशाने ते दुरुस्त झाल्याचा फोटो दिला आहे.` |
| M7 | Seen again after it was recorded as attended. | `दखल घेतल्याची नोंद झाल्यानंतर पुन्हा आढळले.` |
| M8 | Seen {n} times since {date}. | `{date} पासून {n} वेळा आढळले.` |
| M9 | The authority was told on {date}. | `प्राधिकरणाला {date} रोजी कळवले गेले.` |
| M10 | Nobody has recorded telling the authority about this yet, so no deadline is running. | `याबद्दल प्राधिकरणाला कळवल्याची नोंद अजून नाही, त्यामुळे कोणतीही मुदत सुरू झालेली नाही.` |
| M11 | {hours} hours to attend ({short_ref}). That time has passed. | `कार्यवाहीसाठी {hours} तास ({short_ref}). ही मुदत संपली आहे.` |
| M12 | {hours} hours to attend ({short_ref}). | `कार्यवाहीसाठी {hours} तास ({short_ref}).` |
| M13 | {hours} hours to attend. That time has passed — {citation} | `कार्यवाहीसाठी {hours} तास. ही मुदत संपली आहे — {citation}` |
| M14 | {hours} hours to attend — {citation} | `कार्यवाहीसाठी {hours} तास — {citation}` |
| M15 | {hours} hours to attend once the authority is told — {citation} | `प्राधिकरणाला कळवल्यापासून कार्यवाहीसाठी {hours} तास — {citation}` |
| M16 | Every fact above is on this page, with the document it came from: | `वरील प्रत्येक तपशील, तो ज्या कागदपत्रातून आला आहे त्यासह, या पानावर आहे:` |

### The words for a category

Used when the specific problem has no word of its own. These fill the `{thing}` blank, so they have
to work inside M1 and M2 — *"K/E विभागात ______."*

| ID | English | Marathi draft |
|---|---|---|
| MC1 | Road defect | `रस्त्यातील दोष` |
| MC2 | Uncollected waste | `न उचललेला कचरा` |
| MC3 | Water and drainage | `पाणी आणि मलनिस्सारण` |
| MC4 | Street furniture | `रस्त्यावरील सुविधा` |
| MC5 | Structural | `बांधकामविषयक` |
| MC6 | Environment | `पर्यावरण` |

### The words for a specific problem

Same blank, same test — *"K/E विभागात ______."*

| ID | English | Marathi draft |
|---|---|---|
| MT1 | Pothole | `खड्डा` |
| MT2 | Cracked road | `रस्त्याला तडा` |
| MT3 | Road subsidence | `रस्ता खचला` |
| MT4 | Missing manhole cover | `मॅनहोलचे झाकण गायब` |
| MT5 | Open manhole | `उघडे मॅनहोल` |
| MT6 | Damaged manhole cover | `मॅनहोलचे झाकण तुटले` |
| MT7 | Damaged speed breaker | `तुटलेला गतिरोधक` |
| MT8 | Unrepaired utility dig | `खोदकामानंतर दुरुस्ती नाही` |
| MT9 | Faded road markings | `पुसट झालेले रस्ता पट्टे` |

### Month names

Used in a date such as *2 सप्टेंबर 2026*. The day and year stay as Latin digits.

| ID | | ID | | ID | |
|---|---|---|---|---|---|
| MM1 | `जानेवारी` | MM5 | `मे` | MM9 | `सप्टेंबर` |
| MM2 | `फेब्रुवारी` | MM6 | `जून` | MM10 | `ऑक्टोबर` |
| MM3 | `मार्च` | MM7 | `जुलै` | MM11 | `नोव्हेंबर` |
| MM4 | `एप्रिल` | MM8 | `ऑगस्ट` | MM12 | `डिसेंबर` |

### Questions where we would particularly value your judgement

- **MQ1 · Punctuation.** Every sentence currently ends in a full stop `.`. Should Marathi use `.`
  or the danda `।` here? Please answer for the message as a whole, not sentence by sentence.
- **MQ2 · M8 agreement.** With `{n}` filled in as `2`, `4`, `11`, does `आढळले` still read right, or
  does it need to change with the number? (The line only appears from 2 upward, never with 1.)
- **MQ3 · M1 and the ward code.** `{ward}` is filled with a Latin-script code — `K/E`, `R/C`,
  `F/N`, `P/N`, or a bare letter such as `A` or `T`. Does `{ward} विभागात` join correctly to all of
  those, or does a bare letter need different handling?
- **MQ4 · Register.** Is this the right level overall for a message a resident sends to a ward
  office and also forwards to neighbours — or is it too formal, or not formal enough?

---

## 4. Hindi

### What it produces today

```
K/E वार्ड में गड्ढा. प्राधिकरण: BMC.
@mybmc — यह आपका वार्ड है.
अभी भी लंबित.
2 सितंबर 2026 से 4 बार दर्ज.
प्राधिकरण को 10 सितंबर 2026 को बताया गया.
कार्रवाई के लिए 48 घंटे. यह समय बीत चुका है — Bombay High Court order in PIL 85 of 2013
ऊपर का हर तथ्य, उस दस्तावेज़ के साथ जिससे वह आया है, इस पेज पर है:
https://tracesarkar.org/i/abc
#KEWard #MumbaiRoads
```

### The sentences

| ID | What it has to say in English | Hindi draft |
|---|---|---|
| H1 | {thing} in {ward} ward. | `{ward} वार्ड में {thing}.` |
| H2 | {thing} in {ward} ward, {authority}. | `{ward} वार्ड में {thing}. प्राधिकरण: {authority}.` |
| H3 | {handles} — this is your ward. | `{handles} — यह आपका वार्ड है.` |
| H4 | Still open. | `अभी भी लंबित.` |
| H5 | The authority has recorded this as attended. No resident has photographed it fixed. | `प्राधिकरण ने इसे देख लिए जाने के रूप में दर्ज किया है. किसी निवासी ने इसके ठीक होने की तस्वीर नहीं दी है.` |
| H6 | A resident photographed it fixed. | `एक निवासी ने इसके ठीक होने की तस्वीर दी है.` |
| H7 | Seen again after it was recorded as attended. | `देख लिए जाने के दर्ज होने के बाद फिर दिखा.` |
| H8 | Seen {n} times since {date}. | `{date} से {n} बार दर्ज.` |
| H9 | The authority was told on {date}. | `प्राधिकरण को {date} को बताया गया.` |
| H10 | Nobody has recorded telling the authority about this yet, so no deadline is running. | `इसके बारे में प्राधिकरण को बताए जाने का कोई रिकॉर्ड नहीं है, इसलिए कोई समयसीमा शुरू नहीं हुई है.` |
| H11 | {hours} hours to attend ({short_ref}). That time has passed. | `कार्रवाई के लिए {hours} घंटे ({short_ref}). यह समय बीत चुका है.` |
| H12 | {hours} hours to attend ({short_ref}). | `कार्रवाई के लिए {hours} घंटे ({short_ref}).` |
| H13 | {hours} hours to attend. That time has passed — {citation} | `कार्रवाई के लिए {hours} घंटे. यह समय बीत चुका है — {citation}` |
| H14 | {hours} hours to attend — {citation} | `कार्रवाई के लिए {hours} घंटे — {citation}` |
| H15 | {hours} hours to attend once the authority is told — {citation} | `प्राधिकरण को बताए जाने के बाद कार्रवाई के लिए {hours} घंटे — {citation}` |
| H16 | Every fact above is on this page, with the document it came from: | `ऊपर का हर तथ्य, उस दस्तावेज़ के साथ जिससे वह आया है, इस पेज पर है:` |

### The words for a category

| ID | English | Hindi draft |
|---|---|---|
| HC1 | Road defect | `सड़क में खराबी` |
| HC2 | Uncollected waste | `न उठाया गया कचरा` |
| HC3 | Water and drainage | `पानी और नाली` |
| HC4 | Street furniture | `सड़क की सुविधाएं` |
| HC5 | Structural | `ढांचागत` |
| HC6 | Environment | `पर्यावरण` |

### The words for a specific problem

| ID | English | Hindi draft |
|---|---|---|
| HT1 | Pothole | `गड्ढा` |
| HT2 | Cracked road | `सड़क में दरार` |
| HT3 | Road subsidence | `सड़क धंसी` |
| HT4 | Missing manhole cover | `मैनहोल का ढक्कन गायब` |
| HT5 | Open manhole | `खुला मैनहोल` |
| HT6 | Damaged manhole cover | `मैनहोल का ढक्कन टूटा` |
| HT7 | Damaged speed breaker | `टूटा स्पीड ब्रेकर` |
| HT8 | Unrepaired utility dig | `खुदाई के बाद मरम्मत नहीं` |
| HT9 | Faded road markings | `मिट चुकी सड़क पट्टियां` |

### Month names

| ID | | ID | | ID | |
|---|---|---|---|---|---|
| HM1 | `जनवरी` | HM5 | `मई` | HM9 | `सितंबर` |
| HM2 | `फरवरी` | HM6 | `जून` | HM10 | `अक्तूबर` |
| HM3 | `मार्च` | HM7 | `जुलाई` | HM11 | `नवंबर` |
| HM4 | `अप्रैल` | HM8 | `अगस्त` | HM12 | `दिसंबर` |

### Questions where we would particularly value your judgement

- **HQ1 · Punctuation.** Every sentence currently ends in `.`. Should these use the danda `।`?
- **HQ2 · HM10.** `अक्तूबर` or `अक्टूबर`?
- **HQ3 · HC2 / H1.** `न उठाया गया कचरा` reads to us like a translation rather than a sentence.
  Does something like `कचरा उठाया नहीं गया` read better — and if so, can it still fill the `{thing}`
  blank in *"R/C वार्ड में ______."*?
- **HQ4 · H7** is the line we are least confident about.
- **HQ5 · Register.** Mumbai Hindi, for a ward office and for neighbours on WhatsApp. Is this it, or
  is it drifting Sanskritised?

---

## 5. Words that do not exist yet in either language

The platform can classify a photograph into any of the 36 problems below. Only **9** of them have a
Marathi or Hindi word — all 9 are road defects, above. For the other **27**, the message falls back
to the category word, so a felled tree currently reads *"H/W विभागात पर्यावरण."* — "Environment in
H/W ward", which says nothing useful.

**This is a bigger job than checking the tables above and is entirely optional.** If you have the
time, the English column is the source text; one or two words per row is plenty. Each one fills the
same blank — *"K/E विभागात ______."* / *"K/E वार्ड में ______."*

| ID | English | Marathi | Hindi |
|---|---|---|---|
| W1 | Uncollected garbage | | |
| W2 | Overflowing bin | | |
| W3 | Illegal dumping | | |
| W4 | Dumped debris | | |
| W5 | Burning waste | | |
| W6 | Dead animal | | |
| W7 | Waterlogging | | |
| W8 | Blocked drain | | |
| W9 | Silted nallah | | |
| W10 | Sewage overflow | | |
| W11 | Burst water pipeline | | |
| W12 | Water leakage | | |
| W13 | Street light not working | | |
| W14 | Broken footpath | | |
| W15 | Damaged railing | | |
| W16 | Damaged signage | | |
| W17 | Missing bollard | | |
| W18 | Cracked bridge or foot-overbridge | | |
| W19 | Collapsed wall | | |
| W20 | Distressed building | | |
| W21 | Unsafe scaffolding | | |
| W22 | Tree felling | | |
| W23 | Air pollution source | | |
| W24 | Effluent discharge | | |
| W25 | Dumping in a creek | | |
| W26 | Mangrove destruction | | |
| W27 | Encroachment on a water body | | |

Two notes on this table. **W9 "nallah"** — the English keeps the local word on purpose; please use
whatever a resident and a ward office would actually say. **W18** is one entry covering both a
road bridge and a pedestrian foot-overbridge, because a photograph does not always distinguish them.

---

## 6. Where these live, for whoever applies the answers

| File | What it holds |
|---|---|
| [`internal/share/i18n/share.mr.json`](../../internal/share/i18n/share.mr.json) | Everything in §3 |
| [`internal/share/i18n/share.hi.json`](../../internal/share/i18n/share.hi.json) | Everything in §4 |
| [`internal/share/i18n/share.en.json`](../../internal/share/i18n/share.en.json) | The English the two are checked against |

Inside each file, §3's sentences are under `sentences`, the category words under `categories`, the
specific problems under `things`, and the month names under `months`.

When a language has been reviewed, change `"review": "pending"` to `"reviewed"` in that file and
fill in `reviewed_by` and `reviewed_on`. That is what removes the "not yet read by a native speaker"
notice from the screen, and nothing else does
([D137](../00-overview/05-decision-log.md), [ADR 0020](../04-adr/0020-share-text-is-generated-not-written-by-a-model.md)).

A separate and smaller review, only relevant if the model-assisted rewording is ever switched on, is
the per-language guidance in
[`internal/polish/prompts/share_polish_v1.md`](../../internal/polish/prompts/share_polish_v1.md)
and the `register` map in
[`internal/polish/prompt.go`](../../internal/polish/prompt.go). Those are instructions *about*
Marathi and Hindi, written by the same non-speaker, and they deserve the same reading
([ADR 0021](../04-adr/0021-a-model-may-reword-a-generated-post-never-author-one.md)).
