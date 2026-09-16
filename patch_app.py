import sys
with open('wails-app/app.go', 'r') as f:
    content = f.read()

# 1. Update imports
if '"wails-app/core/engine"' not in content:
    content = content.replace(
        '"wails-app/backend/window"',
        '"wails-app/backend/window"\n\t"wails-app/core/engine"\n\t"wails-app/adapters/llmadapter"\n\t"wails-app/adapters/cacheadapter"\n\t"wails-app/adapters/eventsadapter"'
    )

# 2. Update App struct
content = content.replace(
    '''	qaCache *backend.VectorDB
	llmBusy sync.Mutex

	// UI state — polled by frontend via GetState()
	stateMu          sync.RWMutex
	latestTranscript string
	latestResponse   string
	latestThinking   bool
''',
    '''	engine *engine.StealthEngine
'''
)

# 3. Update NewApp instantiation
new_app_replace = '''	eng := engine.New(
		engine.Config{SystemPrompt: llm.DefaultSystemPrompt},
		stt.NewSTTManager(initialEngine),
		llmadapter.NewOpenAIAdapter(),
		cacheadapter.NewSQLiteVecAdapter(db),
		nil,
	)

	return &App{
		remoteServer:   remote.NewServer(http.FS(assets), sessMgr),
		sttManager:     stt.NewSTTManager(initialEngine),
		audioCapture:   captureEngine,
		engine:         eng,
	}'''
content = content.replace(
    '''	return &App{
		remoteServer:   remote.NewServer(http.FS(assets), sessMgr),
		sttManager:     stt.NewSTTManager(initialEngine),
		audioCapture:   captureEngine,
		qaCache:        db,
		sessionManager: sessMgr,
	}''',
    new_app_replace
)

# 4. Update startup(ctx)
content = content.replace(
    '''func (a *App) startup(ctx context.Context) {
	a.ctx = ctx''',
    '''func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.engine.SetEventsAdapter(eventsadapter.NewWailsEventAdapter(ctx))'''
)

# 5. Thin out SetAudioDevice
set_audio_device_original = '''func (a *App) SetAudioDevice(id int, isLoopback bool) error {
	err := a.audioCapture.StartCapture(id, isLoopback, func(samples []float32) {
		ch, err := a.sttManager.TranscribeStream(samples)
		if err != nil {
			log.Printf("Failed to transcribe stream: %v\\n", err)
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

					log.Printf("🎤 STT OUTPUT: %q\\n", transcript)

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
		log.Printf("Failed to set audio device %d: %v\\n", id, err)
		return err
	}

	log.Printf("Successfully started capturing device %d (loopback: %v)\\n", id, isLoopback)
	return nil
}'''

set_audio_device_new = '''func (a *App) SetAudioDevice(id int, isLoopback bool) error {
	err := a.audioCapture.StartCapture(id, isLoopback, func(samples []float32) {
		_ = a.engine.ProcessAudio(samples)
	})

	if err != nil {
		log.Printf("Failed to set audio device %d: %v\\n", id, err)
		return err
	}

	log.Printf("Successfully started capturing device %d (loopback: %v)\\n", id, isLoopback)
	return nil
}'''
content = content.replace(set_audio_device_original, set_audio_device_new)

# 6. Thin out GetState
get_state_original = '''func (a *App) GetState() map[string]interface{} {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()

	cachedPairs := 0
	if a.qaCache != nil {
		cachedPairs = a.qaCache.Count()
	}

	return map[string]interface{}{
		"transcript":             a.latestTranscript,
		"response":               a.latestResponse,
		"thinking":               a.latestThinking,
		"cached_pairs":           cachedPairs,
		"estimated_tokens_saved": cachedPairs * 250,
	}
}'''
get_state_new = '''func (a *App) GetState() map[string]interface{} {
	s := a.engine.GetState()
	return map[string]interface{}{
		"transcript":             s.Transcript,
		"response":               s.Response,
		"thinking":               s.Thinking,
		"cached_pairs":           s.CachedPairs,
		"estimated_tokens_saved": s.CachedPairs * 250,
	}
}'''
content = content.replace(get_state_original, get_state_new)

# Remove a.sessionManager usage in captureScreen?
# The instructions do not say to modify it. Wait, app.go might still need session manager? No, "Remove the now-redundant fields: latestTranscript, latestResponse, latestThinking, stateMu, llmBusy, qaCache, sessionManager"
# Need to check if a.sessionManager is used elsewhere.
with open('wails-app/app.go', 'w') as f:
    f.write(content)
