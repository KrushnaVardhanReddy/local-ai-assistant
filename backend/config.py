import os
import urllib.parse
from dataclasses import dataclass

PROVIDER_CONFIG = {
    "ollama":   {"url": "http://localhost:11434/v1", "key_env": None},
    "lmstudio": {"url": "http://localhost:1234/v1",  "key_env": None},
    "llamacpp": {"url": "http://localhost:8080/v1",  "key_env": None},
    "openai":   {"url": "https://api.openai.com/v1", "key_env": "OPENAI_API_KEY"},
    "groq":     {"url": "https://api.groq.com/openai/v1", "key_env": "GROQ_API_KEY"},
    "gemini":   {"url": "https://generativelanguage.googleapis.com/v1beta/openai/", "key_env": "GEMINI_API_KEY"},
    "anthropic":{"url": "https://api.anthropic.com/v1", "key_env": "ANTHROPIC_API_KEY"}
}

def _load_env():
    env_dict = {}
    # Assume config.py is in backend/ and .env is in the repo root.
    env_path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), ".env.local")
    if os.path.exists(env_path):
        with open(env_path, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("#"):
                    continue
                if "=" in line:
                    key, value = line.split("=", 1)
                    # Strip whitespace from keys and values
                    env_dict[key.strip()] = value.strip()
    return env_dict

# Merge env file dict into os.environ at startup so all modules see env vars uniformly
for k, v in _load_env().items():
    if k not in os.environ:
        os.environ[k] = v

@dataclass
class Config:
    LLM_PROVIDER: str = "auto"
    LLM_MODEL: str = "llama3"
    LLM_BASE_URL: str = ""
    LLM_API_KEY: str = ""
    VISION_MODEL: str = "gpt-4o-mini"
    GEMINI_API_KEY: str = ""
    GEMINI_LIVE_MODEL: str = "gemini-2.0-flash-live-001"

    STT_MODEL: str = "base"
    STT_DEVICE: str = "cuda"
    STT_COMPUTE_TYPE: str = "float16"
    STT_PROVIDER: str = "local"  # "local" = faster-whisper on device, "groq" = Groq Whisper API
    STT_DIARIZE: bool = False

    AUDIO_SAMPLE_RATE: int = 16000
    AUDIO_CHUNK_SECONDS: float = 0.5

    WS_HOST: str = "127.0.0.1"
    WS_PORT: int = 8765

    SUPABASE_JWT_SECRET: str = ""
    SUPABASE_URL: str = ""
    SUPABASE_SERVICE_KEY: str = ""
    ENCRYPTION_KEY: str = ""
    STRIPE_WEBHOOK_SECRET: str = ""

    CHROMA_DIR: str = ""
    UPLOAD_DIR: str = ""
    EMBEDDING_MODEL: str = "all-MiniLM-L6-v2"
    RAG_TOP_K: int = 4

    RAG_ENABLED: bool = True
    WEB_SEARCH_ENABLED: bool = False
    SYSTEM_PROMPT: str = "You are a stealth interview assistant. The user is in a live technical interview. You must provide EXTREMELY concise, direct solutions. DO NOT repeat the question or the constraints. DO NOT output conversational filler. If code is needed, provide only the core snippet in the requested programming language."

    # Smart Audio Filter config
    SILENCE_THRESHOLD_SECONDS: float = 1.5
    MIN_WORDS: int = 4
    SMART_FILTER_ENABLED: bool = True

    COACHING_ENABLED: bool = False
    JOB_DESCRIPTION: str = ""

    def __post_init__(self):
        # Override fields with os.environ
        self.RAG_ENABLED = os.environ.get("RAG_ENABLED", "true").lower() == "true"
        self.WEB_SEARCH_ENABLED = os.environ.get("WEB_SEARCH_ENABLED", "false").lower() == "true"
        self.SYSTEM_PROMPT = os.environ.get("SYSTEM_PROMPT", self.SYSTEM_PROMPT)

        self.SILENCE_THRESHOLD_SECONDS = float(os.environ.get("SILENCE_THRESHOLD_SECONDS", self.SILENCE_THRESHOLD_SECONDS))
        self.MIN_WORDS = int(os.environ.get("MIN_WORDS", self.MIN_WORDS))
        self.SMART_FILTER_ENABLED = os.environ.get("SMART_FILTER_ENABLED", "true").lower() == "true"

        self.COACHING_ENABLED = os.environ.get("COACHING_ENABLED", "false").lower() == "true"

        self.LLM_PROVIDER = os.environ.get("LLM_PROVIDER", self.LLM_PROVIDER)
        self.LLM_MODEL = os.environ.get("LLM_MODEL", self.LLM_MODEL)
        self.LLM_BASE_URL = os.environ.get("LLM_BASE_URL", self.LLM_BASE_URL)
        self.LLM_API_KEY = os.environ.get("LLM_API_KEY", self.LLM_API_KEY)
        self.VISION_MODEL = os.environ.get("VISION_MODEL", self.VISION_MODEL)
        self.GEMINI_API_KEY = os.environ.get("GEMINI_API_KEY", self.GEMINI_API_KEY)
        self.GEMINI_LIVE_MODEL = os.environ.get("GEMINI_LIVE_MODEL", self.GEMINI_LIVE_MODEL)

        self.STT_MODEL = os.environ.get("STT_MODEL", self.STT_MODEL)
        self.STT_DEVICE = os.environ.get("STT_DEVICE", self.STT_DEVICE)
        self.STT_COMPUTE_TYPE = os.environ.get("STT_COMPUTE_TYPE", self.STT_COMPUTE_TYPE)
        self.STT_PROVIDER = os.environ.get("STT_PROVIDER", self.STT_PROVIDER)
        self.STT_DIARIZE = os.environ.get("STT_DIARIZE", "false").lower() == "true"

        self.AUDIO_SAMPLE_RATE = int(os.environ.get("AUDIO_SAMPLE_RATE", self.AUDIO_SAMPLE_RATE))
        self.AUDIO_CHUNK_SECONDS = float(os.environ.get("AUDIO_CHUNK_SECONDS", self.AUDIO_CHUNK_SECONDS))

        self.WS_HOST = os.environ.get("WS_HOST", self.WS_HOST)
        self.WS_PORT = int(os.environ.get("WS_PORT", self.WS_PORT))

        self.SUPABASE_JWT_SECRET = os.environ.get("SUPABASE_JWT_SECRET", self.SUPABASE_JWT_SECRET)
        self.SUPABASE_URL = os.environ.get("SUPABASE_URL", self.SUPABASE_URL)
        self.SUPABASE_SERVICE_KEY = os.environ.get("SUPABASE_SERVICE_KEY", self.SUPABASE_SERVICE_KEY)
        self.ENCRYPTION_KEY = os.environ.get("ENCRYPTION_KEY", self.ENCRYPTION_KEY)
        self.STRIPE_WEBHOOK_SECRET = os.environ.get("STRIPE_WEBHOOK_SECRET", self.STRIPE_WEBHOOK_SECRET)

        self.CHROMA_DIR = os.environ.get("CHROMA_DIR", "./data/chroma")
        self.UPLOAD_DIR = os.environ.get("UPLOAD_DIR", "./data/uploads")
        self.EMBEDDING_MODEL = os.environ.get("EMBEDDING_MODEL", "all-MiniLM-L6-v2")
        self.RAG_TOP_K = int(os.environ.get("RAG_TOP_K", self.RAG_TOP_K))

        # Create necessary directories
        import pathlib
        pathlib.Path(self.CHROMA_DIR).mkdir(parents=True, exist_ok=True)
        pathlib.Path(self.UPLOAD_DIR).mkdir(parents=True, exist_ok=True)

    def resolved_llm(self) -> dict:
        is_cloud = self.LLM_PROVIDER.lower() not in {"auto", "ollama", "lmstudio", "llamacpp"}

        if self.LLM_PROVIDER == "auto":
            return {
                "base_url": self.LLM_BASE_URL,
                "api_key": self.LLM_API_KEY,
                "provider": "auto",
                "is_cloud": is_cloud
            }

        provider_info = PROVIDER_CONFIG.get(self.LLM_PROVIDER.lower())

        base_url = self.LLM_BASE_URL
        api_key = self.LLM_API_KEY

        if provider_info:
            if not base_url:
                base_url = provider_info["url"]
            if not api_key and provider_info["key_env"]:
                api_key = os.environ.get(provider_info["key_env"], "")

        return {
            "base_url": base_url,
            "api_key": api_key,
            "provider": self.LLM_PROVIDER,
            "is_cloud": is_cloud
        }

    def summary(self):
        print("===============================================================")
        print(f"{'Config Summary':^63}")
        print("===============================================================")
        for k, v in self.__dict__.items():
            if "KEY" in k or "SECRET" in k:
                display_v = "***" if v else ""
            else:
                display_v = v
            print(f"{k:<30} | {display_v}")
        print("===============================================================")

        if self.resolved_llm()["is_cloud"]:
            print("⚠️  PRIVACY WARNING: LLM responses are routed through an external cloud provider.")
            print("===============================================================")

config = Config()
