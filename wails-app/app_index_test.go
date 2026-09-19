//go:build !cgo

package main

import (
	"os"
	"path/filepath"
	"testing"
	"wails-app/backend"
	"wails-app/core/engine"
	cacheadapter "wails-app/adapters/cache"
)

// In a real environment with working CGO, these would run.
func TestApp_Indexing(t *testing.T) {
	// Initialize memory DB for test
	db, err := backend.NewVectorDB(":memory:")
	if err != nil {
		t.Skip("Skipping DB test - memory sqlite unsupported in this test env", err)
	}
	defer db.Close()
	
	cacheAdapter := cacheadapter.NewSQLiteVecAdapter(db)
	
	// Stub engine with just cache adapter
	eng := engine.New(
		engine.Config{},
		nil,
		nil,
		cacheAdapter,
		nil,
	)
	
	// Minimal app instance
	app := &App{
		engine: eng,
	}

	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(txtFile, []byte("Hello world indexing test"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = app.IndexFile(txtFile)
	if err != nil {
		t.Errorf("IndexFile failed: %v", err)
	}

	paths := app.GetIndexedPaths()
	if len(paths) == 0 {
		t.Errorf("GetIndexedPaths returned empty")
	}

	found := false
	for _, p := range paths {
		if p == txtFile {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GetIndexedPaths did not contain test file")
	}

	err = app.RemoveIndexedPath(txtFile)
	if err != nil {
		t.Errorf("RemoveIndexedPath failed: %v", err)
	}
	
	paths = app.GetIndexedPaths()
	for _, p := range paths {
		if p == txtFile {
			t.Errorf("GetIndexedPaths contained test file after removal")
		}
	}
}

func TestApp_Indexing_CacheNil(t *testing.T) {
	eng := engine.New(
		engine.Config{},
		nil,
		nil,
		nil,
		nil,
	)
	
	app := &App{
		engine: eng,
	}

	err := app.IndexFile("dummy.txt")
	if err == nil {
		t.Errorf("IndexFile expected error with nil cache")
	}

	paths := app.GetIndexedPaths()
	if len(paths) != 0 {
		t.Errorf("GetIndexedPaths expected 0 paths with nil cache")
	}

	err = app.RemoveIndexedPath("dummy.txt")
	if err != nil {
		t.Errorf("RemoveIndexedPath expected nil with nil cache")
	}
}
