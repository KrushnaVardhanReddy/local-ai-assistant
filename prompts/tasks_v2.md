# 🏗️ BarnOwl AI v2 — Jules Task Tracker (Hexagonal Engine)

> Submit tasks to Jules one-by-one or in parallel batches:
> ```bash
> python3 scripts/jules_submit.py --task P54-T1
> python3 scripts/jules_submit.py --list   # see all tasks
> ```
> Update status and PR numbers here as Jules returns PRs.

## Legend
| Symbol | Meaning |
|--------|---------|
| ✅ | Done / PR merged |
| 🔄 | Jules working on it |
| ⬜ | Not started |
| ⚡ | Can run in PARALLEL with other ⚡ tasks in same batch |

---

## Phase 64 — Launch Prep & Distribution 🚀

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P64-T1 | `wails-app/wails.json`, `Makefile` | **App Packaging & Build** — Configure Wails to produce production-ready binaries: `.dmg` (macOS), `.exe` (Windows), and `.AppImage` (Linux). Add app icons, metadata, and cross-compilation scripts to the Makefile. | ✅ | — |
| P64-T2 | `scripts/deploy_supabase.sh` | **Supabase Production Setup Scripts** — Create SQL migration files for the required tables (`dev_allowlist`, `user_entitlements`) and a deployment script (`scripts/deploy_supabase.sh`) for the Edge Function. | ✅ | — |
| P64-T3 | `supabase/functions/paddle-webhook/index.ts` | **Paddle Billing Webhook Setup** — Create a Supabase Edge Function to process Paddle webhook events (`transaction.completed`, `subscription.activated`, etc). Verify the webhook signature using `PADDLE_WEBHOOK_SECRET`. Update the `user_entitlements` table with `paddle_subscription_id` and `plan_type` based on the payload. | ✅ | — |
| P64-T4 | Paddle Dashboard (Manual) | **Paddle Sandbox Setup** — Create the "BarnOwl AI Lifetime" product and SaaS products. Configure the webhook destination. | ✅ | — |
| P64-T5 | `wails-app/frontend/src/lib/auth.svelte.ts`, `wails-app/frontend/tests/entitlement.spec.ts` | **Entitlement E2E Tests** — Expose `authState` to Playwright via `window.__authState` and write UI E2E tests for Developer Mode, Active License, Usage Meter calculations, Overage warnings, and Stripe subscription status. | 🔄 | — |
| P64-T7 ⚡ | `wails-app/frontend/index.html`, `wails-app/frontend/src/lib/components/AuthModal.svelte` | **Frontend Paddle Checkout Integration** — Embed Paddle.js directly in the frontend HTML. Update AuthModal's "Buy Lifetime License" button to invoke `Paddle.Checkout.open()` and inject `customData: { user_id: authState.user.id }` so that webhook transactions can be tied to the correct user. | ✅ | — |

---
| P64-T6 | `wails-app/main.go`, `wails-app/frontend/src/lib/components/AuthModal.svelte` | **Auth Window UX Redesign** — Modify Wails app to start with `AlwaysOnTop: false` and `BackgroundColour` set to a solid color when in Auth mode. Add a custom draggable HTML titlebar to `AuthModal.svelte` with minimize, maximize, and close buttons that call Wails runtime functions. Toggle `AlwaysOnTop` and background color dynamically when authentication succeeds. | ✅ | #228 |

---

## Phase 69 — Alternative Local Models (Research & Evaluation) 🔬

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|-----|
| P69-T1 | `wails-app/backend/classifier/gemma.go` | **Evaluate Apertus Mini 0.5B** — Test the Apertus Mini 0.5B (Swiss AI) model as an alternative to `gemma-3-270m` for the VAD turn-detection sidecar. Compare RAM usage, battery drain, and accuracy in detecting end-of-turn for non-English speakers and heavy accents. | ⬜ | — |

---

## Phase 70 — Mock Interview Mode (Go Port) 🎤

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|-----|
| P70-T1 | `backend/tts/`, `backend/engine.go`, `frontend/` | **Port Mock Interview Mode to Go:** Re-implement the Phase 21 Mock Interview feature in the Go/Wails architecture. Add a TTS adapter (wrapping edge-tts), flip the LLM prompt to "Interviewer", and add the UI toggle in IDEShell. | ✅ | #240 |

## Phase 71 — CI/CD Pipeline Fixes 🛠️

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P71-T1 | `.github/workflows/*` | **Fix CGO Compilation in CI:** Add C++ toolchains to Windows runners and inject Whisper CGO flags into all build steps. | ✅ | #241 |

## Phase 64-E2E — Full Application Test Suite Split 🧪

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P64-T5a | `workspace.spec.ts` | **Core IDE E2E:** Test file tree, code editor, and tabs. | ✅ | — |
| P64-T5b | `copilot.spec.ts` | **Assistant UI E2E:** Test STAR method presets and drawer. | ✅ | — |
| P64-T5c | `cache.spec.ts` | **Cache E2E:** Test cache manager and RAG stats. | ✅ | — |
| P64-T5d | `mock_interview.spec.ts` | **Mock Interview E2E:** Test persona switch and UI update. | ✅ | — |
| P64-T5e | `entitlement.spec.ts` | **Billing E2E:** Test dev mode, lifetime, and SaaS states. | ✅ | — |

