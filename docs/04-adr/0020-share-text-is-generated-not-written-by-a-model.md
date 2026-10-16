# ADR 0020 — Share text is generated from the fact set, not written by a model

**Status:** Accepted, amended in part by
[ADR 0021](0021-a-model-may-reword-a-generated-post-never-author-one.md) ·
**Date:** 2026-10-14 · **Related:**
[ADR 0006](0006-claude-for-vision-and-extraction.md) ·
[ADR 0007](0007-never-auto-file.md) ·
[share kit](../02-product/05-share-kit.md) ·
[accessibility and vernacular](../02-product/10-accessibility-and-vernacular.md)

## Context

The platform's distribution thesis is that the share is the channel, and that the friction which
kills sharing is *writing the post*. So the platform writes it — in the user's language, per
channel, pre-loaded with the facts that make the post land.

That makes the share kit the first surface where the platform composes prose **about a named party**
that will be forwarded beyond any page it can correct. A permalink can be fixed. A WhatsApp forward
cannot.

[ADR 0006](0006-claude-for-vision-and-extraction.md) already puts a model in the pipeline for
vision and extraction, and the obvious move was to use one here too: give it the fact set and the
tone rules and ask for three languages. It would write better sentences than a template.

It would also, eventually, write a fact.

## Decision

**No model is in the generation path. Every post is assembled from fields the platform can already
prove, by code, in `internal/share`.**

- The package is pure: it takes a `Facts` value and returns text. It reads no database, fetches no
  image and calls nothing over the network.
- `Facts` is a copy of `store.PublicIssueView` — the same coarsened, reference-free view the public
  permalink is built from — so a fact the public page may not show cannot reach a post that will be
  forwarded past it. The exact position and the complaint reference are absent by construction, not
  by filtering.
- Each language is one reviewable JSON file under `internal/share/i18n/`, carrying its own review
  state. Text is generated per language from the fact set, never translated from the English string.
- The prohibited-pattern list from the share kit spec runs over the **finished string**, and
  `Compose` refuses to return text that fails it.

## Rationale

- **What a post may say is bounded by what the database holds**, rather than by how well a prompt
  was written. The bound is structural, and it does not drift.
- **It is testable.** "A deadline with no citation is not stated at all" is an assertion in a test,
  not an instruction in a prompt that a model follows most of the time. So are "an escalation handle
  waits for an elapsed deadline", "an unlodged issue never claims anybody was told", and "the
  citation survives every language unchanged".
- **It is deterministic.** The same issue produces the same bytes, so a post can be cited, and two
  people reading the same share kit are reading the same thing.
- **The failure mode is a stilted sentence, not a false one.** The spec's own answer to a stilted
  sentence is already in place: the text is editable, and a user who rewrites it is likelier to
  send it.
- The tone check runs anyway, because the templates are not the only author — the authority name
  comes from a database, the subcategory from a classifier, and the user may edit before sending.

## Alternatives

| Option | Why not |
|---|---|
| **A model writes the post from the fact set** | Better prose, and an unbounded surface. The thing that must never happen — an asserted fact about a named party in a forwarded image — is exactly what a fluent generator is good at producing, and no amount of prompt instruction converts into a test. Revisit only with an eval set that measures invented facts, not tone. **That condition was met on 16 October 2026 for the narrower question of *rewording*, and only that: see [ADR 0021](0021-a-model-may-reword-a-generated-post-never-author-one.md). Authorship from the fact set remains refused.** |
| **A model writes it, then the prohibited-pattern list filters it** | The list catches accusation, insult and party reference. It does not catch a confident, well-worded, wrong contract number. Filtering catches the vocabulary of a bad post, not its claims. |
| **Machine-translate the English post into Marathi and Hindi** | Explicitly rejected by the vernacular doc, and obviously so for text carrying hashtags, handles and a court citation. |
| **Withhold Marathi and Hindi until a native speaker has reviewed them** | Helps nobody in Mumbai in the meantime. They ship as drafts, labelled in the API and on the page, and editable. |
| **Burn the fact block onto the image in all three languages** | The Go fonts carry no Devanagari and nothing in the standard library shapes a complex script, so conjuncts and matras would misplace. A card with broken Marathi is worse than an English card beside a Marathi post. |

## Amendment, 16 October 2026

[ADR 0021](0021-a-model-may-reword-a-generated-post-never-author-one.md) permits a model to
**reword** lines this package composed, under `polish.Verify`, and keeps everything above intact
for **authorship**. The reading that changed: "no model in the generation path" was the right rule
for what a post may *say* and too broad a rule for how it may be *worded*, and the cost of the
broad version fell almost entirely on Marathi and Hindi.

## Consequences

- **The card is English** while the post carries the language. Recorded here so it is a known
  limitation rather than an oversight; shaping Devanagari needs a text-shaping dependency and a
  font, which is a deliberate decision to take later.
- **The vocabulary is a review task, not a code task.** `internal/share/i18n/share.mr.json` and
  `share.hi.json` are what a native speaker is handed, and `review: "pending"` is what they change.
- **The handle registry is a maintenance commitment.** Rows carry a `verify_by` date and stop being
  served when it passes, so a registry nobody re-checks degrades into silence rather than into
  tagging a stranger ([D136](../00-overview/05-decision-log.md)).
- A contract attribution is **absent** from every post today. `internal/attribute` decides a match,
  but no public surface carries one yet, and the share kit prints a contract only when it has a
  source reference to print with it.
