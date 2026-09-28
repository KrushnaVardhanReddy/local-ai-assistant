package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"wails-app/backend"
	"wails-app/backend/audio"
)

// Server is the local REST+WS HTTP server.
type Server struct {
	hub      *Hub
	app      AppInterface // interface to call App methods (see below)
	srv      *http.Server
	port     int
	portChan chan int // sends the bound port back to app.go
}

// AppInterface defines the subset of App methods the REST server needs.
// This decouples the api package from app.go to avoid import cycles.
type AppInterface interface {
	GetState() map[string]interface{}
	GetSystemStatus() map[string]interface{}
	GetMachineId() string
	GetCacheStats() map[string]interface{}
	GetCacheItems() []backend.CacheItem
	GetIndexedPaths() []string
	GetBuddyURL() string
	GetIDEState() map[string]interface{}

	GetAudioDevices() []audio.AudioDevice
	SetAudioDevice(id int, isLoopback bool) error
	CheckLicense() string
	ActivateLicense(key string) error
	DeactivateLicense() error
	LoadToken() string
	SaveToken(token map[string]interface{})
	DeleteToken()
	SetProxyToken(token string)
	StartBackend() error
	StopBackend() error
	QuitApp()
	ToggleStealth(opts map[string]interface{})
	SetIncludeActiveDocContext(include bool)
	RemoveIndexedPath(path string) error

	ToggleMic() bool
	FlushQuestionBuffer() error
	AppendToBuffer(text string) error
	SendChat(text string) error
	SetManualMode(enabled bool)
	SetRawMode(enabled bool)
	SetAppMode(mode string) error
	SetAudioMode(mode string) error
	ClearState()
	ClearCache() error
	DeleteCacheItems(ids []string) error
	ExportSession() (string, error)
	ToggleMockInterviewMode(enabled bool)
	ToggleMockTTS(enabled bool)
	EndSession() (map[string]interface{}, error)
	StartBuddyMode() error
	StopBuddyMode()
	CaptureScreen() string
	AnalyzeVision(base64Image string, prompt string) error
	SummarizeTranscript() error
}

func NewServer(app AppInterface) *Server {
	hub := NewHub()
	return &Server{
		hub:      hub,
		app:      app,
		portChan: make(chan int, 1),
	}
}

// Start binds to a random available port on localhost, starts the hub and HTTP server.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("api: listen: %w", err)
	}
	s.port = listener.Addr().(*net.TCPAddr).Port
	s.portChan <- s.port

	mux := http.NewServeMux()
	RegisterRoutes(mux, s)

	s.srv = &http.Server{
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go s.hub.Run()
	go s.srv.Serve(listener)
	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

// Port returns the bound port (call after Start()).
func (s *Server) Port() int { return s.port }

// Hub returns the broadcast hub so app.go can push events.
func (s *Server) Hub() *Hub { return s.hub }
