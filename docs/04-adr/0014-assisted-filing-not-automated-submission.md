# ADR 0014 — Assisted filing, never automated submission

**Status:** Accepted · **Date:** 2026-08-25
**Extends:** [ADR 0007](0007-never-auto-file.md), which established *that* the platform never files.
This record establishes *how* filing works instead, and closes off a specific implementation route.

## Context

[ADR 0007](0007-never-auto-file.md) forbids the platform submitting a legal or administrative
instrument. It did not say what the filing adapters may technically do, and the obvious engineering
answer — drive the portal with a headless browser on the server — keeps resurfacing because it is
easy and it would work.

The [August 2026 data availability audit](../01-research/08-data-availability-audit.md) §3 settled
the surrounding facts. **No civic body in the Mumbai Metropolitan Region exposes a citizen-facing
complaint API.** CPGRAMS has an integration path, but it is government-portal-to-government-portal
only. So every filing route is a web form built for a human.

The audit also established what those forms require:

| Destination | Access requirement (surveyed Aug 2026) |
|---|---|
| RTI Online Maharashtra | Registered account **and** a fee payment through net banking, card or UPI |
| CPGRAMS | Registered login |
| Aaple Sarkar | Registered account; a token number is issued on submission |
| MyBMC / MARG | Mobile OTP |
| RailMadad, Swachhata | App or web session |

A separate proposal was to drive these forms with an **anti-detection browser** — a build such as
Camoufox, whose purpose is to defeat bot fingerprinting and present automated traffic as an ordinary
human session.

## Decision

**1. The platform never submits to a government system on a citizen's behalf, by any mechanism.**
This restates ADR 0007 and extends it explicitly to server-side browser automation, headless
drivers, and request replay.

**2. The platform never stores, proxies, or relays a citizen's credentials for a government
portal** — no passwords, no session cookies, no OTP interception, no "we'll log in for you". There
is no code path that holds a government identity belonging to someone else.

**3. The platform never uses anti-detection or fingerprint-spoofing tooling against any
government system**, for filing or for collection. Where a site deploys a bot defence, that defence
is treated as the operator's answer.

**4. Filing is assisted, and the assistance happens on the citizen's own device.** The permitted
pattern, in order of preference:

| Pattern | What it does | Where it runs |
|---|---|---|
| **Prefilled draft + deep link** | Generate the complaint or application text, the correct addressee, the category codes; open the portal in the citizen's browser | Their device, their session |
| **Share-target / clipboard handoff** | One tap to paste the whole body into the portal's field | Their device |
| **Reference capture** | They photograph or paste the acknowledgement; OCR extracts the number; the statutory clock starts | Our server, on their artefact |
| **Read-only status check** | Poll a status endpoint with their own reference number, conservatively rate-limited, subject to a terms review | Our server |
| **Negotiated API** | A real integration under a written agreement | Either |

The citizen's tap is the filing event, and it is logged as such: who, when, which instrument
version, which destination.

## Rationale

**The legal reasons come first, and they are not about the automation.**

1. **Responsibility follows the signature.** An RTI application, a grievance, a tribunal
   application — each is filed by a named person answerable for its contents. The platform cannot
   be that person, and should not appear to be.
2. **A wrongly filed instrument is unrecoverable.** A misaddressed RTI burns the fee and the 30-day
   clock, and the appeal ladder stops. Multiply that by automation and the damage is silent.
3. **Bulk automated filing invites dismissal of the whole platform as vexatious**, which is the
   failure mode that ends the project. Not a lawsuit — a paragraph in an affidavit describing us as
   a machine that generates complaints.

**Then the credential problem, which is larger than the automation problem.**

Because every portal requires an account, and several require a payment, server-side submission
means holding a citizen's government-portal credentials and intercepting their one-time passwords.
That is a credential vault for government identities, inside the DPDP Act's scope, at an
organisation with no reason to hold it. The breach consequence is not "complaints leak" — it is
"someone else's government identity is compromised." No filing convenience is worth that, and
[ADR 0011](0011-phone-only-identity.md) already commits us to holding as little identity as
possible.

**Finally, the anti-detection point specifically.**

A bot defence is a statement of intent by the site operator, whatever the terms of use say or omit.
Circumventing it converts automated access into circumvention, which is a materially different thing
to defend — potentially characterisable as access without authorisation — and it is fatal in exactly
the setting where this platform's output is supposed to matter. The first question opposing counsel
asks about a filed instrument is how it was produced. "A stealth browser impersonating a human
session" is not an answer that survives, and it would taint every other record we publish.

It also contradicts what we have already decided elsewhere:
[hard rule 8](../../CLAUDE.md) and the audit's acquisition decisions record Mahatenders as
**no automated collection** precisely because it deploys a captcha and a `Disallow: /`. Defeating
the same class of control on a filing path would make that decision incoherent.

## Consequences

- Filing adapters are **draft generators plus deep links**, not robots. This is a product
  constraint, and the screen spec already reflects it (S20, S23).
- **We are slower than a hypothetical auto-filer, and that is the accepted cost.** The product's
  claim is not "we file faster"; it is "we know what to file, when, citing what."
- The **reference-number capture screen (S21) becomes load-bearing.** Without the number there is no
  clock and no ladder, and we cannot obtain the number ourselves. The UI nags, gently and forever.
- Status polling needs a terms review per destination, recorded in the source register, before it
  ships. Read is cheaper to justify than write, not free.
- A negotiated API is the only path to true one-tap filing. BMC already operates a WhatsApp
  complaint channel, which means the institutional capability exists. That is a partnership
  conversation to have once there is a filing volume and a clean record to point at — see
  [GTM](../05-delivery/06-gtm-and-partnerships.md).
- Any pull request adding a headless browser driver, a stealth browser dependency, or credential
  storage for a third-party portal is rejected on sight, with a link to this record.

## Alternatives considered

**Server-side automation with the citizen's explicit consent per submission.** Rejected. Consent
fixes the authorisation question between us and the citizen; it does nothing about the authority's
position, the credential custody, or the vexatious-at-scale problem. And a consent checkbox in front
of an automated filer is precisely the dark pattern ADR 0007 exists to prevent.

**A browser extension that fills and submits in one click.** Partially adopted — filling is fine and
genuinely useful; the submit stays a human click. An extension that clicks submit is server-side
automation wearing a costume.

**Filing through a partner NGO's account.** Rejected. It moves the responsibility rather than
resolving it, and it puts a partner's standing at risk on our software's correctness.
