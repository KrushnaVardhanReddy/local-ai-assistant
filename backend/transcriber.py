import io
import os
import sys
import time
import wave
import numpy as np
from config import config

class Transcriber:
    """
    Dual-backend transcriber.

    STT_PROVIDER=local  → faster-whisper running on CPU/GPU (offline)
    STT_PROVIDER=groq   → Groq Whisper API (cloud, ~200ms latency)
    """

    def __init__(self, model_size: str, device: str, compute_type: str, provider: str = "local", diarize: bool = False):
        self.model_size = model_size
        self.device = device
        self.compute_type = compute_type
        self.provider = provider.lower()
        self.model = None  # only used for local provider
        self.diarize = diarize

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def load(self) -> None:
        if self.provider == "groq":
            self._load_groq()
        else:
            self._load_local()

    def transcribe_with_speaker(self, audio_bytes: bytes, channels: int = 1) -> list[dict]:
        """
        Returns list of dicts: [{"speaker": "INTERVIEWER"|"CANDIDATE", "text": "..."}]
        When diarize=False or channels=1 without pyannote, returns single segment with speaker=None.
        """
        if not self.diarize:
            text = self.transcribe(audio_bytes)
            return [{"speaker": None, "text": text}] if text else []

        # STEREO MODE: split L/R channels
        if channels == 2:
            return self._transcribe_stereo(audio_bytes)

        # MONO MODE: use pyannote if available
        return self._transcribe_diarized_mono(audio_bytes)

    def _transcribe_stereo(self, audio_bytes: bytes) -> list[dict]:
        """Split stereo PCM into L (interviewer) and R (candidate) channels."""
        import numpy as np
        if len(audio_bytes) % 4 != 0:
            audio_bytes += b'\x00' * (4 - len(audio_bytes) % 4)

        # Stereo 16-bit PCM: interleaved L, R, L, R ...
        audio_np = np.frombuffer(audio_bytes, dtype=np.int16)
        left = audio_np[0::2].astype(np.float32) / 32768.0   # Interviewer
        right = audio_np[1::2].astype(np.float32) / 32768.0  # Candidate

        results = []
        for channel_audio, speaker in [(left, "INTERVIEWER"), (right, "CANDIDATE")]:
            rms = np.sqrt(np.mean(channel_audio**2))
            if rms < 0.01:  # silence gate
                continue
            # Re-encode to bytes for transcription
            channel_bytes = (channel_audio * 32768).astype(np.int16).tobytes()
            text = self.transcribe(channel_bytes)
            if text:
                results.append({"speaker": speaker, "text": text})
        return results

    def _transcribe_diarized_mono(self, audio_bytes: bytes) -> list[dict]:
        """
        Mono fallback: try pyannote diarization, else return undiarized.
        Requires HF_TOKEN env var for pyannote model download.
        """
        try:
            # Attempt pyannote (optional dep)
            from pyannote.audio import Pipeline
            import os
            hf_token = os.environ.get("HF_TOKEN", "")
            if not hf_token:
                raise ImportError("HF_TOKEN not set")
            # pyannote integration is complex — for now fall back gracefully
            raise ImportError("pyannote mono diarization not yet implemented")
        except (ImportError, Exception):
            # Fallback: undiarized mono
            text = self.transcribe(audio_bytes)
            return [{"speaker": None, "text": text}] if text else []

    def transcribe(self, audio_bytes: bytes) -> str:
        if not audio_bytes:
            return ""

        # Silence gate — skip sending silent audio to any backend
        if len(audio_bytes) % 2 != 0:
            audio_bytes += b'\x00'
        audio_np = np.frombuffer(audio_bytes, dtype=np.int16).astype(np.float32) / 32768.0
        if len(audio_np) > 0 and np.sqrt(np.mean(audio_np**2)) < 0.01:
            return ""

        if self.provider == "groq":
            return self._transcribe_groq(audio_bytes, audio_np)
        else:
            return self._transcribe_local(audio_np)

    # ------------------------------------------------------------------
    # Local (faster-whisper) backend
    # ------------------------------------------------------------------

    def _load_local(self) -> None:
        from faster_whisper import WhisperModel
        import multiprocessing
        cpu_threads = min(multiprocessing.cpu_count(), 8)

        start = time.perf_counter()
        try:
            self.model = WhisperModel(
                self.model_size,
                device=self.device,
                compute_type=self.compute_type,
                cpu_threads=cpu_threads
            )
        except Exception as e:
            if "cuda" in str(e).lower():
                print("⚠️  CUDA unavailable — falling back to CPU", file=sys.stderr)
                self.device = "cpu"
                self.compute_type = "int8"
                self.model = WhisperModel(
                    self.model_size,
                    device=self.device,
                    compute_type=self.compute_type,
                    cpu_threads=cpu_threads
                )
            else:
                raise

        elapsed = time.perf_counter() - start
        print(f"✅ STT model loaded in {elapsed:.2f}s on {self.device} ({cpu_threads} threads)", file=sys.stderr)

    def _transcribe_local(self, audio_np: np.ndarray) -> str:
        if not self.model:
            raise RuntimeError("Model not loaded. Call load() first.")

        start = time.perf_counter()
        segments, _ = self.model.transcribe(
            audio_np,
            beam_size=1,
            language="en",
            condition_on_previous_text=False
        )
        text = " ".join(seg.text for seg in segments).strip()
        if not text:
            return ""

        elapsed = time.perf_counter() - start
        print(f"STT: {elapsed*1000:.0f}ms → '{text[:60]}'", file=sys.stderr)
        return text

    # ------------------------------------------------------------------
    # Groq Whisper API backend
    # ------------------------------------------------------------------

    def _load_groq(self) -> None:
        self._groq_api_key = os.environ.get("GROQ_API_KEY", "")
        if not self._groq_api_key:
            raise RuntimeError(
                "STT_PROVIDER=groq but GROQ_API_KEY is not set. "
                "Add GROQ_API_KEY to .env.local."
            )
        # Model name: use STT_MODEL if it looks like a Groq Whisper model,
        # otherwise default to whisper-large-v3-turbo (fastest on Groq).
        groq_whisper_models = {"whisper-large-v3", "whisper-large-v3-turbo"}
        if self.model_size in groq_whisper_models:
            self._groq_model = self.model_size
        else:
            self._groq_model = "whisper-large-v3-turbo"

        print(f"✅ STT provider: Groq ({self._groq_model}) — no local model needed", file=sys.stderr)

    def _transcribe_groq(self, audio_bytes: bytes, audio_np: np.ndarray) -> str:
        import httpx

        # Wrap raw PCM in a WAV container so Groq can decode it
        wav_buf = io.BytesIO()
        with wave.open(wav_buf, "wb") as wf:
            wf.setnchannels(1)
            wf.setsampwidth(2)   # 16-bit PCM
            wf.setframerate(config.AUDIO_SAMPLE_RATE)
            wf.writeframes(audio_bytes)
        wav_bytes = wav_buf.getvalue()

        start = time.perf_counter()
        try:
            with httpx.Client(timeout=10) as client:
                resp = client.post(
                    "https://api.groq.com/openai/v1/audio/transcriptions",
                    headers={"Authorization": f"Bearer {self._groq_api_key}"},
                    files={"file": ("audio.wav", wav_bytes, "audio/wav")},
                    data={"model": self._groq_model, "language": "en", "response_format": "json"},
                )
                if resp.status_code != 200:
                    print(f"[Groq STT] Error {resp.status_code}: {resp.text}", file=sys.stderr)
                    return ""
                text = resp.json().get("text", "").strip()
        except Exception as e:
            print(f"[Groq STT] Error: {e}", file=sys.stderr)
            return ""

        elapsed = time.perf_counter() - start
        if text:
            print(f"STT: {elapsed*1000:.0f}ms → '{text[:60]}'", file=sys.stderr)
        return text


if __name__ == "__main__":
    from audio_listener import AudioListener

    t = Transcriber(
        model_size=config.STT_MODEL,
        device=config.STT_DEVICE,
        compute_type=config.STT_COMPUTE_TYPE,
        provider=config.STT_PROVIDER
    )
    t.load()

    print("Testing silence...", file=sys.stderr)
    assert t.transcribe(b'\x00' * 8000) == "", "Expected empty string for silence"
    print("Silence test passed.", file=sys.stderr)

    listener = AudioListener(config.AUDIO_SAMPLE_RATE, config.AUDIO_CHUNK_SECONDS)
    listener.start()
    print("Speak for 3 seconds...", file=sys.stderr)
    try:
        chunks = [listener.get_chunk() for _ in range(6)]
        result = t.transcribe(b"".join(chunks))
        print(f"Result: {result}")
    finally:
        listener.stop()
