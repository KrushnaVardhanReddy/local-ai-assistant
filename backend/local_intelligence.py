import os
import asyncio
import sys
from typing import Optional
from config import config

_instance: Optional["LocalIntelligence"] = None

SYSTEM_PROMPT_INJECTIONS = {
    "behavioral": "Structure your response using STAR format (Situation, Task, Action, Result). Be concise — 4 bullet points maximum.",
    "coding": "Provide: 1) Pseudocode or a minimal code snippet, 2) Time/Space complexity, 3) One edge case to watch for. No lengthy prose.",
    "system_design": "Structure as: 1) Clarify requirements, 2) High-level components, 3) Data flow, 4) Scalability considerations. Use bullet points.",
    "conceptual": "Give a concise 1-sentence definition, then 2-3 key properties or features. No padding.",
    "opinion": "State a clear 1-sentence opinion, then give 2 concrete reasons from your experience.",
    "noise": None
}

def get_local_intelligence() -> "LocalIntelligence":
    global _instance
    if _instance is None:
        _instance = LocalIntelligence()
    return _instance

class LocalIntelligence:
    def __init__(self):
        self._llm = None
        self._enabled = config.SMOLLM2_ENABLED
        if self._enabled:
            self._load_model()

    def _load_model(self):
        if not self._enabled:
            return

        try:
            from llama_cpp import Llama
            model_path = config.SMOLLM2_MODEL_PATH

            if not os.path.exists(model_path):
                print("[SmolLM2] First run detected. Downloading ~100MB local intelligence model...", file=sys.stderr)
                os.makedirs(os.path.dirname(model_path), exist_ok=True)

                import urllib.request
                url = "https://huggingface.co/bartowski/SmolLM2-135M-Instruct-GGUF/resolve/main/SmolLM2-135M-Instruct-Q4_K_M.gguf"

                def _progress_hook(count, block_size, total_size):
                    percent = int(count * block_size * 100 / total_size)
                    sys.stderr.write(f"\r[SmolLM2] Downloading: {percent}%")
                    sys.stderr.flush()

                try:
                    urllib.request.urlretrieve(url, model_path, reporthook=_progress_hook)
                    print("\n[SmolLM2] Download complete.", file=sys.stderr)
                except Exception as e:
                    print(f"\n[SmolLM2] Download failed: {e}. Disabling.", file=sys.stderr)
                    self._enabled = False
                    if os.path.exists(model_path):
                        os.remove(model_path)
                    return

            self._llm = Llama(
                model_path=model_path,
                n_ctx=512,
                n_threads=4,
                embedding=True,
                verbose=False,
            )
            print(f"[SmolLM2] Local intelligence loaded from {model_path}", file=sys.stderr)
        except ImportError:
            print("[SmolLM2] llama-cpp-python not installed. Disabling.", file=sys.stderr)
            self._enabled = False
        except Exception as e:
            print(f"[SmolLM2] Failed to load model: {e}", file=sys.stderr)
            self._enabled = False

    def is_complete(self, text: str) -> bool:
        """Returns True if the transcript is a complete thought. Falls back to True if disabled."""
        # The 135M model struggles heavily with zero-shot classification and drops valid queries.
        # Bypassing this for now so the app is actually usable.
        return True

    def encode(self, text: str) -> list:
        """Returns a semantic embedding vector. Returns empty list if disabled."""
        if not self._enabled or self._llm is None:
            return []
        try:
            res = self._llm.embed(text)
            # llama_cpp.embed returns a list of embeddings (list of lists) even for a single string
            if res and isinstance(res, list) and isinstance(res[0], list):
                return res[0]
            return res
        except Exception as e:
            print(f"[SmolLM2] encode error: {e}", file=sys.stderr)
            return []

    async def is_question(self, text: str) -> bool:
        """Returns True if the text appears to be a meaningful question or request using SmolLM2."""
        # The 135M model struggles heavily with zero-shot classification and drops valid queries.
        # Bypassing this for now; smart_filter.py already catches basic conversational filler.
        return True

    def classify_question(self, transcript: str) -> str:
        if not self._enabled or self._llm is None:
            return "conceptual"

        prompt = f"""Classify this interview question into exactly one category.
Categories: behavioral, coding, system_design, conceptual, opinion, noise
Question: {transcript}
Answer with only the category name."""

        try:
            res = self._llm.create_chat_completion(
                messages=[{"role": "user", "content": prompt}],
                max_tokens=10,
                temperature=0.0
            )
            output = res["choices"][0]["message"]["content"].strip().lower()

            valid = ["behavioral", "coding", "system_design", "conceptual", "opinion", "noise"]
            for category in valid:
                if category in output:
                    return category
            return "conceptual"
        except Exception as e:
            print(f"[SmolLM2] classify error: {e}", file=sys.stderr)
            return "conceptual"
