package llmadapter

import "wails-app/core/ports/driven"

// MockLLMAdapter is a test double that replays scripted responses.
// Use in unit tests to avoid real API calls.
type MockLLMAdapter struct {
	// Tokens is the list of token strings to emit via onToken.
	Tokens []string
	// Err is returned from StreamCompletion if non-nil.
	Err error
	// Calls records each question passed to StreamCompletion for assertion.
	Calls []string
}

// StreamCompletion records the call and replays Tokens or returns Err.
func (m *MockLLMAdapter) StreamCompletion(
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
		onToken(tok)
	}
	onDone()
	return nil
}

// StreamVision delegates to StreamCompletion for recording/replaying.
func (m *MockLLMAdapter) StreamVision(base64Image, prompt string, onToken driven.StreamCallback, onDone func()) error {
	return m.StreamCompletion(prompt, "", nil, onToken, onDone)
}

// Compile-time interface check.
var _ driven.LLMPort = (*MockLLMAdapter)(nil)
