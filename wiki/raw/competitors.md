# Feature Comparison: Stealth Interview AI Assistants

This matrix compares the **Local AI Assistant** (Project Parakeet) against existing commercial competitors in the real-time interview copilot space.

## Feature Matrix

| Feature | Local AI Assistant (Ours) | Final Round AI | Sensei Copilot | Ecoute (Open Source) | Otter.ai / Enterprise |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Pricing** | **Freemium + BYOK + Pro Subscription** | $30 - $100+/mo | Subscription | Free / API Costs | $10 - $20+/mo |
| **STT Processing** | Local (`faster-whisper` / `parakeet`) | Cloud | Cloud | Local (Whisper) | Cloud |
| **Turn Detection & Caching** | **Yes (Local SmolLM2 Engine)** | No | No | No | No |
| **Enterprise Knowledge Base** | **Yes (Shared RAG Playbooks)** | No | No | No | Basic Search |
| **Native Mobile App** | **Yes (iOS/Android via Capacitor)** | No | No | No | Yes |
| **Export Options** | **Universal (MD, Email, Copy)** | Web/PDF | Web | Copy | Web/PDF/Email |
| **End-to-End Latency** | **Zero / Sub-second** | 3-5+ Seconds | 1-2 Seconds | 2-3 Seconds | 2-3 Seconds |
| **Stealth Mode** | **Portable ZIP + Ghost Cursor** | Browser/App | Chrome Extension | Basic Python UI | None |
| **Speaker Diarization** | **Yes (Interviewer vs Candidate)** | No / Basic | No | No | Yes |
| **Coding/LeetCode Extraction** | **Yes (Vision Copilot)** | Yes | Yes | No | No |
| **Resume Extraction** | Yes (Local LLM Pass / Context) | Yes | Yes | No | No |
| **Mock Interview Mode** | ❌ Missing | Yes | Yes | No | No |

---

## Why Our Architecture Wins

1. **Undetectable Stealth (Portable App + Ghost Cursor):**
   Competitors rely on Chrome extensions or heavy desktop applications which are easily flagged by proctoring software. Our assistant uses a zero-install **Portable ZIP** combined with advanced hotkeys and a **Ghost Cursor** (via Remote Helper). The candidate never touches their mouse, making detection practically impossible via screen-share or OS monitoring.

2. **Ultra-Low Latency (Local STT + SmolLM2 Edge Caching):** 
   In an interview, a 5-second delay is fatal. By running STT locally on-device (`faster-whisper` or `parakeet`) and using our `SmolLM2` Gatekeeper to filter out partial sentences, we drastically cut down on wasted cloud calls. Furthermore, if a question hits our local ChromaDB semantic cache, the answer is served instantly (zero latency), completely skipping the cloud LLM bottleneck.

3. **Speaker Diarization & Routing:**
   Unlike competitors that just read a wall of text, our app splits the audio channels to tag `[INTERVIEWER]` and `[CANDIDATE]`, ensuring the LLM understands the flow of the conversation and can even provide real-time coaching on the candidate's answers.

4. **Invisible Screen Extraction (Vision Copilot):**
   By hitting `Ctrl+Shift+S`, the app silently captures the screen in memory and pipes it to a Vision model to extract LeetCode problems or System Design diagrams instantly, without any UI overlay blocking your screen.

---

## 🚀 Potential Roadmap Additions (Gap Analysis)

Based on checking what features drive sales for Final Round AI and Sensei Copilot, here are the major gaps we should consider for **Phase 21+**:

1. **Mock Interview Mode**
   * **The Gap:** Users want to practice with the tool before risking it on a real interview.
   * **The Fix:** Add a mode where the LLM speaks questions aloud (via TTS) and evaluates the user's spoken answers in a simulated environment, ending with the Session Scorecard.

2. **Explicit Multilingual Support**
   * **The Gap:** Sensei advertises 30+ languages as a core feature.
   * **The Fix:** Add a dropdown in our Settings UI to force the STT/LLM pipeline into specific languages (Spanish, Hindi, Mandarin) for international interviews.
