package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"wails-app/backend/parser"
	"wails-app/core/ports/driving"
)

// allowedExtensions for document indexing
var allowedExtensions = map[string]bool{
	".txt":    true,
	".md":     true,
	".pdf":    true,
	".pptx":   true,
	".go":     true,
	".py":     true,
	".js":     true,
	".ts":     true,
	".jsx":    true,
	".tsx":    true,
	".svelte": true,
	".json":   true,
	".rs":     true,
	".cpp":    true,
	".c":      true,
	".h":      true,
	".html":   true,
	".css":    true,
	".yaml":   true,
	".yml":    true,
	".sql":    true,
	".sh":     true,
}

// isIgnored directory or file
func isIgnored(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true // ignores .git, .DS_Store, hidden files
	}
	if name == "node_modules" {
		return true
	}
	return false
}

// OpenDirectory recursively scans the dirPath to build the FileNode tree.
// It filters for allowed extensions and valid directories.
func (e *StealthEngine) OpenDirectory(dirPath string) (*driving.FileNode, error) {
	node, _, err := buildFileTree(dirPath)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, fmt.Errorf("directory %s does not contain valid document files", dirPath)
	}

	e.workspaceMu.Lock()
	if e.workspaceTree == nil {
		e.workspaceTree = make([]*driving.FileNode, 0)
	}
	// We replace the workspace tree with this new root (or append, but typical is open a workspace)
	e.workspaceTree = []*driving.FileNode{node}
	e.workspaceMu.Unlock()

	if e.events != nil {
		e.events.Emit("on_workspace_tree_updated", e.workspaceTree)
	}

	return node, nil
}

// buildFileTree recursively builds the FileNode tree.
// Returns the node (if it or any children are valid), a boolean indicating if it contains valid files, and error.
func buildFileTree(path string) (*driving.FileNode, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, false, err
	}

	if isIgnored(info.Name()) {
		return nil, false, nil
	}

	node := &driving.FileNode{
		ID:        path, // using path as simple unique ID
		Name:      info.Name(),
		Path:      path,
		IsDir:     info.IsDir(),
		Extension: strings.ToLower(filepath.Ext(info.Name())),
		Size:      info.Size(),
	}

	if !info.IsDir() {
		// It's a file
		if allowedExtensions[node.Extension] {
			return node, true, nil
		}
		return nil, false, nil
	}

	// It's a directory, scan children
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, false, err
	}

	var children []*driving.FileNode
	hasValidFiles := false

	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		childNode, valid, err := buildFileTree(childPath)
		if err != nil {
			continue // skip errors on children (like permission denied)
		}
		if valid && childNode != nil {
			children = append(children, childNode)
			hasValidFiles = true
		}
	}

	if !hasValidFiles {
		return nil, false, nil // don't include empty dirs or dirs without valid files
	}

	// Sort children: folders first, then files alphabetically
	sort.Slice(children, func(i, j int) bool {
		if children[i].IsDir && !children[j].IsDir {
			return true
		}
		if !children[i].IsDir && children[j].IsDir {
			return false
		}
		return strings.Compare(strings.ToLower(children[i].Name), strings.ToLower(children[j].Name)) < 0
	})

	node.Children = children

	return node, true, nil
}

