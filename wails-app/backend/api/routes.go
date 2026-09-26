package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) setupRoutes() {
	// Add REST API routes
	s.router.HandleFunc("/api/v1/mic/toggle", s.handleToggleMic).Methods("POST")
	s.router.HandleFunc("/api/v1/llm/flush", s.handleFlushLLM).Methods("POST")
	s.router.HandleFunc("/api/v1/llm/append", s.handleAppendLLM).Methods("POST")
	s.router.HandleFunc("/api/v1/state", s.handleGetState).Methods("GET")

	// Add WebSocket route handling
	if s.wsHandler != nil {
		s.router.HandleFunc("/ws", s.wsHandler).Methods("GET")
	}
}

func (s *Server) handleToggleMic(w http.ResponseWriter, r *http.Request) {
	started := s.app.ToggleMic()
	json.NewEncoder(w).Encode(map[string]interface{}{"started": started})
}

func (s *Server) handleFlushLLM(w http.ResponseWriter, r *http.Request) {
	err := s.app.FlushQuestionBuffer()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAppendLLM(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := s.app.AppendToBuffer(req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	state := s.app.GetState()
	json.NewEncoder(w).Encode(state)
}
