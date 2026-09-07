"""
SessionManager — In-memory store for interview session history.

Tracks question/answer pairs with timestamps for post-session scoring.
Reset with clear(). Singleton — imported as `session` from this module.
"""
import time
from dataclasses import dataclass
from typing import Optional


@dataclass
class Turn:
    """One Q&A exchange in an interview session."""
    turn_index: int
    transcript: str          # What the user said / question heard
    response: str            # LLM's answer
    started_at: float        # Unix timestamp when transcript was received
    answered_at: float       # Unix timestamp when LLM response completed
    speaker: str = "user"    # "user" or "interviewer" (for diarized sessions)

    @property
    def latency_ms(self) -> int:
        return int((self.answered_at - self.started_at) * 1000)

    def to_dict(self) -> dict:
        return {
            "turn": self.turn_index,
            "transcript": self.transcript,
            "response": self.response,
            "started_at": self.started_at,
            "answered_at": self.answered_at,
            "latency_ms": self.latency_ms,
            "speaker": self.speaker,
        }


class SessionManager:
    """
    Singleton session manager. Access via module-level `session` instance.

    Usage:
        from session_manager import session
        session.start_turn("what is a binary tree?")
        session.complete_turn("A binary tree is...")
        data = session.export()
    """

    def __init__(self):
        self._turns: list[Turn] = []
        self._current_transcript: Optional[str] = None
        self._current_response_parts: list[str] = []
        self._turn_started_at: Optional[float] = None
        self._session_started_at: float = time.time()

    def start_turn(self, transcript: str, speaker: str = "user") -> None:
        """Call when a new question/transcript is received from STT."""
        self._current_transcript = transcript
        self._current_response_parts = []
        self._turn_started_at = time.time()

    def append_response_token(self, token: str) -> None:
        """Call for each streaming token from the LLM."""
        self._current_response_parts.append(token)

    def complete_turn(self) -> Optional[Turn]:
        """
        Call when the LLM response is complete.
        Saves the turn and returns it. Returns None if no turn was in progress.
        """
        if not self._current_transcript or self._turn_started_at is None:
            return None

        turn = Turn(
            turn_index=len(self._turns),
            transcript=self._current_transcript,
            response="".join(self._current_response_parts),
            started_at=self._turn_started_at,
            answered_at=time.time(),
        )
        self._turns.append(turn)
        self._current_transcript = None
        self._current_response_parts = []
        self._turn_started_at = None
        return turn

    def clear(self) -> None:
        """Reset for a new interview session."""
        self.__init__()

    def export(self) -> dict:
        """Return full session data as a JSON-serializable dict."""
        return {
            "session_started_at": self._session_started_at,
            "session_duration_s": int(time.time() - self._session_started_at),
            "turn_count": len(self._turns),
            "turns": [t.to_dict() for t in self._turns],
        }

    @property
    def turns(self) -> list[Turn]:
        return list(self._turns)

    @property
    def has_data(self) -> bool:
        return len(self._turns) > 0


# Module-level singleton — import this everywhere
session = SessionManager()
