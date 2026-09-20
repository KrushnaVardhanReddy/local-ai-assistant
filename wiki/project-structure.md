# Project Structure

The Local AI Assistant project is structured into several distinct components, built around a pure Go Hexagonal Architecture.

```text
Local_AI_Assistant/
├── wails-app/                  # Go Backend + Wails Shell
│   ├── core/                   # The Hexagonal Domain
│   │   ├── ports/              # Interfaces (driving & driven)
│   │   └── engine/             # StealthEngine (business logic)
│   ├── adapters/               # Infrastructure implementations
│   │   ├── llm/                # OpenAI/Local LLM adapters
│   │   ├── cache/              # SQLiteVec cache adapters
│   │   ├── events/             # Wails event emitters
│   │   └── window/             # OS-level stealth window APIs
│   ├── backend/                # Utility packages (STT, Parser, audio)
│   │   ├── stt/                # Whisper.cpp integration
│   │   └── parser/             # Document extraction (.pdf, .pptx, .md)
│   ├── products/               # Product-specific wiring & prompts
│   │   ├── presenter/          # StealthPresenter configuration
│   │   └── interview/          # BarnOwl AI configuration (Paused)
│   └── frontend/               # Svelte 5 UI
│       ├── src/
│       │   ├── products/       # Product-specific Svelte components
│       │   └── lib/            # Shared UI components
├── prompts/                    # Jules task prompts for async AI development
│   ├── tasks_v2.md             # Master task tracker (Phase 54+)
│   └── tasks/                  # One prompt file per Jules task
├── scripts/
│   └── jules_submit.py         # Submit tasks to Jules API → GitHub PRs
└── README.md
```

## Key Directories

### `wails-app/core/` (The Engine)
This is the heart of the application. The `StealthEngine` resides here, orchestrating audio processing, STT transcription, and LLM querying. It relies entirely on interfaces defined in `core/ports/`, ensuring that it is completely decoupled from any specific UI framework or database.

### `wails-app/adapters/` (The Infrastructure)
Implementations of the ports. For example, `adapters/llm/openai_adapter.go` implements the `driven.LLMPort` interface, supporting the hybrid local/cloud model approach. The Cache adapter powers the Cache Manager UI using `sqlite-vec` for RAG and vector search.

### `wails-app/products/` (The Products)
The engine is generic. The products define what the engine *does*. The `products/presenter/` directory contains the specific system prompts, UI configurations, and initializations required to turn the generic `StealthEngine` into **StealthPresenter**. This allows us to build multiple applications (MentorGlass, GovBrief) on the exact same backend.

### `wails-app/frontend/`
Contains the user interface, built with Svelte 5. It connects to the Go backend via Wails' IPC (Inter-Process Communication) event bus, acting purely as a dumb terminal to display data and capture UI events.

## Build and Packaging

The application includes a root `Makefile` that wraps the Wails CLI commands for creating production-ready binaries for different platforms.

* `make build-mac`: Builds a universal binary (`darwin/universal`) for macOS.
* `make build-windows`: Builds a binary (`windows/amd64`) for Windows.
* `make build-linux`: Builds a binary (`linux/amd64`) for Linux.
* `make build-all`: Runs the build commands for all three platforms sequentially.

*Note: For the time being, Apple code signing is not handled via these commands; they strictly build the raw `.app` and `.exe` bundles.*
