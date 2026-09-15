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
			Content json.RawMessage `json:"content"`
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

type ImageURLContent struct {
	URL string `json:"url"`
}

type VisionMessageContent struct {
	Type     string           `json:"type"`
	Text     string           `json:"text,omitempty"`
	ImageURL *ImageURLContent `json:"image_url,omitempty"`
}

type VisionChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type VisionChatRequest struct {
	Model    string              `json:"model"`
	Messages []VisionChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
}

func StreamVisionCompletion(base64Image string, prompt string, onToken StreamCallback, onDone func()) error {
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

	if prompt == "" {
		prompt = DefaultVisionPrompt
	}

	reqBody := VisionChatRequest{
		Model: model,
		Messages: []VisionChatMessage{
			{
				Role:    "system",
				Content: "You are an expert AI vision coding assistant. Analyze screenshots carefully and answer concisely.",
			},
			{
				Role: "user",
				Content: []VisionMessageContent{
					{
						Type: "text",
						Text: prompt,
					},
					{
						Type: "image_url",
						ImageURL: &ImageURLContent{
							URL: base64Image,
						},
					},
				},
			},
		},
		Stream: true,
	}

	bodyBytes, err := jsonMarshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal vision request: %w", err)
	}

	req, err := httpNewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create vision request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := httpClientDo(client, req)
	if err != nil {
		return fmt.Errorf("failed to send vision request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Vision API request failed with status: %s", resp.Status)
	}

	log.Printf("[LLM-Vision] Streaming response for screenshot...")

	scanner := bufio.NewScanner(resp.Body)
	tokenCount := 0
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
				raw := chunk.Choices[0].Delta.Content
				var content string
				if len(raw) >= 2 && raw[0] == '"' {
					if err := json.Unmarshal(raw, &content); err == nil && content != "" && onToken != nil {
						tokenCount++
						onToken(content)
					}
				}
			}
		}
	}

	log.Printf("[LLM-Vision] Stream complete — %d tokens emitted", tokenCount)

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading vision stream: %w", err)
	}

	return nil
}

func StreamCompletionWithContext(question string, category string, history []ChatMessage, onToken StreamCallback, onDone func()) error {
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

	messages := make([]ChatMessage, 0, len(history)+2)
	messages = append(messages, ChatMessage{Role: "system", Content: BuildSystemPrompt(category)})
	messages = append(messages, history...)
	messages = append(messages, ChatMessage{Role: "user", Content: question})

	reqBody := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
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
	tokenCount := 0
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
				raw := chunk.Choices[0].Delta.Content
				// Only process if it's a JSON string (not a number/token-ID)
				var content string
				if len(raw) >= 2 && raw[0] == '"' {
					if err := json.Unmarshal(raw, &content); err == nil && content != "" && onToken != nil {
						tokenCount++
						onToken(content)
					}
				}
			}
		}
	}

	log.Printf("[LLM] Stream complete — %d tokens emitted", tokenCount)

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

func StreamCompletion(question string, onToken StreamCallback, onDone func()) error {
	return StreamCompletionWithContext(question, "", nil, onToken, onDone)
}
