# System Architecture

The Local AI Assistant operates on a **Hexagonal Architecture** (Ports and Adapters), designed to completely isolate the heavy AI orchestration from infrastructure concerns (like how events are emitted or how LLM APIs are called).

## High-Level Flow (Hexagonal Model)

```
                            ┌────────────────────────────────────┐
                            │      Driving Adapters (Inputs)     │
                            │  (Audio Device, HTTP, UI Events)   │
                            └─────────────────┬──────────────────┘
                                              │
                                   ┌──────────▼──────────┐
                                   │    Driving Ports    │
                                   │  (PipelinePort etc) │
                                   └──────────┬──────────┘
                                              │
 ┌──────────────────────┐          ┌──────────▼──────────┐         ┌──────────────────────┐
 │                      │          │                     │         │                      │
 │    Driven Ports      │◄─────────┤   Stealth Engine    ├─────────►     Driven Ports     │
 │  (LLMPort, Cache)    │          │  (Core/Domain Logic)│         │ (Events, WindowPort) │
 │                      │          │                     │         │                      │
 └─────────┬────────────┘          └─────────────────────┘         └─────────┬────────────┘
           │                                                                 │
           │                                                                 │
           ▼                                                                 ▼
 ┌──────────────────────┐                                          ┌──────────────────────┐
 │  Driven Adapters     │                                          │  Driven Adapters     │
 │  (OpenAI, SQLiteVec) │                                          │  (Wails, X11/Win32)  │
 └──────────────────────┘                                          └──────────────────────┘
```

## 1. The Core Domain (`wails-app/core/engine/`)
The `StealthEngine` is the brain. It is responsible for:
- Orchestrating native system audio loopback capture, Voice Activity Detection (VAD), and routing to STT (via Groq or local Whisper bindings).
- Constructing prompts and routing them to the LLM.
- Handling local semantic caching to save on API costs.
- **Rule:** The engine cannot import *any* external libraries or Wails packages. It only communicates through interface definitions located in `core/ports/`.

## 2. Infrastructure Adapters (`wails-app/adapters/`)
Adapters plug into the core engine.
- **LLM Adapter**: Implements `driven.LLMPort`. We support a hybrid approach: local models (via Ollama/llama.cpp) or cloud models (Groq, Gemini, OpenRouter, Cloudflare Workers AI) configurable via `.env.local`.
- **Cache Adapter**: Implements `driven.CachePort`. Uses `sqlite-vec` to store embeddings locally for instant semantic Q&A lookup, as well as holding workspace RAG documents like resumes and job descriptions.
- **Events Adapter**: Implements `driven.EventsPort`. Usually powered by the Wails event bus, streaming updates to the frontend UI.
- **Window Adapter**: Implements `driven.WindowPort`. Uses OS-specific syscalls (like `SetCaptureExcluded`) to make the UI completely invisible to screen sharing.

## 3. The Frontend (Svelte 5 / Wails)
The frontend serves purely as a dumb terminal/display layer for the backend's AI output.
- **Svelte 5**: Provides a reactive, lightweight UI using the new runes reactivity system.
- **Wails v2 (Desktop)**: Provides a pure native desktop application shell using Go, completely replacing older web/Tauri concepts, consuming around 10-30MB of RAM (compared to Electron's 150MB+ footprint).

## Product Skins (`wails-app/products/`)
Because the `StealthEngine` is completely generic, we can create multiple distinct applications that share the same backend. A product (like **StealthPresenter** or **MentorGlass**) simply defines:
1. Which Svelte UI component to load.
2. The specific system prompt to inject into the LLM adapter.
3. The specific setup/teardown logic for that tool.
