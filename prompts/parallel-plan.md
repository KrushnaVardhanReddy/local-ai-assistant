# ⚡ Jules Parallel Execution Plan

> Use this file to know **which tasks can be submitted to Jules simultaneously**.
> Jules opens each task as a separate branch/PR — tasks in the same wave have no
> file-level conflicts and can all be in-flight at once.

## How to submit a wave

```bash
# Submit all tasks in a wave at once
python3 scripts/jules_submit.py --task P1-T2
python3 scripts/jules_submit.py --task P4-T1    # run in parallel terminal
python3 scripts/jules_submit.py --task P6-T1    # run in parallel terminal

# Then wait for those PRs → review → merge → move to Wave 2
bash scripts/merge_prs.sh <start_pr> <end_pr>

# Then submit Wave 2, and so on
```

---

## Dependency Graph

```
P1-T2 ──────────────────────────────────────────┐
                                                  ├── P1-T1
                                                  ├── P2-T1 ──── P2-T2 ──┐
                                                  ├── P2-T3               ├── P3-T2 ──── P5-T1
                                                  └── P3-T1 ──────────────┘         └── P6-T4 ──── P6-T5

P4-T1 ──── P4-T2 ──── P4-T3 ──── P4-T4 ──────────────────────────────────────────────── P6-T5
       └── P4-T5 (parallel with P4-T2/T3)

P6-T1 ──── P6-T2 ─────────────────────────────── P6-T4
       └── P6-T3 (parallel with P6-T2)
```

---

## Wave 1 — No dependencies (submit all at once 🚀)

| Task | What it creates | Why independent |
|---|---|---|
| **P1-T2** | `backend/config.py` | Foundation module — no prior files needed |
| **P4-T1** | `frontend/` scaffold | Pure Rust/JS — no backend dependency |
| **P6-T1** | `web/` SvelteKit scaffold | Separate app — no backend dependency |

```bash
python3 scripts/jules_submit.py --task P1-T2
python3 scripts/jules_submit.py --task P4-T1
python3 scripts/jules_submit.py --task P6-T1
```

⏳ **Wait for all Wave 1 PRs to merge before Wave 2.**

---

## Wave 2 — After Wave 1 merges (5 tasks in parallel 🚀🚀)

| Task | Depends on | What it creates |
|---|---|---|
| **P1-T1** | P1-T2 | `backend/llm_health_check.py` |
| **P2-T1** | P1-T2 | `backend/audio_listener.py` |
| **P2-T3** | P1-T2 | `backend/requirements.txt`, `backend/README.md` |
| **P3-T1** | P1-T2 | `backend/llm_client.py` |
| **P4-T2** | P4-T1 | `frontend/src/lib/ws.ts` |
| **P4-T5** | P4-T1 | Tauri stealth mode in `main.rs` *(touches different files than P4-T2)* |
| **P6-T2** | P6-T1 | Supabase auth + DB schema |
| **P6-T3** | P6-T1 | Stripe billing *(different files from P6-T2)* |

> ⚠️ P4-T2 and P4-T5 both modify `main.rs` — submit them sequentially or expect a merge conflict. Submit P4-T2 first, merge, then P4-T5.

```bash
python3 scripts/jules_submit.py --task P1-T1
python3 scripts/jules_submit.py --task P2-T1
python3 scripts/jules_submit.py --task P2-T3
python3 scripts/jules_submit.py --task P3-T1
python3 scripts/jules_submit.py --task P4-T2   # merge before P4-T5
python3 scripts/jules_submit.py --task P6-T2
python3 scripts/jules_submit.py --task P6-T3   # parallel with P6-T2 (different files)
```

⏳ **Wait for Wave 2 PRs to merge. Then submit P4-T5.**

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
python3 scripts/jules_submit.py --task P4-T3   # parallel with P2-T2 (different stack)
```

---

## Wave 4 — After Wave 3 merges

| Task | Depends on | What it creates |
|---|---|---|
| **P3-T2** | P2-T1 + P2-T2 + P3-T1 | `backend/app.py` (main server) |
| **P4-T4** | P4-T3 + P4-T2 | `frontend/src/App.svelte` (root component) |

```bash
python3 scripts/jules_submit.py --task P3-T2
python3 scripts/jules_submit.py --task P4-T4   # parallel with P3-T2 (different stack)
```

---

## Wave 5 — After Wave 4 merges

| Task | Depends on | What it creates |
|---|---|---|
| **P5-T1** | P3-T2 + P4-T4 | Push-to-talk (touches both backend and frontend) |
| **P6-T4** | P3-T2 + P6-T2 + P6-T3 | FastAPI JWT middleware + key storage |

```bash
python3 scripts/jules_submit.py --task P5-T1
python3 scripts/jules_submit.py --task P6-T4   # parallel with P5-T1 (different files)
```

---

## Wave 6 — After Wave 5 merges (Final SaaS integration)

| Task | Depends on | What it creates |
|---|---|---|
| **P6-T5** | P6-T4 + P4-T4 | Tauri app auth + OS keychain |

```bash
python3 scripts/jules_submit.py --task P6-T5
```

---

## Summary Table

| Wave | Tasks | Can run in parallel? | Gate |
|---|---|---|---|
| **1** | P1-T2, P4-T1, P6-T1 | ✅ All 3 | — |
| **2** | P1-T1, P2-T1, P2-T3, P3-T1, P4-T2, P6-T2, P6-T3 | ✅ Most (see P4-T2/T5 note) | Wave 1 merged |
| **2b** | P4-T5 | ✅ Alone | P4-T2 merged |
| **3** | P2-T2, P4-T3 | ✅ Both | Wave 2 merged |
| **4** | P3-T2, P4-T4 | ✅ Both | Wave 3 merged |
| **5** | P5-T1, P6-T4 | ✅ Both | Wave 4 merged |
| **6** | P6-T5 | — | Wave 5 merged |

**Total tasks:** 16 &nbsp;|&nbsp; **Waves:** 6 &nbsp;|&nbsp; **Max parallel in Wave 2:** 7 tasks

---

## Tips

- If a Jules PR has merge conflicts, rebase it against `main` locally and push — don't re-submit.
- Always review the PR diff before merging — Jules occasionally adds extra boilerplate.
- `P3-T2` (app.py) is the most complex task — give it extra review time before merging.
- Phases 1–5 are the local desktop app. Phase 6 (SaaS) can start in parallel from Wave 2 onwards since it touches a separate `web/` directory.
