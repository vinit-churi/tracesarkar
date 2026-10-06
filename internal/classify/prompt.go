package classify

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed prompts
var promptFS embed.FS

// Prompt is a versioned system prompt, loaded from a file rather than written
// inline so it can be diffed, reviewed and evaluated. The version travels with
// every classification it produces (ADR 0006).
type Prompt struct {
	Name    string
	Version string
	Text    string
}

// LoadPrompt reads a prompt by file name, e.g. "road_defect_v1".
func LoadPrompt(name string) (Prompt, error) {
	raw, err := promptFS.ReadFile("prompts/" + name + ".md")
	if err != nil {
		return Prompt{}, fmt.Errorf("load prompt %s: %w", name, err)
	}

	text := string(raw)
	version := name

	// Strip the front matter and take the version from it, so the version
	// recorded against an artefact is the one in the file rather than one a
	// caller passed in.
	if strings.HasPrefix(text, "---\n") {
		if end := strings.Index(text[4:], "\n---\n"); end >= 0 {
			header := text[4 : 4+end]
			text = strings.TrimLeft(text[4+end+5:], "\n")
			for _, line := range strings.Split(header, "\n") {
				if v, ok := strings.CutPrefix(strings.TrimSpace(line), "version:"); ok {
					version = strings.TrimSpace(v)
				}
			}
		}
	}

	return Prompt{Name: name, Version: version, Text: text}, nil
}

// WithTaxonomy appends the routable vocabulary to the prompt.
//
// It is generated from the same map the guardrails check against, so the
// prompt and the validation can never drift apart — a category the model is
// told about is by construction one the platform can route.
func (p Prompt) WithTaxonomy() string {
	var b strings.Builder
	b.WriteString(p.Text)
	// Written as a closed list, one value per line, with the instruction
	// naming the subcategory field. A comma-joined run after the category read
	// as description, and the model described the photograph in its own words
	// instead of choosing — every such answer has to be confirmed by hand and
	// cannot be scored against an evaluation set labelled from this vocabulary.
	b.WriteString("\n## Categories and subcategories\n\n")
	b.WriteString("`category` must be exactly one of the names below.\n")
	b.WriteString("`subcategory` must be exactly one of the values listed under " +
		"that category, copied verbatim. Do not combine two of them, do not add " +
		"words, do not describe the photograph here — there is a `rationale` " +
		"field for that. If none fits, choose the closest and say why in " +
		"`rationale`.\n")
	for _, c := range TopCategories() {
		fmt.Fprintf(&b, "\n### %s\n\n", c)
		for _, sub := range Subcategories(c) {
			fmt.Fprintf(&b, "- %s\n", sub)
		}
	}
	return b.String()
}
