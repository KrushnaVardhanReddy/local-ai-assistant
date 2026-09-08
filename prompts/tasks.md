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
| P5-T3 | `backend/app.py` | Inject RAG context into LLM system prompt | ✅ | #19 |
| P5-T4 | `frontend/src/lib/KnowledgeBase.svelte` | Drag-and-drop file upload UI | ✅ | #21 |
| P5-T5 | `backend/rag/web_search.py` | Web Search via duckduckgo-search injected into context | ✅ | #27 |

---

## Phase 6 — Vision Copilot (Code Screen Reader) 👁️

> Captures the screen via Tauri hotkey and sends to a Vision LLM (e.g. GPT-4o) 
> to analyze coding problems without speaking them aloud.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P6-T1 | `main.rs`, `Assistant.svelte` | Tauri screenshot capture hotkey (`Ctrl+Shift+S`) to Base64 | ✅ | #22 |
| P6-T2 | `backend/app.py`, `llm_client.py` | `/vision/analyze` endpoint mapping to Vision LLM | ✅ | #25 |

---

## Phase 7 — Polish & Packaging 🎁

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P7-T1 | `main.rs`, `ws.ts`, `Assistant.svelte` | Stealth Hotkeys: PTT, Scroll (Up/Down), Panic Clear | ✅ | #28 |
| P7-T2 | `frontend/src/lib/Settings.svelte` | Settings panel: model selector, mic selector | ✅ | #24 |
| P7-T3 | `scripts/build.sh` | Validate PyInstaller + Tauri bundles | ✅ | #26 |
| P7-T4 | `README.md` | Final pass — screenshots, install instructions | ✅ | #23 |
| P7-T5 | `scripts/start_remote.sh`, `backend/app.py` | Remote Helper Mode — serve UI statically + Cloudflare tunnel | ✅ | #29 |
| P7-T6 | `tests/`, `e2e/`, `playwright.config.ts` | E2E test suite — pytest API tests + Playwright frontend tests | ✅ | #30 |
| P7-T7 | `frontend/src-tauri/src/main.rs`, `ws.ts` | Advanced Stealth Hotkeys — `Ctrl+Shift+1-6` to send pending transcript chips (P14). Map hidden inputs (e.g. arrow keys + Enter when UI hidden) to control Ghost Cursor for mouse-less selection. | ⬜ | — |

---

## Phase 8 — BYOK SaaS (Auth + Billing) 🚀

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P8-T1 | `web/` | SvelteKit web app scaffold | ✅ | — |
| P8-T2 | `web/src/routes/login/` | Supabase auth + DB schema | ✅ | — |
| P8-T3 | `web/src/routes/api/billing/` | Stripe billing integration | ✅ | — |
| P8-T4 | `backend/auth.py`, `keys.py` | FastAPI JWT middleware + encrypted key storage | ✅ | #18 |
| P8-T5 | `frontend/src/lib/auth.ts` | Tauri app auth flow — OS keychain | ✅ | #20 |

---

## Phase 9 — Smart Audio Intelligence 🧠🎙️

> Eliminates wasted LLM calls by filtering noise, fillers, and ambient audio.
> Prevents interruptions while the user is reading a streamed response.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P9-T1 | `backend/smart_filter.py`, `backend/app.py`, `backend/config.py` | Smart Audio Filter — VAD silence buffer + busy guard + intent heuristic | ✅ | #31 |

---

## Phase 10 — Modular Dashboard UI 🎨

> Complete UI overhaul from a vertical chat to a multi-panel dashboard optimized for stealth.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P10-T1 | `frontend/src/lib/Assistant.svelte` | Modular UI — Top toolbar, Left Pane (Live STT), Right Pane (LLM Response) | ✅ | — |


---

## Phase 11 — Personalization (Resume Parsing) 📄

> Uses client-side PDF parsing to extract the candidate's resume and inject it directly into the LLM system prompt as an in-memory persona.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P11-T1 | `frontend/src/lib/Settings.svelte`, `backend/app.py` | pdfjs-dist resume parsing & backend memory extraction | ✅ | #32 |

---

## Phase 12 — Interview UX Polish (Code Rendering + Language Control) 💎

