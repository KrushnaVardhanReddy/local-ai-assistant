# 🏗️ BarnOwl AI v2 — Jules Task Tracker (Hexagonal Engine)

> Submit tasks to Jules one-by-one or in parallel batches:
> ```bash
> python3 scripts/jules_submit.py --task P54-T1
> python3 scripts/jules_submit.py --list   # see all tasks
> ```
> Update status and PR numbers here as Jules returns PRs.

## Legend
| Symbol | Meaning |
|--------|---------|
| ✅ | Done / PR merged |
| 🔄 | Jules working on it |
| ⬜ | Not started |
| ⚡ | Can run in PARALLEL with other ⚡ tasks in same batch |

---

## Phase 62 — OAuth Authentication & Device Licensing 🔐

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|  
| P62-T1 ⚡ | `wails-app/app.go`, `wails-app/backend/auth/` | **Go OAuth Loopback Server & Machine ID** — Implement local HTTP loopback server for OAuth redirect capture. Implement `GetMachineId()` using OS hardware UUID. Implement `SaveToken`, `LoadToken`, `DeleteToken` via OS keychain (`zalando/go-keyring`). Add `StartOAuthFlow(provider string) error`. | ✅ | — |
| P62-T2 ⚡ | `wails-app/frontend/src/lib/auth.svelte.ts`, `wails-app/frontend/src/lib/components/AuthModal.svelte` | **Frontend OAuth UI & Device Registration** — Replace email/password form with "Sign in with Google" button. Listen for `on_auth_complete` Wails event. Register hashed machine ID in Supabase `device_registrations` table. Fetch and cache user entitlement profile. | ✅ | — |
| P62-T3 | `wails-app/frontend/src/App.svelte`, `wails-app/frontend/src/lib/Settings.svelte` | **License Gate & Entitlement UI** — Show full-screen lock banner when `byok_pass_active == false && remaining_sessions <= 0`. Display license status and "Recharge Pass" button in Settings. | ✅ | — |

---

## Phase 63 — SaaS Backend Proxy (Demo Mode) 🛡️

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P63-T1 | `supabase/functions/llm-proxy/`, `wails-app/backend/llm/openai.go` | **Edge Function LLM Proxy** — Create a Supabase Deno Edge Function to proxy OpenAI requests. Verify JWT and `demo_expires_at`. Inject `OPENAI_API_KEY`. For SaaS product users (non-demo), report session duration (seconds) to Stripe Metered Billing via `POST /v1/subscription_items/{id}/usage_records`. Increment `usage_seconds` in `user_entitlements` table for in-app usage meter. | ✅ | — |
| P63-T2 | `wails-app/frontend/src/lib/Settings.svelte` | **Usage Meter UI (SaaS Products)** — Add a live usage meter to the Settings panel for SaaS products (MentorGlass, CounselDesk, ClinicHUD). Show hours used vs included allocation, estimated overage cost this month, and next billing date. Read from `user_entitlements.usage_seconds`. Only visible when `productMode !== 'interview'`. | ✅ | — |

---

## Phase 64 — Launch Prep & Distribution 🚀

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P64-T1 | `wails-app/wails.json`, `Makefile` | **App Packaging & Build** — Configure Wails to produce production-ready binaries: `.dmg` (macOS), `.exe` (Windows), and `.AppImage` (Linux). Add app icons, metadata, and cross-compilation scripts to the Makefile. | ✅ | — |
| P64-T2 | `scripts/deploy_supabase.sh` | **Supabase Production Setup Scripts** — Create SQL migration files for the required tables (`dev_allowlist`, `user_entitlements`) and a deployment script (`scripts/deploy_supabase.sh`) for the Edge Function. | ✅ | — |
| P64-T3 | `supabase/functions/paddle-webhook/index.ts` | **Paddle Billing Webhook Setup** — Create a Supabase Edge Function to process Paddle webhook events (`transaction.completed`, `subscription.activated`, etc). Verify the webhook signature using `PADDLE_WEBHOOK_SECRET`. Update the `user_entitlements` table with `paddle_subscription_id` and `plan_type` based on the payload. | ✅ | — |
| P64-T4 | Paddle Dashboard (Manual) | **Paddle Sandbox Setup** — Create the "BarnOwl AI Lifetime" product and SaaS products. Configure the webhook destination. | ✅ | — |
| P64-T5 | `wails-app/frontend/src/lib/auth.svelte.ts`, `wails-app/frontend/tests/entitlement.spec.ts` | **Entitlement E2E Tests** — Expose `authState` to Playwright via `window.__authState` and write UI E2E tests for Developer Mode, Active License, Usage Meter calculations, Overage warnings, and Stripe subscription status. | 🔄 | — |
| P64-T7 ⚡ | `wails-app/frontend/index.html`, `wails-app/frontend/src/lib/components/AuthModal.svelte` | **Frontend Paddle Checkout Integration** — Embed Paddle.js directly in the frontend HTML. Update AuthModal's "Buy Lifetime License" button to invoke `Paddle.Checkout.open()` and inject `customData: { user_id: authState.user.id }` so that webhook transactions can be tied to the correct user. | ✅ | — |

---
| P64-T6 | `wails-app/main.go`, `wails-app/frontend/src/lib/components/AuthModal.svelte` | **Auth Window UX Redesign** — Modify Wails app to start with `AlwaysOnTop: false` and `BackgroundColour` set to a solid color when in Auth mode. Add a custom draggable HTML titlebar to `AuthModal.svelte` with minimize, maximize, and close buttons that call Wails runtime functions. Toggle `AlwaysOnTop` and background color dynamically when authentication succeeds. | ✅ | #228 |

