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
- **Wave 1 & Wave 2 are completely finished and merged.**
- We have the SvelteKit scaffolding, Supabase Auth schema, and Stripe Webhooks.
- We have the Python audio listener, async LLM client, and health checks.
- The Svelte 5 reactive WebSocket store (`ws.svelte.ts`) is fully functional.

**⏳ What is IN PROGRESS (Waiting on Jules):**
- We just triggered **Wave 3** via `jules_submit.py`. We are currently waiting for Jules to open PRs for:
  1. `P2-T2`: The `faster-whisper` transcriber logic.
  2. `P4-T3`: The main `Assistant.svelte` floating UI.
  3. `P4-T5`: The Tauri Stealth Mode window hooks.

---

## ⏭️ Next Steps to Resume

1. **Review and Merge Wave 3:** 
   When you return, check GitHub for PRs related to `P2-T2`, `P4-T3`, and `P4-T5`. You can easily merge them using the helper script:
   ```bash
   bash scripts/merge_prs.sh <start_pr_number> <end_pr_number>
   ```
2. **Trigger Wave 4 (The Glue):**
   Once Wave 3 is merged, run:
   ```bash
   python3 scripts/jules_submit.py --task P3-T2
   python3 scripts/jules_submit.py --task P4-T4
   ```
   *Note: `P3-T2` builds the `app.py` FastAPI server. This is the critical piece that finally glues the audio, transcriber, and LLM together. Once this merges, the app will be fully testable end-to-end!*
3. **Trigger Phase 5 (Local RAG):**
   After the main app loop works, trigger `P5-T1` to build the ChromaDB document ingestion pipeline.

---

## 📂 Important Files to Know
- `prompts/parallel-plan.md` — The exact order to submit Jules tasks to avoid merge conflicts.
- `prompts/tasks.md` — The global tracker (I just updated it so all completed tasks are checked off).
- `LAUNCH_POST.md` — The marketing copy you can use to blast Reddit, Twitter, and Product Hunt when you launch!

Good luck, and get ready to launch a game-changing product!
