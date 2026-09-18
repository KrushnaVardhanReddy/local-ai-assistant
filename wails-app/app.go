package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	cacheadapter "wails-app/adapters/cache"
	eventsadapter "wails-app/adapters/events"
	llmadapter "wails-app/adapters/llm"
	"wails-app/backend"
	"wails-app/backend/audio"
	"wails-app/backend/hotkeys"
	"wails-app/backend/llm"
	"wails-app/backend/remote"
	"wails-app/backend/session"
	"wails-app/backend/stt"
	"wails-app/backend/system"
	"wails-app/backend/window"
	"wails-app/core/engine"
	"wails-app/core/ports/driving"

	"github.com/kbinani/screenshot"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	remoteServer *remote.Server
	ctx          context.Context

	backendCmd *exec.Cmd
	cmdMutex   sync.Mutex

	sttManager   *stt.STTManager
	audioCapture *audio.CaptureEngine

	engine *engine.StealthEngine

	isClickthrough bool
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
			"../models/ggml-base.en.bin",
			"../models/ggml-tiny.bin",
			"models/ggml-base.en.bin",
			"models/ggml-tiny.bin",
		}
		for _, candidate := range candidates {
			if _, statErr := os.Stat(candidate); statErr == nil {
				modelPath = candidate
				break
			}
		}
	}
	sttProvider := os.Getenv("STT_PROVIDER")
	if sttProvider == "groq" {
		groqKey := os.Getenv("GROQ_API_KEY")
		groqModel := os.Getenv("STT_MODEL")
		initialEngine = stt.NewGroqEngine(groqKey, groqModel)
		log.Printf("☁️ Using Groq Cloud STT Engine (Model: %s)\n", groqModel)
	} else {
		validPath, err := stt.EnsureWhisperModel(modelPath)
		if err != nil {
			log.Printf("⚠️  Could not ensure Whisper model: %v. STT will be disabled.", err)
		} else {
			log.Printf("🧠 Loading local Whisper model from: %s\n", validPath)
			whisperEngine, err := stt.LoadWhisperEngine(validPath)
			if err == nil {
				initialEngine = whisperEngine
				log.Println("✅ Whisper model loaded successfully!")
			} else {
				log.Printf("❌ Failed to load Whisper model: %v\n", err)
			}
		}
	}

	captureEngine := audio.NewCaptureEngine()
	// Ignore init errors since hardware might not be present.
	_ = captureEngine.Initialize()

	err := os.MkdirAll("./data", 0755)
	if err != nil {
		log.Printf("Failed to create data directory: %v", err)
	}
	db, err := backend.NewVectorDB("./data/qa_cache.db")
	if err != nil {
		log.Printf("Failed to init QA cache DB: %v", err)
	}

	sessMgr := session.NewSessionManager()

	eng := engine.New(
		engine.Config{SystemPrompt: llm.DefaultSystemPrompt},
		stt.NewSTTManager(initialEngine),
		llmadapter.NewOpenAIAdapter(),
		cacheadapter.NewSQLiteVecAdapter(db),
		nil,
	)

	return &App{
		remoteServer: remote.NewServer(http.FS(assets), sessMgr),
		sttManager:   stt.NewSTTManager(initialEngine),
		audioCapture: captureEngine,
		engine:       eng,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.engine.SetEventsAdapter(eventsadapter.NewWailsEventAdapter(ctx))

	// Hide from taskbar for maximum stealth
	if err := window.HideFromTaskbar(ctx); err != nil {
		log.Printf("Failed to hide from taskbar: %v\n", err)
	}

	// Initialize hotkeys Wails runtime dependencies
	hotkeys.WindowGetPosition = wailsruntime.WindowGetPosition
	hotkeys.WindowSetPosition = wailsruntime.WindowSetPosition
	hotkeys.EventsEmit = wailsruntime.EventsEmit
	hotkeys.ToggleClickthrough = func(ctx context.Context) {
		a.ToggleClickthroughMode()
	}

	// Preload the ONNX embedding model during startup so it doesn't log on first microphone input
	go backend.InitEmbeddings()

	// Start hotkeys
	if err := hotkeys.Start(ctx); err != nil {
		log.Printf("Failed to start hotkeys (this is expected in CI without a display): %v\n", err)
	}

	// Start background download for STT models
	if a.remoteServer != nil {
		a.remoteServer.Start()
	}

	sttProvider := os.Getenv("STT_PROVIDER")
	if sttProvider == "parakeet" {
		go system.StartBackgroundDownload(ctx, func(progress float32) {
			// Example: Emit progress event to frontend
			wailsruntime.EventsEmit(ctx, "download_progress", progress)
		})
	}

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

// GetState is polled by the frontend every 200ms to get the latest transcript/response state.
func (a *App) GetState() map[string]interface{} {
	s := a.engine.GetState()
	return map[string]interface{}{
		"transcript":             s.Transcript,
		"response":               s.Response,
		"thinking":               s.Thinking,
		"cached_pairs":           s.CachedPairs,
		"estimated_tokens_saved": s.CachedPairs * 250,
	}
}

// ClearState resets the current transcript, AI response, and thinking flags.
func (a *App) ClearState() {
	a.engine.ClearState()
}

// GetCacheStats returns the number of cached Q&A pairs and estimated tokens saved
func (a *App) GetCacheStats() map[string]interface{} {
	s := a.engine.GetState()
	return map[string]interface{}{
		"cached_pairs":           s.CachedPairs,
		"estimated_tokens_saved": s.CachedPairs * 250,
	}
}

// GetCacheItems returns all cached items for UI management
func (a *App) GetCacheItems() []backend.CacheItem {
	cacheAdapter, ok := a.engine.GetCache().(interface {
		GetAllItems() ([]backend.CacheItem, error)
	})
	if !ok {
		return []backend.CacheItem{}
	}
	items, err := cacheAdapter.GetAllItems()
	if err != nil {
		return []backend.CacheItem{}
	}
	return items
}

// DeleteCacheItems removes specified question IDs from the cache
func (a *App) DeleteCacheItems(ids []string) error {
	cacheAdapter, ok := a.engine.GetCache().(interface{ DeleteItem(string) error })
	if !ok {
		return nil
	}
	for _, id := range ids {
		_ = cacheAdapter.DeleteItem(id)
	}
	return nil
}

// ClearCache clears all cached Q&A pairs
func (a *App) ClearCache() error {
	cacheAdapter, ok := a.engine.GetCache().(interface{ ClearAll() error })
	if !ok {
		return nil
	}
	return cacheAdapter.ClearAll()
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
	log.Println("📸 [Go] CaptureScreen requested...")
	n := screenshot.NumActiveDisplays()
	log.Printf("📸 [Go] Active displays count: %d\n", n)
	if n <= 0 {
		log.Println("❌ [Go] CaptureScreen: No active displays found")
		return ""
	}

	bounds := screenshot.GetDisplayBounds(0)
	log.Printf("📸 [Go] Display 0 bounds: %+v\n", bounds)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		log.Printf("❌ [Go] CaptureScreen error: %v\n", err)
		return ""
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("❌ [Go] CaptureScreen PNG encode error: %v\n", err)
		return ""
	}

	// Save screenshot to disk in ./data/screenshots
	if err := os.MkdirAll("./data/screenshots", 0755); err == nil {
		fileName := fmt.Sprintf("./data/screenshots/snip_%d.png", time.Now().UnixNano())
		if err := os.WriteFile(fileName, buf.Bytes(), 0644); err == nil {
			log.Printf("💾 [Go] Screenshot saved to disk: %s\n", fileName)
		}
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	log.Printf("✅ [Go] Screen captured successfully! PNG size: %d bytes (base64 len: %d)\n", buf.Len(), len(encoded))
	return "data:image/png;base64," + encoded
}

// AnalyzeVision receives a base64 screenshot data URI and streams the LLM vision response to the frontend UI
func (a *App) AnalyzeVision(base64Image string, prompt string) error {
	log.Println("🤖 [Go] AnalyzeVision starting LLM completion for screenshot...")

	if a.ctx != nil {
		wailsruntime.EventsEmit(a.ctx, "on_response_start", nil)
	}

	go func() {
		var answerBuilder strings.Builder
		ctx := a.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		err := llm.StreamVisionCompletion(ctx, base64Image, prompt, func(token string) {
			answerBuilder.WriteString(token)
			a.engine.UpdateState("📸 [Screenshot Snip Captured]", answerBuilder.String(), true)
			if a.ctx != nil {
				wailsruntime.EventsEmit(a.ctx, "on_response_token", map[string]interface{}{"text": token})
			}
		}, func() {
			a.engine.UpdateState("📸 [Screenshot Snip Captured]", answerBuilder.String(), false)
			if a.ctx != nil {
				wailsruntime.EventsEmit(a.ctx, "on_response_end", nil)
			}
		})

		if err != nil {
			log.Printf("❌ [Go] AnalyzeVision failed: %v", err)
			a.engine.UpdateState("📸 [Screenshot Snip Captured]", fmt.Sprintf("Vision analysis error: %v", err), false)
		} else {
			log.Println("✅ [Go] AnalyzeVision stream complete!")
		}
	}()

	return nil
}

func (a *App) ToggleClickthroughMode() bool {
	a.cmdMutex.Lock()
	a.isClickthrough = !a.isClickthrough
	curr := a.isClickthrough
	a.cmdMutex.Unlock()

	log.Printf("🖱️ [Go] ToggleClickthroughMode -> enable = %v\n", curr)
	wailsruntime.WindowSetAlwaysOnTop(a.ctx, true)
	wailsruntime.WindowShow(a.ctx)
	if err := window.SetIgnoreMouseEvents(a.ctx, curr); err != nil {
		log.Printf("❌ SetIgnoreMouseEvents error: %v\n", err)
	}
	wailsruntime.EventsEmit(a.ctx, "toggle-clickthrough", curr)
	return curr
}

func (a *App) SetClickthrough(enable bool) {
	a.cmdMutex.Lock()
	a.isClickthrough = enable
	a.cmdMutex.Unlock()

	log.Printf("🖱️ [Go] SetClickthrough -> enable = %v\n", enable)
	wailsruntime.WindowSetAlwaysOnTop(a.ctx, true)
	wailsruntime.WindowShow(a.ctx)
	if err := window.SetIgnoreMouseEvents(a.ctx, enable); err != nil {
		log.Printf("❌ SetIgnoreMouseEvents error: %v\n", err)
	}
	wailsruntime.EventsEmit(a.ctx, "toggle-clickthrough", enable)
}

func (a *App) ToggleStealth(opts map[string]interface{}) {
	a.ToggleClickthroughMode()
}

func (a *App) EndSession() (map[string]interface{}, error) {
	if a.engine.GetSessionManager() == nil {
		return nil, fmt.Errorf("session manager not configured")
	}

	sessionData := a.engine.GetSessionManager().Export()
	turnCount, _ := sessionData["turn_count"].(int)
	if turnCount == 0 {
		return map[string]interface{}{
			"session":   sessionData,
			"scorecard": nil,
		}, nil
	}

	scorecard, err := llm.GenerateScorecard(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate scorecard: %w", err)
	}

	return map[string]interface{}{
		"session":   sessionData,
		"scorecard": scorecard,
	}, nil
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
		_ = a.engine.ProcessAudio(samples)
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

// PromptOpenDirectory opens a native folder dialog and loads it as a workspace
func (a *App) PromptOpenDirectory() (path string, err error) {
	if a.ctx == nil {
		return "", fmt.Errorf("no valid wails context")
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("runtime error: %v", r)
		}
	}()
	dirpath, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Open Workspace Folder",
	})
	if err != nil || dirpath == "" {
		return "", err
	}
	_, err = a.engine.OpenDirectory(dirpath)
	if err != nil {
		return "", err
	}
	return dirpath, nil
}