> Improves the moment-of-need experience during live interviews.
> T1: Renders LLM markdown as formatted HTML with VS Code-quality syntax highlighting.
> T2: Lets the user lock the LLM to their preferred coding language, auto-detected from resume.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P12-T1 | `frontend/src/lib/markdownRenderer.ts`, `frontend/src/lib/Assistant.svelte` | Markdown + Shiki IDE-style syntax highlighting in The Brain panel | ✅ | #47 |
| P12-T2 | `backend/app.py`, `frontend/src/lib/Settings.svelte` | Language preference dropdown — injects code language into system prompt, auto-detected from resume | ✅ | #48 |
| P12-T3 | `frontend/src/app.css` | UI Polish: Stealth Text Contrast — add subtitle-style text shadows and sheer backdrop blur so text is readable over any IDE background | ✅ | #49 |

---

## Phase 13 — Audio Routing & Loopback 🎧

> Allows the assistant to hear the interviewer by capturing system speaker output instead of just the microphone.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P13-T1 | `backend/audio_listener.py`, `backend/app.py`, `frontend/src/lib/Settings.svelte` | Audio device selector + Windows WASAPI loopback support | ✅ | #50 |

---

## Phase 14 — Transcript Chip Bar (Click-to-Send) 🏷️

> Shows pending transcripts as clickable chips between the toolbar and panels.
> User can click any chip to manually send it to the LLM as a priority question,
> bypassing all audio filter layers via the existing `sendChat()` bypass.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P14-T1 | `frontend/src/lib/ws.svelte.ts`, `frontend/src/lib/Assistant.svelte` | Clickable transcript chip bar — accumulate, click-to-send, dismiss, auto-clear | ✅ | #36 |

---

## Phase 15 — Speaker Diarization (Dual-Voice Mode) 🎙️🎙️

> Inspired by Project Parakeet. When loopback audio is active (Phase 13),
> separate the interviewer's voice from the candidate's voice in the mixed stream.
> Interviewer speech → routed as "question context" to LLM.
> Candidate speech → optionally evaluated for answer quality.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P15-T1 | `backend/transcriber.py`, `backend/config.py` | Add `STT_DIARIZE=true` flag — use faster-whisper or pyannote speaker diarization on loopback stream to label each segment as `[INTERVIEWER]` or `[CANDIDATE]` | ✅ | #53 |
| P15-T2 | `backend/app.py`, `backend/smart_filter.py` | Route diarized segments differently — interviewer speech bypasses intent filter and is always forwarded as context; candidate speech goes through normal VAD pipeline | ✅ | #52 |
| P15-T3 | `frontend/src/lib/Assistant.svelte` | UI: Show speaker label badges on transcript chips (`👤 You` vs `🎤 Interviewer`) so user can tell the system is hearing both sides | ✅ | #51 |

---

## Phase 16 — Gemini Live Mode (Unified Audio + LLM Pipeline) ⚡

> Additive feature — does NOT replace the existing STT + LLM pipeline.
> Two modes, user picks in Settings:
>
> **Mode A — Gemini Live** (`LLM_PROVIDER=gemini`):
> Audio chunks stream directly into Gemini — no separate STT step.
> Gemini handles transcription + reasoning in one unified call. Sub-200ms.
>
> **Mode B — Everything else** (Ollama / Groq / OpenAI / LM Studio):
> Existing pipeline unchanged: STT_PROVIDER (local / groq) → LLM_PROVIDER.
> User picks STT and LLM independently.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P16-T1 | `backend/gemini_live_client.py` | New `GeminiLiveClient` class — streams 500ms PCM audio chunks to Gemini Live API via `google-genai` SDK, returns transcription + LLM response tokens in a single stream. Activated only when `LLM_PROVIDER=gemini`. | ✅ | #54 |
| P16-T2 | `backend/app.py` | In WebSocket handler: detect `LLM_PROVIDER=gemini` at startup and swap the `transcriber → llm_client` chain for `GeminiLiveClient`. All other providers continue through the existing pipeline untouched. | ✅ | #55 |
| P16-T3 | `backend/config.py`, `.env.local` | Add `GEMINI_API_KEY` and `GEMINI_LIVE_MODEL` config vars (default: `gemini-2.0-flash-live`). Document both modes in README. | ✅ | #46 |
| P16-T4 | `frontend/src/lib/Settings.svelte` | Settings panel: when `LLM_PROVIDER=gemini` is selected, hide the STT provider dropdown (not needed) and show a "Gemini Live — unified audio mode" badge. For all other LLM providers, show STT selector as normal. | ✅ | #57 |

---

## Phase 17 — Session Intelligence & Export 📊

