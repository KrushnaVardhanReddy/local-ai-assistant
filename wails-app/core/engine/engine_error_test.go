package engine_test

import (
	"fmt"
	"testing"
	"wails-app/backend/stt"
	"wails-app/core/engine"
)

type errorSTT struct{}

func (m *errorSTT) Close() error { return nil }
func (m *errorSTT) TranscribeStream(audio []float32) (chan string, error) {
	return nil, fmt.Errorf("mock error")
}

func TestProcessAudio_STTError(t *testing.T) {
	_ = stt.NewSTTManager(&errorSTT{})
	// Mock process to fail? wait, TranscribeStream in sttManager is asynchronous.
	// We might need to induce error in TranscribeStream... Actually, STTManager.TranscribeStream only errors if engine is nil or missing.
	// Let's pass nil to NewSTTManager.
	sttMgr2 := stt.NewSTTManager(nil)
	eng := engine.New(engine.Config{}, sttMgr2, nil, nil, nil)
	err := eng.ProcessAudio([]float32{0})
	if err == nil {
		t.Errorf("Expected STT error from nil engine")
	}
}
