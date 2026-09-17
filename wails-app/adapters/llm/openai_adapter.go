package llmadapter

import (
	"context"
	"wails-app/backend/llm"
	"wails-app/core/ports/driven"
)

type OpenAIAdapter struct{}

func NewOpenAIAdapter() *OpenAIAdapter {
	return &OpenAIAdapter{}
}

func (a *OpenAIAdapter) StreamCompletion(
	ctx context.Context,
	question string,
	systemPrompt string,
	history []driven.ChatMessage,
	onToken driven.StreamCallback,
	onDone func(),
) error {
	msgs := make([]llm.ChatMessage, len(history))
	for i, m := range history {
		msgs[i] = llm.ChatMessage{Role: m.Role, Content: m.Content}
	}
	return llm.StreamCompletionWithContext(ctx, question, systemPrompt, msgs, llm.StreamCallback(onToken), onDone)
}

func (a *OpenAIAdapter) StreamVision(
	ctx context.Context,
	base64Image string,
	prompt string,
	onToken driven.StreamCallback,
	onDone func(),
) error {
	return llm.StreamVisionCompletion(ctx, base64Image, prompt, llm.StreamCallback(onToken), onDone)
}

var _ driven.LLMPort = (*OpenAIAdapter)(nil)
