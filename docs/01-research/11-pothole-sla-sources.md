# What actually obliges BMC to fix a pothole, and by when

**Researched 29 September 2026.** Every item below was verified first-hand and archived.

The platform shows a 48-hour clock. This records where that comes from, what else binds, and three
traps that would otherwise be walked into.

---

## 1. The court order — binding, and the strongest

**48 hours**, from the Bombay High Court in PIL 71/2013, order of 13 October 2025, paragraph
70(ix). Neutral citation `2025:BHC-OS:18736-DB`. Reaffirmed verbatim on 6 July 2026.

> All potholes … shall be attended to forthwith and, in any event, within forty-eight hours

**Retrieval trap, and it is a real one.** The order spells it **"forty-eight hours"**, in words.
Any watcher grepping for `48 hours` finds nothing. Worse, paragraph 70(iv) uses the same phrase for
an unrelated police-notification duty, so a naive word match returns the wrong paragraph. Match on
the paragraph number and the surrounding phrase, not the interval alone.

---

## 2. BMC's own contract — 24 hours, with a penalty

Found while looking for something else, and more useful than what we were looking for.

BMC's published road tender **ETH_8000040832** (archived
`archive/bmc_portal/2026-09-29/road-tender-ETH_8000040832-7b29526ff6db.pdf`, source
`https://www.mcgm.gov.in/irj/go/km/docs/documents/Tenders/ETH/ETH_8000040832_020523.pdf`) carries,
verbatim:

| Clause | What it says |
|---|---|
| §10.11 | "All Defective work must be rectified **within 24 hours** of intimation on Whatsapp Messenger or Telegram Messenger or Telephone or Email or by memo in writing" |
| §10.73 | "during monsoon period if any pothole / settled trench is observed, it shall be binding on Contractors to attend it **within 24 / 48 hours** as the case may be" |
| §7.4 | "A penalty of **Rs. 5000/- per day per Defect/Pothole/Trench** if not attended in stipulated time" |

Three things follow, and all three matter to the product.

**The contractor's deadline is shorter than the court's.** 24 hours, contractually, during the
defect liability period. Where a stretch is under contract, that is the number that binds the
party who has to hold the shovel.

**A WhatsApp message is valid intimation.** BMC wrote that into its own contract. The clock can be
started by a channel a citizen already uses — which is exactly the share-kit surface the product
already plans.

**There is a priced consequence.** ₹5,000 per day per pothole is a fact about money, from a public
document, with a source. That is the kind of fact this platform exists to surface.

---

## 3. The "24 hours from the Commissioner" story — not a document

Press reported in June 2026 that BMC would resolve pothole complaints within 24 hours. Traced: it
was a **verbal instruction at a review meeting on 19 June 2026** by the Additional Municipal
Commissioner (Projects). No circular number exists. Aggregator sites attributing it to the
Municipal Commissioner conflict with the first-tier report.

**Do not record it as a constant.** It is context, not an obligation.

**Separate red herring:** `HomePage Data/24 Hours GR dt 01.10.2025.pdf` on BMC's homepage is a
Shops and Establishments Act circular about 24-hour trading. Nothing to do with roads.

---

## 4. Major vs minor road — seasonal, not width

This corrects an assumption in the department mapping.

From the Chief Engineer (Roads & Traffic) manual, chapter 2 §2.5(iv): **outside the monsoon the
ward maintains both major and minor roads. During the monsoon, major roads and bus routes pass to
contractors supervised by central Roads & Traffic zonal staff, while minor roads stay with the
ward.**

So **which department owns a defect changes with the date.** The two rows in
`authority_departments` for R/C are both correct, and which applies depends on when the report was
made.

**There is no published width threshold.** Classification is per-road enumerated lists inside each
ward's manual, and width does not separate them reliably — H/West lists a 19.6 m road as minor and
10 m roads as major. **Do not derive jurisdiction from road width.**

---

## 5. The three circulars: confirmed unobtainable

`MGC/F/1074` (06.07.2013), `CA/FDT/59` (16.03.2013), `CA/FRD/7` (17.05.2013) exist publicly only as
bare citations in BMC's ward RTI manuals — 15 of the 17 ward manuals carry the identical
boilerplate table, with the remarks column empty in every one.

Searched and not found: Indian Kanoon (control-tested — `MGC/F` returns 16 genuine BMC Commissioner
orders, so slashed references do index), Google, Bing, BMC's public circulars page (six items, all
HR), Wayback's index of `/documents/Circulars/` (102 files, none road-related), and the PIL 71/2013
judgments.

**Residual uncertainty:** an engineering-circular folder does exist at
`.../Roads and Traffic/Docs/Circulars/` — it hosts, for example, a footpath construction guideline.
Directory listing is off and filename guesses failed. A pothole circular may be sitting there,
unreachable.

**Conclusion: the RTI is the only route to their text**, and is
[drafted](../06-operations/rti-01-pothole-sla-circulars.md). Its urgency is now lower, because §2
above already gives a citable BMC deadline with a penalty.

---

## What is in `legal_constants`

Loaded 3 October 2026 by migration 0019. The table had been empty until then,
which meant every clock the product showed was either absent or a literal
somewhere — the thing [hard rule 6](../../CLAUDE.md) exists to prevent.

| Key | Value | Citation | Note |
|---|---|---|---|
| `pothole.attend_hours.court` | 48 | `2025:BHC-OS:18736-DB` para 70(ix), 13 Oct 2025 | Binds the corporation, from notification |
| `pothole.attend_hours.contract` | 24 | BMC tender ETH_8000040832 §10.11 | Binds the contractor in DLP, from intimation |
| `pothole.penalty_per_day_paise` | 500000 | Same tender §7.4 | ₹5,000/day per pothole |
| `waste.refuse_removal_hours` | 24 | R/C SWM RTI manual §4(1)(b)(iii) | Sweeping and refuse removal |
| `waste.silt_debris_removal_hours` | 24 | Same manual | Silt and debris removal |

### Waste has its own published deadline, and it is BMC's own

Found 3 October 2026 while mapping departments, in BMC's R/Central Solid Waste
Management RTI handbook, Section 4(1)(b)(iii):

| Activity | Time limit | Basis |
|---|---|---|
| Sweeping of roads & removal of refuse | **Within 24 hours** | MMC Act 1888 §365(a); circular DMC/ENV SWM/4345 dt. 16.03.2006 |
| Removal of silt & debris | **Within 24 hours** | MMC Act 1888 §375(A); same circular |

This is the garbage equivalent of the pothole clock and it needed no RTI — BMC
publishes it. The circular reference is worth requesting in its own right.

### Water supply has no published repair deadline, and that is a finding

The R/Central Water Works handbook publishes time limits for **granting a
connection, reading a meter and disconnecting a supply** — fifteen days, ten
days, four days. It publishes **none for repairing a leak or a burst main**.
Its "Public Complaints" paragraph says only that complaints may be registered
with the ward Complaint Officer and the city Water Control Office.

So for a water leak the platform can name the department and must say plainly
that it has no deadline to show. Borrowing the pothole's 48 hours would be
inventing one. This is the clearest candidate for the next RTI.

Where both apply, **the shorter binds** and the platform should say which is which — a citizen
reading "24 hours" deserves to know it is the contractor's own contract, not a court order.
