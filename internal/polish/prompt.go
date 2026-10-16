package polish

import (
	"embed"
	"fmt"
	"strings"

	"github.com/vinit-churi/tracesarkar/internal/share"
)

//go:embed prompts
var promptFS embed.FS

// Prompt is a versioned system prompt, loaded from a file rather than written
// inline so it can be diffed, reviewed and evaluated, and so the version
// travels with whatever it produced (CLAUDE.md, "Working on the AI pipeline").
type Prompt struct {
	Name    string
	Version string
	Text    string
}

// LoadPrompt reads a prompt by file name, e.g. "share_polish_v1".
func LoadPrompt(name string) (Prompt, error) {
	raw, err := promptFS.ReadFile("prompts/" + name + ".md")
	if err != nil {
		return Prompt{}, fmt.Errorf("load prompt %s: %w", name, err)
	}
	text, version := string(raw), name
	if strings.HasPrefix(text, "---\n") {
		if end := strings.Index(text[4:], "\n---\n"); end >= 0 {
			header := text[4 : 4+end]
			text = strings.TrimLeft(text[4+end+5:], "\n")
			for _, l := range strings.Split(header, "\n") {
				if v, ok := strings.CutPrefix(strings.TrimSpace(l), "version:"); ok {
					version = strings.TrimSpace(v)
				}
			}
		}
	}
	return Prompt{Name: name, Version: version, Text: text}, nil
}

// register is what each language needs said that the general instructions
// cannot say.
//
// Every claim here now comes from the native-speaker review of 16 October
// 2026 ([D175]) rather than from a non-speaker's sense of good style, which is
// what it was built on and which was wrong in at least one place: it told the
// model to use the danda in Hindi, and the review settled that the platform's
// copy uses the full stop in both languages. A prompt that fights the reviewed
// vocabulary produces a post in two registers, and the guard cannot see it
// because punctuation is neither a span nor a number.
//
// [D175]: ../../docs/00-overview/05-decision-log.md
var register = map[share.Lang]string{
	share.EN: "The English is already written by a person and is usually fine. " +
		"Change a line only where it is genuinely ungrammatical or ambiguous.",

	share.MR: "Write Marathi as it is written in Mumbai, for a municipal " +
		"office. Every point below was settled by a native speaker reviewing " +
		"this platform's own copy; follow them rather than your own " +
		"preference.\n\n" +
		"- **Punctuate with the full stop `.`** Modern Marathi — newspapers, " +
		"government circulars, WhatsApp — uses it exclusively. The danda `।` " +
		"belongs to Sanskrit and old poetry and is wrong here. Never convert " +
		"a `.` to a `।`.\n" +
		"- The corporation's own word for an administrative ward is " +
		"**विभाग**. **प्रभाग** is an *electoral* division and is a different " +
		"thing: never use it here, however natural it sounds.\n" +
		"- **Do not use प्राधिकरण in running prose.** Nobody says it. The " +
		"natural form is संबंधित विभाग / संबंधित विभागास — \"the concerned " +
		"department\". यंत्रणा: is the label beside the authority's name.\n" +
		"- Civic and administrative vocabulary — प्रलंबित, दखल, मुदत, " +
		"कार्यवाही — is right for this reader and is understood by ordinary " +
		"residents. Do not simplify it away.\n" +
		"- Prefer the plain word over the Sanskritised one where both exist " +
		"and the plain one is what people say: सांडपाणी or ड्रेनेज, not " +
		"मलनिस्सारण.\n" +
		"- Keep case endings and postpositions correct on place names and " +
		"ward codes. A ward code is a Latin-script token with a Marathi " +
		"postposition attached, and getting that join right is most of what " +
		"makes a line read naturally.\n" +
		"- Do not translate an English department name, court name or " +
		"citation into Marathi. Leave it in English inside the Marathi " +
		"sentence, which is how it is written and said.",

	share.HI: "Write Hindi as it is written in Mumbai, for a municipal " +
		"office. Every point below was settled by a native speaker reviewing " +
		"this platform's own copy; follow them rather than your own " +
		"preference.\n\n" +
		"- **Punctuate with the full stop `.`, not the danda `।`.** The danda " +
		"is correct standard Hindi and either is acceptable in principle, but " +
		"this platform uses `.` throughout — it is what urban Mumbai Hindi " +
		"uses beside Latin ward codes and URLs, and it matches the Marathi. " +
		"Never convert a `.` to a `।`.\n" +
		"- **Do not use प्राधिकरण.** Nobody in Mumbai says प्राधिकरण को बताया " +
		"गया. Use संबंधित विभाग, and सूचित किया गया rather than बताया गया. " +
		"विभाग: is the label beside the authority's name.\n" +
		"- Everyday Mumbai Hindi, not Sanskritised Hindi and not Urdu-leaning " +
		"Hindustani. A ward officer and a neighbour on WhatsApp should both " +
		"find it ordinary.\n" +
		"- Avoid literal renderings of English grammar. A relative clause " +
		"translated word for word (\"the document it came from\") reads as " +
		"translationese; say it the way Hindi says it.\n" +
		"- Keep postpositions and gender agreement correct, including on words " +
		"borrowed from English.\n" +
		"- Do not translate an English department name, court name or " +
		"citation into Hindi. Leave it in English inside the Hindi sentence.",
}

// WithLanguage appends what this language needs said, plus the channel's own
// constraint. Generated rather than written into the prompt file so a new
// language is one map entry and cannot be forgotten.
func (p Prompt) WithLanguage(lang share.Lang, ch share.Channel) string {
	var b strings.Builder
	b.WriteString(p.Text)
	fmt.Fprintf(&b, "\n\n## This request: %s\n\n", langName(lang))
	if r := register[lang]; r != "" {
		b.WriteString(r)
		b.WriteString("\n")
	}
	if ch == share.X {
		fmt.Fprintf(&b, "\nThe whole message must fit %d characters including the "+
			"link and the tags, which are not shown to you. Every line you return "+
			"must therefore be the same length as the one you were given or "+
			"shorter.\n", share.XLimit)
	}
	return b.String()
}

func langName(l share.Lang) string {
	switch l {
	case share.MR:
		return "Marathi (मराठी)"
	case share.HI:
		return "Hindi (हिन्दी)"
	default:
		return "English"
	}
}
