package cacheadapter

import (
	"wails-app/backend"
	"wails-app/core/ports/driven"
)

// SQLiteVecAdapter wraps backend.VectorDB behind driven.CachePort.
type SQLiteVecAdapter struct {
	db *backend.VectorDB
}

// NewSQLiteVecAdapter creates a new SQLiteVecAdapter instance.
func NewSQLiteVecAdapter(db *backend.VectorDB) *SQLiteVecAdapter {
	return &SQLiteVecAdapter{db: db}
}

// Search calls db.SearchByEmbedding with the given threshold.
// Returns the cached answer and true if similarity >= threshold.
func (a *SQLiteVecAdapter) Search(embedding []float32, threshold float64) (string, bool) {
	if a.db == nil {
		return "", false
	}
	return a.db.SearchByEmbedding(embedding, float32(threshold))
}

// Store calls db.Store to save a question-answer pair.
func (a *SQLiteVecAdapter) Store(question, answer string) error {
	if a.db == nil {
		return nil
	}
	return a.db.Store(question, answer)
}

// Count returns the number of cached Q&A pairs via db.GetCount().
func (a *SQLiteVecAdapter) Count() int {
	if a.db == nil {
		return 0
	}
	return a.db.GetCount()
}

// GetAllItems returns all cached Q&A pairs via db.GetAllItems().
func (a *SQLiteVecAdapter) GetAllItems() ([]backend.CacheItem, error) {
	if a.db == nil {
		return nil, nil
	}
	return a.db.GetAllItems()
}

// ClearAll removes all cached Q&A pairs
func (a *SQLiteVecAdapter) ClearAll() error {
	if a.db == nil {
		return nil
	}
	return a.db.ClearAll()
}

// DeleteItem removes a single item from the cache
func (a *SQLiteVecAdapter) DeleteItem(id string) error {
	if a.db == nil {
		return nil
	}
	return a.db.DeleteItem(id)
}

// IndexDocumentChunk inserts a document chunk into the vector database.
func (a *SQLiteVecAdapter) IndexDocumentChunk(path, text string, embedding []float32) error {
	if a.db == nil {
		return nil
	}
	return a.db.InsertDocumentChunk(path, text, embedding)
}

// GetIndexedPaths returns all paths that have been indexed.
func (a *SQLiteVecAdapter) GetIndexedPaths() ([]string, error) {
	if a.db == nil {
		return nil, nil
	}
	return a.db.GetIndexedPaths()
}

// RemoveIndexedPath removes all document chunks for a given path.
func (a *SQLiteVecAdapter) RemoveIndexedPath(path string) error {
	if a.db == nil {
		return nil
	}
	return a.db.DeleteDocumentsByPath(path)
}

// SemanticSearch queries BOTH 'qa' and 'document' rows, returning the best matches regardless of type.
func (a *SQLiteVecAdapter) SemanticSearch(embedding []float32, limit int, threshold float64) ([]string, error) {
	if a.db == nil {
		return nil, nil
	}
	results, err := a.db.SemanticSearch(embedding, limit, float32(threshold))
	if err != nil {
		return nil, err
	}
	var chunks []string
	for _, res := range results {
		// Only return document chunks, as QA cache hits are handled separately
		if res.SourceType == "document" {
			chunks = append(chunks, res.Metadata)
		}
	}
	return chunks, nil
}

// Compile-time interface check.
var _ driven.CachePort = (*SQLiteVecAdapter)(nil)
