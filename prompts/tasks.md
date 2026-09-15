# 🤖 BarnOwl AI — Jules Task Tracker

> Submit tasks to Jules one-by-one:
> ```bash
> python3 scripts/jules_submit.py --task P1-T1
> python3 scripts/jules_submit.py --list   # see all tasks
> ```
> Update status and PR numbers here as Jules returns PRs.

## Legend
| Symbol | Meaning |
|
---
## Phase 37 — Embedded Stealth Terminal

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P37-T1 | `frontend/src-tauri/tauri.conf.json`, `frontend/src/lib/StealthTerminal.svelte` | **Embedded UI Terminal:** Run backend silently within Tauri via Sidecar/Shell plugin | ✅ | #116 |
| P37-T2 | `backend/app.py`, `frontend/src/lib/Settings.svelte` | **Model Status API & Display:** Add an endpoint to expose the active STT/LLM configurations and display them in the Settings UI. | ✅ | #117 |
| P37-T3 | `backend/requirements.txt`, `Makefile` | **Optimize Build Size:** Switch to CPU-only PyTorch and exclude massive unused modules in PyInstaller to radically shrink the `.exe`. | ⬜ | — |

---
## Phase 38 — V2 Native Rust Migration (Future Pipeline)

*Note: This phase is incredibly complex (est. 4-6 weeks) and should only be started once the Python-based V1 is fully shipped and scaled.*

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P38-T1 | `src-tauri/Cargo.toml` | **Rust AI Environment:** Integrate `whisper-rs` (C++ bindings) and `ort` (Microsoft ONNX Runtime) to handle local ML models natively. | ⬜ | — |
| P38-T2 | `src-tauri/src/main.rs` | **API to Tauri IPC:** Rewrite all Python FastAPI HTTP/WebSocket endpoints as native Tauri Commands (`#[tauri::command]`) for zero IPC overhead. | ⬜ | — |
| P38-T3 | `src-tauri/src/vector_db.rs` | **Vector DB Migration:** Replace Python `chromadb` with a Rust-native embedded alternative (e.g., Qdrant or `sqlite-vec`). | ⬜ | — |
| P38-T4 | `src-tauri/src/ml_routing.rs` | **Intent Router Migration:** Export the Scikit-learn `question_classifier.pkl` to ONNX and rewrite the tensor math pipeline in Rust. | ⬜ | — |



---

## Phase 46 — Hybrid Auto-Detect STT Engine 🎤

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P46-T1 | `stt/engine.go`, `whisper.go` | **STT Interface Abstraction:** Refactor the existing Whisper setup into a clean `STTEngine` interface and implement a core STT Manager. | ✅ | #150 |
| P46-T2 | `system/profiler.go`, `downloader.go` | **Hardware Profiler & Downloader:** Build a startup routine that checks for 8GB RAM + AVX2 and downloads Parakeet/sherpa-onnx DLLs from GitHub if supported. | ✅ | #151 |
| P46-T3 | `stt/parakeet.go` | **Parakeet CGO & Hot-Swap:** Implement `sherpa-onnx` bindings and hot-swap logic to seamlessly switch from Whisper to Parakeet mid-stream when downloaded. | ⬜ | — |

---

## Phase 47 — Remote Helper Mode (Wails) 🌐

> Rebuilding the Remote Helper Mode (originally P7-T5) for the new Wails architecture. Allows a friend to access the UI remotely via Cloudflare tunnel.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P47-T1-T2 | `wails-app/backend/app.go` | **Embedded Server & WS Bridge:** Spin up an HTTP server on port 8000 alongside Wails and expose a `/ws` route to mirror Wails IPC events to the remote browser. | ✅ | #157 |
| P47-T3 | `scripts/start_remote.sh` | **Update Tunnel Script:** Refactor the existing script to point the Cloudflare tunnel directly to the new embedded Go HTTP port (8000). | ✅ | #155 |

---

## Phase 48 — Automated E2E Testing 🧪

> Implement a "Split E2E" testing strategy to validate both the Svelte frontend and the Go local ML backend.

| Task ID | File(s) | Description | Status | PR |
| P48-T1 | `wails-app/frontend/playwright.config.ts`, `tests/` | **Split E2E UI Tests:** Setup Playwright in the frontend to test the Svelte UI flows while mocking the backend Go calls. | ✅ | #158 |

---

## Phase 49 — Global Hotkeys & Window Management ⌨️

> Implementing global OS-level hotkeys so the user can interact with and move the transparent Wails window without needing a mouse.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P49-T1 | `wails-app/backend/hotkeys.go` | **Global Hotkey Movement:** Use `golang.design/x/hotkey` to register global hotkeys (e.g. `Ctrl+Alt+Right`) that nudge the Wails window coordinates. | ✅ | #156 |

---

## Phase 50 — Go Native Audio Loopback 🎧

> Re-implementing the driverless system audio loopback feature for the Go backend so users can capture the interviewer's voice natively (e.g. Windows WASAPI).

| Task ID | File(s) | Description | Status | PR |
| P50-T1 | `wails-app/backend/stt/`, `app.go` | **Native Audio Loopback:** Upgrade Go audio capture to support Windows WASAPI loopback and expose device selector bindings to Wails. | ✅ | #159 |

---

## Phase 51 — Go STT Pipeline (VAD + Smart Filter + Cache + LLM) 🎤🧠

> Porting the complete Python audio intelligence pipeline into the Go backend:
> Silence-based VAD, smart filter (filler/embedding), QA cache (sqlite-vec), and OpenAI streaming.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P51-T1 | `backend/filter/`, `backend/llm/`, `backend/audio/capture.go`, `backend/vector_db.go`, `app.go`, `frontend/ws.svelte.ts` | **Full STT Pipeline:** VAD silence buffer, smart filter (filler + Nomic embedding noise detection), sqlite-vec QA cache lookup, OpenAI streaming LLM with busy guard. | ⬜ | — |
| P51-T2 | `backend/filter/classifier.go`, `backend/filter/filter.go` | **Native Go Intent Classifier:** Replace manual questionWords heuristic with an ML-based Nearest Centroid classifier in Go based on embeddings. | ⬜ | — |

---

## Phase 52 — Interview Intelligence (Prompts, Multi-Turn Context & Scorecard) 🧠📊

> Porting the interview intelligence system from `feature/krushna`:
> Categorized system prompts (STAR, coding, system design), multi-turn context (last 2-3 turns), thread-safe session manager, and end-of-session AI scorecard generator.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P52-T1 | `backend/llm/prompts.go`, `backend/llm/openai.go` | **Categorized Prompts & Multi-Turn Context:** Specialized prompt injections (STAR, coding, system design) and 2–3 turn context window injection. | ✅ | #165 |
| P52-T2 | `backend/session/session.go`, `backend/llm/scorecard.go`, `backend/remote/server.go`, `app.go` | **Session Manager & Scorecard:** Thread-safe session tracking, post-interview JSON scorecard evaluation, and `/session/end` endpoint. | ✅ | #166 |