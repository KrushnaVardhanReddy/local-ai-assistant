package remote

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
    "net/http"

	"github.com/gorilla/websocket"
)

func TestNewServer(t *testing.T) {
	s := NewServer(nil, nil)
	if s == nil {
		t.Fatal("Expected server instance, got nil")
	}
	if s.httpServer.Addr != "0.0.0.0:8000" {
		t.Errorf("Expected address 0.0.0.0:8000, got %s", s.httpServer.Addr)
	}
}

func TestServer_StartStop(t *testing.T) {
	s := NewServer(nil, nil)
	s.httpServer.Addr = "127.0.0.1:8001"

	s.Start()
	time.Sleep(100 * time.Millisecond) // Give it time to start

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := s.Stop(ctx)
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

func TestServer_StartError(t *testing.T) {
	s1 := NewServer(nil, nil)
	s1.httpServer.Addr = "127.0.0.1:8002"
	s1.Start()
	time.Sleep(50 * time.Millisecond)

	s2 := NewServer(nil, nil)
	s2.httpServer.Addr = "127.0.0.1:8002"
	s2.Start()
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	s1.Stop(ctx)
	s2.Stop(ctx)
}

func TestServer_WebSocket(t *testing.T) {
	s := NewServer(nil, nil)

	ts := httptest.NewServer(s.httpServer.Handler)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}
	defer ws.Close()

	time.Sleep(50 * time.Millisecond)

	s.clientsMu.Lock()
	clientCount := len(s.clients)
	s.clientsMu.Unlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 client, got %d", clientCount)
	}

	s.Broadcast(websocket.TextMessage, []byte("test message"))

	ws.SetReadDeadline(time.Now().Add(time.Second))
	msgType, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}
	if msgType != websocket.TextMessage {
		t.Errorf("Expected text message type, got %d", msgType)
	}
	if string(msg) != "test message" {
		t.Errorf("Expected 'test message', got '%s'", string(msg))
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte("hello server"))
	if err != nil {
		t.Fatalf("Failed to write message: %v", err)
	}

	ws.Close()

	time.Sleep(50 * time.Millisecond)

	s.clientsMu.Lock()
	clientCount = len(s.clients)
	s.clientsMu.Unlock()

	if clientCount != 0 {
		t.Errorf("Expected 0 clients after disconnect, got %d", clientCount)
	}
}

func TestServer_UpgradeError(t *testing.T) {
	s := NewServer(nil, nil)
	ts := httptest.NewServer(s.httpServer.Handler)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/ws", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        t.Fatalf("Failed to do request: %v", err)
    }
    defer resp.Body.Close()
}

func TestServer_BroadcastError(t *testing.T) {
	s := NewServer(nil, nil)
	ts := httptest.NewServer(s.httpServer.Handler)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	// Create a real connection
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial WebSocket: %v", err)
	}

	time.Sleep(50 * time.Millisecond) // wait for client to register

	// Get the server-side connection
	s.clientsMu.Lock()
	var serverConn *websocket.Conn
	for c := range s.clients {
		serverConn = c
		break
	}
	s.clientsMu.Unlock()

	if serverConn == nil {
		t.Fatal("Expected server to have registered the connection")
	}

	// Make the connection un-writable
	serverConn.Close()

	// We need to bypass the lock temporarily to inject a dummy write failure
	s.clientsMu.Lock()
	// This connection is already closed, writing to it will fail
	s.clientsMu.Unlock()

	// Run broadcast in goroutine to not panic/block
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.Broadcast(websocket.TextMessage, []byte("test"))
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Broadcast timed out")
	}

	// Check if the connection was removed
	s.clientsMu.Lock()
	_, exists := s.clients[serverConn]
	s.clientsMu.Unlock()

	if exists {
		t.Errorf("Expected client to be removed after broadcast error")
	}
    ws.Close()
}

func TestServer_StopWithClients(t *testing.T) {
	s := NewServer(nil, nil)
	ts := httptest.NewServer(s.httpServer.Handler)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer ws.Close()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = s.Stop(ctx)
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}

func TestServer_FileSystem(t *testing.T) {
	fs := http.Dir(".")
	s := NewServer(fs, nil)

	ts := httptest.NewServer(s.httpServer.Handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/server.go")
	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
