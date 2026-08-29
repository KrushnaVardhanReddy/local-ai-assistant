import sys
import queue
import math
import threading
import numpy as np
import sounddevice as sd

from config import config

def get_audio_devices() -> list[dict]:
    import sounddevice as sd
    devices = []
    try:
        for idx, dev in enumerate(sd.query_devices()):
            hostapi = sd.query_hostapis(dev['hostapi'])['name']
            is_wasapi = 'WASAPI' in hostapi
            if dev['max_input_channels'] > 0 or (is_wasapi and dev['max_output_channels'] > 0):
                devices.append({
                    "id": idx,
                    "name": f"{dev['name']} ({hostapi})",
                    "is_loopback_capable": is_wasapi and dev['max_output_channels'] > 0
                })
    except Exception as e:
        print(f"Error enumerating devices: {e}", file=sys.stderr)
    return devices


class AudioListener:
    def __init__(self, sample_rate: int, chunk_seconds: float, device: int | None = None, is_loopback: bool = False):
        self.sample_rate = sample_rate
        self.chunk_seconds = chunk_seconds
        self.chunk_frames = int(sample_rate * chunk_seconds)
        self.device = device
        self.is_loopback = is_loopback
        self._queue: queue.Queue[bytes] = queue.Queue()
        self._stream = None
        self._paused = threading.Event()
        self._paused.set()

    def start(self) -> None:
        try:
            if self.device is not None:
                device_info = sd.query_devices(self.device)
            else:
                device_info = sd.query_devices(kind='input')

            device_name = device_info.get('name', 'Default Microphone')

            chunk_ms = int(self.chunk_seconds * 1000)
            print(f"🎤 Listening on: {device_name} | {self.sample_rate}Hz | {chunk_ms}ms chunks")

            kwargs = {}
            if getattr(self, 'is_loopback', False):
                kwargs['loopback'] = True

            self._stream = sd.RawInputStream(
                device=self.device,
                dtype="int16",
                channels=1,
                samplerate=self.sample_rate,
                blocksize=self.chunk_frames,
                callback=self._callback,
                **kwargs
            )
            self._stream.start()
        except sd.PortAudioError as e:
            print(f"Error initializing microphone: {e}", file=sys.stderr)
            pass
        except Exception as e:
            print(f"Unexpected error initializing microphone: {e}", file=sys.stderr)
            pass

    def _callback(self, indata, frames, time, status):
        if not self._paused.is_set():
            return

        if status:
            print(f"Audio status: {status}", file=sys.stderr)

        # indata is a buffer/memoryview, so we copy it to bytes
        self._queue.put(bytes(indata))

    def pause(self):
        self._paused.clear()
        while not self._queue.empty():
            try:
                self._queue.get_nowait()
            except queue.Empty:
                break

    def resume(self):
        self._paused.set()

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
