package engine

import (
	"fmt"
	"log"
	"strings"
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
			e.handleTranscript(transcript)
		}
	}()
	return nil
}

// AskQuestion bypasses audio and sends a typed question to the LLM pipeline.
func (e *StealthEngine) AskQuestion(question string) error {
	e.handleTranscript(question)
	return nil
}

// handleTranscript contains the core pipeline: filter → cache → LLM.
// This is the logic extracted from app.go SetAudioDevice (lines 444–551).
func (e *StealthEngine) handleTranscript(raw string) {
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

	if !e.llmBusy.TryLock() {
		log.Printf("[BUSY] Discarded (LLM streaming): %q", cleanTranscript)
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
			e.llmBusy.Unlock()
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

	go func(q string) {
		defer e.llmBusy.Unlock()
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

		err := e.llm.StreamCompletion(q, sysPrompt, history, func(token string) {
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
	}(cleanTranscript)
}
