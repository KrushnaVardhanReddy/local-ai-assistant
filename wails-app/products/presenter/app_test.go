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

func (d *DummyCache) Search(embedding []float32, threshold float64) (string, bool)    { return "", false }
func (d *DummyCache) Store(question, answer string) error                             { return nil }
func (d *DummyCache) Count() int                                                      { return 0 }
func (d *DummyCache) IndexDocumentChunk(path, text string, embedding []float32) error { return nil }
func (d *DummyCache) GetIndexedPaths() ([]string, error)                              { return nil, nil }
func (d *DummyCache) RemoveIndexedPath(path string) error                             { return nil }
func (d *DummyCache) SemanticSearch(embedding []float32, limit int, threshold float64) ([]string, error) {
	return nil, nil
}
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

func TestPresenterAppWorkspace(t *testing.T) {
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
	app.ctx = context.Background() // For pure non-UI logic this is fine if dialogs are not mocked, but dialogs require runtime context which errors.

	// Test non-dialog workspace methods
	tempDir := t.TempDir()
	txtPath := filepath.Join(tempDir, "workspace_doc.txt")
	os.WriteFile(txtPath, []byte("Hello Workspace"), 0644)
	mdPath := filepath.Join(tempDir, "workspace_doc.md")
	os.WriteFile(mdPath, []byte("# Markdown"), 0644)

	// OpenDirectory (via engine since PromptOpenDirectory needs runtime)
	_, err := eng.OpenDirectory(tempDir)
	if err != nil {
		t.Fatalf("OpenDirectory failed: %v", err)
	}

	tree := app.GetWorkspaceTree()
	if len(tree) == 0 {
		t.Fatalf("expected non-empty workspace tree")
	}

	// OpenFile
	doc, err := app.OpenFile(txtPath)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	if doc.Content != "Hello Workspace" {
		t.Errorf("expected content Hello Workspace, got %v", doc.Content)
	}

	// SetActiveDocument
	_, err = app.OpenFile(mdPath)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}

	activeDoc, err := app.SetActiveDocument(txtPath)
	if err != nil {
		t.Fatalf("SetActiveDocument failed: %v", err)
	}
	if activeDoc.Path != txtPath {
		t.Errorf("expected active doc path %s, got %s", txtPath, activeDoc.Path)
	}

	// CloseFile
	err = app.CloseFile(mdPath)
	if err != nil {
		t.Fatalf("CloseFile failed: %v", err)
	}
}
