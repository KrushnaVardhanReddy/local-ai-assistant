package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockApp struct {
	micStarted bool
	flushed    bool
	appended   string
}

func (m *mockApp) ToggleMic() bool {
	m.micStarted = !m.micStarted
	return m.micStarted
}

func (m *mockApp) FlushQuestionBuffer() error {
	m.flushed = true
	return nil
}

func (m *mockApp) AppendToBuffer(text string) error {
	m.appended = text
	return nil
}

func (m *mockApp) GetState() map[string]interface{} {
	return map[string]interface{}{"status": "ok"}
}

func TestAPIServer(t *testing.T) {
	app := &mockApp{}
	server := NewServer(8081, app, nil)

	// Start the server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	time.Sleep(100 * time.Millisecond) // Give it time to start

	// Test ToggleMic
	req, _ := http.NewRequest("POST", "/api/v1/mic/toggle", nil)
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Test Append
	body, _ := json.Marshal(map[string]string{"text": "hello"})
	req, _ = http.NewRequest("POST", "/api/v1/llm/append", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	if app.appended != "hello" {
		t.Errorf("Expected appended text to be 'hello', got '%s'", app.appended)
	}

	// Test Flush
	req, _ = http.NewRequest("POST", "/api/v1/llm/flush", nil)
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	if !app.flushed {
		t.Errorf("Expected flushed to be true")
	}

	// Test GetState
	req, _ = http.NewRequest("GET", "/api/v1/state", nil)
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}
