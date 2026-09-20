package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGenerateScorecard_NoTurns(t *testing.T) {
	sessionData := map[string]interface{}{}
	_, err := GenerateScorecard(sessionData)
	if err == nil || !strings.Contains(err.Error(), "no valid turns found") {
		t.Errorf("Expected 'no valid turns found' error, got %v", err)
	}

	sessionData["turns"] = []interface{}{}
	_, err = GenerateScorecard(sessionData)
	if err == nil || !strings.Contains(err.Error(), "no valid turns found") {
		t.Errorf("Expected 'no valid turns found' error for empty turns array, got %v", err)
	}
}

func TestGenerateScorecard_ValidOutput(t *testing.T) {
	mockResponse := `
{
  "overall_score": 8,
  "overall_summary": "Good",
  "topic_breakdown": {"System Design": 8},
  "key_strengths": ["a"],
  "critical_gaps": ["b"],
  "recommended_topics_to_review": ["c"],
  "turns": [
    {
      "turn": 1,
      "topic_domain": "System Design",
      "interviewer_question": "Q1",
      "candidate_response": "R1",
      "score": 4,
      "verdict": "Good",
      "what_was_good": "c",
      "what_was_missing": "d",
      "suggested_improvement": "e"
    }
  ]
}
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mockChunk := map[string]interface{}{
			"id": "chatcmpl-123",
			"choices": []map[string]interface{}{
				{
					"delta": map[string]interface{}{
						"content": mockResponse,
					},
					"index": 0,
				},
			},
		}
		b, _ := json.Marshal(mockChunk)
		w.Write([]byte("data: "))
		w.Write(b)
		w.Write([]byte("\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	os.Setenv("LLM_PROVIDER", "openai")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("LLM_BASE_URL", server.URL)
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("LLM_BASE_URL")
	}()

	sessionData := map[string]interface{}{
		"turns": []interface{}{
			map[string]interface{}{
				"interviewer_question": "Hello",
				"candidate_response":   "Hi",
				"ai_suggestion":        "Greeting",
			},
		},
	}

	scorecard, err := GenerateScorecard(sessionData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if scorecard.OverallScore != 8 {
		t.Errorf("Expected 8, got %d", scorecard.OverallScore)
	}
	if scorecard.TopicBreakdown["System Design"] != 8 {
		t.Errorf("Expected TopicBreakdown 'System Design' to be 8, got %v", scorecard.TopicBreakdown["System Design"])
	}
	if len(scorecard.Turns) != 1 {
		t.Fatalf("Expected 1 turn evaluation, got %d", len(scorecard.Turns))
	}
}

func TestGenerateScorecard_MarkdownRemoval(t *testing.T) {
	mockResponse := "```json\n" + `
{
  "overall_score": 9,
  "overall_summary": "Great",
  "topic_breakdown": {},
  "key_strengths": ["a"],
  "critical_gaps": ["b"],
  "recommended_topics_to_review": [],
  "turns": []
}
` + "\n```"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mockChunk := map[string]interface{}{
			"id": "chatcmpl-123",
			"choices": []map[string]interface{}{
				{
					"delta": map[string]interface{}{
						"content": mockResponse,
					},
					"index": 0,
				},
			},
		}
		b, _ := json.Marshal(mockChunk)
		w.Write([]byte("data: "))
		w.Write(b)
		w.Write([]byte("\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	os.Setenv("LLM_PROVIDER", "openai")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("LLM_BASE_URL", server.URL)
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("LLM_BASE_URL")
	}()

	// Testing the type struct slice parsing logic branch
	type MockTurn struct {
		InterviewerQuestion string `json:"interviewer_question"`
		CandidateResponse   string `json:"candidate_response"`
		AISuggestion        string `json:"ai_suggestion"`
	}
	sessionData := map[string]interface{}{
		"turns": []MockTurn{
			{InterviewerQuestion: "Hello", CandidateResponse: "Hi", AISuggestion: "Greeting"},
		},
	}

	scorecard, err := GenerateScorecard(sessionData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if scorecard.OverallScore != 9 {
		t.Errorf("Expected 9, got %d", scorecard.OverallScore)
	}
}

func TestGenerateScorecard_Error(t *testing.T) {
	mockResponse := "invalid json"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mockChunk := map[string]interface{}{
			"id": "chatcmpl-123",
			"choices": []map[string]interface{}{
				{
					"delta": map[string]interface{}{
						"content": mockResponse,
					},
					"index": 0,
				},
			},
		}
		b, _ := json.Marshal(mockChunk)
		w.Write([]byte("data: "))
		w.Write(b)
		w.Write([]byte("\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	os.Setenv("LLM_PROVIDER", "openai")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("LLM_BASE_URL", server.URL)
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("LLM_BASE_URL")
	}()

	sessionData := map[string]interface{}{
		"turns": []interface{}{
			map[string]interface{}{
				"interviewer_question": "Hello",
				"candidate_response":   "Hi",
			},
		},
	}

	_, err := GenerateScorecard(sessionData)
	if err == nil {
		t.Fatalf("Expected error for invalid json, got nil")
	}
}

func TestGenerateScorecard_StreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	os.Setenv("LLM_PROVIDER", "openai")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("LLM_BASE_URL", server.URL)
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("LLM_BASE_URL")
	}()

	sessionData := map[string]interface{}{
		"turns": []interface{}{
			map[string]interface{}{
				"interviewer_question": "Hello",
				"candidate_response":   "Hi",
			},
		},
	}

	_, err := GenerateScorecard(sessionData)
	if err == nil || !strings.Contains(err.Error(), "failed to generate scorecard") {
		t.Fatalf("Expected error for server error, got %v", err)
	}
}

