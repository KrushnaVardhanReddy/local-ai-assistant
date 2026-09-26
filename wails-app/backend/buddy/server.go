package buddy

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// Message types exchanged over the buddy WebSocket
const (
	MsgTypeTranscript = "transcript" // candidate interview transcript → friend
	MsgTypeHint       = "hint"       // friend → candidate
	MsgTypeState      = "state"      // buddy session state update → friend
)

// BuddyMessage is the JSON structure sent over the WebSocket.
type BuddyMessage struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Speaker string `json:"speaker,omitempty"` // "interviewer" or "candidate"
}

// Server manages the buddy HTTP+WebSocket server and the cloudflared subprocess.
type Server struct {
	mu         sync.Mutex
	httpServer *http.Server
	upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]bool
	clientsMu  sync.Mutex
	token      string    // random secret token for the shared URL
	publicURL  string    // e.g. "https://quick-fox-123.trycloudflare.com"
	cfCmd      *exec.Cmd // the cloudflared subprocess
	port       int       // local HTTP port for the buddy server (default 8765)
	isRunning  bool
	onURL      func(url string)  // callback invoked when the public URL is known
	onHint     func(hint string) // callback invoked when a friend sends a hint
}

// NewServer creates a new BuddyServer with the given port.
func NewServer(port int, onURL func(string), onHint func(string)) *Server {
	return &Server{
		clients: make(map[*websocket.Conn]bool),
		port:    port,
		onURL:   onURL,
		onHint:  onHint,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Start launches the local HTTP server and the cloudflared tunnel.
// Returns an error if the server is already running or fails to start.
func (s *Server) Start() error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("buddy server is already running")
	}

	// Generate a cryptographically random 16-byte hex token
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to generate token: %w", err)
	}
	s.token = hex.EncodeToString(tokenBytes)
	s.isRunning = true
	s.mu.Unlock()

	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ [Buddy] HTTP server error: %v\n", err)
		}
	}()

	// Start cloudflared tunnel in a goroutine
	go s.startTunnel()

	return nil
}

// Stop gracefully shuts down the HTTP server and kills the cloudflared subprocess.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}
	s.isRunning = false
	s.publicURL = ""
	s.token = ""

	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
		s.httpServer = nil
	}

	if s.cfCmd != nil && s.cfCmd.Process != nil {
		s.cfCmd.Process.Kill()
		s.cfCmd = nil
	}

	// Close all connected buddy clients
	s.clientsMu.Lock()
	for conn := range s.clients {
		conn.Close()
	}
	s.clients = make(map[*websocket.Conn]bool)
	s.clientsMu.Unlock()
}

// IsRunning returns whether the buddy server is currently active.
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunning
}

// GetPublicURL returns the current Cloudflare public URL with the embedded token.
// Returns empty string if the tunnel is not yet established.
func (s *Server) GetPublicURL() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.publicURL == "" || s.token == "" {
		return ""
	}
	return fmt.Sprintf("%s/ws?token=%s", s.publicURL, s.token)
}

// BroadcastTranscript sends a new transcript segment to all connected buddy clients.
func (s *Server) BroadcastTranscript(text, speaker string) {
	msg := BuddyMessage{
		Type:    MsgTypeTranscript,
		Text:    text,
		Speaker: speaker,
	}
	s.broadcast(msg)
}

// broadcast marshals a message to JSON and sends it to all connected clients.
func (s *Server) broadcast(msg BuddyMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("❌ [Buddy] Failed to marshal message: %v\n", err)
		return
	}
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for conn := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("⚠️ [Buddy] Failed to write to client: %v\n", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

// handleWebSocket upgrades the HTTP connection to a WebSocket and validates the token.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Validate the security token from query params
	s.mu.Lock()
	expectedToken := s.token
	s.mu.Unlock()

	clientToken := r.URL.Query().Get("token")
	if clientToken == "" || clientToken != expectedToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ [Buddy] WebSocket upgrade error: %v\n", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	log.Printf("✅ [Buddy] Friend connected from %s\n", r.RemoteAddr)

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
		log.Printf("👋 [Buddy] Friend disconnected\n")
	}()

	// Listen for incoming hints from the friend
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var msg BuddyMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}
		if msg.Type == MsgTypeHint && msg.Text != "" && s.onHint != nil {
			s.onHint(msg.Text)
		}
	}
}

// startTunnel spawns `cloudflared tunnel --url http://127.0.0.1:<port>` and
// parses its stderr output to extract the public URL.
// cloudflared must be installed and available on PATH.
func (s *Server) startTunnel() {
	cmd := exec.Command("cloudflared", "tunnel", "--url", fmt.Sprintf("http://127.0.0.1:%d", s.port))
	s.mu.Lock()
	s.cfCmd = cmd
	s.mu.Unlock()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("❌ [Buddy] Failed to get cloudflared stderr pipe: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("❌ [Buddy] Failed to start cloudflared: %v\n", err)
		return
	}

	log.Println("🌐 [Buddy] cloudflared tunnel starting...")

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[cloudflared] %s\n", line)
		// cloudflared prints the URL in a line like:
		// "... | INFO | +----------------------------...+\n | https://quick-fox-123.trycloudflare.com |"
		// or more commonly: "Your quick Tunnel has been created! Visit it at: https://..."
		// We scan for the trycloudflare.com URL pattern.
		if idx := strings.Index(line, "trycloudflare.com"); idx != -1 {
			// Walk backwards to find the start of the URL (https://)
			start := strings.LastIndex(line[:idx], "https://")
			if start == -1 {
				start = strings.LastIndex(line[:idx], "http://")
			}
			if start != -1 {
				rawURL := line[start : idx+len("trycloudflare.com")]
				// Trim any trailing whitespace or pipes
				rawURL = strings.TrimRight(rawURL, " |")
				s.mu.Lock()
				s.publicURL = rawURL
				s.mu.Unlock()
				log.Printf("🔗 [Buddy] Public URL established: %s\n", rawURL)
				if s.onURL != nil {
					s.onURL(s.GetPublicURL())
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("⚠️ [Buddy] cloudflared exited: %v\n", err)
	}
}
