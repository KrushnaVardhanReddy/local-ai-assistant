package remote

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	httpServer *http.Server
	clients    map[*websocket.Conn]bool
	clientsMu  sync.Mutex
	upgrader   websocket.Upgrader
}

func NewServer(fs http.FileSystem) *Server {
	s := &Server{
		clients: make(map[*websocket.Conn]bool),
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

	s.httpServer = &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: mux,
	}

	return s
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
