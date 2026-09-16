# 🎯 StealthPresenter — The Ultimate B2B Presentation HUD

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white)
![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=flat-square&logo=svelte&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-2.15-ED2737?style=flat-square&logo=wails&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

**StealthPresenter** is an invisible, AI-powered teleprompter and co-pilot designed for high-stakes B2B sales pitches, webinars, and founder fundraising.

Unlike generic teleprompter apps, StealthPresenter operates as a completely invisible HUD (Heads-Up Display) that tracks your voice, manages your presentation notes, and acts as an instant "Panic Button" for live Q&A—all running 100% locally on your machine for maximum privacy.

---

## ✨ The 3 Core Moats

### 1. True OS-Level Stealth (Zero-Risk Screen Sharing)
Other apps use basic window transparency. If you accidentally share your "Entire Screen" in Zoom, your audience sees your script. 
StealthPresenter uses native OS APIs (like `SetCaptureExcluded` on Windows) to redact the window at the driver level. **It is completely invisible to screen capture tools.** 

### 2. The Real-Time LLM "Panic Button"
StealthPresenter doesn't just scroll text—it listens. If an audience member asks a complex question about a competitor, the local Whisper STT catches it. Hit a hotkey, and the LLM instantly RAG-searches your battlecards and pops the answer into your invisible HUD.

### 3. 100% Offline Privacy
Sales reps and founders cannot upload sensitive slide decks or confidential IP to a cloud teleprompter. Because StealthPresenter uses local Whisper (STT) and local LLMs (via Ollama/Llama.cpp), **no data ever hits a server**. 

---

## 🏗️ Architecture & Tech Stack

StealthPresenter is built on top of the **StealthEngine**, a pure-Go Hexagonal Architecture pipeline designed to support multiple AI products.

1. **Frontend (Svelte 5 + Wails v2):** 
   - A highly optimized, glassmorphism UI overlay that consumes minimal RAM (~30MB).
   - Global stealth hotkeys for mouse-less interaction.
2. **StealthEngine (Backend):** 
   - Pure Go, completely decoupled from the UI.
   - **Core Ports:** `driving.PipelinePort`, `driven.LLMPort`, `driven.CachePort`, `driven.EventsPort`.
   - Captures microphone audio using native bindings and pipes it to local STT (`whisper.cpp`).
3. **Data Layer (SQLite + ChromaDB):**
   - SQLite handles local caching and vector embeddings for instant document retrieval.

---

## ⚡ Quick Start (Developer Setup)

### Prerequisites
- [Go 1.25+](https://go.dev/)
- [Node.js v18+](https://nodejs.org/)
- [Wails CLI v2+](https://wails.io/)

### 1. Clone & Install
```bash
git clone https://github.com/KrushnaVardhanReddy/local-ai-assistant.git
cd local-ai-assistant
```

### 2. Configure Environment
Copy `.env.example` to `.env.local` and configure any necessary paths for local models (e.g., ONNX embedding models, Whisper models).

### 3. Launch the Application
```bash
cd wails-app
wails dev
```

---

## ⌨️ Advanced Stealth Hotkeys

StealthPresenter is designed to be operated entirely without a mouse while presenting, ensuring you never look distracted.

| Shortcut | Action | Description |
|---|---|---|
| **`Ctrl+Shift+Space`** | Panic Button (Q&A) | Instantly searches your script/deck for an answer to the last heard question. |
| **`Ctrl+Shift+X`** | Panic Hide | Clears the transcript and hides the app instantly. |
| **`Ctrl+Shift+Down`** | Force Scroll | Manually bumps the teleprompter down if voice-tracking lags. |

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
