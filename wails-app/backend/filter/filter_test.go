package filter

import (
	"os"
	"sync"
	"testing"
)

func TestCheck(t *testing.T) {
	os.Setenv("MIN_WORDS", "3")

	// Set up mock centroids for Check
	centroidsOnce = sync.Once{}
	centroids = map[string][]float32{
		"noise": make([]float32, 768),
	}
	for i := 0; i < 768; i++ {
		centroids["noise"][i] = 1.0
	}
	// Bypass initCentroids replacing it
	centroidsOnce.Do(func() {})

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
			embedding: make([]float32, 768),
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
			if tt.name == "accepted" {
				// make the embedding completely orthogonal to noise centroid to not get classified as noise
				tt.embedding[0] = 1.0 // noise is all 1.0, this is just {1,0,0...}
				// wait, if noise is all 1.0, any positive number will have some cosine similarity
				// to get orthogonal, we can make part of it negative so dot product is zero, or just use another category
				// we will just set bestScore to something that doesn't trigger "noise".
				// Actually, we can add a dummy centroid for "behavioral" so it matches that instead of "noise".
				centroids["behavioral"] = make([]float32, 768)
				for i := 0; i < 768; i++ {
					centroids["behavioral"][i] = 1.0
					tt.embedding[i] = 1.0
				}
				// if we have two identical, it's non-deterministic which one it picks if they both have 1.0 score.
				// let's explicitly make it match behavioral better.
				for i := 0; i < 768; i++ {
					centroids["noise"][i] = -1.0
				}
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

	// Set up mock centroids for Check
	centroidsOnce = sync.Once{}
	centroids = map[string][]float32{
		"noise": make([]float32, 768),
	}
	for i := 0; i < 768; i++ {
		centroids["noise"][i] = 1.0
	}
	// Bypass initCentroids replacing it
	centroidsOnce.Do(func() {})

	emb := make([]float32, 768)
	for i := 0; i < 768; i++ {
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
