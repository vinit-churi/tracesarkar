# Security and privacy

TraceSarkar holds photographs of public places taken by identifiable people at identifiable times,
joined to their precise location. That is a sensitive dataset with a real threat model, governed by
the DPDP Act 2023 and the DPDP Rules 2025.

---

## 1. Data inventory and classification

| Data | Class | Public? | Retention |
|---|---|---|---|
| Phone number | **Sensitive PII** | Never | Hashed at rest; plaintext only during the OTP window and for consented notifications |
| Account handle | Pseudonymous | Optional, user-controlled | Until account deletion |
| Report precise coordinates | **Sensitive PII** (location) | **Never** | Retained internally; see §5 |
| Report coarsened coordinates | Low | Yes | Indefinite |
| Photograph — archival original | **Sensitive** (may contain faces, plates) | Never | See §5 |
| Photograph — public derivative | Medium (redacted) | Yes | Indefinite |
| EXIF | Sensitive | Never | Archival copy only |
| Device / IP hashes | Sensitive | Never | 180 days |
| Chat transcripts | Medium | Never | 30 days |
| Contract / contractor data | Public record | Yes | Indefinite |
| Official designations | Public record | Yes | Indefinite |
| Individual officials' names | Sensitive by policy | Only where the official record names them in that capacity | — |

---

## 2. DPDP compliance

The DPDP Rules 2025 were notified on 13 November 2025 (G.S.R. 846(E)) with phased compliance —
the Data Protection Board from Nov 2025, consent-manager registration from Nov 2026, and full
substantive obligations with Schedule 1 penalties (up to ₹250 crore) from 13 May 2027.

### Obligations and implementation

| Obligation | Implementation |
|---|---|
| **Itemised consent notice** | Shown at first capture, not buried in a ToS. Lists each data category, each purpose, retention period, and the withdrawal mechanism, in the user's language. Versioned; re-consent on material change. |
| **Purpose limitation** | Each purpose is a separate consent scope: `report_processing`, `public_display`, `notifications`, `research_publication`. A user can decline `public_display` and still report. |
| **Data minimisation** | We do not collect: name, email, address, date of birth, gender, or any government ID. |
| **Data-principal rights** | Access (export my data), correction, erasure, grievance — implemented as first-class API endpoints and UI, not an email address. |
| **Withdrawal of consent** | Withdrawing `public_display` removes the report from public surfaces; the issue survives as an anonymised aggregate. |
| **Retention limits** | Per the table in §5, enforced by scheduled jobs, not by intention. |
| **Pre-erasure notification** | Inactive accounts are notified before scheduled erasure. |
| **Breach notification** | Documented procedure with defined timelines to the Board and to affected principals. |
| **Grievance officer** | Named, published, with response SLAs — also required under the IT Rules 2021. |

### The photograph problem

A photograph of a public street is not inherently personal data. It becomes personal data when it
contains an identifiable person or a number plate.

**Rule: redact before the public derivative exists.**

```
capture ──▶ archival original (private, encrypted, access-controlled)
              │
              └─▶ detect faces + plates ──▶ blur ──▶ public derivative ──▶ CDN
```

- Detection runs on-device where possible, server-side (on our own infrastructure) as a fallback —
  never via a third-party API before redaction.
- If detection fails or is uncertain, the image is held for review rather than published.
- The archival original is retained because it may be needed as evidence in a filing, under strict
  access control and audit logging.

---

## 3. Location privacy

The most under-appreciated risk. A public feed of precise, timestamped locations from one account is
a home-address inference dataset.

| Control | Detail |
|---|---|
| **Coarsening** | Public coordinates are snapped to a ~100 m grid with deterministic jitter, so the same issue always shows the same coarsened point |
| **No reporter linkage by default** | Public issues do not link to a reporter's other reports |
| **Aggregation floors** | Ward and heatmap statistics suppress cells below a minimum count |
| **Home-proximity heuristic** | Reports clustered tightly around one point over time, from one account, are flagged internally and never surfaced as a pattern publicly |
| **API guarantees** | The public API cannot return precise coordinates. Enforced at the serialisation layer, with a test |

---

## 4. Access control

| Role | Can |
|---|---|
| `citizen` | Own reports, public data, own escalations |
| `moderator` | Moderation queue, disputes, public data; **cannot** see phone numbers or precise coordinates of others |
| `investigator` | Precise coordinates and archival originals — **only** with a logged justification, time-boxed |
| `admin` | Configuration, source register, user administration |
| `api` | Scoped public read |

Principles: least privilege; no shared accounts; MFA on `moderator` and above; every privileged
action written to `audit_log` with actor, target, reason, and time. Access to archival media is
break-glass, not routine.

---

## 5. Retention schedule

| Data | Retention | Trigger |
|---|---|---|
| Archival original media | 7 years for issues that reached an escalation; 3 years otherwise | Legal evidentiary value |
| Public derivative media | Indefinite while the issue is public | — |
| Precise coordinates | Same as archival media | — |
| Device / IP hashes | 180 days | Abuse investigation window |
| Chat transcripts | 30 days | — |
| Phone plaintext | OTP window only (10 min) | — |
| Phone hash | Until account deletion + 30 days | Ban evasion window |
| Scraped procurement documents | Indefinite | Public records, evidentiary |
| Scraped news pages | 24 months | — |
| Audit log | 7 years | — |
| Deleted account data | Purged within 30 days; anonymised aggregates retained | — |

