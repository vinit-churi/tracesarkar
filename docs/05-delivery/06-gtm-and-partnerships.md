# Go-to-market and partnerships

Distribution is the hardest unsolved problem in Indian civic tech. Every predecessor has a large
install base and a small active base.

---

## 1. The distribution thesis

**Growth comes from the share, not from marketing.** A civic app that must buy users cannot afford
to exist. The share kit is the growth engine, and it works only if the shared artefact carries a
fact worth forwarding.

```
report ──▶ attribution fact ──▶ share ──▶ WhatsApp group ──▶ new reporter ──▶ report
                                    │
                                    └──▶ public pressure ──▶ resolution ──▶ proof it works
```

Both loops must close. The second one — visible resolution — is what converts a curious first-time
reporter into a returning one.

---

## 2. Launch sequence

### Phase 1 — one ward, hand-to-hand

Kandivali West. No press, no launch post.

| Action | Why |
|---|---|
| Recruit 5–10 **RWA secretaries and society-federation organisers** directly | Each is worth several hundred individual users and will onboard their own group |
| Walk one corridor with each of them, reporting together | Reveals capture friction faster than any usability study |
| Join 3–5 existing ward WhatsApp groups (with permission) | This is where civic complaint behaviour already lives |
| Ship weekly, visibly, on their feedback | Organisers evangelise tools that respond to them |

**Target:** 200 reports from non-team members. Success is depth in one ward, not breadth.

### Phase 2 — the first story

One journalist, one story, sourced from platform data.

| Action | Why |
|---|---|
| Identify 3–5 Mumbai reporters who cover civic issues | Small, identifiable set |
| Offer a concrete, checkable finding — not a product demo | Reporters want a number, not a pitch |
| Provide the provenance manifest with every figure | Survives an editor and a PR pushback |
| Be reachable by phone for verification | Non-negotiable |

One published story does more for legitimacy — and for the political cost of ignoring the platform —
than 10,000 installs.

### Phase 3 — ward by ward

Expand only where the prerequisites exist:

- [ ] Ward boundary verified
- [ ] Department mapping and SLA confirmed
- [ ] PIO details obtained
- [ ] At least one local organiser recruited
- [ ] Contract data ingested for the ward

**Never launch a ward whose boundary data is unverified.** A wrong-routing reputation is
unrecoverable and travels faster than the product.

### Phase 4 — the second corporation

TMC or KDMC. The point is to prove the model generalises beyond BMC — a different portal, a different
department structure, a different ward scheme. Expect it to be harder than it looks.

---

## 3. Channel strategy

| Channel | Role | Notes |
|---|---|---|
| **WhatsApp** | Primary | Intake bot + share cards. Where the behaviour already is. Requires Business API onboarding, template approval, and a per-message budget |
| **X / Twitter** | Public pressure | Tagged posts to institutional handles. Escalation handles gated on SLA breach |
| **Instagram** | Reach among younger users | Story-format share cards |
| **Local press** | Legitimacy | Phase 2 onward |
| **RWA / society federations** | Organiser-led acquisition | The highest-leverage channel |
| **College civic clubs** | Volunteer verification missions | Good for coverage, needs supervision |
| **Search / SEO** | Long tail | Public permalink pages with structured data; "pothole complaint Kandivali" should find us |
| **Paid acquisition** | **None** | Cannot be afforded and would mask a broken loop |

---

## 4. Partnership targets

Ordered by expected value.

| Partner type | Ask | Offer |
|---|---|---|
| **Newsrooms** (Mumbai dailies, Marathi press) | Use the data; cite it | Watchdog TUI access, saved alerts, embeddable heatmaps, verification support |
| **RTI and civic activists** | Use the escalation tools; tell us where they break | Automation of the paperwork they do by hand |
| **Legal aid / pro-bono counsel** | Review the escalation templates; take referred cases | A pipeline of well-documented, well-timed cases |
| **CivicDataLab / Open Contracting Partnership** | OCDS mapping guidance | An MMR procurement corpus published as OCDS |
| **DataMeet** | Boundary data collaboration | Digitised MMR municipal boundaries contributed back |
| **IITM / IIT-B (MESONET, mumbaiflood.in)** | Data access terms for a public platform | Ground-truth flood reports from citizens |
| **Academic urban planning** | Analysis and validation | Bulk exports, stable schemas, versioned datasets |
| **Municipal corporations** | Eventually: an Open311 endpoint, or consume our API | Deduplicated, correctly-routed, evidence-backed reports; proof-of-work upload path |

**Not partners:** political parties, contractors, and any organisation whose association would
compromise perceived neutrality. This is not negotiable and should be stated publicly.

---

## 5. The municipal conversation

The framing matters enormously. TraceSarkar is adversarial to opacity, not to municipal staff.

**Opening position:**

> "The Bombay High Court has directed the State to maintain a single-window complaint system with
> photograph upload, tracking, multiple reporting modes, and 48-hour pothole response. That direction
> dates back a decade in this PIL line and the Court has repeatedly found compliance inadequate. We
> have built something that does part of that. We are not asking for money. We are asking whether
> you want the deduplicated, correctly-routed feed."

What we offer a corporation:

- Deduplicated issues instead of 200 tickets for one pothole
- Correct routing, so a ward engineer is not blamed for an MSRDC road
- A proof-of-work upload path so a genuine fix is publicly recorded
- Free API access; no SaaS contract, ever

What we will not do:

- Suppress adverse data as a condition of cooperation
- Give any authority the ability to close an issue without citizen confirmation
- Build a paid dashboard (see [personas §P4](../02-product/01-personas-and-jtbd.md))

---

## 6. Sustainability

The platform must survive without compromising its accountability function.

| Model | Verdict |
|---|---|
| **Grants** (civic tech, journalism, open data foundations) | **Primary.** Aligned incentives; well-documented open work is fundable |
| **Individual donations** | Secondary; supports independence and is a legitimacy signal |
| **Institutional API access** (newsrooms, researchers, at cost) | Possible; must never gate the public data tier |
| **Municipal SaaS** | **Rejected.** Would make the customer the entity being held accountable |
| **Advertising** | Rejected |
| **Selling citizen data** | Rejected absolutely |
| **Contractor "reputation management"** | Rejected absolutely; the offer will be made |

The cost envelope is deliberately small (see [cost model](../06-operations/04-cost-model.md)) so that
survival does not require compromise.

---

## 7. Anti-goals

| Anti-goal | Why |
|---|---|
| Viral growth before the loop closes | Users who report once and see nothing happen do not come back, and do not come back a second time either |
| National expansion before MMR works | The jurisdiction and procurement work is region-specific and does not generalise cheaply |
| Feature announcements as marketing | The only announcement worth making is a confirmed fix or a published story |
| Partisan framing | Guarantees capture and destroys the neutrality the data depends on |
| Overstating impact | The credibility is the asset. Publish the caveats with the numbers |

---

## 8. Twelve-month narrative goal

By month twelve, the sentence we want a Mumbai journalist to be able to write without our help:

> *"According to TraceSarkar, which cross-references citizen reports against municipal contract data,
> 38 of the 61 potholes reported in R/S ward this monsoon are on stretches still under contractor
> warranty."*

Everything in the roadmap is in service of making that sentence true, checkable, and defensible.
