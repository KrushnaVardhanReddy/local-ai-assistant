package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
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
