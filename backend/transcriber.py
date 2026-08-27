import sys
import time
import numpy as np
from faster_whisper import WhisperModel
from config import config

class Transcriber:
    def __init__(self, model_size: str, device: str, compute_type: str):
        self.model_size = model_size
        self.device = device
        self.compute_type = compute_type
        self.model = None

    def load(self) -> None:
        start_time = time.perf_counter()
        try:
            self.model = WhisperModel(self.model_size, device=self.device, compute_type=self.compute_type)
        except Exception as e:
            if "cuda" in str(e).lower():
                print("⚠️  CUDA unavailable — falling back to CPU")
                self.device = "cpu"
                self.compute_type = "float32"
                self.model = WhisperModel(self.model_size, device=self.device, compute_type=self.compute_type)
            else:
                raise e
        elapsed = time.perf_counter() - start_time
        print(f"✅ STT model loaded in {elapsed:.2f}s on {self.device}")

    def transcribe(self, audio_bytes: bytes) -> str:
        if not self.model:
            raise RuntimeError("Model not loaded. Call load() first.")

        # Pad with zero if uneven
        if len(audio_bytes) % 2 != 0:
            audio_bytes += b'\x00'

        audio_np = np.frombuffer(audio_bytes, dtype=np.int16).astype(np.float32) / 32768.0

        # Silence detection
        if len(audio_np) > 0:
            rms = np.sqrt(np.mean(audio_np**2))
            if rms < 0.01:
                return ""

        start_time = time.perf_counter()
        segments, info = self.model.transcribe(audio_np, beam_size=5, language="en")

        segments_list = list(segments)
        text = " ".join([segment.text for segment in segments_list]).strip()

        if not text.strip():
            return ""

        elapsed = time.perf_counter() - start_time
        print(f"STT: {elapsed*1000:.0f}ms → '{text[:50]}'", file=sys.stderr)

        return text

if __name__ == "__main__":
    from audio_listener import AudioListener

    transcriber = Transcriber(
        model_size=config.STT_MODEL,
        device=config.STT_DEVICE,
        compute_type=config.STT_COMPUTE_TYPE
    )
    transcriber.load()

    # Test silence
    print("Testing silence...", file=sys.stderr)
    silent_bytes = b'\x00' * 8000
    res = transcriber.transcribe(silent_bytes)
    assert res == "", "Expected empty string for silence"
    print("Silence test passed.", file=sys.stderr)

    listener = AudioListener(sample_rate=config.AUDIO_SAMPLE_RATE, chunk_seconds=config.AUDIO_CHUNK_SECONDS)
    listener.start()

    print("Speak for 3 seconds...", file=sys.stderr)
    try:
        audio_chunks = []
        for _ in range(6): # 6 * 0.5s = 3 seconds
            audio_chunks.append(listener.get_chunk())

        full_audio = b''.join(audio_chunks)
        print("Transcribing...", file=sys.stderr)
        result = transcriber.transcribe(full_audio)
        print(f"Result: {result}")
    finally:
        listener.stop()
