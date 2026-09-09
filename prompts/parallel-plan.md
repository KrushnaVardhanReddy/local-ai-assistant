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


> 📦 Previous waves (1-9) for Phases 1-25 have been completed.

## Phase 26 — Enterprise B2B (Wave 10)

| Wave | Tasks | Parallel? | Gate |
|---|---|---|---|
| **10** | P26-T1 · P26-T2 · P26-T3 | ✅ T1 & T2 parallel. T3 runs alone (touches app.py) | main up to date |

```bash
# T1 and T2 can run together — they only touch web/ files
python3 scripts/jules_submit.py --task P26-T1
python3 scripts/jules_submit.py --task P26-T2
# After T1 & T2 merge, then:
python3 scripts/jules_submit.py --task P26-T3
```

---

## Phase 27 — Voice Conversational Mode (Wave 11)

| Wave | Tasks | Parallel? | Gate |
|---|---|---|---|
| **11a** | P27-T1 | ❌ Sequential | Wave 10 merged |
| **11b** | P27-T2 | ❌ Sequential | P27-T1 merged |
| **11c** | P27-T3 | ✅ Parallel with 11b | P27-T1 merged |

```bash
python3 scripts/jules_submit.py --task P27-T1
# After P27-T1 merged — these two can run in parallel:
python3 scripts/jules_submit.py --task P27-T2
python3 scripts/jules_submit.py --task P27-T3
```

---

## Phase 28 — Agentic Computer Control (Wave 12)

| Wave | Tasks | Parallel? | Gate |
|---|---|---|---|
| **12a** | P28-T1 | ❌ Sequential | Wave 11 merged |
| **12b** | P28-T2 | ✅ Parallel safe (new files only) | P28-T1 merged |
| **12c** | P28-T3 | ❌ Sequential | P28-T2 merged |
| **12d** | P28-T4 | ❌ Sequential | P28-T3 merged |

```bash
python3 scripts/jules_submit.py --task P28-T1
# After P28-T1 merged:
python3 scripts/jules_submit.py --task P28-T2
# After P28-T2 merged:
python3 scripts/jules_submit.py --task P28-T3
# After P28-T3 merged:
python3 scripts/jules_submit.py --task P28-T4
```

> 💡 **Tip for P28:** Since P28-T2 only creates NEW files in `backend/tools/`, it can
> safely run in parallel with unrelated frontend-only tasks if you have any queued up.
