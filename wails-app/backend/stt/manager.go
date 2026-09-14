package stt

import (
	"errors"
	"sync"
)

// STTManager manages the active STTEngine and allows hot-swapping between implementations
type STTManager struct {
	activeEngine STTEngine
	mu           sync.RWMutex
}

// NewSTTManager creates a new STTManager initialized with the given engine
func NewSTTManager(initialEngine STTEngine) *STTManager {
	return &STTManager{
		activeEngine: initialEngine,
	}
}

// SwapEngine cleanly closes the old engine (if any) and sets the new engine as active
func (m *STTManager) SwapEngine(newEngine STTEngine) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cleanly close the old engine if it exists
	if m.activeEngine != nil {
		if err := m.activeEngine.Close(); err != nil {
			// Even if closing fails, we still want to swap out the engine,
			// but we return the error to let the caller know it failed to close gracefully.
			m.activeEngine = newEngine
			return err
		}
	}

	m.activeEngine = newEngine
	return nil
}

// TranscribeStream routes the audio to the active engine in a thread-safe way
func (m *STTManager) TranscribeStream(audio []float32) (chan string, error) {
	m.mu.RLock()
	engine := m.activeEngine
	m.mu.RUnlock()

	if engine == nil {
		return nil, errors.New("no active STT engine")
	}

	return engine.TranscribeStream(audio)
}
