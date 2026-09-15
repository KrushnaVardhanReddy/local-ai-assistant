package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFileAtomic(t *testing.T) {
	content := "test content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "test.txt")

	err := downloadFileAtomic(context.Background(), server.URL, dest)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(b) != content {
		t.Fatalf("Expected %s, got: %s", content, string(b))
	}
}

func TestDownloadFileAtomic_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	dest := filepath.Join(t.TempDir(), "test.txt")

	err := downloadFileAtomic(context.Background(), server.URL, dest)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestEnsureNomicModelFiles_AlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	baseDir := "models/nomic-embed-text-v1.5"
	os.MkdirAll(baseDir, 0755)

	tokPath := filepath.Join(baseDir, "tokenizer.json")
	modelPath := filepath.Join(baseDir, "model.onnx")

	os.WriteFile(tokPath, []byte("tok"), 0644)
	os.WriteFile(modelPath, []byte("mod"), 0644)

	tPath, mPath, err := ensureNomicModelFiles()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if tPath != tokPath || mPath != modelPath {
		t.Fatalf("Expected paths %s and %s, got %s and %s", tokPath, modelPath, tPath, mPath)
	}
}

func TestEnsureNomicModelFiles_DownloadsFiles(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	serverTok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("tok"))
	}))
	defer serverTok.Close()

	serverMod := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("mod"))
	}))
	defer serverMod.Close()

	origTokUrl := NomicTokenizerURL
	origModUrl := NomicModelURL
	NomicTokenizerURL = serverTok.URL
	NomicModelURL = serverMod.URL
	defer func() {
		NomicTokenizerURL = origTokUrl
		NomicModelURL = origModUrl
	}()

	tPath, mPath, err := ensureNomicModelFiles()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	b, _ := os.ReadFile(tPath)
	if string(b) != "tok" {
		t.Fatalf("Expected tok, got: %s", string(b))
	}

	b, _ = os.ReadFile(mPath)
	if string(b) != "mod" {
		t.Fatalf("Expected mod, got: %s", string(b))
	}
}

func TestEnsureNomicModelFiles_DownloadError(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(origWd)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	origTokUrl := NomicTokenizerURL
	NomicTokenizerURL = server.URL
	defer func() {
		NomicTokenizerURL = origTokUrl
	}()

	_, _, err := ensureNomicModelFiles()
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}
