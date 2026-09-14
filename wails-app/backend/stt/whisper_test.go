package stt

import (
	"testing"
)

func TestSTT_MissingModel(t *testing.T) {
	_, err := LoadWhisperEngine("nonexistent_model.bin")
	if err == nil {
		t.Error("Expected error when loading nonexistent model, got nil")
	}
}

func TestSTT_InvalidEngine(t *testing.T) {
	s := &WhisperEngine{model: nil}

	// Test TranscribeStream when model is nil
	samples := []float32{0.0, 0.1, 0.2}
	ch, err := s.TranscribeStream(samples)
	if err == nil {
		t.Error("Expected error when transcribing stream with nil model")
	}
	if ch != nil {
		t.Errorf("Expected nil channel, got %v", ch)
	}
}

func TestSTT_CloseEngine(t *testing.T) {
	s := &WhisperEngine{model: nil}
	err := s.Close()
	if err != nil {
		t.Errorf("Expected no error when closing nil model, got %v", err)
	}
}
