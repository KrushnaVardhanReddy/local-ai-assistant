package llmadapter

import (
	"wails-app/backend/llm"
	"wails-app/core/ports/driven"
)

// OpenAIAdapter wraps the existing backend/llm package functions
// behind the driven.LLMPort interface.
type OpenAIAdapter struct{}

// NewOpenAIAdapter creates a new OpenAIAdapter.
func NewOpenAIAdapter() *OpenAIAdapter {
	return &OpenAIAdapter{}
}

// StreamCompletion delegates to llm.StreamCompletionWithContext.
// The systemPrompt is passed as the category parameter to BuildSystemPrompt.
// NOTE: history []driven.ChatMessage must be converted to []llm.ChatMessage.
func (a *OpenAIAdapter) StreamCompletion(
	question string,
	systemPrompt string,
	history []driven.ChatMessage,
	onToken driven.StreamCallback,
	onDone func(),
) error {
	// Convert driven.ChatMessage slice to llm.ChatMessage slice.
	msgs := make([]llm.ChatMessage, len(history))
	for i, m := range history {
		msgs[i] = llm.ChatMessage{Role: m.Role, Content: m.Content}
	}
	// Use raw system prompt as the "category" override.
	// llm.BuildSystemPrompt returns DefaultSystemPrompt if category unknown.
	return llm.StreamCompletionWithContext(question, systemPrompt, msgs, llm.StreamCallback(onToken), onDone)
}

// StreamVision delegates to llm.StreamVisionCompletion.
func (a *OpenAIAdapter) StreamVision(
	base64Image string,
	prompt string,
	onToken driven.StreamCallback,
	onDone func(),
) error {
	return llm.StreamVisionCompletion(base64Image, prompt, llm.StreamCallback(onToken), onDone)
}

// Compile-time interface check.
var _ driven.LLMPort = (*OpenAIAdapter)(nil)
