package session

import (
	"sync"
	"time"
)

// Turn represents a single interaction turn in a session.
type Turn struct {
	TurnIndex  int       `json:"turn"`
	Transcript string    `json:"transcript"`
	Response   string    `json:"response"`
	StartedAt  time.Time `json:"started_at"`
	AnsweredAt time.Time `json:"answered_at"`
	LatencyMs  int       `json:"latency_ms"`
}

// SessionManager manages the state and turns of an interview session.
type SessionManager struct {
	mu                   sync.RWMutex
	turns                []Turn
	currentTranscript    string
	currentResponseParts []string
	turnStartedAt        time.Time
	sessionStartedAt     time.Time
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		turns:            make([]Turn, 0),
		sessionStartedAt: time.Now(),
	}
}

// StartTurn begins a new turn with the given transcript.
func (sm *SessionManager) StartTurn(transcript string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.currentTranscript = transcript
	sm.currentResponseParts = make([]string, 0)
	sm.turnStartedAt = time.Now()
}

// AppendToken appends a token to the current response.
func (sm *SessionManager) AppendToken(token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.currentResponseParts = append(sm.currentResponseParts, token)
}

// CompleteTurn completes the current turn, calculates latency, and appends it to the turn history.
func (sm *SessionManager) CompleteTurn() *Turn {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentTranscript == "" && len(sm.currentResponseParts) == 0 {
		return nil
	}

	response := ""
	for _, part := range sm.currentResponseParts {
		response += part
	}

	now := time.Now()
	latencyMs := int(now.Sub(sm.turnStartedAt).Milliseconds())

	turn := Turn{
		TurnIndex:  len(sm.turns) + 1,
		Transcript: sm.currentTranscript,
		Response:   response,
		StartedAt:  sm.turnStartedAt,
		AnsweredAt: now,
		LatencyMs:  latencyMs,
	}

	sm.turns = append(sm.turns, turn)

	// Reset current turn state
	sm.currentTranscript = ""
	sm.currentResponseParts = nil

	return &turn
}

// GetRecentTurns returns up to the last N completed turns in chronological order.
func (sm *SessionManager) GetRecentTurns(n int) []Turn {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	numTurns := len(sm.turns)
	if numTurns == 0 {
		return []Turn{}
	}

	if n > numTurns {
		n = numTurns
	}

	recent := make([]Turn, n)
	copy(recent, sm.turns[numTurns-n:])
	return recent
}

// Export returns the session data as a map for JSON serialization.
func (sm *SessionManager) Export() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"session_started_at": sm.sessionStartedAt,
		"turn_count":         len(sm.turns),
		"turns":              sm.turns,
	}
}

// Clear resets the session manager state.
func (sm *SessionManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.turns = make([]Turn, 0)
	sm.currentTranscript = ""
	sm.currentResponseParts = nil
	sm.sessionStartedAt = time.Now()
}
