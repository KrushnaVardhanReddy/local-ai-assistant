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
| P64-T3 | LemonSqueezy (Manual) | **LemonSqueezy Store Setup (BarnOwl AI only)** — Create the "BarnOwl AI Lifetime" product. Configure three pricing tiers: Founder $49 (max 1,000 sales), Early $79 (max 5,000 sales), Standard $99 (unlimited). Enable "Generate License Keys" on purchase. | ⬜ | — |
| P64-T4 | Stripe Dashboard (Manual) | **Stripe Metered Billing Setup (SaaS Products)** — Create Stripe products for MentorGlass ($29/mo base), CounselDesk ($99/mo base), ClinicHUD ($49/mo base). Each product has TWO price items: (1) a flat recurring fee and (2) a metered price for overage usage (per minute, charged at month end). Configure Stripe Customer Portal so users can self-manage subscriptions and view billing history without contacting support. | ⬜ | — |
| P64-T5 | `wails-app/frontend/src/lib/auth.svelte.ts`, `wails-app/frontend/tests/entitlement.spec.ts` | **Entitlement E2E Tests** — Expose `authState` to Playwright via `window.__authState` and write UI E2E tests for Developer Mode, Active License, Usage Meter calculations, Overage warnings, and Stripe subscription status. | 🔄 | — |

---
