# 🚀 BarnOwl AI — Project Handoff

*This document summarizes the current state of BarnOwl. Read this to quickly resume development in a new AI session.*

---

## 🎯 Architecture Summary (The Great Pivot)
We recently completed a massive architectural pivot (Phase 44) from Python/Rust/Tauri to a **Unified Go Single-Binary** using Wails v2.

**Core Stack:**
- **Backend:** Go 1.25+, Wails v2 (`github.com/wailsapp/wails/v2`).
- **Frontend:** Svelte 5, Vite, TailwindCSS v3.
- **Local AI / STT:** `whisper.cpp` (via Go CGO bindings) and `onnxruntime_go` for local embeddings.
- **SaaS / Web:** SvelteKit + Supabase + Stripe.
- **Distribution:** Compiles into a single ~20MB executable.

**Key Architectural Notes:**
- **Zero Python/Rust:** All legacy Python and Rust code was permanently deleted in `P44-T1`.
- **Hybrid Auto-Detect STT (Upcoming):** We are building a hybrid STT engine that defaults to `whisper.cpp` but dynamically downloads and hot-swaps to Parakeet (`sherpa-onnx`) if the user's hardware supports it.
- **Stealth Mode UI:** The Wails window is configured in `wails-app/main.go` to be transparent, frameless, and always-on-top.

---

## 🚦 Current Status

### Recently Completed ✅
- **Phase 44 (The Great Deletion):** Fully purged the old Tauri/Python codebase, re-wired the Makefile, and successfully passed full E2E Wails integration tests (PR #149 merged).
- **Documentation:** Updated `README.md` and `wiki/` to reflect the new Go/Wails architecture.

### In Progress (Delegated to Jules) ⏳
- **P45-T1 (Native Click-Through):** Jules is implementing OS-level CGO hooks in `wails-app/main.go` (X11/Win32/Cocoa) to ignore mouse events, restoring the stealth click-through feature.
- **Phase 46 (Hybrid STT Engine):** Tasks `P46-T1`, `P46-T2`, and `P46-T3` are fully scoped, planned, and queued up for Jules. They involve creating an STT interface, a hardware profiler/downloader, and dynamic `sherpa-onnx` hot-swapping.

### Up Next ⏭️
- Wait for Jules to submit PRs for `P45-T1` and `Phase 46`. 
- You can bulk merge Jules' PRs by running `bash scripts/merge_prs.sh <start> <end>`.
- Continue closing out any remaining UI/UX issues in the Svelte frontend.

---

## 🛠️ How to Resume Development
1. **Run locally:** `make dev`
2. **Task Tracker:** Check `prompts/tasks.md` for the roadmap.
3. **Submit to Jules:** Use `python3 scripts/jules_submit.py --task <ID>`. The script automatically enforces 100% test coverage and idiomatic Go architecture.
