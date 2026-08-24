# ADR 0013 — What the platform may publish about elected representatives and candidates

**Status:** Accepted · **Date:** 2026-08-24

## Context

Two things changed at once.

First, **Mumbai has elected representatives again.** The administrator period ran from 7 March 2022
to January 2026; BMC polled on 15 January 2026 for 227 seats and a mayor took office in February.
Every "who is responsible for this ward" surface now has a corporator layer that the design docs do
not model.

Second, a question was raised directly: candidate disclosure data is public, so should the platform
publish representative track records, manifesto promises, and some measure of credibility?

The data is genuinely available. Form 26, prescribed under Rule 4A of the Conduct of Election Rules
1961 and section 33A of the Representation of the People Act 1951, requires every candidate to
declare pending criminal cases and convictions, assets and liabilities of the candidate, spouse and
dependants, education, PAN and income-tax particulars, and government dues. Returning Officers
upload it; the Election Commission publishes it at `affidavit.eci.gov.in`; ADR has republished
parsed versions at `myneta.info` for two decades.

The Supreme Court has repeatedly held that publication is the *point*. *Union of India v. ADR*
(2002) 5 SCC 294 and *PUCL v. Union of India* (2003) 4 SCC 399 grounded a voter's right to know in
Article 19(1)(a). *Public Interest Foundation* (2018) and *Rambabu Singh Thakur v. Sunil Arora*
(13 February 2020) went further and ordered **parties themselves** to publish their candidates'
criminal antecedents in newspapers, on television, on their websites and on their social media, with
reasons for selecting a candidate with a criminal record.

So the question is not whether this data can be republished. It is **which layer on top of it the
platform is willing to defend.**

Existing decisions pull the other way. The brand doc says the platform does not model politicians.
The feature catalog cuts political-party affiliation tagging outright and places a councillor/MLA
accountability page at v1+ with a risk score of 5. Those decisions were made for a reason:
party-neutral perception is a functional requirement in Maharashtra, not an aesthetic one.

## Decision

The platform publishes **records and joins about the office**, never **judgments about the person**.

### Permitted — build these

1. **Representative resolution.** Ward, prabhag, assembly and parliamentary constituency → the
   person holding that office, with the State Election Commission or ECI record as the source and a
   retrieval date. Never from the stale `COUNCILLOR` field in BMC's GIS layer.
2. **Verbatim Form 26 republication.** The declared fields, unaltered, each rendered next to a link
   to the source affidavit and a retrieval date. No cleaning, no reinterpretation, no inference.
   Pending cases are labelled *pending, charges framed on <date>* and never as "criminal".
3. **On-record legislative activity.** Assembly questions and answers about a ward's civic
   conditions, from the Vidhan Sabha's own published lists. This is the government answering itself
   on the record.
4. **Fund and works disclosure.** MPLADS and equivalent constituency-fund allocation and utilisation,
   as published.
5. **A commitment register** — see the constraints below.

### Forbidden — do not build these

| Not building | Why |
|---|---|
| **A credibility, integrity or performance score for any person** | It is the platform's own inference about a named individual. It is outside the DPDP Act's publicly-available-data exemption, it is a "decision specific to a Data Principal" so the research carve-out does not apply, and it is exactly the conclusion the platform has committed never to draw |
| **"Promise broken" as a status** | A negative finding stated in the platform's voice. It is not covered by the right-to-know cases, which concern statutorily mandated disclosures, not manifesto fulfilment |
| **Party affiliation as a filter, colour or grouping anywhere in the product** | Guarantees capture and destroys the neutrality the whole product depends on. Affiliation may appear only as one sourced field on a representative's record |
| **Attributing a specific civic failure to a named politician** | The platform states that a contract covers a location and that an office holds a duty. It does not say a person caused a pothole |
| **Asset-growth narrative** | Charting declared assets between two affidavits is permissible as a sourced delta with links. Any prose around it that implies a cause is not |

### The commitment register — the only promise-tracking we will build

A promise becomes a record only if all of the following hold:

- The commitment is **quoted verbatim** from a sourced, dated, archived artefact: a party manifesto,
  a written constituency manifesto, or an on-record public statement with a citation.
- It is **specific and checkable** — a named work, place, or measurable quantity. Aspirations are not
  recorded at all.
