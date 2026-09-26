package filter

import (
	_ "embed"
	"encoding/json"
	"log"
	"sync"
	"wails-app/backend"
)

var (
	ClassifierGenerateEmbedding = backend.GenerateEmbedding
	centroids                   map[string][]float32
	centroidsOnce               sync.Once
)

//go:embed centroids.json
var centroidsJSON []byte

func initCentroids() {
	centroids = make(map[string][]float32)
	if err := json.Unmarshal(centroidsJSON, &centroids); err != nil {
		log.Printf("[FILTER] Failed to unmarshal centroids.json: %v", err)
	}
}

func Classify(embedding []float32) string {
	centroidsOnce.Do(initCentroids)

	if len(embedding) != 768 || len(centroids) == 0 {
		return "unknown"
	}

	bestCategory := "unknown"
	var bestScore float32 = -1.0

	for category, centroid := range centroids {
		score := cosineSimilarity(embedding, centroid)
		if score > bestScore {
			bestScore = score
			bestCategory = category
		}
	}

	return bestCategory
}
