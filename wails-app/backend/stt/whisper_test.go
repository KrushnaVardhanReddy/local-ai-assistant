package stt

import (
	"testing"
)

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
)

func TestSTT_MissingModel(t *testing.T) {
	_, err := LoadWhisperEngine("nonexistent_model.bin")
	if err == nil {
		t.Error("Expected error when loading nonexistent model, got nil")
	}
}

func TestEnsureWhisperModel_AlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	modelFile := filepath.Join(tempDir, "existing.bin")
	os.WriteFile(modelFile, []byte("dummy"), 0644)

	path, err := EnsureWhisperModel(modelFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path != modelFile {
		t.Fatalf("Expected path %s, got %s", modelFile, path)
	}
}

func TestEnsureWhisperModel_DownloadAtomic(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("whisper model content"))
	}))
	defer server.Close()

	origURL := WhisperModelURL
	WhisperModelURL = server.URL
	defer func() { WhisperModelURL = origURL }()

	path, err := EnsureWhisperModel("")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path != "models/ggml-base.en.bin" {
		t.Fatalf("Expected fallback path models/ggml-base.en.bin, got %s", path)
	}

	content, err := os.ReadFile("models/ggml-base.en.bin")
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if string(content) != "whisper model content" {
		t.Fatalf("Expected downloaded content, got %s", string(content))
	}
}

func TestEnsureWhisperModel_DownloadError(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	origURL := WhisperModelURL
	WhisperModelURL = server.URL
	defer func() { WhisperModelURL = origURL }()

	_, err := EnsureWhisperModel("")
	if err == nil {
		t.Fatalf("Expected error when downloading fails, got nil")
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
