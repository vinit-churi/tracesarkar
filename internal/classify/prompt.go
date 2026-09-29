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
	b.WriteString("\n## Categories\n\nUse exactly these. Nothing else is routable.\n\n")
	for _, c := range TopCategories() {
		fmt.Fprintf(&b, "- `%s` — %s\n", c, strings.Join(Taxonomy()[c], ", "))
	}
	return b.String()
}
