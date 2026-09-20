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
| P62-T1 ⚡ | `wails-app/app.go`, `wails-app/backend/auth/` | **Go OAuth Loopback Server & Machine ID** — Implement local HTTP loopback server for OAuth redirect capture. Implement `GetMachineId()` using OS hardware UUID. Implement `SaveToken`, `LoadToken`, `DeleteToken` via OS keychain (`zalando/go-keyring`). Add `StartOAuthFlow(provider string) error`. | 🔄 | — |
| P62-T2 ⚡ | `wails-app/frontend/src/lib/auth.svelte.ts`, `wails-app/frontend/src/lib/components/AuthModal.svelte` | **Frontend OAuth UI & Device Registration** — Replace email/password form with "Sign in with Google" button. Listen for `on_auth_complete` Wails event. Register hashed machine ID in Supabase `device_registrations` table. Fetch and cache user entitlement profile. | 🔄 | — |
| P62-T3 | `wails-app/frontend/src/App.svelte`, `wails-app/frontend/src/lib/Settings.svelte` | **License Gate & Entitlement UI** — Show full-screen lock banner when `byok_pass_active == false && remaining_sessions <= 0`. Display license status and "Recharge Pass" button in Settings. | ⬜ | — |

---

## Phase 63 — SaaS Backend Proxy (Demo Mode) 🛡️

| Task | Files | Description | Status | PR |
|------|-------|-------------|--------|----|
| P63-T1 | `supabase/functions/llm-proxy/`, `wails-app/backend/llm/openai.go` | **Edge Function LLM Proxy** — Create a Supabase Deno Edge Function to proxy OpenAI requests. Secure it using the user's JWT and verify `demo_expires_at > now()`. Inject our secure `OPENAI_API_KEY` into the proxied request. Update Go LLM backend to route requests to the Edge proxy instead of `api.openai.com` when the frontend passes the `demo` license status and OAuth token. | ⬜ | — |

---
