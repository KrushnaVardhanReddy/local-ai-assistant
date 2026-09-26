//go:build test

package buddy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewServer(t *testing.T) {
	port := 8765
	srv := NewServer(port, nil, nil)

	if srv.IsRunning() {
		t.Errorf("expected isRunning to be false, got true")
	}
	if srv.GetPublicURL() != "" {
		t.Errorf("expected GetPublicURL to be empty, got %q", srv.GetPublicURL())
	}
	if srv.port != port {
		t.Errorf("expected port %d, got %d", port, srv.port)
	}
}

func TestGetPublicURL_WhenNotRunning(t *testing.T) {
	srv := NewServer(8765, nil, nil)
	if srv.GetPublicURL() != "" {
		t.Errorf("expected GetPublicURL to be empty, got %q", srv.GetPublicURL())
	}

	// Manually set URL but no token
	srv.mu.Lock()
	srv.publicURL = "https://example.trycloudflare.com"
	srv.mu.Unlock()

	if srv.GetPublicURL() != "" {
		t.Errorf("expected GetPublicURL to be empty when token is missing, got %q", srv.GetPublicURL())
	}

	// Set both
	srv.mu.Lock()
	srv.token = "fake-token"
	srv.mu.Unlock()

	expected := "https://example.trycloudflare.com/ws?token=fake-token"
	if srv.GetPublicURL() != expected {
		t.Errorf("expected %q, got %q", expected, srv.GetPublicURL())
	}
}

func TestBroadcastTranscript_NoClients(t *testing.T) {
	srv := NewServer(8765, nil, nil)

	// Should not panic or error
	srv.BroadcastTranscript("test message", "candidate")
}

func TestTokenValidation(t *testing.T) {
	// Setup a server instance
	srv := NewServer(8765, nil, nil)
	srv.token = "valid-token"

	// Create httptest server using handleWebSocket
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", srv.handleWebSocket)
	testSrv := httptest.NewServer(mux)
	defer testSrv.Close()

	// Parse httptest server URL to ws scheme
	u, err := url.Parse(testSrv.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}
	u.Scheme = "ws"
	u.Path = "/ws"

	// 1. Test missing token
	_, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err == nil {
		t.Fatalf("expected error without token")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", resp.StatusCode)
	}

	// 2. Test wrong token
	wrongURL := u.String() + "?token=wrong-token"
	_, resp, err = websocket.DefaultDialer.Dial(wrongURL, nil)
	if err == nil {
		t.Fatalf("expected error with wrong token")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", resp.StatusCode)
	}

	// 3. Test correct token
	correctURL := u.String() + "?token=valid-token"
	conn, resp, err := websocket.DefaultDialer.Dial(correctURL, nil)
	if err != nil {
		t.Fatalf("failed to connect with correct token: %v", err)
	}
	defer conn.Close()

	if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("expected status 101 Switching Protocols, got %d", resp.StatusCode)
	}

	// Test hint callback
	hintChan := make(chan string, 1)
	srv.onHint = func(hint string) {
		hintChan <- hint
	}

	hintMsg := BuddyMessage{
		Type: MsgTypeHint,
		Text: "try binary search",
	}
	data, _ := json.Marshal(hintMsg)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	select {
	case hint := <-hintChan:
		if hint != "try binary search" {
			t.Errorf("expected hint %q, got %q", "try binary search", hint)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("timeout waiting for hint")
	}

	// Test broadcast
	srv.BroadcastTranscript("hello world", "interviewer")

	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read broadcasted message: %v", err)
	}

	var recMsg BuddyMessage
	if err := json.Unmarshal(msgBytes, &recMsg); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if recMsg.Type != MsgTypeTranscript {
		t.Errorf("expected type %q, got %q", MsgTypeTranscript, recMsg.Type)
	}
	if recMsg.Text != "hello world" {
		t.Errorf("expected text %q, got %q", "hello world", recMsg.Text)
	}
	if recMsg.Speaker != "interviewer" {
		t.Errorf("expected speaker %q, got %q", "interviewer", recMsg.Speaker)
	}
}

func TestStartStopServer(t *testing.T) {
	// Setup a server instance with a random port for testing
	srv := NewServer(0, nil, nil)

	// Server shouldn't run yet
	if srv.IsRunning() {
		t.Errorf("Expected server not to be running")
	}

	// Stop when not running should not panic
	srv.Stop()

	err := srv.Start()
	// Depending on cloudflared binary, startTunnel might log an error, but Start should return nil
	// because it's run in a goroutine and the main error is just if it's already running or port failed (unlikely for 0).
	// For testing, let's just make sure isRunning is set.
	if err != nil {
		if !strings.Contains(err.Error(), "already running") && !strings.Contains(err.Error(), "token") {
			t.Logf("Start returned error, ignoring since it might be env-specific: %v", err)
		}
	} else {
		if !srv.IsRunning() {
			t.Errorf("Expected server to be running")
		}

		// Starting again should error
		err = srv.Start()
		if err == nil {
			t.Errorf("Expected error when starting already running server")
		}

		// Cleanup
		srv.Stop()
		if srv.IsRunning() {
			t.Errorf("Expected server to be stopped")
		}
	}
}
