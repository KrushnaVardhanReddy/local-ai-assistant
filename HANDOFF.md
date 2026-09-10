# 🚀 BarnOwl (formerly Local AI Assistant) — Project Handoff

*This document summarizes the current state of the BarnOwl project as of September 2026. Use this to quickly resume development in a new session.*

---

## 🎯 Architecture Summary
We have evolved the initial Local Whisper concept into a full **Anti-Piracy, Split-Compute SaaS Platform** with a robust Edge Intelligence layer.

**Core Stack:**
- **Backend:** Python + FastAPI + `faster-whisper` + ChromaDB + SmolLM2 (Turn-taking/Caching).
- **Desktop Client:** Tauri v2 + Svelte 5.
- **Mobile Client:** Capacitor wrapping the Svelte 5 frontend (iOS/Android).
- **SaaS Web App:** SvelteKit + Supabase + Stripe.
- **Integrations:** Chrome Extension (Google Meet).

**Key Innovations:**
- **Local Intelligence Layer:** Uses `SmolLM2-135M` locally to detect complete thoughts and filter out conversational filler before hitting expensive cloud LLMs.
- **Semantic Q&A Cache:** Stores past responses in ChromaDB; serves instant zero-latency answers for repeated or semantically similar questions.
- **Stealth Mode:** OS-level kernel hooks to hide the app from screen shares, plus global CSS overrides (`cursor-default`) to prevent pointer changes from giving away the invisible overlay.
- **Split-Compute (Remote Helper):** Backend runs on a heavy machine (or a remote friend's machine via Cloudflare Tunnels), while the candidate runs the lightweight UI locally.

---

## 🚦 Current Status (As of Session End)

We are currently wrapping up **Phase 32 (SmolLM2 Local Intelligence Layer)** and moving towards **Phase 34 (Enterprise Timeline)**.

### Recently Completed ✅
- **Phase 30:** Built and merged the Google Meet Chrome Extension (`P30-T1`).
- **Phase 32:** Merged the SmolLM2 turn gatekeeper, Semantic Q&A Cache, Cache UI, and full unit/E2E test suites for the local intelligence layer (`P32-T1` through `P32-T7`).
- **UI & Stealth Polish:** Renamed all user-facing UI from "Local AI" to "BarnOwl" and fixed a critical stealth leak where the mouse pointer would change over invisible buttons.

### In Progress (Jules Tasks) ⏳
Two tasks are currently delegated to Jules and should have open PRs shortly:
1. **P32-T8 (Mind Reader):** Proactive cache pre-warming endpoint (`POST /api/cache/prewarm`) using Resume + JD.
2. **P32-T10 (SmolLM2 Auto-Download):** Backend startup logic to automatically download the 100MB GGUF model on first run if enabled, keeping the portable release `.exe` small.

### Up Next ⏭️
- Merge PRs for `P32-T8` and `P32-T10` once Jules finishes.
- Move into **Phase 34 (Enterprise Timeline)** to build the persistent SQLite transcript memory and Timeline UI.

---

## 🛠️ How to Resume Development
1. Start the stack: `make dev-all`
2. Check `prompts/tasks.md` for the current roadmap and task statuses.
3. Check GitHub for Jules' PRs on P32-T8 and P32-T10. Run `scripts/merge_prs.sh 109 110` (or whatever the PR numbers end up being) to validate and merge them.
