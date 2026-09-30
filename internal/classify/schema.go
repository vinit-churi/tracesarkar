package classify

// ResultSchema is the JSON schema the model must answer in.
//
// It is built from the same taxonomy the guardrails validate against, so the
// model is constrained to categories the platform can actually recognise.
func ResultSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required": []string{
			"category", "subcategory", "severity", "hazard_to_life",
			"people_present", "image_quality", "is_civic_issue", "confidence",
			"rationale",
		},
		"properties": map[string]any{
			"category":       map[string]any{"type": "string", "enum": TopCategories()},
			"subcategory":    map[string]any{"type": "string"},
			"severity":       map[string]any{"type": "string", "enum": []string{"low", "medium", "high", "critical"}},
			"hazard_to_life": map[string]any{"type": "boolean"},
			"surface_type":   map[string]any{"type": "string"},
			"water_present":  map[string]any{"type": "boolean"},
			"people_present": map[string]any{"type": "boolean"},
			"image_quality":  map[string]any{"type": "string", "enum": []string{"good", "poor", "unusable"}},
			"is_civic_issue": map[string]any{"type": "boolean"},
			"confidence":     map[string]any{"type": "number", "minimum": 0, "maximum": 1},
			"alternatives":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"rationale":      map[string]any{"type": "string"},
		},
	}
}
