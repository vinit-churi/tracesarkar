# Marathi and Hindi copy for the share post — review record

**Reviewed and applied on 16 October 2026.** Both languages are now
`"review": "reviewed"`, the "not yet read by a native speaker" notice no longer appears on any
post, and all 36 problems have a word in both languages
([D175](../00-overview/05-decision-log.md)).

This file is kept for two reasons: it records *why* the copy reads the way it does, which a later
reviewer needs, and sections 1 and 2 are the brief to reuse for the next language.

---

## 0. The verdict

The reviewer's summary was that **the Marathi draft was close to the mark and the Hindi draft
suffered from "Google Translate syndrome"** — over-relying on `प्राधिकरण` and on clunky passive
phrasing such as `देख लिए जाने के रूप में`. That is visible in the counts: 19 of 43 Marathi strings
changed and 20 of 43 Hindi ones, but the Marathi changes are mostly register and the Hindi ones are
mostly grammar.

Four themes run through the corrections, and they are worth knowing before reading the tables.

- **`प्राधिकरण` is not a word people use.** Nobody in Mumbai says *प्राधिकरण को बताया गया*. The
  natural form is `संबंधित विभाग` / `संबंधित विभागास` — "the concerned department" — with
  `यंत्रणा:` (Marathi) and `विभाग:` (Hindi) as the label beside the authority's name.
- **A category word needs `समस्या`.** `बांधकामविषयक` is an adjective and `पर्यावरण` is an abstract
  noun; neither reads as a problem inside *"{ward} विभागात {thing}."*. Without it, *"K/E वार्ड में
  सड़क की सुविधाएं"* reads as a positive announcement — "Street facilities in K/E ward!".
- **"Repaired" does not fit most of what this platform reports.** `ते दुरुस्त झाल्याचा फोटो` makes
  sense for a pothole or a railing and is nonsense for uncollected garbage or a felled tree. Both
  languages now say "a photo of the work being completed" — `काम झाल्याचा` / `काम पूरा होने` —
  which also sidesteps the gender agreement problem described in rule 8 below.
- **The English relative clauses were translated literally and should not have been.**
  `तो ज्या कागदपत्रातून आला आहे त्यासह` is a word-for-word rendering of "with the document it came
  from" and is very clunky; `मूळ कागदपत्रांसह` says the same thing in a third of the length.

### Answers to the questions a non-speaker could not settle

| | Question | Answer |
|---|---|---|
| **MQ1** | Does Marathi want `.` or the danda `।`? | **`.`** Modern Marathi — newspapers, government circulars, WhatsApp — uses the full stop exclusively. The danda belongs to Sanskrit or old poetry and would look out of place. |
| **MQ2** | Does `आढळले` agree with `{n}`? | **No change needed.** It is an impersonal neuter construction, "it was spotted", and reads naturally at 2, 4 or 11 regardless of the gender of `{thing}`, because it sits in a standalone sentence separated from the subject. |
| **MQ3** | Does `{ward} विभागात` join to a bare letter? | **Yes.** `K/E विभागात`, `A विभागात`, `T विभागात` all join cleanly. The code is Latin script with a space before `विभागात`, so no connector or suffix handling is needed. |
| **MQ4** | Is the register right? | **Close to ideal.** Civic and administrative Marathi — `प्रलंबित`, `दखल`, `मुदत`, `कार्यवाही` — is widely understood and used by ordinary residents on WhatsApp when the subject is BMC. The only systematic fix was `प्राधिकरण`. |
| **HQ1** | Does Hindi want `.` or `।`? | **Either, consistently.** `।` is technically the correct standard-Hindi full stop; `.` is very common in urban Mumbai Hindi, especially mixed with Latin ward codes, numbers and URLs, and saves space. **`.` kept**, so the platform has one convention across both languages. |
| **HQ2** | `अक्तूबर` or `अक्टूबर`? | **`अक्टूबर`.** Overwhelmingly preferred in modern Indian media and everyday use; `अक्तूबर` is the older NCERT standard. |
| **HQ3** | Is `न उठाया गया कचरा` a translation rather than a sentence? | **Yes.** `कचरा नहीं उठाया गया` — "garbage has not been picked up" — works seamlessly inside the blank: *"R/C वार्ड में कचरा नहीं उठाया गया."* |
| **HQ4** | Is H7 as bad as it looks? | **Worse — "virtually unreadable".** Now `कार्रवाई दर्ज होने के बाद यह समस्या फिर दिखी.` |
| **HQ5** | Is the Hindi register right? | **No.** Drifting too Sanskritised in places (`प्राधिकरण`, `तथ्य`) and too literal in others (`देख लिए जाने के रूप में`). Both are fixed below. |

