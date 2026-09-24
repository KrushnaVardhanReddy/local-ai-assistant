package engine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"wails-app/backend"
	"wails-app/backend/filter"
	"wails-app/backend/llm"
	"wails-app/core/ports/driven"
)

// ProcessAudio runs a PCM audio chunk through the full pipeline:
// audio → STT transcription → smart filter → (cache lookup) → LLM stream.
// All results are delivered via the EventPort asynchronously.
func (e *StealthEngine) ProcessAudio(samples []float32) error {
	return e.ProcessAudioTagged(samples, "")
}

// ProcessAudioTagged processes audio and associates the transcript with a speaker tag.
func (e *StealthEngine) ProcessAudioTagged(samples []float32, speaker string) error {
	ch, err := e.sttManager.TranscribeStream(samples)
	if err != nil {
		return fmt.Errorf("STT error: %w", err)
	}
	go func() {
		for transcript := range ch {
			e.mu.RLock()
			isTranscriptMode := e.isTranscriptMode
			e.mu.RUnlock()

			if isTranscriptMode {
				cleanTranscript := strings.TrimSpace(transcript)
				if cleanTranscript == "" ||
					cleanTranscript == "[BLANK_AUDIO]" ||
					cleanTranscript == " [BLANK_AUDIO]" ||
					strings.Contains(cleanTranscript, "[MUSIC]") ||
					strings.Contains(cleanTranscript, "[INAUDIBLE]") {
					continue
				}

				taggedLine := cleanTranscript
				if speaker != "" {
					taggedLine = fmt.Sprintf("%s: %s", speaker, cleanTranscript)
				}

				e.AppendTranscriptLog(taggedLine)

				if e.events != nil {
					e.events.Emit("on_transcript_log", map[string]interface{}{"text": taggedLine})
				}
			} else {
				e.handleTranscriptTagged(transcript, true, speaker)
			}
		}
	}()
	return nil
}

// AskQuestion bypasses audio and sends a typed question to the LLM pipeline.
func (e *StealthEngine) AskQuestion(question string) error {
	e.handleTranscriptTagged(question, false, "")
	return nil
}

// handleTranscriptTagged contains the core pipeline: filter → cache → LLM.
// This is the logic extracted from app.go SetAudioDevice (lines 444–551).
func (e *StealthEngine) handleTranscriptTagged(raw string, isAuto bool, speaker string) {
	if raw == "" ||
		raw == "[BLANK_AUDIO]" ||
		raw == " [BLANK_AUDIO]" ||
		strings.Contains(raw, "[MUSIC]") ||
		strings.Contains(raw, "[INAUDIBLE]") {
		return
	}

	cleanTranscript := strings.TrimSpace(raw)
	if speaker != "" {
		cleanTranscript = fmt.Sprintf("%s: %s", speaker, cleanTranscript)
	}

	emb := backend.GenerateEmbedding(cleanTranscript)

	e.mu.RLock()
	rawMode := e.rawMode
	e.mu.RUnlock()

	var isNoise bool
	filterRes := filter.Check(cleanTranscript, emb)
	if !filterRes.ShouldSend {
		isNoise = true
		if rawMode {
			log.Printf("[RAW MODE] Would have filtered, but bypassing for: %q", cleanTranscript)
			// In raw mode we process it anyway, so we don't treat it as dropped for logic flow,
			// though we might still want to tag it. We'll keep isNoise true so the UI can style it,
			// but we won't return early.
		}
	}

	if isNoise && !rawMode && isAuto {
		log.Printf("🎤 STT OUTPUT (FILTERED): %q\n", cleanTranscript)
		if e.events != nil {
			e.events.Emit("on_transcript", map[string]interface{}{"text": cleanTranscript, "is_noise": true})
			e.events.Emit("on_chip", map[string]interface{}{
				"id":       fmt.Sprintf("%d", time.Now().UnixNano()),
				"text":     cleanTranscript,
				"is_noise": true,
			})
		}
		return
	}

	log.Printf("🎤 STT OUTPUT: %q\n", cleanTranscript)

	// Update state for frontend polling
	e.mu.Lock()
	e.transcriptBuffer = append(e.transcriptBuffer, cleanTranscript)
	if len(e.transcriptBuffer) > 5 {
		e.transcriptBuffer = e.transcriptBuffer[len(e.transcriptBuffer)-5:]
	}
	e.transcript = cleanTranscript
	e.response = ""
	e.thinking = false
	e.mu.Unlock()

	// Also emit via EventsEmit (belt-and-suspenders)
	if e.events != nil {
		e.events.Emit("on_transcript", map[string]interface{}{"text": cleanTranscript, "is_noise": isNoise})
	}

	e.mu.RLock()
	manual := e.manualMode
	e.mu.RUnlock()

	// Always emit suggestion chip for accepted transcripts
	if e.events != nil {
		e.events.Emit("on_chip", map[string]interface{}{
			"id":       fmt.Sprintf("%d", time.Now().UnixNano()),
			"text":     cleanTranscript,
			"is_noise": isNoise,
		})
	}

	e.mu.RLock()
	isMockMode := e.isMockMode
	isTTSPlaying := e.isTTSPlaying
	e.mu.RUnlock()

	if isTTSPlaying {
		log.Printf("🔇 [STT] Ignoring transcript because TTS is playing: %q", cleanTranscript)
		return
	}

	if isAuto {
		if manual {
			log.Printf("🎛️ [Manual Mode] Transcript accepted but LLM call bypassed: %q", cleanTranscript)
			return
		}
		if isMockMode {
			log.Printf("🎭 [Mock Mode] Transcript accepted but auto-submit bypassed: %q", cleanTranscript)
			if e.questionBuffer != nil {
				e.questionBuffer.AddChunk(cleanTranscript)
			}
			return
		}
		if e.questionBuffer != nil {
			e.questionBuffer.AddChunk(cleanTranscript)
		}
	} else {
		// Manual query bypassing buffer (typed follow-up or chip submit — always allowed)
		e.triggerLLMWithQuestion(cleanTranscript)
	}
}

