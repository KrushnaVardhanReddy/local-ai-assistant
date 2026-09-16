package driving

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFileNode_JSON(t *testing.T) {
	node := FileNode{
		ID:        "node-1",
		Name:      "test.txt",
		Path:      "/tmp/test.txt",
		IsDir:     false,
		Extension: ".txt",
		Size:      1024,
		Children: []*FileNode{
			{ID: "node-2", Name: "child.txt", Path: "/tmp/child.txt", IsDir: false, Extension: ".txt", Size: 256},
		},
	}

	data, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Failed to marshal FileNode: %v", err)
	}

	var parsedNode FileNode
	if err := json.Unmarshal(data, &parsedNode); err != nil {
		t.Fatalf("Failed to unmarshal FileNode: %v", err)
	}

	if !reflect.DeepEqual(node, parsedNode) {
		t.Errorf("Unmarshaled FileNode doesn't match original. Got %+v, want %+v", parsedNode, node)
	}
}

func TestWorkspaceDocument_JSON(t *testing.T) {
	doc := WorkspaceDocument{
		ID:       "doc-1",
		Path:     "/tmp/doc1.txt",
		Name:     "doc1.txt",
		Content:  "Hello world",
		LoadedAt: 1234567890,
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Failed to marshal WorkspaceDocument: %v", err)
	}

	var parsedDoc WorkspaceDocument
	if err := json.Unmarshal(data, &parsedDoc); err != nil {
		t.Fatalf("Failed to unmarshal WorkspaceDocument: %v", err)
	}

	if !reflect.DeepEqual(doc, parsedDoc) {
		t.Errorf("Unmarshaled WorkspaceDocument doesn't match original. Got %+v, want %+v", parsedDoc, doc)
	}
}

// mockWorkspace implements WorkspacePort for testing interface compliance
type mockWorkspace struct{}

func (m *mockWorkspace) OpenDirectory(dirPath string) (*FileNode, error) {
	return nil, nil
}

func (m *mockWorkspace) OpenFile(filePath string) (*WorkspaceDocument, error) {
	return nil, nil
}

func (m *mockWorkspace) GetWorkspaceTree() []*FileNode {
	return nil
}

func (m *mockWorkspace) GetActiveDocument() *WorkspaceDocument {
	return nil
}

func (m *mockWorkspace) SetActiveDocument(path string) (*WorkspaceDocument, error) {
	return nil, nil
}

func (m *mockWorkspace) CloseFile(path string) error {
	return nil
}

func TestWorkspacePort_Compliance(t *testing.T) {
	// Compile-time check
	var _ WorkspacePort = (*mockWorkspace)(nil)

	m := &mockWorkspace{}
	if m == nil {
		t.Errorf("Failed to instantiate mockWorkspace")
	}
}
