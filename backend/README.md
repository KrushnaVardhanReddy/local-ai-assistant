# Local AI Assistant Backend

Backend service powering the Local AI Assistant with FastAPI, STT, and LLM integrations.

## Setup
`pip install -r requirements.txt`

## Run
`python3 app.py` (Starts FastAPI + WebSocket server on ws://127.0.0.1:8765)

## Required Environment Variables
| Variable       | Default  | Description                           |
|----------------|----------|---------------------------------------|
| `LLM_PROVIDER` | `auto`   | LLM provider (auto/openai/ollama/etc) |
| `LLM_MODEL`    | `llama3` | LLM model to use                      |
| `STT_MODEL`    | `base`   | STT model (base/small/medium/large)   |
| `STT_DEVICE`   | `cuda`   | STT device (`cuda` or `cpu`)          |
| `WS_PORT`      | `8765`   | WebSocket server port                 |

## Optional SaaS Environment Variables
- `SUPABASE_JWT_SECRET`: Secret to verify Supabase JWTs.
- `ENCRYPTION_KEY`: BYOK storage key. (If blank, auth/encryption are disabled).

## Hardware Support
**CUDA note:** If CUDA is unavailable, STT falls back to CPU automatically.

## Health Check
Run `python3 llm_health_check.py` to verify STT and LLM connectivity.
