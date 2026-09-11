import os
import asyncio
import sys
import pickle
from typing import Optional
from config import config

_instance: Optional["LocalIntelligence"] = None

SYSTEM_PROMPT_INJECTIONS = {
    "behavioral": "Structure your response using STAR format (Situation, Task, Action, Result). Be concise — 4 bullet points maximum.",
    "coding": "Provide: 1) Pseudocode or a minimal code snippet, 2) Time/Space complexity, 3) One edge case to watch for. No lengthy prose.",
    "system_design": "Structure as: 1) Clarify requirements, 2) High-level components, 3) Data flow, 4) Scalability considerations. Use bullet points.",
    "conceptual": "Give a clear definition, followed by a brief example or core syntax snippet, and 1-2 key properties. Keep it concise but practical.",
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
        self._classifier = None
        self._enabled = config.LOCAL_EMBEDDING_ENABLED

        # Cleanup old SmolLM2 model if it exists
        smollm_path = os.path.join(os.path.dirname(__file__), "models", "smollm2-135m-instruct-q4_k_m.gguf")
        if os.path.exists(smollm_path):
            try:
                os.remove(smollm_path)
                print(f"[Nomic] Removed deprecated SmolLM2 model at {smollm_path}", file=sys.stderr)
            except Exception as e:
                print(f"[Nomic] Failed to remove deprecated SmolLM2 model: {e}", file=sys.stderr)

        if self._enabled:
            self._load_model()
        self._load_classifier()

    def _load_classifier(self):
        try:
            model_path = os.path.join(os.path.dirname(__file__), "models", "question_classifier.pkl")
            if os.path.exists(model_path):
                with open(model_path, "rb") as f:
                    self._classifier = pickle.load(f)
                print(f"[Nomic] Loaded classifier from {model_path}", file=sys.stderr)
            else:
                print(f"[Nomic] Classifier not found at {model_path}", file=sys.stderr)
        except Exception as e:
            print(f"[Nomic] Failed to load classifier: {e}", file=sys.stderr)

    def _load_model(self):
        if not self._enabled:
            return

        try:
            from llama_cpp import Llama
            model_path = config.LOCAL_EMBEDDING_MODEL_PATH

            if not os.path.exists(model_path):
                print("[Nomic] First run detected. Downloading ~80MB local intelligence model...", file=sys.stderr)
                os.makedirs(os.path.dirname(model_path), exist_ok=True)

                import urllib.request
                url = "https://huggingface.co/nomic-ai/nomic-embed-text-v1.5-GGUF/resolve/main/nomic-embed-text-v1.5.Q4_K_M.gguf"

                def _progress_hook(count, block_size, total_size):
                    percent = int(count * block_size * 100 / total_size)
                    sys.stderr.write(f"\r[Nomic] Downloading: {percent}%")
                    sys.stderr.flush()

                try:
                    urllib.request.urlretrieve(url, model_path, reporthook=_progress_hook)
                    print("\n[Nomic] Download complete.", file=sys.stderr)
                except Exception as e:
                    print(f"\n[Nomic] Download failed: {e}. Disabling.", file=sys.stderr)
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
            print(f"[Nomic] Local intelligence loaded from {model_path}", file=sys.stderr)
        except ImportError:
            print("[Nomic] llama-cpp-python not installed. Disabling.", file=sys.stderr)
            self._enabled = False
        except Exception as e:
            print(f"[Nomic] Failed to load model: {e}", file=sys.stderr)
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
            print(f"[Nomic] encode error: {e}", file=sys.stderr)
            return []

    async def is_question(self, text: str) -> bool:
        """Returns True if the text appears to be a meaningful question or request using Nomic."""
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

    _KEYWORD_RULES = [
        # coding — unambiguous imperative verb openers
        (r'^\s*(write|implement|code|create a function|build a function|design a function)\b', "coding"),
        # system_design — design-at-scale openers
        (r'^\s*(design\s+(?:the|a|an)\s+(?:architecture|system|database|infrastructure)|how would you (design|architect))\b', "system_design"),
        # behavioral — STAR signal phrases
        (r'\b(tell me about a time|describe a situation|give me an example of|walk me through a (time|challenge))\b', "behavioral"),
        # opinion — preference/opinion openers
        (r'^\s*(do you prefer|what (is your|do you) (opinion|take|preference|preferred)|which do you prefer|what do you think about)\b', "opinion"),
        # conceptual — definition/explanation openers
        (r'^\s*(what is (the|a|an) |explain (how|what|the|why)|how does .+ work)\b', "conceptual"),
    ]

    def _keyword_classify(self, transcript: str) -> str | None:
        """Fast keyword pre-filter. Returns category if confident, else None."""
        import re
        t = transcript.strip().lower()
        for pattern, category in self._KEYWORD_RULES:
            if re.search(pattern, t, re.IGNORECASE):
                return category
        return None

    def classify_question(self, transcript: str) -> str:
        """Classify transcript into a category using keyword pre-filter + ML classifier."""
        if not self._enabled or self._llm is None:
            return "conceptual"
        try:
            # Stage 1: fast keyword rules for unambiguous patterns
            keyword_result = self._keyword_classify(transcript)
            if keyword_result is not None:
                return keyword_result

            # Stage 2: ML classifier on top of Nomic embeddings
            query_vec = self.encode(transcript)
            if not query_vec:
                return "conceptual"

            if self._classifier is not None:
                try:
                    prediction = self._classifier.predict([query_vec])[0]
                    return prediction
                except Exception as e:
                    print(f"[Nomic] classifier prediction error: {e}", file=sys.stderr)

            # Stage 3: cosine similarity fallback
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
            print(f"[Nomic] classify error: {e}", file=sys.stderr)
            return "conceptual"
