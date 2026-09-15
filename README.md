# 🦉 BarnOwl AI — The Stealth Interview & Meeting Co-Pilot

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white)
![Svelte](https://img.shields.io/badge/Svelte-5-FF3E00?style=flat-square&logo=svelte&logoColor=white)
![Wails](https://img.shields.io/badge/Wails-2.15-ED2737?style=flat-square&logo=wails&logoColor=white)
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

BarnOwl AI is a production-ready SaaS application compiled into a unified single-binary architecture:

1. **Frontend (Svelte 5 + Vite):** 
   - A highly optimized, glassmorphism UI overlay that consumes minimal RAM (~30MB).
   - Global stealth hotkeys for mouse-less interaction.
2. **Backend (Go + Wails v2):** 
   - Captures microphone audio using native CGo bindings (`go-audio`).
   - Integrates local ML embeddings (ONNX) and local STT (`whisper.cpp`).
   - Streams live tokens and handles OS-level window management natively.
3. **Web Dashboard (SvelteKit + Supabase):**
   - Handles JWT authentication, device lock management (anti-sharing), and Stripe billing subscriptions.

---

## ⚡ Quick Start (Developer Setup)

### Prerequisites

- [Go 1.25+](https://go.dev/)
- [Node.js v18+](https://nodejs.org/)
- [Wails CLI v2+](https://wails.io/)

### 1. Clone & Install

```bash
git clone https://github.com/yourusername/BarnOwl_AI.git
cd BarnOwl_AI
make install
```

### 2. Configure Environment

Copy `.env.example` to `.env.local` and set up your Supabase credentials (for auth) and LLM API keys. 
For local testing without a cloud provider, you can use Ollama or LM Studio.

### 3. Launch the Application

```bash
make dev
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

| Tier | Price | Access Period | Key Model & Features |
|---|---|---|---|
| **Demo Trial** | Free | 15 Min Live Demo | Managed Key (Instant trial out of the box) |
| **30-Day BYOK Pass** | $10 (Prepaid) | 30 Days | Bring Your Own Key, unlimited usage |
| **Annual BYOK Pass** | $50 (Prepaid) | 365 Days | Bring Your Own Key, unlimited usage |
| **SaaS Managed Pass** | $19 (Prepaid) | 8 Full Sessions | Fully Managed LLM + STT keys, zero API setup |

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
