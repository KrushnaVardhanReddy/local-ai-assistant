1.  **Create `wails-app/core/engine/engine.go`**:
    *   Define `Config`, `StealthEngine`.
    *   Implement `New()`, `GetState()`, `ClearState()`, `SetStealth()`, `SetEventsAdapter()`.
    *   Ensure compile-time check for `driving.PipelinePort`.
2.  **Create `wails-app/core/engine/pipeline.go`**:
    *   Implement `ProcessAudio()`, `AskQuestion()`, `handleTranscript()`.
    *   Refactor the pipeline logic from `app.go` (`SetAudioDevice`) into `handleTranscript()`.
3.  **Modify `wails-app/app.go`**:
    *   Add `engine *engine.StealthEngine` to `App`.
    *   Update `NewApp()` to instantiate `StealthEngine` and pass it to `App`.
    *   Update `startup(ctx)` to set the `eventsadapter.NewWailsEventAdapter(ctx)` on the engine.
    *   Refactor `SetAudioDevice()` to call `a.engine.ProcessAudio()`.
    *   Refactor `GetState()` to delegate to `a.engine.GetState()`.
    *   Remove redundant fields from `App` (`latestTranscript`, etc.).
4.  **Create `wails-app/core/engine/engine_test.go`**:
    *   Write 100% test coverage for `engine.go` and `pipeline.go` using mock adapters.
5.  **Pre-commit steps**:
    *   Run tests (`go test -v ./core/engine/...` and `go test -v ./...`).
    *   Run `go build ./...`.
    *   Ensure proper testing, verification, review, and reflection are done using `pre_commit_instructions`.