---

## Phase 65 — Platform Architecture 🏗️

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|-----|
| P65-T1 | `wails-app/frontend/src/lib/Assistant.svelte`, `InterviewHUD.svelte`, `ConvPanel.svelte`, `CopilotDrawer.svelte`, `StatusBar.svelte`, `Settings.svelte` | **Stealth Mode Feature Flag (Frontend)** — Conditionally hide stealth-related UI buttons/controls based on `VITE_STEALTH_MODE` env var. When false, hide Stealth button, clickthrough toggle, and stealth status indicator. Go backend already done. | ✅ | #230 |

---

## Phase 66 — Multi-Device Referral Engine 🔗

> Strategy: Lock the 2nd device slot behind 1 successful referral. The referrer gets their `allowed_devices` bumped from 1 → 2. The new buyer gets $10 off their Lifetime License via a Paddle discount coupon. No cash payouts — zero cost to fulfill.

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|-----|
| P66-T1 | `supabase/migrations/20261000000000_add_referral_system.sql` | **DB Migration** — Add `referral_code` (auto-generated `BARN-XXXX` via trigger), `allowed_devices` (default 1), `referred_by_code`, `referral_rewarded_at` columns to `user_entitlements`. | ✅ | #231 |
| P66-T2 ⚡ | `supabase/functions/paddle-webhook/index.ts` | **Webhook Reward Logic** — In `transaction.completed` handler, read `custom_data.referred_by`, look up the referrer, set `allowed_devices = 2` and stamp `referral_rewarded_at` (idempotent — never double-reward). | ✅ | #233 |
| P66-T3 ⚡ | `auth.svelte.ts`, `AuthModal.svelte`, `Settings.svelte` | **Frontend UI** — (A) Add `referralCode`, `allowedDevices`, `deviceLimitReached` to `authState`. (B) Add referral code input above "Buy a Lifetime License" in AuthModal — passes `referred_by` + `discountId` to Paddle. (C) Add "Refer a Friend" card in Settings showing the user's own code with a Copy button and device slot status. | ✅ | #232 |

## Phase 67 — Gemma "System One" Turn-Detection Engine 🧠

> Strategy: Replace the current silence-timer LLM trigger with an intelligent rolling buffer powered by Gemma 3 270M (GGUF) running locally on the CPU via a llama-server subprocess. Model is downloaded at first launch (~500MB) — not bundled. This enables accurate end-of-turn detection even when interviewers pause mid-question, stutter, or ask multi-part questions with breaks.

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|-----|
| P67-T1 | `wails-app/backend/classifier/gemma.go`, `llama_server.go` | **llama-server Sidecar + Gemma Download** — Runtime-download Gemma 3 270M GGUF and a pre-compiled llama-server binary (per OS/arch) using `os.UserCacheDir()`. Start llama-server as a managed subprocess on `localhost:18080`. Provide `IsQuestionComplete(ctx, text) (bool, error)` via the OpenAI-compat `/v1/chat/completions` endpoint with a BNF grammar that forces `{"result":true/false}` output. | ✅ | #234 |
| P67-T2 | `wails-app/backend/classifier/buffer.go` | **Rolling Question Buffer** — Implement `QuestionBuffer` that accumulates transcript chunks and calls `IsQuestionComplete()` after each chunk. Flushes the aggregated question text to the LLM callback only when Gemma returns true. Hard 45-second failsafe watchdog ensures the buffer always flushes eventually even if Gemma is unavailable. | ✅ | #236 |
| P67-T3 | `wails-app/backend/lib/engine/engine.go`, `manager.go` | **Engine Wiring** — Remove the existing silence/timer-based LLM trigger. Wire `QuestionBuffer.AddChunk()` into the transcript pipeline. Start/stop the `DefaultLlamaServer` alongside the engine lifecycle. Preserve ManualMode guard and ClearState buffer reset. Graceful degradation: if llama-server fails to start, the 45s watchdog keeps the app functional. | ✅ | #237 |
| P67-T4 | `wails-app/backend/llm/prompts.go` | **Conversational Context Window** — Add `BuildContextBlock(turns []Turn) string` and `BuildFullSystemPrompt(category string, turns []Turn) string` to the prompt builder. Injects the last 3 completed turns (Q + AI answer) as a compact, token-limited context block into every LLM system prompt. Fixes follow-up questions like "Why?" or "Explain that" that reference previous turns. Independent of T1/T2/T3 — can run in parallel. | ✅ | #235 |
| P67-T5 | `app.go`, `engine.go`, `llama_server.go`, `ws.svelte.ts`, `ConvPanel.svelte` | **LIVE Status Sync & Download Progress UI** — Expose `is_listening` from the engine to Svelte via polling. Pass an `EventPort` down to `llama_server.go` to emit `on_download_progress` from `system.DownloadFileAtomic` during Gemma/Llama download. Listen to this event in the frontend to display a progress indicator while the models download. | ⬜ | — |

## Phase 68 — Environment Configuration Refactor ⚙️

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|-----|
| P68-T1 | `wails-app/backend/config/config.go`, `main.go`, `openai.go` | **Environment Configuration Refactoring** — Introduce `caarlos0/env` to parse environment variables into a strongly-typed `AppConfig` struct. Implement a fail-fast Must pattern on startup. Replace scattered `os.Getenv` and `getEnvOrDefault` calls across the codebase, injecting `AppConfig` instead. | ✅ | — |
