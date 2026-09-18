package ports_test

import (
	"context"
	"wails-app/core/ports/driven"
	"wails-app/core/ports/driving"
)

// Ensure the driving interfaces are valid Go interfaces.
var _ driving.PipelinePort = (*mockPipeline)(nil)
var _ driving.SessionPort = (*mockSession)(nil)

// Ensure the driven interfaces are valid Go interfaces.
var _ driven.LLMPort = (*mockLLM)(nil)
var _ driven.CachePort = (*mockCache)(nil)
var _ driven.EventPort = (*mockEvent)(nil)
var _ driven.WindowPort = (*mockWindow)(nil)

// Mock implementations
type mockPipeline struct{}

func (m *mockPipeline) ProcessAudio(samples []float32) error { return nil }
func (m *mockPipeline) AskQuestion(question string) error    { return nil }
func (m *mockPipeline) SetStealth(enable bool) error         { return nil }
func (m *mockPipeline) GetState() driving.PipelineState      { return driving.PipelineState{} }
func (m *mockPipeline) ClearState()                          {}

type mockSession struct{}

func (m *mockSession) StartTurn(input string) {}
func (m *mockSession) CompleteTurn()          {}
func (m *mockSession) EndSession() (map[string]interface{}, error) {
	return map[string]interface{}{"turn_count": 0, "turns": []string{}}, nil
}

type mockLLM struct{}

func (m *mockLLM) StreamCompletion(ctx context.Context, question string, systemPrompt string, history []driven.ChatMessage, onToken driven.StreamCallback, onDone func()) error {
	return nil
}
func (m *mockLLM) StreamVision(ctx context.Context, base64Image string, prompt string, onToken driven.StreamCallback, onDone func()) error {
	return nil
}

type mockCache struct{}

func (m *mockCache) Search(embedding []float32, threshold float64) (string, bool) { return "", false }
func (m *mockCache) Store(question, answer string) error                          { return nil }
func (m *mockCache) Count() int                                                   { return 0 }

type mockEvent struct{}

func (m *mockEvent) Emit(event string, payload any) {}

type mockWindow struct{}

func (m *mockWindow) SetCaptureExcluded(ctx context.Context, excluded bool) error { return nil }
func (m *mockWindow) HideFromTaskbar(ctx context.Context) error                   { return nil }
