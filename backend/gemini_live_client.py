"""
GeminiLiveClient — Unified audio → transcription + LLM response via Gemini Live API.

Activated when LLM_PROVIDER=gemini. Replaces the separate STT + LLM pipeline
with a single streaming call. Requires GEMINI_API_KEY.
"""
import asyncio
import sys
import time
import os
from typing import AsyncGenerator


class GeminiLiveClient:
    """
    Streams PCM audio chunks to Gemini Live and yields response tokens.
    Also yields the transcription of the user's speech as it becomes available.

    Usage:
        client = GeminiLiveClient(api_key=..., model=..., system_prompt=...)
        await client.connect()
        await client.send_audio(pcm_bytes)
        async for token in client.receive():
            print(token)
        await client.disconnect()
    """

    def __init__(self, api_key: str, model: str, system_prompt: str, sample_rate: int = 16000):
        self.api_key = api_key
        self.model = model
        self.system_prompt = system_prompt
        self.sample_rate = sample_rate
        self._session = None
        self._client = None
        self._session_ctx = None

    async def connect(self) -> None:
        """Establish a persistent Gemini Live session."""
        try:
            from google import genai
            from google.genai import types
        except ImportError:
            raise RuntimeError(
                "google-genai not installed. Run: pip install google-genai"
            )

        self._client = genai.Client(api_key=self.api_key, http_options={"api_version": "v1beta"})

        config = types.LiveConnectConfig(
            response_modalities=["TEXT"],
            system_instruction=types.Content(
                parts=[types.Part(text=self.system_prompt)],
                role="user"
            ),
            speech_config=types.SpeechConfig(
                voice_config=types.VoiceConfig(
                    prebuilt_voice_config=types.PrebuiltVoiceConfig(voice_name="Puck")
                )
            ),
            input_audio_transcription=types.AudioTranscriptionConfig(),
        )

        self._session_ctx = self._client.aio.live.connect(
            model=self.model, config=config
        )
        self._session = await self._session_ctx.__aenter__()
        print(f"✅ Gemini Live session connected (model: {self.model})", file=sys.stderr)

    async def disconnect(self) -> None:
        """Close the Gemini Live session."""
        if self._session_ctx:
            try:
                await self._session_ctx.__aexit__(None, None, None)
            except Exception:
                pass
        self._session = None
        self._client = None
        self._session_ctx = None

    async def send_audio(self, pcm_bytes: bytes) -> None:
        """Send a PCM audio chunk to the active Gemini session."""
        if not self._session:
            raise RuntimeError("Not connected. Call connect() first.")
        from google.genai import types
        import base64
        audio_b64 = base64.b64encode(pcm_bytes).decode()
        await self._session.send(
            input=types.LiveClientRealtimeInput(
                media_chunks=[types.Blob(data=audio_b64, mime_type=f"audio/pcm;rate={self.sample_rate}")]
            )
        )

    async def end_of_turn(self) -> None:
        """Signal end of user speech turn — triggers Gemini to respond."""
        if self._session:
            await self._session.send(input=".", end_of_turn=True)

    async def receive(self) -> AsyncGenerator[dict, None]:
        """
        Yield response events from Gemini. Each yielded dict has:
          {"type": "transcript", "text": "..."}   — user's speech transcription
          {"type": "token", "text": "..."}         — LLM response token
          {"type": "done"}                         — response complete
        """
        if not self._session:
            return

        async for response in self._session.receive():
            # Input transcription (what the user said)
            if response.server_content and response.server_content.input_transcription:
                text = response.server_content.input_transcription.text
                if text:
                    yield {"type": "transcript", "text": text}

            # Model response tokens
            if response.text:
                yield {"type": "token", "text": response.text}

            # Turn complete
            if response.server_content and response.server_content.turn_complete:
                yield {"type": "done"}

    async def send_text(self, text: str) -> None:
        """Send a text message directly (for click-to-send chip bypass)."""
        if self._session:
            await self._session.send(input=text, end_of_turn=True)

    @property
    def is_connected(self) -> bool:
        return self._session is not None
