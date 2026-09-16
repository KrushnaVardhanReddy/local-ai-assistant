package presenter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wails-app/adapters/events"
	llmadapter "wails-app/adapters/llm"
	"wails-app/backend/stt"
	"wails-app/core/engine"
)

// Define dummy cache adapter so we don't import the one that needs CGO
type DummyCache struct{}
func (d *DummyCache) Search(embedding []float32, threshold float64) (string, bool) { return "", false }
func (d *DummyCache) Store(question, answer string) error { return nil }
func (d *DummyCache) Count() int { return 0 }
func (d *DummyCache) Close() error { return nil }

func TestPresenterAppLoadDocument(t *testing.T) {
	sttManager := stt.NewSTTManager(nil)
	llmAdapter := llmadapter.NewOpenAIAdapter()

	eng := engine.New(
		engine.Config{SystemPrompt: SystemPrompt},
		sttManager,
		llmAdapter,
		&DummyCache{},
		&eventsadapter.NoopEventAdapter{},
	)

	mockAudio := &MockAudioCapture{}
	mockWindow := &MockWindowPort{}
	app := NewPresenterAppWithPorts(eng, mockAudio, mockWindow)
	app.ctx = context.Background()

	tempDir := t.TempDir()
	txtPath := filepath.Join(tempDir, "doc.txt")
	os.WriteFile(txtPath, []byte("Parsed stealth document content."), 0644)

	content, err := app.LoadDocument(txtPath)
	if err != nil {
		t.Fatalf("LoadDocument failed: %v", err)
	}
	if !strings.Contains(content, "stealth") {
		t.Errorf("unexpected content: %q", content)
	}

	state := app.GetState()
	if scriptStr, ok := state["script"].(string); !ok || scriptStr != content {
		t.Errorf("expected state script to be %q, got %v", content, state["script"])
	}

	app.ClearState()
	stateAfterClear := app.GetState()
	if scriptStr, ok := stateAfterClear["script"].(string); !ok || scriptStr != "" {
		t.Errorf("expected script to be empty after clear, got %v", stateAfterClear["script"])
	}
}
