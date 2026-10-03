# Scaling beyond one ward

**Written 3 October 2026**, prompted by two questions worth asking early: *why only potholes*, and
*how would this work in Virar?*

The second question generalises. This file is the answer to "how do we cover Maharashtra", and its
conclusion is that the question decomposes into five layers which scale completely differently —
and that most of the platform's value is **already statewide**, gated on one dataset rather than on
three hundred research projects.

It does not change [the roadmap](01-roadmap.md). v0.1 is still one ward, and
[§5 of the MVP](02-milestone-v0-mvp.md) still governs. This is what comes after the hypothesis is
tested, written down now so the one-ward work is not accidentally built in a way that cannot widen.

---

## 1. The shape of the problem

| Tier | Count | Governing Act |
|---|---|---|
| Municipal corporations | 29 | Maharashtra Municipal Corporations Act 1949 — **except BMC**, on the Mumbai Municipal Corporation Act 1888 |
| Municipal councils | 246 | Maharashtra Municipal Councils, Nagar Panchayats and Industrial Townships Act 1965 |
| Nagar panchayats | 42 | Same Act |
| Zilla parishads | 34 | Maharashtra Zilla Parishads and Panchayat Samitis Act 1961 |
| Gram panchayats | ~28,000 | Maharashtra Village Panchayats Act 1958 |

Plus the parastatals that own roads nobody's municipality owns: PWD, MSRDC, MMRDA, NHAI, MIDC.

**The important column is the third one.** Three Acts cover every urban body in the state. The work
of finding that refuse removal sits under MMC Act 1888 §365(a) — done first-hand for BMC on
3 October 2026 — is done **once per Act, not once per body**.

---

## 2. The five layers

### Layer 1 — the law. Already statewide. No per-body work at all.

The Bombay High Court's pothole order is not a Mumbai order. Its own text names
**Municipal Corporations, Municipal Councils, the District Collector** (outside municipal limits),
**PWD, MSRDC, MMRDA and NHAI**, and constitutes a compensation committee for each tier with the
District Legal Services Authority. See the
[archived paragraphs](../01-research/sources/2026-08-24-bombay-hc-pil-71-2013-paras-69-70.md) and
[the SLA note](../01-research/11-pothole-sla-sources.md).

So the 48-hour clock, the ₹6,00,000 death compensation and the DLSA route **already apply in Virar,
in Nagpur, and on a village road under a Collector**. Nothing has to be researched per body to make
them true.

What remains at this layer is bounded and one-off:

- Read the three municipal Acts for the statutory duty behind each of the six categories, with
  section numbers. Three readings, not 317.
- Obtain the **Maharashtra RTS Act 2015 notified-services list**. Statewide, and it carries a
  personal penalty of up to ₹5,000 on the *named officer* — sharper than anything else available.
  This is the highest-value unanswered question in the
  [legal framework](../01-research/04-legal-framework.md).

### Layer 2 — boundaries. One dataset, not 317.

This is the only thing standing between the platform and a useful answer anywhere in Maharashtra.
Today a capture in Virar resolves to *"outside every ward boundary we hold"* — measured, the nearest
loaded ward is R/N, 20.9 km away.

Candidate sources, in order of expected tractability:

1. **State Election Commission** — ward delimitation was published for every local body for the
   2025–26 local elections. Recent, authoritative, and complete by construction.
2. **MRSAC** (Maharashtra Remote Sensing Applications Centre) — runs the state GIS.
3. Per-body GIS, as BMC's `geo/getwardlayer` already provides ([D070](../00-overview/05-decision-log.md)).

One acquisition moves every city in the state from level 0 to level 1 on the ladder below.

### Layer 3 — departments and PIOs. The long tail, and it is mechanical.

Every public authority **must** publish Section 4(1)(b) handbooks. BMC's sit at predictable URL
patterns; the two added on 3 October 2026 were found by pattern-guessing a third. Where a handbook
is missing or silent, the RTI route exists and the first appeal is free.

This is the layer the platform's own users can carry. A citizen in Nagpur who files one RTI for one
department contributes that mapping permanently, for everyone, in that ward.

### Layer 4 — contracts. The hard one, and the differentiator.

BMC is **exceptional** in publishing a works API carrying contractor, work code, dates and geometry.
Nothing suggests the other 28 corporations do.

