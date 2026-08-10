# Landscape: India

What already exists, what it does well, and precisely where each one stops.

---

## 1. Swachhata (MoHUA)

| | |
|---|---|
| Operator | Ministry of Housing and Urban Affairs; technology built and maintained by Janaagraha |
| Scale | ~1.8 crore users, ~2.39 crore complaints, ~2.24 crore reported resolved (>93% claimed resolution rate) |
| Coverage | Effectively all ULBs nationally |
| Routing | Ward-level, to the sanitary inspector |

**Does well:** distribution at national scale; a genuine, working ULB-side workflow; the ward-officer
routing model is proven.

**Stops at:**
- **Sanitation only.** No roads, drains, streetlights, structural safety, or environment.
- **Self-certified closure.** The same ULB that owns the failure marks it resolved. A 93% resolution
  rate that citizens do not recognise on the ground is the definition of the metric TraceSarkar
  refuses to optimise.
- **No procurement link.** The complaint never touches the contract that was supposed to prevent it.
- **No escalation.** When the ULB closes it wrongly, the citizen's recourse inside the app is nil.

---

## 2. IChangeMyCity (Janaagraha)

| | |
|---|---|
| Scale (as of 2017 reporting) | >1.3M app downloads; ~608k users; ~245k complaints; ~174k connected to agencies |
| Notable | Complaint dataset published on the AWS Registry of Open Data |

**Does well:** a real community layer (neighbourhood groups, campaigns); early and principled
commitment to open data; the organisational credibility that produced Swachhata.

**Stops at:** grievance-centric data model; no parastatal jurisdiction resolution; no legal
escalation; no procurement engine.

---

## 3. MyBMC MARG / MyBMC 24x7 (BMC)

| | |
|---|---|
| Coverage | Greater Mumbai only |
| Categories | 114 grievance types |
| Channels | App (iOS/Android), web portal, helpline 1916 |
| Languages | English + Marathi |

**Does well:** it is the official channel with an official ticket, which matters for the evidentiary
chain; status tracking from "Logged" to "Resolved"; broad category coverage.

**Stops at:** single corporation; unilateral closure; no contract attribution; no cross-corporation
memory; the Bombay High Court has repeatedly found the underlying single-window mechanism ineffective.

**TraceSarkar's relationship to it:** *complementary, not competitive.* We file into MARG, capture the
official ticket number, and then track the lifecycle independently — including what happens after
BMC says "resolved".

---

## 4. Aaple Sarkar (Government of Maharashtra)

| | |
|---|---|
| URL | `grievances.maharashtra.gov.in` |
| Scope | State departments |
| SLA | 21 working days (stated) |
| Mechanism | Token number + status tracking + feedback rating |

**Does well:** state-wide; a stated timeline; a feedback loop.

**Stops at:** department-oriented, not asset-oriented; citizens rarely know which department owns a
physical problem; no geospatial layer.

---

## 5. RailMadad (Ministry of Railways)

| | |
|---|---|
| Channels | Web, app, SMS, helpline 139 |
| Linkage | PNR / UTS number |
| Constraints | 1,000-character limit; no modification after submission |

**Does well:** integrated into the railway grievance workflow; genuinely fast for in-journey issues.

**Stops at:** no persistent public record of station-level infrastructure failure; complaints are not
aggregated into a station-condition view; no procurement link; the character limit and immutability
make it useless as an evidence store.

**TraceSarkar's role:** a "Transit Mode" overlay triggered by geofencing station premises, filing
into RailMadad while keeping an independent, aggregatable record.

---

## 6. BBMP FixMyStreet / PACE (Bengaluru) and AI pothole surveys

Bengaluru is the closest Indian analogue and is worth tracking:

- BBMP launched a FixMyStreet-branded app; reported ~25,000 complaints in the first 15 days
- Later apps (PACE) added proof-of-work photographs after complaint closure
- The civic body ran an **AI-based road survey**: camera-mounted vehicles across ~1,600 km of East
  Zone roads using a YOLO-family detector plus a vision-language model to measure pothole length and
  width, and to identify utility-dig damage; 15 AI-camera vehicles deployed for arterial and
  sub-arterial roads

**Lessons:**
1. **Proof-of-work photographs after closure are table stakes.** Design for them.
2. **Municipal bodies are already procuring AI road surveys.** This is validation, and it is also a
   partnership surface: a citizen-sourced defect stream is complementary to a fleet-sourced one.
3. **Volume alone changes nothing.** 25,000 complaints in 15 days did not fix Bengaluru's roads. The
   missing element is consequence — which is precisely TraceSarkar's thesis.

---

## 7. Open contracting in India

| Effort | What it demonstrates |
|---|---|
| **CivicDataLab + Open Contracting Partnership** — Open Contracting India | State procurement data can be mapped to OCDS and published usefully |
| **Assam Public Procurement Explorer** | An OCDS publication with analytics, used in flood-management spending analysis |
| **Himachal Pradesh Health Procurement Index** | Indicator design on top of OCDS |
| **OCP Data Registry** | Assam's Finance Department appears as a registered OCDS publisher |

**Consequence:** OCDS mapping of Indian procurement is a solved problem with working precedent.
TraceSarkar should adopt OCDS as the internal normalisation target rather than inventing a schema,
and should aim to *publish* its normalised MMR contracts corpus as OCDS. That publication is
independently valuable and is a strong reason for researchers to engage.

---

## 8. Adjacent civic-tech efforts worth knowing

| Project | Relevance |
|---|---|
| **Open City (`opencity.in`)** | Practical guidance on accessing Indian government GIS data; a model for the data-access documentation TraceSarkar must maintain |
| **DataMeet** | The community that produced the municipal boundary datasets we depend on |
| **Bhashini / ULCA** | Government language stack; the realistic path to genuine vernacular support |
| **India Open Data Association / data.gov.in** | Variable-quality dataset host; occasionally useful |
| **Zero-FIR / e-Courts Phase III** | The broader digitisation of the justice pipeline; structured civic evidence becomes more valuable as courts digitise |

---

## 9. Where the gap is

Plot every Indian platform on two axes — **breadth of jurisdiction handled** and **depth of
consequence produced** — and the top-right quadrant is empty:

```
consequence
    ▲
    │                                     ┌─────────────────┐
    │                                     │  TraceSarkar    │
    │                                     │  (unoccupied)   │
    │                                     └─────────────────┘
    │
    │   RTI activists
    │   (deep, manual,
    │    unscalable)
    │
    │                    MyBMC ●
    │      Aaple Sarkar ●        ● Swachhata
    │            RailMadad ●   ● IChangeMyCity
    └──────────────────────────────────────────────▶ jurisdictional breadth
```

The individual RTI activist already does what TraceSarkar does — manually, for one case at a time,
after years of learning the process. The product is the industrialisation of that person's workflow.

## 10. Research tasks

- [ ] Survey the grievance app and channel for each of the 9 MMR corporations and 9 councils
- [ ] Obtain and analyse the IChangeMyCity open complaint dataset for MMR-relevant patterns
- [ ] Track BBMP's AI road-survey procurement (vendor, method, output format) as a partnership model
- [ ] Contact CivicDataLab about OCDS mapping precedent for Maharashtra
- [ ] Determine whether any MMR corporation publishes closure proof-of-work photographs
