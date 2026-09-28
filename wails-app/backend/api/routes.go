package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow localhost
}

func RegisterRoutes(mux *http.ServeMux, s *Server) {
	// ── WebSocket (primary real-time channel) ──────────────────────
	mux.HandleFunc("/ws", s.handleWS)

	// ── State queries ──────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/state", jsonHandler(func(r *http.Request) (any, error) { return s.app.GetState(), nil }))
	mux.HandleFunc("GET /api/v1/status", jsonHandler(func(r *http.Request) (any, error) { return s.app.GetSystemStatus(), nil }))
	mux.HandleFunc("GET /api/v1/cache/stats", jsonHandler(func(r *http.Request) (any, error) { return s.app.GetCacheStats(), nil }))
	mux.HandleFunc("GET /api/v1/cache/items", jsonHandler(func(r *http.Request) (any, error) { return s.app.GetCacheItems(), nil }))
	mux.HandleFunc("GET /api/v1/indexed-paths", jsonHandler(func(r *http.Request) (any, error) { return s.app.GetIndexedPaths(), nil }))
	mux.HandleFunc("GET /api/v1/buddy/url", jsonHandler(func(r *http.Request) (any, error) { return s.app.GetBuddyURL(), nil }))
	mux.HandleFunc("GET /api/v1/ide/state", jsonHandler(func(r *http.Request) (any, error) { return s.app.GetIDEState(), nil }))

	// ── Audio / session commands ───────────────────────────────────
	mux.HandleFunc("POST /api/v1/mic/toggle", jsonHandler(func(r *http.Request) (any, error) { return s.app.ToggleMic(), nil }))
	mux.HandleFunc("POST /api/v1/llm/flush", jsonHandler(func(r *http.Request) (any, error) { return nil, s.app.FlushQuestionBuffer() }))
	mux.HandleFunc("POST /api/v1/llm/append", s.handleAppendToBuffer)
	mux.HandleFunc("POST /api/v1/llm/chat", s.handleSendChat)
	mux.HandleFunc("POST /api/v1/session/clear", jsonHandler(func(r *http.Request) (any, error) { s.app.ClearState(); return "ok", nil }))
	mux.HandleFunc("POST /api/v1/session/export", jsonHandler(func(r *http.Request) (any, error) { return s.app.ExportSession() }))
	mux.HandleFunc("POST /api/v1/session/end", jsonHandler(func(r *http.Request) (any, error) { return s.app.EndSession() }))
	mux.HandleFunc("POST /api/v1/session/summarize", jsonHandler(func(r *http.Request) (any, error) { return nil, s.app.SummarizeTranscript() }))
	mux.HandleFunc("POST /api/v1/mode/app", s.handleSetAppMode)
	mux.HandleFunc("POST /api/v1/mode/audio", s.handleSetAudioMode)
	mux.HandleFunc("POST /api/v1/mode/manual", s.handleSetManualMode)
	mux.HandleFunc("POST /api/v1/mode/raw", s.handleSetRawMode)
	mux.HandleFunc("POST /api/v1/mode/mock", s.handleSetMockMode)
	mux.HandleFunc("POST /api/v1/cache/clear", jsonHandler(func(r *http.Request) (any, error) { return nil, s.app.ClearCache() }))
	mux.HandleFunc("POST /api/v1/cache/delete", s.handleDeleteCacheItems)
	mux.HandleFunc("POST /api/v1/buddy/start", jsonHandler(func(r *http.Request) (any, error) { return nil, s.app.StartBuddyMode() }))
	mux.HandleFunc("POST /api/v1/buddy/stop", jsonHandler(func(r *http.Request) (any, error) { s.app.StopBuddyMode(); return "ok", nil }))
	mux.HandleFunc("POST /api/v1/screen/capture", jsonHandler(func(r *http.Request) (any, error) { return s.app.CaptureScreen(), nil }))
	mux.HandleFunc("POST /api/v1/screen/analyze", s.handleAnalyzeVision)

	// ── Audio Devices ──────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/audio/devices", jsonHandler(func(r *http.Request) (any, error) { return s.app.GetAudioDevices(), nil }))
	mux.HandleFunc("POST /api/v1/audio/device", s.handleSetAudioDevice)

	// ── License & Auth ─────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/license/status", jsonHandler(func(r *http.Request) (any, error) { return s.app.CheckLicense(), nil }))
	mux.HandleFunc("POST /api/v1/license/activate", s.handleActivateLicense)
	mux.HandleFunc("POST /api/v1/license/deactivate", jsonHandler(func(r *http.Request) (any, error) { return nil, s.app.DeactivateLicense() }))
	mux.HandleFunc("GET /api/v1/system/machine-id", jsonHandler(func(r *http.Request) (any, error) { return s.app.GetMachineId(), nil }))
	mux.HandleFunc("GET /api/v1/auth/token", jsonHandler(func(r *http.Request) (any, error) { return s.app.LoadToken(), nil }))
	mux.HandleFunc("POST /api/v1/auth/token", s.handleSaveToken)
	mux.HandleFunc("DELETE /api/v1/auth/token", jsonHandler(func(r *http.Request) (any, error) { s.app.DeleteToken(); return "ok", nil }))
	mux.HandleFunc("POST /api/v1/auth/proxy", s.handleSetProxyToken)

	// ── System / App Settings ──────────────────────────────────────
	mux.HandleFunc("POST /api/v1/system/start", jsonHandler(func(r *http.Request) (any, error) { return nil, s.app.StartBackend() }))
	mux.HandleFunc("POST /api/v1/system/stop", jsonHandler(func(r *http.Request) (any, error) { return nil, s.app.StopBackend() }))
	mux.HandleFunc("POST /api/v1/system/quit", jsonHandler(func(r *http.Request) (any, error) { s.app.QuitApp(); return "ok", nil }))
	mux.HandleFunc("POST /api/v1/settings/stealth", s.handleToggleStealth)
	mux.HandleFunc("POST /api/v1/settings/context", s.handleSetIncludeActiveDocContext)

	// ── Cache/Indexed Paths ────────────────────────────────────────
	mux.HandleFunc("DELETE /api/v1/indexed-paths", s.handleRemoveIndexedPath)
}

