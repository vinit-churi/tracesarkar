# ADR 0021 — A model may reword a generated post, never author one

**Status:** Accepted · **Date:** 2026-10-16 · **Amends:**
[ADR 0020](0020-share-text-is-generated-not-written-by-a-model.md) ·
**Related:** [ADR 0006](0006-claude-for-vision-and-extraction.md) ·
[share kit](../02-product/05-share-kit.md) ·
[accessibility and vernacular](../02-product/10-accessibility-and-vernacular.md) ·
[AI pipeline](../03-architecture/04-ai-pipeline.md)

## Context

[ADR 0020](0020-share-text-is-generated-not-written-by-a-model.md) kept every model out of the
share path and gave a reason that has not weakened: the share kit is the first surface that
composes prose **about a named party** which will be forwarded beyond any page the platform can
correct, and a fluent generator is good at exactly the thing that must never happen. Its own
alternatives table left the door open on one condition — *"Revisit only with an eval set that
measures invented facts, not tone."*

Two months of use show where the decision was right and where it was too broad.

It was right about **authorship**. What a post may say should be bounded by what the database
holds, not by how well a prompt was written.

It was too broad about **wording**. The English templates read acceptably because a person wrote
them. The Marathi and Hindi templates were written by the same person, who speaks neither, and the
consequence is visible in the composed output: grammatical but stilted, with English sentence
shapes carried across. Worse, the register errors are the kind a BMC desk notices — the glossary
already records one, that **विभाग** is an administrative ward and **प्रभाग** is an electoral
division, and the only reason that particular error is not in the files is that somebody happened
to know about it. The vernacular doc's answer is a native-speaker review, which is the right answer
and has no date on it.

A stilted post in the reader's own language is a post that does not get sent, and a share that does
not happen is the failure mode this whole system exists to prevent. ADR 0020's consolation — "the
failure mode is a stilted sentence, not a false one" — undercounts the cost of a stilted sentence
in a product whose distribution thesis is the share.

## Decision

**A model may reword the sentences the platform composed. It may not compose one, and nothing it
returns is published unless it can be proved to say the same facts.**

The job is narrowed to the smallest form the problem admits, in `internal/polish`:

- **The model never sees a fact set.** It receives the finished prose lines of a post that
  `internal/share` already composed, numbered, and returns the same numbers reworded. One line in,
  one line out.
- **The permalink and the hashtag run are never sent at all.** They are tokens other people resolve
  and search for, there is no better wording for them, and withholding the permalink means the
  provider is never told which issue the sentences belong to.
- **Every returned line is checked against the line that went out**, by `polish.Verify`:
  - every protected span — legal citation, short reference, department, ward code, authority,
    `@handle` — byte for byte, as often as it appeared;
  - every number as a multiset, so `48` cannot become `72` and a sighting count cannot be dropped;
  - no new `@handle`, `#tag` or URL;
  - the right script for the language;
  - the prohibited-pattern list from the share kit spec, the same one that runs on the platform's
    own text;
  - a line-length cap, and the channel's own window on the finished post.
- **A single failure anywhere discards the whole candidate** and the composed post is served. A
  half-polished post is refused too: two registers in one message read as two authors.
- **`internal/share` stays pure.** It reads no database, fetches nothing and calls nothing. The
  model is a layer above it, and the deterministic composition remains the thing that is always
  correct.
- **The feature is off unless configured**, with no default model, and a missing model at startup
  is a hard error rather than a silent fallback.
- **Rewording is not review.** `Post.Wording` records who chose the words; `Post.Review` records
  whether a native speaker has read the language. A model-worded Marathi post is still
  `review: "draft"`, and a test asserts it.

## Rationale

- **The bound is still structural.** It has moved from *"no model"* to *"the model's output must be
  provably the same facts"*, and both are properties of the code rather than of a prompt. Nothing
  in `polish` asks a model whether another model behaved; that would be the same trust problem one
  layer down.
- **The eval set ADR 0020 asked for exists and measures the right thing.** `polish.Corpus` is 16
  fact sets — 96 posts across all six categories, both channels and three languages — including the
  states that
  *remove* a sentence — never lodged, no deadline, no handle, one sighting — because those are where
  a generator is most likely to put one back. `polish.Measure` reports refusals by kind and
  `Score.InventedFacts` separates a changed claim from a long line. A live test runs it against a
  real provider and independently re-verifies every post that was published.
- **The determinism loss is bounded and paid for.** Two readers of the same issue should read the
  same post, which a model threatens. An in-process cache keyed on the hash of exactly what would
  be sent keeps an answer stable for its lifetime, and the deterministic text is what the cache
  falls back to.
- **The failure modes are all the cheap one.** A provider that is down, slow, rate-limited or
  wrong returns the post ADR 0020 shipped. There is no path on which the share kit fails because a
  model did.

## Alternatives

| Option | Why not |
|---|---|
| **Leave ADR 0020 as it stands and wait for the native-speaker review** | The right answer for correctness and an open-ended delay for everyone in Mumbai in the meantime. It is also not either/or: the review is still needed, still outstanding, and polish does not advance it by one line. |
| **A model writes the post from the fact set** | ADR 0020's answer holds unchanged. An unbounded surface, and no amount of prompt instruction converts into a test. |
| **Machine-translate the English post** | Rejected by the vernacular doc and obvious for text carrying hashtags, handles and a court citation. Note that polish is not this: each language's own composed lines are reworded in place, never derived from English. |
| **Let the model return the whole post as one string and diff it** | A diff cannot tell a reworded line from a moved one, so the only enforceable contract would be "nothing changed". Numbering the lines is what makes span-level verification possible. |
| **Send the permalink so the model can place it naturally** | It is already last, it needs no placement, and sending it would tell the provider which report this is. |
| **Score the rewording with a second model** | The judge needs the same trust the writer needed. Every check here is structural and cheap for that reason. |
| **Polish the vocabulary files offline instead of the posts at runtime** | Genuinely better in one way — it would be reviewable, diffable and free at request time — and worse in another: a template has to serve every fact set, so the model would be rewording sentences with `{hours}` and `{citation}` holes in them and could not see the sentence a reader sees. Worth doing *as well*, as a proposal a person applies; it is not a substitute, and it is not built yet. |

## Consequences

- **The native-speaker review is still open and still the only thing that can close it.** This ADR
  must not be read as having finished it. `review: "pending"` in
  `internal/share/i18n/share.mr.json` and `share.hi.json` is unchanged.
- **A second prompt file is now part of the product surface.**
  `internal/polish/prompts/share_polish_v1.md` carries register rules per language, and those rules
  are claims about Marathi and Hindi that deserve the same native-speaker reading as the vocabulary
  files.
- **The share kit now has a cost per issue.** Bounded by the cache, the breaker and a 10-second
  budget for the whole kit, and zero when the feature is off.
- **The model behind it is an open decision.** There is deliberately no default, and whatever is
  configured has to be named in the decision log with the measurement that justified it.
- **The card is still English.** Unchanged by this ADR; shaping Devanagari still needs a
  text-shaping dependency and a font.
