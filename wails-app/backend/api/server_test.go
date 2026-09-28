package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"wails-app/backend"
)

// mockApp implements AppInterface for testing
type mockApp struct {
	micEnabled bool
	appMode    string
	audioMode  string
	buffer     []string
}

func (m *mockApp) GetState() map[string]interface{} {
	return map[string]interface{}{"status": "ok"}
}
func (m *mockApp) GetSystemStatus() map[string]interface{} { return nil }
func (m *mockApp) GetMachineId() string                    { return "test-machine" }
func (m *mockApp) GetCacheStats() map[string]interface{}   { return nil }
func (m *mockApp) GetCacheItems() []backend.CacheItem      { return nil }
func (m *mockApp) GetIndexedPaths() []string               { return nil }
func (m *mockApp) GetBuddyURL() string                     { return "" }
func (m *mockApp) GetIDEState() map[string]interface{}     { return nil }
func (m *mockApp) ToggleMic() bool {
	m.micEnabled = !m.micEnabled
	return m.micEnabled
}
func (m *mockApp) FlushQuestionBuffer() error { return nil }
func (m *mockApp) AppendToBuffer(text string) error {
	m.buffer = append(m.buffer, text)
	return nil
}
func (m *mockApp) SendChat(text string) error                            { return nil }
func (m *mockApp) SetManualMode(enabled bool)                            {}
func (m *mockApp) SetRawMode(enabled bool)                               {}
func (m *mockApp) SetAppMode(mode string) error                          { m.appMode = mode; return nil }
func (m *mockApp) SetAudioMode(mode string) error                        { m.audioMode = mode; return nil }
func (m *mockApp) ClearState()                                           {}
func (m *mockApp) ClearCache() error                                     { return nil }
func (m *mockApp) DeleteCacheItems(ids []string) error                   { return nil }
func (m *mockApp) ExportSession() (string, error)                        { return "", nil }
func (m *mockApp) ToggleMockInterviewMode(enabled bool)                  {}
func (m *mockApp) ToggleMockTTS(enabled bool)                            {}
func (m *mockApp) EndSession() (map[string]interface{}, error)           { return nil, nil }
func (m *mockApp) StartBuddyMode() error                                 { return nil }
func (m *mockApp) StopBuddyMode()                                        {}
func (m *mockApp) CaptureScreen() string                                 { return "" }
func (m *mockApp) AnalyzeVision(base64Image string, prompt string) error { return nil }
func (m *mockApp) SummarizeTranscript() error                            { return nil }

func TestHub(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{send: make(chan []byte, 10)}
	hub.register <- client

	// Wait a moment for register
	time.Sleep(10 * time.Millisecond)

	hub.Broadcast("test_event", "hello")

	select {
	case msg := <-client.send:
		var result map[string]interface{}
		if err := json.Unmarshal(msg, &result); err != nil {
			t.Fatalf("Failed to parse msg: %v", err)
		}
		if result["type"] != "test_event" {
			t.Errorf("Expected event type test_event, got %v", result["type"])
		}
		if result["payload"] != "hello" {
			t.Errorf("Expected payload hello, got %v", result["payload"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}

	hub.unregister <- client
}

func TestServerStartStop(t *testing.T) {
	mock := &mockApp{}
	srv := NewServer(mock)

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	port := srv.Port()
	if port == 0 {
		t.Fatal("Expected bound port to be non-zero")
	}

	// Just enough to allow HTTP server to spin up fully
	time.Sleep(10 * time.Millisecond)

	if err := srv.Stop(context.Background()); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
}

func TestRESTEndpoints(t *testing.T) {
	mock := &mockApp{}
	srv := NewServer(mock)
	mux := http.NewServeMux()
	RegisterRoutes(mux, srv)

	// Test GET /api/v1/state
	req := httptest.NewRequest("GET", "/api/v1/state", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"status":"ok"`) {
		t.Errorf("Expected body to contain status ok, got %s", rr.Body.String())
	}

	// Test POST /api/v1/mic/toggle
	req = httptest.NewRequest("POST", "/api/v1/mic/toggle", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if !mock.micEnabled {
		t.Error("Expected ToggleMic to set micEnabled to true")
	}

	// Test POST /api/v1/llm/append
	reqBody := bytes.NewBuffer([]byte(`{"text":"test append"}`))
	req = httptest.NewRequest("POST", "/api/v1/llm/append", reqBody)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if len(mock.buffer) == 0 || mock.buffer[0] != "test append" {
		t.Errorf("Expected buffer to contain 'test append', got %v", mock.buffer)
	}
}

func TestWebSocketUpgrade(t *testing.T) {
	mock := &mockApp{}
	srv := NewServer(mock)

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop(context.Background())

	// Allow HTTP server to bind
	time.Sleep(50 * time.Millisecond)

	// Build the real WS url based on the dynamic port bound
	port := srv.Port()
	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/ws", port)

	// Actually dial it using gorilla/websocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial websocket upgrade: %v", err)
	}
	defer conn.Close()

	// Wait to confirm registration to the hub
	time.Sleep(10 * time.Millisecond)

	srv.Hub().mu.RLock()
	clientCount := len(srv.Hub().clients)
	srv.Hub().mu.RUnlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 client registered to hub, got %d", clientCount)
	}
}
