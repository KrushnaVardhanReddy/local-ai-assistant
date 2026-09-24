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
- Orchestrating native system audio loopback capture, Voice Activity Detection (VAD), and routing to STT (via Groq or local `whisperfile` Cosmopolitan binary).
- Constructing prompts and routing them to the LLM. Includes injecting a **Conversational Context Window** (rolling turn memory, default 3 turns) into the system prompt to allow the LLM to understand contextual follow-up questions.
- Handling local semantic caching to save on API costs.
- **Rule:** The engine cannot import *any* external libraries or Wails packages. It only communicates through interface definitions located in `core/ports/`.
- **System One Classifier**: A lightweight turn-detection mechanism using Gemma 3 270M running on a local CPU-only `llama-server` sidecar subprocess. This engine decides exactly when an interviewer has finished speaking to optimize expensive Cloud LLM triggers. It uses a **Rolling Question Buffer** that aggregates incoming audio chunks and continuously evaluates context instead of relying on a simplistic silence timer, ensuring complete multi-part questions are processed together.

## 2. Infrastructure Adapters (`wails-app/adapters/`)
Adapters plug into the core engine.
- **LLM Adapter**: Implements `driven.LLMPort`. We support a hybrid approach: local models (via single-file `llamafile` / Ollama) or cloud models (Groq, Gemini, OpenRouter, Cloudflare Workers AI) configurable via `wails-app/backend/config` (`AppConfig`) loaded from `.env.local`.
- **Cache Adapter**: Implements `driven.CachePort`. Uses `sqlite-vec` to store embeddings locally for instant semantic Q&A lookup, as well as holding workspace RAG documents like resumes and job descriptions.
- **Events Adapter**: Implements `driven.EventsPort`. Usually powered by the Wails event bus, streaming updates to the frontend UI.
- **Window Adapter**: Implements `driven.WindowPort`. Uses OS-specific syscalls (like `SetCaptureExcluded`) to make the UI completely invisible to screen sharing.

