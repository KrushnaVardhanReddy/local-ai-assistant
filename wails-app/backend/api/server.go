package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

type Server struct {
	mu         sync.Mutex
	httpServer *http.Server
	port       int
	isRunning  bool
	router     *mux.Router
	app        AppInterface
	wsHandler  http.HandlerFunc
}

type AppInterface interface {
	ToggleMic() bool
	FlushQuestionBuffer() error
	AppendToBuffer(text string) error
	GetState() map[string]interface{}
}

func NewServer(port int, app AppInterface, wsHandler http.HandlerFunc) *Server {
	s := &Server{
		port:      port,
		router:    mux.NewRouter(),
		app:       app,
		wsHandler: wsHandler,
	}
	s.setupRoutes()
	return s
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("API server is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: s.router,
	}

	go func() {
		log.Printf("🚀 Starting API server on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ API server error: %v\n", err)
		}
	}()

	return nil
}

func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return
	}
	s.isRunning = false

	if s.httpServer != nil {
		s.httpServer.Shutdown(context.Background())
		s.httpServer = nil
	}
}
