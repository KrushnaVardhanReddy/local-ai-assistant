# System Architecture

The Local AI Assistant operates on a **dual-process architecture** designed to completely isolate the heavy AI computation from the user interface.

## High-Level Flow

```
┌────────────────────────────────┐       ┌────────────────────────────────┐
│    Frontend (Svelte 5 + Tauri) │ ◄───► │       Backend (Python)          │
│    (Floating Desktop Overlay)  │  WS   │  (Audio Capture & AI Orchestr) │
└────────────────────────────────┘       └───────────────┬────────────────┘
                                                          │
                           ┌──────────────────────────────┴──────────────────────────────┐
                           ▼                                                              ▼
               ┌─────────────────────────────┐                       ┌─────────────────────────────┐
               │    Speech-to-Text (STT)     │                       │     Local Brain (Ollama)     │
               │ (Faster-Whisper / Parakeet) │                       │    (Llama 3 / Mistral 7B)   │
               └─────────────────────────────┘                       └─────────────────────────────┘
                                                                            ▲
                                                                            │
               ┌─────────────────────────────┐                       ┌──────┴──────────────────────┐
               │  Local Intelligence Layer   │ ◄───────────────────► │ Semantic Cache (ChromaDB)   │
               │ (SmolLM2-135M Turn Gating)  │                       │ (Zero-latency Q&A lookups)  │
               └─────────────────────────────┘                       └─────────────────────────────┘
```

## 1. The Backend (Python / FastAPI)
The Python backend is the core engine of the assistant. It is responsible for:
- **Audio Capture**: Listening to the system microphone.
- **Speech-to-Text (STT)**: Converting voice to text instantly. It supports `faster-whisper` for standard hardware and NVIDIA's **Parakeet-TDT** for extreme low-latency inference on CUDA devices.
- **Local Intelligence Layer**: Uses an on-device `SmolLM2` model to act as a gatekeeper. It checks if the user has finished their thought before allowing expensive cloud LLM calls, and handles vector embeddings for the semantic QA cache.
- **Semantic Q&A Cache**: Stores past Q&A embeddings in `ChromaDB` locally. If a matching question is detected, it serves the answer instantly, saving cloud costs.
- **LLM Orchestration**: Taking the transcribed text (if not cached) and routing it to the configured LLM provider (Ollama, LM Studio, or cloud APIs like OpenAI/Gemini/Groq).
- **WebSocket Server**: Streaming the response tokens back to the frontend in real-time.

## 2. The Frontend (Svelte 5 / Tauri)
The frontend serves purely as a dumb terminal/display layer for the backend's AI output.
- **Svelte 5**: Provides a reactive, lightweight UI using the new runes reactivity system.
- **Tauri 2.0 (Desktop)**: Wraps the web app in a Rust-based native shell, consuming around 10-30MB of RAM (compared to Electron's 150MB+ footprint).
- **Capacitor (Mobile)**: Wraps the same Svelte app into native iOS and Android apps, leveraging native microphone APIs and background-audio tasks.

## Hardware Targets
The system is optimized for an 8GB VRAM footprint:
- ~5GB for the LLM (e.g., Llama 3 8B quantized to 4-bit)
- ~1-2GB for the STT model
- ~1GB reserved for desktop OS overhead
