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

    # Anchor phrases per category — used for embedding-based semantic routing.
    # More anchors = better accuracy. Averages are taken per category.
    _CATEGORY_ANCHORS = {
        "behavioral": [
            "Tell me about a time you failed.",
            "Describe a situation where you showed leadership.",
            "Give me an example of how you handled a conflict at work.",
            "Walk me through a challenge you overcame.",
        ],
        "coding": [
            "Write a function to reverse a string.",
            "Implement a binary search algorithm.",
            "Write a React hook that debounces a value.",
            "Code a solution for the two-sum problem.",
        ],
        "system_design": [
            "How would you design a scalable chat application?",
            "Design the architecture for a URL shortener.",
            "Walk me through designing a distributed caching system.",
            "How would you architect a real-time collaborative editor?",
        ],
        "conceptual": [
            "What is the Virtual DOM?",
            "Explain how HTTP works.",
            "What is the difference between useMemo and useCallback?",
            "How does garbage collection work in JavaScript?",
        ],
        "opinion": [
            "Do you prefer React or Angular?",
            "What is your opinion on using TypeScript?",
            "Which state management library do you prefer and why?",
            "What do you think about microservices vs monoliths?",
        ],
        "noise": [
            "Okay.",
            "Sounds good, thank you.",
            "Hmm, let me think about that.",
            "Mhm, sure.",
            "Got it.",
        ],
    }
    # Cache for pre-computed anchor embeddings (computed once on first classify call)
    _anchor_embeddings: dict = {}

    def _cosine_similarity(self, a: list, b: list) -> float:
        """Compute cosine similarity between two vectors."""
        dot = sum(x * y for x, y in zip(a, b))
        norm_a = sum(x * x for x in a) ** 0.5
        norm_b = sum(x * x for x in b) ** 0.5
        if norm_a == 0 or norm_b == 0:
            return 0.0
        return dot / (norm_a * norm_b)

    def _get_anchor_embeddings(self) -> dict:
        """Returns pre-computed (and cached) average embeddings for each category."""
        if self._anchor_embeddings:
            return self._anchor_embeddings
        for category, phrases in self._CATEGORY_ANCHORS.items():
            vecs = [self.encode(p) for p in phrases]
            vecs = [v for v in vecs if v]  # filter out empty
            if not vecs:
                continue
            dim = len(vecs[0])
            avg = [sum(v[i] for v in vecs) / len(vecs) for i in range(dim)]
            self._anchor_embeddings[category] = avg
        return self._anchor_embeddings

    def classify_question(self, transcript: str) -> str:
        """Classify transcript into a category using embedding cosine similarity."""
        if not self._enabled or self._llm is None:
            return "conceptual"
        try:
            query_vec = self.encode(transcript)
            if not query_vec:
                return "conceptual"
            anchor_embeddings = self._get_anchor_embeddings()
            best_category = "conceptual"
            best_score = -1.0
            for category, anchor_vec in anchor_embeddings.items():
                score = self._cosine_similarity(query_vec, anchor_vec)
                if score > best_score:
                    best_score = score
                    best_category = category
            return best_category
        except Exception as e:
            print(f"[SmolLM2] classify error: {e}", file=sys.stderr)
            return "conceptual"
