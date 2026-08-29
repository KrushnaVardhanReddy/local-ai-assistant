"""
Smart Audio Filter — three-layer guard for deciding whether a transcript
should be forwarded to the LLM or silently dropped.
"""
import re
import time
import sys
import threading
from config import config

FILLER_PHRASES = {
    "okay", "ok", "yeah", "yes", "no", "mm-hmm", "hmm", "uh", "um",
    "right", "sure", "thanks", "thank you", "bye", "alright", "all right",
    "got it", "i see", "i know", "oh", "ah", "so", "yep", "nope",
}

QUESTION_SIGNALS = [
    r"\?$",
    r"\b(what|how|why|when|where|which|who|explain|describe)\b",
    r"\b(can you|could you|should|is there|are there|tell me|help me)\b",
    r"\b(difference between|what is|what are|how do|how does|how can)\b",
    r"\b(give me|show me|write|create|generate|provide)\b",
]

def is_question(text: str) -> bool:
    """Returns True if the text appears to be a meaningful question or request."""
    lower = text.lower().strip()
    
    # Check for basic negations that negate a command
    negation_pattern = r"\b(don't|doesn't|didn't|can't|won't|never)\b\s+\w*\s*(tell me|show me|explain|give me|write|create|generate)\b"
    if re.search(negation_pattern, lower):
        return False

    for pattern in QUESTION_SIGNALS:
        if re.search(pattern, lower):
            return True
    return False

def is_filler(text: str) -> bool:
    """Returns True if the text is a known noise or filler phrase."""
    normalized = text.lower().strip().rstrip(" .?!,;:")
    return normalized in FILLER_PHRASES

def passes_filter(text: str) -> tuple[bool, str]:
    """
    Run all heuristic filter layers.
    Returns (should_send: bool, reason: str).
    """
    if not config.SMART_FILTER_ENABLED:
        return True, "filter_disabled"

    # Strip non-alphanumeric chars for accurate word count
    clean_text = re.sub(r'[^\w\s]', '', text)
    word_count = len(clean_text.split())

    if word_count < config.MIN_WORDS:
        print(f"[FILTER] Dropped (too short / noise): '{text}'", file=sys.stderr)
        return False, "too_short"

    if is_filler(text):
        print(f"[FILTER] Dropped (filler): '{text}'", file=sys.stderr)
        return False, "filler"

    if not is_question(text):
        print(f"[FILTER] Dropped (not a question): '{text}'", file=sys.stderr)
        return False, "not_question"

    print(f"[INTENT] Accepted: '{text}'", file=sys.stderr)
    return True, "accepted"


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
