# Feature Comparison: Stealth Interview AI Assistants

This matrix compares the **Local AI Assistant** against existing open-source and commercial competitors in the real-time interview copilot space.

## Feature Matrix

| Feature | Local AI Assistant (Ours) | Final Round AI | Ecoute | Generic Copilots (Electron) |
| :--- | :--- | :--- | :--- | :--- |
| **Pricing** | Free / Open Source | $30 - $100+/mo | Free / API Costs | Variable ($10-$50/mo) |
| **STT Processing** | Local (Faster-Whisper) | Cloud (Deepgram/etc.) | Local (Whisper) | Cloud |
| **LLM Processing** | Local (Ollama/LM Studio) | Cloud (GPT-4) | Cloud (OpenAI API) | Cloud (GPT-4/Claude) |
| **End-to-End Latency** | **Zero / Sub-second** | 3-5+ Seconds | 2-3 Seconds | 3-5+ Seconds |
| **Data Privacy** | **100% Offline (No data leaves machine)** | Highly Invasive | High Risk (API) | High Risk |
| **Stealth UI (OS-Level)** | **Yes (Tauri Transparent Native Window)** | Browser-based (Detectable) | Basic Python UI (Visible) | Varies (Usually visible) |
| **Resource Usage** | Lightweight (Tauri / Rust) | Heavy (Chrome Extension) | Moderate (Python) | Very Heavy (Electron) |
| **Resume Extraction** | **Yes (Local LLM One-Shot Pass)** | Yes (Cloud) | No | Rarely |
| **Offline Knowledge Base (RAG)** | **Yes (Local Embeddings & Vector DB)** | No | No | No |
| **Smart Audio Filter (VAD)** | **Yes (Filters noise & partial sentences)** | Basic | Yes | Basic |

---

## Why These Features Matter for Live Interviews

1. **End-to-End Latency:** 
   During a technical interview, every second counts. If the interviewer asks a question and the AI takes 5 seconds to bounce audio to the cloud, transcribe it, send the text to GPT-4, and stream the response back, the resulting silence is awkward and suspicious. By keeping STT and the LLM completely local, our tool achieves **sub-second latency**, streaming hints before the interviewer even finishes speaking.

2. **Data Privacy (100% Offline):**
   Commercial tools require you to stream your screen and microphone to a third-party server. In many corporate environments, recording an interview under NDA and sending it to a cloud startup is a massive liability. **Local AI Assistant never sends a single byte over the internet.**

3. **Stealth UI (Tauri vs Electron):**
   Browser-based tools and heavy Electron apps are easily detected by standard proctoring software. Our assistant uses a Tauri-based native transparent window that hooks directly into the OS, allowing it to bypass screen-sharing visibility on most setups.

4. **Context-Awareness (Resume & RAG):**
   Generic copilots answer questions like a textbook. Thanks to our new **Resume Extraction** feature and **Local RAG**, this assistant instantly grounds its answers in your personal experience level, tech stack, and personal notes without requiring you to manually write long prompts.
