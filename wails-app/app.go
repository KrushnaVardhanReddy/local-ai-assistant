package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"wails-app/backend/hotkeys"
	"wails-app/backend/stt"
	"wails-app/backend/system"
	"wails-app/backend/window"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context

	backendCmd *exec.Cmd
	cmdMutex   sync.Mutex

	sttManager *stt.STTManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Initialize default STT engine if model path is provided via env var or exists.
	// For now we'll just try to load a default path or leave it nil if it fails.
	var initialEngine stt.STTEngine
	modelPath := os.Getenv("WHISPER_MODEL_PATH")
	if modelPath == "" {
		modelPath = "models/ggml-base.en.bin"
	}
	engine, err := stt.LoadWhisperEngine(modelPath)
	if err == nil {
		initialEngine = engine
	}

	return &App{
		sttManager: stt.NewSTTManager(initialEngine),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize hotkeys Wails runtime dependencies
	hotkeys.WindowGetPosition = wailsruntime.WindowGetPosition
	hotkeys.WindowSetPosition = wailsruntime.WindowSetPosition
	hotkeys.EventsEmit = wailsruntime.EventsEmit

	// Start hotkeys
	if err := hotkeys.Start(ctx); err != nil {
		log.Printf("Failed to start hotkeys (this is expected in CI without a display): %v\n", err)
	}

	// Start background download for STT models
	go system.StartBackgroundDownload(ctx, func(progress float32) {
		// Example: Emit progress event to frontend
		wailsruntime.EventsEmit(ctx, "download_progress", progress)
	})
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) StartBackend() error {
	// Legacy python spawning logic has been removed.
	return nil
}

func (a *App) StopBackend() error {
	// Legacy python spawning logic has been removed.
	return nil
}

func (a *App) GetMachineId() string {
	return "wails-machine-id"
}

func (a *App) LoadToken() string {
	return ""
}

func (a *App) SaveToken(token map[string]interface{}) {
	// Not implemented
}

func (a *App) DeleteToken() {
	// Not implemented
}

func (a *App) QuitApp() {
	wailsruntime.Quit(a.ctx)
}

func (a *App) CaptureScreen() string {
	return ""
}

func (a *App) SetClickthrough(opts map[string]interface{}) {
	if enable, ok := opts["enable"].(bool); ok {
		wailsruntime.WindowSetAlwaysOnTop(a.ctx, enable)
		window.SetIgnoreMouseEvents(a.ctx, enable)
	}
}

func (a *App) ToggleStealth(opts map[string]interface{}) {
	// Not implemented
}
