package classifier

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

type QuestionBuffer struct {
	mu          sync.Mutex
	chunks      []string
	lastChunkAt time.Time
	maxAge      time.Duration
	minChunks   int
	onFlush     func(string)
	classifyFn  func(ctx context.Context, text string) (bool, error)

	stopCh           chan struct{}
	watchdogInterval time.Duration
}

func NewQuestionBuffer(onFlush func(string)) *QuestionBuffer {
	b := &QuestionBuffer{
		chunks:           make([]string, 0),
		lastChunkAt:      time.Now(),
		maxAge:           45 * time.Second,
		minChunks:        1,
		onFlush:          onFlush,
		classifyFn:       IsQuestionComplete,
		stopCh:           make(chan struct{}),
		watchdogInterval: 5 * time.Second,
	}

	return b
}

// Start starts the internal watchdog goroutine. It must be called after configuring the buffer (e.g., in tests).
func (b *QuestionBuffer) Start() {
	go b.watchdog()
}

func (b *QuestionBuffer) watchdog() {
	ticker := time.NewTicker(b.watchdogInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ticker.C:
			b.mu.Lock()
			shouldFlush := len(b.chunks) > 0 && time.Since(b.lastChunkAt) > b.maxAge
			b.mu.Unlock()

			if shouldFlush {
				b.flush()
			}
		}
	}
}

func (b *QuestionBuffer) AddChunk(chunk string) {
	b.mu.Lock()
	b.chunks = append(b.chunks, chunk)
	b.lastChunkAt = time.Now()

	if len(b.chunks) < b.minChunks {
		b.mu.Unlock()
		return
	}

	fullText := strings.Join(b.chunks, " ")
	b.mu.Unlock()

	go func(text string) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		isComplete, err := b.classifyFn(ctx, text)
		if err != nil {
			log.Printf("[Buffer] Classification error: %v", err)
			return
		}

		if isComplete {
			b.flush()
		}
	}(fullText)
}

func (b *QuestionBuffer) flush() {
	b.mu.Lock()
	if len(b.chunks) == 0 {
		b.mu.Unlock()
		return
	}

	fullText := strings.Join(b.chunks, " ")
	b.chunks = make([]string, 0)
	b.mu.Unlock()

	if b.onFlush != nil {
		b.onFlush(fullText)
	}
}

func (b *QuestionBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.chunks = make([]string, 0)
}

func (b *QuestionBuffer) Stop() {
	close(b.stopCh)
}

func (b *QuestionBuffer) SetClassifier(fn func(ctx context.Context, text string) (bool, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.classifyFn = fn
}

func (b *QuestionBuffer) GetChunks() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	chunks := make([]string, len(b.chunks))
	copy(chunks, b.chunks)
	return chunks
}