// PromptOpenFile opens a native file dialog and opens the file in the workspace
func (a *App) PromptOpenFile() (path string, err error) {
	if a.ctx == nil {
		return "", fmt.Errorf("no valid wails context")
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("runtime error: %v", r)
		}
	}()
	filepath, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Open Document",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "All Supported Files", Pattern: "*.txt;*.md;*.pdf;*.pptx;*.py;*.go;*.ts;*.js;*.jsx;*.tsx;*.svelte;*.json;*.rs;*.cpp;*.c;*.h;*.html;*.css;*.yaml;*.yml;*.sql;*.sh"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil || filepath == "" {
		return "", err
	}
	_, err = a.engine.OpenFile(filepath)
	if err != nil {
		return "", err
	}
	return filepath, nil
}

// OpenDirectory calls engine.OpenDirectory
func (a *App) OpenDirectory(path string) (*driving.FileNode, error) {
	return a.engine.OpenDirectory(path)
}

// OpenFile calls engine.OpenFile
func (a *App) OpenFile(path string) (*driving.WorkspaceDocument, error) {
	return a.engine.OpenFile(path)
}

// CloseDocument calls engine.CloseFile
func (a *App) CloseDocument(path string) error {
	return a.engine.CloseFile(path)
}

// SetActiveDocument calls engine.SetActiveDocument
func (a *App) SetActiveDocument(path string) error {
	_, err := a.engine.SetActiveDocument(path)
	return err
}

