import asyncio
import base64
import os
import tempfile
import sys
import edge_tts
from typing import AsyncGenerator, Tuple

from config import config
from llm_client import LLMClient
from session_manager import session as interview_session

class MockInterviewer:
    def __init__(self, llm_client: LLMClient):
        self.llm_client = llm_client
        self.voice = "en-US-AriaNeural"  # Default edge-tts voice

    async def generate_audio(self, text: str) -> str:
        """
        Converts text to speech using edge-tts, saves to a temp file,
        and returns the Base64 encoded MP3.
        """
        if not text:
            return ""

        communicate = edge_tts.Communicate(text, self.voice)
        fd, path = tempfile.mkstemp(suffix=".mp3")
        os.close(fd)

        try:
            await communicate.save(path)
            with open(path, "rb") as f:
                audio_data = f.read()
            return base64.b64encode(audio_data).decode("utf-8")
        except Exception as e:
            print(f"[MockInterviewer] Audio generation failed: {e}", file=sys.stderr)
            return ""
        finally:
            if os.path.exists(path):
                os.remove(path)

    async def get_next_question(self, user_answer: str = "", resume_context: str = "") -> AsyncGenerator[str, None]:
        """
        Streams the LLM's next question based on the user's answer (if any) and interview state.
        """
        system_prompt = (
            "You are an expert technical interviewer conducting a mock interview. "
            "Ask ONE concise question at a time. Do not provide the answer. "
            "If the user just answered, use standard technical interview scorecard logic (correctness, completeness, communication) to evaluate their answer before asking the next question. "
            "Keep your responses spoken-friendly and concise."
        )

        if resume_context:
            system_prompt += f"\n\nCandidate's resume context:\n{resume_context}"

        # Build message history using session
        messages = [{"role": "system", "content": system_prompt}]
        for turn in interview_session.turns:
            messages.append({"role": "user", "content": turn.transcript})
            messages.append({"role": "assistant", "content": turn.response})

        if user_answer:
            messages.append({"role": "user", "content": user_answer})

        if not messages and not user_answer:
            # Very first question
            messages.append({"role": "user", "content": "Hi, I'm ready to start the mock interview."})

        async for token in self.llm_client.stream(messages):
            yield token
