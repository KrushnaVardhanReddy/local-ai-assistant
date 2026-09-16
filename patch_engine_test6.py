import sys

with open('wails-app/core/engine/engine_test.go', 'r') as f:
    content = f.read()

# Fix TestHandleTranscript_LLMBusy empty stub
content = content.replace('''func TestHandleTranscript_LLMBusy(t *testing.T) {
	eng := engine.New(engine.Config{}, nil, nil, nil, nil)

	// manually lock so it's busy
	// Wait, we can't lock e.llmBusy easily without calling a function that doesn't return immediately
	// But we can trigger AskQuestion twice!

	// Create a mock LLM that blocks until we tell it to finish
	blockCh := make(chan struct{})
	llm := &MockLLM{StreamCalled: false} // We will just overwrite methods if needed, or use a custom one.
}''', '')

with open('wails-app/core/engine/engine_test.go', 'w') as f:
    f.write(content)