func TestGenerateScorecard_InvalidTurnType(t *testing.T) {
	sessionData := map[string]interface{}{
		"turns": []interface{}{
			"invalid_string_turn",
		},
	}
	_, err := GenerateScorecard(sessionData)
	if err == nil || !strings.Contains(err.Error(), "no valid turns found") {
		t.Errorf("Expected 'no valid turns found' error, got %v", err)
	}
}

func TestGenerateScorecard_MarshalError(t *testing.T) {
	sessionData := map[string]interface{}{
		"turns": make(chan int),
	}
	_, err := GenerateScorecard(sessionData)
	if err == nil || !strings.Contains(err.Error(), "no valid turns found") {
		t.Errorf("Expected 'no valid turns found' error, got %v", err)
	}
}

func TestGenerateScorecard_MarkdownRemoval_NoJson(t *testing.T) {
	mockResponse := "```\n" + `
{
  "overall_score": 7,
  "overall_summary": "Great",
  "topic_breakdown": {},
  "key_strengths": ["a"],
  "critical_gaps": ["b"],
  "recommended_topics_to_review": [],
  "turns": []
}
` + "\n```"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mockChunk := map[string]interface{}{
			"id": "chatcmpl-123",
			"choices": []map[string]interface{}{
				{
					"delta": map[string]interface{}{
						"content": mockResponse,
					},
					"index": 0,
				},
			},
		}
		b, _ := json.Marshal(mockChunk)
		w.Write([]byte("data: "))
		w.Write(b)
		w.Write([]byte("\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	os.Setenv("LLM_PROVIDER", "openai")
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("LLM_BASE_URL", server.URL)
	defer func() {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("LLM_BASE_URL")
	}()

	sessionData := map[string]interface{}{
		"turns": []interface{}{
			map[string]interface{}{
				"interviewer_question": "Hello",
				"candidate_response":   "Hi",
			},
		},
	}

	scorecard, err := GenerateScorecard(sessionData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if scorecard.OverallScore != 7 {
		t.Errorf("Expected 7, got %d", scorecard.OverallScore)
	}
}
