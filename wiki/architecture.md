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

### Workspace
- **Workspace Tree**: The workspace manages files loaded into the application. Standalone files opened (e.g., resumes, code) are appended to the workspace tree as root nodes, allowing them to be indexed for RAG.

### Authentication & Licensing UI
The frontend authentication system (`src/lib/auth.svelte.ts`) and modal UI (`AuthModal.svelte`) are designed to support two distinct operational modes controlled by the `VITE_PRODUCT` environment variable:
1. **BarnOwl AI (Lifetime + Demo Mode):** For `VITE_PRODUCT=interview`, the UI presents a dual-option modal. Users can either activate a lifetime license key (via LemonSqueezy) or start a 15-minute free demo via Google OAuth. The `dev_allowlist` Supabase table enables machine IDs to bypass checks. Entitlements and 15-minute expirations are tracked in the `user_entitlements` table.
2. **SaaS Products:** For other products (e.g., MentorGlass, CounselDesk), the UI strictly presents a "Continue with Google" OAuth sign-in, which tracks standard Stripe subscription plans via `user_entitlements`.

## Authentication and Licensing (`wails-app/backend/auth/`)
The application supports dual authentication schemes:
1. **Google OAuth**: A local OAuth loopback server (using an HTML trampoline page to parse URL fragments) provides access to the 15-minute free demo and SaaS tiers via Supabase.
2. **LemonSqueezy License Validation**: For Lifetime Deals, the app generates a deterministic hardware ID using the OS's native machine UUID, HMAC-hashed for privacy. This ID ensures licenses cannot be shared across physical devices.

All sensitive tokens and activation secrets are stored natively on the user's OS Keychain using `zalando/go-keyring`. During the 15-minute free demo, the app uses a Supabase Edge Function (`llm-proxy`) to securely proxy LLM API requests and inject our server-side API key without exposing it to the client binary.

## SaaS Entitlements
For our SaaS products (MentorGlass, CounselDesk, ClinicHUD), user entitlements (such as `usage_seconds` and `plan_type`) are synchronized from Supabase and tracked globally in the frontend via `authState.userEntitlements`. This powers the live Usage & Billing meter UI in the Settings panel.

## Supabase Deployment
Supabase production setup is fully automated. The SQL migrations (e.g., creating `dev_allowlist` and `user_entitlements` tables) and Edge Functions (`llm-proxy`) can be automatically deployed using the included bash script.
Run `scripts/deploy_supabase.sh` to link your Supabase project, push all schema migrations, deploy Edge Functions, and set required secrets.
