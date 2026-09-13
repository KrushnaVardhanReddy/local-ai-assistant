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

---|---|---|---|---|
| P36-T2 | `backend/app.py`, `backend/local_intelligence.py` | **Smart Context Pipeline:** SmolLM2 multi-class routing + ChromaDB RAG | ✅ | #113 |

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
## Phase 39 — Dynamic Cache Pre-Warmer

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P39-T1 | `scripts/dynamic_prewarmer.py` | **Dynamic Cache Pre-Warmer:** Generate 100+ Q&A pairs using an LLM + JD + Resume and store them in ChromaDB. | ✅ | #121 |
| P39-T2 | `backend/app.py`, `backend/qa_cache.py`, `frontend/...` | **Selective Cache Management:** Add API and UI to view and selectively delete specific Q&A pairs from the cache. | ✅ | #122 |

---

## Phase 40 — Cloud Edition (Lightweight Tauri Build) ☁️

> Builds a second Tauri product from the same codebase — no local backend, all requests routed to a Cloudflare Worker.
> Target: ~10-15MB installer. **Pricing: Usage-based per interview — actual LLM API cost (pass-through) + $2 service fee per interview session.**

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P40-T1 | `frontend/src/lib/api.ts` | **Centralized API Config:** Create a single `api.ts` module with `getApiUrl()` / `getWsUrl()` helpers driven by `VITE_API_BASE` env var. Foundation for all other tasks. | ✅ | #123 |
| P40-T2 | `frontend/src/lib/*.svelte`, `frontend/src/lib/ws.svelte.ts` | **Refactor Hardcoded URLs:** Replace all ~25 hardcoded `127.0.0.1:8765` references across 6 files with the new `getApiUrl()` / `getWsUrl()` helpers. | ✅ | #124 |
| P40-T3 | `cloud-worker/` | **Cloudflare Worker Backend:** Implement a Worker that mirrors all FastAPI endpoints (`/ws`, `/api/ask`, `/api/status`, `/api/cache`, `/api/resume/context`, etc.) calling Groq/OpenAI. Auth via Cloudflare KV API keys. | ✅ | #125 |
| P40-T4 | `cloud-worker/`, `frontend/src/lib/auth.svelte.ts` | **Cloud User Auth & Session Management:** User signup → API key provisioned in Cloudflare KV → key stored in Tauri `localStorage` → sent as `Authorization: Bearer` on every request. | ✅ | #126 |
| P40-T5 | `src-tauri/tauri.cloud.conf.json`, `.env.cloud` | **Tauri Cloud Build Config:** Create cloud-specific Tauri config (no sidecar, product name `BarnOwl Cloud`, new bundle ID) and `.env.cloud` pointing `VITE_API_BASE` at the Worker URL. | ✅ | #127 |
| P40-T6 | `.github/workflows/build-windows-cloud.yml` | **CI/CD Cloud Build Pipeline:** Fast (~2 min) GitHub Actions workflow — no PyInstaller, just `npm build` + `cargo tauri build`. Produces a `~10-15MB` `.msi` installer. | ✅ | #128 |
| P40-T7 | `frontend/src/lib/Settings.svelte`, `frontend/src/lib/StealthTerminal.svelte` | **Hide Local-Only UI in Cloud Build:** Use `VITE_BUILD_FLAVOR=cloud` to conditionally hide `StealthTerminal`, Audio Devices selector, and STT Engine selector in the Cloud Edition. | ⬜ | — |
| P40-T8 | `cloud-worker/`, `checkout-server/` | **Stripe Payment Integration (Supabase Edition):** Stripe Checkout for usage-based billing ($2 service fee per interview session + LLM API pass-through). On success, increment `payg_sessions` in Supabase `profiles` table. | ⬜ | — |
| P40-T9 | `cloud-worker/src/vector_cache.ts` | **Multi-Tenant Vector Cache (Supabase pgvector):** Replace ChromaDB with Supabase `pgvector`. Cache vectors are stored in the `qa_cache` table. All queries filter by `user_id` to guarantee strict user isolation. | ⬜ | — |
| P40-T10 | `cloud-worker/src/index.ts` | **Cloud Worker Backend Auth (Supabase Validation):** Validate the frontend Bearer token against the Supabase `user_api_keys` table and check the `payg_sessions` balance before allowing API access. | ✅ | #129 |