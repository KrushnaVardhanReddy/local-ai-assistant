package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestStreamCompletionWithContext(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and decode the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		var reqBody ChatRequest
		if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		// Verify the messages order and content
		if len(reqBody.Messages) != 3 {
			t.Errorf("Expected 3 messages, got %d", len(reqBody.Messages))
		} else {
			if reqBody.Messages[0].Role != "system" {
				t.Errorf("Expected message 0 role to be 'system', got %q", reqBody.Messages[0].Role)
			}
			if !strings.Contains(reqBody.Messages[0].Content, "STAR format strictly") {
				t.Errorf("Expected system message to contain 'STAR format strictly', got %q", reqBody.Messages[0].Content)
			}
			if reqBody.Messages[1].Role != "assistant" {
				t.Errorf("Expected message 1 role to be 'assistant', got %q", reqBody.Messages[1].Role)
			}
			if reqBody.Messages[2].Role != "user" {
				t.Errorf("Expected message 2 role to be 'user', got %q", reqBody.Messages[2].Role)
			}
			if reqBody.Messages[2].Content != "Test question" {
				t.Errorf("Expected message 2 content to be 'Test question', got %q", reqBody.Messages[2].Content)
			}
		}

		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello \"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"world!\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer ts.Close()

	os.Setenv("LLM_BASE_URL", ts.URL)
	defer os.Unsetenv("LLM_BASE_URL")

	var tokens []string
	var doneCalled bool

	history := []ChatMessage{
		{Role: "assistant", Content: "Previous answer"},
	}

	err := StreamCompletionWithContext(context.Background(), "Test question", "behavioral", history, func(token string) {
		tokens = append(tokens, token)
	}, func() {
		doneCalled = true
	})

	if err != nil {
		t.Fatalf("StreamCompletionWithContext failed: %v", err)
	}

	if !doneCalled {
		t.Error("Expected onDone to be called")
	}

	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	} else {
		if tokens[0] != "Hello " {
			t.Errorf("Expected token 0 to be 'Hello ', got %q", tokens[0])
		}
		if tokens[1] != "world!" {
			t.Errorf("Expected token 1 to be 'world!', got %q", tokens[1])
		}
	}
}

func TestStreamCompletion(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hello \"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer ts.Close()

	os.Setenv("LLM_BASE_URL", ts.URL)
	defer os.Unsetenv("LLM_BASE_URL")

	err := StreamCompletion(context.Background(), "Test question", nil, nil)
	if err != nil {
		t.Fatalf("StreamCompletion failed: %v", err)
	}
}

func TestStreamCompletion_NoKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	err := StreamCompletion(context.Background(), "Test", nil, nil)
	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestStreamCompletion_ErrorCases(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Bad URL
	os.Setenv("LLM_BASE_URL", "http://inv\x00alid-url")
	err := StreamCompletion(context.Background(), "Test", nil, nil)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	// Non-200 Response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	os.Setenv("LLM_BASE_URL", ts.URL)
	err = StreamCompletion(context.Background(), "Test", nil, nil)
	if err == nil {
		t.Error("Expected error for non-200 status code")
	}
	ts.Close()
	os.Unsetenv("LLM_BASE_URL")
}

func TestStreamCompletion_JsonParseError(t *testing.T) {
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
	err := StreamCompletion(context.Background(), "Test question", func(token string) {
		tokens = append(tokens, token)
	}, nil)

	if err != nil {
		t.Fatalf("StreamCompletion failed: %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error mock")
}
func (errReader) Close() error {
	return nil
}

func TestStreamCompletion_MockErrors(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// jsonMarshal error
	originalJsonMarshal := jsonMarshal
	jsonMarshal = func(v interface{}) ([]byte, error) {
		return nil, errors.New("mock marshal error")
	}
	err := StreamCompletion(context.Background(), "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock jsonMarshal")
	}
	jsonMarshal = originalJsonMarshal

	// httpNewRequestWithContext error
	originalHttpNewRequest := httpNewRequestWithContext
	httpNewRequestWithContext = func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
		return nil, errors.New("mock request error")
	}
	err = StreamCompletion(context.Background(), "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock httpNewRequestWithContext")
	}
	httpNewRequestWithContext = originalHttpNewRequest

	// httpClientDo error
	originalHttpClientDo := httpClientDo
	httpClientDo = func(c *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("mock do error")
	}
	err = StreamCompletion(context.Background(), "Test", nil, nil)
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
	err = StreamCompletion(context.Background(), "Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock scanner read")
	}
	httpClientDo = originalHttpClientDo
}
