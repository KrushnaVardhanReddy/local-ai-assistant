package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"wails-app/backend/config"
)

func TestStreamVisionCompletion(t *testing.T) {
	SetProxyToken("")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		var reqBody VisionChatRequest
		if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Look \"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"here!\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer ts.Close()

	SetConfig(mockCfg(ts.URL))
	defer SetConfig(nil)

	var tokens []string
	var doneCalled bool

	err := StreamVisionCompletion(context.Background(), "base64data", "Analyze this", func(token string) {
		tokens = append(tokens, token)
	}, func() {
		doneCalled = true
	})

	if err != nil {
		t.Fatalf("StreamVisionCompletion failed: %v", err)
	}

	if !doneCalled {
		t.Error("Expected onDone to be called")
	}

	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
}

func TestStreamVisionCompletion_EmptyPrompt(t *testing.T) {
	SetProxyToken("")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		var reqBody VisionChatRequest
		if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		// Check if prompt defaulted correctly
		contentMap := reqBody.Messages[1].Content.([]interface{})
		textItem := contentMap[0].(map[string]interface{})
		if textItem["text"].(string) != DefaultVisionPrompt {
			t.Errorf("Expected prompt to be DefaultVisionPrompt, got %s", textItem["text"])
		}

		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer ts.Close()

	SetConfig(mockCfg(ts.URL))
	defer SetConfig(nil)

	err := StreamVisionCompletion(context.Background(), "base64data", "", nil, nil)
	if err != nil {
		t.Fatalf("StreamVisionCompletion failed: %v", err)
	}
}

func TestStreamVisionCompletion_NoKey(t *testing.T) {
	SetProxyToken("")
	SetConfig(&config.AppConfig{LLMProvider: "openai", LLMBaseURL: "https://api.openai.com/v1", OpenAIAPIKey: ""})
	defer SetConfig(nil)
	err := StreamVisionCompletion(context.Background(), "b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestStreamVisionCompletion_ErrorCases(t *testing.T) {
	SetProxyToken("")

	// Bad URL
	SetConfig(mockCfg("http://inv\x00alid-url"))
	err := StreamVisionCompletion(context.Background(), "b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	// Non-200 Response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	SetConfig(mockCfg(ts.URL))
	err = StreamVisionCompletion(context.Background(), "b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error for non-200 status code")
	}
	ts.Close()
	SetConfig(nil)
}

func TestStreamVisionCompletion_JsonParseError(t *testing.T) {
	SetProxyToken("")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Good\"}}]}\n\n")
		fmt.Fprint(w, "data: invalid_json\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"bye\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer ts.Close()

	SetConfig(mockCfg(ts.URL))
	defer SetConfig(nil)

	var tokens []string
	err := StreamVisionCompletion(context.Background(), "b64", "Test question", func(token string) {
		tokens = append(tokens, token)
	}, nil)

	if err != nil {
		t.Fatalf("StreamVisionCompletion failed: %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
}

func TestStreamVisionCompletion_MockErrors(t *testing.T) {
	SetConfig(mockCfg("https://api.openai.com/v1"))
	defer SetConfig(nil)

	// jsonMarshal error
	originalJsonMarshal := jsonMarshal
	jsonMarshal = func(v interface{}) ([]byte, error) {
		return nil, errors.New("mock marshal error")
	}
	err := StreamVisionCompletion(context.Background(), "b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock jsonMarshal")
	}
	jsonMarshal = originalJsonMarshal

	// httpNewRequestWithContext error
	originalHttpNewRequest := httpNewRequestWithContext
	httpNewRequestWithContext = func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
		return nil, errors.New("mock request error")
	}
	err = StreamVisionCompletion(context.Background(), "b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock httpNewRequestWithContext")
	}
	httpNewRequestWithContext = originalHttpNewRequest

	// httpClientDo error
	originalHttpClientDo := httpClientDo
	httpClientDo = func(c *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("mock do error")
	}
	err = StreamVisionCompletion(context.Background(), "b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock httpClientDo")
	}
	httpClientDo = originalHttpClientDo

	// scanner error - returning an error from Read
	httpClientDo = func(c *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       errReader{},
		}, nil
	}
	err = StreamVisionCompletion(context.Background(), "b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock scanner read")
	}
	httpClientDo = originalHttpClientDo
}
