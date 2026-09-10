"""
Smart Audio Filter — three-layer guard for deciding whether a transcript
should be forwarded to the LLM or silently dropped.
"""
import re
import time
import sys
import threading
from config import config
from local_intelligence import get_local_intelligence

FILLER_PHRASES = {
    "okay", "ok", "yeah", "yes", "no", "mm-hmm", "hmm", "uh", "um",
    "right", "sure", "thanks", "thank you", "bye", "alright", "all right",
    "got it", "i see", "i know", "oh", "ah", "so", "yep", "nope",
<<<<<<< HEAD
=======
    # Greetings — should not trigger a code response
    "hi", "hello", "hey", "hi there", "hello there", "hey there",
    "good morning", "good afternoon", "good evening", "good night",
    "how are you", "how's it going", "what's up", "sup",
>>>>>>> origin/main
}

def is_filler(text: str) -> bool:
    """Returns True if the text is a known noise or filler phrase."""
    normalized = text.lower().strip().rstrip(" .?!,;:")
    return normalized in FILLER_PHRASES

async def passes_filter(text: str) -> tuple[bool, str]:
    """
    Run all heuristic filter layers.
    Returns (should_send: bool, reason: str).
    """
<<<<<<< HEAD
=======
    t0 = time.time()
>>>>>>> origin/main
    if not config.SMART_FILTER_ENABLED:
        return True, "filter_disabled"

    # Strip non-alphanumeric chars for accurate word count
    clean_text = re.sub(r'[^\w\s]', '', text)
    word_count = len(clean_text.split())

    if word_count < config.MIN_WORDS:
<<<<<<< HEAD
        print(f"[FILTER] Dropped (too short / noise): '{text}'", file=sys.stderr)
        return False, "too_short"

    if is_filler(text):
        print(f"[FILTER] Dropped (filler): '{text}'", file=sys.stderr)
=======
        print(f"[FILTER] {int((time.time() - t0) * 1000)}ms Dropped (too short / noise): '{text}'", file=sys.stderr)
        return False, "too_short"

    if is_filler(text):
        print(f"[FILTER] {int((time.time() - t0) * 1000)}ms Dropped (filler): '{text}'", file=sys.stderr)
>>>>>>> origin/main
        return False, "filler"

    li = get_local_intelligence()
    if not await li.is_question(text):
<<<<<<< HEAD
        print(f"Ignored conversational filler", file=sys.stderr)
        return False, "not_question"

    if not li.is_complete(text):
        print(f"[INTENT] Dropped (incomplete turn): '{text}'", file=sys.stderr)
        return False, "incomplete_turn"

    print(f"[INTENT] Accepted: '{text}'", file=sys.stderr)
    return True, "accepted"


=======
        print(f"[FILTER] {int((time.time() - t0) * 1000)}ms Ignored conversational filler", file=sys.stderr)
        return False, "not_question"

    if not li.is_complete(text):
        print(f"[INTENT] {int((time.time() - t0) * 1000)}ms Dropped (incomplete turn): '{text}'", file=sys.stderr)
        return False, "incomplete_turn"

    print(f"[INTENT] {int((time.time() - t0) * 1000)}ms Accepted: '{text}'", file=sys.stderr)
    return True, "accepted"



>>>>>>> origin/main
class SilenceBuffer:
    """
    Accumulates transcript fragments. When SILENCE_THRESHOLD_SECONDS of
    silence is detected (no new fragment arrives), flushes the buffer
    as a single assembled thought.
    """

    def __init__(self):
        self._buffer: list[str] = []
        self._last_activity: float = 0.0
        self._lock = threading.Lock()

    def add(self, text: str) -> None:
        with self._lock:
            # Simple deduplication of identical consecutive fragments
            if not self._buffer or self._buffer[-1] != text:
                self._buffer.append(text)
            self._last_activity = time.monotonic()

    def is_silent(self) -> bool:
        with self._lock:
            if not self._buffer:
                return False
            return (time.monotonic() - self._last_activity) >= config.SILENCE_THRESHOLD_SECONDS

    def flush(self) -> str:
        """Returns the assembled text and resets the buffer."""
        with self._lock:
            assembled = " ".join(self._buffer).strip()
            self._buffer = []
            self._last_activity = 0.0
        print(f"[VAD] Assembled: '{assembled}'", file=sys.stderr)
        return assembled

    def is_empty(self) -> bool:
        with self._lock:
            return len(self._buffer) == 0

async def passes_filter_for_speaker(text: str, speaker: str | None) -> tuple[bool, str]:
    """
    When speaker is INTERVIEWER: always passes (return True).
    When speaker is CANDIDATE or None: use normal passes_filter() logic.
    """
    if speaker == "INTERVIEWER":
        return True, "interviewer_bypass"
    return await passes_filter(text)