> End-of-session features inspired by Project Parakeet's instant scorecard concept.
> The moment you close a session, a full structured debrief is ready —
> transcript with speaker labels, answer quality scores, missed topics, follow-up suggestions.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P17-T1 | `backend/app.py`, `backend/session_manager.py` | Session manager — accumulate full conversation (questions, answers, timestamps) in memory per WebSocket session. Add `POST /session/end` endpoint that returns structured JSON. | ✅ | #45 |
| P17-T2 | `backend/app.py`, `backend/llm_client.py` | Scorecard generation — on `POST /session/end`, send full transcript to LLM with a scorecard prompt: rate each answer (1–5), flag gaps, suggest what should have been said | ✅ | #45 |
| P17-T3 | `frontend/src/lib/Assistant.svelte`, `frontend/src/lib/SessionReport.svelte` | Session Report panel — triggered by `Ctrl+Shift+E` or button. Renders the scorecard as a formatted report with per-question breakdown. Has a copy-to-clipboard button. | ⏳ | — |
| P17-T4 | `backend/app.py` | Live Answer Coaching — after candidate finishes speaking (VAD silence detected), optionally send the answer to a fast LLM call and stream back a brief coaching hint: "✅ Good — also mention X" or "⚠️ Incomplete — you missed Y" | ⏳ | — |

---

## Phase 18 — Portable App & Stealth Identity 🕵️📦

> No installer, no Add/Remove Programs entry, no obvious process name.
> User downloads a ZIP, unzips, double-clicks — app runs immediately.
> Process name in Task Manager is configurable so it doesn't betray itself
> during a live interview if the candidate opens Task Manager or appwiz.cpl.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P18-T1 | `frontend/src-tauri/tauri.conf.json` | Switch Windows bundle target to portable — no NSIS installer, no registry writes, no appwiz.cpl entry. Output: a ZIP of `AppName.exe` + `resources/`. User unzips and runs directly. | ✅ | #41 |
| P18-T2 | `frontend/src-tauri/Cargo.toml`, `tauri.conf.json` | Change default `productName` to a neutral name (e.g. `"AudioService"`). This controls the EXE filename, Task Manager process name, and window title. | ✅ | #41 |
| P18-T3 | `frontend/src-tauri/tauri.conf.json`, `backend/config.py` | User-configurable process alias — read `APP_DISPLAY_NAME` from `.env.local` or a local `settings.json` at launch. Lets each user personalise their own process name without rebuilding. | ⏳ | — |
| P18-T4 | `scripts/build.sh`, `Makefile` | Update build pipeline: `make build-portable` target — runs PyInstaller on backend → Tauri portable build → zips both into a single `parakeet-portable-win.zip` release artifact. | ✅ | #44 |

---

## Phase 19 — Pricing Tiers & Billing 💰

> Three paid tiers + a free demo layer.
> All tiers are BYOK — user supplies their own API keys.
> Platform cost = auth/billing infra only (no GPU/API spend per user).
>
> | Tier | Price | Limit | Notes |
> |---|---|---|---|
> | **Demo** | Free | 3 questions or 15 min | No account needed, referral link |
> | **Pay-as-you-go** | $5 / session | 1 session = 90 min | Session token expires after 90 min |
> | **Monthly** | $19 / month | Unlimited sessions | Cancel anytime — main recurring revenue |
> | **Founding Member** | $49 one-time | Unlimited, while supported | Early adopter deal, not marketed as "lifetime" |

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P19-T1 | `web/src/routes/pricing/` | Public pricing page — show all 4 tiers with a comparison table. CTA buttons link to Stripe checkout. | ⬜ | — |
| P19-T2 | `web/src/routes/api/billing/`, `backend/auth.py` | Stripe integration: `price_payg` ($5 one-time session token), `price_monthly` ($19/mo sub), `price_founding` ($49 one-time). Webhook updates user's `plan` field in Supabase on payment success. | ⬜ | — |
| P19-T3 | `backend/auth.py`, `backend/app.py` | Session token enforcement — Pay-as-you-go users get a JWT with `expires_at = now + 90min`. Backend validates on every WebSocket message. When token expires, send `{"type": "session_expired"}` to frontend. | ⬜ | — |
| P19-T4 | `frontend/src/lib/Assistant.svelte` | Session expiry UI — when `session_expired` event received, show a non-intrusive overlay: "Session ended — extend for $5 or upgrade to Monthly". Has a direct Stripe payment link. | ⬜ | — |
| P19-T5 | `web/src/routes/demo/` | Demo / Referral mode — every user gets a unique referral link (`parakeet.app/ref/[code]`). New visitor clicks link → gets 15 min free demo. Track referral source + conversion in Supabase (`referrals` table: referrer_id, referee_id, status, converted_at). | ⬜ | — |
| P19-T6 | `backend/auth.py` | Plan-based feature gating — `demo`: 3 question limit + watermark; `payg`: full session, no coaching; `monthly`/`founding`: all features including Live Answer Coaching (P17-T4) and Session Report (P17-T3). | ⬜ | — |
| P19-T7 | `backend/auth.py`, `web/src/routes/api/referral/` | Referral reward system — when a referred user completes signup AND uses their first demo session, the referrer automatically receives +15 min of demo credit added to their account (`demo_credits_minutes` in Supabase). No cap — each successful referral = +15 min. Paid users bank the credits for when friends haven't upgraded yet. Send referrer a notification email: "Your friend joined — you earned 15 bonus minutes! 🎉" | ⬜ | — |

