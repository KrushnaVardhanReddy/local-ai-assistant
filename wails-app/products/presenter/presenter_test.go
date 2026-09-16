package presenter

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	cacheadapter "wails-app/adapters/cache"
	eventsadapter "wails-app/adapters/events"
	llmadapter "wails-app/adapters/llm"
	"wails-app/backend/stt"
	"wails-app/core/engine"
)

// MockAudioCapture implements driven.AudioCapturePort for testing.
type MockAudioCapture struct {
	initErr        error
	startErr       error
	callbackCalled bool
}

func (m *MockAudioCapture) Initialize() error {
	return m.initErr
}

func (m *MockAudioCapture) StartCapture(deviceIndex int, includeLoopback bool, callback func([]float32)) error {
	if m.startErr != nil {
		return m.startErr
	}
	// Synchronously execute callback
	callback([]float32{0.0, 0.0})
	m.callbackCalled = true
	return nil
}

func (m *MockAudioCapture) Stop() {}

// MockWindowPort implements driven.WindowPort for testing.
type MockWindowPort struct{}

func (m *MockWindowPort) SetCaptureExcluded(ctx context.Context, excluded bool) error { return nil }
func (m *MockWindowPort) HideFromTaskbar(ctx context.Context) error                   { return nil }

func TestPrompts(t *testing.T) {
	if !strings.Contains(SystemPrompt, "invisible") {
		t.Error("SystemPrompt missing 'invisible'")
	}
	if !strings.Contains(SystemPrompt, "HUD") {
		t.Error("SystemPrompt missing 'HUD'")
	}
	if !strings.Contains(VisionPrompt, "slide") {
		t.Error("VisionPrompt missing 'slide'")
	}
	if !strings.Contains(VisionPrompt, "talking points") {
		t.Error("VisionPrompt missing 'talking points'")
	}
}

func TestNewPresenterApp(t *testing.T) {
	app := NewPresenterApp()
	if app == nil {
		t.Fatal("NewPresenterApp() returned nil")
	}
	if app.engine == nil {
		t.Fatal("PresenterApp.engine is nil")
	}
	if app.audioCapture == nil {
		t.Fatal("PresenterApp.audioCapture is nil")
	}
}

func TestNewPresenterAppWithGroqKey(t *testing.T) {
	t.Setenv("GROQ_API_KEY", "test-key")
	app := NewPresenterApp()
	if app == nil {
		t.Fatal("NewPresenterApp() returned nil")
	}
	// We expect Groq engine to be used, but we just want coverage for the branch
}

func TestPresenterAppMethods(t *testing.T) {
	// Setup app with mocked dependencies
	sttManager := stt.NewSTTManager(nil)
	llmAdapter := llmadapter.NewOpenAIAdapter()
	cacheAdapter := &cacheadapter.SQLiteVecAdapter{} // Just need a stub

	eng := engine.New(
		engine.Config{SystemPrompt: SystemPrompt},
		sttManager,
		llmAdapter,
		cacheAdapter,
		&eventsadapter.NoopEventAdapter{},
	)

	mockAudio := &MockAudioCapture{}
	mockWindow := &MockWindowPort{}
	app := NewPresenterAppWithPorts(eng, mockAudio, mockWindow)

	// Test GetState
	state := app.GetState()
	if _, ok := state["transcript"]; !ok {
		t.Error("GetState() missing 'transcript' key")
	}

	// Test ClearState
	app.ClearState()

	// Test AskQuestion
	// Should not crash, just pass through
	_ = app.AskQuestion("test question")
}

func TestPresenterAppStartup_Success(t *testing.T) {
	sttManager := stt.NewSTTManager(nil)
	llmAdapter := llmadapter.NewOpenAIAdapter()
	cacheAdapter := &cacheadapter.SQLiteVecAdapter{}

	eng := engine.New(
		engine.Config{SystemPrompt: SystemPrompt},
		sttManager,
		llmAdapter,
		cacheAdapter,
		&eventsadapter.NoopEventAdapter{},
	)

	mockAudio := &MockAudioCapture{}
	mockWindow := &MockWindowPort{}
	app := NewPresenterAppWithPorts(eng, mockAudio, mockWindow)

	app.startup(context.Background())

	// Give the goroutine time to run the callback
	time.Sleep(10 * time.Millisecond)

	if !mockAudio.callbackCalled {
		t.Error("StartCapture callback was not called")
	}
}

func TestPresenterAppStartup_NilContext(t *testing.T) {
	sttManager := stt.NewSTTManager(nil)
	llmAdapter := llmadapter.NewOpenAIAdapter()
	cacheAdapter := &cacheadapter.SQLiteVecAdapter{}

	eng := engine.New(
		engine.Config{SystemPrompt: SystemPrompt},
		sttManager,
		llmAdapter,
		cacheAdapter,
		&eventsadapter.NoopEventAdapter{},
	)

	mockAudio := &MockAudioCapture{}
	mockWindow := &MockWindowPort{}
	app := NewPresenterAppWithPorts(eng, mockAudio, mockWindow)

	app.startup(nil)
}


func TestPresenterAppStartup_InitError(t *testing.T) {
	sttManager := stt.NewSTTManager(nil)
	llmAdapter := llmadapter.NewOpenAIAdapter()
	cacheAdapter := &cacheadapter.SQLiteVecAdapter{}

	eng := engine.New(
		engine.Config{SystemPrompt: SystemPrompt},
		sttManager,
		llmAdapter,
		cacheAdapter,
		&eventsadapter.NoopEventAdapter{},
	)

	mockAudio := &MockAudioCapture{initErr: errors.New("init error")}
	mockWindow := &MockWindowPort{}
	app := NewPresenterAppWithPorts(eng, mockAudio, mockWindow)

	// Should not panic or block
	app.startup(context.Background())
}
