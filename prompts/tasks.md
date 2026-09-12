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
## Phase 26 — Enterprise B2B Features 🏢

> Shifting from B2C to B2B. Requires Supabase authentication upgrades and SSO integration.
> Allows IT Admins to manage seats, and creates a Team Knowledge Base.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P26-T5 | `web/tests/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright and Pytest tests for Enterprise B2B features. | ✅ | #114 |

---

## Phase 27 — Voice Conversational Mode (Alexa for Work) 🗣️

> Transforms the app from a passive stealth listener to an active Voice-In/Voice-Out assistant.
> Users can leave it running all day and ask it questions hands-free.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P27-T1 | `backend/audio_listener.py` | **Wake Word Detection:** Integrate Picovoice Porcupine to detect the wake word (e.g. "Hey Owl") before sending audio to the LLM. | ⬜ | — |
| P27-T2 | `backend/tts.py`, `backend/app.py` | **Text-to-Speech (TTS):** Pipe LLM responses through Edge-TTS (or ElevenLabs for premium users) and play audio through system speakers. | ⬜ | — |
| P27-T3 | `frontend/src/lib/Assistant.svelte` | **Voice Mode UI:** Add a visual indicator (like a glowing orb) when the assistant is actively listening/speaking, bypassing the stealth chat UI. | ⬜ | — |
| P27-T4 | `frontend/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright and Pytest tests for Voice Mode. | ⬜ | — |

---

## Phase 28 — Agentic Computer Control (OS Level) 🤖

> Gives BarnOwl "hands". Upgrades the LLM client to support tool-calling (function calling) so it can execute Python scripts to control the OS.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P28-T2 | `backend/tools/` | **OS Interaction Toolkit:** Build Python functions using `subprocess` and `pyautogui` to open applications, manage windows, type text, and retrieve OS status. | ⬜ | — |
| P28-T3 | `backend/app.py` | **Agentic Event Loop:** Intercept tool-call responses from the LLM, execute the local Python function, and feed the result back to the LLM to continue the conversation. | ⬜ | — |
| P28-T4 | `backend/mcp/` | **Model Context Protocol (MCP):** Add an MCP client to the Tool Registry to dynamically discover and use tools from external enterprise MCP servers (e.g., GitHub, DBs). | ⬜ | — |
| P28-T5 | `tests/e2e/` | **E2E Tests:** Pytest tests for Agentic Tool Registry and Loop. | ⬜ | — |

---

## Phase 30 — Platform Integrations 🔌

> BarnOwl becomes the "layer on top" of every tool teams already use. Instead of competing with Zoom AI or Slack AI, BarnOwl embeds inside them — bringing your private, context-aware AI into the tools you're already in.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P30-T1 | `chrome-extension/` | **Google Meet Chrome Extension:** Injects a BarnOwl sidebar into Google Meet. Reads live captions via DOM MutationObserver and streams them to the local backend for real-time suggestions. Pure HTML/CSS/JS, load as unpacked extension. | ✅ | #107 |
| P30-T2 | `slack-bot/` | **Slack Bot:** `/barnowl ask`, `/barnowl summarize`, and `@BarnOwl` mention support. Uses Slack Bolt SDK with Socket Mode (no public URL needed). Posts session summaries as rich Block Kit messages. | ⬜ | — |
| P30-T3 | `zoom-app/` | **Zoom App (In-Meeting Sidebar):** Embeds BarnOwl as a Zoom Apps iframe panel. Uses the Zoom JS SDK to get meeting context and user identity. Minimal Express server to host the app for local testing. | ⬜ | — |
| P30-T4 | `frontend/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright tests for Chrome Extension sidebar, Pytest stubs for Slack bot and Zoom app. | ⬜ | — |

---|---|---|---|---|
| P32-T6 | `tests/test_local_intelligence.py`, `tests/test_qa_cache.py` | **Unit Tests:** Tests for turn-detection accuracy (COMPLETE/INCOMPLETE), cache hit/miss logic, similarity threshold, and cache clearing via the API endpoint. | ✅ | #108 |
| P32-T7 | `frontend/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright tests for Cache UI (stats display, Clear Cache button, confirmation dialog). Pytest integration tests for `GET /api/cache/stats` and `DELETE /api/cache` endpoints. | ✅ | #106 |
| P32-T8 | `backend/app.py`, `frontend/src/lib/Assistant.svelte` | **Proactive Cache Pre-Warming (Mind Reader):** Add a `POST /api/cache/prewarm` endpoint. It takes the candidate's Resume and Job Description, asks the cloud LLM to generate the 50 most likely interview questions + perfect answers, and bulk-inserts them into the ChromaDB `qa_cache` prior to the interview. | ⏳ | — |
| P32-T10 | `backend/local_intelligence.py`, `backend/app.py` | **SmolLM2 Auto-Download on First Run:** On backend startup, if `SMOLLM2_ENABLED=true` but the GGUF model file is missing from `backend/models/`, automatically download `SmolLM2-135M-Instruct-Q4_K_M.gguf` (~100MB) from HuggingFace with a progress bar logged to stderr. Keep the ZIP small — ship without the model, download it once silently on first launch. Also add `llama-cpp-python` to `backend/requirements.txt` as an optional dep. | ⏳ | — |

---

## Phase 34 — Meeting Memory & Task Sync (Enterprise Timeline) 🧠🗓️

> Bridges the gap between a real-time copilot and an async enterprise assistant (Otter/Limitless style).
> Automatically records, indexes, and extracts actionable items from every conversation.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P34-T1 | `backend/database/` | **Persistent SQLite Timeline:** Update the backend to silently log all spoken transcripts into a localized timeline DB, allowing for infinite searchable memory across all days and sessions. | ⬜ | — |
| P34-T2 | `frontend/src/lib/MemoryTimeline.svelte` | **Timeline UI:** Create a timeline view where users can scrub back through their entire day's audio transcripts and search for specific keywords from past meetings. | ⬜ | — |
| P34-T3 | `backend/task_extractor.py` | **Action Item Extraction:** Run an async LLM post-processing job over the `faster-whisper` transcripts every 15 minutes to automatically extract Action Items and To-Dos. | ⬜ | — |
| P34-T4 | `backend/integrations/` | **Jira & Trello Sync:** Add OAuth/API sync capability to automatically push extracted Action Items into the user's Jira board or Trello list. | ⬜ | — |

---|---|---|---|---|
| P35-T1 | `frontend/src/` | **Hotkeys Side Panel:** Slide-out drawer for hotkeys so users don't lose context. | ✅ | #109 |

---

## Phase 36 — Auto STAR & Smart Context Pipeline

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
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
