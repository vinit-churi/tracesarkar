// Package share builds the post a citizen sends, from facts the platform can
// already prove.
//
// The thesis is that the share is the distribution channel and the friction
// that kills sharing is writing the post, so the platform writes it
// (docs/02-product/05-share-kit.md). Everything here is pure: it takes a fact
// set and returns text. Nothing in this package reads a database, fetches an
// image, or calls a model — which is what makes every rule below testable.
//
// Three rules shape the whole package:
//
//   - A fact about a named party travels with its source or does not appear
//     (hard rule 2). That is why Handle carries SourceRef and RetrievedAt, and
//     why a deadline with no citation is dropped rather than shown as a bare
//     number.
//   - Text is generated per language from the fact set, never translated from
//     the English string (docs/02-product/10-accessibility-and-vernacular.md).
//     A translated hashtag-and-handle post is garbage, and a machine-
//     translated legal citation is a hazard, so citations pass through every
//     language unchanged.
//   - Nothing asserts wrongdoing (hard rule 3). Check enforces it on the
//     finished string, and Compose refuses to return text that fails.
package share

import (
	"fmt"
	"sort"
	"time"
)

// Lang is a language the kit can be generated in.
type Lang string

const (
	EN Lang = "en"
	MR Lang = "mr"
	HI Lang = "hi"
)

// Languages lists what can be generated, in the order the UI offers them.
// Marathi before Hindi: this is Mumbai.
func Languages() []Lang { return []Lang{EN, MR, HI} }

// Channel is where the post is going. The text differs by channel because the
// constraints differ, not because the facts do.
type Channel string

const (
	// WhatsApp is the primary channel: the card must stand alone, because it
	// will be forwarded far from the link (share kit §5).
	WhatsApp Channel = "whatsapp"
	X        Channel = "x"
)

// Channels lists the channels in the order the UI offers them. WhatsApp first,
// because that is where MMR civic complaint behaviour already lives.
func Channels() []Channel { return []Channel{WhatsApp, X} }

// XLimit is the post window on X, in runes rather than bytes — a Devanagari
// post counted in bytes would be rejected at a third of its real length.
const XLimit = 280

// HandleKind separates a desk from a minister.
type HandleKind string

const (
	// Institutional handles are the authority's own account. Safe on a first
	// report.
	Institutional HandleKind = "institutional"
	// Escalation handles are unlocked by an elapsed deadline and never before:
	// tagging a chief minister on every pothole trains everybody to ignore the
	// platform (share kit §4).
	Escalation HandleKind = "escalation"
)

// Handle is one verified account, with the provenance that lets it be printed.
//
// There is deliberately no way to carry a handle without a source: a social
// account is a fact about a named party, and hard rule 2 applies to it exactly
// as it applies to a contract value.
type Handle struct {
	Platform    string
	Handle      string
	Kind        HandleKind
	SourceRef   string
	RetrievedAt time.Time
	// FirstHand records whether the authority itself published this handle, as
	// opposed to it being attested by somebody else. It is printed in the
	// source list so a reader can weigh it, and it is never used to suppress a
	// handle — the source register's ✅/○ discipline, applied here.
	FirstHand bool
}

// Facts is everything a share kit may state.
//
// It is a copy of what the platform knows, not a pointer into the database, so
// that composing a post cannot widen what a public surface is allowed to see.
// Absence is meaningful throughout: an empty DeadlineCitation does not mean
// "look it up", it means no deadline may be stated.
type Facts struct {
	IssueID   string
	Permalink string

	Authority   string
	Ward        string
	Category    string
	Subcategory string

	FirstSeen time.Time
	Sightings int
	Status    string

	// Lodged is whether anybody has recorded telling the authority. The clock
	// runs from this and from nothing else, and the platform never does it.
	Lodged   bool
	LodgedOn *time.Time

	// DeadlineHours is stated only alongside a citation. DeadlineShortRef is
	// the form that fits in a post; DeadlineCitation is the full one. Neither
	// is ever translated.
	DeadlineHours    int
	DeadlineShortRef string
	DeadlineCitation string
	// DeadlineState mirrors issue.ClockState: not_started, running, elapsed,
	// no_deadline.
	DeadlineState string

	// Handles have already been through the registry. This package filters by
	// kind and channel; it never invents one.
	Handles []Handle

	// Contract is the published contract attribution, or nil when there is
	// none to publish. Nil is the normal case today: internal/attribute
	// decides a match but no public surface carries it yet, and a contractor
	// named on a thin match is the expensive error this platform is built to
	// avoid.
	Contract *Contract
}

