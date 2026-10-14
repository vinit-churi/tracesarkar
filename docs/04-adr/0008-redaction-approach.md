# ADR 0008 — On-device-first redaction with a self-hosted fallback

**Status:** Accepted · **Date:** 2026-08-10

## Context

Citizen photographs of public places routinely contain identifiable faces and vehicle number plates.
Under the DPDP Act 2023 and the DPDP Rules 2025, those make the photograph personal data. Public
display of un-redacted images would be a compliance failure and an ethical one.

The redaction must happen **before** any public derivative exists, and ideally before the image
leaves the device.

## Decision

A two-tier redaction pipeline:

1. **On-device (preferred).** A small face and plate detector runs in the mobile client. The public
   derivative is produced locally with blur applied; both the archival original and the redacted
   derivative are uploaded.
2. **Server-side fallback.** For web uploads, WhatsApp intake, and devices where on-device inference
   is unavailable, detection runs on **our own infrastructure** using a locally-hosted model — never
   via a third-party API.

If detection fails or returns low confidence, the image is **held for review** rather than published.

Redaction metadata (`model_version`, counts, confidence) is stored on `report_media.redaction`, so a
detector improvement can trigger a re-run over the corpus.

## Rationale

- Sending an un-redacted image containing an identifiable person to a third-party API in order to
  find the person to redact is self-defeating.
- On-device redaction also reduces upload size and improves the offline-capture experience.
- Holding uncertain images is the right default: a delayed publication is a small cost, an
  un-redacted face is not.

## Alternatives

| Option | Why not |
|---|---|
| **No redaction, rely on "public place" reasoning** | Legally weak under DPDP; ethically wrong; and a single viral image of an identifiable bystander would be a reputational event. |
| **Third-party redaction API** | Sends the exact data we are protecting to a third party, and adds a per-image cost to every upload. |
| **Manual moderation only** | Does not scale past the first few hundred reports per day. |
| **Blur everything indiscriminately** | Destroys the evidentiary value of the photograph. |

## Consequences

- **This is the one place the Go-only decision ([ADR 0002](0002-go-backend.md)) has a seam.** The
  server-side fallback needs an ONNX runtime binding or a small sidecar service. Accepted; it is a
  contained, well-understood dependency with no network egress.
- Model choice and licensing for the detector must be reviewed (permissive licence required, given
  AGPL distribution).
- An adversarial evaluation set is required before launch: faces at distance, partial occlusion,
  reflections, plates at angle, night conditions, and Indian plate formats specifically.
- The archival original retains faces and plates. Access to it is break-glass, logged, and
  justified — see [security and privacy](../03-architecture/11-security-and-privacy.md).

## What was built, 13 October 2026

The server-side tier exists; the on-device tier does not. Both detectors run in the Go process with
no network egress, which keeps the seam this ADR anticipated smaller than expected — no ONNX
runtime and no sidecar.

| | What runs | Licence |
|---|---|---|
| Faces | pigo's Pico cascade, embedded in the binary, scanned upright and about eighteen degrees either way | MIT |
| Plates | An edge-density localiser written here: blur, Sobel-x, close the runs, label, filter by aspect and fill | — |

Three things this ADR said, now enforced rather than intended:

- **Held rather than published on any doubt.** A photograph checked for faces but not plates is
  treated as unchecked, and the classifier seeing people where the detector found no face holds it.
- **A record, not a flag.** `report_media` carries the tool, the time, the output object and the
  verdict (migration 0025), so a detector improvement re-scans by matching the tool string. A CHECK
  constraint refuses a published verdict with no output object.
- **The permissive-licence review** this ADR required: pigo is MIT, vendored at
  `internal/redact/cascade/` with its licence beside it.

**The adversarial set is still outstanding, and it is the gate.** Run over every stored capture on
13 October, the detector reached a verdict on all five real ones and produced four false positives
on litter and pavement — the harmless direction. It measured no recall at all, because none of the
five contains a face or a plate. Faces at distance, partial occlusion, reflections, night, and
Indian plate formats at angle remain unmeasured, and no public image should exist until they are
not.

## The set, 14 October 2026

The set this ADR required now exists as a surface rather than as a requirement: `/truth/`, where
the person who took a photograph drags a box around every face and every number plate in it, marks
the conditions that make it hard from a fixed vocabulary, and says whether anybody is visible at
all. The judgement is stored in fractions of the image (migration 0026), so it survives the
photograph being resized, and `ingest redact-eval` runs both detectors over the originals and
scores them against it.

Three rules the harness keeps:

- **Recall is destroyed pixels, not box overlap.** A face counts as covered when 95% of it is gone,
  by any region, whatever kind the detector thought it was finding ([D133](../00-overview/05-decision-log.md)).
- **A figure with nothing behind it is withheld.** Recall over a set containing no face prints as a
  dash, not as 100%.
- **The headline is a count.** How many faces and plates would have reached a public image, given
  what the gate would have decided about that photograph.

**The set is still empty of positives, and that is still the gate.** Nothing in the five stored
captures contains a face or a plate, so the field walk owes this the one kind of photograph it has
not been producing: people and vehicles, deliberately, at distance, at night, partly occluded, and
Indian plates at an angle.

**What one annotated photograph already bought.** The first run found the plate localiser blind to
any plate large in the frame — it closes gaps between character strokes with a constant in pixels,
and a plate two metres away has gaps wider than that constant. It found a plate across the road and
missed the legible one. The pass now runs at full size, half and quarter; the tool string is
`edge-density-2`, and `ingest redact --rescan` re-scanned every stored capture so the corpus is not
left holding the old answer ([D134](../00-overview/05-decision-log.md)).

