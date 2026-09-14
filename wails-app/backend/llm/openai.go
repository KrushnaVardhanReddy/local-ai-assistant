package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type StreamCallback func(token string)

func getEnvOrDefault(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatResponseChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// Variables for mocking in tests
var (
	jsonMarshal   = json.Marshal
	httpNewRequest = http.NewRequest
	httpClientDo  = func(c *http.Client, req *http.Request) (*http.Response, error) {
		return c.Do(req)
	}
)

func StreamCompletion(question string, onToken StreamCallback, onDone func()) error {
	defer func() {
		if onDone != nil {
			onDone()
		}
	}()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}
	baseURL := getEnvOrDefault("LLM_BASE_URL", "https://api.openai.com/v1")
	model := getEnvOrDefault("LLM_MODEL", "gpt-4o")

	if !strings.HasSuffix(baseURL, "/") {
	    baseURL += "/"
	}
	endpoint := baseURL + "chat/completions"

	reqBody := ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are an expert technical interview assistant. Answer the following interview question concisely, clearly, and with code examples where appropriate.",
			},
			{
				Role:    "user",
				Content: question,
			},
		},
		Stream: true,
	}

	bodyBytes, err := jsonMarshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := httpNewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := httpClientDo(client, req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	log.Printf("[LLM] Streaming response for: %q", question)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk ChatResponseChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" && onToken != nil {
					onToken(content)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}