// handleWS upgrades HTTP to WebSocket and registers the client with the hub.
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256)}
	s.hub.register <- client

	go func() {
		defer func() {
			s.hub.unregister <- client
			conn.Close()
		}()
		for msg := range client.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				break
			}
		}
	}()

	// Read loop (to detect disconnects and handle client→server messages)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
	s.hub.unregister <- client
}

// ── Request body helpers ───────────────────────────────────────────────────────

func (s *Server) handleAppendToBuffer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, nil, s.app.AppendToBuffer(body.Text))
}

func (s *Server) handleSendChat(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, nil, s.app.SendChat(body.Text))
}

func (s *Server) handleSetAppMode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, nil, s.app.SetAppMode(body.Mode))
}

func (s *Server) handleSetAudioMode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, nil, s.app.SetAudioMode(body.Mode))
}

func (s *Server) handleSetManualMode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.app.SetManualMode(body.Enabled)
	writeJSON(w, "ok", nil)
}

func (s *Server) handleSetRawMode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.app.SetRawMode(body.Enabled)
	writeJSON(w, "ok", nil)
}

func (s *Server) handleSetMockMode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled bool `json:"enabled"`
		TTS     bool `json:"tts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.app.ToggleMockInterviewMode(body.Enabled)
	s.app.ToggleMockTTS(body.TTS)
	writeJSON(w, "ok", nil)
}

func (s *Server) handleDeleteCacheItems(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, nil, s.app.DeleteCacheItems(body.IDs))
}

func (s *Server) handleAnalyzeVision(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Image  string `json:"image"`
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, nil, s.app.AnalyzeVision(body.Image, body.Prompt))
}

func (s *Server) handleSetAudioDevice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID         string `json:"id"`
		IsLoopback bool   `json:"isLoopback"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	idInt, _ := strconv.Atoi(body.ID)
	writeJSON(w, nil, s.app.SetAudioDevice(idInt, body.IsLoopback))
}

func (s *Server) handleActivateLicense(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, nil, s.app.ActivateLicense(body.Key))
}

func (s *Server) handleSaveToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.app.SaveToken(map[string]interface{}{"token": body.Token})
	writeJSON(w, "ok", nil)
}

func (s *Server) handleSetProxyToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.app.SetProxyToken(body.Token)
	writeJSON(w, "ok", nil)
}

func (s *Server) handleToggleStealth(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enable bool `json:"enable"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.app.ToggleStealth(map[string]interface{}{"enable": body.Enable})
	writeJSON(w, "ok", nil)
}

func (s *Server) handleSetIncludeActiveDocContext(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Include bool `json:"include"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.app.SetIncludeActiveDocContext(body.Include)
	writeJSON(w, "ok", nil)
}

func (s *Server) handleRemoveIndexedPath(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeJSON(w, nil, fmt.Errorf("path is required"))
		return
	}
	writeJSON(w, nil, s.app.RemoveIndexedPath(path))
}

// ── Response helpers ───────────────────────────────────────────────────────────

func jsonHandler(fn func(*http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := fn(r)
		writeJSON(w, result, err)
	}
}

func writeJSON(w http.ResponseWriter, result any, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"data": result})
}
