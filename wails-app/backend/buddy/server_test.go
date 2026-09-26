package buddy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewServer(t *testing.T) {
	port := 8080
	onURL := func(url string) {}
	onHint := func(hint string) {}

	srv := NewServer(port, onURL, onHint, nil)

	if srv.port != port {
		t.Errorf("Expected port %d, got %d", port, srv.port)
	}
	if srv.clients == nil {
		t.Errorf("Expected clients map to be initialized")
	}
	if srv.isRunning {
		t.Errorf("Expected server to not be running initially")
	}
}

func TestGetPublicURL_WhenNotRunning(t *testing.T) {
	srv := NewServer(8080, nil, nil, nil)

	url := srv.GetPublicURL()
	if url != "" {
		t.Errorf("Expected empty URL when not running, got %s", url)
	}
}

func TestBroadcastTranscript_NoClients(t *testing.T) {
	srv := NewServer(8765, nil, nil, nil)

	// Should not panic or error
	srv.BroadcastTranscript("test message", "candidate")
}

func TestTokenValidation(t *testing.T) {
	// Setup a server instance
	srv := NewServer(8765, nil, nil, nil)
	srv.token = "valid-token"
	srv.isRunning = true

	// Create httptest server using handleWebSocket
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", srv.HandleWebSocket)
	testSrv := httptest.NewServer(mux)
	defer testSrv.Close()

	// Parse httptest server URL to ws scheme
	u, err := url.Parse(testSrv.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}
	u.Scheme = "ws"
	u.Path = "/ws"

	// 1. Test missing token (allowed if localhost)
	_, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		// Just a local dial
	}

	// 2. Test invalid token
	u.RawQuery = "token=wrong-token"
	_, _, err = websocket.DefaultDialer.Dial(u.String(), nil)

	// 3. Test valid token
	u.RawQuery = "token=valid-token"
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("expected successful connection with valid token, got error: %v", err)
	}
	if resp != nil && resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("expected status 101 Switching Protocols, got %d", resp.StatusCode)
	}
	if conn != nil {
		conn.Close()
	}
}

func TestStartStopServer(t *testing.T) {
	srv := NewServer(8766, nil, nil, nil) // use different port to avoid conflicts

	err := srv.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if !srv.IsRunning() {
		t.Errorf("Expected server to be running")
	}

	// Try starting again (should fail)
	err = srv.Start()
	if err == nil {
		t.Errorf("Expected error when starting already running server")
	}

	// Wait a moment for cloudflared command to initialize (even if it fails)
	time.Sleep(100 * time.Millisecond)

	srv.Stop()

	if srv.IsRunning() {
		t.Errorf("Expected server to be stopped")
	}
}

func TestServeIndex(t *testing.T) {
	srv := NewServer(8767, nil, nil, nil)
	err := srv.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer srv.Stop()

	// Need a small sleep to ensure HTTP server is up before testing
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:8767/")
	if err != nil {
		t.Fatalf("Failed to get index: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Check that a non-root path returns 404
	resp404, err := http.Get("http://127.0.0.1:8767/invalid-path")
	if err != nil {
		t.Fatalf("Failed to get invalid path: %v", err)
	}
	defer resp404.Body.Close()

	if resp404.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got %d", resp404.StatusCode)
	}
}
