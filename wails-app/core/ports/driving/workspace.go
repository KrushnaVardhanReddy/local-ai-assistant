package driving

type FileNode struct {
	ID        string      `json:"id"`        // Unique clean path or UUID
	Name      string      `json:"name"`      // Display name
	Path      string      `json:"path"`      // Absolute path on disk
	IsDir     bool        `json:"isDir"`     // True if folder
	Extension string      `json:"extension"` // e.g. ".md", ".pdf", ".txt", ".pptx"
	Size      int64       `json:"size"`      // File size in bytes
	Children  []*FileNode `json:"children,omitempty"` // Child nodes if folder
}

type WorkspaceDocument struct {
	ID       string `json:"id"`
	Path     string `json:"path"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	LoadedAt int64  `json:"loadedAt"`
}

type WorkspacePort interface {
	// OpenDirectory reads a directory from disk and returns the hierarchical FileNode tree.
	// Only supported document types (.txt, .md, .pdf, .pptx) and subfolders are indexed.
	OpenDirectory(dirPath string) (*FileNode, error)

	// OpenFile parses a specific file and returns a WorkspaceDocument.
	OpenFile(filePath string) (*WorkspaceDocument, error)

	// GetWorkspaceTree returns the current root file nodes.
	GetWorkspaceTree() []*FileNode

	// GetActiveDocument returns the currently active document.
	GetActiveDocument() *WorkspaceDocument

	// SetActiveDocument sets the active document by ID/Path.
	SetActiveDocument(path string) (*WorkspaceDocument, error)

	// CloseFile closes an open document.
	CloseFile(path string) error
}