---

## Phase 20 — Anti-Sharing: Device Lock & Session Enforcement 🔐

> Prevents credential sharing between users.
> Infrastructure is already partially built:
> - `lib.rs` → `get_machine_id()` generates a per-machine UUID stored in OS keyring ✅
> - `ws.svelte.ts` → sends `machine_id` on every WebSocket auth handshake ✅
> - Just needs backend enforcement + Supabase schema.
>
> **Device limits per plan:**
> | Plan | Max Registered Devices | Concurrent Sessions |
> |---|---|---|
> | Demo | 1 (machine-bound, no login) | 1 |
> | Pay-as-you-go | 1 | 1 |
> | Monthly | 2 | 1 |
> | Founding Member | 3 | 1 |

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P20-T1 | `backend/supabase_schema.sql` | Add `user_devices` table: `(user_id, machine_id, device_label, registered_at, last_seen_at)`. Add `active_session_id` column to `users` table to track current live session. | ⬜ | — |
| P20-T2 | `backend/auth.py` | Device registration — on first WebSocket auth from an unknown `machine_id`, check device count against plan limit. If under limit: register device and allow. If at limit: reject with `{"type": "device_limit_reached", "max": N}` — user must remove a device from dashboard first. | ⬜ | — |
| P20-T3 | `backend/auth.py`, `backend/app.py` | Concurrent session lock — on WebSocket connect, write `active_session_id = new_session_uuid` to Supabase. If account already has a different `active_session_id`, reject new connection with `{"type": "already_active"}`. On disconnect, clear `active_session_id`. | ⬜ | — |
| P20-T4 | `web/src/routes/dashboard/devices/` | Device management page — lists all registered devices with label, last-seen date. User can remove a device (frees up a slot). Useful when switching machines legitimately. | ⬜ | — |
| P20-T5 | `frontend/src/lib/Assistant.svelte`, `frontend/src/lib/auth.svelte.ts` | Handle rejection events in frontend: `device_limit_reached` → show "Max devices reached, manage at dashboard.parakeet.app/devices"; `already_active` → show "Another session is already active — close it first or wait 5 minutes for it to expire automatically." | ⬜ | — |


---

## Phase 21 — Mock Interviews & Internationalization 🌍

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P21-T1 | `frontend/src/lib/Settings.svelte`, `backend/config.py` | **Multilingual Support:** Add language override dropdown. Updates STT prompt and LLM system prompt to enforce target language. | ⬜ | — |
| P21-T2 | `backend/app.py`, `frontend/src/lib/Assistant.svelte` | **Mock Interview Mode:** Toggle that flips the LLM from "Answerer" to "Interviewer". Uses Edge-TTS to speak questions aloud. Feeds user answers back for evaluation. | ⬜ | — |

---

## Phase 22 — Dynamic Job Context Grounding 🏢

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P22-T1 | `frontend/src/lib/Settings.svelte`, `backend/config.py`, `backend/app.py` | **Job Context Injection:** Add a text area in Settings for the Job Description. The backend dynamically injects this into the LLM system prompt so all real-time answers are tailored specifically to the company and role requirements. | ⬜ | — |

---

## Phase 23 — Comprehensive E2E Testing Suite 🧪

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P23-T1 | `frontend/e2e/`, `tests/e2e/`, `Makefile` | **Playwright + Pytest E2E:** Build out comprehensive UI tests mocking Tauri IPC, and backend integration tests mocking WS connections to guarantee reliability before final release. | ⬜ | — |
