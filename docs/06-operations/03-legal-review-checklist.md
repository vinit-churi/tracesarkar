# Legal review checklist

Run this before any feature that **publishes a name**, **generates an instrument**, or **collects new
data**. It is not legal advice; it is a set of engineering gates derived from
[legal framework](../01-research/04-legal-framework.md) and
[security and privacy](../03-architecture/11-security-and-privacy.md).

---

## A. Publishing anything about a named party

Applies to: contractor names, firm names, official designations, blacklisting records, scorecards.

- [ ] **Source exists.** The fact comes from a public record we have archived, with a URL, retrieval
      timestamp, and SHA-256.
- [ ] **Source is cited in the UI**, not just in the database. A user can click through to see where
      it came from.
- [ ] **Confidence threshold met.** For contract attribution: `published = true` requires confidence
      ≥ 0.75 **and** an acceptable geocode derivation. Below that, the name is not rendered.
- [ ] **Statement is factual, not conclusory.** "Contract X covers this location; its DLP is active"
      ✔. "This contractor did substandard work" ✘.
- [ ] **Platform statements are visually distinct** from citizen-generated content.
- [ ] **Derived metrics show their inputs.** A scorecard grade displays its formula version and the
      values that produced it.
- [ ] **Minimum-data threshold.** No grade is shown below the configured minimum contract count and
      observation window.
- [ ] **Dispute link present** on the record.
- [ ] **Right-of-reply process reachable** and staffed.
- [ ] **Individuals:** designations by default; a person's name only where the official record names
      them in that capacity.
- [ ] **No political affiliation** stated or implied anywhere.

**Fail any of these → the fact is not rendered.** Enforced in the serialiser, with a test.

---

## B. Generating a legal or administrative instrument

Applies to: RTI, first appeal, RTS appeal, NGT OA, Lokayukta complaint, consumer complaint, HC
compensation claim.

- [ ] **Fee and format verified against the current rules** at generation time, not hard-coded from
      an old constant. Maharashtra RTI fees changed under the 2026 rules; assume any figure may be
      stale.
- [ ] **Legal constants come from `legal_constants`**, with a citation and an effective-from date.
- [ ] **Limitation period computed and displayed**, with the cause-of-action date shown.
- [ ] **Correct forum and addressee** for the resolved authority (state vs central RTI portal; NGT
      Western Zone for MMR).
- [ ] **Applicant is the citizen**, never the platform.
- [ ] **Draft status is unambiguous** in the UI and on the rendered document.
- [ ] **Disclaimer present:** generated draft, not legal advice, review before filing.
- [ ] **No auto-submission path exists.** See [ADR 0007](../04-adr/0007-never-auto-file.md).
- [ ] **Review checklist shown** to the citizen before they file.
- [ ] **Structured content, deterministic rendering.** The model fills a schema; a template engine
      produces the document.
- [ ] **Every statutory citation traceable** to the constants table, not to model output.

---

## C. Collecting or processing new data

- [ ] **Purpose identified** and mapped to an existing or new consent scope.
- [ ] **Consent notice updated** if the purpose is new; re-consent triggered on material change.
- [ ] **Minimisation applied.** Is every field necessary? Can it be derived instead of collected?
- [ ] **Retention period set** and enforced by the retention job.
- [ ] **Legal-hold interaction considered.**
- [ ] **Classification assigned** in the data inventory.
- [ ] **Public exposure decided** — and if public, coarsening or redaction applied.
- [ ] **Data-principal rights covered** (access, correction, erasure include the new field).
- [ ] **Logging reviewed** — the new field must not appear in logs, metrics, or traces.

---

## D. Adding a data source

- [ ] **Terms of use read and recorded** in the source register, with reviewer and date.
- [ ] **`robots.txt` checked** and obeyed.
- [ ] **Automated access permitted?** If not, the ingester is not built; RTI becomes the channel, and
      that decision is recorded.
- [ ] **Rate limits set** at or below the framework default (1 req / 3 s / host).
- [ ] **Attribution requirements** identified (e.g. DataMeet's CC BY-SA 2.5 IN; OSM's ODbL) and
      implemented in the UI and exports.
- [ ] **Licence compatibility** with our AGPL/CC BY-SA posture confirmed.
- [ ] **Archive path configured**; artefacts stored write-once with hashes.
- [ ] **PII in the source?** If yes, is our processing of it lawful for our purpose?

---

## E. Adding a filing adapter

- [ ] **Target platform's terms permit** automated submission on a user's behalf.
- [ ] **User's explicit action** triggers each submission.
- [ ] **Credentials** — does the adapter need the user's account? If so, how are credentials handled?
      (Preference: do not hold user credentials for government portals at all; generate and hand off.)
- [ ] **Reference number captured** and verified; an issue is never marked filed without one.
- [ ] **Failure degrades** to draft + deep link, never to silent failure.
- [ ] **Rate limits** respect the target platform.
- [ ] **Audit log entry** for every submission.

---

## F. Publishing an aggregate or a dataset

- [ ] **Aggregation floor applied** — no cell or group below the minimum count.
- [ ] **No re-identification path.** Can an individual reporter be inferred from the aggregate,
      cross-referenced with anything public?
- [ ] **Provenance manifest attached** — sources, retrieval times, dataset version, caveats.
- [ ] **Caveats are generated, not boilerplate** — thresholds applied, records excluded, known gaps.
- [ ] **Licence stated** (CC BY-SA 4.0) with attribution requirements.
- [ ] **Citizen photographs are not included** in bulk dataset releases.

---

## G. Standing obligations (verify quarterly)

- [ ] Grievance officer named, contactable, and responding within published SLAs.
- [ ] Takedown workflow operational; court/competent-authority orders actioned within the IT Rules
      2021 window (36 hours for the specified grounds).
- [ ] Correction log public and current.
- [ ] Transparency report published (takedowns, data demands, accounts actioned, disputes,
      corrections, incidents).
- [ ] Retention jobs running; legal holds respected.
- [ ] Consent notice current with actual processing.
- [ ] `SECURITY.md` contact working.
- [ ] Access review completed.

---

## H. Escalation triggers — stop and get advice

Do not proceed on engineering judgement alone if any of these occur:

| Trigger | Action |
|---|---|
| First legal notice of any kind | Preserve everything; do not modify the record under pressure; obtain counsel |
| A government takedown or data demand | Follow the documented process; minimum necessary disclosure; log for the transparency report |
| A dispute alleging a factual error in a published attribution | Freeze that attribution pending review, not just that record |
| A journalist reports that a published number is wrong | Treat as a P1 incident; correct publicly and log |
| A pattern of disputes on one contractor | Review the matching methodology, not just the individual records |
| Any indication a reporter has been identified or harassed | Immediate review of coarsening, linkage, and account footprint controls |

---

## I. Blocking gates by milestone

| Milestone | Gate |
|---|---|
| v0.1 launch | Sections A (minimal — names withheld unless ≥0.95 confidence), C, D, G |
| v0.3 (RTI generation) | Section B in full; fee verification live |
| v0.5 (scorecard) | Section A in full **plus** documented external review of the scorecard methodology |
| v0.5 (public API) | Section F in full |
| v0.6 (NGT compiler) | Section B, with counsel review of the compiled brief format |