### One value that departs from the review

**`effluent discharge` in Hindi** is `दूषित पानी छोड़ना` and not the reviewer's
`गंदा रासायनिक पानी छोड़ना`. A photograph of a discharge does not establish that it is *chemical*,
and this platform does not state more than its records support (rule 4). The Marathi
(`दूषित सांडपाणी सोडणे`) makes no such claim either, so the two now match. Everything else is the
review as given, taking the reviewer's first option wherever alternatives were offered.

---

## 1. What this copy is for

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

The message is built from a fixed set of sentences with blanks in them, and it is editable by the
sender before they send it.

## 2. Rules the copy has to follow

These are not style preferences — they are what keeps the message truthful and usable. Rules 1 to 7
were in the brief the reviewer answered; **rule 8 is the reviewer's own addition** and should be in
any future brief.

1. **A blank in braces must survive, spelled exactly.** `{ward}`, `{thing}`, `{hours}`, `{date}`,
   `{n}`, `{authority}`, `{citation}`, `{short_ref}`, `{handles}`, `{ward_word}`. They are filled in
   at the moment the message is written. Move them if the sentence reads better, but do not rename,
   translate or drop one.
2. **Digits stay as Latin numerals** — `48`, `2026` — in every language. These are checked against
   government documents that print them that way.
3. **A legal citation is never translated**, and the sentence ends **before** it rather than putting
   punctuation after it. `{citation}` is filled with English text such as *Bombay High Court order
   in PIL 85 of 2013*, which stays English inside a Marathi or Hindi sentence.
4. **Nobody is accused of anything.** The message states what happened and what the record says. It
   never says or implies that anyone is corrupt, negligent, lazy or at fault — not even when it says
   a deadline has passed. *That a deadline has passed* is a fact; *that somebody ignored it* is an
   accusation. This is also why `distressed building` is `जीर्ण झालेली इमारत` and not
   `धोकादायक इमारत`: a photograph does not establish that a building is dangerous.
5. **Nothing is softened either.** No pleading, no apology, no "we humbly request". Flat and
   factual.
6. **Shorter is better.** One of the two channels has a hard 280-character window, and everything is
   read on a phone.
7. **The corporation's own word for an administrative ward is विभाग.** **प्रभाग** is an *electoral*
   division and is a different thing — do not use it here, however natural it sounds.
8. **Watch the grammar of `{thing}`.** In English "photographed it fixed" uses the genderless
   pronoun "it". Marathi and Hindi inflect for gender and number (`ते` / `तो` / `ती`), and
   "repaired" (`दुरुस्त`) only makes sense for a broken object — not for uncollected garbage or a
   felled tree. A universal noun phrase such as `काम झाल्याचा फोटो` / `काम पूरा होने की तस्वीर`
   sidesteps both problems at once.

Two notes from the reviewer on the brief itself: calling out विभाग vs प्रभाग (rule 7) and the
no-punctuation-after-`{citation}` rule (rule 3) **saved a round trip**, because a native speaker
would otherwise instinctively try to "fix" both.

---

## 3. What changed

### Marathi — what changed

