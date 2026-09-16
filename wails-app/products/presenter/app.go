package presenter

import (
	"context"
	"os"

	cacheadapter "wails-app/adapters/cache"
	eventsadapter "wails-app/adapters/events"
	llmadapter "wails-app/adapters/llm"
	windowadapter "wails-app/adapters/window"
	"wails-app/backend"
	"wails-app/backend/audio"
	"wails-app/backend/stt"
	"wails-app/core/engine"
	"wails-app/core/ports/driven"
	"wails-app/backend/parser"
)

// PresenterApp is the Wails App struct for the StealthPresenter product.
// It is thin — it owns only Wails-facing concerns (context, window, audio).
// All pipeline logic is delegated to the embedded StealthEngine.
type PresenterApp struct {
	ctx          context.Context
	engine       *engine.StealthEngine
	audioCapture driven.AudioCapturePort
	window       driven.WindowPort
}

func NewPresenterApp() *PresenterApp {
	// Build adapters
	db, _ := backend.NewVectorDB("./data/presenter_cache.db")
	llmAdapter := llmadapter.NewOpenAIAdapter()
	cacheAdapter := cacheadapter.NewSQLiteVecAdapter(db)

	// STT: prefer Groq if key set, else fall back to local Whisper
	var sttEngine stt.STTEngine
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		sttEngine = stt.NewGroqEngine(key, "whisper-large-v3-turbo")
	} else {
		validPath, err := stt.EnsureWhisperModel("")
		if err == nil {
			sttEngine, _ = stt.LoadWhisperEngine(validPath)
		}
	}

	eng := engine.New(
		engine.Config{SystemPrompt: SystemPrompt},
		stt.NewSTTManager(sttEngine),
		llmAdapter,
		cacheAdapter,
		&eventsadapter.NoopEventAdapter{}, // replaced in startup()
	)

	return &PresenterApp{
		engine:       eng,
		audioCapture: audio.NewCaptureEngine(),
		window:       windowadapter.NewWailsWindowAdapter(),
	}
}

// NewPresenterAppWithPorts provides a constructor for testing that accepts the port
func NewPresenterAppWithPorts(eng *engine.StealthEngine, ac driven.AudioCapturePort, wp driven.WindowPort) *PresenterApp {
	return &PresenterApp{engine: eng, audioCapture: ac, window: wp}
}

// Startup is called by Wails on app start.
func (p *PresenterApp) Startup(ctx context.Context) {
	p.ctx = ctx
	p.engine.SetEventsAdapter(eventsadapter.NewWailsEventAdapter(ctx))

	// Hide from taskbar for stealth
	_ = p.window.HideFromTaskbar(ctx)

	// Auto-start microphone
	_ = p.audioCapture.Initialize()
	go p.audioCapture.StartCapture(-1, false, func(samples []float32) {
		_ = p.engine.ProcessAudio(samples)
	})
}

// GetState delegates to the engine
func (p *PresenterApp) GetState() map[string]interface{} {
	state := p.engine.GetState()
	return map[string]interface{}{
		"transcript":  state.Transcript,
		"response":    state.Response,
		"thinking":    state.Thinking,
		"cachedPairs": state.CachedPairs,
	}
}

// ClearState delegates to the engine
func (p *PresenterApp) ClearState() {
	p.engine.ClearState()
}

// AskQuestion delegates to the engine
func (p *PresenterApp) AskQuestion(q string) error {
	return p.engine.AskQuestion(q)
}

// LoadDocument parses a document and adds its text to the engine's context.
func (p *PresenterApp) LoadDocument(filepath string) (string, error) {
	text, err := parser.ParseDocument(filepath)
	if err != nil {
		return "", err
	}

	// Store in engine context
	p.engine.AddContext(text)

	// Optionally also update the state so the HUD can display it
	p.engine.UpdateState("Document loaded", text, false)

	return text, nil
}
