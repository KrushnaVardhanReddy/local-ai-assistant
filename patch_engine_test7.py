import sys

with open('wails-app/core/engine/engine_test.go', 'r') as f:
    content = f.read()

content += '''
func TestGetState_NoCache(t *testing.T) {
	eng := engine.New(engine.Config{}, nil, nil, nil, nil)
	state := eng.GetState()
	if state.CachedPairs != 0 {
		t.Errorf("Expected 0 cached pairs when cache is nil")
	}
}

func TestLLMError(t *testing.T) {
	llm := &MockLLMError{}
	eng := engine.New(engine.Config{}, nil, llm, nil, nil)
	eng.AskQuestion("hello this is a valid question 3")
}

type MockLLMError struct {
}
func (m *MockLLMError) StreamCompletion(q string, sp string, hist []driven.ChatMessage, onToken driven.StreamCallback, onDone func()) error {
	return fmt.Errorf("mock llm stream error")
}
func (m *MockLLMError) StreamVision(b64 string, prompt string, onToken driven.StreamCallback, onDone func()) error {
	return nil
}
'''
with open('wails-app/core/engine/engine_test.go', 'w') as f:
    f.write(content)
