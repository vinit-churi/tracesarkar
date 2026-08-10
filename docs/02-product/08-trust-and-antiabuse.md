# Trust and anti-abuse

A platform that names contractors and officials, in a region with active political competition, will
be attacked. Not "might be" — will be. This document specifies the threat model and the design
response.

**Design stance:** trust is a v1 concern. Retrofitting it after the first coordinated attack means
rebuilding the data model under fire.

---

## 1. Threat model

| # | Adversary | Goal | Method |
|---|---|---|---|
| T1 | Political operative | Make an opponent's ward look bad | Bulk-report real-but-trivial issues; report old photos as new; concentrate reports before an election |
| T2 | Political operative | Make own ward look good | Mass-confirm fake resolutions; dispute genuine issues; suppress via mass-flagging |
| T3 | Contractor / agency | Remove or discredit adverse records | Bulk disputes; legal notices; astroturfed "resolved" confirmations; DMCA/defamation takedown abuse |
| T4 | Individual griefer | Amusement / harassment | Fake photos, unrelated images, offensive content, targeting a private individual's property |
| T5 | Competitor firm | Damage a rival | Fabricate defects attributed to a rival's contracts |
| T6 | Scraper | Harvest citizen data | Bulk API pulls; correlate report locations to infer home addresses |
| T7 | State actor | Suppress the platform | Takedown orders, intermediary-liability pressure, data demands |
| T8 | Insider | Data leak or manipulation | Privileged access misuse |

---

## 2. Identity: the deliberate choice

**Decision: phone-number verification, no Aadhaar, no KYC.**

| Option | Verdict |
|---|---|
| Anonymous | ✘ Sybil-trivial; makes contractor-naming indefensible |
| Phone (OTP) | ✔ **Chosen.** Real friction, near-universal, no sensitive identifier stored |
| Email | ✘ Free to mint at scale |
| Social login | ✘ Adds a dependency and a data-sharing surface for no sybil resistance |
| Aadhaar / DigiLocker | ✘ Rejected — see below |

**Why Aadhaar is rejected:** it excludes exactly the populations most affected by civic failure, it
makes the platform a high-value breach target under DPDP, it chills reporting against powerful
parties, and its marginal sybil resistance over phone verification is small. Reconsider only if
sybil attacks demonstrably defeat the phone-based model — and record that reconsideration as an ADR.

Phone numbers are stored **hashed with a per-deployment pepper**; the plaintext is retained only for
the OTP window and for notification delivery under an explicit consent scope.

---

## 3. Trust score

Every account carries a score in [0, 1], recomputed on every relevant event.

### Inputs

| Signal | Direction | Weight class |
|---|---|---|
| Reports later verified by independent corroboration | ↑ | high |
| Reports whose resolution was later citizen-confirmed | ↑ | high |
| Re-verification missions completed accurately | ↑ | high |
| Account age and steady activity | ↑ | low |
| Reports rejected as not-a-civic-issue | ↓ | medium |
| Reports found to be duplicates the user was shown and dismissed | ↓ | low |
| Reports found to use reused / manipulated / stale imagery | ↓↓ | severe |
| Reports in wards where the account has no other activity footprint | ↓ | low (weak signal alone) |
| Clustering with other accounts on device / IP / behavioural timing | ↓↓ | high |
| Moderator-upheld abuse reports against the account | ↓↓ | severe |

### Effects

| Trust band | Effect |
|---|---|
| `< 0.2` | Reports accepted but not publicly displayed until moderated; hard rate limits |
| `0.2 – 0.5` | Normal; requires corroboration for verification |
| `0.5 – 0.8` | Reports self-verify for low-risk categories |
| `> 0.8` | Reports self-verify for most categories; eligible for verification missions; eligible to review the moderation queue |

**Never expose the numeric score.** Publish a coarse badge at most. A visible score is a target to be
farmed and a status game that corrupts reporting behaviour.

---

## 4. Verification policy by category

The original concept's "three independent users within a timeframe" rule is right for reputational
claims and dangerously wrong for life-safety ones.

| Category | Public display | Routing |
|---|---|---|
| `hazard_to_life = true` (open manhole, live wire, imminent collapse) | Immediate, labelled `unverified — urgent` | **Routed immediately.** Corroboration is sought in parallel, never as a gate. A delayed open-manhole alert is worse than a false one. |
| Routine road / waste / street furniture | On verification | Routed on verification |
| Environment, illegal construction, anything naming a private party | After corroboration **and** moderator review | Routed after review |
| Anything naming an individual (not an entity) | Not displayed publicly by default | Handled as a moderation case |

---

## 5. Corroboration integrity

Corroboration only means something if the corroborators are genuinely independent. Reports are
treated as **non-independent** when they share:

- The same device fingerprint or install ID
- The same IP prefix within a short window (with care — carrier-grade NAT and shared Wi-Fi are common
  and produce false positives; this signal is weak alone and must be combined)
- Referral chains (account A invited B invited C)
- Near-identical capture timing and heading (a group photographing the same thing from one spot)
- Perceptually near-identical images (the same photo re-uploaded)

The independence check is a graph problem. It reduces trust weight; it does not auto-ban. Bans are a
human decision. See [moderation policy](../06-operations/02-moderation-policy.md).

