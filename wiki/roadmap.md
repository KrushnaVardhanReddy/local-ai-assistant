# Development Roadmap

This document outlines the master task list for the Local AI Assistant, broken down into developmental phases.

## Phase 1-18: Core, UX, Gemini, Intelligence & Portable (✅ Complete)
- **Phase 1-3 (The Brain & Ears)**: Local LLM integration, audio capture, and STT pipelines.
- **Phase 4 (UI)**: Svelte 5 + Tauri floating overlay.
- **Phase 5-6 (RAG & Vision)**: Document ingestion, web search, and screen capture (`Ctrl+Shift+S`).
- **Phase 7 (Polish)**: Stealth hotkeys, remote helper mode, and testing.
- **Phase 8-11 (SaaS & UX)**: Supabase/Stripe auth, smart audio filtering (noise/filler), and resume parsing.
- **Phase 12 (UX Polish)**: Syntax highlighting for code, language preference controls, and stealth UI contrast.
- **Phase 13 (Audio Routing)**: Audio loopback (hearing the interviewer's voice).
- **Phase 14**: Transcript Chip bar (clickable priority questions).
- **Phase 15**: Speaker Diarization (separating interviewer vs candidate voices).
- **Phase 16**: Gemini Live Mode integration for sub-second, STT-less reasoning.
- **Phase 17**: Session scorecards and real-time coaching.
- **Phase 18**: Portable App packaging (no installation required).

## Phase 19-25: SaaS, Advanced Features & Competitor Parity (✅ Complete)
- **Phase 19-20**: Pricing tiers, referral engine, and device limit locks.
- **Phase 21-23**: Mock interviews, job context grounding, and E2E testing suite.
- **Phase 24**: Lifetime "Bring Your Own Key" (BYOK) system.
- **Phase 25**: Competitive gaps (STAR framework formatting, rolling transcript buffer, session history trendline, email drafting).

## Phase 26-31: Enterprise & Agentic PLG (🚧 In Progress)
- **Phase 26**: Enterprise B2B Features (SSO, seat management, team RAG).
- **Phase 27**: Voice Conversational Mode (hands-free wake words, TTS).
- **Phase 28**: Agentic OS-Level Control (tool calling, PyAutoGUI, MCP).
- **Phase 29**: Auto-Tailored Resume Builder (dual-pane markdown editor, PDF export) *(✅ Complete)*.
- **Phase 30**: Platform Integrations (Chrome Extension, Slack Bot, Zoom App).
- **Phase 31**: Freemium PLG (paywalling cloud AI, forcing local STT for free tier) *(✅ Complete)*.

## Phase 32-34: Edge Intelligence & Mobile Apps (🚧 In Progress)
- **Phase 32**: SmolLM2 Local Intelligence Layer (on-device turn-taking gating, semantic Q&A cache using ChromaDB).
- **Phase 33**: Native Mobile App via Capacitor (iOS/Android wrapping, background audio plugins).
- **Phase 34**: Enterprise Timeline (infinite SQLite persistent searchable memory, Action Item extraction).
