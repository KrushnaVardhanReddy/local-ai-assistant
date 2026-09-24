package classifier

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

type mockEventPort struct{}

func (m *mockEventPort) Emit(event string, payload any) {}

func TestEnsureGemmaModelFile_AlreadyExists(t *testing.T) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Skip("Cannot get user cache dir")
	}

	modelDir := filepath.Join(cacheDir, "barnowl-ai", "models")
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	modelPath := filepath.Join(modelDir, "gemma-3-270m-it-Q8_0.gguf")
	// Create a dummy file to simulate already existing
	if err := os.WriteFile(modelPath, []byte("dummy model content"), 0644); err != nil {
		t.Fatalf("Failed to write dummy model file: %v", err)
	}
	defer os.Remove(modelPath)

	// Keep old URL and restore it later to ensure clean state for other tests,
	// but we don't actually need to mock it if the file exists. We will mock it just to be safe.
	origURL := GemmaModelURL
	GemmaModelURL = "http://invalid-url-should-not-be-called"
	defer func() { GemmaModelURL = origURL }()

	gotPath, err := EnsureGemmaModelFile(context.Background(), &mockEventPort{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if gotPath != modelPath {
		t.Errorf("Expected path %s, got %s", modelPath, gotPath)
	}
}

func TestEnsureLlamaServerBinary_AlreadyExists(t *testing.T) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Skip("Cannot get user cache dir")
	}

	binDir := filepath.Join(cacheDir, "barnowl-ai", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	binName := "llamafile"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)

	if err := os.WriteFile(binPath, []byte("dummy binary content"), 0755); err != nil {
		t.Fatalf("Failed to write dummy bin file: %v", err)
	}
	defer os.Remove(binPath)

	// Mock URLs to ensure we don't download
	origURLs := GemmaModelURL
	GemmaModelURL = "http://invalid-url-should-not-be-called"
	defer func() { GemmaModelURL = origURLs }()

	gotPath, err := EnsureLlamaServerBinary(context.Background(), &mockEventPort{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if gotPath != binPath {
		t.Errorf("Expected path %s, got %s", binPath, gotPath)
	}
}

func TestLlamaServerProcess_StartStop(t *testing.T) {
	// Create a dummy model file
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Skip("Cannot get user cache dir")
	}
	modelDir := filepath.Join(cacheDir, "barnowl-ai", "models")
	os.MkdirAll(modelDir, 0755)
	modelPath := filepath.Join(modelDir, "gemma-3-270m-it-Q8_0.gguf")
	os.WriteFile(modelPath, []byte("dummy"), 0644)
	defer os.Remove(modelPath)

	// Create a dummy shell script (or bat on Windows) to act as the binary
	binDir := filepath.Join(cacheDir, "barnowl-ai", "bin")
	os.MkdirAll(binDir, 0755)
	binName := "llamafile"
	var scriptContent string
	if runtime.GOOS == "windows" {
		binName += ".bat"
		// Loop forever
		scriptContent = "@echo off\n:loop\nping 127.0.0.1 -n 2 > nul\ngoto loop\n"
	} else {
		scriptContent = "#!/bin/sh\nsleep 30\n"
	}
	binPath := filepath.Join(binDir, binName)
	os.WriteFile(binPath, []byte(scriptContent), 0755)
	defer os.Remove(binPath)

	// Mock URLs to use the local files directly (bypassing download logic for testing Start itself,
	// but the Start function calls EnsureLlamaServerBinary which will see the file exists!)
	origGemmaModelURL := GemmaModelURL
	GemmaModelURL = "http://invalid"
	defer func() { GemmaModelURL = origGemmaModelURL }()

	origURLs := GemmaModelURL
	// We need EnsureLlamaServerBinary to look for our dummy bin name.
	// EnsureLlamaServerBinary explicitly adds ".exe" on Windows, but our script is ".bat".
	// To fix this test for Windows, we will just rename it to .exe and let it be treated as an executable.
	// Wait, Windows can't run a .bat if it's named .exe easily via exec.Command without some shell.
	// Actually, Go's exec.Command on Windows might not like a script named .exe.
	// Let's just create an empty .exe and run it? No, it will exit immediately.
	// Let's use a very simple go program compiled as the dummy? No, too slow.
	// Let's just mock the HTTP health endpoint, and let the process run whatever.
	// If the file exists, EnsureLlamaServerBinary returns it.
	if runtime.GOOS == "windows" {
		binName = "llama-server.exe"
		binPath = filepath.Join(binDir, binName)
		// Try a basic batch-like thing but named .exe, it might fail.
		// Let's just use "ping" as the dummy binary if on Windows.
		// Actually, we can't easily override what EnsureLlamaServerBinary returns without it checking os.UserCacheDir.
		// EnsureLlamaServerBinary returns cacheDir/barnowl-ai/bin/llama-server.exe
		// Let's just write to that path! If it's not a real exe, exec.Command might fail on Windows.
		// Let's write a small batch script and then in our mock of EnsureLlamaServerBinary, we'd need to change things.
		// But we cannot mock EnsureLlamaServerBinary.
		// Let's just write "ping.exe" content? No.
		// We'll write an empty file and hope exec.Command doesn't immediately fail before we mock the HTTP?
		// No, exec.Command will fail if it's not an executable format.
		// We can compile a tiny go program right here.
		src := `package main
import "time"
func main() { time.Sleep(30 * time.Second) }`
		srcPath := filepath.Join(binDir, "dummy.go")
		os.WriteFile(srcPath, []byte(src), 0644)
		cmd := exec.Command("go", "build", "-o", binPath, srcPath)
		if err := cmd.Run(); err != nil {
			t.Logf("Failed to compile dummy go program, using simple script fallback: %v", err)
			os.WriteFile(binPath, []byte(scriptContent), 0755)
		}
		os.Remove(srcPath)
	}

	GemmaModelURL = "http://invalid-url-should-not-be-called"
	defer func() { GemmaModelURL = origURLs }()

	// Start a mock HTTP server for the health check
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
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

	p := &LlamaServerProcess{port: port}
	err = p.Start(context.Background(), &mockEventPort{})
	if err != nil {
		t.Fatalf("Failed to start process: %v", err)
	}

	if !p.started {
		t.Errorf("Expected started=true")
	}

	// Test idempotency
	err = p.Start(context.Background(), &mockEventPort{})
	if err != nil {
		t.Fatalf("Expected no error on second start, got: %v", err)
	}

	p.Stop()

	if p.started {
		t.Errorf("Expected started=false after Stop()")
	}
}
