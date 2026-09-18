package engine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"wails-app/backend"
	"wails-app/backend/filter"
	"wails-app/core/ports/driven"
)

// ProcessAudio runs a PCM audio chunk through the full pipeline:
// audio → STT transcription → smart filter → (cache lookup) → LLM stream.
// All results are delivered via the EventPort asynchronously.
func (e *StealthEngine) ProcessAudio(samples []float32) error {
	ch, err := e.sttManager.TranscribeStream(samples)
	if err != nil {
		return fmt.Errorf("STT error: %w", err)
	}
	go func() {
		for transcript := range ch {
			e.handleTranscript(transcript, true)
		}
	}()
	return nil
}

// AskQuestion bypasses audio and sends a typed question to the LLM pipeline.
func (e *StealthEngine) AskQuestion(question string) error {
	e.handleTranscript(question, false)
	return nil
}

// handleTranscript contains the core pipeline: filter → cache → LLM.
// This is the logic extracted from app.go SetAudioDevice (lines 444–551).
func (e *StealthEngine) handleTranscript(raw string, isAuto bool) {
	if raw == "" ||
		raw == "[BLANK_AUDIO]" ||
		raw == " [BLANK_AUDIO]" ||
		strings.Contains(raw, "[MUSIC]") ||
		strings.Contains(raw, "[INAUDIBLE]") {
		return
	}

	cleanTranscript := strings.TrimSpace(raw)
	emb := backend.GenerateEmbedding(cleanTranscript)

	filterRes := filter.Check(cleanTranscript, emb)
	if !filterRes.ShouldSend {
		return
	}

	log.Printf("🎤 STT OUTPUT: %q\n", cleanTranscript)

	// Update state for frontend polling
	e.mu.Lock()
	e.transcript = cleanTranscript
	e.response = ""
	e.thinking = false
	e.mu.Unlock()

	// Also emit via EventsEmit (belt-and-suspenders)
	if e.events != nil {
		e.events.Emit("on_transcript", map[string]interface{}{"text": cleanTranscript})
	}

	e.inFlightMu.Lock()
	if e.cancelInFlight != nil {
		log.Printf("[PREEMPT] Interrupting active LLM stream for new question: %q", cleanTranscript)
		e.cancelInFlight()
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.cancelInFlight = cancel
	e.inFlightCtx = ctx
	e.inFlightMu.Unlock()

	// Always emit suggestion chip for accepted transcripts
	if e.events != nil {
		e.events.Emit("on_chip", map[string]interface{}{
			"id":   fmt.Sprintf("%d", time.Now().UnixNano()),
			"text": cleanTranscript,
		})
	}

	e.mu.RLock()
	manual := e.manualMode
	e.mu.RUnlock()

	if isAuto && manual {
		log.Printf("🎛️ [Manual Mode] Transcript accepted but LLM call bypassed: %q", cleanTranscript)
		return
	}

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

	go func(q string, streamCtx context.Context, streamCancel context.CancelFunc) {
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

		sysPrompt := e.cfg.SystemPrompt
		if activeDocBlock != "" {
			sysPrompt += activeDocBlock
		}

		err := e.llm.StreamCompletion(streamCtx, q, sysPrompt, history, func(token string) {
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
		}
	}(cleanTranscript, ctx, cancel)
}