// OpenFile parses a specific file and returns a WorkspaceDocument.
func (e *StealthEngine) OpenFile(filePath string) (*driving.WorkspaceDocument, error) {
	e.workspaceMu.Lock()
	_, exists := e.openDocuments[filePath]
	e.workspaceMu.Unlock()

	if exists {
		// Already open, just set as active
		return e.SetActiveDocument(filePath)
	}

	content, err := parser.ParseDocument(filePath)
	if err != nil {
		return nil, err
	}

	newDoc := &driving.WorkspaceDocument{
		ID:       filePath,
		Path:     filePath,
		Name:     filepath.Base(filePath),
		Content:  content,
		LoadedAt: time.Now().Unix(),
	}

	e.workspaceMu.Lock()
	e.openDocuments[filePath] = newDoc
	e.activeDoc = newDoc

	// Check if file is already in workspace tree
	found := false
	var walk func(nodes []*driving.FileNode)
	walk = func(nodes []*driving.FileNode) {
		for _, node := range nodes {
			if node.Path == filePath {
				found = true
				return
			}
			if len(node.Children) > 0 {
				walk(node.Children)
			}
		}
	}
	walk(e.workspaceTree)

	var newlyAdded bool
	if !found {
		// Not found in tree, create a new node and append
		info, err := os.Stat(filePath)
		if err == nil {
			node := &driving.FileNode{
				ID:        filePath,
				Name:      info.Name(),
				Path:      filePath,
				IsDir:     false,
				Extension: strings.ToLower(filepath.Ext(info.Name())),
				Size:      info.Size(),
			}
			e.workspaceTree = append(e.workspaceTree, node)
			newlyAdded = true
		}
	}

	e.workspaceMu.Unlock()

	if newlyAdded && e.events != nil {
		e.events.Emit("on_workspace_tree_updated", e.workspaceTree)
	}

	// If cache is enabled, we could potentially index it here.
	// We'll store it as a single chunk for simplicity in this implementation,
	// or just let it be retrieved. The prompt mentioned "cache document contents in the engine's memory/vector cache"
	if e.cache != nil {
		// Cache as Q&A pair where Q is the filename and A is the content, or just store it.
		e.cache.Store(fmt.Sprintf("Document: %s", newDoc.Name), content)
	}

	if e.events != nil {
		e.events.Emit("on_active_document_changed", newDoc)
	}

	return newDoc, nil
}

// GetWorkspaceTree returns the current root file nodes.
func (e *StealthEngine) GetWorkspaceTree() []*driving.FileNode {
	e.workspaceMu.RLock()
	defer e.workspaceMu.RUnlock()
	return e.workspaceTree
}

// GetActiveDocument returns the currently active document.
func (e *StealthEngine) GetActiveDocument() *driving.WorkspaceDocument {
	e.workspaceMu.RLock()
	defer e.workspaceMu.RUnlock()
	return e.activeDoc
}

// GetOpenDocuments returns all currently open documents.
func (e *StealthEngine) GetOpenDocuments() []*driving.WorkspaceDocument {
	e.workspaceMu.RLock()
	defer e.workspaceMu.RUnlock()

	docs := make([]*driving.WorkspaceDocument, 0, len(e.openDocuments))
	for _, doc := range e.openDocuments {
		docs = append(docs, doc)
	}

	// Sort tabs alphabetically by Path to ensure deterministic UI rendering order
	sort.Slice(docs, func(i, j int) bool {
		return strings.Compare(strings.ToLower(docs[i].Path), strings.ToLower(docs[j].Path)) < 0
	})

	return docs
}

// SetActiveDocument sets the active document by ID/Path.
func (e *StealthEngine) SetActiveDocument(path string) (*driving.WorkspaceDocument, error) {
	e.workspaceMu.Lock()
	doc, exists := e.openDocuments[path]
	if !exists {
		e.workspaceMu.Unlock()
		return nil, errors.New("document not open")
	}
	e.activeDoc = doc
	e.workspaceMu.Unlock()

	if e.events != nil {
		e.events.Emit("on_active_document_changed", doc)
	}

	return doc, nil
}

// CloseFile closes an open document.
func (e *StealthEngine) CloseFile(path string) error {
	e.workspaceMu.Lock()
	defer e.workspaceMu.Unlock()

	if _, exists := e.openDocuments[path]; !exists {
		return errors.New("document not found")
	}

	delete(e.openDocuments, path)

	if e.activeDoc != nil && e.activeDoc.Path == path {
		e.activeDoc = nil
	}

	return nil
}

// Compile-time check
var _ driving.WorkspacePort = (*StealthEngine)(nil)
