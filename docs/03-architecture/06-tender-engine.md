# Tender engine — "follow the money"

**Answers:** *which public contract was supposed to prevent this defect, who holds it, for how much,
and is it still under warranty?*

This is the platform's differentiator and its hardest technical problem. If contract-to-geometry
matching cannot be made to work at usable accuracy, the product degrades to "a better complaint app"
— which is why this component gets its own milestone gate. See
[roadmap](../05-delivery/01-roadmap.md).

---

## 1. Pipeline

```
 ┌──────────────┐
 │ 1. INGEST    │  Mahatenders · BMC portal · CPPP · GeM
 │              │  → raw_documents (S3, sha256, write-once)
 └──────┬───────┘
        ▼
 ┌──────────────┐
 │ 2. PARSE     │  structured HTML fields → contracts (OCDS-shaped)
 └──────┬───────┘
        ▼
 ┌──────────────┐
 │ 3. EXTRACT   │  VLM/OCR over PDFs → DLP, work-site description,
 │              │  milestones, penalty clauses, contractor details
 └──────┬───────┘
        ▼
 ┌──────────────┐
 │ 4. GEOCODE   │  work-site text → geometry (contract_sites)
 └──────┬───────┘
        ▼
 ┌──────────────┐
 │ 5. MATCH     │  issue.location × contract_sites → issue_contract_matches
 └──────┬───────┘
        ▼
 ┌──────────────┐
 │ 6. SCORE     │  contractor_scorecard (published methodology)
 └──────────────┘
```

Stages 4 and 5 are the hard ones. Stage 3 is where the legally decisive data comes from.

---

## 2. Stage 1–2: ingest and parse

See [ingestion and scrapers](07-ingestion-and-scrapers.md) for the general framework. Contract-specific
notes:

| Source | Yields | Difficulty |
|---|---|---|
| Mahatenders (GePNIC) | Tender notices, corrigenda, award-of-contract | Medium — GePNIC has a consistent structure across states; captcha/session handling likely |
| BMC portal | BMC tenders and awards | Higher — SAP portal URLs are brittle |
| CPPP | Central bodies (Railways, port) | Medium |
| RTI | Work orders, completion certificates, measurement books, payment records | Slow but authoritative; the only reliable route to some fields |

