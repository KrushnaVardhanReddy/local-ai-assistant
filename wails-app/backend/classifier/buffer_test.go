package classifier

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestQuestionBuffer_FlushesOnTrueClassification(t *testing.T) {
	var flushedText string
	var flushCount int
	var mu sync.Mutex

	onFlush := func(text string) {
		mu.Lock()
		defer mu.Unlock()
		flushedText = text
		flushCount++
	}

	b := NewQuestionBuffer(onFlush)
	b.maxAge = 45 * time.Second
	b.watchdogInterval = 5 * time.Second
	b.Start()
	defer b.Stop()

	// Use a lock in the mock to delay return until we added all chunks
	var startClassify sync.WaitGroup
	startClassify.Add(1)

	b.classifyFn = func(ctx context.Context, text string) (bool, error) {
		startClassify.Wait() // wait until all chunks are added
		return true, nil
	}

	b.AddChunk("Tell me", false)
	b.AddChunk("about Redis", false)
	b.AddChunk("caching please.", false)

	startClassify.Done()

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if flushCount != 1 {
		t.Fatalf("expected 1 flush, got %d", flushCount)
	}

	expected := "Tell me about Redis caching please."
	if flushedText != expected {
		t.Errorf("expected flushed text %q, got %q", expected, flushedText)
	}

	b.mu.Lock()
	if len(b.chunks) != 0 {
		t.Errorf("expected buffer to be empty, got %d chunks", len(b.chunks))
	}
	b.mu.Unlock()
}

func TestQuestionBuffer_DoesNotFlushOnFalse(t *testing.T) {
	var flushCount int
	var mu sync.Mutex

	onFlush := func(text string) {
		mu.Lock()
		defer mu.Unlock()
		flushCount++
	}

	b := NewQuestionBuffer(onFlush)

	b.maxAge = 45 * time.Second
	b.watchdogInterval = 5 * time.Second
	b.Start()
	defer b.Stop()

	b.classifyFn = func(ctx context.Context, text string) (bool, error) {
		return false, nil
	}

	b.AddChunk("chunk1", false)
	b.AddChunk("chunk2", false)
	b.AddChunk("chunk3", false)

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if flushCount != 0 {
		t.Fatalf("expected 0 flush, got %d", flushCount)
	}

	b.mu.Lock()
	if len(b.chunks) != 3 {
		t.Errorf("expected buffer to have 3 chunks, got %d", len(b.chunks))
	}
	b.mu.Unlock()
}

func TestQuestionBuffer_MinChunksGuard(t *testing.T) {
	var flushCount int
	var classifyCount int
	var mu sync.Mutex

	onFlush := func(text string) {
		mu.Lock()
		defer mu.Unlock()
		flushCount++
	}

	b := NewQuestionBuffer(onFlush)

	b.minChunks = 3
	b.maxAge = 45 * time.Second
	b.watchdogInterval = 5 * time.Second
	b.Start()
	defer b.Stop()

	b.classifyFn = func(ctx context.Context, text string) (bool, error) {
		mu.Lock()
		classifyCount++
		mu.Unlock()
		return true, nil
	}

	b.AddChunk("chunk1", false)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if classifyCount != 0 {
		t.Errorf("expected 0 classification calls, got %d", classifyCount)
	}
	if flushCount != 0 {
		t.Errorf("expected 0 flush calls, got %d", flushCount)
	}
}

func TestQuestionBuffer_FailsafeWatchdog(t *testing.T) {
	var flushCount int
	var mu sync.Mutex

	onFlush := func(text string) {
		mu.Lock()
		defer mu.Unlock()
		flushCount++
	}

	b := NewQuestionBuffer(onFlush)

	b.maxAge = 100 * time.Millisecond
	b.watchdogInterval = 50 * time.Millisecond // check more often in tests
	b.Start()
	defer b.Stop()

	b.classifyFn = func(ctx context.Context, text string) (bool, error) {
		return false, nil
	}

	b.AddChunk("chunk1", false)
	b.AddChunk("chunk2", false)
	b.AddChunk("chunk3", false)

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if flushCount == 0 {
		t.Fatal("expected flush by watchdog, got 0")
	}

	b.mu.Lock()
	if len(b.chunks) != 0 {
		t.Errorf("expected empty buffer, got %d", len(b.chunks))
	}
	b.mu.Unlock()
}

func TestQuestionBuffer_Reset(t *testing.T) {
	var flushCount int
	var mu sync.Mutex

	onFlush := func(text string) {
		mu.Lock()
		defer mu.Unlock()
		flushCount++
	}

	b := NewQuestionBuffer(onFlush)
	b.maxAge = 45 * time.Second
	b.watchdogInterval = 5 * time.Second
	b.Start()
	defer b.Stop()

	b.classifyFn = func(ctx context.Context, text string) (bool, error) {
		return false, nil
	}

	for i := 0; i < 5; i++ {
		b.AddChunk(fmt.Sprintf("chunk%d", i), false)
	}

	b.Reset()

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if flushCount != 0 {
		t.Errorf("expected 0 flush calls, got %d", flushCount)
	}

	b.mu.Lock()
	if len(b.chunks) != 0 {
		t.Errorf("expected empty buffer after reset, got %d", len(b.chunks))
	}
	b.mu.Unlock()
}

func TestQuestionBuffer_ClassifyError(t *testing.T) {
	var flushCount int
	var mu sync.Mutex

	onFlush := func(text string) {
		mu.Lock()
		defer mu.Unlock()
		flushCount++
	}

	b := NewQuestionBuffer(onFlush)
	b.maxAge = 45 * time.Second
	b.watchdogInterval = 5 * time.Second
	b.Start()
	defer b.Stop()

	b.classifyFn = func(ctx context.Context, text string) (bool, error) {
		return false, fmt.Errorf("gemma down")
	}

	b.AddChunk("chunk1", false)
	b.AddChunk("chunk2", false)
	b.AddChunk("chunk3", false)

	// The error in classifyFn should not crash the app, and should not flush.
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if flushCount != 0 {
		t.Errorf("expected 0 flush calls, got %d", flushCount)
	}
}
