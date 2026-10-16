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

// register is what each language needs said about it that the general
// instructions cannot say.
//
// These are the things a stilted draft gets wrong, written down from the
// glossary and from BMC's own usage rather than from a sense of good style —
// so the prompt can be reviewed against a document, and so a native speaker
// reading `internal/share/i18n` and a native speaker reading this file are
// being asked to check the same claims.
var register = map[share.Lang]string{
	share.EN: "The English is already written by a person and is usually fine. " +
		"Change a line only where it is genuinely ungrammatical or ambiguous.",

	share.MR: "Write Marathi as it is written in Mumbai, for a municipal " +
		"office.\n\n" +
		"- The corporation's own word for an administrative ward is **विभाग**. " +
		"**प्रभाग** is an *electoral* division and is a different thing: never " +
		"use it here, however natural it sounds.\n" +
		"- Prefer the plain word over the Sanskritised one where both exist and " +
		"the plain one is what people say. This is a complaint, not an " +
		"application.\n" +
		"- Punctuate as Marathi is punctuated in print today. Do not carry " +
		"English punctuation across mechanically.\n" +
		"- Keep case endings and postpositions correct on place names and ward " +
		"codes — a ward code is a Latin-script token with a Marathi " +
		"postposition attached, and getting that join right is most of what " +
		"makes the line read naturally.\n" +
		"- Do not translate an English department name, court name or citation " +
		"into Marathi. Leave it in English inside the Marathi sentence, which " +
		"is how it is written and said.",

	share.HI: "Write Hindi as it is written in Mumbai, for a municipal " +
		"office.\n\n" +
		"- Everyday Mumbai Hindi, not Sanskritised Hindi and not Urdu-leaning " +
		"Hindustani. A reader at a ward office and a neighbour on WhatsApp " +
		"should both find it ordinary.\n" +
		"- Use the punctuation Hindi actually uses, including the danda (।) " +
		"where a full stop belongs, unless the line ends in a Latin-script " +
		"reference, in which case end the sentence before the reference " +
		"instead of putting a danda after it.\n" +
		"- Keep postpositions and gender agreement correct, including on words " +
		"borrowed from English.\n" +
		"- Do not translate an English department name, court name or citation " +
		"into Hindi. Leave it in English inside the Hindi sentence.",
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
