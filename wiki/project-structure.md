# Project Structure

The Local AI Assistant project is structured into several distinct components:

```text
Local_AI_Assistant/
├── wails-app/backend/          # Go Backend (AI & Audio Processing)
│   ├── app.py                  # FastAPI server + WebSocket broadcaster
│   ├── audio_listener.py       # Mic capture and live transcription
│   └── requirements.txt        # Go dependencies
├── frontend/                   # Svelte 5 + Wails Desktop App
│   ├── src/
│   │   ├── App.svelte          # Root component (floating overlay)
│   │   ├── lib/
│   │   │   ├── Assistant.svelte   # Main chat/response panel
│   │   │   └── ws.ts              # WebSocket store (reactive)
│   │   └── main.ts             # App entrypoint
│   ├── wails-app/              # Wails Go shell config
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

## Key Directories

### `backend/`
Contains the Go-based AI orchestration layer. This layer interacts directly with hardware (microphone, CUDA GPUs) to perform Speech-to-Text (using faster-whisper or Parakeet) and interfaces with the LLM API (Ollama, LM Studio, etc.). It exposes a WebSocket connection for the frontend.

### `frontend/`
Contains the user interface, built with Svelte 5 and bundled as a native desktop application using Wails v2. The UI is designed as a minimalist floating overlay (stealth mode) that connects to the backend over WebSocket to receive streaming LLM tokens.

### `prompts/`
A critical directory for the AI-assisted development workflow. It contains the master roadmap (`tasks.md`) and individual task prompts sent to Jules (Google's async coding agent) for execution via `scripts/jules_submit.py`.