| ID | Was | Now |
|---|---|---|
| M1 | `{ward} {ward_word}ात {thing}.` | `{ward} विभागात {thing}.` |
| M2 | `{ward} {ward_word}ात {thing}. प्राधिकरण: {authority}.` | `{ward} विभागात {thing}. यंत्रणा: {authority}.` |
| M3 | `{handles} — हा तुमचा {ward_word} आहे.` | `{handles} — हा तुमचा विभाग आहे.` |
| M5 | `प्राधिकरणाने याची दखल घेतल्याची नोंद केली आहे. कोणत्याही रहिवाशाने ते दुरुस्त झाल्याचा फोटो दिलेला नाही.` | `याची दखल घेतल्याची नोंद आहे. पण काम झाल्याचा फोटो कोणत्याही रहिवाशाने दिलेला नाही.` |
| M6 | `एका रहिवाशाने ते दुरुस्त झाल्याचा फोटो दिला आहे.` | `काम झाल्याचा फोटो एका रहिवाशाने दिला आहे.` |
| M7 | `दखल घेतल्याची नोंद झाल्यानंतर पुन्हा आढळले.` | `दखल घेतल्याच्या नोंदीनंतरही समस्या पुन्हा आढळली.` |
| M9 | `प्राधिकरणाला {date} रोजी कळवले गेले.` | `संबंधित विभागास {date} रोजी कळवण्यात आले.` |
| M10 | `याबद्दल प्राधिकरणाला कळवल्याची नोंद अजून नाही, त्यामुळे कोणतीही मुदत सुरू झालेली नाही.` | `याची माहिती संबंधित विभागास दिल्याची नोंद अजून नाही, त्यामुळे कोणतीही मुदत सुरू झालेली नाही.` |
| M15 | `प्राधिकरणाला कळवल्यापासून कार्यवाहीसाठी {hours} तास — {citation}` | `संबंधित विभागास कळवल्यापासून कार्यवाहीसाठी {hours} तास — {citation}` |
| M16 | `वरील प्रत्येक तपशील, तो ज्या कागदपत्रातून आला आहे त्यासह, या पानावर आहे:` | `वरील सर्व माहिती मूळ कागदपत्रांसह या पानावर आहे:` |
| MC1 | `रस्त्यातील दोष` | `रस्त्याची समस्या` |
| MC3 | `पाणी आणि मलनिस्सारण` | `पाणी आणि सांडपाणी समस्या` |
| MC4 | `रस्त्यावरील सुविधा` | `रस्त्यावरील नागरी सुविधांची समस्या` |
| MC5 | `बांधकामविषयक` | `बांधकामविषयक समस्या` |
| MC6 | `पर्यावरण` | `पर्यावरणीय समस्या` |
| MT2 | `रस्त्याला तडा` | `रस्त्याला तडे` |
| MT3 | `रस्ता खचला` | `खचलेला रस्ता` |
| MT6 | `मॅनहोलचे झाकण तुटले` | `मॅनहोलचे तुटलेले झाकण` |
| MT9 | `पुसट झालेले रस्ता पट्टे` | `रस्त्यावरील पुसट झालेले पट्टे` |

### Hindi — what changed

