package filter

import (
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"wails-app/backend"
)

type FilterResult struct {
	ShouldSend bool
	Reason     string
}

// Allow mocking GenerateEmbedding in tests
var GenerateEmbedding = backend.GenerateEmbedding

var (
	noiseCentroid []float32
	once          sync.Once
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

func initNoiseCentroid() {
	anchors := []string{
		"Okay.",
		"Sounds good, thank you.",
		"Hmm, let me think about that.",
		"Mhm, sure.",
		"Got it.",
	}
	embeddings := make([][]float32, 0, len(anchors))
	for _, anchor := range anchors {
		emb := GenerateEmbedding(anchor)
		if len(emb) == 384 {
			embeddings = append(embeddings, emb)
		}
	}
	if len(embeddings) > 0 {
		noiseCentroid = make([]float32, 384)
		for _, emb := range embeddings {
			for i := 0; i < 384; i++ {
				noiseCentroid[i] += emb[i]
			}
		}
		for i := 0; i < 384; i++ {
			noiseCentroid[i] /= float32(len(embeddings))
		}
	}
}

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
	minWordsStr := os.Getenv("MIN_WORDS")
	minWords := 3
	if minWordsStr != "" {
		if val, err := strconv.Atoi(minWordsStr); err == nil {
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
	// Bypass embedding noise-gate if the text is obviously a question
	isQuestion := strings.HasSuffix(strings.TrimSpace(text), "?")
	lowerText := strings.ToLower(strings.TrimSpace(text))
	questionWords := []string{"what", "how", "why", "where", "who", "when", "can", "could", "should", "would"}
	for _, word := range questionWords {
		if strings.HasPrefix(lowerText, word+" ") {
			isQuestion = true
			break
		}
	}

	if !isQuestion {
		once.Do(initNoiseCentroid)
		if len(noiseCentroid) == 768 && len(embedding) == 768 {
			similarity := cosineSimilarity(embedding, noiseCentroid)
			if similarity > 0.82 {
				log.Printf("[FILTER] Dropped (noise): %q", text)
				return FilterResult{ShouldSend: false, Reason: "noise"}
			}
		}
	}

	log.Printf("[FILTER] Accepted: %q", text)
	return FilterResult{ShouldSend: true, Reason: ""}
}
