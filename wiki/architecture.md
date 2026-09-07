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
```

## 1. The Backend (Python / FastAPI)
The Python backend is the core engine of the assistant. It is responsible for:
- **Audio Capture**: Listening to the system microphone.
- **Speech-to-Text (STT)**: Converting voice to text instantly. It supports `faster-whisper` for standard hardware and NVIDIA's **Parakeet-TDT** for extreme low-latency inference on CUDA devices.
- **LLM Orchestration**: Taking the transcribed text and routing it to the configured LLM provider (Ollama, LM Studio, or cloud APIs like OpenAI/Gemini/Groq).
- **WebSocket Server**: Streaming the response tokens back to the frontend in real-time.

## 2. The Frontend (Svelte 5 / Tauri)
The frontend serves purely as a dumb terminal/display layer for the backend's AI output.
- **Svelte 5**: Provides a reactive, lightweight UI using the new runes reactivity system.
- **Tauri 2.0**: Wraps the web app in a Rust-based native shell, consuming around 10-30MB of RAM (compared to Electron's 150MB+ footprint).

## Hardware Targets
The system is optimized for an 8GB VRAM footprint:
- ~5GB for the LLM (e.g., Llama 3 8B quantized to 4-bit)
- ~1-2GB for the STT model
- ~1GB reserved for desktop OS overhead
