package remote

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"wails-app/backend/llm"
	"wails-app/backend/session"
)

type Server struct {
	httpServer     *http.Server
	clients        map[*websocket.Conn]bool
	clientsMu      sync.Mutex
	upgrader       websocket.Upgrader
	sessionManager *session.SessionManager
	historyPath    string
}

func NewServer(fs http.FileSystem, sessionManager *session.SessionManager) *Server {
	s := &Server{
		clients:        make(map[*websocket.Conn]bool),
		sessionManager: sessionManager,
		historyPath:    "./data/session_history.json",
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	mux := http.NewServeMux()
	if fs != nil {
		mux.Handle("/", http.FileServer(fs))
	}
	mux.HandleFunc("/ws", s.handleWS)
	mux.HandleFunc("/session/end", s.handleSessionEnd)
	mux.HandleFunc("/session/history", s.handleSessionHistory)
	mux.HandleFunc("/session/clear", s.handleSessionClear)

	s.httpServer = &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: mux,
	}

	return s
}

func (s *Server) handleSessionEnd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.sessionManager == nil {
		http.Error(w, "Session manager not configured", http.StatusInternalServerError)
		return
	}

	sessionData := s.sessionManager.Export()
	turnCount, _ := sessionData["turn_count"].(int)
	if turnCount == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"session": sessionData,
			"scorecard": nil,
		})
		return
	}

	scorecard, err := llm.GenerateScorecard(sessionData)
	if err != nil {
		http.Error(w, "Failed to generate scorecard: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Persist to history
	var histories []map[string]interface{}

	// Create data dir based on historyPath directory
	importPathParts := []byte(s.historyPath)
	dirPath := "./data"
	for i := len(importPathParts) - 1; i >= 0; i-- {
		if importPathParts[i] == '/' {
			dirPath = string(importPathParts[:i])
			break
		}
	}

	if err := os.MkdirAll(dirPath, 0755); err == nil {
		if b, err := os.ReadFile(s.historyPath); err == nil {
			json.Unmarshal(b, &histories)
		}

		histories = append(histories, map[string]interface{}{
			"session": sessionData,
			"scorecard": scorecard,
		})

		if b, err := json.MarshalIndent(histories, "", "  "); err == nil {
			os.WriteFile(s.historyPath, b, 0644)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session": sessionData,
		"scorecard": scorecard,
	})
}

func (s *Server) handleSessionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var histories []map[string]interface{}
	if b, err := os.ReadFile(s.historyPath); err == nil {
		json.Unmarshal(b, &histories)
	}

	if histories == nil {
		histories = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(histories)
}

func (s *Server) handleSessionClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.sessionManager != nil {
		s.sessionManager.Clear()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "cleared",
	})
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WS: %v", err)
		return
	}

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
		conn.Close()
	}()

	for {
		// Basic connection manager: read messages and discard
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) Start() {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Remote server error: %v", err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	s.clientsMu.Lock()
	for conn := range s.clients {
		conn.Close()
	}
	s.clientsMu.Unlock()

	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Broadcast(messageType int, message []byte) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	for conn := range s.clients {
		err := conn.WriteMessage(messageType, message)
		if err != nil {
			log.Printf("Error writing to WS: %v", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}
