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

| Task ID | File(s) | Description | Parallel? | Status | PR |
|---------|---------|-------------|-----------|--------|----|
| P54-T1 | `core/ports/driving/`, `core/ports/driven/` | **Core Port Interfaces:** Define all driving + driven port interfaces. Pure Go, zero external dependencies. | ❌ Sequential (first) | ⬜ | — |
| P54-T2 | `adapters/llm/openai_adapter.go` | **LLM Adapter:** Wrap existing `backend/llm/openai.go` behind `LLMPort` interface. | ⚡ Batch 2 | ⬜ | — |
| P54-T3 | `adapters/cache/sqlitevec_adapter.go` | **Cache Adapter:** Wrap existing `backend/vector_db.go` behind `CachePort` interface. | ⚡ Batch 2 | ⬜ | — |
| P54-T4 | `adapters/events/wails_adapter.go`, `adapters/events/noop_adapter.go` | **Events Adapter:** Create `WailsEventAdapter` wrapping `wailsruntime.EventsEmit` and a `NoopEventAdapter` for tests. | ⚡ Batch 2 | ⬜ | — |
| P54-T5 | `core/engine/engine.go`, `core/engine/pipeline.go` | **StealthEngine Core:** Wire all adapters into a single `StealthEngine` struct. Extract pipeline logic from `app.go` into `core/engine/pipeline.go`. | ❌ Sequential (after Batch 2) | ⬜ | — |
| P54-T6 | `products/presenter/`, `frontend/src/products/presenter/` | **StealthPresenter Product Skin:** First product using the engine — new system prompt + thin `PresenterApp` Wails shell + `PresenterHUD.svelte` scaffold. | ❌ Sequential (after T5) | ⬜ | — |

---

## Notes

- **STT already hexagonal** ✅ — `stt/engine.go`, `stt/manager.go`, `stt/groq.go`, `stt/whisper.go` are already Ports & Adapters. Do NOT touch them in this phase.
- **Window already hexagonal** ✅ — `backend/window/window_linux.go` already exists. T1 just adds the `WindowPort` interface.
- **No frontend changes in T1–T5** — UI scaffold only added in T6.
- Each task **must compile and pass `go test ./...`** before submitting PR.
- Prompt files for each task: `prompts/tasks/phase_54_hex_engine/P54_T{N}_*.txt`
