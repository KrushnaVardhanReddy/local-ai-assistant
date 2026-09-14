# 🤖 Local AI Assistant — Jules Task Tracker

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

## Phase 43 — Go Native AI Backend (Local ML) 🧠

> Re-implementing the Python local AI server in pure Go.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P43-T1 | `wails-app/backend/vector_db.go` | **Local Vector DB:** Replace ChromaDB with embedded SQLite + `sqlite-vec` extension in Go. | ✅ | #138 |
| P43-T2 | `wails-app/backend/embeddings.go` | **Embeddings & Intent:** Use `onnxruntime-go` to run all-MiniLM-L6-v2 ONNX models directly in the Go process for fast vector generation. | ✅ | #144 |
| P43-T3 | `wails-app/backend/stt.go` | **Local STT:** Integrate `whisper.go` (CGO bindings for whisper.cpp) for offline Speech-to-Text inference, replacing `faster-whisper`. | ✅ | #139 |
| P43-T4 | `wails-app/tests/e2e/` | **E2E Local ML Test:** Feed real audio through `whisper.go`, generate embeddings via `onnx`, and query `sqlite-vec` with zero mocking. | 🔄 | — |
| P43-T5 | `wails-app/backend/stt_test.go` | **Go Backend Unit Tests:** Write comprehensive unit tests for the Speech-to-Text (`stt.go`) implementation and `app.go` bindings. | ✅ | #145 |


---

## Phase 45 — Advanced UI Features ✨

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P45-T1 | `wails-app/main.go` | **Native Click-Through:** Implement CGO/OS-level hooks to allow mouse clicks to pass completely through the transparent Wails window. | ⏳ | — |

---

## Phase 46 — Hybrid Auto-Detect STT Engine 🎤

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P46-T1 | `stt/engine.go`, `whisper.go` | **STT Interface Abstraction:** Refactor the existing Whisper setup into a clean `STTEngine` interface and implement a core STT Manager. | ✅ | #150 |
| P46-T2 | `system/profiler.go`, `downloader.go` | **Hardware Profiler & Downloader:** Build a startup routine that checks for 8GB RAM + AVX2 and downloads Parakeet/sherpa-onnx DLLs from GitHub if supported. | ⏳ | — |
| P46-T3 | `stt/parakeet.go` | **Parakeet CGO & Hot-Swap:** Implement `sherpa-onnx` bindings and hot-swap logic to seamlessly switch from Whisper to Parakeet mid-stream when downloaded. | ⬜ | — |

---

## Phase 47 — Remote Helper Mode (Wails) 🌐

> Rebuilding the Remote Helper Mode (originally P7-T5) for the new Wails architecture. Allows a friend to access the UI remotely via Cloudflare tunnel.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P47-T1 | `wails-app/backend/app.go` | **Embedded HTTP Server:** Spin up a lightweight Go HTTP server alongside Wails to serve the compiled Svelte frontend assets on port 8000. | ⬜ | — |
| P47-T2 | `wails-app/backend/app.go`, `wails-app/frontend/src/lib/ws.svelte.ts` | **Remote WebSocket Bridge:** Expose a WebSocket route on the Go server that mirrors the Wails IPC events (audio stream, LLM tokens) to the remote browser connection. | ⬜ | — |
| P47-T3 | `scripts/start_remote.sh` | **Update Tunnel Script:** Refactor the existing script to point the Cloudflare tunnel directly to the new embedded Go HTTP port (8000). | ⬜ | — |