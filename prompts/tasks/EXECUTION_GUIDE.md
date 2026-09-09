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


> 📦 Previous batches (1-5) for Phases 1-25 have been completed.

### BATCH 6 — Phase 26 Enterprise (Parallel Safe — All touch different files)
**No prerequisites beyond main being fully up to date.**

| Priority | File | Task | Touches |
|---|---|---|---|
| 🔴 1st | `phase_26_enterprise/P26_T1_sso_integration` | Enterprise SSO (SAML/Supabase) | `web/src/routes/login/`, `web/src/routes/api/enterprise/sso-init/`, `web/supabase/migrations/` |
| 🔴 2nd | `phase_26_enterprise/P26_T2_seat_management` | Admin seat management dashboard | `web/src/routes/dashboard/admin/`, `web/src/routes/api/enterprise/` (different files) |
| 🟠 3rd | `phase_26_enterprise/P26_T3_team_knowledge_base` | Team RAG (shared knowledge base) | `backend/rag/team_retriever.py` (NEW), `backend/rag/retriever.py`, `backend/app.py`, `frontend/src/lib/KnowledgeBase.svelte` |

⚠️ Note: P26-T3 touches `backend/app.py`. Do NOT run any other task that also touches `backend/app.py` in the same batch.

```bash
python3 scripts/jules_submit.py --task P26-T1
python3 scripts/jules_submit.py --task P26-T2
python3 scripts/jules_submit.py --task P26-T3
# After P26-T3 and P26-T4 merged:
python3 scripts/jules_submit.py --task P26-T5
```

---

### BATCH 7 — Phase 27 Voice Mode (Sequential — build on each other)
**Wait for Batch 6 to merge.**

| Priority | File | Task | Touches |
|---|---|---|---|
| 🔴 1st | `phase_27_voice_mode/P27_T1_wake_word` | Wake word detection | `backend/wake_word.py` (NEW), `backend/config.py`, `backend/app.py`, `frontend/src/lib/Settings.svelte` |
| 🔴 2nd | `phase_27_voice_mode/P27_T2_tts_engine` | TTS engine | `backend/tts.py` (NEW), `backend/config.py`, `backend/smart_filter.py`, `backend/app.py` |

⚠️ P27-T1 and P27-T2 BOTH touch `backend/config.py` and `backend/app.py`. Submit them SEQUENTIALLY (wait for P27-T1 to merge before P27-T2).

```bash
# Submit P27-T1 first and wait for it to merge, THEN:
python3 scripts/jules_submit.py --task P27-T1
# After P27-T1 merged:
python3 scripts/jules_submit.py --task P27-T2
# After P27-T2 merged:
python3 scripts/jules_submit.py --task P27-T3
# After P27-T3 merged:
python3 scripts/jules_submit.py --task P27-T4
```

| 🟡 3rd | `phase_27_voice_mode/P27_T3_voice_mode_ui` | Glowing orb UI | `frontend/src/lib/Assistant.svelte` only |
| 🟢 4th | `phase_27_voice_mode/P27_T4_e2e_tests` | E2E Tests | `frontend/e2e/`, `tests/e2e/` |

---

### BATCH 8 — Phase 28 Agentic (Sequential — strict dependency chain)
**Wait for Batch 7 to merge.**

| Priority | File | Task | Touches |
|---|---|---|---|
| 🔴 1st | `phase_28_agentic/P28_T1_tool_registry` | Tool calling infra + BaseTool | `backend/tools/base_tool.py` (NEW), `backend/tools/__init__.py` (NEW), `backend/llm_client.py`, `backend/config.py` |
| 🔴 2nd | `phase_28_agentic/P28_T2_os_tool_pack` | 6 built-in OS tools | `backend/tools/open_app.py`, `type_text.py`, `get_clipboard.py`, `set_clipboard.py`, `web_search.py`, `show_notification.py` (ALL NEW) |
| 🔴 3rd | `phase_28_agentic/P28_T3_agentic_loop` | Agent loop + UI toggle | `backend/app.py`, `frontend/src/lib/Assistant.svelte` |
| 🔴 4th | `phase_28_agentic/P28_T4_mcp_support` | MCP Client Integration | `backend/mcp/client.py` (NEW), `backend/tools/registry.py`, `backend/config.py`, `frontend/src/lib/Settings.svelte` |
| 🟢 5th | `phase_28_agentic/P28_T5_e2e_tests` | E2E Tests | `tests/e2e/` |

⚠️ P28 tasks MUST be submitted in strict order (T1 → T2 → T3 → T4). Each depends on the previous.
P28-T2 is fully safe to parallelize with any Phase 26 tasks (only creates new files).

```bash
# Submit strictly in order:
python3 scripts/jules_submit.py --task P28-T1
# After P28-T1 merged:
python3 scripts/jules_submit.py --task P28-T2
# After P28-T2 merged:
python3 scripts/jules_submit.py --task P28-T3
# After P28-T3 merged:
python3 scripts/jules_submit.py --task P28-T4
# After P28-T4 merged:
python3 scripts/jules_submit.py --task P28-T5
```

---

### BATCH 9 — Phase 29 Resume Builder (Parallel Safe)
**No prerequisites beyond main being fully up to date.**

| Priority | File | Task | Touches |
|---|---|---|---|
| 🔴 1st | `phase_29_resume_builder/P29_T1_resume_gen_api` | Backend LLM Generation API | `backend/resume_builder.py` (NEW), `backend/app.py` |
| 🔴 2nd | `phase_29_resume_builder/P29_T2_resume_editor_ui` | Frontend Resume Editor | `frontend/src/lib/ResumeBuilder.svelte` (NEW), `frontend/src/lib/Assistant.svelte` |
| 🟡 3rd | `phase_29_resume_builder/P29_T3_pdf_export` | 5-10 CSS Templates & PDF Export | `frontend/src/lib/resume-styles.css` (NEW), `frontend/src/lib/ResumeBuilder.svelte` |
| 🟢 4th | `phase_29_resume_builder/P29_T4_e2e_tests` | E2E Tests | `frontend/e2e/`, `tests/e2e/` |

⚠️ P29-T1 and P29-T2 can be submitted in parallel (backend vs frontend). Submit P29-T3 sequentially after P29-T2 merges.

```bash
python3 scripts/jules_submit.py --task P29-T1
python3 scripts/jules_submit.py --task P29-T2
# After P29-T2 merged:
python3 scripts/jules_submit.py --task P29-T3
# After P29-T3 merged:
python3 scripts/jules_submit.py --task P29-T4
```

