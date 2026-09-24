# Reading Cooldown

## Overview
BarnOwl AI implements a dynamic "Reading Cooldown" to prevent the Live Mode auto-submit feature from instantly overwriting the screen after the LLM generates a response.

## Logic
1. **Cooldown Calculation**: Once an LLM stream finishes, we count the number of words. The reading cooldown is `wordCount * 300ms`, clamped to a minimum of 3 seconds and a maximum of 20 seconds.
2. **Buffer Behavior**: While the cooldown is active, STT transcript chunks continue to accumulate in the `QuestionBuffer` but are intentionally not passed to the Turn-Detection Classifier (Gemma). This prevents auto-flushes.
3. **Manual Bypass**: Manual submissions (like clicking "Submit Answer" or manually typing a follow-up) will bypass and clear the cooldown immediately.