| ID | Was | Now |
|---|---|---|
| H1 | `{ward} {ward_word} में {thing}.` | `{ward} वार्ड में {thing}.` |
| H2 | `{ward} {ward_word} में {thing}. प्राधिकरण: {authority}.` | `{ward} वार्ड में {thing}. विभाग: {authority}.` |
| H3 | `{handles} — यह आपका {ward_word} है.` | `{handles} — यह आपका वार्ड है.` |
| H5 | `प्राधिकरण ने इसे देख लिए जाने के रूप में दर्ज किया है. किसी निवासी ने इसके ठीक होने की तस्वीर नहीं दी है.` | `विभाग के रिकॉर्ड में इस पर ध्यान दिया गया है. किसी भी निवासी ने काम पूरा होने की तस्वीर नहीं दी है.` |
| H6 | `एक निवासी ने इसके ठीक होने की तस्वीर दी है.` | `एक निवासी ने काम पूरा होने की तस्वीर दी है.` |
| H7 | `देख लिए जाने के दर्ज होने के बाद फिर दिखा.` | `कार्रवाई दर्ज होने के बाद यह समस्या फिर दिखी.` |
| H8 | `{date} से {n} बार दर्ज.` | `{date} से {n} बार दिखा.` |
| H9 | `प्राधिकरण को {date} को बताया गया.` | `संबंधित विभाग को {date} को सूचित किया गया.` |
| H10 | `इसके बारे में प्राधिकरण को बताए जाने का कोई रिकॉर्ड नहीं है, इसलिए कोई समयसीमा शुरू नहीं हुई है.` | `संबंधित विभाग को इसकी सूचना दिए जाने का अभी कोई रिकॉर्ड नहीं है, इसलिए कोई समयसीमा शुरू नहीं हुई है.` |
| H15 | `प्राधिकरण को बताए जाने के बाद कार्रवाई के लिए {hours} घंटे — {citation}` | `संबंधित विभाग को सूचित किए जाने के बाद कार्रवाई के लिए {hours} घंटे — {citation}` |
| H16 | `ऊपर का हर तथ्य, उस दस्तावेज़ के साथ जिससे वह आया है, इस पेज पर है:` | `ऊपर दी गई हर जानकारी, मूल दस्तावेज़ों के साथ, इस पेज पर है:` |
| HC2 | `न उठाया गया कचरा` | `कचरा नहीं उठाया गया` |
| HC3 | `पानी और नाली` | `पानी और ड्रेनेज की समस्या` |
| HC4 | `सड़क की सुविधाएं` | `सड़क की सुविधाओं में खराबी` |
| HC5 | `ढांचागत` | `ढांचागत समस्या` |
| HC6 | `पर्यावरण` | `पर्यावरण संबंधी समस्या` |
| HT3 | `सड़क धंसी` | `धंसी हुई सड़क` |
| HT6 | `मैनहोल का ढक्कन टूटा` | `मैनहोल का टूटा ढक्कन` |
| HT9 | `मिट चुकी सड़क पट्टियां` | `सड़क की धुंधली पट्टियां` |
| HM10 | `अक्तूबर` | `अक्टूबर` |

---

## 4. The 27 problems that had no word in any language

The platform can classify a photograph into any of 36 problems. Before this review, **9** had a
Marathi or Hindi word — all 9 road defects — and the other 27 fell back to the category word, so a
felled tree composed as *"H/W विभागात पर्यावरण."* — "Environment in H/W ward", which says nothing.
It now composes as *"H/W विभागात झाड तोडणे."*

Each of these was checked by the reviewer inside both frames: *"K/E विभागात ______."* and
*"K/E वार्ड में ______."*

