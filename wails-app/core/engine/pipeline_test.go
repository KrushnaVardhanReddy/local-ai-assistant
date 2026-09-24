package engine

import (
	"context"
	"testing"
	"time"

	"wails-app/core/ports/driven"
)

type mockPreemptLLM struct {
	tokens  []string
	blockCh chan struct{}
	calls   []string
}

func (m *mockPreemptLLM) StreamCompletion(ctx context.Context, q string, sp string, hist []driven.ChatMessage, onToken driven.StreamCallback, onDone func()) error {
	m.calls = append(m.calls, q)

	for _, t := range m.tokens {
		select {
		case <-ctx.Done():
			if onDone != nil {
				onDone() // Some stream processors call onDone on cancellation
			}
			return ctx.Err()
		case <-m.blockCh:
			onToken(t)
		}
	}
	if onDone != nil {
		onDone()
	}
	return nil
}

func (m *mockPreemptLLM) StreamVision(ctx context.Context, b64 string, prompt string, onToken driven.StreamCallback, onDone func()) error {
	return nil
}

func TestPipeline_PreemptionOnInterruption(t *testing.T) {
	blockCh := make(chan struct{})
	llm := &mockPreemptLLM{
		tokens:  []string{"A", "B", "C"},
		blockCh: blockCh,
	}
	events := &MockEvents{Emitted: make(map[string]int)}
	eng := New(Config{}, nil, llm, nil, events)

	// Send Q1
	eng.AskQuestion("How do you pass data between the go routines?")
	time.Sleep(50 * time.Millisecond) // Give time to start stream

	// Send Q2 to preempt Q1
	eng.AskQuestion("What is the best way to handle concurrent operations?")

	// Close blockCh so the preempting question can finish its stream.
	// The original question stream should return immediately on context cancellation.
	close(blockCh)
	time.Sleep(100 * time.Millisecond) // Let both finish

	if len(llm.calls) != 2 {
		t.Errorf("Expected 2 calls to LLM, got %d", len(llm.calls))
	}

	if events.Emitted["on_chip"] < 2 {
		t.Errorf("Expected on_chip emitted at least twice, got %d", events.Emitted["on_chip"])
	}
}

func TestPipeline_ClearStateCancelsInFlight(t *testing.T) {
	blockCh := make(chan struct{})
	llm := &mockPreemptLLM{
		tokens:  []string{"A", "B"},
		blockCh: blockCh,
	}
	eng := New(Config{}, nil, llm, nil, nil)

	eng.AskQuestion("How do you handle context cancellation properly in Go?")
	time.Sleep(50 * time.Millisecond) // Stream is blocking on blockCh

	// ClearState should cancel the inflight context, causing the LLM loop to abort
	// without needing us to unblock the channel.
	eng.ClearState()
	time.Sleep(50 * time.Millisecond) // Let the goroutine exit

	eng.inFlightMu.Lock()
	inFlightCtx := eng.inFlightCtx
	eng.inFlightMu.Unlock()

	if inFlightCtx != nil {
		t.Errorf("Expected inFlightCtx to be nil after clear state context cancellation and goroutine return")
	}
}

func TestPipeline_CacheBypassWhenActive(t *testing.T) {
	cache := &MockCache{Hit: true, Answer: "Cached"}
	llm := &mockPreemptLLM{}
	eng := New(Config{}, nil, llm, cache, nil)

	eng.AskQuestion("What is the difference between a mutex and a channel?")
	time.Sleep(50 * time.Millisecond)

	state := eng.GetState()
	if state.Response != "Cached" {
		t.Errorf("Expected cached answer, got %s", state.Response)
	}

	eng.inFlightMu.Lock()
	if eng.inFlightCtx != nil {
		t.Errorf("Expected inFlightCtx to be nil after cache hit")
	}
	eng.inFlightMu.Unlock()
}

// Dummy MockEvents for testing pipeline preemption
type MockEvents struct {
	Emitted map[string]int
}

func (m *MockEvents) Emit(eventName string, data any) {
	if m.Emitted == nil {
		m.Emitted = make(map[string]int)
	}
	m.Emitted[eventName]++
}
func (m *MockEvents) Subscribe(eventName string, callback func(...interface{})) {}
func (m *MockEvents) Unsubscribe(eventName string)                              {}

// Dummy MockCache for testing cache bypass
type MockCache struct {
	Hit    bool
	Answer string
}

func (m *MockCache) Search(embedding []float32, threshold float64) (string, bool) {
	if m.Hit {
		return m.Answer, true
	}
	return "", false
}
func (m *MockCache) Store(question, answer string) error { return nil }
func (m *MockCache) Count() int                          { return 1 }

func TestPipeline_ManualMode(t *testing.T) {
	events := &MockEvents{Emitted: make(map[string]int)}
	eng := New(Config{}, nil, nil, nil, events)
	eng.SetManualMode(true)

	// In manual mode, AskQuestion should bypass LLM call and return immediately if isAuto is true
	eng.handleTranscript("Hello manually", true)

	// And if isAuto is false, it should continue normally
	// eng.AskQuestion calls handleTranscript with false, so it won't hit the bypass early return
	// but we just pass nil llm to fail naturally if it proceeds, so we know if it bypassed or not.
	// We'll just verify no panic happens when bypassed
}

func TestPipeline_RollingTranscriptBuffer(t *testing.T) {
	blockCh := make(chan struct{})
	close(blockCh) // unblocked immediately
	llm := &mockPreemptLLM{
		tokens:  []string{"A"},
		blockCh: blockCh,
	}
	events := &MockEvents{Emitted: make(map[string]int)}
	eng := New(Config{}, nil, llm, nil, events)

	// Pre-populate buffer with 3 strings
	eng.mu.Lock()
	eng.transcriptBuffer = []string{
		"Tell me about",
		"your experience with",
		"distributed systems",
	}
	eng.mu.Unlock()

	// Send a 4th transcript
	eng.handleTranscript("at scale.", false)
	time.Sleep(50 * time.Millisecond) // Give time for goroutine to start and call LLM

	if len(llm.calls) != 1 {
		t.Fatalf("Expected 1 call to LLM, got %d", len(llm.calls))
	}

	capturedQuestion := llm.calls[0]

	// Check if all entries are correctly present
	if !strings.Contains(capturedQuestion, "[1] \"Tell me about\"") {
		t.Errorf("Missing entry 1. Got: %s", capturedQuestion)
	}
	if !strings.Contains(capturedQuestion, "[2] \"your experience with\"") {
		t.Errorf("Missing entry 2. Got: %s", capturedQuestion)
	}
	if !strings.Contains(capturedQuestion, "[3] \"distributed systems\"") {
		t.Errorf("Missing entry 3. Got: %s", capturedQuestion)
	}
	if !strings.Contains(capturedQuestion, "[4] \"at scale.\"") {
		t.Errorf("Missing entry 4. Got: %s", capturedQuestion)
	}
}

func TestPipeline_MockMode(t *testing.T) {
	events := &MockEvents{Emitted: make(map[string]int)}
	eng := New(Config{}, nil, nil, nil, events, nil)
	eng.ToggleMockInterviewMode(true)

	// In mock mode, handleTranscript with isAuto=true should add to buffer but bypass auto-submit
	eng.handleTranscript("Hello mock mode", true)

	if len(eng.GetQuestionBuffer().GetChunks()) == 0 {
		t.Fatalf("Expected transcript to be added to buffer in mock mode")
	}
}
