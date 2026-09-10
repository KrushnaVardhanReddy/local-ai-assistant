import os
import asyncio
import sys
from typing import Optional
from config import config

_instance: Optional["LocalIntelligence"] = None

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
        if not self._enabled or self._llm is None:
            return True
        prompt = (
            f"<|system|>You are a turn-detection classifier. Reply with exactly one word.<|end|>\n"
            f"<|user|>Transcript: \"{text}\"\n"
            f"Is this a complete thought? Reply COMPLETE or INCOMPLETE.<|end|>\n"
            f"<|assistant|>"
        )
        try:
            output = self._llm(prompt, max_tokens=5, temperature=0.0, stop=["\n", "<"])
            answer = output["choices"][0]["text"].strip().upper()
            return answer == "COMPLETE"
        except Exception as e:
            print(f"[SmolLM2] is_complete error: {e}", file=sys.stderr)
            return True

    def encode(self, text: str) -> list:
        """Returns a semantic embedding vector. Returns empty list if disabled."""
        if not self._enabled or self._llm is None:
            return []
        try:
            return self._llm.embed(text)
        except Exception as e:
            print(f"[SmolLM2] encode error: {e}", file=sys.stderr)
            return []

    async def is_question(self, text: str) -> bool:
        """Returns True if the text appears to be a meaningful question or request using SmolLM2."""
        if not self._enabled or self._llm is None:
            return True

        prompt = (
            f"You are a strict classification engine. Does the following text require a factual answer, explanation, or response from an AI? Reply ONLY with YES or NO. Text: '{text}'"
        )

        try:
            def _run_llm():
                return self._llm(prompt, max_tokens=5, temperature=0.0)

            output = await asyncio.to_thread(_run_llm)
            answer = output["choices"][0]["text"].strip().upper()
            return answer.startswith("YES")
        except Exception as e:
            print(f"[SmolLM2] is_question error: {e}", file=sys.stderr)
            return True
