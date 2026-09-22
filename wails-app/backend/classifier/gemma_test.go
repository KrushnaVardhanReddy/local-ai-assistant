package classifier

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestIsQuestionComplete_True(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Mocking the OpenAI-compatible response format
			fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"result\":true}"}}]}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Extract port from ts.URL robustly
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}
	portStr := u.Port()
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Failed to parse port: %v", err)
	}

	// Override the port temporarily
	origPort := LlamaServerPort
	LlamaServerPort = port
	defer func() { LlamaServerPort = origPort }()

	ctx := context.Background()
	result, err := IsQuestionComplete(ctx, "What is your greatest weakness?")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result {
		t.Errorf("Expected result to be true, got false")
	}
}

func TestIsQuestionComplete_False(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Mocking the OpenAI-compatible response format
			fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"result\":false}"}}]}`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// Extract port from ts.URL robustly
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}
	portStr := u.Port()
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Failed to parse port: %v", err)
	}

	// Override the port temporarily
	origPort := LlamaServerPort
	LlamaServerPort = port
	defer func() { LlamaServerPort = origPort }()

	ctx := context.Background()
	result, err := IsQuestionComplete(ctx, "So, I was thinking")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result {
		t.Errorf("Expected result to be false, got true")
	}
}

func TestIsQuestionComplete_ServerDown(t *testing.T) {
	// Pick a port that is highly unlikely to be used by any server during the test
	origPort := LlamaServerPort
	LlamaServerPort = 48291 // Random high port
	defer func() { LlamaServerPort = origPort }()

	ctx := context.Background()
	result, err := IsQuestionComplete(ctx, "Hello?")

	if err == nil {
		t.Fatalf("Expected an error when the server is down, got nil")
	}

	if result {
		t.Errorf("Expected result to be false on error, got true")
	}
}
