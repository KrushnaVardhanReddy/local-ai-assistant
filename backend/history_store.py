"""
HistoryStore — Lightweight local JSON persistence for session scorecard history.

Stores a rolling log of the last MAX_ENTRIES session summaries in a
platform-appropriate user data directory. No database required.
"""
import json
import os
import time
from pathlib import Path

MAX_ENTRIES = 50


def _get_history_path() -> Path:
    """Return platform-appropriate path for the history JSON file."""
    if os.name == "nt":  # Windows
        base = Path(os.environ.get("APPDATA", Path.home() / "AppData" / "Roaming"))
    else:  # Linux / macOS
        base = Path(os.environ.get("XDG_DATA_HOME", Path.home() / ".local" / "share"))
    data_dir = base / "parakeet"
    data_dir.mkdir(parents=True, exist_ok=True)
    return data_dir / "history.json"


def append_session(scorecard: dict, session_data: dict) -> None:
    """
    Append a lightweight summary of a completed session to the history file.
    Silently no-ops if the file cannot be written (e.g., permissions).
    """
    try:
        path = _get_history_path()
        # Load existing entries
        entries = []
        if path.exists():
            try:
                entries = json.loads(path.read_text(encoding="utf-8"))
                if not isinstance(entries, list):
                    entries = []
            except (json.JSONDecodeError, OSError):
                entries = []

        # Build lightweight summary
        entry = {
            "date_iso": time.strftime(
                "%Y-%m-%dT%H:%M:%S",
                time.localtime(session_data.get("session_started_at", time.time()))
            ),
            "overall_score": scorecard.get("overall_score", 0),
            "turn_count": session_data.get("turn_count", 0),
            "duration_s": session_data.get("session_duration_s", 0),
            "strengths": scorecard.get("strengths", [])[:2],   # top 2 only
            "gaps": scorecard.get("gaps", [])[:2],              # top 2 only
            "overall_summary": scorecard.get("overall_summary", "")[:200],
        }

        entries.append(entry)

        # Cap at MAX_ENTRIES — drop oldest
        if len(entries) > MAX_ENTRIES:
            entries = entries[-MAX_ENTRIES:]

        path.write_text(json.dumps(entries, indent=2, ensure_ascii=False), encoding="utf-8")
    except Exception:
        pass  # History is best-effort — never crash the main flow


def load_history() -> list[dict]:
    """
    Return the last MAX_ENTRIES session summaries, newest first.
    Returns empty list if the file doesn't exist or is corrupt.
    """
    try:
        path = _get_history_path()
        if not path.exists():
            return []
        entries = json.loads(path.read_text(encoding="utf-8"))
        if not isinstance(entries, list):
            return []
        return list(reversed(entries))  # newest first
    except Exception:
        return []
