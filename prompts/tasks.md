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

## Phase 41 — Cloudflare Worker LLM & Voice Integration 🗣️

> Migrates the core real-time interview loop (WebSocket connection, Auth, and LLM streaming) from the old Python backend to the Cloudflare Worker.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P41-T1 | `cloud-worker/src/index.ts` | **Worker WebSocket Upgrade:** Implement the `/ws` endpoint in the worker to accept incoming WebSocket upgrade requests and handle connection lifecycle. | ✅ | #134 |
| P41-T2 | `cloud-worker/src/index.ts` | **Supabase Session Validation:** Handle the initial `{"type": "auth"}` message over WS. Validate the API Key against Supabase and check if the user has `payg_sessions > 0`. | ✅ | #135 |
| P41-T3 | `cloud-worker/src/index.ts` | **LLM Streaming & Vector Cache Integration:** Handle the `{"type": "chat"}` message over WS. Generate embeddings, query the `pgvector` cache, stream to Groq, and stream tokens back. | ✅ | #136 |

---

## Phase 42 — The Wails Pivot (Frontend & Desktop Core) 🚀

> Replacing Tauri (Rust) with Wails (Go) for a unified single-binary architecture.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P42-T1 | `wails-app/` | **Init Wails:** Run the Wails CLI to generate a new Svelte+TS template (`wails init -n local-ai-assistant -t svelte-ts`). | ✅ | #137 |
| P42-T2 | `wails-app/frontend/src/` | **UI Porting:** Copy existing Svelte components, lib files, and Tailwind configuration into the new Wails frontend directory. Fix any imports. | ✅ | #140 |
| P42-T3 | `wails-app/frontend/src/lib/` | **API Swap:** Replace Tauri frontend calls (e.g., `invoke('command')`, `WebviewWindow`) with Wails Go bindings (`@wailsio/runtime`). | ✅ | #143 |
| P42-T4 | `wails-app/frontend/src/lib/api.ts` | **Cloud Worker Proxy (Wails Edition):** Ensure the Cloud Edition Svelte logic correctly points to the existing Cloudflare Worker URL. | ✅ | #141 |
| P42-T5 | `wails-app/frontend/src/lib/` | **Cloud Auth UI:** Implement a Login and Registration modal in Svelte that uses Supabase Auth to register users for the Cloud Edition. | ✅ | #142 |
| P42-T6 | `wails-app/tests/e2e/` | **E2E Wails UI Test:** Wails native E2E test verifying the Svelte app mounts and successfully routes API calls with zero mocking. | 🔄 | — |

---

## Phase 43 — Go Native AI Backend (Local ML) 🧠

> Re-implementing the Python local AI server in pure Go.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P43-T1 | `wails-app/backend/vector_db.go` | **Local Vector DB:** Replace ChromaDB with embedded SQLite + `sqlite-vec` extension in Go. | ✅ | #138 |
| P43-T2 | `wails-app/backend/embeddings.go` | **Embeddings & Intent:** Use `onnxruntime-go` to run all-MiniLM-L6-v2 ONNX models directly in the Go process for fast vector generation. | ✅ | #144 |
| P43-T3 | `wails-app/backend/stt.go` | **Local STT:** Integrate `whisper.go` (CGO bindings for whisper.cpp) for offline Speech-to-Text inference, replacing `faster-whisper`. | ✅ | #139 |
| P43-T4 | `wails-app/tests/e2e/` | **E2E Local ML Test:** Feed real audio through `whisper.go`, generate embeddings via `onnx`, and query `sqlite-vec` with zero mocking. | 🔄 | — |

---

## Phase 44 — Finalization & Cleanup 🧹

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P44-T1 | `frontend/`, `backend/`, `src-tauri/` | **The Great Deletion:** Safely remove the legacy Tauri frontend, Rust backend, and Python sidecar directories. | ⬜ | — |
| P44-T2 | `Makefile`, `.github/workflows/` | **Update CI Pipelines:** Switch build scripts to use `wails build` instead of `cargo tauri build` and `PyInstaller`. | ⬜ | — |
| P44-T3 | `wails-app/tests/e2e/` | **E2E Full System Test:** End-to-end Wails desktop test verifying the entire offline ML pipeline within the compiled single binary. | ⬜ | — |