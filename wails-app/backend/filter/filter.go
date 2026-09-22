package filter

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"wails-app/backend/config"
)

var (
	pkgConfig   *config.AppConfig
	pkgConfigMu sync.RWMutex
)

// SetConfig stores the application config for use by the filter.
func SetConfig(cfg *config.AppConfig) {
	pkgConfigMu.Lock()
	defer pkgConfigMu.Unlock()
	pkgConfig = cfg
}

func getCfg() *config.AppConfig {
	pkgConfigMu.RLock()
	defer pkgConfigMu.RUnlock()
	return pkgConfig
}

type FilterResult struct {
	ShouldSend bool
	Reason     string
}

var (
	fillerPhrases = map[string]bool{
		"okay": true, "ok": true, "yeah": true, "yes": true, "no": true,
		"mmhmm": true, "hmm": true, "uh": true, "um": true, "right": true,
		"sure": true, "thanks": true, "thank you": true, "bye": true,
		"alright": true, "all right": true, "got it": true, "i see": true,
		"i know": true, "oh": true, "ah": true, "so": true, "yep": true,
		"nope": true, "hi": true, "hello": true, "hey": true,
		"good morning": true, "good afternoon": true, "good evening": true,
		"how are you": true, "hows it going": true, "whats up": true, "sup": true,
	}
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
)

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct, normA, normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / float32(math.Sqrt(float64(normA))*math.Sqrt(float64(normB)))
}

func countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

func cleanText(text string) string {
	return strings.TrimSpace(nonAlphanumericRegex.ReplaceAllString(strings.ToLower(text), ""))
}

func Check(text string, embedding []float32) FilterResult {
	// Stage 1: MIN_WORDS
	minWords := 3
	if cfg := getCfg(); cfg != nil && cfg.MinWords != "" {
		if val, err := strconv.Atoi(cfg.MinWords); err == nil {
			minWords = val
		}
	}
	if countWords(text) < minWords {
		log.Printf("[FILTER] Dropped (too_short): %q", text)
		return FilterResult{ShouldSend: false, Reason: "too_short"}
	}

	// Stage 2: Filler / Noise phrase
	cleaned := cleanText(text)
	if fillerPhrases[cleaned] {
		log.Printf("[FILTER] Dropped (filler): %q", text)
		return FilterResult{ShouldSend: false, Reason: "filler"}
	}

	// Stage 3: Embedding Similarity
	predictedClass := Classify(embedding)
	if predictedClass == "noise" {
		log.Printf("[FILTER] Dropped (noise): %q", text)
		return FilterResult{ShouldSend: false, Reason: "noise"}
	}

	log.Printf("[FILTER] Accepted: %q", text)
	return FilterResult{ShouldSend: true, Reason: ""}
}