The statewide equivalent is **MahaTenders**, the state e-procurement portal: one source, every state
body, carrying contractor, award value and a free-text work description — and **no geometry**.
Turning *"Improvement of road from X chowk to Y temple, Ward 12"* into a line on a map is the actual
research problem, and it is already the documented stop-and-replan gate at v0.2.

**Until this layer exists somewhere, the platform in that place is a complaint app** — which is what
every predecessor already is, and what this project exists not to be.

### Layer 5 — rural. Honestly out of scope for a long time.

34 zilla parishads and roughly 28,000 gram panchayats, under different Acts, with far weaker
publication. The HC order covers them through the District Collector, so the *legal* layer is
present; everything else is not. Recording this as out of scope is more useful than pretending a
plan exists.

---

## 3. The coverage ladder

The product does not need all five layers to be useful. It needs to know, and say, where a given
point sits.

| Level | What the platform can say | What it takes |
|---|---|---|
| **0** | Nothing — "this is outside every boundary we hold" | — |
| **1** | Which body and which ward | layer 2 |
| **2** | Which department, which PIO, how to escalate | layer 3 |
| **3** | What the deadline is and what breaching it costs | layer 1 |
| **4** | Which contract, which contractor, whether it is still in warranty | layer 4 |

Where things actually stand, as of 3 October 2026:

| Place | Roads | Waste | Water supply |
|---|---|---|---|
| **R/C, Borivali** | 4 | 3 | 2 — the department is known, and [no repair deadline is published](../01-research/11-pothole-sla-sources.md) |
| Rest of Greater Mumbai | 1 | 1 | 1 |
| Virar, Nagpur, anywhere else | 0 | 0 | 0 |

**The useful observation:** Nagpur reaches level 3 with *no Nagpur-specific research whatsoever*,
once layers 1 and 2 land — because level 3 is statutory. Most of the value is already written; it is
waiting on a boundary file.

---

## 4. Order of work

1. **Finish the hypothesis in one ward.** Do not scale an unproven product. That is what the v0.2
   gate exists for.
2. **Statewide boundaries.** One effort, unlocks level 1 everywhere. This is what Virar needs.
3. **The three Acts and the RTS notified-services list.** One effort, unlocks level 3 everywhere.
4. **Handbook harvesting**, crowd-assisted through RTI.
5. **MahaTenders and work-description geocoding.** The gate.

---

## 5. What not to do

**Do not repeat the per-ward department research 317 times.** Adding two departments to one ward on
3 October 2026 took about two hours of reading and archiving. At that rate, 317 bodies across six
categories is on the order of **four thousand hours**. That path does not end, and it is not what
the runtime department lookup was built to avoid ([D096](../00-overview/05-decision-log.md)).

**Do not widen the taxonomy to widen coverage.** The classifier has recognised six categories since
it was built. Coverage is a function of `authority_departments`, not of the model.

**Do not let coverage outrun sourcing.** Hard rule 2 — every displayed fact about a named party
carries a source reference and a retrieval timestamp — is harder across 317 portals that drop URLs
without notice. BMC's own Citizen's Charter URL already 404s. The archive is the evidentiary record,
not the parsed row.

---

## 6. The risk that grows fastest

Statewide means far more named parties on public surfaces, in far more jurisdictions, with far more
variable source quality. The exposure tiers and feature flags stop being theoretical the moment the
platform names a contractor outside Mumbai, and the moderation and takedown workflow becomes load-
bearing rather than planned. See the [legal framework](../01-research/04-legal-framework.md) §8.

---

## 7. Open questions this raised

- [ ] **How many urban local bodies are there?** 29 corporations + 246 councils + 42 nagar
      panchayats = **317**, and the state's own ULB data portal says **424**. The councils and
      panchayat figures were confirmed against the Directorate of Municipal Administration; the
      discrepancy is unexplained and should be settled before anything is planned against it.
- [ ] **Does the State Election Commission's 2025–26 ward delimitation exist as geometry**, or only
      as descriptive text? The answer decides whether layer 2 is an afternoon or a quarter.
- [ ] **Is pothole repair a notified service under the RTS Act?** Still the highest-value unanswered
      question in the project, and it is statewide.
- [ ] **Does MahaTenders carry any location field** beyond free text?
