package llm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestStreamVisionCompletion(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and decode the body
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

	os.Setenv("LLM_BASE_URL", ts.URL)
	defer os.Unsetenv("LLM_BASE_URL")

	var tokens []string
	var doneCalled bool

	err := StreamVisionCompletion("base64data", "Analyze this", func(token string) {
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
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and decode the body
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

	os.Setenv("LLM_BASE_URL", ts.URL)
	defer os.Unsetenv("LLM_BASE_URL")

	err := StreamVisionCompletion("base64data", "", nil, nil)

	if err != nil {
		t.Fatalf("StreamVisionCompletion failed: %v", err)
	}
}

func TestStreamVisionCompletion_NoKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	err := StreamVisionCompletion("b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestStreamVisionCompletion_ErrorCases(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Bad URL
	os.Setenv("LLM_BASE_URL", "http://inv\x00alid-url")
	err := StreamVisionCompletion("b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	// Non-200 Response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	os.Setenv("LLM_BASE_URL", ts.URL)
	err = StreamVisionCompletion("b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error for non-200 status code")
	}
	ts.Close()
	os.Unsetenv("LLM_BASE_URL")
}

func TestStreamVisionCompletion_JsonParseError(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Good\"}}]}\n\n")
		fmt.Fprint(w, "data: invalid_json\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"bye\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer ts.Close()

	os.Setenv("LLM_BASE_URL", ts.URL)
	defer os.Unsetenv("LLM_BASE_URL")

	var tokens []string
	err := StreamVisionCompletion("b64", "Test question", func(token string) {
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
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// jsonMarshal error
	originalJsonMarshal := jsonMarshal
	jsonMarshal = func(v interface{}) ([]byte, error) {
		return nil, errors.New("mock marshal error")
	}
	err := StreamVisionCompletion("b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock jsonMarshal")
	}
	jsonMarshal = originalJsonMarshal

	// httpNewRequest error
	originalHttpNewRequest := httpNewRequest
	httpNewRequest = func(method, url string, body io.Reader) (*http.Request, error) {
		return nil, errors.New("mock request error")
	}
	err = StreamVisionCompletion("b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock httpNewRequest")
	}
	httpNewRequest = originalHttpNewRequest

	// httpClientDo error
	originalHttpClientDo := httpClientDo
	httpClientDo = func(c *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("mock do error")
	}
	err = StreamVisionCompletion("b64", "Test", nil, nil)
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
	err = StreamVisionCompletion("b64", "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock scanner read")
	}
	httpClientDo = originalHttpClientDo
}
