package driven

// CachePort is the driven port for the semantic QA cache (vector DB).
type CachePort interface {
	// Search finds a cached answer by semantic similarity of the embedding.
	// threshold is the minimum cosine similarity (e.g. 0.88).
	// Returns the cached answer and true if a hit is found.
	Search(embedding []float32, threshold float64) (answer string, hit bool)

	// Store saves a question-answer pair into the cache.
	Store(question, answer string) error

	// Count returns the number of cached pairs.
	Count() int
}
