package filter

import (
	"sync"
	"testing"
)

func TestInitCentroids(t *testing.T) {
	// Reset the sync.Once and state
	centroidsOnce = sync.Once{}
	centroids = nil

	// Mock embedding to return vector of 1s
	ClassifierGenerateEmbedding = func(text string) []float32 {
		emb := make([]float32, 768)
		for i := 0; i < 768; i++ {
			emb[i] = 1.0
		}
		return emb
	}

	initCentroids()

	if len(centroids) == 0 {
		t.Errorf("Expected centroids to be initialized")
	}

	if _, ok := centroids["noise"]; !ok {
		t.Errorf("Expected 'noise' centroid to exist")
	}

	if _, ok := centroids["behavioral"]; !ok {
		t.Errorf("Expected 'behavioral' centroid to exist")
	}

	// Test case where embedding generation fails (returns nil or empty)
	centroidsOnce = sync.Once{}
	centroids = nil
	ClassifierGenerateEmbedding = func(text string) []float32 {
		return nil
	}

	initCentroids()

	if len(centroids) != 0 {
		t.Errorf("Expected centroids to be empty since embeddings failed")
	}
}

func TestClassify(t *testing.T) {
	// Reset the sync.Once and state
	centroidsOnce = sync.Once{}
	centroids = nil

	// Mock embedding to return a unique vector per category
	// For testing Classify, we will just manually set the centroids map directly
	// instead of relying on initCentroids to generate them.

	// First, let initCentroids run with a dummy embedding so it doesn't crash
	ClassifierGenerateEmbedding = func(text string) []float32 {
		emb := make([]float32, 768)
		return emb
	}

	// Force the once to trigger so we can override centroids
	Classify([]float32{})

	// Now set up our controlled centroids for testing similarity
	centroids = map[string][]float32{
		"catA": make([]float32, 768),
		"catB": make([]float32, 768),
	}

	for i := 0; i < 768; i++ {
		centroids["catA"][i] = 1.0 // vector of 1s
		centroids["catB"][i] = 0.5 // vector of 0.5s
	}

	tests := []struct {
		name      string
		embedding []float32
		want      string
	}{
		{
			name:      "wrong length embedding",
			embedding: []float32{1.0, 2.0},
			want:      "unknown",
		},
		{
			name:      "matches catA",
			embedding: centroids["catA"],
			want:      "catA", // Cosine similarity will be 1.0 with catA
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.embedding)
			if got != tt.want {
				t.Errorf("Classify() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test with nil centroids map
	centroids = nil
	if got := Classify(make([]float32, 768)); got != "unknown" {
		t.Errorf("Expected 'unknown' for nil centroids, got %v", got)
	}
}
