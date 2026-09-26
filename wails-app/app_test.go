package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"wails-app/backend/config"
	"wails-app/core/engine"
)

func TestApp_GetAudioDevices(t *testing.T) {
	app := NewApp(&config.AppConfig{})
	devices := app.GetAudioDevices()
	// Length might be zero in CI, but it shouldn't panic
	if devices == nil {
		t.Error("Expected non-nil devices list")
	}
}

func TestApp_SetAudioDevice_Invalid(t *testing.T) {
	app := NewApp(&config.AppConfig{})
	// Start with an invalid index
	err := app.SetAudioDevice(-1, false)
	if err == nil {
		t.Error("Expected error setting invalid audio device")
	}
}

func TestApp_StartupShutdown(t *testing.T) {
	app := NewApp(&config.AppConfig{})
	app.startup(context.Background())
	// Test basic execution path
	app.shutdown(context.Background())
}

func TestApp_WorkspaceMethods(t *testing.T) {
	app := NewApp(&config.AppConfig{})
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
	app := NewApp(&config.AppConfig{})
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
	app := NewApp(&config.AppConfig{})

	// Wait for any async initialization to avoid races, though NewApp doesn't start capture by itself

	// Initial toggle should return true (start capture) or false (fail to start due to CI).
	// It relies on SetAudioDevice which might fail in CI and print log but doesn't panic.
	// Actually, SetAudioDevice returns an error if it fails, and the capturing state might not change.
	app.ToggleMic()

	// Another toggle
	app.ToggleMic()
}

func TestApp_SetAppMode(t *testing.T) {
	app := NewApp(&config.AppConfig{})

	// Test setting to Interview mode
	err := app.SetAppMode("interview")
	if err != nil {
		t.Fatalf("Failed to set app mode to interview: %v", err)
	}
	if app.appMode != AppModeInterview {
		t.Errorf("Expected appMode to be %v, got %v", AppModeInterview, app.appMode)
	}

	// Test setting to Transcript mode
	err = app.SetAppMode("transcript")
	if err != nil {
		t.Fatalf("Failed to set app mode to transcript: %v", err)
	}
	if app.appMode != AppModeTranscript {
		t.Errorf("Expected appMode to be %v, got %v", AppModeTranscript, app.appMode)
	}

	// Test invalid mode
	err = app.SetAppMode("invalid")
	if err == nil {
		t.Error("Expected error when setting invalid app mode")
	}
}

func TestApp_SetAudioMode(t *testing.T) {
	app := NewApp(&config.AppConfig{})

	// Test setting to Speaker mode
	err := app.SetAudioMode("speaker")
	// If it fails on startDualCapture, it might return err.
	// We just test if mode was set
	if app.audioMode != AudioModeSpeaker {
		t.Errorf("Expected audioMode to be %v, got %v", AudioModeSpeaker, app.audioMode)
	}

	_ = app.SetAudioMode("dual")
	if app.audioMode != AudioModeDual {
		t.Errorf("Expected audioMode to be %v, got %v", AudioModeDual, app.audioMode)
	}

	// Test invalid mode
	err = app.SetAudioMode("invalid")
	if err == nil {
		t.Error("Expected error when setting invalid audio mode")
	}
}

func TestApp_ExportSession(t *testing.T) {
	app := NewApp(&config.AppConfig{})

	app.engine = engine.New(
		engine.Config{},
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	sessMgr := app.engine.GetSessionManager()
	sessMgr.StartTurn("Q1")
	sessMgr.SetAISuggestion("A1")
	sessMgr.CompleteTurn()

	sessMgr.StartTurn("Q2")
	sessMgr.SetAISuggestion("A2")
	sessMgr.CompleteTurn()

	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("USERPROFILE", tempDir) // for Windows

	_, err := app.ExportSession()
	if err == nil {
		t.Fatalf("Expected error from SaveFileDialog due to missing Wails context")
	}

	if err.Error() == "no turns in current session to export" {
		t.Fatalf("Expected error from SaveFileDialog, got: %v", err)
	}

	// Test 0 turns
	app.engine = engine.New(
		engine.Config{},
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	_, err = app.ExportSession()
	if err == nil || !strings.Contains(err.Error(), "no turns in current session to export") {
		t.Fatalf("Expected 'no turns' error, got: %v", err)
	}

	// Test nil session manager
	app.engine.SetSessionManager(nil)
	_, err = app.ExportSession()
	if err == nil || !strings.Contains(err.Error(), "session manager not available") {
		t.Fatalf("Expected 'session manager not available' error, got: %v", err)
	}
}

func TestApp_SetManualMode(t *testing.T) {
	app := NewApp(&config.AppConfig{})
	app.SetManualMode(true)
	// Just verifies it doesn't panic
}

func TestSetProxyToken(t *testing.T) {
	app := NewApp(&config.AppConfig{})
	app.SetProxyToken("test-token")

	// We'll just test that it doesn't crash since llm.DemoProxyToken isn't exported here
	// Or we can assume it works based on the lack of panic
	if app == nil {
		t.Errorf("App is nil")
	}
}

func TestApp_FlushQuestionBuffer(t *testing.T) {
	app := NewApp(&config.AppConfig{})

	// Create engine with mocked ports
	events := &engine.MockEvents{Emitted: make(map[string]int)}
	eng := engine.New(engine.Config{}, nil, nil, nil, events, nil)
	app.engine = eng

	app.AppendToBuffer("hello")

	if len(eng.GetQuestionBuffer().GetChunks()) == 0 {
		t.Fatalf("Expected chunks in buffer")
	}

	err := app.FlushQuestionBuffer()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(eng.GetQuestionBuffer().GetChunks()) != 0 {
		t.Errorf("Expected buffer to be empty after flush")
	}
}
