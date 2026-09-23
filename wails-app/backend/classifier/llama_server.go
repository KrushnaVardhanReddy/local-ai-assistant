package classifier

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"wails-app/backend/system"
	"wails-app/core/ports/driven"
)

var (
	// Gemma 3 270M Q8_0 GGUF from unsloth on Hugging Face.
	// Primary URL is GitHub Releases CDN, fallback is HuggingFace direct.
	GemmaModelURL         = "https://github.com/KrushnaVardhanReddy/local-ai-assistant/releases/download/v1.0-models/gemma-3-270m-it-Q8_0.gguf"
	GemmaModelFallbackURL = "https://huggingface.co/unsloth/gemma-3-270m-it-GGUF/resolve/main/gemma-3-270m-it-Q8_0.gguf"

	// llamafile binary (Cosmopolitan Libc single-file executable). One URL for ALL platforms.
	LlamafileURL = "https://github.com/mozilla-Ocho/llamafile/releases/download/0.10.6/llamafile-0.10.6"

	LlamaServerPort = 18080
)

func EnsureGemmaModelFile(ctx context.Context, events driven.EventPort) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache dir: %w", err)
	}

	modelPath := filepath.Join(cacheDir, "barnowl-ai", "models", "gemma-3-270m-it-Q8_0.gguf")

	if _, err := os.Stat(modelPath); err == nil {
		return modelPath, nil
	}

	if err := os.MkdirAll(filepath.Dir(modelPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directories for model: %w", err)
	}

	var lastProgress int
	callback := func(progress float32) {
		current := int(progress)
		if current > lastProgress {
			lastProgress = current
			if events != nil {
				events.Emit("on_download_progress", map[string]interface{}{
					"component": "Gemma 3 270M Model",
					"progress":  progress,
				})
			}
		}
	}

	log.Printf("[Classifier] Downloading Gemma 3 270M model (~500MB)...")
	if err := system.DownloadFileAtomic(ctx, GemmaModelURL, modelPath, callback); err != nil {
		log.Printf("[Classifier] Primary download failed (%v), trying fallback...", err)
		if err := system.DownloadFileAtomic(ctx, GemmaModelFallbackURL, modelPath, callback); err != nil {
			return "", fmt.Errorf("failed to download Gemma model: %w", err)
		}
	}

	return modelPath, nil
}

func EnsureLlamaServerBinary(ctx context.Context, events driven.EventPort) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache dir: %w", err)
	}

	// Llamafile is a polyglot binary. Windows expects a .exe extension to run it properly.
	binName := "llamafile"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(cacheDir, "barnowl-ai", "bin", binName)

	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	if err := os.MkdirAll(filepath.Dir(binPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directories for binary: %w", err)
	}

	var lastProgress int
	callback := func(progress float32) {
		current := int(progress)
		if current > lastProgress {
			lastProgress = current
			if events != nil {
				events.Emit("on_download_progress", map[string]interface{}{
					"component": "llama-server",
					"progress":  progress,
				})
			}
		}
	}

	log.Printf("[Classifier] Downloading llamafile binary (cross-platform)...")
	if err := system.DownloadFileAtomic(ctx, LlamafileURL, binPath, callback); err != nil {
		return "", fmt.Errorf("failed to download llamafile binary: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(binPath, 0755); err != nil {
			return "", fmt.Errorf("failed to make llama-server executable: %w", err)
		}
	}

	return binPath, nil
}

type LlamaServerProcess struct {
	cmd       *exec.Cmd
	port      int
	modelPath string
	mu        sync.Mutex
	started   bool
}

var DefaultLlamaServer = &LlamaServerProcess{port: LlamaServerPort}

func (p *LlamaServerProcess) Start(ctx context.Context, events driven.EventPort) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return nil
	}

	modelPath, err := EnsureGemmaModelFile(ctx, events)
	if err != nil {
		return err
	}
	p.modelPath = modelPath

	binPath, err := EnsureLlamaServerBinary(ctx, events)
	if err != nil {
		return err
	}

	p.cmd = exec.CommandContext(ctx, binPath, "--model", p.modelPath, "--port", fmt.Sprintf("%d", p.port), "--host", "127.0.0.1", "--ctx-size", "2048", "--threads", "2", "--no-mmap", "-ngl", "0")
	p.cmd.Stdout = log.Writer()
	p.cmd.Stderr = log.Writer()

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start llama-server: %w", err)
	}

	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", p.port)
	client := http.Client{Timeout: 1 * time.Second}
	ready := false

	for i := 0; i < 60; i++ { // 30 seconds max
		resp, err := client.Get(healthURL)
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				ready = true
				resp.Body.Close()
				break
			}
			resp.Body.Close()
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !ready {
		// Cannot call p.Stop() here while holding the lock because Stop() also tries to acquire it.
		if p.cmd != nil && p.cmd.Process != nil {
			p.cmd.Process.Kill()
			p.cmd.Wait()
		}
		return fmt.Errorf("llama-server failed to become ready within 30 seconds")
	}

	p.started = true
	return nil
}

func (p *LlamaServerProcess) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
		p.cmd.Wait()
	}
	p.started = false
}
