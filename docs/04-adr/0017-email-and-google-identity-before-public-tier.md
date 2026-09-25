# ADR 0017 — Email/password and Google sign-in before the public tier

**Status:** Accepted · **Date:** 2026-09-25 · **Amends:** [ADR 0011](0011-phone-only-identity.md)

## Context

[ADR 0011](0011-phone-only-identity.md) makes phone-number verification the *only* identity
requirement, and rejects email ("free to mint at scale; no sybil resistance at all") and social
login ("a third-party data-sharing surface … for negligible sybil resistance"). That reasoning is
about **published** citizen reports: the platform names contractors, so the reports behind those
claims must be hard to manufacture.

The v0.1 client needs an account *now*, and at Phase 0 nothing it collects is published. Under
[D044](../00-overview/05-decision-log.md) every surface starts at the `personal` tier — the only
account is the operator's own. There is no sybil threat against a database of one, and no claim
about a named party rests on these captures yet.

Meanwhile phone OTP is not free: it needs a DLT-registered sender ID and a paid SMS route in India,
which is a procurement task, not an afternoon's work. Making that the gate on the first working
client would stall the field kit and the evaluation set behind a vendor.

## Decision

**Email/password and Google sign-in are accepted for the `personal` and `flagged` tiers. Phone
verification remains the gate for the `public` tier, unchanged.**

- An account may hold any of: a password credential, a Google subject, a verified phone. The
  `auth_providers` array on the account records which, so a surface can ask "is this account
  phone-verified?" rather than assuming.
- **No report reaches a public surface from an account that is not phone-verified.** ADR 0011 is
  intact where it was actually load-bearing; this ADR only says it does not apply before publication
  exists.
- Passwords: bcrypt, length-only policy (minimum 10). No composition rules — they push people toward
  predictable substitutions without adding entropy.
- Google sign-in verifies the ID token against Google's JWKS, and rejects any token whose `alg` is
  not `RS256`, whose `aud` is not our client ID, or whose `iss` is not Google. It is optional: with
  no `GOOGLE_CLIENT_ID` the endpoint refuses cleanly and email/password still works.
- Sign-in answers identically for an unknown address and a wrong password. On a civic platform the
  list of who is registered is itself worth something.
- Email is stored alongside a normalised form for uniqueness. ADR 0011's rule that accounts are
  **pseudonymous** stands: the display name is user-chosen and the email is never a public surface.

## Alternatives

| Option | Why not |
|---|---|
| **Wait for phone OTP** | Blocks the field kit, the evaluation set, and every downstream stage on an SMS vendor. The thing being protected — published claims about named parties — does not exist yet |
| **No accounts at Phase 0; a static bearer token only** | What the field kit already does. It does not survive more than one person, and gives no per-account attribution for captures, which the evaluation set needs |
| **Email now, and never require a phone** | Abandons ADR 0011's actual argument. When reports start naming contractors, "anonymous internet photos" is exactly the rebuttal we would have earned |
| **Google sign-in only** | A hard dependency on one third party for access to a civic tool, and unusable for anyone avoiding a Google account — a reasonable stance for someone reporting against local interests |

## Consequences

- Sybil resistance at `personal`/`flagged` is effectively nil. Accepted: nothing published depends
  on it.
- A gate must exist before the first public surface ships, and it is now a real piece of work rather
  than an assumption: the publication path must check `auth_providers` for a verified phone and
  refuse otherwise. Phase 1 does not exit without it.
- We now hold email addresses, which ADR 0011 was designed to avoid. They are personal data under
  DPDP: never logged, never rendered on a public surface, and deletable with the account.
- Google sign-in adds a runtime dependency on Google's JWKS endpoint. The key set is cached, so a
  brief outage degrades new Google sign-ins only, not sessions or email sign-in.
- A password reset flow does not exist yet. Until it does, a forgotten password means a new account.
