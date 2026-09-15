package llm

import (
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	tests := []struct {
		name          string
		category      string
		expectedParts []string
		exactMatch    bool
	}{
		{
			name:       "Empty category",
			category:   "",
			exactMatch: true,
		},
		{
			name:       "Unknown category",
			category:   "unknown_category_xyz",
			exactMatch: true,
		},
		{
			name:     "Coding category",
			category: "coding",
			expectedParts: []string{
				DefaultSystemPrompt,
				"Approach (2-3 sentences)",
				"Complexity:",
			},
		},
		{
			name:     "Behavioral category (case-insensitive)",
			category: "BehavioRal",
			expectedParts: []string{
				DefaultSystemPrompt,
				"STAR format strictly",
				"Situation (1-2 sentences)",
			},
		},
		{
			name:     "System design category",
			category: "system_design",
			expectedParts: []string{
				DefaultSystemPrompt,
				"Clarify Requirements",
				"High-Level Architecture",
			},
		},
		{
			name:     "Conceptual category",
			category: "conceptual",
			expectedParts: []string{
				DefaultSystemPrompt,
				"Definition (2-3 sentences)",
				"How It Works",
			},
		},
		{
			name:     "Opinion category",
			category: "opinion",
			expectedParts: []string{
				DefaultSystemPrompt,
				"Position (1 sentence)",
				"Reason 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSystemPrompt(tt.category)

			if tt.exactMatch {
				if result != DefaultSystemPrompt {
					t.Errorf("Expected exactly DefaultSystemPrompt, got: %s", result)
				}
			} else {
				for _, part := range tt.expectedParts {
					if !strings.Contains(result, part) {
						t.Errorf("Expected prompt to contain %q, but it didn't. Result:\n%s", part, result)
					}
				}
			}
		})
	}
}
