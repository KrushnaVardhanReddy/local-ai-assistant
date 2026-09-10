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


> 📦 View completed phases in `tasks_completed.md`.

## Phase 26 — Enterprise B2B Features 🏢

> Shifting from B2C to B2B. Requires Supabase authentication upgrades and SSO integration.
> Allows IT Admins to manage seats, and creates a Team Knowledge Base.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P26-T1 | `backend/auth.py`, `backend/app.py`, `frontend/src/routes/login/+page.svelte` | **SSO Integration:** Integrate SAML/SSO via Supabase for enterprise login, auto-provisioning a Seat based on the email domain. | ✅ | — |
| P26-T2 | `frontend/src/lib/AdminDashboard.svelte` | **Seat Management:** Simple admin panel for the Org Admin to view active seats, invite via email, and instantly revoke API access. | ✅ | — |
| P26-T3 | `backend/rag/` | **Team Knowledge Base:** Expand RAG to allow uploading company-wide Playbooks/Docs to a shared vector database. | ✅ | — |
| P26-T4 | `frontend/src/lib/SessionReport.svelte` | **Universal Export Menu:** Add an export dropdown menu containing "Copy to Clipboard", "Download as Markdown", and "Draft as Email" options in the Session Report panel. | ✅ | — |
| P26-T5 | `web/tests/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright and Pytest tests for Enterprise B2B features. | ⬜ | — |

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
| P28-T1 | `backend/llm_client.py` | **Tool Calling Integration:** Add `tools` and `tool_choice` parameters to the OpenAI API requests to allow the LLM to emit function calls instead of text responses. | ✅ | — |
| P28-T2 | `backend/tools/` | **OS Interaction Toolkit:** Build Python functions using `subprocess` and `pyautogui` to open applications, manage windows, type text, and retrieve OS status. | ⬜ | — |
| P28-T3 | `backend/app.py` | **Agentic Event Loop:** Intercept tool-call responses from the LLM, execute the local Python function, and feed the result back to the LLM to continue the conversation. | ⬜ | — |
| P28-T4 | `backend/mcp/` | **Model Context Protocol (MCP):** Add an MCP client to the Tool Registry to dynamically discover and use tools from external enterprise MCP servers (e.g., GitHub, DBs). | ⬜ | — |
| P28-T5 | `tests/e2e/` | **E2E Tests:** Pytest tests for Agentic Tool Registry and Loop. | ⬜ | — |

---

## Phase 29 — Auto-Tailored Resume Builder 📄

> Generates an ATS-optimized, tailored resume in Markdown using the user's base resume and target Job Description. Features a dual-pane editor and PDF export with 5-10 professional CSS templates.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P29-T1 | `backend/resume_builder.py`, `backend/app.py` | **Resume Generation API:** Endpoint that uses the LLM to output a perfectly tailored Markdown resume. | ✅ | — |
| P29-T2 | `frontend/src/lib/ResumeBuilder.svelte` | **Resume Editor UI:** Dual-pane view with a Markdown editor on the left and a live preview on the right. | ✅ | — |
| P29-T3 | `frontend/src/lib/resume-styles.css` | **Templates & Export:** 5-10 selectable CSS themes and a `window.print()` PDF export button. | ✅ | — |
| P29-T4 | `frontend/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright and Pytest tests for Resume Builder. | ✅ | — |

---

## Phase 30 — Platform Integrations 🔌

