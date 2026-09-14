package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"wails-app/backend/hotkeys"
	"wails-app/backend/stt"
	"wails-app/backend/audio"
	"wails-app/backend/remote"
	"net/http"
	"wails-app/backend/system"
	"wails-app/backend/window"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {

	remoteServer *remote.Server
	ctx context.Context

	backendCmd *exec.Cmd
	cmdMutex   sync.Mutex

	sttManager   *stt.STTManager
	audioCapture *audio.CaptureEngine
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Initialize default STT engine if model path is provided via env var or exists.
	// For now we'll just try to load a default path or leave it nil if it fails.
	var initialEngine stt.STTEngine
	modelPath := os.Getenv("WHISPER_MODEL_PATH")
	if modelPath == "" {
		// Try several candidate paths relative to the working directory
		candidates := []string{
			"../models/ggml-tiny.bin",
			"../models/ggml-base.en.bin",
			"models/ggml-tiny.bin",
			"models/ggml-base.en.bin",
		}
		for _, candidate := range candidates {
			if _, statErr := os.Stat(candidate); statErr == nil {
				modelPath = candidate
				break
			}
		}
	}
	if modelPath == "" {
		log.Println("⚠️  No Whisper model found! Set WHISPER_MODEL_PATH env var. STT will be disabled.")
	} else {
		log.Printf("🧠 Loading Whisper model from: %s\n", modelPath)
	}
	whisperEngine, err := stt.LoadWhisperEngine(modelPath)
	if err == nil {
		initialEngine = whisperEngine
		log.Println("✅ Whisper model loaded successfully!")
	} else {
		log.Printf("❌ Failed to load Whisper model: %v\n", err)
	}

	captureEngine := audio.NewCaptureEngine()
	// Ignore init errors since hardware might not be present.
	_ = captureEngine.Initialize()

	return &App{
		remoteServer: remote.NewServer(http.FS(assets)),
		sttManager:   stt.NewSTTManager(initialEngine),
		audioCapture: captureEngine,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Hide from taskbar for maximum stealth
	if err := window.HideFromTaskbar(ctx); err != nil {
		log.Printf("Failed to hide from taskbar: %v\n", err)
	}

	// Initialize hotkeys Wails runtime dependencies
	hotkeys.WindowGetPosition = wailsruntime.WindowGetPosition
	hotkeys.WindowSetPosition = wailsruntime.WindowSetPosition
	hotkeys.EventsEmit = wailsruntime.EventsEmit

	// Start hotkeys
	if err := hotkeys.Start(ctx); err != nil {
		log.Printf("Failed to start hotkeys (this is expected in CI without a display): %v\n", err)
	}

	// Start background download for STT models
	if a.remoteServer != nil {
		a.remoteServer.Start()
	}
	go system.StartBackgroundDownload(ctx, func(progress float32) {
		// Example: Emit progress event to frontend
		wailsruntime.EventsEmit(ctx, "download_progress", progress)
	})

	// Auto-start default microphone capture
	go func() {
		if a.sttManager == nil || a.audioCapture == nil {
			log.Println("⚠️  STT manager or audio capture not ready, skipping auto-start")
			return
		}
		log.Println("🎙️ Auto-starting default microphone capture...")
		err := a.SetAudioDevice(-1, false)
		if err != nil {
			log.Printf("❌ Failed to auto-start microphone: %v\n", err)
		} else {
			log.Println("✅ Microphone auto-started successfully!")
		}
	}()
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

func (a *App) GetAudioDevices() []audio.AudioDevice {
	devices, err := a.audioCapture.GetDevices()
	if err != nil {
		log.Printf("Failed to get audio devices: %v\n", err)
		return []audio.AudioDevice{}
	}
	return devices
}

func (a *App) SetAudioDevice(id int, isLoopback bool) error {
	err := a.audioCapture.StartCapture(id, isLoopback, func(samples []float32) {
		ch, err := a.sttManager.TranscribeStream(samples)
		if err != nil {
			log.Printf("Failed to transcribe stream: %v\n", err)
			return
		}

		go func() {
			for transcript := range ch {
				if transcript != "" &&
					transcript != "[BLANK_AUDIO]" &&
					transcript != " [BLANK_AUDIO]" &&
					!strings.Contains(transcript, "[MUSIC]") &&
					!strings.Contains(transcript, "[INAUDIBLE]") {
					log.Printf("🎤 STT OUTPUT: %q\n", transcript)
					payload := map[string]interface{}{
						"text": strings.TrimSpace(transcript),
					}
					wailsruntime.EventsEmit(a.ctx, "on_transcript", payload)
				}
			}
		}()
	})

	if err != nil {
		log.Printf("Failed to set audio device %d: %v\n", id, err)
		return err
	}

	log.Printf("Successfully started capturing device %d (loopback: %v)\n", id, isLoopback)
	return nil
}

func (a *App) shutdown(ctx context.Context) {
	if a.audioCapture != nil {
		a.audioCapture.Terminate()
	}
	if a.remoteServer != nil {
		_ = a.remoteServer.Stop(ctx)
	}
}
