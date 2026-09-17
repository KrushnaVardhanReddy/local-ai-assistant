package llmadapter

import (
	"context"
	"wails-app/core/ports/driven"
)

// MockLLMAdapter is a test double that replays scripted responses.
// Use in unit tests to avoid real API calls.
type MockLLMAdapter struct {
	Tokens []string
	Err    error
	Calls  []string
}

func (m *MockLLMAdapter) StreamCompletion(
	ctx context.Context,
	question, systemPrompt string,
	history []driven.ChatMessage,
	onToken driven.StreamCallback,
	onDone func(),
) error {
	m.Calls = append(m.Calls, question)
	if m.Err != nil {
		return m.Err
	}
	for _, tok := range m.Tokens {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			onToken(tok)
		}
	}
	onDone()
	return nil
}

func (m *MockLLMAdapter) StreamVision(ctx context.Context, base64Image, prompt string, onToken driven.StreamCallback, onDone func()) error {
	return m.StreamCompletion(ctx, prompt, "", nil, onToken, onDone)
}

var _ driven.LLMPort = (*MockLLMAdapter)(nil)
