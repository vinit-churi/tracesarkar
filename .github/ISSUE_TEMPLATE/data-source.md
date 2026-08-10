---
name: Data source
about: Propose a new external data source
title: "[D] Source: "
labels: data, needs-legal-review
---

## Source

- **Name:**
- **URL:**
- **Category:** procurement / geospatial / weather / news / corporate / other
- **What it gives us:**

## Why we need it

<!-- Which feature depends on this? Link the doc. -->

## Acquisition

- [ ] Documented API
- [ ] Bulk download
- [ ] Scraping
- [ ] RTI
- [ ] Manual

**Cadence:**

## Legal review (required before any ingester is built)

- [ ] Terms of use read — link:
- [ ] Automated access permitted? yes / no / unclear
- [ ] `robots.txt` checked
- [ ] Licence: <!-- and any attribution obligation -->
- [ ] Licence compatible with AGPL-3.0 / CC BY-SA 4.0
- [ ] Contains personal data? If yes, is our processing lawful for our purpose?

> **If automated access is not permitted, do not build an ingester.** Record the decision and use RTI.

## Format and risks

- **Format:** HTML / JSON / PDF (scanned?) / GeoJSON / other
- **Known fragility:**
- **Rate limits observed:**
- **Historical availability:**

## Source register entry

```yaml
- id:
  name:
  url:
  category:
  acquisition:
  cadence:
  licence:
  terms_reviewed_on:
  terms_reviewed_by:
  robots_ok:
  archive_prefix:
  status: planned
```
