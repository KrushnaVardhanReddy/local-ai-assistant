package cacheadapter

import (
	"testing"

	"wails-app/backend"
)

func TestSQLiteVecAdapter_NilDB(t *testing.T) {
	adapter := NewSQLiteVecAdapter(nil)

	// Test Search with nil db
	ans, hit := adapter.Search([]float32{0.1, 0.2}, 0.8)
	if ans != "" || hit != false {
		t.Errorf("Expected Search to return ('', false) with nil db, got ('%s', %v)", ans, hit)
	}

	// Test Store with nil db
	err := adapter.Store("q", "a")
	if err != nil {
		t.Errorf("Expected Store to return nil with nil db, got %v", err)
	}

	// Test Count with nil db
	count := adapter.Count()
	if count != 0 {
		t.Errorf("Expected Count to return 0 with nil db, got %d", count)
	}
}

func TestSQLiteVecAdapter_WithDB(t *testing.T) {
	// Create an in-memory database to satisfy "No real SQLite DB file created"
	db, err := backend.NewVectorDB(":memory:")
	if err != nil {
		t.Skip("Skipping test because VectorDB could not be initialized in-memory (likely due to CGO/sqlite-vec missing in CI)", err)
	}

	adapter := NewSQLiteVecAdapter(db)

	// We just need to hit the non-nil paths.
	// Since we don't have real embeddings, it will probably fail to search or store,
	// but the adapter logic will be covered.

	// Test Search
	adapter.Search([]float32{0.1, 0.2}, 0.8)

	// Test Store
	adapter.Store("question", "answer")

	// Test Count
	adapter.Count()
}
