# RTI 1 — the pothole circulars and BMC's own repair timeline

**Status:** ready to file · **Drafted:** 29 September 2026 · **Route corrected:** 29 September 2026
**Filed:** _(record the date and registration number here after filing)_

---

## Correction: BMC is not on the state RTI portal

An earlier draft of this file said to file at `rtionline.maharashtra.gov.in`. **That is wrong.**

That portal serves a fixed list of authorities and names them on its front page. Checked
29 September 2026: it carries **26 municipal bodies** — Thane, Navi Mumbai, Kalyan-Dombivli,
Vasai-Virar, Panvel, Mira-Bhayandar, Pune, Nagpur and others — and **Brihanmumbai Municipal
Corporation is not among them.** The two "Greater Mumbai" entries on that page are the Deputy
Director of Town Planning and the State Information Commission, neither of which is BMC.

BMC runs its own RTI system instead, and it is filed with directly.

---

## How to file

### The online form does not serve Borivali

BMC's RTI Request Form at `portal.mcgm.gov.in` → RTI → **RTI Application** does load in a real
browser. It is unusable for this RTI for a simpler reason: checked 29 September 2026, its
mandatory **Ward** dropdown offers exactly two options — **KW Ward** and **MW Ward**. K/West is
Andheri West and Juhu; M/West is Chembur.

**R/Central is not on it.** BMC's online RTI covers 2 of its 24 wards. It is a pilot, not a
service.

Do not select K/W as a workaround. That PIO has no jurisdiction over Borivali roads and would have
to transfer the application under Section 6(3) — five days lost, and a real chance of it being
dropped in the handoff.

### The route that works: on paper, to the PIO

Filing on paper is always valid under Section 6(1) and is what this draft assumes.

**Address it to:**

```
The Public Information Officer
Assistant Engineer (Roads) R/Central
Office of the Dy. Ch. Eng. Roads (Western Suburbs)
5th Floor, P/S Ward Office Building
S. V. Road, Goregaon (West)
Mumbai 400062
```

**Fee:** a **₹10 court-fee stamp** affixed to the application is the standard method for a
Maharashtra public authority. Our own research has the amount contested — ₹10 is long-standing, a
₹30 figure circulates with no gazette notification found (open question Q2). **Take a photograph of
the stamped application before submitting**; that photograph settles Q2 either way.

**Submit** in person at the office above, or by **registered post with acknowledgement due**. Keep
the AD card or the counter receipt: the 30-day clock under Section 7(1) runs from the date of
receipt, and the receipt is what proves it.

---

## If a form limits you to 150 words

Some RTI forms cap the "information sought" field. This version fits in about 145 words. Attach the
long version below as an annexure and reference it.

```
Under Section 6(1) of the Right to Information Act, 2005, please provide:

1. Copies of Circular MGC/F/1074 dated 06.07.2013, Circular CA/FDT/59 dated
   16.03.2013, and Circular CA/FRD/7 dated 17.05.2013, cited under the heading
   "Pothole" in the Section 4(1)(b)(v) rules table of the RTI manual of the
   Maintenance Department, R/Central Ward.

2. A copy of any circular or office order issued after 17.05.2013 and in force
   today which prescribes the time within which a reported pothole is to be
   attended to.

3. A blank copy of the pro forma "Daily Report of Potholes and Bad Patches",
   listed as a record held in the RTI manual of the Chief Engineer
   (Roads and Traffic).

4. The criteria by which a road in R/Central ward is classified as a major road
   or a minor road for maintenance purposes.

Please supply the information in electronic form to the email address given.
```

---

## The full application

For the paper route, where no word limit applies.

```
To,
The Public Information Officer
Assistant Engineer (Roads) R/Central
Office of the Dy. Ch. Eng. Roads (Western Suburbs)
5th Floor, P/S Ward Office Building, S. V. Road
Goregaon (West), Mumbai 400062

Subject: Application under Section 6(1) of the Right to Information Act, 2005 —
         circulars governing pothole repair and the prescribed timeline for
         attending potholes in R/Central ward

Sir / Madam,

I am a citizen of India and seek the following information under the Right to
Information Act, 2005.

1. A copy of Circular no. MGC/F/1074 dated 06.07.2013.

2. A copy of Circular no. CA/FDT/59 dated 16.03.2013.

3. A copy of Circular no. CA/FRD/7 dated 17.05.2013.

   (All three are cited under the heading "Pothole" in the rules table at
   Section 4(1)(b)(v) of the RTI manual published by the Maintenance
   Department, R/Central Ward, on the MCGM portal.)

4. A copy of any circular, office order or standing instruction issued after
   17.05.2013 and in force on the date of this application, which prescribes
   the time within which a reported pothole is to be attended to.

5. A blank copy of the pro forma titled "Daily Report of Potholes and Bad
   Patches", listed as a record held in the RTI manual of the Chief Engineer
   (Roads and Traffic).

6. The number of pothole complaints received for R/Central ward between
   1 June 2026 and 30 September 2026, and the number recorded as attended.

7. A copy of the criteria by which a road in R/Central ward is classified as a
   major road or a minor road for maintenance purposes.

I request the information in electronic form, to the email address below.
A court-fee stamp of the prescribed amount is affixed.

Yours faithfully,

[ your name ]
[ your address, Borivali West ]
[ your mobile ]
[ your email ]

Date:  [ date ]
Place: Mumbai
```

---

## Why it is worded this way

Each question asks for **a named document or a count** — the form that survives. Questions inviting
opinion, inference or a reason are refused under the Act, and a refusal costs a month.

- **The circulars are asked for by number.** BMC cited them itself, so it cannot claim not to
  identify them.
- **Question 4 catches supersession.** A 2013 circular may have been replaced; asking for "any
  circular in force" closes that.
- **Question 7 matters most for the product.** The major/minor split decides which of the two
  departments owns a given pothole. The platform currently has to state both, because it cannot
  tell them apart.
- **Seven questions.** Long omnibus applications invite rejection.

---

## Why this RTI at all

The department mapping did **not** need one — BMC publishes it in its Section 4(1)(b) handbooks,
and R/C's road-defect owners are recorded with sources in `authority_departments`.

What is missing is **BMC's own pothole timeline**. The 48 hours the platform shows comes from the
Bombay High Court's October 2025 direction — a court order *on* the corporation, not the
corporation's internal operating rule. R/Central's own manual cites three circulars under a
"Pothole" heading and publishes none of them. If BMC's internal timeline is shorter than 48 hours,
the shorter one binds.

---

## What to do with the reply

| Outcome | Next step |
|---|---|
| Circulars supplied with a timeline | Record in `legal_constants` with citation and effective-from date. If it differs from 48 hours, **both** apply and the shorter binds |
| Refused, or no reply in 30 days | First appeal to the **Executive Engineer (Roads) K/E, R/C & R/N**. Section 7(1): silence past 30 days is deemed refusal, and the first appeal is free |
| "Information not held" | Useful in itself. It means BMC has no internal pothole timeline and the High Court's 48 hours stands alone — record that |
| The fee receipt | Photograph it regardless. It closes open question Q2 |

**Record the registration number here when you file.** The 30-day clock starts on receipt, and the
deadline wallet (S35) will need it.
