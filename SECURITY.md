# Security policy

TraceSarkar holds photographs taken by identifiable people at identifiable places and times. A
vulnerability here can expose a citizen who reported a problem to a powerful party. We take that
seriously and we would rather hear from you than not.

---

## Reporting a vulnerability

**Do not open a public issue.**

| Channel | Detail |
|---|---|
| Email | `security@tracesarkar.org` *(to be provisioned before launch)* |
| GitHub | Private security advisory on this repository |

Please include:

- What you found and where
- Steps to reproduce
- What an attacker could do with it
- Whether you accessed any real user data (and if so, please stop and tell us immediately)

---

## What to expect

| Stage | Target |
|---|---|
| Acknowledgement | 3 working days |
| Initial assessment and severity | 7 working days |
| Fix for critical issues | 30 days |
| Fix for other issues | 90 days |
| Public disclosure | Coordinated, after the fix, with credit if you want it |

---

## Safe harbour

We will not pursue legal action against researchers who:

- Act in good faith and report promptly
- Do not access, modify, or exfiltrate real user data beyond the minimum needed to demonstrate the issue
- Do not degrade service for other users
- Do not use the finding for any purpose other than reporting it
- Give us reasonable time to fix before public disclosure

If you are unsure whether something is in scope, ask first — `security@tracesarkar.org`.

---

## In scope

- The TraceSarkar API, web application, and WhatsApp bot
- The public data API
- Authentication and session handling
- Access control, including the moderation and admin surfaces
- Data exposure: PII, precise coordinates, un-redacted media, archival originals
- The redaction pipeline (a bypass is a high-severity finding)
- The coarsening layer (a way to recover precise coordinates from public data is a high-severity finding)
- Injection of any kind, including into the Watchdog query DSL
- SSRF via the ingestion framework
- Rate-limit and abuse-control bypasses
- Anything that lets one user see another's identity or report history

## Out of scope

- Denial of service by volume
- Findings that require a compromised device or physical access
- Social engineering of maintainers or users
- Missing security headers with no demonstrated impact
- Reports from automated scanners without a demonstrated exploit
- Vulnerabilities in third-party government portals we integrate with (please report those to them;
  tell us too so we can mitigate)

---

## Findings we consider critical

These map directly to real-world harm for a reporter:

1. **Reporter de-anonymisation** — any path from public data to a specific person's identity, phone
   number, or home location
2. **Redaction bypass** — retrieving an un-redacted face or number plate from a public surface
3. **Coarsening bypass** — recovering precise coordinates from public or API data
4. **Archival media access** without a logged, justified break-glass grant
5. **Moderation or admin surface access** by an unauthorised party
6. **Platform statement forgery** — writing a contract attribution or scorecard entry without going
   through the publication gates

---

## Our commitments

- Every report gets a human response
- Every confirmed vulnerability gets a fix and, where users were affected, a notification
- Every security incident is described in the quarterly transparency report
- We publish an SBOM per release and run dependency vulnerability scanning in CI
- We do not hold data we do not need — see
  [security and privacy](docs/03-architecture/11-security-and-privacy.md). The minimisation *is* a
  security control

---

## Supported versions

Pre-1.0: only the current deployed version is supported. Security fixes are applied to `main` and
deployed.
