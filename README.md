# 🦉 BarnOwl AI — The Stealth Interview & Meeting Co-Pilot

![Python](https://img.shields.io/badge/Python-3.10+-3776AB?style=flat-square&logo=python&logoColor=white)
![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=flat-square&logo=svelte&logoColor=white)
![Tauri](https://img.shields.io/badge/Tauri-2.0-FFC131?style=flat-square&logo=tauri&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)
![SaaS](https://img.shields.io/badge/SaaS%20Platform-Stripe%20%7C%20Supabase-blueviolet?style=flat-square)

**BarnOwl AI** is a real-time, context-aware desktop assistant designed for high-stakes interviews and confidential meetings. 

Like a barn owl—famous for its asymmetrical ears and completely silent flight—this app is the ultimate stealth listener. It captures system audio without exposing itself to screen sharing, transcribes conversations in real-time, and provides live coaching, code solutions, and meeting notes.

---

## ✨ The BarnOwl Advantage

Most AI meeting assistants (like Otter or Fireflies) join your calendar invite as a highly visible "Bot". BarnOwl AI takes a different approach:
- **100% Stealth:** Runs as a native desktop overlay. Uses Windows `WDA_EXCLUDEFROMCAPTURE` to remain completely invisible to Zoom, Teams, and Google Meet screen sharing.
- **No Bots Allowed:** Uses WASAPI system loopback to capture meeting audio directly from your speakers. You never have to ask "Is it okay if I record this?"
- **Bring Your Own Key (BYOK):** Ultimate privacy. Use your own API keys (OpenAI, Anthropic, Gemini, Groq, or OpenRouter). You control your data; we don't store your transcripts.

---

## 🚀 Key Features

- **Real-Time Coaching:** Get live answers and bullet points streamed directly to your screen as the interviewer asks the question.
- **Resume Grounding:** Upload your resume (PDF/TXT) and the Target Job Description. BarnOwl AI auto-injects them into the system prompt so every answer is tailored to your exact background.
- **STAR Method Formatting:** One-click toggle forces the AI to structure behavioral answers using the Situation, Task, Action, Result framework.
- **Post-Meeting Intelligence:** The moment you close the session, BarnOwl generates a scorecard, tracks your trend over time, and drafts a personalized follow-up thank you email.
- **"Catch Me Up":** Zoned out? Hit the history button and BarnOwl instantly generates a 3-bullet summary of what you just missed.

---

## 🏗️ Architecture & Tech Stack

BarnOwl AI is a production-ready SaaS application with a dual-process architecture:

1. **Frontend (Svelte 5 + Tauri 2.0):** 
   - A highly optimized, glassmorphism UI overlay that consumes minimal RAM (~30MB).
   - Global stealth hotkeys for mouse-less interaction.
2. **Backend (Python + FastAPI):** 
   - Captures microphone and speaker loopback audio via `sounddevice`/`pyaudio`.
   - VAD (Voice Activity Detection) filters out noise and filler words.
   - Streams live tokens to the frontend via WebSockets.
3. **Web Dashboard (SvelteKit + Supabase):**
   - Handles JWT authentication, device lock management (anti-sharing), and Stripe billing subscriptions.

---

## ⚡ Quick Start (Developer Setup)

### Prerequisites

- [Python 3.10+](https://www.python.org/downloads/)
- [Node.js v18+](https://nodejs.org/)
- [Rust](https://rustup.rs/) (required by Tauri)

### 1. Clone & Install

```bash
git clone https://github.com/yourusername/Local_AI_Assistant.git
cd Local_AI_Assistant
make install
```

### 2. Configure Environment

Copy `.env.example` to `.env.local` and set up your Supabase credentials (for auth) and LLM API keys. 
For local testing without a cloud provider, you can use Ollama or LM Studio.

### 3. Launch the Application

```bash
make dev-all
```

---

## ⌨️ Advanced Stealth Hotkeys

BarnOwl AI is designed to be operated entirely without a mouse while in stealth mode, ensuring you never look distracted.

| Shortcut | Action | Description |
|---|---|---|
| **`Ctrl+Shift+Space`** | Push-To-Talk | Force the mic to listen (bypasses VAD). |
| **`Ctrl+Shift+X`** | Panic Hide | Clears the transcript and hides the app instantly. |
| **`Ctrl+Shift+S`** | Vision Capture | Takes a silent screenshot of your monitor (great for coding rounds). |
| **`Ctrl+Shift+1-6`** | Send Chip | Sends a specific pending transcript snippet directly to the LLM. |
| **`Ctrl+Shift+E`** | Session Report | Ends the session and generates your scorecard and email draft. |

---

## 💰 SaaS Billing Tiers

BarnOwl AI supports a robust subscription model out of the box (managed via Stripe):

| Tier | Price | Features |
|---|---|---|
| **Demo** | Free | 15 min session (Requires referral link) |
| **Pay-as-you-go** | $5 | 90 min one-time session token |
| **Monthly** | $19 / mo | Max 15 sessions/mo, unlocks Live Coaching |
| **Lifetime (BYOK)** | $99 one-time | Bring Your Own Key, unlimited usage |

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
