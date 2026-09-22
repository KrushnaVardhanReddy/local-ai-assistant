package llm

import (
	"strings"
	"testing"
	"wails-app/backend/session"
)

func TestBuildContextBlock_Empty(t *testing.T) {
	result := BuildContextBlock([]session.Turn{})
	if result != "" {
		t.Errorf("Expected empty string, got: %q", result)
	}
}

func TestBuildContextBlock_SingleTurn(t *testing.T) {
	turns := []session.Turn{
		{
			InterviewerQuestion: "Hello",
			AISuggestion:        "World",
		},
	}
	result := BuildContextBlock(turns)

	if !strings.Contains(result, "[Turn 1]") {
		t.Errorf("Expected result to contain '[Turn 1]', got: %s", result)
	}
	if !strings.Contains(result, "Hello") {
		t.Errorf("Expected result to contain 'Hello', got: %s", result)
	}
	if !strings.Contains(result, "World") {
		t.Errorf("Expected result to contain 'World', got: %s", result)
	}
	if !strings.HasPrefix(result, "=== Recent Interview Context") {
		t.Errorf("Expected result to start with '=== Recent Interview Context', got: %s", result)
	}
	if !strings.HasSuffix(result, "=== End Context ===") {
		t.Errorf("Expected result to end with '=== End Context ===', got: %s", result)
	}
}

func TestBuildContextBlock_ThreeTurns(t *testing.T) {
	turns := []session.Turn{
		{InterviewerQuestion: "Q1", AISuggestion: "A1"},
		{InterviewerQuestion: "Q2", AISuggestion: "A2"},
		{InterviewerQuestion: "Q3", AISuggestion: "A3"},
	}
	result := BuildContextBlock(turns)
	if !strings.Contains(result, "[Turn 1]") || !strings.Contains(result, "[Turn 2]") || !strings.Contains(result, "[Turn 3]") {
		t.Errorf("Expected result to contain all 3 turns, got: %s", result)
	}
}

func TestBuildContextBlock_Truncation(t *testing.T) {
	longQ := strings.Repeat("A", 200)
	turns := []session.Turn{
		{InterviewerQuestion: longQ, AISuggestion: "Ans"},
	}
	result := BuildContextBlock(turns)
	if !strings.Contains(result, strings.Repeat("A", 120)+"...") {
		t.Errorf("Expected question to be truncated to 120 chars + '...'. Result: %s", result)
	}
}

func TestBuildContextBlock_OverallSizeLimit(t *testing.T) {
	turns := make([]session.Turn, 10)
	for i := 0; i < 10; i++ {
		turns[i] = session.Turn{
			InterviewerQuestion: strings.Repeat("Q", 200),
			AISuggestion:        strings.Repeat("A", 300),
		}
	}
	result := BuildContextBlock(turns)
	if len(result) > 2000 {
		t.Errorf("Expected result to be <= 2000 chars, got length %d", len(result))
	}
}

func TestBuildFullSystemPrompt_WithContext(t *testing.T) {
	turns := []session.Turn{
		{InterviewerQuestion: "Q1", AISuggestion: "A1"},
		{InterviewerQuestion: "Q2", AISuggestion: "A2"},
	}
	result := BuildFullSystemPrompt("behavioral", turns)

	if !strings.Contains(result, DefaultSystemPrompt) {
		t.Errorf("Expected default prompt to be included")
	}
	if !strings.Contains(result, "=== Recent Interview Context") {
		t.Errorf("Expected context block to be included")
	}
	if !strings.Contains(result, "Q1") || !strings.Contains(result, "Q2") {
		t.Errorf("Expected all questions to be present")
	}
}

func TestBuildFullSystemPrompt_EmptyTurns(t *testing.T) {
	turns := []session.Turn{}
	resultWithEmptyContext := BuildFullSystemPrompt("coding", turns)
	resultWithoutContext := BuildSystemPrompt("coding")

	if resultWithEmptyContext != resultWithoutContext {
		t.Errorf("Expected empty turns to fall back identically to BuildSystemPrompt")
	}
}

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
