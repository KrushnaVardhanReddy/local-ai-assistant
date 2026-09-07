# Jules Prompt Execution Guide — Phases 15–20

Submit prompts to Jules using:
```bash
python3 scripts/jules_submit.py --task <filename_without_.txt>
```

---

## ⚡ Parallel Execution Batches

Each batch can have 2–3 Jules tasks running simultaneously with ZERO merge conflicts.
Wait for ALL tasks in a batch to be merged before starting the next batch.

---

### BATCH 1 — Start Now (All Parallel Safe)
**No prerequisites. Run all 3 simultaneously.**

| Priority | File | Task | Touches |
|---|---|---|---|
| 🔴 1st | `phase_18_portable_stealth/P18_T1_T2_portable_neutral_name` | Portable app + neutral process name | `tauri.conf.json`, `Cargo.toml` |
| 🔴 2nd | `phase_16_gemini_live/P16_T1_gemini_live_client` | New GeminiLiveClient module | `backend/gemini_live_client.py` (NEW) |
| 🟠 3rd | `phase_17_session_intelligence/P17_T1_session_manager` | New SessionManager + 2 endpoints | `backend/session_manager.py` (NEW), `backend/app.py` (new routes only) |

---

### BATCH 2 — After Batch 1 Merged (All Parallel Safe)

| Priority | File | Task | Touches |
|---|---|---|---|
| 🔴 1st | `phase_16_gemini_live/P16_T3_gemini_config` | Gemini config vars | `backend/config.py` (2 new fields), `.env.local` |
| 🟠 2nd | `phase_17_session_intelligence/P17_T2_scorecard_generation` | Scorecard LLM method | `backend/llm_client.py` (new method) |
| 🟡 3rd | `phase_18_portable_stealth/P18_T4_build_pipeline` | Build pipeline script | `Makefile` (new targets), `scripts/build_portable.sh` (NEW) |

---

### BATCH 3 — After Batch 2 Merged (All Parallel Safe)

| Priority | File | Task | Touches |
|---|---|---|---|
| 🔴 1st | `phase_16_gemini_live/P16_T2_gemini_app_pipeline` | Wire Gemini into WS pipeline | `backend/app.py` (startup + WS handler) |
| 🟠 2nd | `phase_15_speaker_diarization/P15_T1_diarization_backend` | Speaker diarization in Transcriber | `backend/transcriber.py`, `backend/config.py` (1 field), `backend/app.py` (1 line) |
| 🟡 3rd | `phase_17_session_intelligence/P17_T3_session_report_ui` | Session Report UI | `frontend/src/lib/SessionReport.svelte` (NEW), `frontend/src/lib/Assistant.svelte` |

---

### BATCH 4 — After Batch 3 Merged (All Parallel Safe)

| Priority | File | Task | Touches |
|---|---|---|---|
| 🟠 | `phase_15_speaker_diarization/P15_T2_diarization_routing` | Route diarized segments | `backend/smart_filter.py`, `backend/app.py` (WS routing) |
| 🟠 | `phase_15_speaker_diarization/P15_T3_diarization_ui` | Speaker badge chips | `frontend/src/lib/ws.svelte.ts`, `frontend/src/lib/Assistant.svelte` |
| 🟡 | `phase_16_gemini_live/P16_T4_gemini_settings_ui` | STT selector hide for Gemini | `frontend/src/lib/Settings.svelte` |
| 🟡 | `phase_17_session_intelligence/P17_T4_answer_coaching` | Live coaching hints | `backend/config.py` (1 field), `backend/app.py` (coaching task) |
| 🟡 | `phase_18_portable_stealth/P18_T3_configurable_name` | User-configurable process name | `frontend/src-tauri/src/lib.rs`, `settings.json.example` (NEW) |

⚠️ Note: P15-T2 and P17-T4 both touch `backend/app.py`. Submit them sequentially within Batch 4, not truly parallel. All others in Batch 4 are fully parallel.

---

### BATCH 5 — Phase 19 & 20 (Needs Auth/Billing Infrastructure)
**Submit Phase 19 tasks in sequence (many share auth.py). Phase 20 after schema applied.**

**Phase 20 — Anti-Sharing (run first within Batch 5):**
1. `phase_20_anti_sharing/P20_T1_supabase_schema` — Apply SQL schema in Supabase (no code)
2. `phase_20_anti_sharing/P20_T2_T3_device_session_lock` — Device + session enforcement
3. `phase_20_anti_sharing/P20_T4_T5_device_dashboard` — Dashboard + frontend handlers

**Phase 19 — Pricing (parallel with Phase 20 where possible):**
> Phase 19 prompts are in `phase_19_pricing_billing/` — see individual files for details.
> These require Stripe account setup and are lower priority than the core app features.

---

## 📋 Priority Color Key
- 🔴 Critical path — blocks next batch
- 🟠 High value — ship ASAP
- 🟡 Important but non-blocking

## 🚫 Conflict Rules
- **Never run two Jules tasks that touch the same file in the same batch**
- `backend/app.py` is the most contested file — check each prompt's "Files touched" row
- `backend/config.py` — only one task per batch should touch it
- `frontend/src/lib/Assistant.svelte` — only one task per batch
