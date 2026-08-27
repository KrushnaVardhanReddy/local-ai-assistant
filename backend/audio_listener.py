import sys
import queue
import math
import numpy as np
import sounddevice as sd

from config import config

class AudioListener:
    def __init__(self, sample_rate: int, chunk_seconds: float):
        self.sample_rate = sample_rate
        self.chunk_seconds = chunk_seconds
        self.chunk_frames = int(sample_rate * chunk_seconds)
        self._queue: queue.Queue[bytes] = queue.Queue()
        self._stream = None

    def start(self) -> None:
        try:
            device_info = sd.query_devices(kind='input')
            device_name = device_info.get('name', 'Default Microphone')

            chunk_ms = int(self.chunk_seconds * 1000)
            print(f"🎤 Listening on: {device_name} | {self.sample_rate}Hz | {chunk_ms}ms chunks")

            self._stream = sd.RawInputStream(
                dtype="int16",
                channels=1,
                samplerate=self.sample_rate,
                blocksize=self.chunk_frames,
                callback=self._callback
            )
            self._stream.start()
        except sd.PortAudioError as e:
            print(f"Error initializing microphone: {e}", file=sys.stderr)
            sys.exit(1)
        except Exception as e:
            print(f"Unexpected error initializing microphone: {e}", file=sys.stderr)
            sys.exit(1)

    def _callback(self, indata, frames, time, status):
        if status:
            print(f"Audio status: {status}", file=sys.stderr)

        # indata is a buffer/memoryview, so we copy it to bytes
        self._queue.put(bytes(indata))

    def stop(self) -> None:
        if self._stream:
            self._stream.stop()
            self._stream.close()
            self._stream = None

    def get_chunk(self) -> bytes:
        # Blocks up to 2 seconds waiting for a chunk of audio
        return self._queue.get(timeout=2)


if __name__ == "__main__":
    listener = AudioListener(config.AUDIO_SAMPLE_RATE, config.AUDIO_CHUNK_SECONDS)
    listener.start()

    try:
        while True:
            try:
                chunk = listener.get_chunk()

                # Convert bytes back to int16 numpy array to calculate RMS
                # We use int32 inside the mean to avoid overflow when squaring
                data = np.frombuffer(chunk, dtype=np.int16).astype(np.int32)

                if len(data) > 0:
                    rms = math.sqrt(np.mean(data**2))
                    print(f"RMS: {rms:.2f}")
                else:
                    print("Empty chunk")

            except queue.Empty:
                print("Queue empty, waiting...")

    except KeyboardInterrupt:
        listener.stop()
        print("\nStopped.")
