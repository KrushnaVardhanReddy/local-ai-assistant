package remote

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"wails-app/backend/session"
)

func TestServer_HandleSessionEnd(t *testing.T) {
	sm := session.NewSessionManager()
	server := NewServer(nil, sm)

	// Test GET method not allowed
	req := httptest.NewRequest(http.MethodGet, "/session/end", nil)
	rr := httptest.NewRecorder()
	server.handleSessionEnd(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}

	// Test with no turns
	req = httptest.NewRequest(http.MethodPost, "/session/end", nil)
	rr = httptest.NewRecorder()
	server.handleSessionEnd(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	var res map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&res)
	if res["scorecard"] != nil {
		t.Errorf("Expected nil scorecard for 0 turns, got %v", res["scorecard"])
	}
}

func TestServer_HandleSessionHistory(t *testing.T) {
	sm := session.NewSessionManager()
	server := NewServer(nil, sm)

	// Set historyPath to temp dir for tests
	tempDir := t.TempDir()
	tempFile := tempDir + "/session_history.json"
	server.historyPath = tempFile

	// Test POST method not allowed
	req := httptest.NewRequest(http.MethodPost, "/session/history", nil)
	rr := httptest.NewRecorder()
	server.handleSessionHistory(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}

	// Test GET method
	req = httptest.NewRequest(http.MethodGet, "/session/history", nil)
	rr = httptest.NewRecorder()

	// Create dummy file
	os.WriteFile(tempFile, []byte(`[{"dummy": "data"}]`), 0644)

	server.handleSessionHistory(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestServer_HandleSessionClear(t *testing.T) {
	sm := session.NewSessionManager()
	server := NewServer(nil, sm)

	// Test GET method not allowed
	req := httptest.NewRequest(http.MethodGet, "/session/clear", nil)
	rr := httptest.NewRecorder()
	server.handleSessionClear(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}

	// Test POST method
	req = httptest.NewRequest(http.MethodPost, "/session/clear", nil)
	rr = httptest.NewRecorder()
	server.handleSessionClear(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}
