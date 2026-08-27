# 🤖 Local AI Assistant — Jules Task Tracker

> Submit tasks to Jules one-by-one:
> ```bash
> python3 scripts/jules_submit.py --task P1-T1
> python3 scripts/jules_submit.py --list   # see all tasks
> ```
> Update status and PR numbers here as Jules returns PRs.

## Legend
| Symbol | Meaning |
|---|---|
| ⬜ | Not started |
| ⏳ | Submitted to Jules / In Progress |
| 🔍 | PR open — under review |
| ✅ | Merged |
| ❌ | Failed / needs rework |

---

## Phase 1 — Local Brain (Config + Health Check) 🧠

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P1-T1 | `backend/ollama_check.py` | Ollama health-check + curl test helper | ✅ | — |
| P1-T2 | `backend/config.py` | Central config module (all env-var overrideable settings) | ✅ | — |

---

## Phase 2 — Ears (STT / Audio Capture) 👂

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P2-T1 | `backend/audio_listener.py` | Mic capture with sounddevice, chunked streaming | ✅ | — |
| P2-T2 | `backend/transcriber.py` | faster-whisper CUDA integration, returns text stream | ✅ | — |
| P2-T3 | `backend/requirements.txt`, `backend/README.md` | All Python deps pinned | ✅ | — |

---

## Phase 3 — Brain Bridge (FastAPI + LLM) 🌁

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P3-T1 | `backend/llm_client.py` | Async OpenAI-compat LLM client with streaming token support | ✅ | — |
| P3-T2 | `backend/app.py` | FastAPI server with `/ws` WebSocket, pipes STT → LLM → frontend | ✅ | #14 |

---

## Phase 4 — Floating UI (Svelte 5 + Tauri) 🖥️

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P4-T1 | `frontend/` | Scaffold Svelte 5 + Tauri 2.0 project | ✅ | — |
| P4-T2 | `frontend/src/lib/ws.ts` | Reactive WebSocket store using Svelte 5 runes | ✅ | — |
| P4-T3 | `frontend/src/lib/Assistant.svelte`, `app.css` | Glassmorphism floating overlay, streams tokens | ✅ | — |
| P4-T4 | `frontend/src/App.svelte`, `main.ts` | Root component, mounts assistant, mic status dot | ✅ | #15 |
| P4-T5 | `src-tauri/src/main.rs` | **Stealth Mode** — screen share safe, dock hidden, global hotkey | ✅ | — |

---

## Phase 5 — Local RAG (Knowledge Base) 📚

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P5-T1 | `backend/rag/ingestor.py` | File upload endpoint, PDF/TXT chunking → ChromaDB | ✅ | #16 |
| P5-T2 | `backend/rag/retriever.py` | Semantic search, top-K retrieval | ✅ | #17 |
| P5-T3 | `backend/app.py` | Inject RAG context into LLM system prompt | ⬜ | — |
| P5-T4 | `frontend/src/lib/KnowledgeBase.svelte` | Drag-and-drop file upload UI | ⬜ | — |
| P5-T5 | `backend/rag/web_search.py` | Web Search via duckduckgo-search injected into context | ⬜ | — |

---

## Phase 6 — Vision Copilot (Code Screen Reader) 👁️

> Captures the screen via Tauri hotkey and sends to a Vision LLM (e.g. GPT-4o) 
> to analyze coding problems without speaking them aloud.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P6-T1 | `main.rs`, `Assistant.svelte` | Tauri screenshot capture hotkey (`Ctrl+Shift+S`) to Base64 | ⬜ | — |
| P6-T2 | `backend/app.py`, `llm_client.py` | `/vision/analyze` endpoint mapping to Vision LLM | ⬜ | — |

---

## Phase 7 — Polish & Packaging 🎁

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P7-T1 | `main.rs`, `ws.ts`, `Assistant.svelte` | Stealth Hotkeys: PTT, Scroll (Up/Down), Panic Clear | ⬜ | — |
| P7-T2 | `frontend/src/lib/Settings.svelte` | Settings panel: model selector, mic selector | ⬜ | — |
| P7-T3 | `scripts/build.sh` | Validate PyInstaller + Tauri bundles | ⬜ | — |
| P7-T4 | `README.md` | Final pass — screenshots, install instructions | ⬜ | — |
| P7-T5 | `scripts/start_remote.sh`, `backend/app.py` | Remote Helper Mode — serve UI statically + Cloudflare tunnel | ⬜ | — |
| P7-T6 | `tests/`, `e2e/`, `playwright.config.ts` | E2E test suite — pytest API tests + Playwright frontend tests | ⬜ | — |

---

## Phase 8 — BYOK SaaS (Auth + Billing) 🚀

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P8-T1 | `web/` | SvelteKit web app scaffold | ✅ | — |
| P8-T2 | `web/src/routes/login/` | Supabase auth + DB schema | ✅ | — |
| P8-T3 | `web/src/routes/api/billing/` | Stripe billing integration | ✅ | — |
| P8-T4 | `backend/auth.py`, `keys.py` | FastAPI JWT middleware + encrypted key storage | ⬜ | — |
| P8-T5 | `frontend/src/lib/auth.ts` | Tauri app auth flow — OS keychain | ⬜ | — |
