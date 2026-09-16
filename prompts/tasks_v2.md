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
| **P54-T3** | Cache Port Adapter | ✅ Merged | ⚡ Parallel | [PR 172](https://jules.google.com/session/529672600929029618) | Wraps backend/vector_db.go behind CachePort |
| **P54-T4** | Events Port Adapter | ✅ Merged | ⚡ Parallel | [PR 173](https://jules.google.com/session/9165291875719402956) | WailsEventAdapter + NoopEventAdapter |
| **P54-T5** | StealthEngine Core Pipeline | ✅ Merged | Sequential | [PR 174](https://jules.google.com/session/11318015680178095278) | Wires all adapters into engine.Start() |
| **P54-T6** | StealthPresenter Skin + HUD | ✅ Merged | Sequential | [PR 175](https://jules.google.com/session/8797094189760034044) | PresenterHUD.svelte + presenter build tag |

---

## Notes

- **STT already hexagonal** ✅ — `stt/engine.go`, `stt/manager.go`, `stt/groq.go`, `stt/whisper.go` are already Ports & Adapters. Do NOT touch them in this phase.
- **Window already hexagonal** ✅ — `backend/window/window_linux.go` already exists. T1 just adds the `WindowPort` interface.
- **No frontend changes in T1–T5** — UI scaffold only added in T6.
- Each task **must compile and pass `go test ./...`** before submitting PR.
- Prompt files for each task: `prompts/tasks/phase_54_hex_engine/P54_T{N}_*.txt`

---

## Phase 55 — StealthPresenter Core UX

> **Goal:** Build the pristine, glassmorphic UI overlay, the document parser (PDF/MD/PPTX), the voice-tracked scroller, and the LLM copilot panel.
>
> **Branch:** `feature/krushna_golang-stealth-ux`

### Dependency Order

```
BATCH 1 (Parallel execution)
  P55-T1 ⚡ The Presenter Shell & Typography (Svelte)
  P55-T2 ⚡ Document Parser Backend (Go)

BATCH 2 (Sequential)
  P55-T3: Voice-Scroller Engine (depends on T1)
  P55-T4: LLM Audience Copilot (depends on T1 & T3)
```

---

| Task ID | Title | Status | Parallel? | PR | Notes |
|---------|-------|--------|-----------|----|-------|
| **P55-T1** | The Presenter Shell (Svelte) | ✅ Merged | Sequential | [PR 176] | UI layout, opacity, font size sliders |
| **P55-T2** | Document Parser Backend | ✅ Merged | ⚡ Parallel | [PR 177] | Parses .txt, .md, .pdf, .pptx |
| **P55-T3** | Voice-Scroller Engine | ✅ Merged | Sequential | [PR 178] | Maps STT events to Svelte auto-scroll |
| **P55-T4** | LLM Audience Copilot | ✅ Merged | Sequential | [PR 179] | Live Q&A side-panel in HUD |

---

## Phase 56 — Multi-Document IDE Workspace (Hexagonal Core + Reusable Svelte)

> **Goal:** Build an IDE-style workspace (file/folder browser, expandable tree, multi-file tabs) in the hexagonal core engine so multiple products can load folders/files and switch active documents.
>
> **Branch:** `feature/krushna_golang`

### Dependency Order

```
BATCH 1 (Core Ports & Types)
  P56-T1: Core Workspace Port & Document Node Hierarchy (Go)

BATCH 2 (Parallel Execution)
  P56-T2 ⚡ StealthEngine Workspace Implementation & RAG Multi-Doc Indexing (Go)
  P56-T3 ⚡ Reusable IDE File Tree & Workspace Svelte Components (UI)

BATCH 3 (Sequential Product Integration)
  P56-T4: StealthPresenter Product Integration & Wails Bindings
```

---

| Task ID | Title | Status | Parallel? | PR | Notes |
|---------|-------|--------|-----------|----|-------|
| **P56-T1** | Core Workspace Port & Types | ✅ Merged | Sequential | [PR 180](https://github.com/KrushnaVardhanReddy/local-ai-assistant/pull/180) | `ports/driving/workspace.go`, `FileNode`, `WorkspacePort` |
| **P56-T2** | Engine Workspace & RAG Indexer | ✅ Merged | ⚡ Parallel | [PR 182](https://github.com/KrushnaVardhanReddy/local-ai-assistant/pull/182) | `core/engine/workspace.go`, directory scanning & doc caching |
| **P56-T3** | Svelte IDE Tree & Tabs UI | ✅ Merged | ⚡ Parallel | [PR 181](https://github.com/KrushnaVardhanReddy/local-ai-assistant/pull/181) | `lib/components/workspace/` (FileTree, Tabs, Sidebar) |
| **P56-T4** | StealthPresenter Integration | ✅ Merged | Sequential | [PR 183](https://github.com/KrushnaVardhanReddy/local-ai-assistant/pull/183) | `PresenterApp` bindings + `PresenterHUD` tabs/tree hookup |
| **P56-T5** | Universal IDE Shell Framework | ✅ Merged | Sequential | [PR 184](https://github.com/KrushnaVardhanReddy/local-ai-assistant/pull/184) | `IDEShell.svelte` (ActivityBar, Explorer, Tabs, StatusBar, Slots) |
