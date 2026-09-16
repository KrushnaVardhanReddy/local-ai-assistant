# 🚀 StealthPresenter — Project Handoff

*This document summarizes the current state of the Local AI Assistant. Read this to quickly resume development in a new AI session.*

---

## 🎯 Architecture Summary
We recently completed Phase 55, successfully building the full **StealthPresenter** product on top of our pure Go/Wails Hexagonal Engine.

**Core Stack:**
- **Backend:** Go 1.25+, Wails v2, `whisper.cpp` (STT), SQLiteVec (Cache).
- **Frontend:** Svelte 5, Vite, TailwindCSS.
- **Engine:** Pure Go `StealthEngine` (Hexagonal Architecture).

---

## 🚦 Current Status

### Recently Completed ✅
- **Phase 55 (StealthPresenter UX & Engine):**
  - **T1:** Transparent Svelte Shell & `StealthTitleBar`.
  - **T2:** Document Parser (`.pdf`, `.pptx`, `.md`, `.txt`).
  - **T3:** Voice-Scroller Engine mapped to local Whisper.
  - **T4:** LLM Audience Copilot (glowing side-drawer that uses RAG to answer audience questions live).
  - Added native OS file picker (`📁` button) to the HUD.

### Up Next (Bug Bash) ⏭️
- The core functionality is completely built, but **many issues remain**. 
- We will now tackle bugs, UI glitches, and stabilization issues **one by one**.
- Known areas to investigate:
  - VAD (Voice Activity Detection) tuning and error handling.
  - Svelte CSS warnings (`.error-msg`, `.animate-shimmer_2s_infinite`).
  - General polishing of the Copilot Drawer and Scroller sync.

---

## 🛠️ How to Resume Development
1. **Run locally:** `make dev-presenter`
2. **Task Tracker:** Check `prompts/tasks_v2.md` for the roadmap.
3. **Submit to Jules:** Use `python3 scripts/jules_submit.py --task <ID>`.