| ID | English | Marathi | Hindi |
|---|---|---|---|
| W1 | Uncollected garbage | `न उचललेला कचरा` | `कचरा नहीं उठाया गया` |
| W2 | Overflowing bin | `ओसंडून वाहणारी कचराकुंडी` | `भरा हुआ कूड़ेदान` |
| W3 | Illegal dumping | `बेकायदेशीर कचरा फेकणे` | `अवैध रूप से कचरा फेंकना` |
| W4 | Dumped debris | `टाकलेला राडारोडा` | `फेंका गया मलबा` |
| W5 | Burning waste | `कचरा जाळणे` | `जलता हुआ कचरा` |
| W6 | Dead animal | `मृत प्राणी` | `मृत जानवर` |
| W7 | Waterlogging | `पाणी साचले` | `जलभराव` |
| W8 | Blocked drain | `तुंबलेली गटार` | `जाम नाली` |
| W9 | Silted nallah | `गाळाने भरलेला नाला` | `गाद से भरा नाला` |
| W10 | Sewage overflow | `सांडपाणी रस्त्यावर वाहणे` | `सीवेज का पानी बहना` |
| W11 | Burst water pipeline | `फुटलेली पाण्याची पाईपलाईन` | `पानी की फूटी पाइपलाइन` |
| W12 | Water leakage | `पाणी गळती` | `पानी का रिसाव` |
| W13 | Street light not working | `बंद पडलेला पथदिवा` | `बंद पड़ी स्ट्रीट लाइट` |
| W14 | Broken footpath | `तुटलेला फुटपाथ` | `टूटा हुआ फुटपाथ` |
| W15 | Damaged railing | `तुटलेली रेलिंग` | `टूटी हुई रेलिंग` |
| W16 | Damaged signage | `तुटलेला दिशादर्शक फलक` | `टूटा हुआ साइनबोर्ड` |
| W17 | Missing bollard | `गायब असलेला खांब (बोलार्ड)` | `गायब बोलार्ड (खंभा)` |
| W18 | Cracked bridge or foot-overbridge | `पुलाला किंवा पादचारी पुलाला तडे` | `पुल या फुट-ओवरब्रिज में दरार` |
| W19 | Collapsed wall | `कोसळलेली भिंत` | `गिरी हुई दीवार` |
| W20 | Distressed building | `जीर्ण झालेली इमारत` | `जर्जर इमारत` |
| W21 | Unsafe scaffolding | `असुरक्षित परातणी (स्कॅफोल्डिंग)` | `असुरक्षित मचान (स्कैफोल्डिंग)` |
| W22 | Tree felling | `झाड तोडणे` | `पेड़ की कटाई` |
| W23 | Air pollution source | `हवा प्रदूषणाचा स्रोत` | `वायु प्रदूषण का स्रोत` |
| W24 | Effluent discharge | `दूषित सांडपाणी सोडणे` | `दूषित पानी छोड़ना` |
| W25 | Dumping in a creek | `खाडीत कचरा किंवा राडारोडा टाकणे` | `खाड़ी में कचरा या मलबा फेंकना` |
| W26 | Mangrove destruction | `कांदळवनाची कत्तल` | `मैंग्रोव की कटाई` |
| W27 | Encroachment on a water body | `जलस्रोतावर अतिक्रमण` | `जलस्रोत पर अतिक्रमण` |

Four notes on the Mumbai vocabulary in that table, from the reviewer:

- **W4 `राडारोडा`** is the universal Marathi word for construction and demolition debris, used by
  BMC and residents alike.
- **W20** — "distressed" is the engineering term, and BMC's own C1/C2A/C2B classification and
  resident complaints say `धोकादायक इमारत` (dangerous) or `जीर्ण इमारत` (dilapidated). The neutral
  one is used here, for the reason in rule 4.
- **W21 `स्कॅफोल्डिंग`** keeps the transliterated English in brackets because residents and ward
  staff say "scaffolding" even when speaking Marathi or Hindi.
- **W26 `कांदळवन`** is the official and widely publicised Maharashtra term for mangroves (also
  `खारफुटी`). Mumbai Hindi and local Hindi newspapers simply write `मैंग्रोव`.

---

## 5. Where this lives

| File | What it holds |
|---|---|
| [`internal/share/i18n/share.mr.json`](../../internal/share/i18n/share.mr.json) | The Marathi copy. Its `note` field records the four decisions above |
| [`internal/share/i18n/share.hi.json`](../../internal/share/i18n/share.hi.json) | The Hindi copy, and the one deviation from the review |
| [`internal/share/i18n/share.en.json`](../../internal/share/i18n/share.en.json) | The English the other two are checked against |

Inside each file the sentences are under `sentences`, the category words under `categories`, the
specific problems under `things`, and the month names under `months`. `review`, `reviewed_by` and
`reviewed_on` are what the API and the issue page read to decide whether to show the draft notice —
and nothing else does
([D137](../00-overview/05-decision-log.md), [ADR 0020](../04-adr/0020-share-text-is-generated-not-written-by-a-model.md)).

**Still unreviewed:** the per-language guidance a model is given when the optional rewording layer
is switched on, in
[`internal/polish/prompts/share_polish_v1.md`](../../internal/polish/prompts/share_polish_v1.md)
and the `register` map in
[`internal/polish/prompt.go`](../../internal/polish/prompt.go). Those are instructions *about*
Marathi and Hindi written by a non-speaker, and at least one of them is now known to be wrong: it
tells the model to use the danda in Hindi and says nothing about Marathi's full stop, which MQ1 and
HQ1 settle the other way ([ADR 0021](../04-adr/0021-a-model-may-reword-a-generated-post-never-author-one.md)).
