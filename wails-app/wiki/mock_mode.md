# Mock Interview Mode

## Overview

"Mock Interview Mode" is a specialized mode within BarnOwl AI designed for users to practice their answers to interview questions. In this mode, the system asks questions and then evaluates the candidate's answers.

## Auto-submit Bypass

During a Live or Stealth session, the Voice Activity Detection (VAD) automatically triggers a submission to the LLM (evaluator) after a period of silence. This ensures seamless conversation flow for an assistant.

However, during a Mock Interview, candidates need time to think and formulate their responses. Therefore, the system bypasses this auto-submit functionality when `isMockMode` is enabled.

Instead, the candidate's speech is continuously accumulated into the `questionBuffer`.

## Manual Submission

To trigger the evaluation of the accumulated response, the user must explicitly submit their answer.

-   **Backend:** `app.go` exposes a `FlushQuestionBuffer` method. This method takes all the accumulated speech chunks from the buffer, combines them, clears the buffer, and sends the full text to the LLM via `AskQuestion`.
-   **Frontend:** In `ConvPanel.svelte`, a "Submit Answer" button is conditionally rendered when `isMockMode` is true. Clicking this button triggers the `FlushQuestionBuffer` method.
