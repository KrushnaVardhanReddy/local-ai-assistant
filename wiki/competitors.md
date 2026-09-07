# Competitor Analysis

This document analyzes the commercial competitors in the real-time AI interview copilot space, contrasting them with the **Local AI Assistant (Project Parakeet)**.

## Competitor Landscape
The two primary commercial competitors are **Final Round AI** ($30-$100+/mo) and **Sensei Copilot**. There is also an open-source alternative, **Ecoute**.

### Why Project Parakeet Wins
Project Parakeet is designed to overcome the critical flaws of these existing solutions:

1. **Undetectable Stealth**: Competitors rely on Chrome extensions or desktop apps that are easily flagged by proctoring software. Project Parakeet uses a zero-install portable ZIP, advanced hotkeys, and a Ghost Cursor (via Remote Helper mode), rendering it completely undetectable.
2. **Ultra-Low Latency**: Traditional STT pipelines introduce a 3-5 second delay. Project Parakeet leverages Gemini Live's streaming API for zero/sub-second latency.
3. **Speaker Diarization**: Unlike tools that process a flat stream of text, Parakeet separates candidate and interviewer audio. This enables the LLM to understand conversational context and provide real-time coaching.
4. **Invisible Screen Extraction**: Using `Ctrl+Shift+S`, the Vision Copilot instantly extracts LeetCode or System Design problems from the screen without any visible UI overlay.

## Future Gap Analysis (Roadmap)
Based on features that drive sales for competitors, the following features have been added to the Parakeet roadmap:
- **Mock Interview Mode**: Simulates interviews via TTS, evaluating the user's spoken answers and generating a scorecard.
- **Multilingual Support**: Allowing the STT/LLM pipeline to be forced into specific languages (e.g., Spanish, Hindi, Mandarin) for international interviews.
