1. **wails-app/core/engine/engine.go**:
   - Add `lastResponseAt time.Time` to `StealthEngine` struct.

2. **wails-app/core/engine/pipeline.go**:
   - In `handleTranscript`:
     - If it's a manual trigger (`!isAuto`), set `e.lastResponseAt = time.Time{}` (zero it out) so manual trigger isn't blocked by cooldown.
     - Add `ignoreClassifier := time.Now().Before(e.lastResponseAt)` inside `handleTranscript` (or just pass `time.Now().Before(e.lastResponseAt)` to `AddChunk`). Wait, the instructions say to add `IsCooldownActive` to `QuestionBuffer` or just use `AddChunkSilently`, or modify `AddChunk(chunk string, ignoreClassifier bool)`.
     - Modifying `AddChunk` seems easiest. Let's modify `QuestionBuffer.AddChunk(chunk string, ignoreClassifier bool)`.
   - In `triggerLLMWithQuestion`, after LLM streams completely (at the end where it stores in cache and finalAns is complete):
     - Compute words: `wordCount := len(strings.Fields(finalAns))`
     - Compute cooldown duration: `cooldown := time.Duration(wordCount) * 300 * time.Millisecond`
     - Clamp to 3s-20s.
     - Set `e.mu.Lock(); e.lastResponseAt = time.Now().Add(cooldown); e.mu.Unlock()`.
     - Make sure manual triggers clear `lastResponseAt` in `AskQuestion` or `triggerLLMWithQuestion`.

3. **wails-app/backend/classifier/buffer.go**:
   - Update `AddChunk(chunk string)` to `AddChunk(chunk string, ignoreClassifier bool)`.
   - If `ignoreClassifier` is true, we just append to `b.chunks`, and do NOT launch the goroutine to run `b.classifyFn`.
   - Also, increase the context timeout from 3s to 10s: `ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)`.

4. **wails-app/backend/classifier/buffer_test.go**:
   - Update tests calling `AddChunk` to pass `false` for `ignoreClassifier`.

5. **wails-app/core/engine/pipeline_test.go** and **wails-app/core/engine/engine_test.go**:
   - Fix tests if they break due to changes, write coverage for cooldown calculation.

6. **Update Wiki (`wails-app/wiki/`)**:
   - Update `architecture.md` or similar to document the reading cooldown logic.

7. **Pre-commit Instructions**:
   - Call `pre_commit_instructions` tool to make sure tests pass.
