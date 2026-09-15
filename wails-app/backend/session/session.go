package session

import (
	"sync"
	"time"
)

// Turn represents a single interaction turn in a session.
type Turn struct {
	TurnIndex           int       `json:"turn"`
	InterviewerQuestion string    `json:"interviewer_question"`
	CandidateResponse   string    `json:"candidate_response"`
	AISuggestion        string    `json:"ai_suggestion"`
	StartedAt           time.Time `json:"started_at"`
	AnsweredAt          time.Time `json:"answered_at"`
	LatencyMs           int       `json:"latency_ms"`
}

// SessionManager manages the state and turns of an interview session.
type SessionManager struct {
	mu                         sync.RWMutex
	turns                      []Turn
	currentInterviewerQuestion string
	currentCandidateResponse   string
	currentAISuggestion        string
	turnStartedAt              time.Time
	sessionStartedAt           time.Time
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		turns:            make([]Turn, 0),
		sessionStartedAt: time.Now(),
	}
}

// StartTurn begins a new turn with the given interviewer question.
func (sm *SessionManager) StartTurn(interviewerQuestion string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.currentInterviewerQuestion = interviewerQuestion
	sm.currentCandidateResponse = ""
	sm.currentAISuggestion = ""
	sm.turnStartedAt = time.Now()
}

// SetCandidateResponse sets the candidate's spoken response.
func (sm *SessionManager) SetCandidateResponse(response string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.currentCandidateResponse = response
}

// SetAISuggestion sets the AI suggestion.
func (sm *SessionManager) SetAISuggestion(suggestion string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.currentAISuggestion = suggestion
}

// CompleteTurn completes the current turn, calculates latency, and appends it to the turn history.
func (sm *SessionManager) CompleteTurn() *Turn {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentInterviewerQuestion == "" && sm.currentCandidateResponse == "" && sm.currentAISuggestion == "" {
		return nil
	}

	now := time.Now()
	latencyMs := int(now.Sub(sm.turnStartedAt).Milliseconds())

	turn := Turn{
		TurnIndex:           len(sm.turns) + 1,
		InterviewerQuestion: sm.currentInterviewerQuestion,
		CandidateResponse:   sm.currentCandidateResponse,
		AISuggestion:        sm.currentAISuggestion,
		StartedAt:           sm.turnStartedAt,
		AnsweredAt:          now,
		LatencyMs:           latencyMs,
	}

	sm.turns = append(sm.turns, turn)

	// Reset current turn state
	sm.currentInterviewerQuestion = ""
	sm.currentCandidateResponse = ""
	sm.currentAISuggestion = ""

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
	sm.currentInterviewerQuestion = ""
	sm.currentCandidateResponse = ""
	sm.currentAISuggestion = ""
	sm.sessionStartedAt = time.Now()
}
