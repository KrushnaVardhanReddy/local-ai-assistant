package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TurnEvaluation represents the evaluation of a single turn.
type TurnEvaluation struct {
	Turn              int    `json:"turn"`
	Score             int    `json:"score"`
	Verdict           string `json:"verdict"`
	WhatWasGood       string `json:"what_was_good"`
	WhatWasMissing    string `json:"what_was_missing"`
	SuggestedAddition string `json:"suggested_addition"`
}

// Scorecard represents the overall evaluation scorecard of an interview.
type Scorecard struct {
	OverallScore   int              `json:"overall_score"`
	OverallSummary string           `json:"overall_summary"`
	Strengths      []string         `json:"strengths"`
	Gaps           []string         `json:"gaps"`
	Turns          []TurnEvaluation `json:"turns"`
}

// GenerateScorecard evaluates the session data and returns a Scorecard.
func GenerateScorecard(sessionData map[string]interface{}) (*Scorecard, error) {
	turnsData, ok := sessionData["turns"].([]interface{})
	if !ok {
		// handle if the array is already of type session.Turn via reflection or cast in JSON if needed
		// Let's attempt to format it safely
		turnsData = nil
	}

	var transcriptBuilder strings.Builder

	// Fallback mechanism to handle both map[string]interface{} (from JSON) and strong typed slices
	if turnsData != nil {
		for i, t := range turnsData {
			turnMap, ok := t.(map[string]interface{})
			if !ok {
				continue
			}
			transcript, _ := turnMap["transcript"].(string)
			response, _ := turnMap["response"].(string)

			transcriptBuilder.WriteString(fmt.Sprintf("Turn %d:\nCandidate said: %s\nAssistant answered: %s\n\n", i+1, transcript, response))
		}
	} else if typedTurns, ok := sessionData["turns"]; ok {
		// When turns is actually a struct array (e.g., from session.Turn directly)
		// We can marshal and unmarshal to convert it into a map for simplicity
		b, err := json.Marshal(typedTurns)
		if err == nil {
			var parsedTurns []map[string]interface{}
			if err := json.Unmarshal(b, &parsedTurns); err == nil {
				for i, turnMap := range parsedTurns {
					transcript, _ := turnMap["transcript"].(string)
					response, _ := turnMap["response"].(string)
					transcriptBuilder.WriteString(fmt.Sprintf("Turn %d:\nCandidate said: %s\nAssistant answered: %s\n\n", i+1, transcript, response))
				}
			}
		}
	}

	transcript := strings.TrimSpace(transcriptBuilder.String())
	if transcript == "" {
		return nil, fmt.Errorf("no valid turns found to evaluate")
	}

	prompt := fmt.Sprintf(`You are an expert technical interview evaluator. Analyze the following interview transcript and return a JSON scorecard.
TRANSCRIPT:
%s
Return ONLY valid JSON with this exact structure (no markdown, no explanation):
{
  "overall_score": <1-10 integer>,
  "overall_summary": "<2-3 sentence overall assessment>",
  "strengths": ["<strength 1>", "<strength 2>"],
  "gaps": ["<gap 1>", "<gap 2>"],
  "turns": [
    {
      "turn": 0,
      "score": <1-5 integer>,
      "verdict": "<Good|Partial|Incomplete|Off-topic>",
      "what_was_good": "<one sentence>",
      "what_was_missing": "<one sentence or null>",
      "suggested_addition": "<one concise sentence or null>"
    }
  ]
}`, transcript)

	msgs := []ChatMessage{
		{Role: "user", Content: prompt},
	}

	var result string
	// StreamCompletionWithContext expects (question string, category string, history []ChatMessage, onToken StreamCallback, onDone func())
	// We pass empty question/category as we are putting our prompt directly in the history msgs
	err := StreamCompletionWithContext("", "scorecard", msgs, func(token string) {
		result += token
	}, func() {})

	if err != nil {
		return nil, fmt.Errorf("failed to generate scorecard: %w", err)
	}

	// Clean up potential markdown blocks
	result = strings.TrimSpace(result)
	if strings.HasPrefix(result, "```json") {
		result = strings.TrimPrefix(result, "```json")
		result = strings.TrimSuffix(result, "```")
	} else if strings.HasPrefix(result, "```") {
		result = strings.TrimPrefix(result, "```")
		result = strings.TrimSuffix(result, "```")
	}
	result = strings.TrimSpace(result)

	var scorecard Scorecard
	if err := json.Unmarshal([]byte(result), &scorecard); err != nil {
		return nil, fmt.Errorf("failed to parse JSON scorecard: %w. Result: %s", err, result)
	}

	return &scorecard, nil
}
