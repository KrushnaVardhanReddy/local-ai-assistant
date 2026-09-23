package llm

import (
	"sync"

	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"wails-app/backend/config"
)

type StreamCallback func(token string)

// pkg-level config store — set once at startup via SetConfig.
var (
	pkgConfig     *config.AppConfig
	pkgConfigOnce sync.Once
	pkgConfigMu   sync.RWMutex
)

// SetConfig stores the application config for use by all LLM functions.
// Must be called once during app initialisation before any LLM calls.
func SetConfig(cfg *config.AppConfig) {
	pkgConfigMu.Lock()
	defer pkgConfigMu.Unlock()
	pkgConfig = cfg
}

func getCfg() *config.AppConfig {
	pkgConfigMu.RLock()
	defer pkgConfigMu.RUnlock()
	return pkgConfig
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
	jsonMarshal               = json.Marshal
	httpNewRequestWithContext = http.NewRequestWithContext
	httpClientDo              = func(c *http.Client, req *http.Request) (*http.Response, error) {
		return c.Do(req)
	}
)

var (
	DemoProxyToken  string
	proxyTokenMutex sync.RWMutex
)

func SetProxyToken(token string) {
	proxyTokenMutex.Lock()
	defer proxyTokenMutex.Unlock()
	DemoProxyToken = token
}

func GetProxyToken() string {
	proxyTokenMutex.RLock()
	defer proxyTokenMutex.RUnlock()
	return DemoProxyToken
}

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

func getProviderConfig() (baseURL string, apiKey string, err error) {
	if proxyToken := GetProxyToken(); proxyToken != "" {
		url := "https://api.barnowl.ai/v1/functions/llm-proxy"
		if cfg := getCfg(); cfg != nil && cfg.SupabaseEdgeURL != "" {
			url = cfg.SupabaseEdgeURL
		}
		return url, proxyToken, nil
	}

	cfg := getCfg()
	if cfg != nil && cfg.LLMProvider == "groq" {
		if cfg.GroqAPIKey == "" {
			return "", "", fmt.Errorf("GROQ_API_KEY environment variable is not set")
		}
		baseURL := "https://api.groq.com/openai/v1"
		if cfg.LLMBaseURL != "" && cfg.LLMBaseURL != "https://api.openai.com/v1" {
			baseURL = cfg.LLMBaseURL
		}
		return baseURL, cfg.GroqAPIKey, nil
	}

	if cfg != nil {
		if cfg.OpenAIAPIKey == "" {
			return "", "", fmt.Errorf("OPENAI_API_KEY environment variable is not set")
		}
		return cfg.LLMBaseURL, cfg.OpenAIAPIKey, nil
	}
	return "", "", fmt.Errorf("LLM config not initialised — call llm.SetConfig at startup")
}

func StreamVisionCompletion(ctx context.Context, base64Image string, prompt string, onToken StreamCallback, onDone func()) error {
	defer func() {
		if onDone != nil {
			onDone()
		}
	}()

	baseURL, apiKey, err := getProviderConfig()
	if err != nil {
		return err
	}
	model := "gpt-4o"
	if cfg := getCfg(); cfg != nil && cfg.LLMModel != "" {
		model = cfg.LLMModel
	}

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

	req, err := httpNewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(bodyBytes))
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

func StreamCompletionWithContext(ctx context.Context, question string, systemPrompt string, history []ChatMessage, onToken StreamCallback, onDone func()) error {
	defer func() {
		if onDone != nil {
			onDone()
		}
	}()

	baseURL, apiKey, err := getProviderConfig()
	if err != nil {
		return err
	}
	model := "gpt-4o"
	if cfg := getCfg(); cfg != nil && cfg.LLMModel != "" {
		model = cfg.LLMModel
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	endpoint := baseURL + "chat/completions"

	messages := make([]ChatMessage, 0, len(history)+2)
	if systemPrompt == "" {
		systemPrompt = DefaultSystemPrompt
	}
	messages = append(messages, ChatMessage{Role: "system", Content: systemPrompt})
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

	req, err := httpNewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	log.Printf("[LLM] Making request to endpoint: %s", endpoint)

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

func StreamCompletion(ctx context.Context, question string, onToken StreamCallback, onDone func()) error {
	return StreamCompletionWithContext(ctx, question, BuildSystemPrompt(""), nil, onToken, onDone)
}
