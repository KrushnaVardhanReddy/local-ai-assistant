package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// TurnEvaluation represents the evaluation of a single turn.
type TurnEvaluation struct {
	Turn                 int    `json:"turn"`
	TopicDomain          string `json:"topic_domain"`
	InterviewerQuestion  string `json:"interviewer_question"`
	CandidateResponse    string `json:"candidate_response"`
	Score                int    `json:"score"`
	Verdict              string `json:"verdict"`
	WhatWasGood          string `json:"what_was_good"`
	WhatWasMissing       string `json:"what_was_missing"`
	SuggestedImprovement string `json:"suggested_improvement"`
}

// Scorecard represents the overall evaluation scorecard of an interview.
type Scorecard struct {
	OverallScore              int              `json:"overall_score"`
	OverallSummary            string           `json:"overall_summary"`
	TopicBreakdown            map[string]int   `json:"topic_breakdown"`
	KeyStrengths              []string         `json:"key_strengths"`
	CriticalGaps              []string         `json:"critical_gaps"`
	RecommendedTopicsToReview []string         `json:"recommended_topics_to_review"`
	Turns                     []TurnEvaluation `json:"turns"`
}

// GenerateScorecard evaluates the session data and returns a Scorecard.
func GenerateScorecard(sessionData map[string]interface{}) (*Scorecard, error) {
	turnsData, ok := sessionData["turns"].([]interface{})
	if !ok {
		turnsData = nil
	}

	var transcriptBuilder strings.Builder

	processTurn := func(i int, turnMap map[string]interface{}) {
		question, _ := turnMap["interviewer_question"].(string)
		response, _ := turnMap["candidate_response"].(string)
		suggestion, _ := turnMap["ai_suggestion"].(string)

		transcriptBuilder.WriteString(fmt.Sprintf("Turn %d:\nInterviewer Question: %s\nCandidate Response: %s\nAI Suggestion: %s\n\n", i+1, question, response, suggestion))
	}

	if turnsData != nil {
		for i, t := range turnsData {
			turnMap, ok := t.(map[string]interface{})
			if !ok {
				continue
			}
			processTurn(i, turnMap)
		}
	} else if typedTurns, ok := sessionData["turns"]; ok {
		b, err := json.Marshal(typedTurns)
		if err == nil {
			var parsedTurns []map[string]interface{}
			if err := json.Unmarshal(b, &parsedTurns); err == nil {
				for i, turnMap := range parsedTurns {
					processTurn(i, turnMap)
				}
			}
		}
	}

	transcript := strings.TrimSpace(transcriptBuilder.String())
	if transcript == "" {
		return nil, fmt.Errorf("no valid turns found to evaluate")
	}

	prompt := fmt.Sprintf(`You are a senior principal technical interviewer and executive career coach. Analyze the following candidate interview transcript where Candidate Spoken Answers were captured during the live session.

TRANSCRIPT:
%s

EVALUATION CRITERIA:
1. Assess candidate's technical correctness, communication clarity, depth, and edge-case handling.
2. Classify each turn into a Topic Domain.
3. Compare candidate's spoken answer against industry standards and AI reference suggestions.
4. Generate actionable, constructive coaching feedback.

Return ONLY valid JSON with exact schema matching Scorecard (no markdown, no explanation):
{
  "overall_score": <1-10 integer>,
  "overall_summary": "<2-3 sentence overall assessment>",
  "topic_breakdown": {"<Topic 1>": <1-10 integer>, "<Topic 2>": <1-10 integer>},
  "key_strengths": ["<strength 1>", "<strength 2>"],
  "critical_gaps": ["<gap 1>", "<gap 2>"],
  "recommended_topics_to_review": ["<topic 1>", "<topic 2>"],
  "turns": [
    {
      "turn": <integer>,
      "topic_domain": "<Topic Domain>",
      "interviewer_question": "<question string>",
      "candidate_response": "<response string>",
      "score": <1-5 integer>,
      "verdict": "<Excellent|Good|Partial|Incomplete|Missed>",
      "what_was_good": "<one sentence>",
      "what_was_missing": "<one sentence or null>",
      "suggested_improvement": "<one concise sentence or null>"
    }
  ]
}`, transcript)

	msgs := []ChatMessage{
		{Role: "user", Content: prompt},
	}

	var result string
	err := StreamCompletionWithContext(context.Background(), "", "scorecard", msgs, func(token string) {
		result += token
	}, func() {})

	if err != nil {
		return nil, fmt.Errorf("failed to generate scorecard: %w", err)
	}

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
