# 🦉 BarnOwl AI — Invisible Interview Copilot & Stealth IDE Workstation
*(Formerly StealthPresenter — teleprompter mode paused; BarnOwl AI in active development)*

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white)
![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=flat-square&logo=svelte&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-2.15-ED2737?style=flat-square&logo=wails&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

**BarnOwl AI** is an undetectable, live AI interview co-pilot and developer workstation designed for high-stakes technical interviews, system design rounds, and live coding challenges.

Running 100% locally or with fast cloud inference (BYOK/Groq/Ollama), BarnOwl AI listens to interviewer speech in real-time, displays question transcript chips, automatically searches active workspace notes & resumes, and generates succinct bullet answers, code solutions, and STAR-method responses directly on an invisible, screen-share-immune HUD.

> ⏸️ **StealthPresenter (Teleprompter Mode):** Paused indefinitely. ShareSpeak and other generic teleprompters have commoditized script-scrolling with low-cost $14.99 lifetime licenses. The underlying auto-scrolling engine and presentation HUD remain preserved in the repository (`make dev-presenter`), but product focus has shifted to the high-value, live-intelligence AI Copilot.

---

## ✨ The 3 Core Moats

### 1. True OS-Level Stealth (Zero-Risk Screen Sharing)
Other apps use basic window transparency. If you accidentally share your "Entire Screen" in Zoom, your audience sees your script. 
StealthPresenter uses native OS APIs to redact the window at the driver level. **It is completely invisible to screen capture tools.** 

### 2. Voice-Tracked Auto-Scroller & Native Parser
Forget manual scrolling. StealthPresenter natively parses your `.pptx`, `.pdf`, or `.md` files without uploading them to a cloud server. The local STT engine (powered by single-file `whisperfile`) listens to your voice and flawlessly auto-scrolls your script as you speak.

### 3. The Real-Time LLM Audience Copilot
StealthPresenter doesn't just listen to you—it listens to your audience. If an interviewer asks a complex question, the Copilot instantly intercepts it, RAG-searches your loaded battlecards/script, and pops the perfect answer into your invisible HUD.

### 4. 100% Offline Privacy
Because StealthPresenter uses local STT (via `whisperfile`) and local LLMs (via `llamafile` / SmolLM / Gemma), **no data ever hits a server**. Perfect for enterprise compliance and confidential IP.

---

## 🏗️ Architecture & Tech Stack

StealthPresenter is built on top of the **StealthEngine**, a pure-Go Hexagonal Architecture pipeline designed to support multiple AI products.

1. **Frontend (Svelte 5 + Wails v2):** 
   - A highly optimized, glassmorphism UI overlay that consumes minimal RAM (~30MB).
   - Global stealth hotkeys for mouse-less interaction.
2. **StealthEngine (Backend):** 
   - Pure Go, completely decoupled from the UI.
   - **Core Ports:** `driving.PipelinePort`, `driven.LLMPort`, `driven.CachePort`, `driven.EventsPort`.
   - Captures microphone audio using native bindings and pipes it to local STT (`whisperfile` Cosmopolitan binary).
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
Copy `.env.example` to `.env.local` and configure any necessary paths for local models (e.g., ONNX embedding models, `llamafile` binaries, or `whisperfile` STT models).

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
