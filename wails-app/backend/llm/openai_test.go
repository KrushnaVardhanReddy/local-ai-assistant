package llm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"errors"
	"io"
)

func TestStreamCompletion(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	err := StreamCompletion("Test question", func(token string) {
		tokens = append(tokens, token)
	}, func() {
		doneCalled = true
	})

	if err != nil {
		t.Fatalf("StreamCompletion failed: %v", err)
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

func TestStreamCompletion_NoKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	err := StreamCompletion("Test", nil, nil)
	if err == nil {
		t.Error("Expected error for missing API key")
	}
}

func TestStreamCompletion_ErrorCases(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

    // Bad URL
    os.Setenv("LLM_BASE_URL", "http://inv\x00alid-url")
    err := StreamCompletion("Test", nil, nil)
    if err == nil {
        t.Error("Expected error for invalid URL")
    }

	// Non-200 Response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	os.Setenv("LLM_BASE_URL", ts.URL)
	err = StreamCompletion("Test", nil, nil)
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
	err := StreamCompletion("Test question", func(token string) {
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
	err := StreamCompletion("Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock jsonMarshal")
	}
	jsonMarshal = originalJsonMarshal

	// httpNewRequest error
	originalHttpNewRequest := httpNewRequest
	httpNewRequest = func(method, url string, body io.Reader) (*http.Request, error) {
		return nil, errors.New("mock request error")
	}
	err = StreamCompletion("Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock httpNewRequest")
	}
	httpNewRequest = originalHttpNewRequest

	// httpClientDo error
	originalHttpClientDo := httpClientDo
	httpClientDo = func(c *http.Client, req *http.Request) (*http.Response, error) {
		return nil, errors.New("mock do error")
	}
	err = StreamCompletion("Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock httpClientDo")
	}
	httpClientDo = originalHttpClientDo

    // scanner error - returning an error from Read
    httpClientDo = func(c *http.Client, req *http.Request) (*http.Response, error) {
		return &http.Response{
            StatusCode: http.StatusOK,
            Body: errReader{},
        }, nil
	}
	err = StreamCompletion("Test", nil, nil)
	if err == nil {
		t.Error("Expected error from mock scanner read")
	}
	httpClientDo = originalHttpClientDo
}
