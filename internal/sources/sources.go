// Package sources loads the source register and enforces, at runtime, the rules
// about which sources may be collected at which exposure tier.
//
// See data/sources.yaml, D027 (Mahatenders is manual only) and D045 (the
// personal tier may run a candidate source whose terms are unreviewed).
package sources

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Exposure tiers (D044).
const (
	TierPersonal = "personal"
	TierFlagged  = "flagged"
	TierPublic   = "public"
)

// Source is one entry in the register.
type Source struct {
	ID              string    `yaml:"id"`
	Name            string    `yaml:"name"`
	URL             string    `yaml:"url"`
	Category        string    `yaml:"category"`
	Acquisition     string    `yaml:"acquisition"`
	Cadence         string    `yaml:"cadence"`
	Licence         string    `yaml:"licence"`
	TermsReviewedOn *string   `yaml:"terms_reviewed_on"`
	TermsReviewedBy *string   `yaml:"terms_reviewed_by"`
	RobotsOK        *bool     `yaml:"robots_ok"`
	Status          string    `yaml:"status"`
	Blocklist       []string  `yaml:"-"`
	RawBlocklist    yaml.Node `yaml:"blocklist"`
	ArchivePrefix   string    `yaml:"archive_prefix"`
	Parser          string    `yaml:"parser"`
	Blocker         string    `yaml:"blocker"`
	Notes           string    `yaml:"notes"`
}

// Register is the whole file, indexed by id.
type Register struct {
	byID  map[string]Source
	order []string
}

type registerFile struct {
	Sources []Source `yaml:"sources"`
}

// Load reads and validates the register.
func Load(path string) (*Register, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read source register: %w", err)
	}
	var file registerFile
	if err := yaml.Unmarshal(body, &file); err != nil {
		return nil, fmt.Errorf("parse source register: %w", err)
	}

	reg := &Register{byID: make(map[string]Source, len(file.Sources))}
	for _, src := range file.Sources {
		if strings.TrimSpace(src.ID) == "" {
			return nil, fmt.Errorf("source register: an entry has no id")
		}
		if _, dup := reg.byID[src.ID]; dup {
			return nil, fmt.Errorf("source register: id %q appears twice", src.ID)
		}
		src.Blocklist = parseBlocklist(src.RawBlocklist)
		src.RawBlocklist = yaml.Node{}
		reg.byID[src.ID] = src
		reg.order = append(reg.order, src.ID)
	}
	if len(reg.byID) == 0 {
		return nil, fmt.Errorf("source register: no sources found")
	}
	return reg, nil
}

// parseBlocklist accepts either a YAML list or a comma-separated string.
func parseBlocklist(node yaml.Node) []string {
	switch node.Kind {
	case yaml.SequenceNode:
		var out []string
		for _, item := range node.Content {
			if v := strings.TrimSpace(item.Value); v != "" {
				out = append(out, v)
			}
		}
		return out
	case yaml.ScalarNode:
		var out []string
		for _, part := range strings.Split(node.Value, ",") {
			if v := strings.TrimSpace(part); v != "" {
				out = append(out, v)
			}
		}
		return out
	default:
		return nil
	}
}

// Get returns one source.
func (r *Register) Get(id string) (Source, bool) {
	src, ok := r.byID[id]
	return src, ok
}

// All returns every source in file order.
func (r *Register) All() []Source {
	out := make([]Source, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.byID[id])
	}
	return out
}

// IDs returns every source id, sorted.
func (r *Register) IDs() []string {
	out := append([]string(nil), r.order...)
	sort.Strings(out)
	return out
}

// runnableStatuses are the states a scheduled collector may run in.
var runnableStatuses = map[string]bool{
	"candidate": true,
	"planned":   true,
	"reviewing": true,
	"active":    true,
}

// MayRun reports whether a scheduled ingester may collect this source at the
// given tier, and why not when it may not.
func MayRun(src Source, tier string) (bool, string) {
	if !runnableStatuses[src.Status] {
		return false, fmt.Sprintf("status is %q", src.Status)
	}
	if strings.EqualFold(src.Acquisition, "manual") {
		return false, "acquisition is manual: documents arrive by hand, never on a schedule"
	}
	if src.RobotsOK != nil && !*src.RobotsOK {
		return false, "robots.txt disallows automated collection (hard rule 8)"
	}
	if tier != TierPersonal {
		if src.TermsReviewedOn == nil || strings.TrimSpace(*src.TermsReviewedOn) == "" {
			return false, fmt.Sprintf("terms are unreviewed, which the %s tier requires (D045)", tier)
		}
	}
	return true, ""
}

// MayPublish reports whether this source's output may reach a flagged or public
// surface.
func MayPublish(src Source) (bool, string) {
	if src.TermsReviewedOn == nil || strings.TrimSpace(*src.TermsReviewedOn) == "" {
		return false, "terms are unreviewed; personal tier only (D045)"
	}
	if !runnableStatuses[src.Status] {
		return false, fmt.Sprintf("status is %q", src.Status)
	}
	return true, ""
}
