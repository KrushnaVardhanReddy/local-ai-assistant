package filter

import (
	"os"
	"testing"
	"sync"
)

func init() {
	// Mock GenerateEmbedding for all tests to avoid tokenizer panics
	GenerateEmbedding = func(text string) []float32 {
		emb := make([]float32, 384)
		for i := 0; i < 384; i++ {
			emb[i] = 1.0 // identical to our mock centroid
		}
		return emb
	}
}

func TestInitNoiseCentroidCoverage(t *testing.T) {
	// Reset the sync.Once
	once = sync.Once{}
	noiseCentroid = nil

	initNoiseCentroid()

	if noiseCentroid == nil {
		t.Errorf("Expected noiseCentroid to be initialized")
	}

	// Test nil returned by GenerateEmbedding (as if tokenizer failed)
	once = sync.Once{}
	noiseCentroid = nil
	GenerateEmbedding = func(text string) []float32 { return nil }
	initNoiseCentroid()
	if noiseCentroid != nil {
		t.Errorf("Expected nil noiseCentroid")
	}

	// Restore mock
	GenerateEmbedding = func(text string) []float32 {
		emb := make([]float32, 384)
		for i := 0; i < 384; i++ {
			emb[i] = 1.0
		}
		return emb
	}
}

func TestCheck(t *testing.T) {
	os.Setenv("MIN_WORDS", "3")

	tests := []struct {
		name      string
		text      string
		embedding []float32
		want      bool
		reason    string
	}{
		{
			name:      "too short",
			text:      "Yes sure",
			embedding: nil,
			want:      false,
			reason:    "too_short",
		},
		{
			name:      "filler (okay) with punct",
			text:      "Okay...",
			embedding: nil,
			want:      false,
			reason:    "too_short",
		},
		{
			name:      "filler (okay) forced enough words",
			text:      "Okay Okay Okay",
			embedding: nil,
			want:      false,
			reason:    "filler",
		},
		{
			name:      "accepted",
			text:      "What are the features of Golang?",
			embedding: make([]float32, 384), // 0s, won't match noise
			want:      true,
			reason:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "filler (okay) forced enough words" {
				os.Setenv("MIN_WORDS", "1")
				tt.text = "Okay..."
			}

			res := Check(tt.text, tt.embedding)
			if res.ShouldSend != tt.want {
				t.Errorf("Check() ShouldSend = %v, want %v", res.ShouldSend, tt.want)
			}
			if res.Reason != tt.reason {
				t.Errorf("Check() Reason = %v, want %v", res.Reason, tt.reason)
			}
		})
	}
	os.Setenv("MIN_WORDS", "3")
}

func TestCheckNoiseCentroid(t *testing.T) {
	os.Setenv("MIN_WORDS", "3")

	once = sync.Once{}
	noiseCentroid = nil

	emb := make([]float32, 384)
	for i := 0; i < 384; i++ {
		emb[i] = 1.0
	}
	res := Check("This is some text that should not be dropped by similarity but who knows", emb)
	if res.Reason != "noise" {
		t.Errorf("Expected dropped as noise, got %v", res.Reason)
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{1, 0, 0}
	if cosineSimilarity(a, b) != 1.0 {
		t.Errorf("expected 1.0")
	}

	c := []float32{0, 1, 0}
	if cosineSimilarity(a, c) != 0 {
		t.Errorf("expected 0")
	}

	if cosineSimilarity([]float32{}, []float32{}) != 0 {
		t.Errorf("expected 0")
	}

	if cosineSimilarity([]float32{0}, []float32{0}) != 0 {
		t.Errorf("expected 0")
	}
}

func TestMinWordsEnv(t *testing.T) {
	os.Setenv("MIN_WORDS", "invalid")
	res := Check("One two", nil)
	if res.Reason != "too_short" {
		t.Errorf("expected too short for invalid min words")
	}

	os.Setenv("MIN_WORDS", "1")
	emb := make([]float32, 1)
	res = Check("Valid", emb)
	if !res.ShouldSend {
		t.Errorf("expected accepted")
	}
}
