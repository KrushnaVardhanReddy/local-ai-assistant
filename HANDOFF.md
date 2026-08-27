# 🚀 Local AI Assistant — Project Handoff

*This document summarizes the current state of the Local AI Assistant project as of August 2026, so you can easily resume development later.*

---

## 🎯 Architecture Summary
We have evolved the initial Local Whisper concept into a full **Anti-Piracy, Split-Compute SaaS Platform**. 

**Core Stack:**
- **Backend:** Python + FastAPI + `faster-whisper` + ChromaDB + Ollama/OpenAI API.
- **Desktop Client:** Tauri v2 + Svelte 5.
- **SaaS Web App:** SvelteKit + Supabase + Stripe.

**Killer Features Designed:**
- **Stealth Mode:** OS-level kernel hooks (`WDA_EXCLUDEFROMCAPTURE`) to hide the app from screen shares. (Linux workaround: Window sharing vs Full Screen).
- **Split-Compute (Remote Helper):** Backend runs on a heavy machine (or a remote friend's machine via Cloudflare Tunnels), while the candidate runs the lightweight UI locally.
- **Hardware ID Locking:** Tauri generates a `machine_id` saved in the OS keychain to prevent credential sharing on the $49 lifetime license.
- **Vision Copilot (Phase 6):** Silent OS screenshots read the Zoom chat/code using GPT-4o.

---

## 🚦 Current Status: Wave 3 In-Flight

We are orchestrating development through **Jules**. 

**✅ What is DONE (Merged to `main`):**
- **Wave 1, Wave 2, and Wave 3 are completely finished and merged.**
- We have the `faster-whisper` transcriber logic.
- We have the main `Assistant.svelte` floating UI.
- We have the Tauri Stealth Mode window hooks.
- We have the SvelteKit scaffolding, Supabase Auth schema, and Stripe Webhooks.
- We have the Python audio listener, async LLM client, and health checks.
- The Svelte 5 reactive WebSocket store (`ws.svelte.ts`) is fully functional.

**⏳ What is IN PROGRESS (Waiting on Jules):**
- We just triggered **Wave 8** via `jules_submit.py`. We are waiting for:
  1. `P5-T4`: `KnowledgeBase.svelte` (drag-and-drop file upload UI)
  2. `P6-T1`: `main.rs`, `Assistant.svelte` (Vision capture hotkey)
  3. `P8-T5`: `frontend/src/lib/auth.ts` (Tauri app auth + OS keychain)

---

## ⏭️ Next Steps to Resume

1. **Review and Merge Wave 8:** 
   When you return, check GitHub for PRs related to `P5-T4`, `P6-T1`, and `P8-T5`. Merge them using:
   ```bash
   bash scripts/merge_prs.sh <start_pr_number> <end_pr_number>
   ```
2. **Test the Pipeline End-to-End!**
   Once Wave 4 is merged, the core pipeline is completely connected. You can start the FastAPI server and test the audio transcription and UI streaming!
3. **Trigger Wave 9 (Final Polish):**
   After testing, run:
   ```bash
   python3 scripts/jules_submit.py --task P7-T1
   python3 scripts/jules_submit.py --task P7-T2
   python3 scripts/jules_submit.py --task P7-T3
   python3 scripts/jules_submit.py --task P7-T4
   python3 scripts/jules_submit.py --task P7-T6
   ```

---

## 📂 Important Files to Know
- `prompts/parallel-plan.md` — The exact order to submit Jules tasks to avoid merge conflicts.
- `prompts/tasks.md` — The global tracker (I just updated it so all completed tasks are checked off).
- `LAUNCH_POST.md` — The marketing copy you can use to blast Reddit, Twitter, and Product Hunt when you launch!

Good luck, and get ready to launch a game-changing product!
