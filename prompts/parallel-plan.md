# ⚡ Jules Parallel Execution Plan

> Use this file to know **which tasks can be submitted to Jules simultaneously**.
> Jules opens each task as a separate branch/PR — tasks in the same wave have no
> file-level conflicts and can all be in-flight at once.

## How to submit a wave

```bash
# Submit all tasks in a wave at once (separate terminals or background jobs)
python3 scripts/jules_submit.py --task P1-T2
python3 scripts/jules_submit.py --task P4-T1    # separate terminal
python3 scripts/jules_submit.py --task P7-T1    # separate terminal

# Then wait for those PRs → review → merge → move to next wave
bash scripts/merge_prs.sh <start_pr> <end_pr>
```

---

## Dependency Graph

```
P1-T2 ──────────────────────────────┐
                                     ├── P1-T1
                                     ├── P2-T1 ──── P2-T2 ──┐
                                     ├── P2-T3               ├── P3-T2 ──────────────────────────────┐
                                     └── P3-T1 ──────────────┘                                        │
                                                                                                        ├── P5-T3 ──── P5-T4 ──── P6-T1
P4-T1 ──── P4-T2 ──── P4-T3 ──── P4-T4 ───────────────────────────────────────────────────────────────┘            └── P7-T5
       └── P4-T5 ↑parallel↑

P5-T1 ──── P5-T2 (after P3-T2 + P4-T4 merged)

P7-T1 ──── P7-T2 ─────────────────────────────── P7-T4
       └── P7-T3 (parallel with P7-T2)       └── P7-T5
```

---

## Wave 1 — No dependencies (submit all 3 at once 🚀)

| Task | What it creates | Why independent |
|---|---|---|
| **P1-T2** | `backend/config.py` | Foundation module — no prior files |
| **P4-T1** | `frontend/` scaffold | Pure Rust/JS — no backend dependency |
| **P7-T1** | `web/` SvelteKit scaffold | Separate app — no backend dependency |

```bash
python3 scripts/jules_submit.py --task P1-T2
python3 scripts/jules_submit.py --task P4-T1
python3 scripts/jules_submit.py --task P7-T1
```

⏳ **Wait for all Wave 1 PRs to merge before Wave 2.**

---

## Wave 2 — After Wave 1 merges (7 tasks in parallel 🚀🚀)

| Task | Depends on | What it creates |
|---|---|---|
| **P1-T1** | P1-T2 | `backend/llm_health_check.py` |
| **P2-T1** | P1-T2 | `backend/audio_listener.py` |
| **P2-T3** | P1-T2 | `backend/requirements.txt` |
| **P3-T1** | P1-T2 | `backend/llm_client.py` |
| **P4-T2** | P4-T1 | `frontend/src/lib/ws.ts` |
| **P7-T2** | P7-T1 | Supabase auth + DB schema |
| **P7-T3** | P7-T1 | Stripe billing *(different files from P7-T2)* |

> ⚠️ Submit **P4-T2 first**. After P4-T2 merges, submit P4-T5 separately (both touch `main.rs`).

```bash
python3 scripts/jules_submit.py --task P1-T1
python3 scripts/jules_submit.py --task P2-T1
python3 scripts/jules_submit.py --task P2-T3
python3 scripts/jules_submit.py --task P3-T1
python3 scripts/jules_submit.py --task P4-T2
python3 scripts/jules_submit.py --task P7-T2
python3 scripts/jules_submit.py --task P7-T3
```

After P4-T2 merges:
```bash
python3 scripts/jules_submit.py --task P4-T5
```

---

## Wave 3 — After Wave 2 merges

| Task | Depends on | What it creates |
|---|---|---|
| **P2-T2** | P2-T1 + P1-T2 | `backend/transcriber.py` |
| **P4-T3** | P4-T2 | `frontend/src/lib/Assistant.svelte` |

```bash
python3 scripts/jules_submit.py --task P2-T2
python3 scripts/jules_submit.py --task P4-T3
```

---

## Wave 4 — After Wave 3 merges

