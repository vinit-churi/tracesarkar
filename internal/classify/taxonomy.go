package classify

import (
	"sort"
	"strings"
)

// The v1 taxonomy, from docs/02-product/04-snap-to-action-pipeline.md.
//
// A category exists here only if something downstream can route it. Adding one
// without a department mapping produces a report that is classified and then
// goes nowhere, which is worse than an honest "we are not sure".
var taxonomy = map[string][]string{
	"road_defect": {
		"pothole", "crack", "subsidence", "missing manhole cover", "open manhole",
		"damaged speed breaker", "utility-dig damage", "missing/faded markings",
	},
	"waste": {
		"uncollected garbage", "illegal dumping", "debris", "overflowing bin",
		"dead animal", "burning waste",
	},
	"water_drainage": {
		"waterlogging", "burst pipeline", "leakage", "blocked drain",
		"silted nallah", "sewage overflow",
	},
	"street_furniture": {
		"streetlight out", "damaged railing", "broken footpath",
		"missing bollard", "damaged signage",
	},
	"structural": {
		"cracked bridge/FOB", "distressed building", "unsafe scaffolding",
		"collapsed wall",
	},
	"environment": {
		"mangrove destruction", "creek dumping", "effluent discharge",
		"tree felling", "air pollution source", "encroachment on water body",
	},
}

// KnownCategory reports whether a category can be routed.
func KnownCategory(c string) bool {
	_, ok := taxonomy[strings.ToLower(strings.TrimSpace(c))]
	return ok
}

// KnownSubcategory reports whether a subcategory belongs to its category.
func KnownSubcategory(category, sub string) bool {
	subs, ok := taxonomy[strings.ToLower(strings.TrimSpace(category))]
	if !ok {
		return false
	}
	for _, s := range subs {
		if strings.EqualFold(s, strings.TrimSpace(sub)) {
			return true
		}
	}
	return false
}

// TopCategories lists the categories to offer when the classifier is unsure.
func TopCategories() []string {
	out := make([]string, 0, len(taxonomy))
	for c := range taxonomy {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// Taxonomy returns the whole vocabulary, for the prompt and the eval harness.
func Taxonomy() map[string][]string {
	out := make(map[string][]string, len(taxonomy))
	for k, v := range taxonomy {
		out[k] = append([]string(nil), v...)
	}
	return out
}
