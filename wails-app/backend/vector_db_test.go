package backend

import (
	"fmt"
	"testing"
)

func TestVectorDBIntegration(t *testing.T) {
	db, err := NewVectorDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}
	defer db.Close()

	emb1 := make([]float32, 384)
	emb1[0] = 1.0

	emb2 := make([]float32, 384)
	emb2[0] = 0.5
	emb2[1] = 0.5

	err = db.InsertVector("doc1", emb1, `{"title":"Document 1"}`)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	err = db.InsertVector("doc2", emb2, `{"title":"Document 2"}`)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	query := make([]float32, 384)
	query[0] = 1.0
	results, err := db.SearchVector(query, 2)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	fmt.Printf("Top match: %s with distance %f\n", results[0].ID, results[0].Distance)
	fmt.Printf("Second match: %s with distance %f\n", results[1].ID, results[1].Distance)

	if results[0].ID != "doc1" {
		t.Errorf("Expected top match to be doc1, got %s", results[0].ID)
	}
}