---

## 6. Image integrity

| Check | Purpose |
|---|---|
| **Perceptual hash** against the corpus | Catch re-uploads of an existing photo |
| **EXIF analysis** | Missing EXIF on a claimed live capture; capture time far from submission time; editing-software markers |
| **Device time vs server time divergence** | Backdating |
| **GPS accuracy and jitter profile** | Spoofed-location apps produce implausibly perfect fixes |
| **Sun position / shadow vs claimed time** | Cheap sanity check for coarse fabrication (advisory only) |
| **VLM cross-check** | "Does the image content match the claimed category and the claimed weather/time?" |
| **Reverse-search of viral imagery** | Old news photos recirculated as fresh reports |

None of these is conclusive alone. They feed a composite `image_integrity_score` that gates public
display and adjusts trust, and every automated downgrade is auditable.

---

## 7. Resolution confirmation — the anti-fraud crux

The moment of greatest manipulation incentive is when an authority marks an issue `claimed_resolved`.

| Rule | Detail |
|---|---|
| Only a **citizen photograph** can move an issue to `citizen_confirmed` | An authority's own claim never confirms itself |
| The confirming photo must be **fresh** | Captured after the resolution claim, with live geotag inside the issue radius |
| The confirmer should not be the original reporter where possible | Verification missions route to nearby high-trust users |
| The **claimed vs confirmed gap** is published per authority and per ward | This ratio is the platform's most important public statistic |
| Re-reports at a confirmed location within the cool-off window auto-`reopen` | And are flagged on the contractor's scorecard |

---

## 8. Dispute and right of reply

Named parties must have a real channel. Bulk-dispute abuse (T3) is countered by process design, not
by refusing disputes.

| Property | Design |
|---|---|
| Single-instance | Disputes are filed per record, by a human, with a stated basis. No bulk endpoint. |
| Evidence-bearing | The disputant must state which specific fact is wrong and why. "This is defamatory" without specifics is triaged, not actioned. |
| Timeboxed | Published SLA for platform response. |
| Transparent | The existence of a dispute is shown on the record while it is under review. |
| Outcome-published | Upheld → the record is corrected and the correction logged. Rejected → the record stands and the response is displayed alongside. |
| Escalation | Court or competent-authority orders are honoured under the IT Rules 2021 (36-hour window for court/government orders); every such action is logged in a transparency report. |

---

## 9. Data minimisation as an anti-abuse control

| Data | Public | Internal |
|---|---|---|
| Reporter identity | Never (pseudonymous handle at most) | Hashed phone + account ID |
| Exact report coordinates | **Coarsened** for public display and API (snapped to a grid; jitter within the accuracy radius) | Exact |
| EXIF | Stripped from public derivatives | Retained in the archival copy |
| Faces / number plates | Blurred before public derivative generation | Blurred in storage too, except where the original is legally required as evidence and access-controlled |
| Reporter's report history | Not publicly linkable across issues by default | Available to the account owner |

**Why coarsening matters:** a public feed of precise, timestamped locations from a single account is
a home-address inference dataset. This is both a T6 risk and a DPDP obligation.

---

## 10. Rate limits and cost controls

| Surface | Limit (initial, tunable) |
|---|---|
| Reports per account | 20/hour, 100/day; trust-scaled |
| Reports per device | 30/hour |
| Reports per IP | 60/hour (loose — shared connections are normal) |
| Classification API calls | Global daily budget with graceful degradation to queue-and-classify-later |
| Public API | Per-key quota; anonymous access heavily limited |
| Dispute filings | 5/day per verified legal-entity contact |
| Share-kit generation | 30/hour per account |

Rate limiting also protects against the cost-attack variant of T4: bulk submissions purely to burn
the VLM budget.

---

## 11. Insider and infrastructure controls

- Role-based access with least privilege; no shared admin accounts
- Every moderation and admin action written to an append-only audit log with actor, time, and reason
- Access to raw (un-redacted) media requires an explicit, logged justification
- Secrets in a managed store, never in the repo; rotation schedule documented
- Backups encrypted; restore tested quarterly
- Published security contact and disclosure policy (`SECURITY.md`)

---

## 12. Transparency report

Published quarterly, covering:

- Takedown requests received, by source and outcome
- Government data demands received and responses
- Accounts actioned, by reason
- Disputes filed, upheld, and rejected
- Corrections issued
- Known incidents and their resolution

This is a cheap, powerful legitimacy asset — and it is the thing that makes the platform's own
accountability claims non-hypocritical.

---

## 13. Open questions

- What is the minimum viable independence signal set that does not misfire on shared-Wi-Fi
  households and carrier-grade NAT?
- How do we handle a genuine mass-reporting event (a real corridor-wide failure) without the
  astroturf heuristics suppressing it? *Proposal: a "burst mode" that switches from independence
  scoring to spatial-spread scoring when volume exceeds a threshold across a wide area.*
- What is the right cool-off window before an issue moves from `citizen_confirmed` to `closed`?
  *Proposal: 90 days for road defects; category-tuned thereafter.*
- Do we ever publish an individual official's name, or only designations? *Current position:
  designations by default; names only where the official record itself names them in that capacity.*
