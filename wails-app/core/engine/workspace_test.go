package engine

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"wails-app/core/ports/driving"
)

// mockCachePort is a mock implementation of driven.CachePort for testing.
type mockCachePort struct {
	stored map[string]string
}

func (m *mockCachePort) Search(embedding []float32, threshold float64) (string, bool) {
	return "", false
}

func (m *mockCachePort) Store(question, answer string) error {
	m.stored[question] = answer
	return nil
}

func (m *mockCachePort) Count() int {
	return len(m.stored)
}

// mockEventPort is a mock implementation of driven.EventPort for testing.
type mockEventPort struct {
	emitted map[string]any
}

func (m *mockEventPort) Emit(event string, payload any) {
	m.emitted[event] = payload
}

func TestStealthEngine_Workspace(t *testing.T) {
	// Setup temporary directory structure
	tempDir := t.TempDir()

	// Create valid files
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content 1"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.md"), []byte("content 2"), 0644)

	// Create unsupported files
	os.WriteFile(filepath.Join(tempDir, "image.png"), []byte("fake image"), 0644)

	// Create hidden files and directories
	os.WriteFile(filepath.Join(tempDir, ".hidden.txt"), []byte("hidden"), 0644)
	gitDir := filepath.Join(tempDir, ".git")
	os.Mkdir(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0644)
	nodeModulesDir := filepath.Join(tempDir, "node_modules")
	os.Mkdir(nodeModulesDir, 0755)
	os.WriteFile(filepath.Join(nodeModulesDir, "package.json"), []byte("{}"), 0644)

	// Create subfolder with valid files
	subDir := filepath.Join(tempDir, "docs")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "doc1.pdf"), []byte("fake pdf"), 0644)

	// Create empty folder
	emptyDir := filepath.Join(tempDir, "empty")
	os.Mkdir(emptyDir, 0755)

	cache := &mockCachePort{stored: make(map[string]string)}
	events := &mockEventPort{emitted: make(map[string]any)}

	engine := New(Config{}, nil, nil, cache, events)

	t.Run("OpenDirectory", func(t *testing.T) {
		node, err := engine.OpenDirectory(tempDir)
		if err != nil {
			t.Fatalf("OpenDirectory failed: %v", err)
		}

		if node == nil {
			t.Fatal("Expected node to not be nil")
		}

		// Verify event emitted
		if payload, ok := events.emitted["on_workspace_tree_updated"]; !ok {
			t.Errorf("Expected on_workspace_tree_updated event to be emitted")
		} else {
			nodes, ok := payload.([]*driving.FileNode)
			if !ok || len(nodes) != 1 || nodes[0].Path != tempDir {
				t.Errorf("Unexpected event payload: %+v", payload)
			}
		}

		// Verify structure
		if len(node.Children) != 3 { // docs folder, file1.txt, file2.md
			t.Errorf("Expected 3 children, got %d", len(node.Children))
		}

		// Check sorting: folders first, then alphabetically
		if node.Children[0].Name != "docs" {
			t.Errorf("Expected first child to be 'docs', got %s", node.Children[0].Name)
		}
		if node.Children[1].Name != "file1.txt" {
			t.Errorf("Expected second child to be 'file1.txt', got %s", node.Children[1].Name)
		}
		if node.Children[2].Name != "file2.md" {
			t.Errorf("Expected third child to be 'file2.md', got %s", node.Children[2].Name)
		}

		// Also check that a file is sorted after a directory
		// We can add a file "a.txt" and directory "z_dir" and it should sort "z_dir" first.
		// We'll create these now in tempDir and re-scan.
		os.WriteFile(filepath.Join(tempDir, "a.txt"), []byte("content"), 0644)
		os.Mkdir(filepath.Join(tempDir, "z_dir"), 0755)
		os.WriteFile(filepath.Join(tempDir, "z_dir", "file.txt"), []byte("content"), 0644)

		node2, err2 := engine.OpenDirectory(tempDir)
		if err2 != nil {
			t.Fatalf("OpenDirectory failed on re-scan: %v", err2)
		}
		if node2.Children[0].Name != "docs" || node2.Children[1].Name != "z_dir" {
			t.Errorf("Expected folders to be first: %s, %s", node2.Children[0].Name, node2.Children[1].Name)
		}
		if node2.Children[2].Name != "a.txt" || node2.Children[3].Name != "file1.txt" || node2.Children[4].Name != "file2.md" {
			t.Errorf("Expected files to be sorted alphabetically after folders")
		}

		// Check subfolder
		if len(node.Children[0].Children) != 1 || node.Children[0].Children[0].Name != "doc1.pdf" {
			t.Errorf("Expected docs folder to have 1 child 'doc1.pdf'")
		}

		// Verify GetWorkspaceTree
		tree := engine.GetWorkspaceTree()
		if len(tree) != 1 || tree[0].Path != tempDir {
			t.Errorf("Unexpected tree from GetWorkspaceTree: %+v", tree)
		}
	})

	t.Run("OpenDirectory - Invalid Path", func(t *testing.T) {
		_, err := engine.OpenDirectory(filepath.Join(tempDir, "does-not-exist"))
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("OpenDirectory - Empty valid files", func(t *testing.T) {
		emptyDir2 := t.TempDir()
		_, err := engine.OpenDirectory(emptyDir2)
		if err == nil {
			t.Error("Expected error for directory with no valid files")
		}
	})

	t.Run("OpenDirectory - Stat Error Child", func(t *testing.T) {
		// To trigger stat error inside buildFileTree on a child, we can create a file,
		// start scanning, but delete the file before child stat or mock it.
		// A simpler way is just to create a directory, then pass a path that we know will fail os.ReadDir if it was possible.
		// Since we want to cover `if err != nil { continue }` in the children loop,
		// we can make a sub-folder that has no read permissions.
		noReadDir := filepath.Join(tempDir, "noread")
		os.Mkdir(noReadDir, 0333) // write and execute, no read

		// Running this should skip the "noread" child and not fail the whole scan,
		// or if buildFileTree on "noread" returns an error, it gets skipped.
		node, err := engine.OpenDirectory(tempDir)
		if err != nil {
			// Clean up permission so it can be deleted later
			os.Chmod(noReadDir, 0755)
			t.Fatalf("Expected no error when skipping unreadable child, got: %v", err)
		}
		if node == nil {
			os.Chmod(noReadDir, 0755)
			t.Fatal("Expected node to not be nil")
		}

		os.Chmod(noReadDir, 0755)
	})

	t.Run("OpenFile", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "file1.txt")

		doc, err := engine.OpenFile(filePath)
		if err != nil {
			t.Fatalf("OpenFile failed: %v", err)
		}

		if doc == nil {
			t.Fatal("Expected doc to not be nil")
		}

		if doc.Content != "content 1" {
			t.Errorf("Expected content 'content 1', got %s", doc.Content)
		}

		if engine.GetActiveDocument() != doc {
			t.Errorf("Expected active document to be %v, got %v", doc, engine.GetActiveDocument())
		}

		// Verify event emitted
		if payload, ok := events.emitted["on_active_document_changed"]; !ok {
			t.Errorf("Expected on_active_document_changed event to be emitted")
		} else {
			if !reflect.DeepEqual(payload, doc) {
				t.Errorf("Unexpected event payload: %+v", payload)
			}
		}

		// Verify cache stored
		key := "Document: file1.txt"
		if val, ok := cache.stored[key]; !ok || val != "content 1" {
			t.Errorf("Expected cache to store %s -> 'content 1', got %v", key, cache.stored)
		}

		// Re-open same file
		doc2, err := engine.OpenFile(filePath)
		if err != nil {
			t.Fatalf("OpenFile (re-open) failed: %v", err)
		}
		if doc2 != doc {
			t.Errorf("Expected same document instance when re-opening")
		}
	})

	t.Run("OpenFile - Error", func(t *testing.T) {
		_, err := engine.OpenFile(filepath.Join(tempDir, "does-not-exist.txt"))
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("SetActiveDocument", func(t *testing.T) {
		filePath1 := filepath.Join(tempDir, "file1.txt")
		filePath2 := filepath.Join(tempDir, "file2.md")

		doc2, err := engine.OpenFile(filePath2)
		if err != nil {
			t.Fatalf("OpenFile failed: %v", err)
		}

		if engine.GetActiveDocument() != doc2 {
			t.Error("Expected file2.md to be active")
		}

		// Set back to file1
		doc1, err := engine.SetActiveDocument(filePath1)
		if err != nil {
			t.Fatalf("SetActiveDocument failed: %v", err)
		}

		if doc1.Path != filePath1 {
			t.Errorf("Expected path %s, got %s", filePath1, doc1.Path)
		}

		if engine.GetActiveDocument() != doc1 {
			t.Error("Expected file1.txt to be active")
		}
	})

	t.Run("SetActiveDocument - Error", func(t *testing.T) {
		_, err := engine.SetActiveDocument("not-open.txt")
		if err == nil {
			t.Error("Expected error for not open document")
		}
	})

	t.Run("CloseFile", func(t *testing.T) {
		filePath1 := filepath.Join(tempDir, "file1.txt")

		err := engine.CloseFile(filePath1)
		if err != nil {
			t.Fatalf("CloseFile failed: %v", err)
		}

		if engine.GetActiveDocument() != nil {
			t.Error("Expected active document to be nil after closing it")
		}

		err = engine.CloseFile("not-open.txt")
		if err == nil {
			t.Error("Expected error when closing non-open document")
		}
	})
}
