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
| P26-T1 | `web/src/routes/api/auth/` | **SSO Integration:** Add SAML/Okta single sign-on support via Supabase Auth for Enterprise customers. | ⏳ | — |
| P26-T2 | `web/src/routes/dashboard/admin/` | **Seat Management:** Admin dashboard to purchase blocks of seats via Stripe, and assign/revoke them to employee emails. | ⏳ | — |
| P26-T3 | `backend/rag/` | **Team Knowledge Base:** Expand RAG to allow uploading company-wide Playbooks/Docs to a shared vector database. | ⬜ | — |
| P26-T4 | `backend/app.py`, `frontend/src/lib/SessionReport.svelte` | **Auto-CRM Sync:** Add a "Push to Salesforce/HubSpot" button in the Session Report panel to log meeting notes directly to the CRM. | ⬜ | — |

---

## Phase 27 — Voice Conversational Mode (Alexa for Work) 🗣️

> Transforms the app from a passive stealth listener to an active Voice-In/Voice-Out assistant.
> Users can leave it running all day and ask it questions hands-free.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P27-T1 | `backend/audio_listener.py` | **Wake Word Detection:** Integrate Picovoice Porcupine to detect the wake word (e.g. "Hey Owl") before sending audio to the LLM. | ⬜ | — |
| P27-T2 | `backend/tts.py`, `backend/app.py` | **Text-to-Speech (TTS):** Pipe LLM responses through Edge-TTS (or ElevenLabs for premium users) and play audio through system speakers. | ⬜ | — |
| P27-T3 | `frontend/src/lib/Assistant.svelte` | **Voice Mode UI:** Add a visual indicator (like a glowing orb) when the assistant is actively listening/speaking, bypassing the stealth chat UI. | ⬜ | — |

---

## Phase 28 — Agentic Computer Control (OS Level) 🤖

> Gives BarnOwl "hands". Upgrades the LLM client to support tool-calling (function calling) so it can execute Python scripts to control the OS.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P28-T1 | `backend/llm_client.py` | **Tool Calling Integration:** Add `tools` and `tool_choice` parameters to the OpenAI API requests to allow the LLM to emit function calls instead of text responses. | ⬜ | — |
| P28-T2 | `backend/tools/` | **OS Interaction Toolkit:** Build Python functions using `subprocess` and `pyautogui` to open applications, manage windows, type text, and retrieve OS status. | ⬜ | — |
| P28-T3 | `backend/app.py` | **Agentic Event Loop:** Intercept tool-call responses from the LLM, execute the local Python function, and feed the result back to the LLM to continue the conversation. | ⬜ | — |
| P28-T4 | `backend/mcp/` | **Model Context Protocol (MCP):** Add an MCP client to the Tool Registry to dynamically discover and use tools from external enterprise MCP servers (e.g., GitHub, DBs). | ⬜ | — |

---

## Phase 29 — Auto-Tailored Resume Builder 📄

> Generates an ATS-optimized, tailored resume in Markdown using the user's base resume and target Job Description. Features a dual-pane editor and PDF export with 5-10 professional CSS templates.

| Task ID | File(s) | Description | Status | PR |
|---|---|---|---|---|
| P29-T1 | `backend/resume_builder.py`, `backend/app.py` | **Resume Generation API:** Endpoint that uses the LLM to output a perfectly tailored Markdown resume. | ⏳ | — |
| P29-T2 | `frontend/src/lib/ResumeBuilder.svelte` | **Resume Editor UI:** Dual-pane view with a Markdown editor on the left and a live preview on the right. | ⏳ | — |
| P29-T3 | `frontend/src/lib/resume-styles.css` | **Templates & Export:** 5-10 selectable CSS themes and a `window.print()` PDF export button. | ⬜ | — |