func (e *StealthEngine) triggerLLMWithQuestion(cleanTranscript string) {
	emb := backend.GenerateEmbedding(cleanTranscript)

	e.inFlightMu.Lock()
	if e.cancelInFlight != nil {
		log.Printf("[PREEMPT] Interrupting active LLM stream for new question: %q", cleanTranscript)
		e.cancelInFlight()
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.cancelInFlight = cancel
	e.inFlightCtx = ctx
	e.inFlightMu.Unlock()

	if e.cache != nil {
		cachedAns, hit := e.cache.Search(emb, 0.88)
		if hit {
			log.Printf("[Cache] Hit (similarity >= 0.88): %q", cleanTranscript)
			e.mu.Lock()
			e.response = cachedAns
			e.thinking = false
			e.mu.Unlock()
			if e.events != nil {
				e.events.Emit("on_response_start", nil)
				e.events.Emit("on_response_end", nil)
			}
			e.inFlightMu.Lock()
			if e.inFlightCtx == ctx {
				e.cancelInFlight = nil
				e.inFlightCtx = nil
			}
			e.inFlightMu.Unlock()

			if e.sessionMgr != nil {
				e.sessionMgr.StartTurn(cleanTranscript)
				e.sessionMgr.SetAISuggestion(cachedAns)
				e.sessionMgr.CompleteTurn()
			}

			cancel()
			return
		}
		log.Printf("[Cache] Miss: %q", cleanTranscript)
	}

	e.mu.Lock()
	e.thinking = true
	e.response = ""
	e.mu.Unlock()
	if e.events != nil {
		e.events.Emit("on_response_start", nil)
	}

	if e.sessionMgr != nil {
		e.sessionMgr.StartTurn(cleanTranscript)
	}

	activeDocBlock := ""
	if e.GetIncludeActiveDocContext() {
		e.workspaceMu.RLock()
		doc := e.activeDoc
		var docName, docContent string
		if doc != nil {
			docName = doc.Name
			docContent = doc.Content
		}
		e.workspaceMu.RUnlock()

		if docContent != "" {
			if len(docContent) > 10000 {
				docContent = docContent[:10000]
			}
			activeDocBlock = fmt.Sprintf("\n\n[ACTIVE WORKSPACE DOCUMENT: %s]\n%s\n[/ACTIVE WORKSPACE DOCUMENT]", docName, docContent)
		}
	}

	ragContextBlock := ""
	if e.cache != nil {
		chunks, err := e.cache.SemanticSearch(emb, 3, 0.75) // Limit to top 3, with 0.75 threshold
		if err == nil && len(chunks) > 0 {
			var sb strings.Builder
			sb.WriteString("\n\n[WORKSPACE RAG CONTEXT]\n")
			sb.WriteString("The following are excerpts from workspace documents relevant to the question:\n\n")
			for i, chunk := range chunks {
				sb.WriteString(fmt.Sprintf("--- Excerpt %d ---\n%s\n", i+1, chunk))
			}
			sb.WriteString("[/WORKSPACE RAG CONTEXT]")
			ragContextBlock = sb.String()
		}
	}

	llmQuestion := cleanTranscript

	go func(q string, llmQuestion string, streamCtx context.Context, streamCancel context.CancelFunc) {
		e.llmBusy.Lock()
		defer e.llmBusy.Unlock()

		defer func() {
			e.inFlightMu.Lock()
			if e.inFlightCtx == streamCtx {
				e.cancelInFlight = nil
				e.inFlightCtx = nil
			}
			e.inFlightMu.Unlock()
			streamCancel()
		}()

		// If cancelled before lock acquired, abort early
		if streamCtx.Err() != nil {
			return
		}

		var answerBuilder strings.Builder

		var history []driven.ChatMessage
		if e.sessionMgr != nil {
			recentTurns := e.sessionMgr.GetRecentTurns(3)
			for _, t := range recentTurns {
				history = append(history, driven.ChatMessage{Role: "user", Content: t.InterviewerQuestion})
				history = append(history, driven.ChatMessage{Role: "assistant", Content: t.AISuggestion})
			}
		}

		e.mu.RLock()
		isMockMode := e.isMockMode
		e.mu.RUnlock()

		sysPrompt := e.cfg.SystemPrompt
		if isMockMode {
			sysPrompt = llm.MockInterviewerPrompt
		}

		if activeDocBlock != "" {
			sysPrompt += activeDocBlock
		}
		if ragContextBlock != "" {
			sysPrompt += ragContextBlock
		}

		if e.sessionMgr != nil {
			recentTurns := e.sessionMgr.GetRecentTurns(llm.MaxContextTurns)
			contextBlock := llm.BuildContextBlock(recentTurns)
			if contextBlock != "" {
				sysPrompt += "\n\n" + contextBlock
			}
		}

		err := e.llm.StreamCompletion(streamCtx, llmQuestion, sysPrompt, history, func(token string) {
			answerBuilder.WriteString(token)
			e.mu.Lock()
			e.response = answerBuilder.String()
			e.thinking = true
			e.mu.Unlock()

			if e.sessionMgr != nil {
				e.sessionMgr.SetAISuggestion(answerBuilder.String())
			}
			if e.events != nil {
				e.events.Emit("on_response_token", map[string]interface{}{"text": token})
			}
		}, func() {
			e.mu.Lock()
			e.thinking = false
			e.mu.Unlock()

			if e.sessionMgr != nil {
				e.sessionMgr.CompleteTurn()
			}
			if e.events != nil {
				e.events.Emit("on_response_end", nil)
			}
		})

		// If cancelled during streaming, log and abort without saving to cache
		if streamCtx.Err() != nil {
			log.Printf("[PREEMPT] LLM streaming aborted for: %q", q)
			return
		}

		if err != nil {
			log.Printf("LLM streaming failed: %v", err)
		} else {
			finalAns := answerBuilder.String()
			if e.cache != nil && len(finalAns) > 0 {
				if storeErr := e.cache.Store(q, finalAns); storeErr != nil {
					log.Printf("[Cache] Error storing Q&A pair: %v", storeErr)
				} else {
					log.Printf("[Cache] ✅ Successfully stored Q&A pair in background")
				}
			}
			log.Printf("[LLM] Stream complete. Stored in cache.")

			e.mu.RLock()
			mockActive := e.isMockMode
			ttsActive := e.isMockTTS
			e.mu.RUnlock()

			if mockActive && ttsActive && e.ttsAdapter != nil {
				go func(text string) {
					log.Printf("🔊 [TTS] Speaking response...")
					e.mu.Lock()
					e.isTTSPlaying = true
					e.mu.Unlock()

					defer func() {
						e.mu.Lock()
						e.isTTSPlaying = false
						e.mu.Unlock()
					}()

					if err := e.ttsAdapter.Speak(text); err != nil {
						log.Printf("❌ [TTS] Failed to speak: %v", err)
					}
				}(finalAns)
			}
		}
	}(cleanTranscript, llmQuestion, ctx, cancel)
}

func (e *StealthEngine) triggerLLMWithCustomPrompt(prompt string, input string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if e.events != nil {
		e.events.Emit("on_response_start", nil)
	}

	var answerBuilder strings.Builder

	err := e.llm.StreamCompletion(ctx, input, prompt, nil, func(token string) {
		answerBuilder.WriteString(token)
		if e.events != nil {
			e.events.Emit("on_response_token", map[string]interface{}{"text": token})
		}
	}, func() {
		if e.events != nil {
			e.events.Emit("on_response_end", nil)
		}
	})

	if err != nil {
		log.Printf("Custom LLM streaming failed: %v", err)
	} else {
		log.Printf("[LLM] Custom stream complete.")
	}
}

func (e *StealthEngine) SummarizeTranscript() error {
	lines := e.FlushTranscriptLog()
	if len(lines) == 0 {
		return fmt.Errorf("no transcript to summarize")
	}

	fullTranscript := strings.Join(lines, "\n")

	go e.triggerLLMWithCustomPrompt(llm.TranscriptSummaryPrompt, fullTranscript)

	return nil
}