**Normalise to OCDS.** Precedent exists in India (CivicDataLab's Assam and Himachal publications).
Using OCDS means the corpus is publishable as a public good and immediately usable by researchers.

Minimum viable OCDS mapping:

| OCDS | Source field |
|---|---|
| `ocid` | `{prefix}-{source}-{external_id}` |
| `tender.title`, `tender.description` | Tender notice |
| `tender.value.amount` | Estimated value |
| `awards[].suppliers[].name` | Award-of-contract |
| `awards[].value.amount` | Awarded value |
| `awards[].date` | Award date |
| `contracts[].period` | Work order → completion |
| `contracts[].implementation.transactions` | Payments (RTI-derived) |
| `planning.budget` | Budget head |

Fields OCDS does not model well — defect liability period, work-site geometry, penalty clauses — go
in an extension namespace.

---

## 3. Stage 3: extraction (the DLP problem)

**The defect liability period is almost never in a machine-readable field.** It lives in the general
conditions of contract, usually a scanned PDF annexure.

Known baselines to sanity-check extractions against (verify per contract):
- BMC concrete roads: guarantee roughly 5–10 years
- BMC asphalt roads: roughly 3 years
- ~20% of contract value typically withheld until the guarantee period ends
- Defects appearing within the DLP must be repaired by the contractor free of cost

### Extraction approach

A vision-language model with structured output over the contract PDF pages, with page-level
provenance:

```jsonc
{
  "defect_liability": {
    "found": true,
    "duration_months": 60,
    "starts_from": "completion_certificate",   // completion_certificate | work_order | handover
    "verbatim": "The Contractor shall be responsible for rectifying any defect …
                 for a period of five (5) years from the date of the completion certificate.",
    "page": 47,
    "confidence": 0.88
  },
  "work_site": {
    "description": "Improvement of Link Road from Kandivali Station Road junction to Charkop naka",
    "road_names": ["Link Road"],
    "landmarks": ["Kandivali Station Road junction", "Charkop naka"],
    "ward": "R/S",
    "chainage": {"from_km": 0.0, "to_km": 2.4},
    "page": 3,
    "confidence": 0.79
  },
  "penalty_clauses": [
    {"trigger": "delay", "rate": "0.5% per week", "cap": "10% of contract value", "page": 51}
  ],
  "retention": {"percent": 20, "release_condition": "expiry of guarantee period", "page": 52}
}
```

**Rules:**
- `verbatim` is mandatory for every extracted legal term. A DLP value without the sentence it came
  from is unciteable and therefore unusable.
- Page numbers are mandatory so the claim can be checked against the archived document.
- Extractions below a confidence threshold go to a human review queue; they are never published.
- Extraction is versioned (`extraction.model_version`, `extraction.prompt_version`) so a prompt
  improvement can trigger a re-run and a diff.

Model configuration is in [AI pipeline](04-ai-pipeline.md).

---

## 4. Stage 4: geocoding work sites

The hardest step. Tender descriptions are written for humans who know the city:

> "Improvement of Link Road from Kandivali Station Road junction to Charkop naka, R/S ward"
> "Desilting of minor nallahs in M-West ward, Group III"
> "Resurfacing of internal roads, Sector 12, Nerul"

### Strategy, in order of preference

| # | Method | Yields | Confidence |
|---|---|---|---|
| 1 | **Explicit coordinates or chainage** in the document | Precise geometry | 0.95 |
| 2 | **Road-name match** against the road network (OSM + BMC road layer), clipped to the named ward | LineString along the named road | 0.75–0.90 |
| 3 | **From–to landmark resolution** — geocode both endpoints, take the road path between them | LineString | 0.70–0.85 |
| 4 | **Ward + work-type polygon** ("all nallahs in M-West, Group III") | The ward polygon, or the union of matching assets in it | 0.35–0.55 |
| 5 | **Ward only** | Ward polygon | 0.25 |

Anything at or below level 4 is **not published as an attribution** on a citizen-facing card. It is
retained for research and for the Watchdog TUI, clearly labelled with its derivation and confidence.

### Gazetteer

A curated MMR gazetteer is required: road names with aliases (Link Road / Swami Vivekanand Marg /
S.V. Road confusion is endemic), landmarks, nakas, junctions, sectors, and nallah names, each with
geometry and a ward. This is a build-once, maintain-forever asset. Bootstrap from OSM, correct by
hand, and grow from unresolved extractions.

---

## 5. Stage 5: matching an issue to a contract

```
match(issue) →
  candidates = contract_sites within (buffer(site.geom, 30 m + issue.accuracy_m))
                 where contract.category compatible with issue.category
                 and   contract.stage in ('completed','in_progress','work_order')
                 and   contract.work_order_on <= issue.first_reported_at

  for each candidate, score:
    + 0.40  spatial containment or distance < 15 m
    + 0.25  category compatibility (road contract ↔ road defect)
    + 0.15  site derivation confidence
    + 0.10  contract is the most recent works contract on this asset
    + 0.10  asset-level linkage (contract explicitly names this asset)
    − 0.20  another contract is a strictly better spatial fit
    ×       site.confidence

  publish if score >= 0.75 and site.derivation in (explicit_coords, road_name_match, from_to)
```

Multiple matches are legitimate and should be shown: a road may have a construction contract and a
separate maintenance contract, and both are relevant.

`in_dlp` is computed as:

```
in_dlp = contract.dlp_ends_on IS NOT NULL
     AND issue.first_reported_at::date <= contract.dlp_ends_on
     AND contract.completed_on IS NOT NULL
```

---

## 6. Publication threshold — the legal gate

| Confidence | Behaviour |
|---|---|
| ≥ 0.75 **and** good derivation | Contractor name published on the issue card, with source and retrieval date |
| 0.50 – 0.75 | Contract shown **without** contractor name: "a road contract covers this area — details are being verified" |
| < 0.50 | Nothing shown; the issue offers to generate an RTI seeking the contract details |

**This threshold is a legal control, not a UX preference.** Naming a firm on a 0.55-confidence
geocode is precisely the fact pattern a defamation claim needs. See
[legal framework](../01-research/04-legal-framework.md#8-publication-risk-defamation-and-intermediary-liability).

Every published attribution carries:
- The source system and external ID
- The retrieval timestamp and archive hash
- The derivation method and confidence
- A "how was this matched?" expansion showing the full `basis` array
- A dispute link

---

## 7. Contractor identity resolution

Contractor names in Indian tender data are inconsistent: `M/s ABC Infra Pvt. Ltd.`,
`ABC Infrastructure Private Limited`, `A.B.C. INFRA PVT LTD`.

```
normalise(name):
  uppercase → strip honorifics (M/S, MESSRS) → expand abbreviations
  (PVT→PRIVATE, LTD→LIMITED, INFRA→INFRASTRUCTURE, ENGG→ENGINEERING)
  → remove punctuation → collapse whitespace
```

Then: exact match on normalised name → trigram similarity ≥ 0.85 → CIN match (authoritative when
available) → human review for anything in between.

**CIN is the real identity key.** Where the tender data carries a CIN or the firm can be matched to
MCA records, use it. Everything else is heuristic and must be treated as such.

### Corporate lineage (v1+)

Detecting blacklist evasion — "same directors, new company" — is genuinely valuable and genuinely
risky. Rules:

- The **fact** published is the overlap: "Directors A and B of Firm X (blacklisted by BMC in 2023)
  are also directors of Firm Y (awarded contract Z in 2026)", sourced to MCA records.
- The **conclusion** ("this is a shell to evade blacklisting") is never asserted by the platform.
- Director-name matching without DIN is unreliable (common names) and must not be published; DIN
  matching is required for publication.

---

## 8. Contractor scorecard

A published, versioned formula — never a black box.

```
score = 100
      − 3 × defects_in_dlp
      − 1 × defects_outside_dlp
      − 15 × blacklistings_active
      − 5 × blacklistings_historic
      − 10 × (contracts_completed_with_zero_evidence / contracts_completed) × 10
      + 5 × (citizen_confirmed_fixes / defects_in_dlp)      -- rewards actually fixing things

grade: A ≥ 85 · B ≥ 70 · C ≥ 55 · D ≥ 40 · E < 40
```

Displayed with:
- Every input value, visible
- The formula version and its changelog
- A minimum-data threshold: no grade is shown below N contracts or M observation-months, because a
  grade computed from two data points is noise presented as judgment
- The dispute link

**Normalisation caution:** a contractor with 200 contracts will accumulate more raw defects than one
with 5. The formula above is deliberately simple for v1; the correct v2 version normalises by
contract-kilometres or contract value and is a research task, not a guess. Until then, the scorecard
displays absolute counts alongside the grade so a reader can see the denominator.

---

## 9. What this engine explicitly does not do

| Not doing | Why |
|---|---|
| Assert that a contractor did substandard work | That is a technical finding requiring inspection, not a spatial join |
| Assert corruption | A legal conclusion the platform is not competent to draw |
| Detect bid rigging automatically | NLP on tender text can flag *restrictive specifications* as a descriptive observation; it cannot establish collusion |
| Satellite verification of road works | Sentinel-2 at 10 m cannot resolve whether a 6 m road was resurfaced. Scope any imagery work to large sites, afforestation, and building footprints, or drop it. See [open questions](../01-research/07-open-questions.md) |
| Publish payment-vs-milestone gaps without the underlying documents | Requires RTI-obtained payment records; inference from partial data is not publishable |

---

## 10. Evaluation

| Metric | Target | Measurement |
|---|---|---|
| DLP extraction accuracy | ≥ 90% on a hand-labelled set of 200 contracts | Human verification against the archived PDF |
| Work-site geocoding precision @ published threshold | ≥ 95% | Hand-verify 200 published attributions |
| Work-site geocoding recall | ≥ 50% initially | Fraction of contracts that reach a publishable geometry |
| Contractor identity resolution | ≥ 98% precision | Manual audit of merges |
| Attribution disputes upheld | < 2% of published attributions | Dispute outcomes |

**Precision over recall, always.** Missing an attribution costs a feature. A wrong published
attribution costs the project.

---

## 11. Implementation order

1. Ingest Mahatenders + BMC for **one ward**, one work category (roads), one financial year
2. Parse structured fields → `contracts`
3. Extract DLP + work-site from the PDFs of that set; hand-label 200 for evaluation
4. Build the gazetteer for that ward
5. Geocode; measure recall and precision
6. Match against existing issues in that ward; hand-verify every match
7. Only then: expand geographically
8. Only then: build the scorecard

**Gate:** if step 5 cannot reach ≥50% recall with ≥95% precision on one ward, stop and re-plan the
roadmap around escalation instead of attribution. That decision point is explicit and scheduled.
