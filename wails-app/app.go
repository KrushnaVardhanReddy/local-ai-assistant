package main

import (
	"context"
	"fmt"
	"log"
	"wails-app/backend/session"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"bytes"
	"encoding/base64"
	"image/png"
	"wails-app/backend"
	"wails-app/backend/audio"
	"wails-app/backend/filter"
	"wails-app/backend/hotkeys"
	"wails-app/backend/llm"
	"wails-app/backend/remote"
	"wails-app/backend/stt"
	"wails-app/backend/system"
	"wails-app/backend/window"

	"github.com/kbinani/screenshot"
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

	qaCache *backend.VectorDB
	llmBusy sync.Mutex

	// UI state — polled by frontend via GetState()
	stateMu          sync.RWMutex
	latestTranscript string
	latestResponse   string
	latestThinking   bool

	sessionManager *session.SessionManager
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
		if modelPath == "" {
			log.Println("⚠️  No Whisper model found! Set WHISPER_MODEL_PATH env var. STT will be disabled.")
		} else {
			log.Printf("🧠 Loading local Whisper model from: %s\n", modelPath)
			whisperEngine, err := stt.LoadWhisperEngine(modelPath)
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

	return &App{
		remoteServer: remote.NewServer(http.FS(assets), sessMgr),
		sttManager:   stt.NewSTTManager(initialEngine),
		audioCapture: captureEngine,
		qaCache:      db,
		sessionManager: sessMgr,
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

// GetState is polled by the frontend every 200ms to get the latest transcript/response state.
func (a *App) GetState() map[string]interface{} {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()

	var count int
	if a.qaCache != nil {
		count = a.qaCache.GetCount()
	}

	return map[string]interface{}{
		"transcript":             a.latestTranscript,
		"response":               a.latestResponse,
		"thinking":               a.latestThinking,
		"cached_pairs":           count,
		"estimated_tokens_saved": count * 250,
	}
}

// ClearState resets the current transcript, AI response, and thinking flags.
func (a *App) ClearState() {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	a.latestTranscript = ""
	a.latestResponse = ""
	a.latestThinking = false
}

// GetCacheStats returns the number of cached Q&A pairs and estimated tokens saved
func (a *App) GetCacheStats() map[string]interface{} {
	if a.qaCache == nil {
		return map[string]interface{}{
			"cached_pairs":           0,
			"estimated_tokens_saved": 0,
		}
	}
	count := a.qaCache.GetCount()
	return map[string]interface{}{
		"cached_pairs":           count,
		"estimated_tokens_saved": count * 250,
	}
}

// GetCacheItems returns all cached items for UI management
func (a *App) GetCacheItems() []backend.CacheItem {
	if a.qaCache == nil {
		return []backend.CacheItem{}
	}
	items, err := a.qaCache.GetAllItems()
	if err != nil {
		return []backend.CacheItem{}
	}
	return items
}

// DeleteCacheItems removes specified question IDs from the cache
func (a *App) DeleteCacheItems(ids []string) error {
	if a.qaCache == nil {
		return nil
	}
	for _, id := range ids {
		_ = a.qaCache.DeleteItem(id)
	}
	return nil
}

// ClearCache clears all cached Q&A pairs
func (a *App) ClearCache() error {
	if a.qaCache == nil {
		return nil
	}
	return a.qaCache.ClearAll()
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

	a.stateMu.Lock()
	a.latestTranscript = "📸 [Screenshot Snip Captured]"
	a.latestResponse = ""
	a.latestThinking = true
	a.stateMu.Unlock()

	wailsruntime.EventsEmit(a.ctx, "on_response_start", nil)

	go func() {
		var answerBuilder strings.Builder
		err := llm.StreamVisionCompletion(base64Image, prompt, func(token string) {
			answerBuilder.WriteString(token)
			a.stateMu.Lock()
			a.latestResponse = answerBuilder.String()
			a.latestThinking = true
			a.stateMu.Unlock()
			wailsruntime.EventsEmit(a.ctx, "on_response_token", map[string]interface{}{"text": token})
		}, func() {
			a.stateMu.Lock()
			a.latestThinking = false
			a.stateMu.Unlock()
			wailsruntime.EventsEmit(a.ctx, "on_response_end", nil)
		})

		if err != nil {
			log.Printf("❌ [Go] AnalyzeVision failed: %v", err)
			a.stateMu.Lock()
			a.latestResponse = fmt.Sprintf("Vision analysis error: %v", err)
			a.latestThinking = false
			a.stateMu.Unlock()
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

func (a *App) SetClickthrough(opts map[string]interface{}) {
	if enable, ok := opts["enable"].(bool); ok {
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
}

func (a *App) ToggleStealth(opts map[string]interface{}) {
	a.ToggleClickthroughMode()
}

func (a *App) EndSession() (map[string]interface{}, error) {
	if a.sessionManager == nil {
		return nil, fmt.Errorf("session manager not configured")
	}

	sessionData := a.sessionManager.Export()
	turnCount, _ := sessionData["turn_count"].(int)
	if turnCount == 0 {
		return map[string]interface{}{
			"session": sessionData,
			"scorecard": nil,
		}, nil
	}

	scorecard, err := llm.GenerateScorecard(sessionData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate scorecard: %w", err)
	}

	return map[string]interface{}{
		"session": sessionData,
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

					cleanTranscript := strings.TrimSpace(transcript)
					emb := backend.GenerateEmbedding(cleanTranscript)

					filterRes := filter.Check(cleanTranscript, emb)
					if !filterRes.ShouldSend {
						continue
					}

					log.Printf("🎤 STT OUTPUT: %q\n", transcript)

					// Update state for frontend polling
					a.stateMu.Lock()
					a.latestTranscript = cleanTranscript
					a.latestResponse = ""
					a.latestThinking = false
					a.stateMu.Unlock()

					// Also emit via EventsEmit (belt-and-suspenders)
					wailsruntime.EventsEmit(a.ctx, "on_transcript", map[string]interface{}{"text": cleanTranscript})

					if !a.llmBusy.TryLock() {
						log.Printf("[BUSY] Discarded (LLM streaming): %q", cleanTranscript)
						continue
					}

					if a.qaCache != nil {
						cachedAns, hit := a.qaCache.SearchByEmbedding(emb, 0.88)
						if hit {
							log.Printf("[Cache] Hit (similarity >= 0.88): %q", cleanTranscript)
							a.stateMu.Lock()
							a.latestResponse = cachedAns
							a.latestThinking = false
							a.stateMu.Unlock()
							wailsruntime.EventsEmit(a.ctx, "on_response_start", nil)
							wailsruntime.EventsEmit(a.ctx, "on_response_end", nil)
							a.llmBusy.Unlock()
							continue
						}
						log.Printf("[Cache] Miss: %q", cleanTranscript)
					}

					a.stateMu.Lock()
					a.latestThinking = true
					a.latestResponse = ""
					a.stateMu.Unlock()
					wailsruntime.EventsEmit(a.ctx, "on_response_start", nil)

					if a.sessionManager != nil {
						a.sessionManager.StartTurn(cleanTranscript)
					}

					go func(q string) {
						var answerBuilder strings.Builder

						var history []llm.ChatMessage
						if a.sessionManager != nil {
							recentTurns := a.sessionManager.GetRecentTurns(3)
							for _, t := range recentTurns {
								history = append(history, llm.ChatMessage{Role: "user", Content: t.Transcript})
								history = append(history, llm.ChatMessage{Role: "assistant", Content: t.Response})
							}
						}

						err := llm.StreamCompletionWithContext(q, "", history, func(token string) {
							answerBuilder.WriteString(token)
							a.stateMu.Lock()
							a.latestResponse = answerBuilder.String()
							a.latestThinking = true
							a.stateMu.Unlock()

							if a.sessionManager != nil {
								a.sessionManager.AppendToken(token)
							}
							wailsruntime.EventsEmit(a.ctx, "on_response_token", map[string]interface{}{"text": token})
						}, func() {
							a.stateMu.Lock()
							a.latestThinking = false
							a.stateMu.Unlock()

							if a.sessionManager != nil {
								a.sessionManager.CompleteTurn()
							}
							wailsruntime.EventsEmit(a.ctx, "on_response_end", nil)
						})
						if err != nil {
							log.Printf("LLM streaming failed: %v", err)
						} else {
							finalAns := answerBuilder.String()
							if a.qaCache != nil && len(finalAns) > 0 {
								go func(questionText, answerText string) {
									if storeErr := a.qaCache.Store(questionText, answerText); storeErr != nil {
										log.Printf("[Cache] Error storing Q&A pair: %v", storeErr)
									} else {
										log.Printf("[Cache] ✅ Successfully stored Q&A pair in background")
									}
								}(q, finalAns)
							}
							log.Printf("[LLM] Stream complete. Stored in cache.")
						}
						a.llmBusy.Unlock()
					}(cleanTranscript)
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