> BarnOwl becomes the "layer on top" of every tool teams already use. Instead of competing with Zoom AI or Slack AI, BarnOwl embeds inside them — bringing your private, context-aware AI into the tools you're already in.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P30-T1 | `chrome-extension/` | **Google Meet Chrome Extension:** Injects a BarnOwl sidebar into Google Meet. Reads live captions via DOM MutationObserver and streams them to the local backend for real-time suggestions. Pure HTML/CSS/JS, load as unpacked extension. | ⬜ | — |
| P30-T2 | `slack-bot/` | **Slack Bot:** `/barnowl ask`, `/barnowl summarize`, and `@BarnOwl` mention support. Uses Slack Bolt SDK with Socket Mode (no public URL needed). Posts session summaries as rich Block Kit messages. | ⬜ | — |
| P30-T3 | `zoom-app/` | **Zoom App (In-Meeting Sidebar):** Embeds BarnOwl as a Zoom Apps iframe panel. Uses the Zoom JS SDK to get meeting context and user identity. Minimal Express server to host the app for local testing. | ⬜ | — |
| P30-T4 | `frontend/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright tests for Chrome Extension sidebar, Pytest stubs for Slack bot and Zoom app. | ⬜ | — |

---

## Phase 31 — Freemium PLG Model 💸

> Implementing the Product-Led Growth freemium model. Free users get unlimited local transcription (faster-whisper) at zero cloud cost. Paid users unlock Cloud LLMs for summaries, chat, and agentic tools.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P31-T1 | `backend/app.py`, `backend/config.py` | **Freemium Engine:** Lock LLM chat, summary, and CRM sync behind a subscription check. Force free users to strictly use local STT (`faster-whisper` or `parakeet`) and disable cloud LLM endpoints. | ✅ | — |
| P31-T2 | `backend/audio_listener.py` | **Multi-Engine Local STT:** Allow power users to select their local STT engine in settings. Support `faster-whisper` (universal/CPU/Mac) and Nvidia `parakeet` (for RTX GPU owners). | ✅ | — |

---

## Phase 32 — SmolLM2 Local Intelligence Layer 🧠

> Adds a tiny on-device SmolLM2 model (135M params, ~100MB RAM) as a local pre-filter and semantic cache.
> This layer runs entirely offline, cuts cloud API costs by 40-60%, and makes the assistant feel instant.
> Two core jobs: (1) turn-taking gatekeeper — stops premature LLM calls mid-sentence; (2) semantic cache encoder — finds similar past Q&As in ChromaDB to serve answers instantly without hitting the cloud.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P32-T1 | `backend/local_intelligence.py` | **SmolLM2 Engine:** Load `SmolLM2-135M-Instruct` (GGUF quantized) via `llama-cpp-python` at server startup. Expose two async methods: `is_complete(text) → bool` (turn detection) and `encode(text) → vector` (semantic embedding). | ✅ | — |
| P32-T2 | `backend/smart_filter.py` | **Turn-Taking Gatekeeper:** Before firing any STT transcript to the cloud LLM, call `local_intelligence.is_complete()`. If `INCOMPLETE`, reset the silence timer and keep listening. Log skipped incomplete turns. | ✅ | — |
| P32-T3 | `backend/qa_cache.py`, `backend/app.py` | **Semantic Q&A Cache:** Use the SmolLM2 encoder to convert each question into a vector and store/look up past Q&A pairs in a dedicated ChromaDB collection (`qa_cache`). On cache hit (cosine similarity > 0.92), serve the cached answer instantly, skipping cloud LLM entirely. | ✅ | — |
| P32-T4 | `backend/app.py` | **Cache Management API:** Add `DELETE /api/cache` endpoint to wipe all entries from the `qa_cache` ChromaDB collection. Add `GET /api/cache/stats` to show total cached pairs and estimated tokens saved. | ⏳ | — |
| P32-T5 | `frontend/src/lib/Assistant.svelte` | **Cache UI Controls:** Add a "Cache" section in the settings panel showing cache stats (e.g. "47 answers cached — ~12,000 tokens saved"). Include a red "Clear Cache" button that calls the `DELETE /api/cache` endpoint with a confirmation dialog. | ⏳ | — |
| P32-T6 | `tests/test_local_intelligence.py`, `tests/test_qa_cache.py` | **Unit Tests:** Tests for turn-detection accuracy (COMPLETE/INCOMPLETE), cache hit/miss logic, similarity threshold, and cache clearing via the API endpoint. | ⬜ | — |
| P32-T7 | `frontend/e2e/`, `tests/e2e/` | **E2E Tests:** Playwright tests for Cache UI (stats display, Clear Cache button, confirmation dialog). Pytest integration tests for `GET /api/cache/stats` and `DELETE /api/cache` endpoints. | ⬜ | — |
| P32-T8 | `backend/app.py`, `frontend/src/lib/Assistant.svelte` | **Proactive Cache Pre-Warming (Mind Reader):** Add a `POST /api/cache/prewarm` endpoint. It takes the candidate's Resume and Job Description, asks the cloud LLM to generate the 50 most likely interview questions + perfect answers, and bulk-inserts them into the ChromaDB `qa_cache` prior to the interview. | ⬜ | — |
| P32-T9 | `backend/local_intelligence.py`, `backend/smart_filter.py` | **Semantic Intent Classification:** Add an `is_question(text) -> bool` method to `SmolLM2` using a zero-shot prompt. Replace the old hardcoded keyword logic in `smart_filter.py` with this semantic check, saving API calls on conversational filler (e.g., "I see", "That's good"). | ⏳ | — |

---

## Phase 33 — Native Mobile App (Capacitor) 📱

> Wraps our existing Svelte web app into a native iOS and Android application using Ionic Capacitor.
> Enables native mobile distribution (App Store / Play Store) with full access to device hardware.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P33-T1 | `capacitor.config.ts`, `package.json` | **Capacitor Scaffolding:** Install `@capacitor/core` and `@capacitor/cli`. Initialize the project and add iOS and Android targets. Configure SvelteKit adapter-static for native builds. | ✅ | — |
| P33-T2 | `frontend/src/lib/audio/` | **Native Microphone:** Replace the Web Audio API with `@capacitor-community/microphone` for seamless audio capture on mobile devices without browser permission prompts. | ✅ | — |
| P33-T3 | `frontend/src/lib/audio/` | **Background Audio Plugin:** Integrate a Capacitor background task plugin so the assistant can continue listening for the wake word even when the phone screen is locked. | ⏳ | — |
| P33-T4 | `frontend/e2e/` | **Mobile E2E Tests:** Add mobile viewport emulation to Playwright tests to ensure the UI remains responsive and functional on smaller screens. | ✅ | — |

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
