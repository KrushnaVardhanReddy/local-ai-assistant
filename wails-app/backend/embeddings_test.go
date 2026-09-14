package backend

import (
	"testing"
)

func TestGenerateEmbedding(t *testing.T) {
	// If tokenizer is missing we cannot run this test natively in CI/tests without pulling model weights.
	// Since initEmbeddings logs an error instead of panicking on failure, tk could be nil.
	initOnce.Do(InitEmbeddings)
	if tk == nil {
		t.Skip("Tokenizer is nil, skipping TestGenerateEmbedding because model assets are unavailable")
	}

	text := "Hello world!"
	embedding := GenerateEmbedding(text)

	if embedding == nil {
		t.Fatalf("Expected embedding to be not nil")
	}

	if len(embedding) != 384 {
		t.Fatalf("Expected embedding length to be 384, got %d", len(embedding))
	}
}
