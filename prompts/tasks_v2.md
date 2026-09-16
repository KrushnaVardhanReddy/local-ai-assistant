# 🏗️ BarnOwl AI v2 — Jules Task Tracker (Hexagonal Engine)

> Submit tasks to Jules one-by-one or in parallel batches:
> ```bash
> python3 scripts/jules_submit.py --task P54-T1
> python3 scripts/jules_submit.py --list   # see all tasks
> ```
> Update status and PR numbers here as Jules returns PRs.

## Legend
| Symbol | Meaning |
|--------|---------|
| ✅ | Done / PR merged |
| 🔄 | Jules working on it |
| ⬜ | Not started |
| ⚡ | Can run in PARALLEL with other ⚡ tasks in same batch |

---

## Phase 54 — Hexagonal Core Engine (Ports & Adapters)

> **Goal:** Refactor the Go backend into a clean Ports & Adapters architecture
> so multiple products (Interview, StealthPresenter, MentorGlass, ClinicHUD)
> can share the same engine with different UI skins and system prompts.
>
> **Branch:** `feature/krushna_golang-hex-engine`

### Dependency Order (What blocks what)

```
BATCH 1 (Foundation — must go first, sequential)
  P54-T1: Core Port Interfaces

BATCH 2 (All 3 can run in PARALLEL after T1 merges)
  P54-T2 ⚡ LLM Port Adapter
  P54-T3 ⚡ Cache Port Adapter
  P54-T4 ⚡ Events Port Adapter

BATCH 3 (Sequential, depends on all of Batch 2)
  P54-T5: StealthEngine (wires adapters into core pipeline)

BATCH 4 (Sequential, depends on T5)
  P54-T6: StealthPresenter Product Skin (first new product)
```

---

| Task ID | Title | Status | Parallel? | PR | Notes |
|---------|-------|--------|-----------|----|-------|
| **P54-T1** | Core Port Interfaces | ✅ Merged | Sequential | [PR 170](https://jules.google.com/session/48906415451042314) | Foundation: ports/driving.go, ports/driven.go |
| **P54-T2** | LLM Port Adapter | ✅ Merged | ⚡ Parallel | [PR 171](https://jules.google.com/session/15853767921854849554) | Wraps backend/llm/ behind LLMPort |
| **P54-T3** | Cache Port Adapter | 🔄 In Progress | ⚡ Parallel | [529672600929029618](https://jules.google.com/session/529672600929029618) | Wraps backend/vector_db.go behind CachePort |
| **P54-T4** | Events Port Adapter | 🔄 In Progress | ⚡ Parallel | [9165291875719402956](https://jules.google.com/session/9165291875719402956) | WailsEventAdapter + NoopEventAdapter |
| **P54-T5** | StealthEngine Core Pipeline | ⬜ Queued | Sequential | — | Wires all adapters into engine.Start() |
| **P54-T6** | StealthPresenter Skin + HUD | ⬜ Queued | Sequential | — | PresenterHUD.svelte + presenter build tag |

---

## Notes

- **STT already hexagonal** ✅ — `stt/engine.go`, `stt/manager.go`, `stt/groq.go`, `stt/whisper.go` are already Ports & Adapters. Do NOT touch them in this phase.
- **Window already hexagonal** ✅ — `backend/window/window_linux.go` already exists. T1 just adds the `WindowPort` interface.
- **No frontend changes in T1–T5** — UI scaffold only added in T6.
- Each task **must compile and pass `go test ./...`** before submitting PR.
- Prompt files for each task: `prompts/tasks/phase_54_hex_engine/P54_T{N}_*.txt`
