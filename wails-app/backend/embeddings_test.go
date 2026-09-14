package backend

import (
	"testing"
)

func TestGenerateEmbedding(t *testing.T) {
	text := "Hello world!"
	embedding := GenerateEmbedding(text)

	if embedding == nil {
		t.Fatalf("Expected embedding to be not nil")
	}

	if len(embedding) != 384 {
		t.Fatalf("Expected embedding length to be 384, got %d", len(embedding))
	}
}