**Legal hold:** data relating to an issue that is the subject of a pending filing or dispute is
exempt from scheduled deletion until the matter closes. Implemented as a flag checked by the
retention job.

---

## 6. Encryption

| Layer | Control |
|---|---|
| In transit | TLS 1.2+ everywhere, including internal service-to-service |
| At rest — database | Full-disk encryption; `phone_hash` is HMAC with a pepper stored in the secret store, not the database |
| At rest — object storage | Server-side encryption; separate keys for public and archival buckets |
| Backups | Encrypted, with keys stored separately from the backup |
| Secrets | Managed secret store; never in the repo, environment files in git, or image layers |

---

## 7. Application security

| Control | Detail |
|---|---|
| Input validation | Schema validation on every endpoint; strict content-type checking |
| SQL | Parameterised queries only. The Watchdog query DSL compiles to parameterised SQL against a field whitelist — **no** dynamic SQL string building |
| File uploads | Content-type sniffing, magic-byte validation, size limits, image re-encoding (which also strips malicious payloads) |
| SSRF | Ingesters use an allowlist of hosts; no user-supplied URL is ever fetched server-side |
| XSS | Output encoding; CSP with no inline scripts |
| CSRF | SameSite cookies + token for any cookie-authenticated surface |
| Dependency management | Automated updates; `govulncheck` in CI; SBOM published per release |
| Rate limiting | Per account, per device, per IP, per endpoint |
| Secrets scanning | Pre-commit and CI |

---

## 8. Threat-specific responses

Beyond the abuse threats in [trust and anti-abuse](../02-product/08-trust-and-antiabuse.md):

| Threat | Response |
|---|---|
| **Legal pressure to remove true information** | Documented takedown workflow; court/competent-authority orders honoured within the IT Rules 2021 window; every action published in the transparency report |
| **Government data demand** | Documented process; minimum necessary disclosure; user notification where legally permitted; disclosed in aggregate in the transparency report |
| **Compelled disclosure of a reporter's identity** | We hold a hashed phone number and no name. The minimisation *is* the control. |
| **Infrastructure seizure or takedown** | Off-site encrypted backups in a separate jurisdiction-appropriate region; documented rebuild procedure |
| **Insider misuse** | Least privilege, audit log, break-glass access to sensitive data, periodic access review |
| **Targeted harassment of a reporter** | Pseudonymous handles by default; no cross-issue linkage; a rapid path to make an account's public footprint private |
| **Supply-chain compromise** | Pinned dependencies, SBOM, signed images, reproducible builds where practical |

---

## 9. Publication safety (recap)

The controls that keep the platform's public statements defensible live in
[legal framework §8](../01-research/04-legal-framework.md#8-publication-risk-defamation-and-intermediary-liability)
and [tender engine §6](06-tender-engine.md#6-publication-threshold--the-legal-gate). Summary of the
technical enforcement points:

1. A contractor name is rendered **only** when `issue_contract_matches.published = true`.
2. `published` is set only above the confidence threshold **and** with an acceptable geocode
   derivation.
3. Every published fact about a named party carries a `sources[]` array; the serialiser refuses to
   emit the fact if the array is empty.
4. Platform-generated statements and citizen-generated content are visually and structurally
   distinct in every surface.
5. Every scorecard displays its formula version and its inputs.

These are code-level invariants with tests, not editorial guidelines.

---

## 10. Incident response

| Phase | Actions |
|---|---|
| **Detect** | Alerting, user reports, security contact, anomaly detection |
| **Triage** | Severity classification within 1 hour; incident commander assigned |
| **Contain** | Isolate, revoke credentials, disable the affected surface |
| **Assess** | What data, how many principals, what exposure window |
| **Notify** | Data Protection Board and affected principals per DPDP timelines; status page immediately |
| **Remediate** | Fix, verify, restore |
| **Review** | Blameless post-mortem published in the transparency report |

Security contact and disclosure policy in `SECURITY.md`. Coordinated disclosure with a 90-day
window; no legal action against good-faith researchers.

---

## 11. Compliance checklist (pre-launch, blocking)

- [ ] DPDP consent notice drafted, reviewed, translated (en/mr/hi)
- [ ] Consent scopes implemented as separable
- [ ] Data-principal rights endpoints implemented and tested
- [ ] Retention jobs implemented with legal-hold support
- [ ] Redaction pipeline verified against an adversarial test set
- [ ] Coarsening enforced at the serialisation layer with a test
- [ ] Grievance officer named and published
- [ ] Takedown and right-of-reply workflows operational
- [ ] Correction log public
- [ ] `SECURITY.md` published with a working contact
- [ ] Backup restore tested end to end
- [ ] Access review completed; MFA enforced
- [ ] Pre-launch penetration test or, at minimum, a documented self-assessment against OWASP ASVS L2
