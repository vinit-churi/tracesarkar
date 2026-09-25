# ADR 0011 — Phone-number identity only; no Aadhaar, no KYC

**Status:** Accepted · **Date:** 2026-08-10 · **Amended by:** [ADR 0017](0017-email-and-google-identity-before-public-tier.md) — email and Google
sign-in are permitted before the public tier; phone verification remains the gate on
publication

## Context

The platform publishes contractor names, ward-level failure statistics, and contractor scorecards.
Those claims are only defensible if the underlying citizen reports are credible, which requires
meaningful sybil resistance. At the same time, the reports are frequently adverse to powerful
interests, and the reporters are ordinary residents.

Identity design is therefore a direct trade between data credibility and reporter safety.

## Decision

**Phone-number verification (OTP) is the only identity requirement.** Specifically:

- Phone numbers are stored as `HMAC(number, pepper)`; the pepper lives in the secret store, not the
  database. Plaintext is retained only for the OTP window and for notification delivery under an
  explicit consent scope.
- Accounts are **pseudonymous**: a user-chosen handle, never a real name.
- No email, no address, no date of birth, no gender, no government identifier.
- Credibility beyond identity is carried by the **trust score** and by **corroboration**, not by
  stronger identity. See [trust and anti-abuse](../02-product/08-trust-and-antiabuse.md).

## Alternatives

| Option | Why not |
|---|---|
| **Anonymous** | Sybil-trivial. Would make contractor naming indefensible: "your evidence is anonymous internet photos." |
| **Email** | Free to mint at scale; no sybil resistance at all. |
| **Social login** | Adds a third-party data-sharing surface and a dependency, for negligible sybil resistance. |
| **Aadhaar / DigiLocker verification** | **Rejected.** (a) Excludes exactly the populations most affected by civic failure. (b) Makes the platform a high-value breach target holding government identifiers, with severe DPDP exposure. (c) Chills reporting against powerful local interests — the reporter must believe they cannot be trivially identified. (d) The marginal sybil resistance over phone verification is small, because a determined operative can obtain multiple verified identities either way. |
| **Optional Aadhaar for a "verified" badge** | Creates a two-tier system where the safest reporters are the least credible. Rejected on the same grounds. |

## Consequences

- Sybil resistance is weaker than a KYC system. Accepted, and compensated by trust scoring, device
  and behavioural clustering, corroboration requirements, and moderation.
- We hold almost nothing that could identify a reporter under compulsion — a hashed phone number and
  a pseudonymous handle. **The minimisation is the safety control.**
- SIM-farm attacks remain possible. Mitigations: rate limits per number and device, trust scoring
  from zero, and clustering detection. Monitored explicitly.
- **Revisit condition:** if sybil attacks demonstrably defeat the phone-based model and corroboration
  cannot compensate, reopen this decision as a new ADR with evidence. Do not weaken it quietly.
