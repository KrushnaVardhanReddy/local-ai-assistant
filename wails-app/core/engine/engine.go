package engine

import (
	"context"
	"sync"
	"wails-app/backend/session"
	"wails-app/backend/stt"
	"wails-app/core/ports/driven"
	"wails-app/core/ports/driving"
)

// Config holds product-specific configuration injected at startup.
type Config struct {
	// SystemPrompt is the product-specific LLM system instruction.
	// e.g. "You are a stealth interview assistant..."
	// e.g. "You are an invisible teleprompter assistant..."
	SystemPrompt string
}

// StealthEngine is the core engine that wires all ports together.
// It implements driving.PipelinePort.
type StealthEngine struct {
	cfg        Config
	sttManager *stt.STTManager
	llm        driven.LLMPort
	cache      driven.CachePort
	events     driven.EventPort

	sessionMgr *session.SessionManager

	// workspace state
	workspaceTree []*driving.FileNode
	openDocuments map[string]*driving.WorkspaceDocument
	activeDoc     *driving.WorkspaceDocument
	workspaceMu   sync.RWMutex

	includeActiveDocMu sync.RWMutex
	includeActiveDoc   bool

	// internal state (mutex-protected)
	mu             sync.RWMutex
	transcript     string
	response       string
	thinking       bool
	llmBusy        sync.Mutex
	inFlightMu     sync.Mutex
	cancelInFlight context.CancelFunc
	inFlightCtx    context.Context
	manualMode     bool
}

func New(
	cfg Config,
	sttMgr *stt.STTManager,
	llm driven.LLMPort,
	cache driven.CachePort,
	events driven.EventPort,
) *StealthEngine {
	return &StealthEngine{
		cfg:              cfg,
		sttManager:       sttMgr,
		llm:              llm,
		cache:            cache,
		events:           events,
		sessionMgr:       session.NewSessionManager(),
		openDocuments:    make(map[string]*driving.WorkspaceDocument),
		includeActiveDoc: true,
	}
}

// SetEventsAdapter sets the events port.
// Used for decoupling initialization when the Wails context is only available later.
func (e *StealthEngine) SetEventsAdapter(events driven.EventPort) {
	e.events = events
}

// GetState returns the current pipeline state for UI polling.
func (e *StealthEngine) GetState() driving.PipelineState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	cachedPairs := 0
	if e.cache != nil {
		cachedPairs = e.cache.Count()
	}
	return driving.PipelineState{
		Transcript:  e.transcript,
		Response:    e.response,
		Thinking:    e.thinking,
		CachedPairs: cachedPairs,
	}
}

// ClearState resets the pipeline state.
func (e *StealthEngine) ClearState() {
	e.inFlightMu.Lock()
	if e.cancelInFlight != nil {
		e.cancelInFlight()
	}
	e.inFlightMu.Unlock()

	e.mu.Lock()
	defer e.mu.Unlock()
	e.transcript = ""
	e.response = ""
	e.thinking = false
}

// SetStealth is a no-op here — window management is handled by the
// product-level App shell which has access to the Wails context.
func (e *StealthEngine) SetStealth(_ bool) error { return nil }

// Compile-time check.
var _ driving.PipelinePort = (*StealthEngine)(nil)

// GetSessionManager returns the session manager instance.
func (e *StealthEngine) GetSessionManager() *session.SessionManager {
	return e.sessionMgr
}

// GetCache returns the cache port instance.
func (e *StealthEngine) GetCache() driven.CachePort {
	return e.cache
}

// UpdateState allows modifying the pipeline state manually, e.g., for vision queries.
func (e *StealthEngine) UpdateState(transcript, response string, thinking bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.transcript = transcript
	e.response = response
	e.thinking = thinking
}

func (e *StealthEngine) SetIncludeActiveDocContext(enabled bool) {
	e.includeActiveDocMu.Lock()
	defer e.includeActiveDocMu.Unlock()
	e.includeActiveDoc = enabled
}

func (e *StealthEngine) GetIncludeActiveDocContext() bool {
	e.includeActiveDocMu.RLock()
	defer e.includeActiveDocMu.RUnlock()
	return e.includeActiveDoc
}

// AddContext adds parsed document text or other context to the engine's system prompt or session.
func (e *StealthEngine) AddContext(contextText string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Append context to the system prompt so the LLM knows about it
	e.cfg.SystemPrompt += "\n\nAdditional Context:\n" + contextText
}

// SetManualMode toggles whether incoming STT transcripts are automatically sent to the LLM.
func (e *StealthEngine) SetManualMode(enabled bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.manualMode = enabled
}
