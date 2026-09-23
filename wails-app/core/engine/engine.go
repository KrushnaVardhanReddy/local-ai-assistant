package engine

import (
	"log"
	"strings"
	"fmt"
	"context"
	"sync"
	"wails-app/backend/session"
	"wails-app/backend/stt"
	"wails-app/core/ports/driven"
	"wails-app/core/ports/driving"
	"wails-app/backend/classifier"
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

	questionBuffer *classifier.QuestionBuffer

	// workspace state
	workspaceTree []*driving.FileNode
	openDocuments map[string]*driving.WorkspaceDocument
	activeDoc     *driving.WorkspaceDocument
	workspaceMu   sync.RWMutex

	includeActiveDocMu sync.RWMutex
	includeActiveDoc   bool

	// internal state (mutex-protected)
	mu               sync.RWMutex
	transcript       string
	transcriptBuffer []string // rolling window of last 5 accepted transcripts
	response         string
	thinking         bool
	llmBusy        sync.Mutex
	inFlightMu     sync.Mutex
	cancelInFlight context.CancelFunc
	inFlightCtx    context.Context
	manualMode     bool
	rawMode        bool
}

func New(
	cfg Config,
	sttMgr *stt.STTManager,
	llm driven.LLMPort,
	cache driven.CachePort,
	events driven.EventPort,
) *StealthEngine {
	e := &StealthEngine{
		cfg:              cfg,
		sttManager:       sttMgr,
		llm:              llm,
		cache:            cache,
		events:           events,
		sessionMgr:       session.NewSessionManager(),
		openDocuments:    make(map[string]*driving.WorkspaceDocument),
		includeActiveDoc: true,
	}
	e.questionBuffer = classifier.NewQuestionBuffer(e.triggerLLMWithQuestion)
	return e
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

	if e.questionBuffer != nil {
		e.questionBuffer.Reset()
	}
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

// SetSessionManager sets the session manager instance.
func (e *StealthEngine) SetSessionManager(sm *session.SessionManager) {
	e.sessionMgr = sm
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

func (e *StealthEngine) SetRawMode(enabled bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rawMode = enabled
}

// SummaryRequest holds the configuration for a single summary template.
type SummaryRequest struct {
	ID     string `json:"id"`
	Prompt string `json:"prompt"`
}

// SummarizeSession uses the LLM to summarize the entire interview session
// using predefined or custom templates. It streams the results asynchronously.
func (e *StealthEngine) SummarizeSession(requests []SummaryRequest) error {
	if e.sessionMgr == nil {
		return fmt.Errorf("session manager not configured")
	}

	turns := e.sessionMgr.GetRecentTurns(10000)
	if len(turns) == 0 {
		return fmt.Errorf("no session history to summarize")
	}

	var transcript strings.Builder
	for _, t := range turns {
		transcript.WriteString(fmt.Sprintf("Interviewer: %s\n", t.InterviewerQuestion))
		if t.CandidateResponse != "" {
			transcript.WriteString(fmt.Sprintf("Candidate: %s\n", t.CandidateResponse))
		}
		transcript.WriteString("\n")
	}

	transcriptStr := transcript.String()

	for _, req := range requests {
		go func(r SummaryRequest) {
			sysPrompt := r.Prompt
			if sysPrompt == "" {
				sysPrompt = "You are a helpful assistant. Summarize the interview."
			}

			// We use a completely detached context so in-flight cancellations do not affect summaries.
			ctx := context.Background()

			if e.events != nil {
				e.events.Emit("on_summary_start", map[string]interface{}{"id": r.ID})
			}

			err := e.llm.StreamCompletion(
				ctx,
				transcriptStr,
				sysPrompt,
				nil,
				func(token string) {
					if e.events != nil {
						e.events.Emit("on_summary_token", map[string]interface{}{
							"id":   r.ID,
							"text": token,
						})
					}
				},
				func() {
					if e.events != nil {
						e.events.Emit("on_summary_end", map[string]interface{}{"id": r.ID})
					}
				},
			)

			if err != nil {
				// Handle error
				if e.events != nil {
					e.events.Emit("on_summary_token", map[string]interface{}{
						"id":   r.ID,
						"text": fmt.Sprintf("\n[Error: %v]", err),
					})
					e.events.Emit("on_summary_end", map[string]interface{}{"id": r.ID})
				}
			}
		}(req)
	}

	return nil
}


func (e *StealthEngine) Start(ctx context.Context) {
	if e.questionBuffer != nil {
		e.questionBuffer.Start()
	}
	if err := classifier.DefaultLlamaServer.Start(ctx); err != nil {
		log.Printf("[Engine] Gemma sidecar unavailable (turn-detection disabled): %v\n", err)
	}
}

func (e *StealthEngine) Stop() {
	if e.questionBuffer != nil {
		e.questionBuffer.Stop()
	}
	classifier.DefaultLlamaServer.Stop()
}

func (e *StealthEngine) GetQuestionBuffer() *classifier.QuestionBuffer {
	return e.questionBuffer
}