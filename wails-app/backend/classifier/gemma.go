package classifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const booleanGrammar = `
root   ::= object
object ::= "{" ws "\"result\"" ws ":" ws value ws "}"
value  ::= "true" | "false"
ws     ::= [ \t\n]*
`

type GemmaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GemmaRequest struct {
	Model       string         `json:"model"`
	Messages    []GemmaMessage `json:"messages"`
	Temperature float32        `json:"temperature"`
	MaxTokens   int            `json:"max_tokens"`
	Grammar     string         `json:"grammar"`
}

type GemmaResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type GemmaResult struct {
	Result bool `json:"result"`
}

func IsQuestionComplete(ctx context.Context, transcriptBuffer string) (bool, error) {
	reqBody := GemmaRequest{
		Model: "gemma",
		Messages: []GemmaMessage{
			{
				Role:    "system",
				Content: "You are a turn-detection classifier. Respond ONLY with valid JSON.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Interview transcript so far:\n\n%s\n\nHas the interviewer finished asking their complete question and is now waiting for the candidate to respond? Answer with {\"result\": true} if yes, {\"result\": false} if no.", transcriptBuffer),
			},
		},
		Temperature: 0,
		MaxTokens:   10,
		Grammar:     booleanGrammar,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/v1/chat/completions", LlamaServerPort)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Server might not be running or error, do not panic
		return false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("llama-server returned status %d", resp.StatusCode)
	}

	var gemmaResp GemmaResponse
	if err := json.NewDecoder(resp.Body).Decode(&gemmaResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(gemmaResp.Choices) == 0 {
		return false, fmt.Errorf("no choices returned")
	}

	content := gemmaResp.Choices[0].Message.Content
	var res GemmaResult
	if err := json.Unmarshal([]byte(content), &res); err != nil {
		return false, fmt.Errorf("failed to parse JSON from content '%s': %w", content, err)
	}

	return res.Result, nil
}
