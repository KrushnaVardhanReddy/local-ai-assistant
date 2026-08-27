# 🤖 Local AI Assistant

![Python](https://img.shields.io/badge/Python-3.10+-3776AB?style=flat-square&logo=python&logoColor=white)
![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=flat-square&logo=svelte&logoColor=white)
![Tauri](https://img.shields.io/badge/Tauri-2.0-FFC131?style=flat-square&logo=tauri&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)
![Status](https://img.shields.io/badge/Status-In%20Development-orange?style=flat-square)
![VRAM](https://img.shields.io/badge/VRAM-8GB%20Target-purple?style=flat-square&logo=nvidia)
![SaaS](https://img.shields.io/badge/SaaS%20Roadmap-BYOK-blueviolet?style=flat-square)

A real-time, context-aware desktop assistant that runs **completely offline** on local hardware. This project uses local Speech-to-Text (STT) and a lightweight Large Language Model (LLM) to process voice commands with **zero latency** and **total privacy** — no cloud, no API keys, no data leaves your machine.

> **Inspired by [Parakeet](https://github.com/NVIDIA/NeMo)** — NVIDIA's state-of-the-art CTC-based ASR model optimized for real-time, low-latency streaming transcription on CUDA hardware.

---

## 🚀 Architecture Overview

This application uses a **dual-process architecture** to separate heavy AI compute from the user interface:

1. **Frontend (UI):** Built with **Svelte 5 + Tauri 2.0** — a lightweight, native desktop shell with reactive UI for the floating assistant overlay.
2. **Backend (AI & Audio):** Built with **Python (FastAPI)** to capture microphone audio, handle real-time transcription, and stream LLM responses via WebSocket.

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

### 🦜 Why Parakeet for STT?

[NVIDIA Parakeet-TDT](https://huggingface.co/nvidia/parakeet-tdt-1.1b) is purpose-built for real-time audio transcription:

| Feature | Parakeet-TDT | Whisper |
|---|---|---|
| Architecture | CTC + Token & Duration Transducer | Encoder-Decoder |
| Latency | Streaming-optimized (chunk-level) | Full-utterance decode |
| CUDA Optimization | First-class NVIDIA NeMo support | Via `faster-whisper` (CTranslate2) |
| Accuracy (en) | State-of-the-art on LibriSpeech | Very strong, but slower |
| Use Case | Real-time voice assistant | Batch transcription |

### 🧩 Why Svelte 5 + Tauri?

| | Plain JS/TS | Electron | **Tauri + Svelte 5** ✅ |
|---|---|---|---|
| RAM Usage | Low | ~150MB+ | ~10–30MB |
| Bundle Size | Small | ~200MB | ~5–15MB |
| Reactive UI | Manual DOM | React/Svelte | **Svelte 5 runes** |
| WebSocket | ✅ | ✅ | ✅ |
| Native desktop | ❌ | ✅ | ✅ |
| Learning curve | Low | Medium | **Low** (if you know Svelte) |

---

## ⚡ Quick Start

### Prerequisites

- [Python 3.10+](https://www.python.org/downloads/)
- [Node.js v18+](https://nodejs.org/)
- [Rust](https://rustup.rs/) (required by Tauri)
- [Ollama](https://ollama.com) — for local LLM hosting
- NVIDIA GPU with CUDA 11.8+ (for full GPU acceleration)

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/Local_AI_Assistant.git
cd Local_AI_Assistant
```

### 2. Start the Local LLM Brain

```bash
# Pull and run the default model (fits in ~5GB VRAM with 4-bit quantization)
ollama run llama3:8b-instruct-q4_K_M
```

### 3. Set Up the Python Backend

```bash
cd backend
pip install -r requirements.txt
python app.py
```

### 4. Launch the Desktop App

```bash
cd frontend
npm install
npm run tauri dev
```

---

## 💻 Hardware Requirements & Target

| Component | Spec |
|---|---|
| **Device** | Dell XPS |
| **GPU VRAM** | 8GB Dedicated |
| **LLM Allocation** | ~5GB VRAM (4-bit quantized 7B/8B model) |
| **STT Allocation** | ~1–2GB VRAM (Whisper Base/Small or Parakeet) |
| **System Overhead** | ~1GB VRAM (desktop environment) |

---

## 🛠️ Tech Stack

### Backend (Python)
| Library | Purpose |
|---|---|
| `faster-whisper` | Optimized STT via CTranslate2 + CUDA |
| `nemo_toolkit` | NVIDIA Parakeet-TDT integration |
| `pyaudio` / `sounddevice` | System microphone capture |
| `fastapi` + `uvicorn` | Local API + WebSocket server |
| `aiohttp` | Async communication with Ollama |

### Frontend (JS/TS)
| Tool | Purpose |
|---|---|
| **Svelte 5** | Reactive UI with runes — minimal boilerplate |
| **Tauri 2.0** | Rust-backed native desktop shell, replaces Electron |
| **WebSocket** | Streams live LLM response tokens from Python backend |

### AI Models
| Role | Model | Source |
|---|---|---|
| 🧠 Brain (LLM) | Any model (local or cloud) | See provider table below |
| 👂 Ears (STT) | Parakeet-TDT 1.1B / Whisper Base | [HuggingFace](https://huggingface.co/nvidia/parakeet-tdt-1.1b) |

---

## 🔌 LLM Provider Support

The app uses the **OpenAI-compatible `/v1/chat/completions` API** — the same client code works for every provider below. Just set `LLM_PROVIDER` in your `.env`.

### 🖥️ Local (Private — data never leaves your machine)
| Provider | Default URL | How to install |
|---|---|---|
| **Ollama** | `http://localhost:11434/v1` | [ollama.com](https://ollama.com) |
| **LM Studio** | `http://localhost:1234/v1` | [lmstudio.ai](https://lmstudio.ai) |
| **llama.cpp** | `http://localhost:8080/v1` | [github.com/ggerganov/llama.cpp](https://github.com/ggerganov/llama.cpp) |

> Set `LLM_PROVIDER=auto` (default) and the app will probe all three and use whichever is running.

### ☁️ Cloud (Requires API key — data leaves your machine)
| Provider | Model examples | API Key Env Var |
|---|---|---|
| **OpenAI** | `gpt-4o`, `gpt-4o-mini` | `OPENAI_API_KEY` |
| **Groq** | `llama-3.1-70b-versatile` | `GROQ_API_KEY` |
| **Google Gemini** | `gemini-1.5-flash` | `GEMINI_API_KEY` |
| **Anthropic** | `claude-3-5-sonnet` | `ANTHROPIC_API_KEY` |

> ⚠️ **Privacy note:** Cloud providers send your voice transcripts to external servers.
> Use local providers if privacy is a concern.

### Configuration (`.env`)
```bash
# Pick ONE of the following:

# Option A — Local (auto-detect)
LLM_PROVIDER=auto
LLM_MODEL=llama3

# Option B — Pin a local backend
LLM_PROVIDER=lmstudio
LLM_MODEL=llama-3-8b-instruct

# Option C — Cloud
LLM_PROVIDER=openai
LLM_MODEL=gpt-4o-mini
OPENAI_API_KEY=sk-...

# Option D — Groq (fast free tier)
LLM_PROVIDER=groq
LLM_MODEL=llama-3.1-70b-versatile
GROQ_API_KEY=gsk_...
```

---

## 📂 Project Structure

```text
Local_AI_Assistant/
├── backend/                    # Python Backend (AI & Audio Processing)
│   ├── app.py                  # FastAPI server + WebSocket broadcaster
│   ├── audio_listener.py       # Mic capture and live transcription
│   └── requirements.txt        # Python dependencies
├── frontend/                   # Svelte 5 + Tauri Desktop App
│   ├── src/
│   │   ├── App.svelte          # Root component (floating overlay)
│   │   ├── lib/
│   │   │   ├── Assistant.svelte   # Main chat/response panel
│   │   │   └── ws.ts              # WebSocket store (reactive)
│   │   └── main.ts             # App entrypoint
│   ├── src-tauri/              # Tauri Rust shell config
│   └── package.json
├── prompts/                    # Jules task prompts for async AI development
│   ├── tasks.md                # Master task tracker
│   └── tasks/                  # One prompt file per Jules task
├── scripts/
│   ├── jules_submit.py         # Submit tasks to Jules API → GitHub PRs
│   ├── merge_prs.sh            # Batch merge Jules PRs
│   └── build.sh                # Production build script
└── README.md
```

---

## 🗺️ Development Roadmap

> Tasks are tracked in [`prompts/tasks.md`](prompts/tasks.md) and executed via Jules one-by-one.

### ✅ Phase 1: Spin Up the Local Brain 🧠
1. Install **Ollama**.
2. Run `ollama run llama3:8b-instruct-q4_K_M` to load an 8B model optimized for your 5GB VRAM slot.
3. Test the local endpoint:
   ```bash
   curl http://localhost:11434/api/generate -d '{"model":"llama3","prompt":"Hello!"}'
   ```

### ⬜ Phase 2: Give the App Ears 👂
1. Create the `backend/` directory.
2. Implement `audio_listener.py` using `faster-whisper` or Parakeet-TDT.
3. Configure CUDA acceleration: `device="cuda"`, `compute_type="float16"`.
4. Verify sub-second transcription prints to terminal on voice input.

### ⬜ Phase 3: Bridge the Systems 🌁
1. Pipe transcription output from Phase 2 into the Ollama API payload in `app.py`.
2. Expose a **FastAPI WebSocket** endpoint so the frontend can receive streamed LLM responses.

### ⬜ Phase 4: Build the Floating UI 🖥️
1. Scaffold a **Svelte 5 + Tauri 2.0** project in `frontend/`.
2. Design a minimalist, transparent floating overlay panel.
3. Connect to the Python WebSocket via a Svelte reactive store to display real-time streamed tokens as you speak.

---

## 🤖 Jules Integration

This project uses [Jules](https://jules.google.com) for async AI-assisted development. Each phase task has a dedicated prompt file in `prompts/tasks/` that Jules executes and returns as a GitHub PR.

```bash
# Submit a single task to Jules
python3 scripts/jules_submit.py --task P1-T1

# Submit all tasks for a phase
python3 scripts/jules_submit.py --phase 1

# Submit a custom prompt file
python3 scripts/jules_submit.py --file prompts/tasks/P2_T1_audio_listener.txt

# Batch merge Jules PRs after review
bash scripts/merge_prs.sh 1 5
```

---

## 📝 Open Decisions

- [ ] **STT model:** Parakeet-TDT (best accuracy, needs NeMo) vs. Faster-Whisper (easier setup)?
- [ ] **Wake word:** Always-on mic vs. push-to-talk vs. hotword detection (e.g. `porcupine`)?
- [ ] **Overlay style:** Minimal pill/bubble vs. sidebar panel?

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

---

## 🙏 Acknowledgements

- [NVIDIA NeMo](https://github.com/NVIDIA/NeMo) — Parakeet ASR model family
- [Ollama](https://ollama.com) — Dead-simple local LLM hosting
- [faster-whisper](https://github.com/SYSTRAN/faster-whisper) — CTranslate2-optimized Whisper
- [Tauri](https://tauri.app) — Lightweight native desktop framework
- [Svelte](https://svelte.dev) — Compiler-first reactive UI framework