// Source is the provenance line for a fact that named somebody.
type Source struct {
	Fact        string    `json:"fact"`
	Ref         string    `json:"ref"`
	RetrievedAt time.Time `json:"retrieved_at"`
	FirstHand   bool      `json:"first_hand"`
}

// ReviewState is how far the copy in a language has been checked.
type ReviewState string

const (
	// Reviewed means a native speaker has read it, or the language is English.
	Reviewed ReviewState = "reviewed"
	// Draft means the platform wrote it and nobody has checked it. It is
	// offered, because withholding it helps nobody, but it is labelled so the
	// client can say so and the user can edit before sending.
	Draft ReviewState = "draft"
)

// Post is one composed message.
type Post struct {
	Lang     Lang        `json:"lang"`
	Channel  Channel     `json:"channel"`
	Text     string      `json:"text"`
	Hashtags []string    `json:"hashtags"`
	Review   ReviewState `json:"review"`
	Sources  []Source    `json:"sources"`
	// Fallbacks names any term this language had no word for, so a gap in the
	// vocabulary is visible rather than silently filled from English.
	Fallbacks []string `json:"fallbacks,omitempty"`
	// Dropped names sentences removed to fit the channel's window, so a short
	// post can be told from a post that lost a fact.
	Dropped []string `json:"dropped,omitempty"`
}

// Kit is every post for one issue, plus what the image should carry.
type Kit struct {
	IssueID string          `json:"issue_id"`
	Posts   map[string]Post `json:"posts"`
	Card    Card            `json:"card"`
}

// Compose builds one post.
func Compose(f Facts, lang Lang, ch Channel) (Post, error) {
	v, err := vocabulary(lang)
	if err != nil {
		return Post{}, err
	}
	switch ch {
	case WhatsApp, X:
	default:
		return Post{}, fmt.Errorf("share: no template for channel %q", ch)
	}

	p := Post{Lang: lang, Channel: ch, Review: v.state(), Hashtags: hashtags(f)}
	lines, fallbacks, sources := build(f, v, ch, &p)
	p.Fallbacks, p.Sources = fallbacks, sources

	text := assemble(lines, f, p.Hashtags)
	if ch == X {
		text = fit(text, lines, f, p.Hashtags, &p)
	}
	p.Text = text

	// The gate, applied to the finished string rather than trusted to the
	// template above. Subcategory and authority are free text from upstream,
	// so a template that is clean today can still carry an accusation.
	if vs := Check(p.Text); len(vs) > 0 {
		return Post{}, fmt.Errorf("share: generated text fails the tone check: %s", vs[0])
	}
	return p, nil
}

// ComposeKit builds every post for an issue, skipping none: a client that has
// to special-case a missing language will show the user a hole.
func ComposeKit(f Facts) (Kit, error) {
	k := Kit{IssueID: f.IssueID, Posts: map[string]Post{}, Card: CardFor(f)}
	for _, lang := range Languages() {
		for _, ch := range Channels() {
			p, err := Compose(f, lang, ch)
			if err != nil {
				return Kit{}, err
			}
			k.Posts[string(lang)+"."+string(ch)] = p
		}
	}
	return k, nil
}

// hashtags are derived from the fact set, never translated: a hashtag is a
// token other people are already using, and localising it splits the stream.
func hashtags(f Facts) []string {
	tags := []string{}
	switch f.Category {
	case "road_defect":
		tags = append(tags, "#MumbaiRoads")
	case "waste":
		tags = append(tags, "#MumbaiWaste")
	case "water_drainage":
		tags = append(tags, "#MumbaiWater")
	}
	if w := wardTag(f.Ward); w != "" {
		tags = append(tags, w)
	}
	sort.Strings(tags)
	return tags
}
