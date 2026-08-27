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
| P1-T1 | `backend/ollama_check.py` | Ollama health-check + curl test helper | ⬜ | — |
| P1-T2 | `backend/config.py` | Central config module (all env-var overrideable settings) | ⏳ | — |

---

## Phase 2 — Ears (STT / Audio Capture) 👂

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P2-T1 | `backend/audio_listener.py` | Mic capture with sounddevice, chunked streaming | ⬜ | — |
| P2-T2 | `backend/transcriber.py` | faster-whisper CUDA integration, returns text stream | ⬜ | — |
| P2-T3 | `backend/requirements.txt`, `backend/README.md` | All Python deps pinned | ⬜ | — |

---

## Phase 3 — Brain Bridge (FastAPI + LLM) 🌁

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P3-T1 | `backend/llm_client.py` | Async OpenAI-compat LLM client with streaming token support | ⬜ | — |
| P3-T2 | `backend/app.py` | FastAPI server with `/ws` WebSocket, pipes STT → LLM → frontend | ⬜ | — |

---

## Phase 4 — Floating UI (Svelte 5 + Tauri) 🖥️

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P4-T1 | `frontend/` | Scaffold Svelte 5 + Tauri 2.0 project | ⏳ | — |
| P4-T2 | `frontend/src/lib/ws.ts` | Reactive WebSocket store using Svelte 5 runes | ⬜ | — |
| P4-T3 | `frontend/src/lib/Assistant.svelte`, `app.css` | Glassmorphism floating overlay, streams tokens | ⬜ | — |
| P4-T4 | `frontend/src/App.svelte`, `main.ts` | Root component, mounts assistant, mic status dot | ⬜ | — |
| P4-T5 | `src-tauri/src/main.rs` | **Stealth Mode** — screen share safe, dock hidden, tab switch silent, global toggle hotkey | ⬜ | — |

---

## Phase 5 — Local RAG (Knowledge Base) 📚

> Fully offline — ChromaDB + local sentence-transformers. Zero cloud needed.
> Users can upload PDFs/TXT/MD and the LLM automatically uses them as context.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P5-T1 | `backend/rag/ingestor.py` | File upload endpoint, PDF/TXT/MD parser, chunking → ChromaDB | ⬜ | — |
| P5-T2 | `backend/rag/retriever.py` | Semantic search, top-K retrieval with source deduplication | ⬜ | — |
| P5-T3 | `backend/app.py` | Inject RAG context into LLM system prompt (async, non-blocking) | ⬜ | — |
| P5-T4 | `frontend/src/lib/KnowledgeBase.svelte` | Drag-and-drop file upload UI + source citation display in assistant | ⬜ | — |

---

## Phase 6 — Polish & Packaging 🎁

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P6-T1 | `main.rs`, `ws.ts`, `Assistant.svelte` | Push-to-talk global hotkey (`Ctrl+Shift+P`) with PTT/always-on toggle | ⬜ | — |
| P6-T2 | `frontend/src/lib/Settings.svelte` | Settings panel: model selector, mic selector, stealth toggle | ⬜ | — |
| P6-T3 | `scripts/build.sh` | Validate PyInstaller + Tauri bundles correctly end-to-end | ⬜ | — |
| P6-T4 | `README.md` | Final pass — screenshots, GIF demo, install instructions | ⬜ | — |

---

## Phase 7 — BYOK SaaS (Auth + Billing + Dashboard) 🚀

> Build on top of the finished local tool. The backend barely changes — just adds a JWT gate.
> Zero regressions: local mode still works with no env vars set.
> Phase 5 (RAG) becomes a Pro-tier upgrade: cloud-synced knowledge bases across devices.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P7-T1 | `web/` | SvelteKit web app — landing page, pricing, user dashboard scaffold | ⏳ | — |
| P7-T2 | `web/src/routes/login/`, `supabase/migrations/` | Supabase auth (email + Google OAuth) + DB schema (profiles, BYOK keys, usage) | ⬜ | — |
| P7-T3 | `web/src/routes/api/billing/` | Stripe billing — checkout, webhook, customer portal | ⬜ | — |
| P7-T4 | `backend/auth.py`, `backend/keys.py` | FastAPI JWT middleware + encrypted BYOK key storage (Fernet AES) | ⬜ | — |
| P7-T5 | `frontend/src/lib/auth.ts`, `main.rs` | Tauri app auth flow — OS keychain token storage, WS JWT handshake | ⬜ | — |

### SaaS Architecture (BYOK model)

```
User's Browser                 Tauri Desktop App          Your Hosted Backend
──────────────                 ─────────────────          ───────────────────
web/ (SvelteKit)               frontend/ (Svelte+Tauri)   backend/ (FastAPI)
  Landing page                   Signs in via Supabase       /ws — validates JWT
  Pricing                        Stores JWT in OS keychain   /api/keys — BYOK storage
  Dashboard                      Sends JWT on WS connect     Uses user's own API key
  API key manager ──saves──►  Supabase DB (encrypted)        Logs usage to Supabase
  Stripe billing
  Knowledge Base sync (Pro) ──────────────────────────────► /rag/* cloud sync
```

**Key principle:** You never touch the user's API keys in plaintext.
They're encrypted (Fernet AES) before storage and decrypted only server-side per-request.

---

## Stealth Features Reference

> Implemented in **P4-T5**. How each feature works:

| Feature | macOS | Windows | Tauri API |
|---|---|---|---|
| Screen Share Safe | `NSWindow.sharingType = NSWindowSharingNone` | `SetWindowDisplayAffinity(WDA_EXCLUDEFROMCAPTURE)` | Custom Rust |
| Dock Hidden | `NSApp.activationPolicy = .accessory` | `WS_EX_TOOLWINDOW` style | `skip_taskbar: true` + Rust |
| Task Manager Invisible | App still in Activity Monitor (process-level limitation) | Window hidden from taskbar | Partial |
| Tab Switch Silent | `.accessory` policy removes from CMD+Tab | `WS_EX_TOOLWINDOW` removes from Alt+Tab | `skip_taskbar: true` |
| Toggle Visibility | `Ctrl+Shift+Space` global shortcut | Same | Tauri global shortcut |

---

## Notes

- Keep tasks **small and atomic** — one file / one feature per Jules session.
- Always update PR column after Jules creates a PR.
- Run `bash scripts/merge_prs.sh <start> <end>` to batch merge after review.
- Phase 5 (RAG) depends on Phase 3 (FastAPI) + Phase 4 (UI) being complete.
- Phase 7 (SaaS) depends on Phases 1–6 being complete. Don't start P7 until P6 is merged.
- In the SaaS tier, the RAG knowledge base can be optionally cloud-synced to Supabase Storage as a Pro feature.