| Task | Depends on | What it creates |
|---|---|---|
| **P3-T2** | P2-T1 + P2-T2 + P3-T1 | `backend/app.py` (main FastAPI server) |
| **P4-T4** | P4-T3 + P4-T2 | `frontend/src/App.svelte` |

```bash
python3 scripts/jules_submit.py --task P3-T2
python3 scripts/jules_submit.py --task P4-T4
```

---

## Wave 5 — RAG Pipeline (after Wave 4 merges)

| Task | Depends on | What it creates |
|---|---|---|
| **P5-T1** | P3-T2 + P1-T2 | `backend/rag/ingestor.py` (ChromaDB + embeddings) |

```bash
python3 scripts/jules_submit.py --task P5-T1
```

---

## Wave 6 — RAG Retriever (after P5-T1 merges)

| Task | Depends on | What it creates |
|---|---|---|
| **P5-T2** | P5-T1 | `backend/rag/retriever.py` (semantic search) |

```bash
python3 scripts/jules_submit.py --task P5-T2
```

---

## Wave 7 — RAG Integration + SaaS Auth (after Wave 6 merges)

| Task | Depends on | What it creates |
|---|---|---|
| **P5-T3** | P5-T2 + P3-T2 | Inject RAG context into LLM prompt |
| **P7-T4** | P3-T2 + P7-T2 + P7-T3 | FastAPI JWT middleware + BYOK key storage |

```bash
python3 scripts/jules_submit.py --task P5-T3
python3 scripts/jules_submit.py --task P7-T4
```

---

## Wave 8 — Polish + Final SaaS (after Wave 7 merges)

| Task | Depends on | What it creates |
|---|---|---|
| **P5-T4** | P5-T3 + P4-T4 | KnowledgeBase.svelte drag-and-drop UI |
| **P6-T1** | P4-T4 + P3-T2 | Push-to-talk hotkey |
| **P7-T5** | P7-T4 + P4-T4 | Tauri app auth + OS keychain |

```bash
python3 scripts/jules_submit.py --task P5-T4
python3 scripts/jules_submit.py --task P6-T1
python3 scripts/jules_submit.py --task P7-T5
```

---

## Wave 9 — Final Polish (after Wave 8 merges)

| Task | Depends on | What it creates |
|---|---|---|
| **P6-T2** | P5-T4 + P6-T1 | Settings panel (model, mic, stealth toggles) |
| **P6-T3** | All backend | Build validation (PyInstaller + Tauri) |
| **P6-T4** | All | README final pass |

```bash
python3 scripts/jules_submit.py --task P6-T2
python3 scripts/jules_submit.py --task P6-T3
python3 scripts/jules_submit.py --task P6-T4
```

---

## Full Summary Table

| Wave | Tasks | Parallel? | Gate |
|---|---|---|---|
| **1** | P1-T2 · P4-T1 · P7-T1 | ✅ All 3 | — |
| **2** | P1-T1 · P2-T1 · P2-T3 · P3-T1 · P4-T2 · P7-T2 · P7-T3 | ✅ 7 at once | Wave 1 merged |
| **2b** | P4-T5 | — | P4-T2 merged |
| **3** | P2-T2 · P4-T3 | ✅ Both | Wave 2 merged |
| **4** | P3-T2 · P4-T4 | ✅ Both | Wave 3 merged |
| **5** | P5-T1 | — | Wave 4 merged |
| **6** | P5-T2 | — | P5-T1 merged |
| **7** | P5-T3 · P7-T4 | ✅ Both | Wave 6 merged |
| **8** | P5-T4 · P6-T1 · P7-T5 | ✅ All 3 | Wave 7 merged |
| **9** | P6-T2 · P6-T3 · P6-T4 | ✅ All 3 | Wave 8 merged |

**Total tasks:** 20 | **Phases:** 7 | **Max parallel in Wave 2:** 7 tasks

---

## Tips

- If a Jules PR has merge conflicts, rebase it against `main` locally and push.
- P3-T2 (`app.py`) is the most complex task — give it extra review time.
- Phases 1–6 are the local desktop app. Phase 7 (SaaS) can start from Wave 2 onwards since it lives in `web/`.
- The RAG phases (P5-T1 → P5-T4) are sequential — each one builds on the previous. Don't parallelize within Phase 5.