// GetWorkspaceTree calls engine.GetWorkspaceTree
func (a *App) GetWorkspaceTree() []*driving.FileNode {
	return a.engine.GetWorkspaceTree()
}

// GetOpenDocuments calls engine.GetOpenDocuments
func (a *App) GetOpenDocuments() []*driving.WorkspaceDocument {
	return a.engine.GetOpenDocuments()
}

// GetActiveDocument calls engine.GetActiveDocument
func (a *App) GetActiveDocument() *driving.WorkspaceDocument {
	return a.engine.GetActiveDocument()
}

// SetIncludeActiveDocContext sets whether the active document is included as LLM context
func (a *App) SetIncludeActiveDocContext(enabled bool) {
	a.engine.SetIncludeActiveDocContext(enabled)
}

// GetIDEState returns the current IDE state for the frontend
func (a *App) GetIDEState() map[string]interface{} {
	activeDoc := a.engine.GetActiveDocument()
	activeDocPath := ""
	activeDocContent := ""
	activeDocName := ""
	if activeDoc != nil {
		activeDocPath = activeDoc.Path
		activeDocContent = activeDoc.Content
		activeDocName = activeDoc.Name
	}

	isListening := false
	if a.audioCapture != nil {
		isListening = a.audioCapture.IsCapturing()
	}

	s := a.engine.GetState()

	a.cmdMutex.Lock()
	clickthrough := a.isClickthrough
	a.cmdMutex.Unlock()

	return map[string]interface{}{
		"workspaceTree":           a.engine.GetWorkspaceTree(),
		"openDocuments":           a.engine.GetOpenDocuments(),
		"activeDocumentPath":      activeDocPath,
		"activeDocumentContent":   activeDocContent,
		"activeDocumentName":      activeDocName,
		"isListening":             isListening,
		"isClickthrough":          clickthrough,
		"cachedPairs":             s.CachedPairs,
		"includeActiveDocContext": a.engine.GetIncludeActiveDocContext(),
	}
}

// ToggleMic toggles the microphone capturing state.
// Returns true if the microphone was started, false if it was stopped.
func (a *App) ToggleMic() bool {
	if a.audioCapture == nil {
		return false
	}
	if a.audioCapture.IsCapturing() {
		a.audioCapture.StopCapture()
		return false
	} else {
		a.SetAudioDevice(-1, false)
		return true
	}
}