- Its status is drawn from a **closed vocabulary**, and no other words are used:
  - `Not yet due` — with the stated date
  - `Evidence found` — with the sourced artefact that evidences it
  - `No verified evidence as of <date>` — a statement about the platform's own records, not about
    the person
- Every status change is **logged with its evidence** and is reversible through the public correction
  log.
- The subject has a **standing right of reply** on the record, published alongside it.

`No verified evidence as of <date>` is deliberately clumsy. It is a statement about what the platform
has been able to verify, which is true and defensible, rather than a finding about a person, which
is neither.

### Election-period freeze

From the start of the ECI-notified campaign period to the close of polling, per constituency and per
phase:

- Editorial-judgment surfaces are **frozen**: no commitment-status changes, no new
  representative-linked analysis, no push notifications or share cards naming a candidate in a
  constituency that is voting.
- Verbatim record display continues, unchanged, with no new framing.
- In the **48-hour silence window** before polling closes, nothing naming a candidate in that
  constituency is published, shared or notified at all.
- An immutable audit log records exactly what was displayed, when, and from which source, for every
  candidate-related surface during a campaign period.

This is driven by section 126 of the Representation of the People Act 1951, which bars displaying
"election matter" — anything "intended or calculated to influence or affect the result of an
election" — in the 48 hours before polling closes. The Election Commission's stated position, cited
in litigation in both the Gujarat and Bombay High Courts, is that websites and social media are
electronic media for this purpose. Whether a non-paid civic surface is a "political advertisement"
requiring MCMC pre-certification is genuinely unresolved, and the freeze is what makes that question
survivable rather than urgent.

## Rationale

1. **The data layer is well precedented; the judgment layer is not.** ADR has republished affidavit
   data for twenty years. No comparable Indian precedent exists for a computed score attached to a
   named politician.
2. **Truth alone is not the defence people assume.** Under the Bharatiya Nyaya Sanhita's defamation
   provision, truth exculpates only when published **for the public good** — a separate, fact-specific
   question a court decides. Sourced records satisfy it comfortably. A derived score is a much harder
   argument.
3. **India has no actual-malice standard and no anti-SLAPP statute.** Public figures are not held to
   a higher bar, a criminal complaint can be quashed only through the inherent-powers route, and a
   civil suit can run for years irrespective of merit. The cost of being right is the real risk here,
   not the risk of being wrong.
4. **Safe harbour does not cover our own content.** Section 79 protects an intermediary hosting third
   party material. A compiled scorecard is the platform's own speech, and carries ordinary publisher
   liability.
5. **The DPDP Act exempts exactly the verbatim layer and nothing above it.** Section 3(c)(ii) excludes
   personal data made public by a person under a legal obligation — which is precisely what a
   Returning Officer does with Form 26. Derived and inferred data about the same person is inside
   the Act, and section 8(3) attaches an accuracy duty where the data is used in a decision affecting
   the person.
6. **Neutrality is load-bearing.** The moment the product ranks politicians, every fact it publishes
   is read as partisan, including the contract data that is the reason it exists.

## Consequences

- The data model gains a representative layer: office, term, person, source, retrieved-at.
- Feature E8 (councillor/MLA accountability page) is redefined as a **record page**, not a scorecard,
  and its risk score drops accordingly. It stays at v1+.
- The commitment register is a new feature, gated behind the same publication threshold and
  right-of-reply machinery as contractor records.
- The pre-commit checklist gains an election-period addendum.
- A Grievance Officer must be appointed before any representative-linked surface ships.
- Four questions go to counsel before the first live election cycle: whether the platform is a
  "publisher of news and current affairs content" under the IT Rules 2021; whether its content can
  be characterised as a political advertisement requiring pre-certification; the notified status and
  compliance timeline of the DPDP Rules; and whether to seek an advance clarification from the
  Election Commission rather than relying on analogy.

## Alternatives considered

**Publish a full scorecard like MyNeta.** Rejected. ADR's model is a purely declarative
republication of affidavits by an organisation whose entire mandate is electoral disclosure. Our
mandate is infrastructure accountability; adding candidate scoring buys us the political-capture
risk without the institutional standing.

**Publish nothing about representatives at all.** Rejected. A citizen asking "who is responsible for
this ward" deserves the answer, and the answer is now a person again. Withholding a sourced public
record is its own failure.

**Defer the whole question to v1+.** Partially adopted. The record page stays at v1+. The
representative resolution layer does not, because escalation routing needs it as soon as corporators
exist.