## Phase 72 — Mock Interview Polish 🎤

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P72-T1 ⚡ | `wails-app/frontend/src/products/interview/ConvPanel.svelte`, `wails-app/frontend/src/products/interview/InterviewHUD.svelte` | **Mock Mode Submit Button:** Disable VAD-based auto-submission in mock mode. Add a "Submit Answer" button so the user controls when the response is evaluated. | ✅ | #248 |
| P72-T2 ⚡ | `wails-app/backend/llm/prompts.go`, `wails-app/core/engine/pipeline.go` | **Mock Interview Persona:** Create a distinct system prompt for Mock Mode where the LLM evaluates the user's response, gives brief constructive feedback, and asks the next relevant follow-up question. | ✅ | #247 |
| P72-T3 ⚡ | `wails-app/backend/tts/`, `wails-app/frontend/src/` | **Interactive Audio TTS:** Integrate TTS (Text-to-Speech) so the mock interviewer reads its questions out loud, making the experience more immersive. | ✅ | #249 |
| P72-T4 | `wails-app/**/*_test.go` | **Fix Golang Unit Tests:** Resolve all compilation and runtime errors across the Golang test suite. Address CGO dependency issues (like whisper.h) by using build tags or proper interface mocking. | 🔄 | — |
| P72-T5 | `wails-app/backend/classifier/buffer.go` | **Fix Context Timeout:** Resolve the `context deadline exceeded` error by increasing the Gemma classification timeout to 10s and clearing the buffer on failure. (Completed locally) | ✅ | — |

## Phase 73 — Dual Mode Audio (Interview vs Granolah Mode) 🎙️

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P73-T1 ⚡ | `backend/audio/dual_capture.go`, `app.go`, `pipeline.go` | **Dual Capture Backend:** Create a new `AppMode` system. In Interview Mode, default to single loopback capture. In Transcript Mode, spin up a `DualCaptureEngine` that captures both Loopback and Mic simultaneously and tags transcripts with `[Interviewer]` and `[Candidate]`. | ✅ | #251 |
| P73-T2 ⚡ | `frontend/.../InterviewHUD.svelte`, `ws.svelte.ts` | **Mode Switcher UI:** Add a frontend toggle in the ActivityBar to switch between Interview Mode and Transcript Mode. In Transcript mode, hide the auto-submit controls and display a "Summarize Session" button. | ✅ | #253 |
| P73-T3 ⚡ | `engine.go`, `pipeline.go`, `app.go` | **Transcript Summarization:** Route tagged dual-audio transcripts into a dedicated `transcriptLog`. Wire the "Summarize Session" button to flush this log to the LLM with a Granolah-style meeting summary system prompt, streaming the notes back to the UI. | ✅ | #254 |
| P73-T4 | `backend/classifier/buffer.go`, `core/engine/engine.go` | **Remove Gemma Classifier:** Remove the llamafile/Gemma-3 local LLM sidecar entirely. Replace `IsQuestionComplete` with an instant always-true function. Delete `gemma.go`, `gemma_test.go`, and `llama_server.go`. Eliminates 5-8s startup, ~600MB RAM, and CPU contention that caused `context deadline exceeded` and dropped questions. | ✅ | #256 |
| P73-T5 | `backend/filter/classifier.go`, `centroids.json` | **O(1) Centroid Embedding:** Shift ONNX centroid calculation to compile-time. Remove the hardcoded dataset map and runtime embedding generation from `initCentroids`. Instead, load pre-calculated 768-dim float arrays instantly via `//go:embed centroids.json`. This allows using massive HuggingFace datasets (10,000+ questions) for training without any startup penalty. | ✅ | #257 |
| P73-T6 | `centroids.json` | **Expand Classifier Dataset:** Use Jules to dynamically write a python script containing ~350 highly diverse interview questions and conversational noise, calculate the embeddings via `nomic-embed-text-v1.5`, and overwrite `centroids.json`. Ensures robust accuracy across all programming languages. | 🔄 | — |

## Phase 74 — Screen-Share Invisibility (OS Level) 👻

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P74-T1 | `backend/window/`, `adapters/window/`, `app.go` | **OS-Level Screen-Share Exclusion:** Implement true OS-level screen capture invisibility using `SetWindowDisplayAffinity` (Windows), `NSWindowSharingNone` (macOS), and X11 compositor hints (Linux) so the app remains invisible during accidental screen shares. | ✅ | #259 |

## Phase 75 — UX Polish (Analytics & Glanceability) ✨

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P75-T1 | `ScorecardModal.svelte`, `InterviewHUD.svelte`, `prompts.go` | **Post-Interview Analytics & Talking Points:** Build a beautiful Scorecard Modal to visualize the JSON analytics generated by `EndSession()`. Modify the LLM prompt to strictly return 3-5 glanceable "Talking Point" bullets instead of dense paragraphs. | 🔄 | — |
