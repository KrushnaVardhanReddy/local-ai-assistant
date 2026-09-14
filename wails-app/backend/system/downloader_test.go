package system

import (
	"archive/zip"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Helper to create a dummy zip file
func createDummyZip(t *testing.T) []byte {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	var files = []struct {
		Name, Body string
	}{
		{"parakeet/model.onnx", "dummy onnx content"},
		{"parakeet/config.json", "{}"},
	}
	for _, f := range files {
		fWriter, err := w.Create(f.Name)
		if err != nil {
			t.Fatal(err)
		}
		_, err = fWriter.Write([]byte(f.Body))
		if err != nil {
			t.Fatal(err)
		}
	}

	err := w.Close()
	if err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func TestStartBackgroundDownload(t *testing.T) {
	// Setup test mock HTTP server
	zipData := createDummyZip(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		importStrconv := "strconv"
		_ = importStrconv
		// Content-Length will be automatically set by w.Write(zipData) if we don't stream
		w.Write(zipData)
	}))
	defer server.Close()

	// Override global variables for testing
	origModelBundleURL := ModelBundleURL
	ModelBundleURL = server.URL
	defer func() { ModelBundleURL = origModelBundleURL }()

	// Mock hardware profiler variables to ensure IsParakeetSupported returns true
	origGoarch := goarch
	origHasAVX2 := hasAVX2
	origGetMemory := getMemory
	defer func() {
		goarch = origGoarch
		hasAVX2 = origHasAVX2
		getMemory = origGetMemory
	}()
	goarch = "amd64"
	hasAVX2 = true
	getMemory = func() (uint64, error) { return 16 * 1024 * 1024 * 1024, nil }

	// Mock HOME dir so we don't mess with real ~/.local
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome) // Linux/macOS
	t.Setenv("USERPROFILE", tempHome) // Windows

	// Override os.UserHomeDir function by hacking standard lib? No, t.Setenv works in Go 1.15+ for os.UserHomeDir().

	var lastProgress float32
	progressCb := func(p float32) {
		lastProgress = p
	}

	// 1. Initial download
	StartBackgroundDownload(context.Background(), progressCb)

	modelsDir := filepath.Join(tempHome, ".local", "share", "barnowl", "models")
	parakeetDir := filepath.Join(modelsDir, "parakeet")

	// Check if directory was created
	if _, err := os.Stat(parakeetDir); os.IsNotExist(err) {
		t.Fatalf("Expected directory %s to be created", parakeetDir)
	}

	// Check if file extracted successfully
	content, err := os.ReadFile(filepath.Join(parakeetDir, "model.onnx"))
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	if string(content) != "dummy onnx content" {
		t.Errorf("Unexpected content: %s", string(content))
	}

	// Ensure progress was updated
	if lastProgress == 0 {
		t.Errorf("Progress callback was not called")
	}

	// 2. Call again, should skip because directory exists (no panic, no re-download ideally, hard to mock network zero byte download without another server, but we can just ensure it doesn't fail).
	StartBackgroundDownload(context.Background(), progressCb)
}

func TestStartBackgroundDownload_Unsupported(t *testing.T) {
	// Mock hardware profiler variables to ensure IsParakeetSupported returns false
	origGoarch := goarch
	defer func() {
		goarch = origGoarch
	}()
	goarch = "386" // unsupported

	var progressCalled bool
	progressCb := func(p float32) {
		progressCalled = true
	}

	StartBackgroundDownload(context.Background(), progressCb)
	if progressCalled {
		t.Errorf("Expected progress callback not to be called for unsupported system")
	}
}

func TestDownloadFileWithRetry_Failure(t *testing.T) {
	// Setup test mock HTTP server that always returns an error or 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "test.zip")

	// Fast fail timeout by passing a canceled context or mock time
	// To avoid waiting 4 seconds (2 retries * 2s), we can just use a shorter timeout for test,
	// actually the sleep is fixed at 2s in `downloadFileWithRetry`.
	// We can cancel context to exit early.
	ctx, cancel := context.WithCancel(context.Background())
	// wait for 1 attempt, then cancel?
	// Let's just run it with maxRetries = 1
	err := downloadFileWithRetry(ctx, server.URL, dest, nil, 1)
	if err == nil {
		t.Errorf("Expected error from 404 response")
	}
	cancel()
}

func TestExtractZip_ZipSlip(t *testing.T) {
	// Create zip with relative path outside
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	fWriter, err := w.Create("../evil.txt")
	if err == nil {
		fWriter.Write([]byte("evil"))
	}
	w.Close()

	zipPath := filepath.Join(t.TempDir(), "evil.zip")
	os.WriteFile(zipPath, buf.Bytes(), 0644)

	err = extractZip(zipPath, t.TempDir())
	if err == nil || err.Error() == "" {
		t.Errorf("Expected ZipSlip error, got nil")
	}
}
