import sys

with open('wails-app/core/engine/engine_test.go', 'r') as f:
    content = f.read()

content += '''
func TestHandleTranscript_LLMBusy(t *testing.T) {
	eng := engine.New(engine.Config{}, nil, nil, nil, nil)

	// manually lock so it's busy
	// Wait, we can't lock e.llmBusy easily without calling a function that doesn't return immediately
	// But we can trigger AskQuestion twice!

	// Create a mock LLM that blocks until we tell it to finish
	blockCh := make(chan struct{})
	llm := &MockLLM{StreamCalled: false} // We will just overwrite methods if needed, or use a custom one.
}

type MockLLMBlock struct {
	BlockCh chan struct{}
}
func (m *MockLLMBlock) StreamCompletion(q string, sp string, hist []driven.ChatMessage, onToken driven.StreamCallback, onDone func()) error {
	<-m.BlockCh
	onDone()
	return nil
}
func (m *MockLLMBlock) StreamVision(b64 string, prompt string, onToken driven.StreamCallback, onDone func()) error {
	return nil
}

func TestLLMBusy(t *testing.T) {
	blockCh := make(chan struct{})
	llm := &MockLLMBlock{BlockCh: blockCh}
	eng := engine.New(engine.Config{}, nil, llm, nil, nil)

	// First request blocks
	eng.AskQuestion("hello this is a valid question 1")

	// Second request should be discarded
	eng.AskQuestion("hello this is a valid question 2")

	close(blockCh)
}

func TestStoreError(t *testing.T) {
    // We need cache.Store to return an error to hit that branch.
}

type MockCacheError struct {
}
func (m *MockCacheError) Search(embedding []float32, threshold float64) (string, bool) { return "", false }
func (m *MockCacheError) Store(question, answer string) error { return fmt.Errorf("mock store error") }
func (m *MockCacheError) Count() int { return 0 }

func TestStoreErrorBranch(t *testing.T) {
	llm := &MockLLM{}
	cache := &MockCacheError{}
	eng := engine.New(engine.Config{}, nil, llm, cache, nil)
	eng.AskQuestion("hello this is a valid question for error branch")
}
'''
with open('wails-app/core/engine/engine_test.go', 'w') as f:
    f.write(content)
