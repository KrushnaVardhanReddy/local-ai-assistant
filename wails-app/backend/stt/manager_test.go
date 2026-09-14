package stt

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// MockSTTEngine is a mock implementation of STTEngine for testing
type MockSTTEngine struct {
	transcribeFunc func(audio []float32) (chan string, error)
	closeFunc      func() error
	closed         bool
}

func (m *MockSTTEngine) TranscribeStream(audio []float32) (chan string, error) {
	if m.transcribeFunc != nil {
		return m.transcribeFunc(audio)
	}
	ch := make(chan string)
	go func() {
		defer close(ch)
		ch <- "mock transcription"
	}()
	return ch, nil
}

func (m *MockSTTEngine) Close() error {
	m.closed = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestNewSTTManager(t *testing.T) {
	mockEngine := &MockSTTEngine{}
	manager := NewSTTManager(mockEngine)

	if manager.activeEngine != mockEngine {
		t.Errorf("Expected active engine to be set, got %v", manager.activeEngine)
	}
}

func TestSTTManager_SwapEngine(t *testing.T) {
	engine1 := &MockSTTEngine{}
	manager := NewSTTManager(engine1)

	engine2 := &MockSTTEngine{}
	err := manager.SwapEngine(engine2)

	if err != nil {
		t.Errorf("Expected no error when swapping engine, got %v", err)
	}

	if !engine1.closed {
		t.Error("Expected original engine to be closed after swap")
	}

	if manager.activeEngine != engine2 {
		t.Errorf("Expected active engine to be engine2, got %v", manager.activeEngine)
	}
}

func TestSTTManager_SwapEngineErrorClose(t *testing.T) {
	expectedErr := errors.New("failed to close")
	engine1 := &MockSTTEngine{
		closeFunc: func() error {
			return expectedErr
		},
	}
	manager := NewSTTManager(engine1)

	engine2 := &MockSTTEngine{}
	err := manager.SwapEngine(engine2)

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	if manager.activeEngine != engine2 {
		t.Errorf("Expected active engine to still swap despite close error, got %v", manager.activeEngine)
	}
}

func TestSTTManager_TranscribeStream(t *testing.T) {
	expectedString := "test stream"
	engine := &MockSTTEngine{
		transcribeFunc: func(audio []float32) (chan string, error) {
			ch := make(chan string, 1)
			ch <- expectedString
			close(ch)
			return ch, nil
		},
	}
	manager := NewSTTManager(engine)

	ch, err := manager.TranscribeStream([]float32{1, 2, 3})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	result := <-ch
	if result != expectedString {
		t.Errorf("Expected result %v, got %v", expectedString, result)
	}
}

func TestSTTManager_TranscribeStreamNoEngine(t *testing.T) {
	manager := NewSTTManager(nil)

	_, err := manager.TranscribeStream([]float32{})
	if err == nil {
		t.Error("Expected error when transcribing with no active engine")
	}
}

func TestSTTManager_ConcurrentSwapAndTranscribe(t *testing.T) {
	// Let's test that SwapEngine and TranscribeStream don't cause race conditions
	engine1 := &MockSTTEngine{
		transcribeFunc: func(audio []float32) (chan string, error) {
			ch := make(chan string, 1)
			ch <- "engine1"
			close(ch)
			return ch, nil
		},
	}
	manager := NewSTTManager(engine1)

	var wg sync.WaitGroup

	// Start 100 goroutines that just continuously TranscribeStream
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, _ = manager.TranscribeStream([]float32{0})
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// While transcribing, start 10 goroutines swapping out the engine continuously
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				newEngine := &MockSTTEngine{
					transcribeFunc: func(audio []float32) (chan string, error) {
						ch := make(chan string, 1)
						ch <- "new engine"
						close(ch)
						return ch, nil
					},
				}
				_ = manager.SwapEngine(newEngine)
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	// If it doesn't panic/race, the test passes
}
