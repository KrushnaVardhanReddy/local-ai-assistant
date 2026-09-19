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

// Compile-time interface check.
var _ driven.CachePort = (*SQLiteVecAdapter)(nil)
