package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"wails-app/backend"
	"wails-app/core/engine"
	"strings"
	"fmt"
)

func TestApp_GetAudioDevices(t *testing.T) {
	app := NewApp()
	devices := app.GetAudioDevices()
	// Length might be zero in CI, but it shouldn't panic
	if devices == nil {
		t.Error("Expected non-nil devices list")
	}
}

func TestApp_SetAudioDevice_Invalid(t *testing.T) {
	app := NewApp()
	// Start with an invalid index
	err := app.SetAudioDevice(-1, false)
	if err == nil {
		t.Error("Expected error setting invalid audio device")
	}
}

func TestApp_StartupShutdown(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())
	// Test basic execution path
	app.shutdown(context.Background())
}

func TestApp_WorkspaceMethods(t *testing.T) {
	app := NewApp()
	// Create a temp workspace directory
	tempDir := t.TempDir()
	txtPath := filepath.Join(tempDir, "testdoc.txt")
	err := os.WriteFile(txtPath, []byte("Hello workspace"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 1. OpenDirectory
	node, err := app.OpenDirectory(tempDir)
	if err != nil {
		t.Fatalf("OpenDirectory failed: %v", err)
	}
	if node == nil || node.Path != tempDir {
		t.Errorf("Expected valid FileNode with path %s", tempDir)
	}

	// 2. GetWorkspaceTree
	tree := app.GetWorkspaceTree()
	if len(tree) == 0 {
		t.Error("Expected non-empty workspace tree")
	}

	// 3. OpenFile
	doc, err := app.OpenFile(txtPath)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	if doc == nil || doc.Path != txtPath {
		t.Errorf("Expected valid WorkspaceDocument with path %s", txtPath)
	}

	// 4. GetOpenDocuments
	docs := app.GetOpenDocuments()
	if len(docs) == 0 {
		t.Error("Expected non-empty open documents")
	}

	// 5. SetActiveDocument & GetActiveDocument
	err = app.SetActiveDocument(txtPath)
	if err != nil {
		t.Fatalf("SetActiveDocument failed: %v", err)
	}

	active := app.GetActiveDocument()
	if active == nil || active.Path != txtPath {
		t.Errorf("Expected active document to be %s", txtPath)
	}

	// 6. GetIDEState
	state := app.GetIDEState()
	if state == nil {
		t.Fatal("GetIDEState returned nil")
	}

	if path, ok := state["activeDocumentPath"].(string); !ok || path != txtPath {
		t.Errorf("Expected activeDocumentPath to be %s, got %v", txtPath, state["activeDocumentPath"])
	}

	if content, ok := state["activeDocumentContent"].(string); !ok || content != "Hello workspace" {
		t.Errorf("Expected activeDocumentContent 'Hello workspace', got %v", state["activeDocumentContent"])
	}

	// 7. CloseDocument
	err = app.CloseDocument(txtPath)
	if err != nil {
		t.Fatalf("CloseDocument failed: %v", err)
	}

	docsAfterClose := app.GetOpenDocuments()
	if len(docsAfterClose) != 0 {
		t.Errorf("Expected 0 open documents after close, got %d", len(docsAfterClose))
	}
}

func TestApp_PromptMethodsContext(t *testing.T) {
	app := NewApp()
	// Should fail because context is nil
	_, err := app.PromptOpenDirectory()
	if err == nil {
		t.Error("Expected error when context is nil")
	}

	_, err = app.PromptOpenFile()
	if err == nil {
		t.Error("Expected error when context is nil")
	}
}

func TestApp_ToggleMic(t *testing.T) {
	app := NewApp()

	// Wait for any async initialization to avoid races, though NewApp doesn't start capture by itself

	// Initial toggle should return true (start capture) or false (fail to start due to CI).
	// It relies on SetAudioDevice which might fail in CI and print log but doesn't panic.
	// Actually, SetAudioDevice returns an error if it fails, and the capturing state might not change.
	app.ToggleMic()

	// Another toggle
	app.ToggleMic()
}


type mockCacheAdapter struct {
    items []backend.CacheItem
    err   error
}

func (m *mockCacheAdapter) GetAllItems() ([]backend.CacheItem, error) {
    return m.items, m.err
}
func (m *mockCacheAdapter) Store(item backend.CacheItem) error { return nil }
func (m *mockCacheAdapter) Search(query string, limit int) ([]backend.CacheItem, error) { return nil, nil }
func (m *mockCacheAdapter) DeleteItem(id string) error { return nil }
func (m *mockCacheAdapter) ClearAll() error { return nil }
func (m *mockCacheAdapter) EnsureIndex() error { return nil }

func TestApp_ExportSession(t *testing.T) {
	app := NewApp()

    mockCache := &mockCacheAdapter{
        items: []backend.CacheItem{
            {Question: "Q1", Answer: "A1"},
            {Question: "Q2", Answer: "A2"},
        },
    }
    app.engine = engine.New(
		engine.Config{},
		nil,
		nil,
		mockCache,
		nil,
	)

    tempDir := t.TempDir()
    t.Setenv("HOME", tempDir)
    t.Setenv("USERPROFILE", tempDir) // for Windows

    filePath, err := app.ExportSession()
    if err != nil {
        t.Fatalf("ExportSession failed: %v", err)
    }

    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        t.Fatalf("Expected file to exist at %s", filePath)
    }

    content, err := os.ReadFile(filePath)
    if err != nil {
        t.Fatalf("Failed to read exported file: %v", err)
    }

    contentStr := string(content)
    if !strings.Contains(contentStr, "## Q1: Q1") {
        t.Errorf("Expected content to contain Q1, got: %s", contentStr)
    }
    if !strings.Contains(contentStr, "A2") {
        t.Errorf("Expected content to contain A2, got: %s", contentStr)
    }

    // Test error case
    mockCache.err = fmt.Errorf("mock error")
    _, err = app.ExportSession()
    if err == nil {
        t.Fatalf("Expected error when GetAllItems fails")
    }

    // Test invalid cache adapter type (doesn't implement GetAllItems)
    app.engine = engine.New(engine.Config{}, nil, nil, nil, nil)
    _, err = app.ExportSession()
    if err == nil {
        t.Fatalf("Expected error when cache adapter doesn't implement GetAllItems")
    }
}

func TestApp_SetManualMode(t *testing.T) {
	app := NewApp()
	app.SetManualMode(true)
	// Just verifies it doesn't panic
}