## 3. The Frontend (Svelte 5 / Wails)
The frontend serves purely as a dumb terminal/display layer for the backend's AI output.
- **Svelte 5**: Provides a reactive, lightweight UI using the new runes reactivity system.
- **Wails v2 (Desktop)**: Provides a pure native desktop application shell using Go, completely replacing older web/Tauri concepts, consuming around 10-30MB of RAM (compared to Electron's 150MB+ footprint).

### Download Progress Synchronization
For heavy models (such as the Gemma turn-detection sidecar and llama-server binary), the backend utilizes an atomic downloader system. The downloader natively supports a progress callback that pushes `on_download_progress` events to the UI via Wails events. Svelte components (like the `InterviewHUD`) intercept these events to display real-time download bars for seamless user feedback during initial local-mode startup.

## Product Skins (`wails-app/products/`)
Because the `StealthEngine` is completely generic, we can create multiple distinct applications that share the same backend. A product (like **StealthPresenter** or **MentorGlass**) simply defines:
1. Which Svelte UI component to load.
2. The specific system prompt to inject into the LLM adapter.
3. The specific setup/teardown logic for that tool.

### Workspace
- **Workspace Tree**: The workspace manages files loaded into the application. Standalone files opened (e.g., resumes, code) are appended to the workspace tree as root nodes, allowing them to be indexed for RAG.

### Authentication & Licensing UI
The frontend authentication system (`src/lib/auth.svelte.ts`) and modal UI (`AuthModal.svelte`) are designed to support two distinct operational modes controlled by the `VITE_PRODUCT` environment variable. To improve UX during login, the authentication window disables the "invisible shield" mode (`AlwaysOnTop: false`) and presents a fully styled, draggable OS-like window using a custom `Titlebar.svelte` component. Once authentication succeeds, the application transitions back into frameless shield mode.

1. **BarnOwl AI (Lifetime + Demo Mode):** For `VITE_PRODUCT=interview`, the UI presents a dual-option modal. Users can either activate a lifetime license key (via Paddle) or start a 15-minute free demo via Google OAuth. The `dev_allowlist` Supabase table enables machine IDs to bypass checks. Entitlements and 15-minute expirations are tracked in the `user_entitlements` table.
2. **SaaS Products:** For other products (e.g., MentorGlass, CounselDesk), the UI strictly presents a "Continue with Google" OAuth sign-in, which tracks Paddle subscription plans and overage tracking via `user_entitlements`.

## Authentication and Licensing (`wails-app/backend/auth/`)
The application supports dual authentication schemes:
1. **Google OAuth**: A local OAuth loopback server (using an HTML trampoline page to parse URL fragments) provides access to the 15-minute free demo and SaaS tiers via Supabase.
2. **Paddle License Validation**: For Lifetime Deals, the app generates a deterministic hardware ID using the OS's native machine UUID, HMAC-hashed for privacy. This ID ensures licenses cannot be shared across physical devices. Paddle acts as the MoR for all billing.

All sensitive tokens and activation secrets are stored natively on the user's OS Keychain using `zalando/go-keyring`. During the 15-minute free demo, the app uses a Supabase Edge Function (`llm-proxy`) to securely proxy LLM API requests and inject our server-side API key without exposing it to the client binary.

## SaaS Entitlements & Paddle Billing
For our SaaS products (MentorGlass, CounselDesk, ClinicHUD), user entitlements (such as `usage_seconds`, `included_seconds`, and `product_mode`) are synchronized from Supabase and tracked globally in the frontend via `authState.userEntitlements`. This powers the live Usage & Billing meter UI in the Settings panel.

We use Paddle as our single unified billing Merchant of Record (MoR). Paddle handles both one-time lifetime deals and monthly SaaS subscriptions. This is supported by two Supabase Edge Functions:
- **`paddle-webhook`**: Receives Paddle webhook events (e.g., `transaction.completed`, `subscription.activated`, `subscription.updated`, `subscription.canceled`) and activates or deactivates entitlements in `user_entitlements`.
- **`paddle-billing-cron`**: A monthly cron job that reads actual usage vs included limits from `user_entitlements` and dynamically charges any overages directly via the Paddle API.

### Multi-Device Referral Engine
Every user gets a unique referral code auto-generated when their entitlements row is created in `user_entitlements`.
When a new user buys the Lifetime License using someone's referral code, the referrer gets their `allowed_devices` count incremented (e.g. from 1 to 2, permanently unlocking a 2nd device).
The new buyer gets a $10 discount via a Paddle coupon. Device limit enforcement is done in the frontend during `syncUserEntitlements()`.

The UI surfaces for the referral engine include:
- A "Refer a Friend" panel in the `Settings.svelte` tab for users to copy their shareable code.
- A "Have a referral code?" input in the `AuthModal.svelte` checkout footer for new users to enter their referrer's code before purchase, passed as custom data to Paddle.

## Supabase Deployment
Supabase production setup is fully automated. The SQL migrations (e.g., creating `dev_allowlist` and `user_entitlements` tables) and Edge Functions (`llm-proxy`) can be automatically deployed using the included bash script.
Run `scripts/deploy_supabase.sh` to link your Supabase project, push all schema migrations, deploy Edge Functions, and set required secrets.

### Phase 70: Mock Interview Mode (Go Port)
- **TTS Adapter**: Implemented `wails-app/backend/tts` with `EdgeTTSAdapter` using `edge-tts` and `ffplay`/`afplay`.
- **Engine State**: `StealthEngine` tracks `isMockMode` and updates `SystemPrompt` dynamically to an interviewer persona when active. Uses `llm.MockInterviewerPrompt` which instructs the LLM to act as a senior technical interviewer, evaluate the response, give brief constructive feedback, and ask a relevant follow-up question.
- **Audio Output**: Final LLM answers are streamed to the TTS adapter when Mock Mode is active.
- **UI Toggle**: Added a Mock Mode toggle to the frontend header in `InterviewHUD.svelte` that issues Wails IPC calls to update backend state.
