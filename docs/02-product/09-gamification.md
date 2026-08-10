# Gamification — carefully

**Warning:** gamification is where civic platforms go wrong. Reward volume and you get volume — noise,
duplicates, and a leaderboard of people who photograph the same street daily. Every predecessor's
"1.8 crore complaints" statistic is a monument to that mistake.

---

## 1. What we reward

| Rewarded | Not rewarded |
|---|---|
| Reports that get **verified** by independent corroboration | Raw report count |
| **Re-checks completed** — going back and confirming or reopening | Time in app |
| **Verification missions** accepted and done accurately | Comment volume |
| Reports that lead to a **confirmed fix** | Upvotes given or received |
| Corrections that improve platform data (a wrong ward flagged, a bad geocode reported) | Streaks and daily logins |
| Onboarding a neighbour who becomes an active reporter | Referral count alone |

**The single rule:** reward *accuracy* and *closure*, never *volume*.

---

## 2. Mechanics

### Civic points

Points accrue for the rewarded actions above. They are:

- **Non-transferable**, non-purchasable, with no monetary value
- **Decaying** — an account inactive for six months decays slowly, so the leaderboard reflects
  current contribution
- **Revocable** — points from a report later found fabricated are removed along with the trust hit

### Levels

Five levels, each unlocking capability rather than status:

| Level | Unlocks |
|---|---|
| 1 · Reporter | Report, share |
| 2 · Verified reporter | Reports self-verify for low-risk categories; fewer rate limits |
| 3 · Verifier | Receives verification missions |
| 4 · Ward watcher | Can create campaigns; ward digest access |
| 5 · Reviewer | Eligible to help triage the moderation queue (with training and a two-person rule) |

Progression is tied to the trust score, which is computed from accuracy — so the levels are a
*visible* expression of an *invisible* score. The numeric trust score itself is never shown; a
visible score becomes a target to farm.

### Badges

Awarded for meaningful, hard things — not for participation:

| Badge | Criterion |
|---|---|
| **First confirmed fix** | A report you filed reached `citizen_confirmed` |
| **Closer** | 10 re-checks completed |
| **Ward memory** | Reported the same asset's repeat failure, establishing a pattern |
| **Corrector** | Flagged a platform data error that was upheld |
| **Monsoon watch** | Active verification during a declared monsoon event |
| **Organiser** | Ran a campaign that reached 25 corroborated issues |

No "10 reports" badge. No "100 reports" badge. Ever.

---

## 3. Leaderboards

**Ward-level only, and opt-in.** No city-wide leaderboard.

Reasons:
- A city-wide leaderboard creates an incentive to report anywhere rather than to fix somewhere.
- Ward-level competition maps onto how civic organising actually works.
- Opt-in avoids exposing a user's activity pattern (a privacy risk — see
  [trust and anti-abuse §9](08-trust-and-antiabuse.md#9-data-minimisation-as-an-anti-abuse-control)).

Ranked by **confirmed fixes contributed**, not by reports filed. A user with 3 confirmed fixes ranks
above one with 200 unverified reports, and that ordering is the whole message.

---

## 4. Community recognition, not just points

Higher-value than any badge:

| Mechanism | Detail |
|---|---|
| **Ward digest attribution** | The weekly ward digest names (with consent) the people whose re-checks closed issues |
| **Campaign credit** | Campaigns show their contributors |
| **The fix story** | When an issue reaches `citizen_confirmed`, everyone who contributed gets a notification showing the before/after and the elapsed time. This is the actual reward |

That last one matters most. The dopamine hit that sustains civic participation is *seeing the thing
get fixed*, not seeing a number increase.

---

## 5. Anti-gaming

| Attack | Control |
|---|---|
| Farming points with trivial reports | Points accrue on verification, not submission |
| Self-corroboration from a second account | Independence checks; clustered accounts do not corroborate each other |
| Mass false "it's fixed" confirmations | Confirmation requires a fresh geotagged photo; a later reopen reverses the points and hits trust |
| Reporting a real issue repeatedly | Dedup merges it; no additional points |
| Racing to be first on a viral issue | Corroborators are rewarded equally with the first reporter |
| Buying an account | Points are non-transferable; trust is not portable across accounts |

---

## 6. Interaction with trust

Points and trust are **related but distinct**:

| | Trust score | Points |
|---|---|---|
| Visible? | No (coarse badge at most) | Yes |
| Purpose | Gates verification thresholds and rate limits | Motivation and recognition |
| Decays? | Slowly, on inactivity | Slowly, on inactivity |
| Recoverable after a hit? | Yes, through accurate reporting | Yes |

Trust is a safety mechanism; points are a motivation mechanism. Conflating them would make the safety
mechanism gameable.

---

## 7. What we deliberately omit

| Omitted | Why |
|---|---|
| Streaks | Manufactures anxiety and daily-login behaviour that has nothing to do with civic value |
| Push notifications to maintain engagement | A civic tool that nags for engagement gets muted, and then cannot deliver an SLA breach alert |
| Public per-user report counts | Rewards volume by display, even if not by points |
| Monetary rewards or vouchers | Attracts exactly the wrong reporters and creates a fraud incentive |
| NFTs / tokens | No |
| Comparative shaming between wards' *citizens* | The comparison that matters is between *authorities*, not between residents |

---

## 8. Measurement

| Metric | Watch for |
|---|---|
| Reports per active user | A sharp rise without a rise in verification = farming |
| Verification rate by level | Should rise with level; if not, levels are mis-tuned |
| Re-check response rate | The mechanic we most want to work |
| Badge distribution | A badge nobody earns is mis-specified; a badge everyone earns is meaningless |
| Leaderboard opt-in rate | Low opt-in suggests the framing is off |

---

## 9. Open questions

- Does any of this move the numbers, or does the fix-notification alone carry the motivation?
  *Test the fix notification first, in isolation, before building points at all.*
- Should organisers (Sunita persona) have a separate progression? *Probably — their contribution is
  distribution, not reporting, and points do not capture it.*
- Is level 5 (moderation queue access) safe? *Only with training, a two-person rule, and a
  conflict-of-interest recusal policy. Consider deferring past v1.*
