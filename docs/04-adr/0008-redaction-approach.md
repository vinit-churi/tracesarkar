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
